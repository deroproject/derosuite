package walletapi

import "os"
import "fmt"
import "time"
import "crypto/rand"
import "crypto/sha1"
import "sync"
import "strings"
import "encoding/hex"
import "encoding/json"
import "encoding/binary"

import "github.com/romana/rlog"
import "github.com/vmihailenco/msgpack"

import bolt "github.com/coreos/bbolt"
import "github.com/blang/semver"
import "golang.org/x/crypto/pbkdf2" // // used to encrypt master password ( so user can change his password anytime)

import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/walletapi/mnemonics"

const FUNDS_BUCKET = "FUNDS"                   // stores all incoming funds, key is global output index encrypted form
const FUNDS_AVAILABLE = "FUNDS_AVAILABLE"      // indices of all funds ready to spend
const FUNDS_SPENT = "FUNDS_SPENT"              // indices of all funds already spent
const FUNDS_SPENT_WHERE = "FUNDS_SPENT_WHERE"  // mapping which output -> spent where
const FUNDS_BUCKET_OUTGOING = "FUNDS_OUTGOING" // stores all tx where our funds were spent

const RING_BUCKET = "RING_BUCKET"         //  to randomly choose ring members when transactions are created
const KEYIMAGE_BUCKET = "KEYIMAGE_BUCKET" // used to track which funds have been spent (only on chain ) and which output was consumed
const SECRET_KEY_BUCKET = "TXSECRETKEY"   // used to keep secret keys for any tx this wallet has created
const TX_OUT_DETAILS_BUCKET = "TX_OUT_DETAILS"   // used to keep secret keys for any tx this wallet has created

const HEIGHT_TO_BLOCK_BUCKET = "H2BLOCK_BUCKET" // used to track height to block hash mapping
const OUR_TX_BUCKET = "OUR_TX_BUCKET"           // this contains all TXs in which we have spent OUR  FUNDS

// the strings are prepended so as there can NEVER be collision between TXID and payment ID
// as paymentID are chosen by users
const TXID = "TXID"   // all TX to output index will have this string prepended
const PAYID = "PAYID" // all payment ID to output index will have this string prepended,

// PAYMENT  ID itself is a bucket, TODO COLLISIONS ???
//const FUNDS_BY_PAYMENT_ID_BUCKET ="PAYMENTID_BUCKET" // each payment id is a bucket itself, which stores OUTPUT_INDEX

//const FUNDS_BY_PAYMENT_ID_BUCKET = 10         // each payment id is a bucket itself
const FUNDS_BY_TX_ID_BUCKET = 11              // each tx id is a bucket
const FUNDS_BY_BLOCK_HEIGHT_BUCKET = 12       // funds sorted by block height
const FUNDS_SPENT_BY_BLOCK_HEIGHT_BUCKET = 13 // funds spent by block height
const NOTES_BY_TXID = 14                      // any notes attached

// the db has 3 buckets, one
// funds ( funds which have arrived ), they can have attached , notes addresses, funds which payment id are linked to
// funds_available, contains only output indices
// funds_spent, contains only output indices
// searchable by payment id and searhable by txid and searchable by block height

// fundstate  : available, spent, spentinpool, fund state works as state machine
// search_by_paymentid ( funds from a specific payment id can be accessed here )
// search_by_txid ( funds from a specific payment id can be accessed here )
// search_by_block_height ( funds from a specific payment id can be accessed here )

// address book will have random number based entries

// see this https://godoc.org/golang.org/x/crypto/pbkdf2
type KDF struct {
	Hashfunction string `json:"hash"` //"SHA1" currently only sha1 is supported
	Keylen       int    `json:"keylen"`
	Iterations   int    `json:"iterations"`
	Salt         []byte `json:"salt"`
}

// this is stored in disk in encrypted form
type Wallet struct {
	Version semver.Version `json:"version"` // database version
	Secret  []byte         `json:"secret"`  // actual unlocker to the DB, depends on password from user, stored encrypted
	// secret key used to encrypt all DB data ( both keys and values )
	// this is always in encrypted form

	KDF KDF `json:"kdf"`

	account           *Account //`json:"-"` // not serialized, we store an encrypted version  // keys, seed language etc settings
	Account_Encrypted []byte   `json:"account_encrypted"`

	pbkdf2_password []byte // used to encrypt metadata on updates
	master_password []byte // single password which never changes

	Daemon_Endpoint   string `json:"-"` // endpoint used to communicate with daemon
	Daemon_Height     uint64 `json:"-"` // used to track  daemon height  ony if wallet in online
	Daemon_TopoHeight int64  `json:"-"` // used to track  daemon topo height  ony if wallet in online

	wallet_online_mode bool // set whether the mode is online or offline
	// an offline wallet can be converted to online mode, calling.
	// SetOffline() and vice versa using SetOnline
	// used to create transaction with this fee rate,
	//if this is lower than network, then created transaction will be rejected by network
	dynamic_fees_per_kb uint64
	quit                chan bool // channel to quit any processing go routines

	db *bolt.DB // access to DB

	rpcserver *RPCServer // reference to RPCserver

	id string // first 8 bytes of wallet address , to put into logs to identify different wallets in case many are active

	transfer_mutex sync.Mutex // to avoid races within the transfer
	//sync.Mutex  // used to syncronise access
	sync.RWMutex
}

const META_BUCKET = "METADATA" // all metadata is stored in this bucket
const BLOCKS_BUCKET = "BLOCKS" // stores height to block hash mapping for later on syncing

var BLOCKCHAIN_UNIVERSE = []byte("BLOCKCHAIN_UNIVERSE") // all main chain txs are stored in this bucket

// when smart contracts are implemented, each will have it's own universe to track and maintain transactions

// this file implements the encrypted data store at rest
func Create_Encrypted_Wallet(filename string, password string, seed crypto.Key) (w *Wallet, err error) {
	rlog.Infof("Creating Wallet from recovery seed")
	w = &Wallet{}
	w.Version, err = semver.Parse("0.0.1-alpha.preview.github")

	if err != nil {
		return
	}

	if _, err = os.Stat(filename); err == nil {
		err = fmt.Errorf("File '%s' already exists", filename)
		rlog.Errorf("err creating wallet %s", err)
		return
	}

	w.db, err = bolt.Open(filename, 0600, &bolt.Options{Timeout: 1 * time.Second})

	if err != nil {
		rlog.Errorf("err opening boltdb file %s", err)
		return
	}

	// generate account keys
	w.account, err = Generate_Account_From_Seed(seed)
	if err != nil {
		return
	}

	// generate a 64 byte key to be used as master Key
	w.master_password = make([]byte, 32, 32)
	_, err = rand.Read(w.master_password)
	if err != nil {
		return
	}

	err = w.Set_Encrypted_Wallet_Password(password) // lock the db with the password

	w.quit = make(chan bool)

	w.id = string((w.account.GetAddress().String())[:8]) // set unique id for logs

	rlog.Infof("Successfully created wallet %s", w.id)
	return
}

// create an encrypted wallet using electrum recovery words
func Create_Encrypted_Wallet_From_Recovery_Words(filename string, password string, electrum_seed string) (w *Wallet, err error) {
	rlog.Infof("Creating Wallet from recovery words")

	language, seed, err := mnemonics.Words_To_Key(electrum_seed)
	if err != nil {
		rlog.Errorf("err parsing recovery words %s", err)
		return
	}
	w, err = Create_Encrypted_Wallet(filename, password, seed)

	if err != nil {
		rlog.Errorf("err creating wallet %s", err)
		return
	}

	w.account.SeedLanguage = language
	rlog.Infof("Successfully created wallet %s", w.id)
	return
}

// create an encrypted wallet using using random data
func Create_Encrypted_Wallet_Random(filename string, password string) (w *Wallet, err error) {
	rlog.Infof("Creating Wallet Randomly")
	w, err = Create_Encrypted_Wallet(filename, password, *crypto.RandomScalar())

	if err != nil {
		rlog.Errorf("err %s", err)
		return
	}
	// TODO setup seed language, default is already english
	rlog.Infof("Successfully created wallet %s", w.id)
	return
}

// create an encrypted wallet using using random data
func Create_Encrypted_Wallet_ViewOnly(filename string, password string, viewkey string) (w *Wallet, err error) {

	var public_spend, private_view crypto.Key
	rlog.Infof("Creating View Only Wallet")
	view_raw, err := hex.DecodeString(strings.TrimSpace(viewkey))
	if len(view_raw) != 64 || err != nil {
		err = fmt.Errorf("View Only key must be 128 chars hexadecimal chars")
		rlog.Errorf("err %s", err)
		return
	}

	copy(public_spend[:], view_raw[:32])
	copy(private_view[:], view_raw[32:64])

	// create encrypted wallet randomly and then swap the keys
	w, err = Create_Encrypted_Wallet(filename, password, *crypto.RandomScalar())

	if err != nil {
		rlog.Errorf("err %s", err)
		return
	}

	// swap the keys
	w.account.Keys.Spendkey_Public = public_spend
	w.account.Keys.Viewkey_Secret = private_view
	w.account.Keys.Viewkey_Public = *(private_view.PublicKey())
	w.account.ViewOnly = true

	w.Save_Wallet() // save wallet data
	rlog.Infof("Successfully created view only wallet %s", w.id)
	return
}

// create an encrypted wallet using using random data
func Create_Encrypted_Wallet_NonDeterministic(filename string, password string, secretkey,viewkey string) (w *Wallet, err error) {

	var secret_spend, secret_view crypto.Key
	rlog.Infof("Creating View Only Wallet")
	spend_raw, err := hex.DecodeString(strings.TrimSpace(secretkey))
	if len(spend_raw) != 32 || err != nil {
		err = fmt.Errorf("View Only key must be 64 chars hexadecimal chars")
		rlog.Errorf("err %s", err)
		return
	}

	copy(secret_spend[:], spend_raw[:32])
	

	view_raw, err := hex.DecodeString(strings.TrimSpace(viewkey))
	if len(view_raw) != 32 || err != nil {
		err = fmt.Errorf("Spend Only key must be 64 chars hexadecimal chars")
		rlog.Errorf("err %s", err)
		return
	}

	copy(secret_view[:], view_raw[:32])

	// create encrypted wallet randomly and then swap the keys
	w, err = Create_Encrypted_Wallet(filename, password, *crypto.RandomScalar())

	if err != nil {
		rlog.Errorf("err %s", err)
		return
	}

	// swap the keys
	w.account.Keys.Spendkey_Secret = secret_spend
	w.account.Keys.Spendkey_Public = *(secret_spend.PublicKey())
	w.account.Keys.Viewkey_Secret = secret_view
	w.account.Keys.Viewkey_Public = *(secret_view.PublicKey())
	
	w.Save_Wallet() // save wallet data
	rlog.Infof("Successfully created view only wallet %s", w.id)
	return
}

// wallet must already be open
func (w *Wallet) Set_Encrypted_Wallet_Password(password string) (err error) {

	if w == nil {
		return
	}
	w.Lock()

	// set up KDF structure
	w.KDF.Salt = make([]byte, 32, 32)
	_, err = rand.Read(w.KDF.Salt)
	if err != nil {
		w.Unlock()
		return
	}
	w.KDF.Keylen = 32
	w.KDF.Iterations = 262144
	w.KDF.Hashfunction = "SHA1"

	// lets generate the bcrypted password

	w.pbkdf2_password = Generate_Key(w.KDF, password)

	w.Unlock()
	w.Save_Wallet() // save wallet data

	return
}

func Open_Encrypted_Wallet(filename string, password string) (w *Wallet, err error) {
	w = &Wallet{}

	if _, err = os.Stat(filename); os.IsNotExist(err) {
		err = fmt.Errorf("File '%s' does NOT exists", filename)
		rlog.Errorf("err opening wallet %s", err)
		return
	}

	w.db, err = bolt.Open(filename, 0600, &bolt.Options{Timeout: 1 * time.Second})

	if err != nil {
		rlog.Errorf("err opening boltdb %s", err)
		return
	}

	// read the metadata from metadat bucket
	w.db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(META_BUCKET))

		v := b.Get([]byte(META_BUCKET))

		if v == nil || len(v) == 0 {
			err = fmt.Errorf("Invalid Database, Could not find meta data")
			rlog.Errorf("err opening wallet %s", err)
			return err
		}

		//fmt.Printf("v %+v\n",string(v)) // DO NOT dump account keys

		// deserialize json data
		err = json.Unmarshal(v, &w)
		if err != nil {
			rlog.Errorf("err parsing metabucket %s", err)
			return err
		}

		w.quit = make(chan bool)
		// todo make any routines necessary, such as sync etc

		return nil
	})

	// try to deseal password and store it
	w.pbkdf2_password = Generate_Key(w.KDF, password)

	// try to decrypt the master password with the pbkdf2
	w.master_password, err = DecryptWithKey(w.pbkdf2_password, w.Secret) // decrypt the master key
	if err != nil {
		rlog.Errorf("err opening secret err: %s ", err)
		err = fmt.Errorf("Invalid Password")
		w.db.Close()
		w = nil
		return
	}

	// password has been  found, open the account

	account_bytes, err := w.Decrypt(w.Account_Encrypted)
	if err != nil {
		rlog.Errorf("err opening account err: %s ", err)
		err = fmt.Errorf("Invalid Password")
		w.db.Close()
		w = nil
		return
	}

	w.account = &Account{} // allocate a new instance
	err = json.Unmarshal(account_bytes, w.account)
	if err != nil {
		return
	}

	w.id = string((w.account.GetAddress().String())[:8]) // set unique id for logs

	return

}

// check whether the already opened wallet can use this password
func (w *Wallet) Check_Password(password string) bool {
	w.Lock()
	defer w.Unlock()
	if w == nil {
		return false
	}

	pbkdf2_password := Generate_Key(w.KDF, password)

	// TODO we can compare pbkdf2_password & w.pbkdf2_password, if they are equal password is vaid

	// try to decrypt the master password with the pbkdf2
	_, err := DecryptWithKey(pbkdf2_password, w.Secret) // decrypt the master key

	if err == nil {
		return true
	}
	rlog.Warnf("%s Invalid Password", w.id)
	return false

}

// save updated copy of wallet
func (w *Wallet) Save_Wallet() (err error) {
	w.Lock()
	defer w.Unlock()
	if w == nil {
		return
	}

	// encrypted the master password with the pbkdf2
	w.Secret, err = EncryptWithKey(w.pbkdf2_password[:], w.master_password) // encrypt the master key
	if err != nil {
		return
	}

	// encrypt the account

	account_serialized, err := json.Marshal(w.account)
	if err != nil {
		return
	}
	w.Account_Encrypted, err = w.Encrypt(account_serialized)
	if err != nil {
		return
	}

	// json marshal wallet data struct, serialize it, encrypt it and store it
	serialized, err := json.Marshal(&w)
	if err != nil {
		return
	}
	//fmt.Printf("Serialized  %+v\n",serialized)

	// let save the secret to DISK in encrypted form
	err = w.db.Update(func(tx *bolt.Tx) (err error) {

		bucket, err := tx.CreateBucketIfNotExists([]byte(META_BUCKET))
		if err != nil {
			return
		}
		err = bucket.Put([]byte(META_BUCKET), serialized)
		return

	})
	rlog.Infof("Saving wallet %s", w.id)
	return
}

// close the wallet
func (w *Wallet) Close_Encrypted_Wallet() {
	close(w.quit)

	time.Sleep(time.Second) // give goroutines some time to quit
	rlog.Infof("Saving and Closing Wallet %s\n", w.id)
	w.Save_Wallet()
	w.db.Sync()

	w.db.Close()
}

// generate key from password
func Generate_Key(k KDF, password string) (key []byte) {
	switch k.Hashfunction {
	case "SHA1":
		return pbkdf2.Key([]byte(password), k.Salt, k.Iterations, k.Keylen, sha1.New)

	default:
		return pbkdf2.Key([]byte(password), k.Salt, k.Iterations, k.Keylen, sha1.New)
	}
}

// check whether a key exists
func (w *Wallet) check_key_exists(universe []byte, subbucket []byte, key []byte) (result bool) {

	//fmt.Printf("Checking %s %s %x \n", string(universe), string(subbucket), key)

	w.db.View(func(tx *bolt.Tx) error {
		universe_bucket := tx.Bucket(w.Key2Key(universe)) //open universe bucket
		if universe_bucket == nil {
			return nil
		}
		bucket := universe_bucket.Bucket(w.Key2Key(subbucket)) // open subbucket
		if bucket == nil {
			return nil
		}

		v := bucket.Get(w.Key2Key(key))

		if v != nil {
			// fmt.Printf("Found\n")
			result = true
		}
		return nil
	})
	return // default is false
}

// delete specified key
func (w *Wallet) delete_key(universe []byte, subbucket []byte, key []byte) {
	rlog.Tracef(1, "Deleting %s %s %x\n", string(universe), string(subbucket), key)

	w.db.Update(func(tx *bolt.Tx) (err error) {
		universe_bucket, err := tx.CreateBucketIfNotExists(w.Key2Key(universe)) //open universe bucket
		if err != nil {
			return
		}
		bucket, err := universe_bucket.CreateBucketIfNotExists(w.Key2Key(subbucket)) // open subbucket
		if err != nil {
			return
		}

		err = bucket.Delete(w.Key2Key(key))

		return err // it will be nil

	})

}

// delete specified key
func (w *Wallet) delete_bucket(universe []byte, subbucket []byte) {
	rlog.Tracef(1, "Deleting bucket %s %s \n", string(universe), string(subbucket))

	w.db.Update(func(tx *bolt.Tx) (err error) {
		universe_bucket, err := tx.CreateBucketIfNotExists(w.Key2Key(universe)) //open universe bucket
		if err != nil {
			return
		}
		err = universe_bucket.DeleteBucket(w.Key2Key(subbucket)) // delete subbucket
		return err                                               // it will be nil

	})

}

// store a key-value, everything is encrypted
func (w *Wallet) store_key_value(universe []byte, subbucket []byte, key []byte, value []byte) error {

	rlog.Tracef(1, "Storing %s %s %x\n", string(universe), string(subbucket), key)

	return w.db.Update(func(tx *bolt.Tx) (err error) {
		universe_bucket, err := tx.CreateBucketIfNotExists(w.Key2Key(universe)) //open universe bucket
		if err != nil {
			return
		}
		bucket, err := universe_bucket.CreateBucketIfNotExists(w.Key2Key(subbucket)) // open subbucket
		if err != nil {
			return
		}

		encrypted_value, err := w.Encrypt(value) // encrypt and seal the value
		if err != nil {
			return
		}
		err = bucket.Put(w.Key2Key(key), encrypted_value)

		return err // it will be nil

	})
}

func (w *Wallet) load_key_value(universe []byte, subbucket []byte, key []byte) (value []byte, err error) {

	rlog.Tracef(1, "loading %s %s %x\n", string(universe), string(subbucket), key)

	w.db.View(func(tx *bolt.Tx) (err error) {
		universe_bucket := tx.Bucket(w.Key2Key(universe)) //open universe bucket
		if universe_bucket == nil {
			return nil
		}
		bucket := universe_bucket.Bucket(w.Key2Key(subbucket)) // open subbucket
		if bucket == nil {
			return nil
		}
		v := bucket.Get(w.Key2Key(key))

		if v == nil {
			return fmt.Errorf("%s %s %x NOT Found", string(universe), string(subbucket), key)
		}

		// fmt.Printf("length of encrypted value %d\n",len(v))

		value, err = w.Decrypt(v)

		return err // it will be nil if everything is alright

	})
	return
}

// enumerate all keys from the bucket
// due to design enumeration is impossible,
// however, all the keys of a specific bucket, where necessary are added as values,
// for example, we never enumerate our funds, if we  donot store them in FUNDS_AVAILABLE or FUNDS_SPENT bucket
// so we find all values in FUNDS_AVAILABLE and FUNDS_SPENT bucket, and then decode value to recover the funds
// this function should only be called for FUNDS_AVAILABLE or FUNDS_SPENT bucket
func (w *Wallet) load_all_values_from_bucket(universe []byte, subbucket []byte) (values [][]byte) {

	w.db.View(func(tx *bolt.Tx) (err error) {
		universe_bucket := tx.Bucket(w.Key2Key(universe)) //open universe bucket
		if universe_bucket == nil {

			return nil
		}
		bucket := universe_bucket.Bucket(w.Key2Key(subbucket)) // open subbucket
		if bucket == nil {
			return nil
		}

		//fmt.Printf("Enumerating Keys\n")
		// Iterate over items
		err = bucket.ForEach(func(k, v []byte) error {
			//  fmt.Printf("Enumerated key\n")
			value, err := w.Decrypt(v)
			if err == nil {
				values = append(values, value)
			}

			return err
		})

		return err // it will be nil if everything is alright

	})
	return
}

func (w *Wallet) load_ring_member(index_global uint64) (r Ring_Member, err error) {

	// store all data
	data_bytes, err := w.load_key_value(BLOCKCHAIN_UNIVERSE, []byte(RING_BUCKET), itob(index_global))

	if err != nil {
		return
	}

	err = msgpack.Unmarshal(data_bytes, &r)

	return
}

// itob returns an 8-byte big endian representation of v.
func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

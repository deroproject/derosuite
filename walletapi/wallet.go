// Copyright 2017-2018 DERO Project. All rights reserved.
// Use of this source code in any form is governed by RESEARCH license.
// license can be found in the LICENSE file.
// GPG: 0F39 E425 8C65 3947 702A  8234 08B2 0360 A03A 9DE8
//
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY
// EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL
// THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
// PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT,
// STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF
// THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package walletapi

import "fmt"
import "net"
import "sort"
import "sync"
import "time"
import "bytes"
import "strings"
import "crypto/rand"
import "encoding/json"
import "encoding/binary"

import "github.com/romana/rlog"
import "github.com/vmihailenco/msgpack"

import "github.com/deroproject/derosuite/config"
import "github.com/deroproject/derosuite/structures"
import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/crypto/ringct"
import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/walletapi/mnemonics"
import "github.com/deroproject/derosuite/address"
import "github.com/deroproject/derosuite/blockchain/inputmaturity"

// used to encrypt payment id
const ENCRYPTED_PAYMENT_ID_TAIL = 0x8d

type _Keys struct {
	Spendkey_Secret crypto.Key `json:"spendkey_secret"`
	Spendkey_Public crypto.Key `json:"spendkey_public"`
	Viewkey_Secret  crypto.Key `json:"viewkey_secret"`
	Viewkey_Public  crypto.Key `json:"viewkey_public"`
}

// all random outputs are stored within wallet in this form
// to be used as ring members
type Ring_Member struct { // structure size is around 74 bytes
	InKey         ringct.CtKey `msgpack:"K"`
	Index_Global  uint64       `msgpack:"I"`
	Height        uint64       `msgpack:"H"`
	Unlock_Height uint64       `msgpack:"U,omitempty"` // this is mostly empty
	Sigtype       uint64       `msgpack:"S,omitempty"` // this is empty for miner tx
}

type Account struct {
	Keys           _Keys   `json:"keys"`
	SeedLanguage   string  `json:"seedlanguage"`
	FeesMultiplier float32 `json:"feesmultiplier"` // fees multiplier accurate to 2 decimals
	Mixin          int     `json:"mixin"`          // default mixn to use for txs

	ViewOnly bool `json:"viewonly"` // is this viewonly wallet

	Index_Global uint64 `json:"index_global"` // till where the indexes have been processed, it must only increase and never decrease
	Height       uint64 `json:"height"`       // block height till where blockchain has been scanned
	TopoHeight   int64  `json:"topoheight"`   // block height till where blockchain has been scanned

	//Wallet_Height    uint64    `json:"wallet_height"`// used to track height till which we have scanned the inputs

	balance_stale  bool   // whether the balance  is stale
	Balance_Mature uint64 `json:"balance_mature"` // total balance of account
	Balance_Locked uint64 `json:"balance_locked"` // balance locked

	random_percent uint64 // number of outputs to store within the db, for mixing, default is 10%

	key_image_checklist map[crypto.Key]bool // key images which need to be monitored, this is updated when new funds arrive

	//Outputs_Array []TX_Wallet_Data // all outputs found in the chain belonging to us, as found in chain

	// uint64 si the Index_Global which is the unique number
	//Outputs_Index    map[uint64]bool               // all outputs which  are ours for deduplication
	//Outputs_Ready    map[uint64]TX_Wallet_Data     // these outputs are ready for consumption ( maturity needs to be checked)
	//Keyimages_Ready  map[crypto.Key]bool           // keyimages which are ready to get consumed, // we monitor them to find which
	//Outputs_Consumed map[crypto.Key]TX_Wallet_Data // the key is the keyimage

	//Random_Outputs  map[uint64]TX_Wallet_Data  // random ring members
	//Random_Outputs_Recent map[uint64]TX_Wallet_Data  // random ring members from recent blocks
	//Ring_Members    map[uint64]bool // ring members

	sync.Mutex // syncronise modifications to this structure
}

// this structure is kept by wallet
type TX_Wallet_Data struct {
	TXdata globals.TX_Output_Data `msgpack:"txdata"` // all the fields of output data

	WAmount      uint64       `msgpack:"wamount"` // actual amount, in case of miner it is verbatim, for other cases it decrypted
	WKey         ringct.CtKey `msgpack:"wkey"`    // key which is used to later send this specific output
	WKimage      crypto.Key   `msgpack:"wkimage"` // key image which gets consumed when this output is spent
	WSpent       bool         `msgpack:"wspent"`  // whether this output has been spent
	WSpentPool   bool         //`msgpack:""`// we built and send out a tx , but it has not been mined
	WPaymentID   []byte       `msgpack:"wpaymentid"`   // payment if if present and decrypted if required
	WSecretTXkey crypto.Key   `msgpack:"wsecrettxkey"` // tx secret which can be be used to prove that the funds have been spent
}

// generate keys from using random numbers
func Generate_Keys_From_Random() (user *Account, err error) {
	user = &Account{Mixin: 5, FeesMultiplier: 1.5}
	seed := crypto.RandomScalar()
	user.Keys = Generate_Keys_From_Seed(*seed)

	return
}

// generate keys from seed which is from the recovery words
// or we feed in direct
func Generate_Keys_From_Seed(Seed crypto.Key) (keys _Keys) {

	// setup main keys
	keys.Spendkey_Secret = Seed
	keys.Spendkey_Public = *(Seed.PublicKey())

	// view keys are generated from secret ( so as single recovery seed is enough )
	hash := crypto.Key(crypto.Keccak256(Seed[:]))
	crypto.ScReduce32(&hash)
	keys.Viewkey_Secret = hash
	keys.Viewkey_Public = *(keys.Viewkey_Secret.PublicKey())

	return
}

// generate user account using recovery seeds
func Generate_Account_From_Recovery_Words(words string) (user *Account, err error) {
	user = &Account{Mixin: 5, FeesMultiplier: 1.5}
	language, seed, err := mnemonics.Words_To_Key(words)
	if err != nil {
		return
	}

	user.SeedLanguage = language
	user.Keys = Generate_Keys_From_Seed(seed)

	return
}

func Generate_Account_From_Seed(Seed crypto.Key) (user *Account, err error) {
	user = &Account{Mixin: 5, FeesMultiplier: 1.5}

	// TODO check whether the seed is invalid
	user.Keys = Generate_Keys_From_Seed(Seed)

	return
}

// generate keys for view only wallet
func Generate_Account_View_Only(Publicspend crypto.Key, ViewSecret crypto.Key) (user *Account, err error) {

	user = &Account{Mixin: 5, FeesMultiplier: 1.5}

	//  TODO check whether seed is valid secret
	user.Keys.Spendkey_Public = Publicspend
	user.Keys.Viewkey_Secret = ViewSecret
	user.Keys.Viewkey_Public = *(ViewSecret.PublicKey())
	user.ViewOnly = true

	return
}

// generate keys for view only wallet
func Generate_Account_NONDeterministic_Only(Secretspend crypto.Key, ViewSecret crypto.Key) (user *Account, err error) {

	user = &Account{Mixin: 5, FeesMultiplier: 1.5}

	//  TODO check whether seed is valid secret
	user.Keys.Spendkey_Secret = Secretspend
	user.Keys.Spendkey_Public = *(Secretspend.PublicKey())
	user.Keys.Viewkey_Secret = ViewSecret
	user.Keys.Viewkey_Public = *(ViewSecret.PublicKey())
	user.ViewOnly = true

	return
}

// convert key to seed using language
func (w *Wallet) GetSeed() (str string) {
	return mnemonics.Key_To_Words(w.account.Keys.Spendkey_Secret, w.account.SeedLanguage)
}

// convert key to seed using language
func (w *Wallet) GetSeedinLanguage(lang string) (str string) {
	return mnemonics.Key_To_Words(w.account.Keys.Spendkey_Secret, lang)
}

// view wallet key consists of public spendkey and private view key
func (w *Wallet) GetViewWalletKey() (str string) {
	return fmt.Sprintf("%s%s", w.account.Keys.Spendkey_Public, w.account.Keys.Viewkey_Secret)
}

func (account *Account) GetAddress() (addr address.Address) {
	switch globals.Config.Name {
	case "testnet":
		addr.Network = config.Testnet.Public_Address_Prefix //choose dETo

	default:
		fallthrough // assume mainnet
	case "mainnet":
		addr.Network = config.Mainnet.Public_Address_Prefix //choose dERo

		//panic(fmt.Sprintf("Unknown Network \"%s\"", globals.Config.Name))
	}

	addr.SpendKey = account.Keys.Spendkey_Public
	addr.ViewKey = account.Keys.Viewkey_Public

	return
}

// convert a user account to address
func (w *Wallet) GetAddress() (addr address.Address) {
	return w.account.GetAddress()
}

// get a random integrated address
func (w *Wallet) GetRandomIAddress8() (addr address.Address) {
	addr = w.account.GetAddress()

	if addr.Network == config.Mainnet.Public_Address_Prefix {
		addr.Network = config.Mainnet.Public_Address_Prefix_Integrated
	} else { // it's a testnet address
		addr.Network = config.Testnet.Public_Address_Prefix_Integrated
	}

	// setup random 8 bytes of payment ID, it must be from non-deterministic RNG namely crypto random
	addr.PaymentID = make([]byte, 8, 8)
	rand.Read(addr.PaymentID[:])

	return
}

// get a random integrated address
func (w *Wallet) GetRandomIAddress32() (addr address.Address) {
	addr = w.account.GetAddress()

	if addr.Network == config.Mainnet.Public_Address_Prefix {
		addr.Network = config.Mainnet.Public_Address_Prefix_Integrated
	} else { // it's a testnet address
		addr.Network = config.Testnet.Public_Address_Prefix_Integrated
	}

	// setup random 32 bytes of payment ID, it must be from non-deterministic RNG namely crypto random
	addr.PaymentID = make([]byte, 32, 32)
	rand.Read(addr.PaymentID[:])

	return
}

// this function is used to encrypt/decrypt payment id
// as the operation is symmetric XOR, is the same in both direction
func EncryptDecryptPaymentID(derivation crypto.Key, tx_public crypto.Key, input []byte) (output []byte) {
	// input must be exactly 8 bytes long
	if len(input) != 8 {
		panic("Encrypted payment ID must be exactly 8 bytes long")
	}

	var tmp_buf [33]byte
	copy(tmp_buf[:], derivation[:]) // copy derivation key to buffer
	tmp_buf[32] = ENCRYPTED_PAYMENT_ID_TAIL

	// take hash
	hash := crypto.Keccak256(tmp_buf[:]) // take hash of entire 33 bytes, 32 bytes derivation key, 1 byte tail

	output = make([]byte, 8, 8)
	for i := range input {
		output[i] = input[i] ^ hash[i] // xor the bytes with the hash
	}

	return
}

// one simple function which does all the crypto to find out whether output belongs to this account
// NOTE: this function only uses view key secret and Spendkey_Public
// output index is the position of vout within the tx list itself
func (w *Wallet) Is_Output_Ours(tx_public crypto.Key, output_index uint64, vout_key crypto.Key) bool {
	derivation := crypto.KeyDerivation(&tx_public, &w.account.Keys.Viewkey_Secret)
	derivation_public_key := derivation.KeyDerivation_To_PublicKey(output_index, w.account.Keys.Spendkey_Public)

	return derivation_public_key == vout_key
}

// only for testing purposes
func (acc *Account) Is_Output_Ours(tx_public crypto.Key, output_index uint64, vout_key crypto.Key) bool {
	derivation := crypto.KeyDerivation(&tx_public, &acc.Keys.Viewkey_Secret)
	derivation_public_key := derivation.KeyDerivation_To_PublicKey(output_index, acc.Keys.Spendkey_Public)
	return derivation_public_key == vout_key
}

// this function does all the keyderivation required for decrypting ringct outputs, generate keyimage etc
// also used when we build up a transaction for mining or sending amount
func (w *Wallet) Generate_Helper_Key_Image(tx_public crypto.Key, output_index uint64) (ephermal_secret, ephermal_public, keyimage crypto.Key) {
	derivation := crypto.KeyDerivation(&tx_public, &w.account.Keys.Viewkey_Secret)
	ephermal_secret = derivation.KeyDerivation_To_PrivateKey(output_index, w.account.Keys.Spendkey_Secret)
	ephermal_public = derivation.KeyDerivation_To_PublicKey(output_index, w.account.Keys.Spendkey_Public)
	keyimage = crypto.GenerateKeyImage(ephermal_public, ephermal_secret)

	return
}

// this function decodes ringCT encoded output amounts
// this is only possible if signature is full or simple
func (w *Wallet) Decode_RingCT_Output(tx_public crypto.Key, output_index uint64, pkkey crypto.Key, tuple ringct.ECdhTuple, sigtype uint64) (amount uint64, mask crypto.Key, result bool) {

	derivation := crypto.KeyDerivation(&tx_public, &w.account.Keys.Viewkey_Secret)
	scalar_key := derivation.KeyDerivationToScalar(output_index)

	switch sigtype {
	case 0: // NOT possible , miner tx outputs are not hidden
		return
	case 1: // ringct MG  // Both ringct outputs can be decoded using the same methods
		// however, original implementation has different methods, maybe need to evaluate more
		fallthrough
	case 2, 4: // ringct sample

		amount, mask, result = ringct.Decode_Amount(tuple, *scalar_key, pkkey)

	default:
		return
	}

	return
}

// add the transaction to our wallet record, so as funds can be later on tracked
// due to enhanced features, we have to wait and watch for all funds
// this will extract secret keys from newly arrived funds to consume them later on
func (w *Wallet) Add_Transaction_Record_Funds(txdata *globals.TX_Output_Data) (amount uint64, result bool) {

	// check for input maturity at every height change
	w.Lock()
	if w.account.Height < txdata.Height {
		w.account.Height = txdata.Height
		w.account.TopoHeight = txdata.TopoHeight
		w.account.balance_stale = true // balance needs recalculation
	}
	w.Unlock()

	w.Store_Height_Mapping(txdata)     // update height to block mapping, also sets wallet height
	w.Add_Possible_Ring_Member(txdata) // add as ringmember for future

	// if our funds have been consumed, remove it from our available list
	for i := range txdata.Key_Images {
		w.Is_Our_Fund_Consumed(txdata.Key_Images[i], txdata)
	}

	// confirm that that data belongs to this user
	if !w.Is_Output_Ours(txdata.Tx_Public_Key, txdata.Index_within_tx, crypto.Key(txdata.InKey.Destination)) {
		return // output is not ours
	}

	// since input is ours, take a lock for processing
	defer w.Save_Wallet()
	w.Lock()
	defer w.Unlock()

	var tx_wallet TX_Wallet_Data

	/*
		// check whether we are deduplicating, is the transaction already in our records, skip it
		// transaction is already in our wallet, skip it for being duplicate
		if w.check_key_exists(BLOCKCHAIN_UNIVERSE, []byte(FUNDS_BUCKET), itob(txdata.Index_Global)) {
			return
		}
	*/

	// setup Amount
	switch txdata.SigType {
	case 0: // miner tx
		tx_wallet.WAmount = txdata.Amount
		tx_wallet.WKey.Mask = ringct.Identity // secret mask for miner tx is Identity

	case 1, 2, 4: // ringct full/simple, simplebulletproof
		tx_wallet.WAmount, tx_wallet.WKey.Mask, result = w.Decode_RingCT_Output(txdata.Tx_Public_Key, txdata.Index_within_tx, crypto.Key(txdata.InKey.Mask), txdata.ECDHTuple,
			txdata.SigType)

		if result == false { // It's an internal error most probably, log more details
			return
		}
	case 3: //fullbulletproof are not supported

	}

	amount = tx_wallet.WAmount // setup amount so it can be properly returned
	tx_wallet.TXdata = *txdata

	// use payment ID if availble
	// due to design we never know for which output the payment id was
	// C daemon places relates it with any outputs
	// so we attach the payment with all outputs
	// tracking it would make wallet slow
	// we cannot send to ourselves with a payment ID
	// if the TX was sent by this wallet, do NOT process payment IDs,to avoid triggerring a critical bug
	// we need a better FIX for this
	// NOTE: this fix needs reqork before adding support for SHADOW addresses
	if w.GetTXKey(txdata.TXID) == "" { // make sure TX was not sent by this wallet
		if !w.check_key_exists(BLOCKCHAIN_UNIVERSE, []byte(OUR_TX_BUCKET), txdata.TXID[:]) { // if wallet is recreated, we track our TX via key-images
			if txdata.Index_within_tx >= 0 {
				switch len(txdata.PaymentID) {
				case 8: // this needs to be decoded
					derivation := crypto.KeyDerivation(&txdata.Tx_Public_Key, &w.account.Keys.Viewkey_Secret)
					tx_wallet.WPaymentID = EncryptDecryptPaymentID(derivation, txdata.Tx_Public_Key, txdata.PaymentID)
				case 32:
					tx_wallet.WPaymentID = txdata.PaymentID
				}
			}
		}
	}

	// if wallet is viewonly, we cannot track when the funds were spent
	// so lets skip the part, since we do not have the keys
	if !w.account.ViewOnly { // it's a full wallet, track spendable and get ready to spend
		secret_key, _, kimage := w.Generate_Helper_Key_Image(txdata.Tx_Public_Key, txdata.Index_within_tx)
		tx_wallet.WKimage = kimage
		tx_wallet.WKey.Destination = secret_key

		if w.check_key_exists(BLOCKCHAIN_UNIVERSE, []byte(KEYIMAGE_BUCKET), kimage[:]) {
			// find the output index to which this key image belong
			value_bytes, err := w.load_key_value(BLOCKCHAIN_UNIVERSE, []byte(KEYIMAGE_BUCKET), kimage[:])
			if err == nil && len(value_bytes) == 8 {
				index := binary.BigEndian.Uint64(value_bytes)

				// now lets load the suitable index data
				value_bytes, err = w.load_key_value(BLOCKCHAIN_UNIVERSE, []byte(FUNDS_BUCKET), itob(index))
				if err == nil {
					var tx_wallet_temp TX_Wallet_Data
					err = msgpack.Unmarshal(value_bytes, &tx_wallet_temp)
					if err == nil {
						if tx_wallet_temp.TXdata.TXID != txdata.TXID { // transaction mismatch
							rlog.Warnf("KICD  %s,%s,  \n%+v  \n%+v",txdata.TXID, tx_wallet_temp.TXdata.TXID, txdata,tx_wallet_temp);
							return 0, false
						}

						if tx_wallet_temp.TXdata.Index_within_tx != txdata.Index_within_tx { // index within tx mismatch
							rlog.Warnf("KICD2  %s,%s,  \n%+v  \n%+v",txdata.TXID, tx_wallet_temp.TXdata.TXID, txdata,tx_wallet_temp);
							return 0, false
						}
					}
				}
			}

		}

		// store the key image so as later on we can find when it is spent
		w.store_key_value(BLOCKCHAIN_UNIVERSE, []byte(KEYIMAGE_BUCKET), kimage[:], itob(txdata.Index_Global))
	}

	// serialize and store the tx, make it available for funds
	serialized, err := msgpack.Marshal(&tx_wallet)
	if err != nil {
		panic(err)
	}
	// store all data about the transfer
	w.store_key_value(BLOCKCHAIN_UNIVERSE, []byte(FUNDS_BUCKET), itob(txdata.Index_Global), serialized)

	// store TX to global output index link
	w.store_key_value(BLOCKCHAIN_UNIVERSE, append([]byte(TXID), txdata.TXID[:]...), itob(txdata.Index_Global), itob(txdata.Index_Global))

	// store payment ID to global output index link, but only if we have payment ID
	if len(tx_wallet.WPaymentID) == 8 || len(tx_wallet.WPaymentID) == 32 {
		w.store_key_value(BLOCKCHAIN_UNIVERSE, append([]byte(PAYID), tx_wallet.WPaymentID[:]...), itob(txdata.Index_Global), itob(txdata.Index_Global))
	}

	// mark the funds as available
	w.store_key_value(BLOCKCHAIN_UNIVERSE, []byte(FUNDS_AVAILABLE), itob(txdata.Index_Global), itob(txdata.Index_Global))

	// TODO we must also sort transactions by payment id, so as they are searchable by payment id
	w.account.balance_stale = true // balance needs recalculation, as new funds have arrived

	result = true
	return
}

// check whether our fund is consumed
// this is done by finding the keyimages floating in blockchain, to what keyimages belong to this account
// if match is found, we have consumed our funds
// NOTE: Funds spent check should always be earlier than TX output check, so as we can handle payment IDs properly

func (w *Wallet) Is_Our_Fund_Consumed(key_image crypto.Key, txdata *globals.TX_Output_Data) (amount uint64, result bool) {

	// if not ours, go back
	if !w.check_key_exists(BLOCKCHAIN_UNIVERSE, []byte(KEYIMAGE_BUCKET), key_image[:]) {
		return 0, false
	}
	defer w.Save_Wallet()
	w.Lock()
	defer w.Unlock()

	spent_index := txdata.Index_Global

	// these are stored in a different bucket so we never collide with incoming funds
	{
		var tx_wallet TX_Wallet_Data
		tx_wallet.TXdata = *txdata

		// yes it is our fund, store relevant info to FUNDS bucket
		serialized, err := msgpack.Marshal(&tx_wallet)
		if err != nil {
			panic(err)
		}
		// store all data
		w.store_key_value(BLOCKCHAIN_UNIVERSE, []byte(FUNDS_BUCKET_OUTGOING), itob(txdata.Index_Global), serialized)
	}

	// mark that this TX was sent by us
	w.store_key_value(BLOCKCHAIN_UNIVERSE, []byte(OUR_TX_BUCKET), txdata.TXID[:], txdata.TXID[:])

	// find the output index to which this key image belong
	value_bytes, err := w.load_key_value(BLOCKCHAIN_UNIVERSE, []byte(KEYIMAGE_BUCKET), key_image[:])
	if err != nil {
		panic(fmt.Sprintf("Error while reading spent keyimage data key_image %s, err %s", key_image, err))
	}
	index := binary.BigEndian.Uint64(value_bytes)

	// now lets load the suitable index data
	value_bytes, err = w.load_key_value(BLOCKCHAIN_UNIVERSE, []byte(FUNDS_BUCKET), itob(index))
	if err != nil {
		panic(fmt.Sprintf("Error while reading spent funds data key_image %s, index %d err %s", key_image, index, err))
	}

	var tx_wallet TX_Wallet_Data
	err = msgpack.Unmarshal(value_bytes, &tx_wallet)
	if err != nil {
		panic(fmt.Sprintf("Error while decoding spent funds data key_image %s, index %d err %s", key_image, index, err))
	}

	// this case should never be possible, until logical or db corruption has already occured
	if key_image != tx_wallet.WKimage {
		fmt.Printf("%+v\n", tx_wallet)
		panic(fmt.Sprintf("Stored key_image %s and loaded key image mismatch  %s index %d err %s", key_image, tx_wallet.WKimage, index, err))
	}

	// move the funds from availble to spent bucket
	w.delete_key(BLOCKCHAIN_UNIVERSE, []byte(FUNDS_AVAILABLE), itob(index))
	w.store_key_value(BLOCKCHAIN_UNIVERSE, []byte(FUNDS_SPENT), itob(index), itob(index))

	w.store_key_value(BLOCKCHAIN_UNIVERSE, []byte(FUNDS_SPENT_WHERE), itob(index), itob(spent_index))

	w.account.balance_stale = true // balance needs recalculation

	return tx_wallet.WAmount, true
}

// add the transaction to record,
// this will mark the funds as consumed on the basis of  keyimages
// locate the transaction  and get the amount , this is O(n), so we can tell how much funds were spent
// cryptnote only allows to spend complete funds, change comes back
func (w *Wallet) Consume_Transaction_Record_Funds(txdata *globals.TX_Output_Data, key_image crypto.Key) bool {
	return false

}

// get the unlocked balance ( amounts which are mature and can be spent at this time )
// offline wallets may get this wrong, since they may not have latest data
// TODO: for offline wallets, we must make all balance as mature
// full resync costly
// TODO URGENT we are still not cleaning up, spent funds,do that asap to recover funds which were spent on alt-xhain
func (w *Wallet) Get_Balance_Rescan() (mature_balance uint64, locked_balance uint64) {
	w.RLock()
	defer w.RUnlock()

	index_list := w.load_all_values_from_bucket(BLOCKCHAIN_UNIVERSE, []byte(FUNDS_AVAILABLE))

	//fmt.Printf("found %d elements in bucket \n", len(index_list))
	for i := range index_list { // load index
		index := binary.BigEndian.Uint64(index_list[i])
		value_bytes, err := w.load_key_value(BLOCKCHAIN_UNIVERSE, []byte(FUNDS_BUCKET), index_list[i])
		if err != nil {
			rlog.Debugf("Error while reading available funds index index %d err %s", index, err)
			continue
		}

		var tx_wallet TX_Wallet_Data
		err = msgpack.Unmarshal(value_bytes, &tx_wallet)
		if err != nil {
			rlog.Debugf("Error while decoding availble funds data index %d err %s", index, err)
			continue
		}

		// check whether the height and block matches with what is stored at this point in time
		local_hash, err := w.load_key_value(BLOCKCHAIN_UNIVERSE, []byte(HEIGHT_TO_BLOCK_BUCKET), itob(uint64(tx_wallet.TXdata.TopoHeight)))
		if err != nil {
			continue
		}

		// input disappeared due to chain soft fork
		if len(local_hash) == 32 && !bytes.Equal(tx_wallet.TXdata.BLID[:], local_hash) {
			// stop tracking the funds everywhere
			w.delete_key(BLOCKCHAIN_UNIVERSE, []byte(FUNDS_BUCKET), index_list[i])           // delete index
			w.delete_key(BLOCKCHAIN_UNIVERSE, []byte(KEYIMAGE_BUCKET), tx_wallet.WKimage[:]) // delete key_image for this
			w.delete_key(BLOCKCHAIN_UNIVERSE, []byte(FUNDS_AVAILABLE), itob(index))
			w.delete_key(BLOCKCHAIN_UNIVERSE, []byte(FUNDS_SPENT), index_list[i])

			w.delete_key(BLOCKCHAIN_UNIVERSE, []byte(FUNDS_SPENT), index_list[i])

			// delete TXID to index mapping
			w.delete_key(BLOCKCHAIN_UNIVERSE, append([]byte(TXID), tx_wallet.TXdata.TXID[:]...), index_list[i])
			// delete payment ID to index mapping
			w.delete_key(BLOCKCHAIN_UNIVERSE, append([]byte(PAYID), tx_wallet.WPaymentID[:]...), index_list[i])

			continue // skip this input
		}

		if inputmaturity.Is_Input_Mature(w.account.Height,
			tx_wallet.TXdata.Height,
			tx_wallet.TXdata.Unlock_Height,
			tx_wallet.TXdata.SigType) {
			mature_balance += tx_wallet.WAmount
		} else {
			locked_balance += tx_wallet.WAmount
		}

	}

	w.account.Balance_Mature = mature_balance
	w.account.Balance_Locked = locked_balance
	w.account.balance_stale = false // balance is updated

	return
}

// get the unlocked balance ( amounts which are mature and can be spent at this time )
// offline wallets may get this wrong, since they may not have latest data

//
func (w *Wallet) Get_Balance() (mature_balance uint64, locked_balance uint64) {
	w.RLock()
	if !w.Is_Balance_Modified() {
		w.RUnlock()
		return w.account.Balance_Mature, w.account.Balance_Locked
	}

	w.RUnlock()
	return w.Get_Balance_Rescan() // call heavy function
}

var old_block_cache crypto.Hash // used as a cache to skip decryptions is possible

func (w *Wallet) Store_Height_Mapping(txdata *globals.TX_Output_Data) {

	w.Lock()
	defer w.Unlock()

	// store height to block hash mapping
	// store all data
	// save block height only if required

	w.account.Height = txdata.Height         // increase wallet height
	w.account.TopoHeight = txdata.TopoHeight // increase wallet topo height
	if old_block_cache != txdata.BLID {      // make wallet scanning as fast as possible
		//fmt.Printf("gStoring height to block %d %s t %s\n", txdata.Height, txdata.BLID, txdata.TXID)
		old_block, err := w.load_key_value(BLOCKCHAIN_UNIVERSE, []byte(HEIGHT_TO_BLOCK_BUCKET), itob(uint64(txdata.TopoHeight)))
		if err != nil || !bytes.Equal(old_block, txdata.BLID[:]) {
			w.store_key_value(BLOCKCHAIN_UNIVERSE, []byte(HEIGHT_TO_BLOCK_BUCKET), itob(uint64(txdata.TopoHeight)), txdata.BLID[:])
			old_block_cache = txdata.BLID

			// fmt.Printf("Storing height to block %d %s\n", txdata.Height, txdata.BLID)
		}
	}

}

// add all random outputs which will be used while creating transactions
// currently we store all ringmembers
func (w *Wallet) Add_Possible_Ring_Member(txdata *globals.TX_Output_Data) {

	if txdata == nil {
		return
	}

	if w.Is_View_Only() { // view only wallets can never construct a transaction
		return
	}
	w.Lock()
	defer w.Unlock()

	w.account.Index_Global = txdata.Index_Global // increment out pointer

	var r Ring_Member
	r.InKey = txdata.InKey
	r.Height = txdata.Height
	r.Index_Global = txdata.Index_Global
	r.Unlock_Height = txdata.Unlock_Height // required for maturity checking
	r.Sigtype = txdata.SigType             // required for maturity checking

	// lets serialize and store the data encrypted, so as no one track any info
	serialized, err := msgpack.Marshal(&r)
	if err != nil {
		panic(err)
	}
	// store all data
	w.store_key_value(BLOCKCHAIN_UNIVERSE, []byte(RING_BUCKET), itob(txdata.Index_Global), serialized)

	return
}

type Entry struct {
	Index_Global uint64 `json:"index_global"`
	Height       uint64 `json:"height"` 
	TopoHeight   int64  `json:"topoheight"` 
	TXID         crypto.Hash `json:"txid"` 
	Amount       uint64  `json:"amount"`
	PaymentID    []byte `json:"payment_id"`
	Status       byte  `json:"status"`
 	Unlock_Time  uint64 `json:"unlock_time"`
	Time         time.Time `json:"time"`
	Secret_TX_Key string `json:"secret_tx_key"`  // can be used to prove if available
	Details structures.Outgoing_Transfer_Details `json:"details"`  // actual details if available
}

	
// finds all inputs which have been received/spent etc
// TODO this code can be easily parallelised and need to be parallelised
// if only the availble is requested, then the wallet is very fast
// the spent tracking may make it slow ( in case of large probably million  txs )
//TODO currently we do not track POOL at all any where ( except while building tx)
// if payment_id is true, only entries with payment ids are returned
func (w *Wallet) Show_Transfers(available bool, in bool, out bool, pool bool, failed bool, payment_id bool, min_height, max_height uint64) (entries []Entry) {

	dero_first_block_time := time.Unix(1512432000, 0) //Tuesday, December 5, 2017 12:00:00 AM

	if max_height == 0 {
		max_height = 5000000000
	}
	if available || in {
		index_list := w.load_all_values_from_bucket(BLOCKCHAIN_UNIVERSE, []byte(FUNDS_AVAILABLE))

		for i := range index_list { // load index
			current_index := binary.BigEndian.Uint64(index_list[i])

			tx, err := w.load_funds_data(current_index, FUNDS_BUCKET)
			if err != nil {
				rlog.Warnf("Error while reading available funds index index %d err %s", current_index, err)
				continue
			}

			if tx.TXdata.Height >= min_height && tx.TXdata.Height <= max_height { // height filter
				var entry Entry
				entry.Index_Global = current_index
				entry.Height = tx.TXdata.Height
				entry.TopoHeight = tx.TXdata.TopoHeight
				entry.TXID = tx.TXdata.TXID
				entry.Amount = tx.WAmount
				entry.PaymentID = tx.WPaymentID
				entry.Status = 0
				entry.Time = time.Unix(int64(tx.TXdata.Block_Time), 0)

				if entry.Height < 95600 { // make up time for pre-atlantis blocks
					duration, _ := time.ParseDuration(fmt.Sprintf("%ds", int64(180*entry.Height)))
					entry.Time = dero_first_block_time.Add(duration)
				}

				if payment_id {

					if len(entry.PaymentID) >= 8 {
						entries = append(entries, entry)
					}
				} else {
					entries = append(entries, entry)
				}

			}

		}
	}

	if in || out {
		// all spent funds will have 2 entries, one for receive/other for spent
		index_list := w.load_all_values_from_bucket(BLOCKCHAIN_UNIVERSE, []byte(FUNDS_SPENT))

		for i := range index_list { // load index
			current_index := binary.BigEndian.Uint64(index_list[i])

			tx, err := w.load_funds_data(current_index, FUNDS_BUCKET)
			if err != nil {
				rlog.Warnf("Error while reading available funds index index %d err %s", current_index, err)
				continue
			}

			/// receipt entry
			var entry Entry
			entry.Index_Global = current_index
			entry.Height = tx.TXdata.Height
			entry.TopoHeight = tx.TXdata.TopoHeight
			entry.TXID = tx.TXdata.TXID
			entry.Amount = tx.WAmount
			entry.PaymentID = tx.WPaymentID
			entry.Status = 0
			entry.Time = time.Unix(int64(tx.TXdata.Block_Time), 0)

			if entry.Height < 95600 { // make up time for pre-atlantis blocks
				duration, _ := time.ParseDuration(fmt.Sprintf("%ds", int64(180*entry.Height)))
				entry.Time = dero_first_block_time.Add(duration)
			}

			if in {
				if tx.TXdata.Height >= min_height && tx.TXdata.Height <= max_height { // height filter
					if payment_id {

						if len(entry.PaymentID) >= 8 {
							entries = append(entries, entry)
						}
					} else {
						entries = append(entries, entry)
					}
				}

			}

			if out {
				// spendy entry
				value_bytes, err := w.load_key_value(BLOCKCHAIN_UNIVERSE, []byte(FUNDS_SPENT_WHERE), index_list[i])
				if err != nil {
					fmt.Printf("Error while reading FUNDS_SPENT_WHERE index %d err %s", current_index, err)
					continue
				}
				spent_index := binary.BigEndian.Uint64(value_bytes)
				tx, err = w.load_funds_data(spent_index, FUNDS_BUCKET_OUTGOING)
				if err != nil {
					rlog.Warnf("Error while reading available funds index index %d err %s", current_index, err)
					continue
				}
				if tx.TXdata.Height >= min_height && tx.TXdata.Height <= max_height { // height filter
					entry.Index_Global = tx.TXdata.Index_Global
					entry.Height = tx.TXdata.Height
					entry.TXID = tx.TXdata.TXID
					entry.TopoHeight = tx.TXdata.TopoHeight
					entry.PaymentID = entry.PaymentID[:0] // payment id needs to be zero or tracked from some where else
					entry.Time = time.Unix(int64(tx.TXdata.Block_Time), 0)

					if entry.Height < 95600 { // make up time for pre-atlantis blocks
						duration, _ := time.ParseDuration(fmt.Sprintf("%ds", int64(180*entry.Height)))
						entry.Time = dero_first_block_time.Add(duration)
					}
					
					// fill tx secret_key 
					entry.Secret_TX_Key =  w.GetTXKey(tx.TXdata.TXID)
                                        entry.Details = w.GetTXOutDetails(tx.TXdata.TXID)

					entry.Status = 1
					entries = append(entries, entry) // spend entry
				}
			}

		}
	}

	sort.SliceStable(entries, func(i, j int) bool { return entries[i].Index_Global > entries[j].Index_Global })

	return

}

// w.delete_key(BLOCKCHAIN_UNIVERSE, append([]byte(TXID),tx_wallet.TXdata.TXID[:]...) , index_list[i])

/* gets all the payments  done to specific payment ID and filtered by specific block height */
func (w *Wallet) Get_Payments_Payment_ID(payid []byte, min_height uint64) (entries []Entry) {
	index_list := w.load_all_values_from_bucket(BLOCKCHAIN_UNIVERSE, append([]byte(PAYID), payid[:]...))

	for i := range index_list {
		current_index := binary.BigEndian.Uint64(index_list[i])

		tx, err := w.load_funds_data(current_index, FUNDS_BUCKET)
		if err != nil {
			rlog.Warnf("Error while reading available funds index index %d err %s", current_index, err)
			continue
		}

		if tx.TXdata.Height > min_height { // height filter
			var entry Entry
			entry.Index_Global = current_index
			entry.TopoHeight = tx.TXdata.TopoHeight
			entry.Height = tx.TXdata.Height
			entry.TXID = tx.TXdata.TXID
			entry.Amount = tx.WAmount
			entry.PaymentID = tx.WPaymentID
			entry.Status = 0
			entry.Unlock_Time = tx.TXdata.Unlock_Height
			
			// fill tx secret_key 
                        entry.Secret_TX_Key =  w.GetTXKey(tx.TXdata.TXID)
                        entry.Details = w.GetTXOutDetails(tx.TXdata.TXID)
			entries = append(entries, entry)
		}
	}

	sort.SliceStable(entries, func(i, j int) bool { return entries[i].Index_Global > entries[j].Index_Global })

	return

}

// return all payments within a tx there can be more than 1 entry, if yes then they will be merged
// NOTE:
func (w *Wallet) Get_Payments_TXID(txid []byte) (entry Entry) {
	index_list := w.load_all_values_from_bucket(BLOCKCHAIN_UNIVERSE, append([]byte(TXID), txid[:]...))

	for i := range index_list {
		current_index := binary.BigEndian.Uint64(index_list[i])

		tx, err := w.load_funds_data(current_index, FUNDS_BUCKET)
		if err != nil {
			rlog.Warnf("Error while reading available funds index index %d err %s", current_index, err)
			continue
		}
		if bytes.Compare(txid, tx.TXdata.TXID[:]) == 0 {

			entry.Index_Global = current_index
			entry.Height = tx.TXdata.Height
			entry.TopoHeight = tx.TXdata.TopoHeight
			entry.TXID = tx.TXdata.TXID
			entry.Amount += tx.WAmount // merge all amounts ( if it was provided in different outputs)
			entry.PaymentID = tx.WPaymentID
			entry.Status = 0
			entry.Unlock_Time = tx.TXdata.Unlock_Height
			
			// fill tx secret_key 
                        entry.Secret_TX_Key =  w.GetTXKey(tx.TXdata.TXID)
                        entry.Details = w.GetTXOutDetails(tx.TXdata.TXID)
		}
	}

	return

}

// get the unlocked balance ( amounts which are mature and can be spent at this time )
// offline wallets may get this wrong, since they may not have latest data
// TODO: for offline wallets, we must make all balance as mature
//
func (w *Wallet) Start_RPC_Server(address string) (err error) {
	w.Lock()
	defer w.Unlock()

	tcpAddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return
	}

	w.rpcserver, err = RPCServer_Start(w, tcpAddr.String())
	if err != nil {
		w.rpcserver = nil
	}

	return
}

func (w *Wallet) Stop_RPC_Server() {
	w.Lock()
	defer w.Unlock()

	if w.rpcserver != nil {
		w.rpcserver.RPCServer_Stop()
		w.rpcserver = nil // remover reference
	}

	return
}

// delete most of the data and prepare for rescan
func (w *Wallet) Clean() {
	w.Lock()
	defer w.Get_Balance_Rescan()
	defer w.Unlock()

	w.delete_bucket(BLOCKCHAIN_UNIVERSE, []byte(FUNDS_BUCKET))
	w.delete_bucket(BLOCKCHAIN_UNIVERSE, []byte(FUNDS_AVAILABLE))
	w.delete_bucket(BLOCKCHAIN_UNIVERSE, []byte(FUNDS_SPENT))
	w.delete_bucket(BLOCKCHAIN_UNIVERSE, []byte(FUNDS_SPENT_WHERE))
	w.delete_bucket(BLOCKCHAIN_UNIVERSE, []byte(FUNDS_BUCKET_OUTGOING))

	w.delete_bucket(BLOCKCHAIN_UNIVERSE, []byte(RING_BUCKET)) //Improves wallet rescan performance.
	w.delete_bucket(BLOCKCHAIN_UNIVERSE, []byte(KEYIMAGE_BUCKET))
	w.delete_bucket(BLOCKCHAIN_UNIVERSE, []byte(HEIGHT_TO_BLOCK_BUCKET))

}

// whether account is view only
func (w *Wallet) Is_View_Only() bool {
	// w.RLock()
	//defer w.RUnlock()
	return w.account.ViewOnly
}

// informs thats balance information may be stale and needs recalculation
func (w *Wallet) Is_Balance_Modified() bool {
	// w.RLock()
	// defer w.RUnlock()
	return w.account.balance_stale
}

// return height of wallet
func (w *Wallet) Get_Height() uint64 {
	w.RLock()
	defer w.RUnlock()
	return w.account.Height
}

// return topoheight of wallet
func (w *Wallet) Get_TopoHeight() int64 {
	w.RLock()
	defer w.RUnlock()
	return w.account.TopoHeight
}

func (w *Wallet) Get_Daemon_Height() uint64 {
	w.Lock()
	defer w.Unlock()

	return w.Daemon_Height
}

// return index position
func (w *Wallet) Get_Index_Global() uint64 {
	w.RLock()
	defer w.RUnlock()
	return w.account.Index_Global
}

func (w *Wallet) Get_Keys() _Keys {
	return w.account.Keys
}

// by default a wallet opens in Offline Mode
// however, if the wallet is in online mode, it can be made offline instantly using this
func (w *Wallet) SetOfflineMode() bool {
	w.Lock()
	defer w.Unlock()

	current_mode := w.wallet_online_mode
	w.wallet_online_mode = false
	return current_mode
}

// return current mode
func (w *Wallet) GetMode() bool {
	w.RLock()
	defer w.RUnlock()

	return w.wallet_online_mode
}

// use the endpoint set  by the program
func (w *Wallet) SetDaemonAddress(endpoint string) string {
	w.Lock()
	defer w.Unlock()

	w.Daemon_Endpoint = endpoint
	return w.Daemon_Endpoint
}

// by default a wallet opens in Offline Mode
// however, It can be made online by calling this
func (w *Wallet) SetOnlineMode() bool {
	w.Lock()
	defer w.Unlock()

	current_mode := w.wallet_online_mode
	w.wallet_online_mode = true

	if current_mode != true { // trigger subroutine if previous mode was offline
		go w.sync_loop() // start sync subroutine
	}
	return current_mode
}

// by default a wallet opens in Offline Mode
// however, It can be made online by calling this
func (w *Wallet) SetMixin(Mixin int) int {
	defer w.Save_Wallet() // save wallet
	w.Lock()
	defer w.Unlock()

	if Mixin >= 5 && Mixin < 14 { //reasonable limits for mixin, atleastt for now, network should bump it to 13 on next HF
		w.account.Mixin = Mixin
	}
	return w.account.Mixin
}

// by default a wallet opens in Offline Mode
// however, It can be made online by calling this
func (w *Wallet) GetMixin() int {
	w.Lock()
	defer w.Unlock()
	if w.account.Mixin < 5 {
		return 5
	}
	return w.account.Mixin
}

// sets a fee multiplier
func (w *Wallet) SetFeeMultiplier(x float32) float32 {
	defer w.Save_Wallet() // save wallet
	w.Lock()
	defer w.Unlock()
	if x < 1.0 { // fee cannot be less than 1.0, base fees
		w.account.FeesMultiplier = 2.0
	} else {
		w.account.FeesMultiplier = x
	}
	return w.account.FeesMultiplier
}

// gets current fee multiplier
func (w *Wallet) GetFeeMultiplier() float32 {
	w.Lock()
	defer w.Unlock()
	if w.account.FeesMultiplier < 1.0 {
		return 1.0
	}
	return w.account.FeesMultiplier
}

// get fees multiplied by multiplier
func (w *Wallet) getfees(txfee uint64) uint64 {
	multiplier := w.account.FeesMultiplier
	if multiplier < 1.0 {
		multiplier = 2.0
	}
	return txfee * uint64(multiplier*100.0) / 100
}

// Ability to change seed lanaguage
func (w *Wallet) SetSeedLanguage(language string) string {
	defer w.Save_Wallet() // save wallet
	w.Lock()
	defer w.Unlock()

	language_list := mnemonics.Language_List()
	for i := range language_list {
		if strings.ToLower(language) == strings.ToLower(language_list[i]) {
			w.account.SeedLanguage = language_list[i]
		}
	}
	return w.account.SeedLanguage
}

// retrieve current seed language
func (w *Wallet) GetSeedLanguage() string {
	w.Lock()
	defer w.Unlock()
	if w.account.SeedLanguage == "" { // default is English
		return "English"
	}
	return w.account.SeedLanguage
}

// retrieve  secret key for any tx we may have created
func (w *Wallet) GetTXKey(txhash crypto.Hash) string {

	key, err := w.load_key_value(BLOCKCHAIN_UNIVERSE, []byte(SECRET_KEY_BUCKET), txhash[:])
	if err != nil {
		return ""
	}

	return fmt.Sprintf("%x", key)
}

// we need better names for functions
func (w *Wallet) GetTXOutDetails(txhash crypto.Hash) (details structures.Outgoing_Transfer_Details) {

	data_bytes, err := w.load_key_value(BLOCKCHAIN_UNIVERSE, []byte(TX_OUT_DETAILS_BUCKET), txhash[:])
	if err != nil {
		return
	}

	if len(data_bytes) > 10 {
		json.Unmarshal(data_bytes, &details)
	}

	return
}

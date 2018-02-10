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
import "sync"

import "github.com/deroproject/derosuite/config"
import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/crypto/ringct"
import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/walletapi/mnemonics"
import "github.com/deroproject/derosuite/address"
import "github.com/deroproject/derosuite/blockchain/inputmaturity"

type _Keys struct {
	Spendkey_Secret crypto.Key
	Spendkey_Public crypto.Key
	Viewkey_Secret  crypto.Key
	Viewkey_Public  crypto.Key
}

type Account struct {
	Keys         _Keys
	SeedLanguage string

	ViewOnly bool // is this viewonly wallet

	Index_Global uint64 // till where the indexes have been processed
	Height       uint64

	Balance        uint64 // total balance of account
	Balance_Locked uint64 // balance locked

	Outputs_Array []TX_Wallet_Data // all outputs found in the chain belonging to us, as found in chain

	// uint64 si the Index_Global which is the unique number
	Outputs_Index    map[uint64]bool               // all outputs which  are ours for deduplication
	Outputs_Ready    map[uint64]TX_Wallet_Data     // these outputs are ready for consumption ( maturity needs to be checked)
	Keyimages_Ready  map[crypto.Key]bool           // keyimages which are ready to get consumed, // we monitor them to find which
	Outputs_Consumed map[crypto.Key]TX_Wallet_Data // the key is the keyimage

	sync.Mutex // syncronise modifications to this structure
}

// this structure is kept by wallet
type TX_Wallet_Data struct {
	TXdata globals.TX_Output_Data // all the fields of output data

	WAmount uint64       // actual amount, in case of miner it is verbatim, for other cases it decrypted
	WKey    ringct.CtKey // key which is used to later send this specific output
	WKimage crypto.Key   // key image which gets consumed when this output is spent
	WSpent  bool         // whether this output has been spent
}

// generate keys from using random numbers
func Generate_Keys_From_Random() (user *Account, err error) {
	user = &Account{}
	seed := crypto.RandomScalar()
	user.Keys = Generate_Keys_From_Seed(*seed)

	// initialize maps now
	user.Outputs_Index = map[uint64]bool{}
	user.Outputs_Ready = map[uint64]TX_Wallet_Data{}
	user.Outputs_Consumed = map[crypto.Key]TX_Wallet_Data{}
	user.Keyimages_Ready = map[crypto.Key]bool{}
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
	user = &Account{}
	language, seed, err := mnemonics.Words_To_Key(words)
	if err != nil {
		return
	}

	user.SeedLanguage = language
	user.Keys = Generate_Keys_From_Seed(seed)

	// initialize maps now

	user.Outputs_Index = map[uint64]bool{}
	user.Outputs_Ready = map[uint64]TX_Wallet_Data{}
	user.Outputs_Consumed = map[crypto.Key]TX_Wallet_Data{}
	user.Keyimages_Ready = map[crypto.Key]bool{}

	return
}

func Generate_Account_From_Seed(Seed crypto.Key) (user *Account, err error) {
	user = &Account{}

	// TODO check whether the seed is invalid
	user.Keys = Generate_Keys_From_Seed(Seed)

	// initialize maps now
	user.Outputs_Index = map[uint64]bool{}
	user.Outputs_Ready = map[uint64]TX_Wallet_Data{}
	user.Outputs_Consumed = map[crypto.Key]TX_Wallet_Data{}
	user.Keyimages_Ready = map[crypto.Key]bool{}

	return
}

// generate keys for view only wallet
func Generate_Account_View_Only(Publicspend crypto.Key, ViewSecret crypto.Key) (user *Account, err error) {

	user = &Account{}

	//  TODO check whether seed is valid secret
	user.Keys.Spendkey_Public = Publicspend
	user.Keys.Viewkey_Secret = ViewSecret
	user.Keys.Viewkey_Public = *(ViewSecret.PublicKey())
	user.ViewOnly = true

	// initialize maps
	user.Outputs_Index = map[uint64]bool{}
	user.Outputs_Ready = map[uint64]TX_Wallet_Data{}
	user.Outputs_Consumed = map[crypto.Key]TX_Wallet_Data{}
	user.Keyimages_Ready = map[crypto.Key]bool{}

	return
}

// convert key to seed using language
func (user *Account) GetSeed() (str string) {
	return mnemonics.Key_To_Words(user.Keys.Spendkey_Secret, user.SeedLanguage)
}

// view wallet key consists of public spendkey and private view key
func (user *Account) GetViewWalletKey() (str string) {
	return fmt.Sprintf("%s%s", user.Keys.Spendkey_Public, user.Keys.Viewkey_Secret)
}

// convert a user account to address
func (user *Account) GetAddress() (addr address.Address) {
	switch globals.Config.Name {
	case "testnet":
		addr.Network = config.Testnet.Public_Address_Prefix //choose dETo

	default:
		fallthrough // assume mainnet
	case "mainnet":
		addr.Network = config.Mainnet.Public_Address_Prefix //choose dERo

		//panic(fmt.Sprintf("Unknown Network \"%s\"", globals.Config.Name))
	}

	addr.SpendKey = user.Keys.Spendkey_Public
	addr.ViewKey = user.Keys.Viewkey_Public

	return
}

// one simple function which does all the crypto to find out whether output belongs to this account
// NOTE: this function only uses view key secret and Spendkey_Public
// output index is the position of vout within the tx list itself
func (user *Account) Is_Output_Ours(tx_public crypto.Key, output_index uint64, vout_key crypto.Key) bool {
	derivation := crypto.KeyDerivation(&tx_public, &user.Keys.Viewkey_Secret)
	derivation_public_key := derivation.KeyDerivation_To_PublicKey(output_index, user.Keys.Spendkey_Public)

	return derivation_public_key == vout_key
}

// this function does all the keyderivation required for decrypting ringct outputs, generate keyimage etc
// also used when we build up a transaction for mining or sending amount
func (user *Account) Generate_Helper_Key_Image(tx_public crypto.Key, output_index uint64) (ephermal_secret, ephermal_public, keyimage crypto.Key) {
	derivation := crypto.KeyDerivation(&tx_public, &user.Keys.Viewkey_Secret)
	ephermal_public = derivation.KeyDerivation_To_PublicKey(output_index, user.Keys.Spendkey_Public)
	ephermal_secret = derivation.KeyDerivation_To_PrivateKey(output_index, user.Keys.Spendkey_Secret)

	keyimage = crypto.GenerateKeyImage(ephermal_public, ephermal_secret)

	return
}

// this function decodes ringCT encoded output amounts
// this is only possible if signature is full or simple
func (user *Account) Decode_RingCT_Output(tx_public crypto.Key, output_index uint64, pkkey crypto.Key, tuple ringct.ECdhTuple, sigtype uint64) (amount uint64, mask ringct.Key, result bool) {

	derivation := crypto.KeyDerivation(&tx_public, &user.Keys.Viewkey_Secret)
	scalar_key := derivation.KeyDerivationToScalar(output_index)

	switch sigtype {
	case 0: // NOT possible , miner tx outputs are not hidden
		return
	case 1: // ringct MG  // Both ringct outputs can be decoded using the same methods
		// however, original implementation has different methods, maybe need to evaluate more
		fallthrough
	case 2: // ringct sample

		amount, mask, result = ringct.Decode_Amount(tuple, ringct.Key(*scalar_key), ringct.Key(pkkey))

	default:
		return
	}

	return
}

// add the transaction to our wallet record, so as funds can be later on tracked
// due to enhanced features, we have to wait and watch for all funds
// this will extract secret keys from newly arrived funds to consume them later on
func (user *Account) Add_Transaction_Record_Funds(txdata *globals.TX_Output_Data) (result bool) {
	user.Lock()
	defer user.Unlock()

	var tx_wallet TX_Wallet_Data
	// confirm once again that data belongs to this user
	if !user.Is_Output_Ours(txdata.Tx_Public_Key, txdata.Index_within_tx, crypto.Key(txdata.InKey.Destination)) {
		return false // output is not ours
	}

	// setup Amount
	switch txdata.SigType {
	case 0: // miner tx
		tx_wallet.WAmount = txdata.Amount
		tx_wallet.WKey.Mask = ringct.ZeroCommitment_From_Amount(txdata.Amount)

	case 1, 2: // ringct full/simple
		tx_wallet.WAmount, tx_wallet.WKey.Mask, result = user.Decode_RingCT_Output(txdata.Tx_Public_Key, txdata.Index_within_tx, crypto.Key(txdata.InKey.Mask), txdata.ECDHTuple,
			txdata.SigType)

		if result == false { // It's an internal error most probably
			return false
		}
	}

	tx_wallet.TXdata = *txdata

	// check whether we are deduplicating, is the transaction already in our records, skip it
	if _, ok := user.Outputs_Index[txdata.Index_Global]; ok { // transaction is already in our wallet, skip it for being duplicate
		return false
	}

	// if wallet is viewonly, we cannot track when the funds were spent
	// so lets skip the part, since we do not have th keys
	if !user.ViewOnly { // it's a full wallet, track spendable and get ready to spend
		secret_key, _, kimage := user.Generate_Helper_Key_Image(txdata.Tx_Public_Key, txdata.Index_within_tx)
		user.Keyimages_Ready[kimage] = true // monitor this key image for consumption

		tx_wallet.WKimage = kimage
		tx_wallet.WKey.Destination = ringct.Key(secret_key)
	}

	// add tx info to wallet
	user.Outputs_Index[txdata.Index_Global] = true // deduplication it if it ever comes again
	user.Outputs_Ready[txdata.Index_Global] = tx_wallet
	user.Outputs_Array = append(user.Outputs_Array, tx_wallet)

	return true
}

// check whether our fund is consumed
// this is done by finding the keyimages floating in blockchain, to what keyimages belong to this account
//  if  match is found, we have consumed our funds
func (user *Account) Is_Our_Fund_Consumed(key_image crypto.Key) (amount uint64, result bool) {
	if _, ok := user.Keyimages_Ready[key_image]; ok {
		user.Lock()
		defer user.Unlock()

		for k, _ := range user.Outputs_Ready {
			if user.Outputs_Ready[k].WKimage == key_image {
				return user.Outputs_Ready[k].WAmount, true // return ammount and success
			}
		}

		fmt.Printf("This case should NOT be possible theoritically\n")
		return 0, true
	}
	return 0, false
}

// add the transaction to record,
// this will mark the funds as consumed on the basis of  keyimages
// locate the transaction  and get the amount , this is O(n), so we can tell how much funds were spent
// cryptnote only allows to spend complete funds, change comes back
func (user *Account) Consume_Transaction_Record_Funds(txdata *globals.TX_Output_Data, key_image crypto.Key) bool {
	var tx_wallet TX_Wallet_Data
	if _, ok := user.Keyimages_Ready[key_image]; ok {
		user.Lock()
		defer user.Unlock()

		for k, _ := range user.Outputs_Ready {
			if user.Outputs_Ready[k].WKimage == key_image { // find the input corressponding to this image

				// mark output as consumed, move it to consumed map, delete it from ready map
				tx_wallet.TXdata = *txdata
				tx_wallet.WAmount = user.Outputs_Ready[k].WAmount // take amount from original TX
				tx_wallet.WSpent = true                           // mark this fund as spent

				delete(user.Outputs_Ready, k)

				user.Outputs_Consumed[key_image] = tx_wallet
				user.Outputs_Array = append(user.Outputs_Array, user.Outputs_Consumed[key_image])

				return true // return success
			}
		}

		fmt.Printf("This case should NOT be possible theoritically\n")
		// locate the transaction  and get the amount , this is O(n)
		return true
	}

	return false
}

// get the unlocked balance ( amounts which are mature and can be spent at this time )
// offline wallets may get this wrong, since they may not have latest data
// TODO: for offline wallets, we must make all balance as mature

func (user *Account) Get_Balance() (mature_balance uint64, locked_balance uint64) {
	user.Lock()
	defer user.Unlock()

	for k := range user.Outputs_Ready {

		if inputmaturity.Is_Input_Mature(user.Height,
			user.Outputs_Ready[k].TXdata.Height,
			user.Outputs_Ready[k].TXdata.Unlock_Height,
			user.Outputs_Ready[k].TXdata.SigType) {
			mature_balance += user.Outputs_Ready[k].WAmount
		} else {
			locked_balance += user.Outputs_Ready[k].WAmount
		}
	}

	return
}

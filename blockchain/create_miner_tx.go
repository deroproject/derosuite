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

package blockchain

import "fmt"

import "github.com/deroproject/derosuite/config"
import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/address"
import "github.com/deroproject/derosuite/crypto/ringct"
import "github.com/deroproject/derosuite/transaction"

// this function creates a miner tx, with specific blockreward
// TODO we should consider hardfork  version while creating a miner tx
func Create_Miner_TX(hf_version, height, reward uint64, miner_address address.Address, reserve_size int) (tx transaction.Transaction, err error) {

	// initialize extra map in empty tx

	tx.Extra_map = map[transaction.EXTRA_TAG]interface{}{}

	//TODO need to fix hard fork version checks
	tx.Version = 2

	tx.Unlock_Time = height + config.MINER_TX_AMOUNT_UNLOCK       // miner tx unlocks in this much time
	tx.Vin = append(tx.Vin, transaction.Txin_gen{Height: height}) // add input height

	// now lets create encrypted keys, so as the miner_address can spend them

	tx_secret_key, tx_public_key := crypto.NewKeyPair() // create new tx key pair

	//tx_public_key is added to extra and serialized
	tx.Extra_map[transaction.TX_PUBLIC_KEY] = *tx_public_key

	derivation := crypto.KeyDerivation(&miner_address.ViewKey, tx_secret_key) // keyderivation using miner address view key

	index_within_tx := uint64(0)

	// this becomes the key within Vout
	ehphermal_public_key := derivation.KeyDerivation_To_PublicKey(index_within_tx, miner_address.SpendKey)

	// added the amount and key in vout
	tx.Vout = append(tx.Vout, transaction.Tx_out{Amount: reward, Target: transaction.Txout_to_key{Key: ehphermal_public_key}})

	// add reserve size if requested
	if reserve_size != 0 {
		if reserve_size <= 255 {
			tx.Extra_map[transaction.TX_EXTRA_NONCE] = make([]byte, reserve_size, reserve_size)
		} else {
			// give a warning that nonce was requested but could not be created
		}
	}

	tx.Extra = tx.Serialize_Extra() // serialize the extra
	// add 0 byte ringct signature
	var sig ringct.RctSig
	tx.RctSignature = &sig

	return
}

// this function creates a miner tx, with specific blockreward
// TODO we should consider hardfork  version while creating a miner tx
// even if the hash changes during hard-forks or between , the TX should still be spendable by the owner
// the tx has might change due to serialization changes/ differences, however keyimage will still be the same and
// thus not double spendable
func Create_Miner_TX2(hf_version, height int64, miner_address address.Address) (tx transaction.Transaction, err error) {

	// initialize extra map in empty tx

	tx.Extra_map = map[transaction.EXTRA_TAG]interface{}{}

	//TODO need to fix hard fork version checks
	switch {
	case hf_version <= 6:
		tx.Version = 2
	default:
		err = fmt.Errorf("NO such hardfork version")
	}

	tx.Unlock_Time = uint64(height) + config.MINER_TX_AMOUNT_UNLOCK       // miner tx unlocks in this much time
	tx.Vin = append(tx.Vin, transaction.Txin_gen{Height: uint64(height)}) // add input height

	// now lets create encrypted keys, so as the miner_address can spend them
	// but also everyone on planet generates the same key

	//tx_secret_key := crypto.Key(seed)
	//crypto.ScReduce32(&tx_secret_key)
	// tx_public_key := tx_secret_key.PublicKey()

	tx_secret_key, tx_public_key := crypto.NewKeyPair() // create new tx key pair

	//tx_public_key is added to extra and serialized
	tx.Extra_map[transaction.TX_PUBLIC_KEY] = *tx_public_key

	derivation := crypto.KeyDerivation(&miner_address.ViewKey, tx_secret_key) // keyderivation using miner address view key

	index_within_tx := uint64(0)

	// this becomes the key within Vout
	ehphermal_public_key := derivation.KeyDerivation_To_PublicKey(index_within_tx, miner_address.SpendKey)

	// added the amount and key in vout
	tx.Vout = append(tx.Vout, transaction.Tx_out{Amount: 0, Target: transaction.Txout_to_key{Key: ehphermal_public_key}})

	tx.Extra = tx.Serialize_Extra() // serialize the extra
	// add 0 byte ringct signature
	var sig ringct.RctSig
	tx.RctSignature = &sig

	return
}

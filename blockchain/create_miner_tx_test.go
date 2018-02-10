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

//import "fmt"

import "testing"

import "github.com/deroproject/derosuite/walletapi"
import "github.com/deroproject/derosuite/crypto"

//import "github.com/deroproject/derosuite/address"
import "github.com/deroproject/derosuite/transaction"

// file to test whether the miner tx is created successfully and can be serialized/decoded successfully by the block miner

func Test_Create_Miner_TX(t *testing.T) {

	for i := 0; i < 1; i++ {

		hf_version := uint64(0)
		height := uint64(i)
		reward := uint64(i + 1)

		account, _ := walletapi.Generate_Keys_From_Random()

		miner_address := account.GetAddress()

		miner_tx_original, err := Create_Miner_TX(hf_version, height, reward, miner_address, 0)

		if err != nil {
			t.Fatalf("error creating miner tx, err :%s", err)
		}

		miner_tx_original_serialized := miner_tx_original.Serialize()

		var miner_tx_parsed transaction.Transaction

		err = miner_tx_parsed.DeserializeHeader(miner_tx_original_serialized)
		if err != nil {
			t.Fatalf("error parsing created miner tx, err :%s", err)
		}

		miner_tx_parsed.Parse_Extra() // parse the extra

		if miner_tx_parsed.Vin[0].(transaction.Txin_gen).Height != height {
			t.Fatalf("miner tx  height mismatch")
		}

		if miner_tx_parsed.Vout[0].Amount != reward {
			t.Fatalf("miner tx  reward mismatch")
		}

		// check whether we can decode it  output

		public_key := miner_tx_parsed.Extra_map[transaction.TX_PUBLIC_KEY].(crypto.Key)
		vout_key := miner_tx_parsed.Vout[0].Target.(transaction.Txout_to_key).Key
		index_within_tx := uint64(0)
		if !account.Is_Output_Ours(public_key, index_within_tx, vout_key) {
			t.Fatalf("miner tx  cannot be decrypted by the wallet")
		}

	}

}

func Test_Create_Miner_TX_with_extra(t *testing.T) {

	for i := 0; i < 1; i++ {

		hf_version := uint64(0)
		height := uint64(i)
		reward := uint64(i + 1)

		account, _ := walletapi.Generate_Keys_From_Random()

		miner_address := account.GetAddress()

		miner_tx_original, err := Create_Miner_TX(hf_version, height, reward, miner_address, 60)

		if err != nil {
			t.Fatalf("error creating miner tx, err :%s", err)
		}

		miner_tx_original_serialized := miner_tx_original.Serialize()

		var miner_tx_parsed transaction.Transaction

		err = miner_tx_parsed.DeserializeHeader(miner_tx_original_serialized)
		if err != nil {
			t.Fatalf("error parsing created miner tx, err :%s", err)
		}

		miner_tx_parsed.Parse_Extra() // parse the extra

		if miner_tx_parsed.Vin[0].(transaction.Txin_gen).Height != height {
			t.Fatalf("miner tx  height mismatch")
		}

		if miner_tx_parsed.Vout[0].Amount != reward {
			t.Fatalf("miner tx  reward mismatch")
		}

		// check whether we can decode it  output

		public_key := miner_tx_parsed.Extra_map[transaction.TX_PUBLIC_KEY].(crypto.Key)
		vout_key := miner_tx_parsed.Vout[0].Target.(transaction.Txout_to_key).Key
		index_within_tx := uint64(0)
		if !account.Is_Output_Ours(public_key, index_within_tx, vout_key) {
			t.Fatalf("miner tx  cannot be decrypted by the wallet")
		}

		extra_data := miner_tx_parsed.Extra_map[transaction.TX_EXTRA_NONCE].([]byte)

		if len(extra_data) != 60 {
			t.Fatalf("miner tx  extra data mismatch")
		}

	}

}

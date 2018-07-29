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

import "encoding/hex"

import "github.com/romana/rlog"

//import "github.com/deroproject/derosuite/address"
import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/block"
import "github.com/deroproject/derosuite/globals"

//import "github.com/deroproject/derosuite/config"

/*
func Create_Miner_Transaction(height uint64, median_size uint64, already_generated_coins uint64,
	current_block_size uint64, fee uint64,
	miner_address address.Address, nonce []byte,
	max_outs uint64, hard_fork uint64) (tx *transaction.Transaction, err error) {

	return nil, nil
}
*/

// genesis transaction  hash  5a18d9489bcd353aeaf4a19323d04e90353f98f0d7cc2a030cfd76e19495547d
// genesis amount 35184372088831
func Generate_Genesis_Block() (bl block.Block) {

	genesis_tx_blob, err := hex.DecodeString(globals.Config.Genesis_Tx)
	if err != nil {
		panic("Failed to hex decode genesis Tx")
	}
	err = bl.Miner_TX.DeserializeHeader(genesis_tx_blob)

	if err != nil {
		panic("Failed to parse genesis tx ")
	}

	//rlog.Tracef(2, "Hash of Genesis Tx %x\n", bl.Miner_tx.GetHash())

	// verify whether tx is coinbase and valid

	// setup genesis block header
	bl.Major_Version = 1
	bl.Minor_Version = 1
	bl.Timestamp = 0 // first block timestamp

	var zerohash crypto.Hash
	_ = zerohash
	//bl.Tips = append(bl.Tips,zerohash)
	//bl.Prev_hash is automatic zero
	bl.Nonce = globals.Config.Genesis_Nonce

	rlog.Tracef(2, "Hash of genesis block is %x", bl.GetHash())

	serialized := bl.Serialize()

	var bl2 block.Block
	err = bl2.Deserialize(serialized)
	if err != nil {
		panic("error while serdes genesis block")
	}
	if bl.GetHash() != bl2.GetHash() {
		panic("hash mismatch serdes genesis block")
	}

	//rlog.Tracef(2, "Genesis Block PoW %x\n", bl.GetPoWHash())

	return
}

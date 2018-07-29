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

package globals

//import  "github.com/vmihailenco/msgpack"

import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/crypto/ringct"

// this structure is way blockchain sends outputs to wallet
// this structure is also used internal by blockchain itself, when TXs are expanded and verified

// this structure needs to be expandable without breaking legacy
// This structure is the only information needed by wallet to decode previous and send new txs
// This opens up the case for mobile wallets without downloading entire blockchain
// Some of the fields get duplicated if a tx has multiple vouts
type TX_Output_Data struct {
	BLID          crypto.Hash `msgpack:"L" json:"blid,omitempty"`       // the block id in which this was found, we are duplicating it for all
	TXID          crypto.Hash `msgpack:"T" json:"txid,omitempty"`       // the transaction id in which this was found, we are duplicating it for all vouts
	Tx_Public_Key crypto.Key  `msgpack:"P"  json:"publickey,omitempty"` // the public key of the tx ,  we are duplicating it for all vouts within the TX
	// this is extracted from the extra field

	// this data is consumed by the blockchain itself while expanding tx
	InKey           ringct.CtKey     `msgpack:"D"  json:"inkey,omitempty"`           // contains the vout key and encrypted commitment
	ECDHTuple       ringct.ECdhTuple `msgpack:"E" json:"ecdhtuple,omitempty"`        // encrypted Amounts, will be empty for miner tx
	SenderPk        crypto.Key       `msgpack:"K" json:"senderpk,omitempty"`         // from the outpk
	Amount          uint64           `msgpack:"A,omitempty" json:"amount,omitempty"` // amount used for miner tx
	SigType         uint64           `msgpack:"S" json:"sigtype,omitempty"`          // which ringct type the output was of
	Index_within_tx uint64           `msgpack:"W" json:"indexwithintx,omitempty"`    // output index  of vout within the tx
	Index_Global    uint64           `msgpack:"G" json:"indexglobal,omitempty"`      // position in global index
	Height          uint64           `msgpack:"H" json:"height,omitempty"`           // height to which this belongs
	Unlock_Height   uint64           `msgpack:"U" json:"unlockheight,omitempty"`     // height at which it will unlock
	TopoHeight      int64            `msgpack:"TH" json:"Topoheight,omitempty"`      // Topoheight
	Block_Time      uint64           `msgpack:"B"`                                   // when was this block found in epoch

	Key_Images []crypto.Key `msgpack:"KI,omitempty"`                           // all the key images consumed within the TX
	PaymentID  []byte       `msgpack:"I,omitempty" json:"paymentid,omitempty"` // payment ID contains both unencrypted (33byte)/encrypted (9 bytes)

	// TODO this structure must also keep payment ids and encrypted payment ids
}

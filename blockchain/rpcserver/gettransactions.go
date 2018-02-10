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

package rpcserver

import "net/http"
import "encoding/hex"
import "encoding/json"

import "github.com/deroproject/derosuite/crypto"

// we definitely need to clear up the MESS that has been created by the MONERO project
// half of their APIs are json rpc and half are http
// for compatibility reasons, we are implementing theirs ( however we are also providin a json rpc implementation)
// we should DISCARD the http

//  NOTE: we have currently not implemented the decode as json parameter
//  it is however on the pending list

type (
	GetTransaction_Handler struct{}
	GetTransaction_Params  struct {
		Tx_Hashes []string `json:"txs_hashes"`
		Decode    uint64   `json:"decode_as_json,omitempty"` // Monero Daemon breaks if this sent
	} // no params
	GetTransaction_Result struct {
		Txs_as_hex  []string          `json:"txs_as_hex"`
		Txs_as_json []string          `json:"txs_as_json"`
		Txs         []Tx_Related_Info `json:"txs"`
		Status      string            `json:"status"`
	}

	Tx_Related_Info struct {
		As_Hex         string   `json:"as_hex"`
		As_Json        string   `json:"as_json"`
		Block_Height   int64    `json:"block_height"`
		In_pool        bool     `json:"in_pool"`
		Output_Indices []uint64 `json:"output_indices"`
		Tx_hash        string   `json:"tx_hash"`
	}
)

func gettransactions(rw http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var p GetTransaction_Params
	err := decoder.Decode(&p)
	if err != nil {
		panic(err)
	}
	defer req.Body.Close()

	result := gettransactions_fill(p)
	//logger.Debugf("Request %+v", p)

	encoder := json.NewEncoder(rw)
	encoder.Encode(result)
}

// fill up the response
func gettransactions_fill(p GetTransaction_Params) (result GetTransaction_Result) {

	for i := 0; i < len(p.Tx_Hashes); i++ {

		hash := crypto.HashHexToHash(p.Tx_Hashes[i])

		// check whether we can get the tx from the pool
		{
			tx := chain.Mempool.Mempool_Get_TX(hash)
			if tx != nil { // found the tx in the mempool
				result.Txs_as_hex = append(result.Txs_as_hex, hex.EncodeToString(tx.Serialize()))

				var related Tx_Related_Info

				related.Block_Height = -1 // not mined
				related.In_pool = true

				for i := 0; i < len(tx.Vout); i++ {
					related.Output_Indices = append(related.Output_Indices, 0) // till the tx is mined we do not get indices
				}

				result.Txs = append(result.Txs, related)

				continue // no more processing required
			}
		}

		tx, err := chain.Load_TX_FROM_ID(hash)
		if err == nil {
			result.Txs_as_hex = append(result.Txs_as_hex, hex.EncodeToString(tx.Serialize()))
			var related Tx_Related_Info

			related.Block_Height = int64(chain.Load_TX_Height(hash))

			index := chain.Find_TX_Output_Index(hash)

			//   logger.Infof("TX hash %s height %d",hash, related.Block_Height)
			for i := 0; i < len(tx.Vout); i++ {
				related.Output_Indices = append(related.Output_Indices, index+uint64(i))
			}

			result.Txs = append(result.Txs, related)

		} else { // we could not fetch the tx, return an empty string
			result.Txs_as_hex = append(result.Txs_as_hex, "")
			result.Status = "TX NOT FOUND"
			return
		}

	}
	result.Status = "OK"
	return
}

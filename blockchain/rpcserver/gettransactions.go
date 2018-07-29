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

import "github.com/romana/rlog"
import "github.com/vmihailenco/msgpack"

import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/structures"
import "github.com/deroproject/derosuite/transaction"

// we definitely need to clear up the MESS that has been created by the MONERO project
// half of their APIs are json rpc and half are http
// for compatibility reasons, we are implementing theirs ( however we are also providin a json rpc implementation)
// we should DISCARD the http

//  NOTE: we have currently not implemented the decode as json parameter
//  it is however on the pending list

type GetTransaction_Handler struct{}

func gettransactions(rw http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var p structures.GetTransaction_Params
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
func gettransactions_fill(p structures.GetTransaction_Params) (result structures.GetTransaction_Result) {

	for i := 0; i < len(p.Tx_Hashes); i++ {

		hash := crypto.HashHexToHash(p.Tx_Hashes[i])

		{ // check if tx is from  blockchain
			tx, err := chain.Load_TX_FROM_ID(nil, hash)
			if err == nil {
				var related structures.Tx_Related_Info

				// check whether tx is orphan

				/*if chain.Is_TX_Orphan(hash) {
					result.Txs_as_hex = append(result.Txs_as_hex, "") // given empty data
					result.Txs = append(result.Txs, related)          // should we have an orphan tx marker
				} else */{

					// topo height at which it was mined
					related.Block_Height = int64(chain.Load_TX_Height(nil, hash))

					if tx.IsCoinbase() { // fill reward but only for coinbase
						blhash, err := chain.Load_Block_Topological_order_at_index(nil, int64(related.Block_Height))
						if err == nil { // if err return err
							related.Reward = chain.Load_Block_Total_Reward(nil, blhash)
						}
					}

					if !tx.IsCoinbase() {
						// expand ring members and provide information
						related.Ring = make([][]globals.TX_Output_Data, len(tx.Vin), len(tx.Vin))
						for i := 0; i < len(tx.Vin); i++ {
							related.Ring[i] = make([]globals.TX_Output_Data, len(tx.Vin[i].(transaction.Txin_to_key).Key_offsets), len(tx.Vin[i].(transaction.Txin_to_key).Key_offsets))
							ring_member := uint64(0)
							for j := 0; j < len(tx.Vin[i].(transaction.Txin_to_key).Key_offsets); j++ {
								ring_member += tx.Vin[i].(transaction.Txin_to_key).Key_offsets[j]

								var ring_data globals.TX_Output_Data
								data_bytes, err := chain.Read_output_index(nil, ring_member)

								err = msgpack.Unmarshal(data_bytes, &ring_data)
								if err != nil {
									rlog.Warnf("RPC err while unmarshallin output index data index = %d  data_len %d err %s", ring_member, len(data_bytes), err)
								}
								related.Ring[i][j] = ring_data
							}
						}
						err = nil
					}

					// also fill where the tx is found and in which block is valid and in which it is invalid

					blocks := chain.Load_TX_blocks(nil, hash)

					for i := range blocks {

						//  logger.Infof("%s tx valid %+v",blocks[i],chain.IS_TX_Valid(nil,blocks[i],hash))
						//  logger.Infof("%s block topo %+v", blocks[i], chain.Is_Block_Topological_order(nil,blocks[i]))
						if chain.IS_TX_Valid(nil, blocks[i], hash) && chain.Is_Block_Topological_order(nil, blocks[i]) {
							related.ValidBlock = blocks[i].String()
						} else {

							related.InvalidBlock = append(related.InvalidBlock, blocks[i].String())
						}

					}

					index := chain.Find_TX_Output_Index(hash)
					// index := uint64(0)

					//   logger.Infof("TX hash %s height %d",hash, related.Block_Height)
					for i := 0; i < len(tx.Vout); i++ {
						if index >= 0 {
							related.Output_Indices = append(related.Output_Indices, uint64(index+int64(i)))
						} else {
							related.Output_Indices = append(related.Output_Indices, 0)
						}
					}
					result.Txs_as_hex = append(result.Txs_as_hex, hex.EncodeToString(tx.Serialize()))
					result.Txs = append(result.Txs, related)
				}
				continue
			}
		}
		// check whether we can get the tx from the pool
		{
			tx := chain.Mempool.Mempool_Get_TX(hash)
			if tx != nil { // found the tx in the mempool
				result.Txs_as_hex = append(result.Txs_as_hex, hex.EncodeToString(tx.Serialize()))

				var related structures.Tx_Related_Info

				related.Block_Height = -1 // not mined
				related.In_pool = true

				for i := 0; i < len(tx.Vout); i++ {
					related.Output_Indices = append(related.Output_Indices, 0) // till the tx is mined we do not get indices
				}

				result.Txs = append(result.Txs, related)

				continue // no more processing required
			}
		}

		{ // we could not fetch the tx, return an empty string
			result.Txs_as_hex = append(result.Txs_as_hex, "")
			result.Status = "TX NOT FOUND"
			return
		}

	}
	result.Status = "OK"
	return
}

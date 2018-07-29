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

package p2p

//import "fmt"
//import "net"
//import "sync"
import "time"

import "encoding/binary"

//import "container/list"

import "github.com/romana/rlog"
import "github.com/vmihailenco/msgpack"

import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/block"
import "github.com/deroproject/derosuite/errormsg"
import "github.com/deroproject/derosuite/transaction"

// Peer has notified us of a new transaction
func (connection *Connection) Handle_Notification_Transaction(buf []byte) {
	var request Notify_New_Objects_Struct

	err := msgpack.Unmarshal(buf, &request)
	if err != nil {
		rlog.Warnf("Error while decoding incoming TX notifcation err %s %s", err, globals.CTXString(connection.logger))
		connection.Exit()
	}

	var tx transaction.Transaction
	err = tx.DeserializeHeader(request.Tx)
	if err != nil { // we have a tx which could not be deserialized ban peer
		rlog.Warnf("Error Incoming TX could not be deserialized err %s %s", err, globals.CTXString(connection.logger))
		connection.Exit()
		return
	}

	// track transaction propagation
	if first_time, ok := tx_propagation_map.Load(tx.GetHash()); ok {
		// block already has a reference, take the time and observe the value
		diff := time.Now().Sub(first_time.(time.Time)).Round(time.Millisecond)
		transaction_propagation.Observe(float64(diff / 1000000))
	} else {
		tx_propagation_map.Store(tx.GetHash(), time.Now()) // if this is the first time, store the tx time
	}

	// try adding tx to pool
	success_pool := chain.Add_TX_To_Pool(&tx)

	// add tx to cache  of the peer who sent us this tx
	connection.TXpool_cache_lock.Lock()
	if success_pool && globals.Arguments["--lowcpuram"].(bool) == false && connection.TXpool_cache != nil {

		txhash := tx.GetHash()
		connection.TXpool_cache[binary.LittleEndian.Uint64(txhash[:])] = uint32(time.Now().Unix())

		//logger.Debugf("Adding %s to cache", tx.GetHash())
	}
	connection.TXpool_cache_lock.Unlock()

	// broadcasting of tx is controlled by mempool

}

// Peer has notified us of a new block
func (connection *Connection) Handle_Notification_Block(buf []byte) {
	var request Notify_New_Objects_Struct

	err := msgpack.Unmarshal(buf, &request)
	if err != nil {
		rlog.Warnf("Error while decoding incoming Block notifcation request err %s %s", err, globals.CTXString(connection.logger))
		connection.Exit()
	}

	var cbl block.Complete_Block // parse incoming block and deserialize it
	var bl block.Block
	// lets deserialize block first and see whether it is the requested object
	cbl.Bl = &bl
	err = bl.Deserialize(request.CBlock.Block)
	if err != nil { // we have a block which could not be deserialized ban peer
		rlog.Warnf("Error Incoming block could not be deserilised err %s %s", err, globals.CTXString(connection.logger))
		connection.Exit()
		return
	}

	blid := bl.GetHash()

	rlog.Infof("Incoming block Notification hash %s %s ", blid, globals.CTXString(connection.logger))

	// track block propagation
	if first_time, ok := block_propagation_map.Load(blid); ok {
		// block already has a reference, take the time and observe the value
		diff := time.Now().Sub(first_time.(time.Time)).Round(time.Millisecond)
		block_propagation.Observe(float64(diff / 1000000))
	} else {
		block_propagation_map.Store(blid, time.Now()) // if this is the first time, store the block
	}

	// object is already is in our chain, we need not relay it
	if chain.Block_Exists(nil, blid) {
		return
	}

	// the block is not in our db,  parse entire block, complete the txs and try to add it
	if len(bl.Tx_hashes) == len(request.CBlock.Txs) {
		connection.logger.Debugf("Received a complete block %s with %d transactions",blid, len(bl.Tx_hashes))
		for j := range request.CBlock.Txs {
			var tx transaction.Transaction
			err = tx.DeserializeHeader(request.CBlock.Txs[j])
			if err != nil { // we have a tx which could not be deserialized ban peer
				rlog.Warnf("Error Incoming TX could not be deserialized err %s %s", err, globals.CTXString(connection.logger))
				connection.Exit()
				return
			}
			cbl.Txs = append(cbl.Txs, &tx)
		}
	} else { // the block is NOT complete, we consider it as an ultra compact block

		connection.logger.Debugf("Received an ultra compact block %s, total %d contains %d skipped %d transactions", blid,len(bl.Tx_hashes), len(request.CBlock.Txs), len(bl.Tx_hashes)-len(request.CBlock.Txs))
		for j := range request.CBlock.Txs {
			var tx transaction.Transaction
			err = tx.DeserializeHeader(request.CBlock.Txs[j])
			if err != nil { // we have a tx which could not be deserialized ban peer
				rlog.Warnf("Error Incoming TX could not be deserialized err %s %s", err, globals.CTXString(connection.logger))
				connection.Exit()
				return
			}
			chain.Add_TX_To_Pool(&tx) // add tx to pool
		}

		// lets build a complete block ( tx from db or mempool )
		for i := range bl.Tx_hashes {
			if tx, err := chain.Load_TX_FROM_ID(nil, bl.Tx_hashes[i]); err == nil {
				cbl.Txs = append(cbl.Txs, tx) // tx is from disk
			} else {
				tx := chain.Mempool.Mempool_Get_TX(bl.Tx_hashes[i]) // tx is from mempool
				if tx != nil {
					cbl.Txs = append(cbl.Txs, tx)
				} else {
					// the tx mentioned in ultra compact block could not be found
					// request a full block

					connection.Send_ObjectRequest([]crypto.Hash{blid}, []crypto.Hash{})
					logger.Debugf("Ultra compact block  %s missing TX %s, requesting full block", blid, bl.Tx_hashes[i])
					return
				}
			}

		}
	}

	// check if we can add ourselves to chain
	if err, ok := chain.Add_Complete_Block(&cbl); ok { // if block addition was successfil
		// notify all peers
		Broadcast_Block(&cbl, connection.Peer_ID) // do not send back to the original peer

	} else { // ban the peer for sometime
		if err == errormsg.ErrInvalidPoW {
			connection.logger.Warnf("This peer should be banned and terminated")
			connection.Exit()
		}
	}

}

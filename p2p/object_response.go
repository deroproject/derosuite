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
import "sync/atomic"
import "time"

//import "container/list"

//import log "github.com/sirupsen/logrus"
import "github.com/romana/rlog"
import "github.com/vmihailenco/msgpack"

//import "github.com/deroproject/derosuite/crypto"
//import "github.com/deroproject/derosuite/globals"

import "github.com/deroproject/derosuite/block"
import "github.com/deroproject/derosuite/errormsg"
import "github.com/deroproject/derosuite/transaction"

// peer has responded with  some objects, we must respond
func (connection *Connection) Handle_ObjectResponse(buf []byte) {
	var response Object_Response_struct

	err := msgpack.Unmarshal(buf, &response)
	if err != nil {
		rlog.Warnf("Error while decoding incoming object response err %s %s", err, connection.logid)
		connection.Exit()
	}

	var expected Queued_Command

	select {
	case expected = <-connection.Objects:

	default: // if nothing is on queue the peer sent us bogus request,
		rlog.Warnf("Peer sent us a chain response, when we didnot request chain, Exiting, may be block the peer %s", connection.logid)
		connection.Exit()
	}

	if expected.Command != V2_COMMAND_OBJECTS_RESPONSE {
		rlog.Warnf("We were waiting for a different object, but peer sent something else, Exiting, may be block the peer %s", connection.logid)
		connection.Exit()
	}

	// we need to verify common and update common

	if len(response.CBlocks) != len(expected.BLID) { // we requested x block , peer sent us y blocks, time to ban peer
		rlog.Warnf("we got %d response for %d requests %s %s", len(response.CBlocks), len(expected.BLID), connection.logid)
	}

	for i := 0; i < len(response.CBlocks); i++ { // process incoming full blocks
		var cbl block.Complete_Block // parse incoming block and deserialize it
		var bl block.Block
		// lets deserialize block first and see whether it is the requested object
		cbl.Bl = &bl
		err := bl.Deserialize(response.CBlocks[i].Block)
		if err != nil { // we have a block which could not be deserialized ban peer
			rlog.Warnf("Error Incoming block could not be deserilised err %s %s", err, connection.logid)
			connection.Exit()
			return
		}

		// check if deserialized hash is same as what we requested
		if bl.GetHash() != expected.BLID[i] { // user is trying to spoof block, ban hime
			connection.logger.Warnf("requested and response block mismatch")
			rlog.Warnf("Error block hash mismatch Actual %s Expected %s err %s %s", bl.GetHash(), expected.BLID[i], connection.logid)
			connection.Exit()
		}

		// give the chain some more time to respond
		atomic.StoreInt64(&connection.LastObjectRequestTime, time.Now().Unix())

		// check whether the object was requested one

		// complete the txs
		for j := range response.CBlocks[i].Txs {
			var tx transaction.Transaction
			err = tx.DeserializeHeader(response.CBlocks[i].Txs[j])
			if err != nil { // we have a tx which could not be deserialized ban peer
				rlog.Warnf("Error Incoming TX could not be deserialized err %s %s", err, connection.logid)
				connection.Exit()

				return
			}
			cbl.Txs = append(cbl.Txs, &tx)
		}

		// check if we can add ourselves to chain
		err, ok := chain.Add_Complete_Block(&cbl)
		if !ok && err == errormsg.ErrInvalidPoW {
			connection.logger.Warnf("This peer should be banned")
			connection.Exit()
		}

		// add the object to object pool from where it will be consume
		// queue_block_received(bl.GetHash(),&cbl)

	}

}

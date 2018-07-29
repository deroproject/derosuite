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

import "github.com/romana/rlog"
import "github.com/vmihailenco/msgpack"

//import "github.com/deroproject/derosuite/crypto"
//import "github.com/deroproject/derosuite/globals"

import "github.com/deroproject/derosuite/crypto"

//import "github.com/deroproject/derosuite/globals"

//import "github.com/deroproject/derosuite/blockchain"

//  we are sending object request
// right now we only send block ids
func (connection *Connection) Send_ObjectRequest(blids []crypto.Hash, txids []crypto.Hash) {

	var request Object_Request_Struct
	fill_common(&request.Common) // fill common info
	request.Command = V2_COMMAND_OBJECTS_REQUEST

	for i := range blids {
		request.Block_list = append(request.Block_list, blids[i])
	}

	for i := range txids {
		request.Tx_list = append(request.Tx_list, txids[i])
	}

	if len(blids) > 0 || len(txids) > 0 {
		serialized, err := msgpack.Marshal(&request) // serialize and send
		if err != nil {
			panic(err)
		}

		// use first object

		command := Queued_Command{Command: V2_COMMAND_OBJECTS_RESPONSE, BLID: blids, TXID: txids}

		connection.Objects <- command
		atomic.StoreInt64(&connection.LastObjectRequestTime, time.Now().Unix())

		// we should add to queue that we are waiting for object response
		//command := Queued_Command{Command: V2_COMMAND_OBJECTS_RESPONSE, BLID: blids, TXID: txids, Started: time.Now()}

		connection.Lock()
		//connection.Command_queue.PushBack(command) // queue command
		connection.Send_Message_prelocked(serialized)
		connection.Unlock()
		rlog.Tracef(3, "object request sent contains %d blids %d txids %s ", len(blids), connection.logid)
	}
}

// peer has requested some objects, we must respond
// if certain object is not in our list we respond with empty buffer for that slot
func (connection *Connection) Handle_ObjectRequest(buf []byte) {
	var request Object_Request_Struct
	var response Object_Response_struct

	err := msgpack.Unmarshal(buf, &request)
	if err != nil {
		rlog.Warnf("Error while decoding incoming object request err %s %s", err, connection.logid)
		connection.Exit()
	}

	if len(request.Block_list) < 1 { // we are expecting 1 block
		rlog.Warnf("malformed object request  received, banning peer %+v %s", request, connection.logid)
		connection.Exit()
	}

	for i := 0; i < len(request.Block_list); i++ { // find the common point in our chain
		var cbl Complete_Block
		if chain.Block_Exists(nil, request.Block_list[i]) {
			bl, _ := chain.Load_BL_FROM_ID(nil, request.Block_list[i])
			cbl.Block = bl.Serialize()
			for j := range bl.Tx_hashes {
				tx, err := chain.Load_TX_FROM_ID(nil, bl.Tx_hashes[j])

				if err != nil {
					//rlog.Tracef(1, "ERR Cannot load tx from DB\n")
					return
				}
				cbl.Txs = append(cbl.Txs, tx.Serialize()) // append all the txs
			}
		}

		response.CBlocks = append(response.CBlocks, cbl)
	}

	// we can serve maximum of 1024 BLID = 32 KB

	// if everything is OK, we must respond with object response
	fill_common(&response.Common) // fill common info
	response.Command = V2_COMMAND_OBJECTS_RESPONSE

	serialized, err := msgpack.Marshal(&response) // serialize and send
	if err != nil {
		panic(err)
	}

	rlog.Tracef(3, "OBJECT RESPONSE SENT  sent size %d %s", len(serialized), connection.logid)
	connection.Send_Message(serialized)
}

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
//import "time"

//import "container/list"

//import log "github.com/sirupsen/logrus"
import "github.com/romana/rlog"
import "github.com/vmihailenco/msgpack"

import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/config"

//import "github.com/deroproject/derosuite/blockchain"

// peer has responded with chain response and we must process the response
func (connection *Connection) Handle_ChainResponse(buf []byte) {
	var response Chain_Response_Struct

	err := msgpack.Unmarshal(buf, &response)
	if err != nil {
		rlog.Warnf("Error while decoding incoming chain response err %s %s", err, connection.logid)
		connection.Exit()
		return
	}

	// check what we we queued is what what we got back
	// for chain request we queue an empty hash

	var expected Queued_Command

	select {
	case expected = <-connection.Objects:

	default: // if nothing is on queue the peer sent us bogus request,
		rlog.Warnf("Peer sent us a chain response, when we didnot request chain, Exiting, may be block the peer %s", connection.logid)
		connection.Exit()
	}

	if expected.Command != V2_COMMAND_CHAIN_RESPONSE {
		rlog.Warnf("We were waiting for a different object, but peer sent something else, Exiting, may be block the peer %s", connection.logid)
		connection.Exit()
	}

	// we were expecting something else ban
	if len(response.Block_list) < 1 {
		rlog.Warnf("Malformed chain response  %s", err, connection.logid)
		connection.Exit()
		return
	}

	rlog.Tracef(2, "Peer wants to give chain from topoheight %d ", response.Start_height)
        _ = config.STABLE_LIMIT

	// we do not need reorganisation if deviation is less than  or equak to 7 blocks
	// only pop blocks if the system has somehow deviated more than 7 blocks
	// if the deviation is less than 7 blocks, we internally reorganise everything
	if chain.Load_TOPO_HEIGHT(nil)-response.Start_topoheight >= config.STABLE_LIMIT && connection.SyncNode {
		// get our top block
		rlog.Infof("rewinding status our %d  peer %d", chain.Load_TOPO_HEIGHT(nil), response.Start_topoheight)
		pop_count := chain.Load_TOPO_HEIGHT(nil) - response.Start_topoheight
		chain.Rewind_Chain(int(pop_count)) // pop as many blocks as necessary

		// we should NOT queue blocks, instead we sent our chain request again
		connection.Send_ChainRequest()
		return

	}

	// response only 128 blocks at a time
	max_blocks_to_queue := 128
	// check whether the objects are in our db or not
	// until we put in place a parallel object tracker, do it one at a time
	for i := range response.Block_list {
		if !chain.Block_Exists(nil, response.Block_list[i]) { // if block is not in our chain, add it to request list
			//queue_block(request.Block_list[i])
			if max_blocks_to_queue >= 0 {
				max_blocks_to_queue--
				connection.Send_ObjectRequest([]crypto.Hash{response.Block_list[i]}, []crypto.Hash{})
				rlog.Tracef(2, "Queuing block %x height %d  %s", response.Block_list[i], response.Start_height+int64(i), connection.logid)
			}
		} else {
			//logger.Warnf("We must have queued %x, but we skipped it at height %d",request.Block_list[i],request.Start_height+int64(i) )
		}
	}

	// request alt-tips ( blocks if we are nearing the main tip )
	if (response.Common.TopoHeight - chain.Load_TOPO_HEIGHT(nil)) <= 5 {
		for i := range response.TopBlocks {
			if !chain.Block_Exists(nil, response.TopBlocks[i]) {
				connection.Send_ObjectRequest([]crypto.Hash{response.TopBlocks[i]}, []crypto.Hash{})
				rlog.Tracef(2, "Queuing ALT-TIP  block %x %s", response.TopBlocks[i], connection.logid)

			}

		}
	}
}

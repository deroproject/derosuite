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
import "github.com/deroproject/derosuite/globals"

//import "github.com/deroproject/derosuite/blockchain"

//  we are sending a chain request, build a packet with our chain data
// so as other side can respond with chain response
func (connection *Connection) Send_ChainRequest() {
	var request Chain_Request_Struct

	fill_common(&request.Common) // fill common info
	request.Command = V2_COMMAND_CHAIN_REQUEST

	// send our blocks, first 10 blocks directly, then decreasing in powers of 2
	start_point := chain.Load_TOPO_HEIGHT(nil)
	for i := int64(0); i < start_point; {

		blid, _ := chain.Load_Block_Topological_order_at_index(nil, start_point-i)
		request.Block_list = append(request.Block_list, blid)
		request.TopoHeights = append(request.TopoHeights, start_point-i)
		rlog.Tracef(3, "Adding block to chain request h %d %s", i, blid)
		switch {
		case len(request.Block_list) < 10:
			i++
		default:
			i = i * 2
		}
	}

	// add genesis block at the end
	request.Block_list = append(request.Block_list, globals.Config.Genesis_Block_Hash)
	request.TopoHeights = append(request.TopoHeights, 0)

	// serialize and send
	serialized, err := msgpack.Marshal(&request)
	if err != nil {
		panic(err)
	}

	// queue command
	command := Queued_Command{Command: V2_COMMAND_CHAIN_RESPONSE}
	connection.Objects <- command

	atomic.StoreInt64(&connection.LastObjectRequestTime, time.Now().Unix())
	//connection.Lock()
	//connection.Command_queue.PushBack(command) // queue command
	connection.Send_Message_prelocked(serialized)
	//connection.Unlock()

	rlog.Tracef(2, "chain request  sent successfully %s", globals.CTXString(connection.logger))
}

// peer has requested chain
func (connection *Connection) Handle_ChainRequest(buf []byte) {
	var request Chain_Request_Struct
	var response Chain_Response_Struct

	err := msgpack.Unmarshal(buf, &request)
	if err != nil {
		rlog.Warnf("Error while decoding incoming chain request err %s %s", err, globals.CTXString(connection.logger))
		connection.Exit()
		return
	}

	//
	if len(request.Block_list) < 1 { // malformed request ban peer
		rlog.Warnf("malformed chain request  received, banning peer %+v %s", request, globals.CTXString(connection.logger))
		connection.Exit()

		return
	}

	if len(request.Block_list) != len(request.TopoHeights) {
		rlog.Warnf("Peer chain request has %d block %d topos, therefore invalid", len(request.Block_list), len(request.TopoHeights))
		connection.Exit()
		return
	}

	if request.Block_list[len(request.Block_list)-1] != globals.Config.Genesis_Block_Hash {
		rlog.Warnf("Peer's genesis block is different from our, so disconnect Actual %s Expected %s", request.Block_list[len(request.Block_list)-1], globals.Config.Genesis_Block_Hash)
		connection.Exit()
		return
	}

	rlog.Tracef(2, "chain request received %s", globals.CTXString(connection.logger))

	// we must give user our version of the chain
	start_height := int64(0)
        start_topoheight := int64(0)

	for i := 0; i < len(request.Block_list); i++ { // find the common point in our chain ( the block is NOT orphan)

		//connection.logger.Infof("Checking block for chain detection %d %s", i, request.Block_list[i])

		if chain.Block_Exists(nil, request.Block_list[i]) && chain.Is_Block_Topological_order(nil, request.Block_list[i]) &&
			request.TopoHeights[i] == chain.Load_Block_Topological_order(nil, request.Block_list[i]) {
			start_height = chain.Load_Height_for_BL_ID(nil, request.Block_list[i])
                        start_topoheight =  chain.Load_Block_Topological_order(nil, request.Block_list[i])
			rlog.Tracef(2, "Found common point in chain at hash %x height %d topoheight %d\n", request.Block_list[i], start_height, start_topoheight)
			break
		}
	}

	// we can serve maximum of 512 BLID = 16K KB
	const MAX_BLOCKS = 512

	// if everything is OK, we must respond with chain response
	//connection.Send_TimedSync(false) // send it as response

	for i := start_topoheight; i <= chain.Load_TOPO_HEIGHT(nil) && len(response.Block_list) <= MAX_BLOCKS; i++ {
		hash, _ := chain.Load_Block_Topological_order_at_index(nil, i)
		response.Block_list = append(response.Block_list, [32]byte(hash))
	}

	// we must also fill blocks for the  last top 10 heights, so client can sync faster to alt tips
	top_height := chain.Get_Height()
	counter := 0
	for ; top_height > 0 && counter <= 10; top_height-- {
		blocks := chain.Get_Blocks_At_Height(nil, top_height)
		for i := range blocks {
			response.TopBlocks = append([][32]byte{blocks[i]}, response.TopBlocks...) // blocks are ordered height wise
		}
		counter++
	}

	response.Start_height = start_height
	response.Start_topoheight = start_topoheight
	fill_common(&response.Common) // fill common info
	response.Command = V2_COMMAND_CHAIN_RESPONSE

	// serialize and send
	serialized, err := msgpack.Marshal(&response)
	if err != nil {
		panic(err)
	}

	// we should add to queue that we are waiting for chain response
	rlog.Tracef(2, "chain response sent due to incoming chain request  sent len response = %d %s", len(serialized), globals.CTXString(connection.logger))
	connection.Send_Message(serialized)
}

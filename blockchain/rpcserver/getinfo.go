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

// get block template handler not implemented

//import "fmt"
import "context"

//import	"log"
//import 	"net/http"

import "github.com/intel-go/fastjson"
import "github.com/osamingo/jsonrpc"

import "github.com/deroproject/derosuite/config"
import "github.com/deroproject/derosuite/globals"

/*
{
  "id": "0",
  "jsonrpc": "2.0",
  "result": {
    "alt_blocks_count": 5,
    "difficulty": 972165250,
    "grey_peerlist_size": 2280,
    "height": 993145,
    "incoming_connections_count": 0,
    "outgoing_connections_count": 8,
    "status": "OK",
    "target": 60,
    "target_height": 993137,
    "testnet": false,
    "top_block_hash": "",
    "tx_count": 564287,
    "tx_pool_size": 45,
    "white_peerlist_size": 529
  }
}*/
type (
	GetInfo_Handler struct{}
	GetInfo_Params  struct{} // no params
	GetInfo_Result  struct {
		Alt_Blocks_Count           uint64 `json:"alt_blocks_count"`
		Difficulty                 uint64 `json:"difficulty"`
		Grey_PeerList_Size         uint64 `json:"grey_peerlist_size"`
		Height                     uint64 `json:"height"`
		Incoming_connections_count uint64 `json:"incoming_connections_count"`
		Outgoing_connections_count uint64 `json:"outgoing_connections_count"`
		Target                     uint64 `json:"target"`
		Target_Height              uint64 `json:"target_height"`
		Testnet                    bool   `json:"testnet"`
		Top_block_hash             string `json:"top_block_hash"`
		Tx_count                   uint64 `json:"tx_count"`
		Tx_pool_size               uint64 `json:"tx_pool_size"`
		White_peerlist_size        uint64 `json:"white_peerlist_size"`

		Status string `json:"status"`
	}
)

// TODO
func (h GetInfo_Handler) ServeJSONRPC(c context.Context, params *fastjson.RawMessage) (interface{}, *jsonrpc.Error) {
	var result GetInfo_Result

	top_id := chain.Get_Top_ID()
	result.Difficulty = chain.Get_Difficulty_At_Block(top_id)
	result.Height = chain.Get_Height() - 1
	result.Status = "OK"
	result.Top_block_hash = top_id.String()
	result.Target = config.BLOCK_TIME
	result.Target_Height = chain.Get_Height()
	result.Tx_pool_size = uint64(len(chain.Mempool.Mempool_List_TX()))

	if globals.Config.Name != config.Mainnet.Name { // anything other than mainnet is testnet at this point in time
		result.Testnet = true
	}

	return result, nil
}

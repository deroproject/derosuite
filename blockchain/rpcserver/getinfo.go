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

//import "fmt"
//import "time"
import "context"

//import	"log"
//import 	"net/http"

import "github.com/intel-go/fastjson"
import "github.com/osamingo/jsonrpc"

import "github.com/deroproject/derosuite/config"
import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/structures"

type GetInfo_Handler struct{}

// TODO
func (h GetInfo_Handler) ServeJSONRPC(c context.Context, params *fastjson.RawMessage) (interface{}, *jsonrpc.Error) {
	var result structures.GetInfo_Result

	//top_id := chain.Get_Top_ID()

	//result.Difficulty = chain.Get_Difficulty_At_Block(top_id)
	result.Height = chain.Get_Height()
	result.StableHeight = chain.Get_Stable_Height()
	result.TopoHeight = chain.Load_TOPO_HEIGHT(nil)

	blid, err := chain.Load_Block_Topological_order_at_index(nil, result.TopoHeight)
	if err == nil {
		result.Difficulty = chain.Get_Difficulty_At_Tips(nil, chain.Get_TIPS()).Uint64()
	}

	result.Status = "OK"
	result.Version = config.Version.String()
	result.Top_block_hash = blid.String()
	result.Target = config.BLOCK_TIME

	if result.TopoHeight > 50 {
		blid50, err := chain.Load_Block_Topological_order_at_index(nil, result.TopoHeight-50)
		if err == nil {
			now := chain.Load_Block_Timestamp(nil, blid)
			now50 := chain.Load_Block_Timestamp(nil, blid50)
			result.AverageBlockTime50 = float32(now-now50) / 50.0
		}
	}

	//result.Target_Height = uint64(chain.Get_Height())
	result.Tx_pool_size = uint64(len(chain.Mempool.Mempool_List_TX()))
	// get dynamic fees per kb, used by wallet for tx creation
	result.Dynamic_fee_per_kb = config.FEE_PER_KB
	result.Median_Block_Size = config.CRYPTONOTE_MAX_BLOCK_SIZE

	result.Total_Supply = chain.Load_Already_Generated_Coins_for_Topo_Index(nil, result.TopoHeight)
	if result.Total_Supply > (1000000 * 1000000000000) {
		result.Total_Supply -= (1000000 * 1000000000000) // remove  premine
	}
	result.Total_Supply = result.Total_Supply / 1000000000000

	if globals.Config.Name != config.Mainnet.Name { // anything other than mainnet is testnet at this point in time
		result.Testnet = true
	}

	return result, nil
}

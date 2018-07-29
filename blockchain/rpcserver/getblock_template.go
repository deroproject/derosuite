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
import "time"
import "context"

//import	"log"
//import 	"net/http"

import "github.com/intel-go/fastjson"
import "github.com/osamingo/jsonrpc"

import "golang.org/x/time/rate"

import "github.com/deroproject/derosuite/config"
import "github.com/deroproject/derosuite/address"
import "github.com/deroproject/derosuite/structures"
import "github.com/deroproject/derosuite/transaction"

type GetBlockTemplate_Handler struct{}

// rate limiter is deployed, in case RPC is exposed over internet
// someone should not be just giving fake inputs and delay chain syncing
var get_block_limiter = rate.NewLimiter(16.0, 8) // 16 req per sec, burst of 8 req is okay


func (h GetBlockTemplate_Handler) ServeJSONRPC(c context.Context, params *fastjson.RawMessage) (interface{}, *jsonrpc.Error) {
	var p structures.GetBlockTemplate_Params
	if err := jsonrpc.Unmarshal(params, &p); err != nil {
		return nil, err
	}
	
	
	if !get_block_limiter.Allow() { // if rate limiter allows, then add block to chain
		logger.Warnf("Too many get block template requests per sec rejected by chain.")
                
                return nil,&jsonrpc.Error{
		Code:    jsonrpc.ErrorCodeInvalidRequest,
		Message: "Too many get block template requests per sec rejected by chain.",
	}
	
            
	}

	/*
		return structures.GetBlockTemplate_Result{
				Status: "GetBlockTemplate_Handler NOT implemented",
			}, nil

	*/

	// validate address
	miner_address, err := address.NewAddress(p.Wallet_Address)
	if err != nil {
		return structures.GetBlockTemplate_Result{
			Status: "Wallet address could not be parsed",
		}, nil
	}

	if p.Reserve_size > 255 || p.Reserve_size < 1 {
		return structures.GetBlockTemplate_Result{
			Status: "Reserve size should be > 0 and < 255",
		}, nil
	}
	
	bl, block_hashing_blob_hex, block_template_hex, reserved_pos := chain.Create_new_block_template_mining(chain.Get_Top_ID(), *miner_address, int(p.Reserve_size))

	prev_hash := ""
	for i := range bl.Tips {
		prev_hash = prev_hash + bl.Tips[i].String()
	}
	return structures.GetBlockTemplate_Result{
		Blocktemplate_blob: block_template_hex,
		Blockhashing_blob:  block_hashing_blob_hex,
		Reserved_Offset:    uint64(reserved_pos),
		Expected_reward:    0, // fill in actual reward
		Height:             bl.Miner_TX.Vin[0].(transaction.Txin_gen).Height,
		Prev_Hash:          prev_hash,
		Epoch:              uint64(uint64(time.Now().UTC().Unix()) + config.BLOCK_TIME), // expiry time of this block
		Difficulty:         chain.Get_Difficulty_At_Tips(nil, bl.Tips).Uint64(),
		Status:             "OK",
	}, nil

}

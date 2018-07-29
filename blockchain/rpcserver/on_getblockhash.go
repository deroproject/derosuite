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

import "fmt"
import "context"

//import	"log"
//import 	"net/http"

import "github.com/intel-go/fastjson"
import "github.com/osamingo/jsonrpc"

import "github.com/deroproject/derosuite/crypto"

type On_GetBlockHash_Handler struct{}

func (h On_GetBlockHash_Handler) ServeJSONRPC(c context.Context, params *fastjson.RawMessage) (interface{}, *jsonrpc.Error) {

	var height [1]uint64
	//var result On_GetBlockHash_Result
	if err := jsonrpc.Unmarshal(params, &height); err != nil {
		return nil, err
	}

	if height[0] < uint64(chain.Get_Height()) {
		/*hash, err := chain.Load_BL_ID_at_Height(int64(height[0]))
		if err == nil {
			result := fmt.Sprintf("%s", hash)
			return result, nil
		} else */{
			logger.Warnf("NOT possible, Could not find block at height %d Chain height %d\n", height[0], chain.Get_Height())
		}
	} else { // we must return 0
		var hash crypto.Hash
		result := fmt.Sprintf("%s", hash)
		return result, nil

	}

	// if we reach here some error is here
	return nil, jsonrpc.ErrInvalidParams()
}

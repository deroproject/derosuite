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

import "fmt"
import "context"

//import	"log"
//import 	"net/http"

import "github.com/intel-go/fastjson"
import "github.com/osamingo/jsonrpc"

//import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/structures"

type GetBlockHeaderByHeight_Handler struct{}

func (h GetBlockHeaderByHeight_Handler) ServeJSONRPC(c context.Context, params *fastjson.RawMessage) (interface{}, *jsonrpc.Error) {
	var p structures.GetBlockHeaderByHeight_Params
	if err := jsonrpc.Unmarshal(params, &p); err != nil {
		return nil, err
	}

	if int64(p.Height) > chain.Load_TOPO_HEIGHT(nil) {
		return nil, &jsonrpc.Error{Code: -2, Message: fmt.Sprintf("Too big height: %d, current blockchain height = %d", p.Height, chain.Get_Height())}
	}

	//return nil, &jsonrpc.Error{Code: -2, Message: fmt.Sprintf("NOT SUPPORTED height: %d, current blockchain height = %d", p.Height, chain.Get_Height())}
	hash, err := chain.Load_Block_Topological_order_at_index(nil, int64(p.Height))
	if err != nil { // if err return err
		logger.Warnf("User requested %d height block, chain topo height %d but err occured %s", p.Height, chain.Get_Height(), err)

		return nil, jsonrpc.ErrInvalidParams()
	}

	block_header, err := chain.GetBlockHeader(hash)
	if err != nil { // if err return err
		logger.Warnf("User requested %d height block, chain  topo height %d but err occured %s", p.Height, chain.Get_Height(), err)

		return nil, jsonrpc.ErrInvalidParams()
	}

	return structures.GetBlockHeaderByHeight_Result{ // return success
		Block_Header: block_header,
		Status:       "OK",
	}, nil

}

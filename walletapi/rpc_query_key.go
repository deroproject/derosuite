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

package walletapi

import "fmt"
import "strings"
import "context"

//import	"log"
//import 	"net/http"

import "github.com/romana/rlog"
import "github.com/intel-go/fastjson"
import "github.com/osamingo/jsonrpc"

import "github.com/deroproject/derosuite/structures"

type Query_Key_Handler struct { // this has access to the wallet
	r *RPCServer
}

func (h Query_Key_Handler) ServeJSONRPC(c context.Context, params *fastjson.RawMessage) (interface{}, *jsonrpc.Error) {

	var p structures.Query_Key_Params
	var result structures.Query_Key_Result
	//var result structures.Transfer_Result
	//var err error

	if errp := jsonrpc.Unmarshal(params, &p); errp != nil {
		rlog.Errorf("Could not parse  query key json, err %s\n", errp)
		return nil, &jsonrpc.Error{Code: -2, Message: fmt.Sprintf("Could not parse  query key json, err %s", errp)}
	}

	// NOTE: can we give the user the spend key Secret
	// this is because we are give away the mnemonic which can anyways recreate the full wallet
	// can we disable mnemonic here
	switch {
	case strings.ToLower(p.Key_type) == "mnemonic":
		result.Key = h.r.w.GetSeed()
	case strings.ToLower(p.Key_type) == "view_key":
		result.Key = h.r.w.account.Keys.Viewkey_Secret.String()
	default:
		return nil, &jsonrpc.Error{Code: -2, Message: fmt.Sprintf("Invalid key type, must be mnemonic or view_key")}

	}
	return result, nil
}

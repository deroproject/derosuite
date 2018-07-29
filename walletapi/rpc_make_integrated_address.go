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
import "context"
import "encoding/hex"

//import	"log"
//import 	"net/http"

import "github.com/romana/rlog"
import "github.com/intel-go/fastjson"
import "github.com/osamingo/jsonrpc"

import "github.com/deroproject/derosuite/structures"

type Make_Integrated_Address_Handler struct { // this has access to the wallet
	r *RPCServer
}

func (h Make_Integrated_Address_Handler) ServeJSONRPC(c context.Context, params *fastjson.RawMessage) (interface{}, *jsonrpc.Error) {

	var p structures.Make_Integrated_Address_Params
	var result structures.Make_Integrated_Address_Result
	//var result structures.Transfer_Result
	//var err error

	if errp := jsonrpc.Unmarshal(params, &p); errp != nil {
		rlog.Errorf("Could not parse make_integrated_address json, err %s\n", errp)
		return nil, &jsonrpc.Error{Code: -2, Message: fmt.Sprintf("Could not parse make_integrated_address json, err %s", errp)}
	}

	var payment_id []byte
	if p.Payment_id != "" {
		payid, err := hex.DecodeString(p.Payment_id)
		if err != nil {
			return nil, &jsonrpc.Error{Code: -2, Message: fmt.Sprintf("%s could NOT be hex decoded err %s", p.Payment_id, err)}
		}

		if !(len(payid) == 8 || len(payid) == 32) {
			return nil, &jsonrpc.Error{Code: -2, Message: fmt.Sprintf("%s not 16 or 64 hex bytes", p.Payment_id)}
		}
		payment_id = payid
	}

	switch len(payment_id) {
	case 8:
		addr := h.r.w.GetRandomIAddress8()
		copy(addr.PaymentID, payment_id)
		result.Integrated_Address = addr.String()
		result.Payment_id = hex.EncodeToString(payment_id)
	case 32:
		addr := h.r.w.GetRandomIAddress32()
		copy(addr.PaymentID, payment_id)
		result.Integrated_Address = addr.String()
		result.Payment_id = hex.EncodeToString(payment_id)
	default:
		addr := h.r.w.GetRandomIAddress8()
		result.Integrated_Address = addr.String() // default return 8 byte encrypted payment ids
		result.Payment_id = hex.EncodeToString(addr.PaymentID)
	}

	return result, nil
}

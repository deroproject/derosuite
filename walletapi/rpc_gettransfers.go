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

import "context"
import "encoding/hex"

//import	"log"
//import 	"net/http"

import "github.com/romana/rlog"
import "github.com/intel-go/fastjson"
import "github.com/osamingo/jsonrpc"

import "github.com/deroproject/derosuite/structures"

type Get_Transfers_Handler struct { // this has access to the wallet
	r *RPCServer
}

func (h Get_Transfers_Handler) ServeJSONRPC(c context.Context, params *fastjson.RawMessage) (interface{}, *jsonrpc.Error) {

	var p structures.Get_Transfers_Params
	var result structures.Get_Transfers_Result
	//var err error

	if errp := jsonrpc.Unmarshal(params, &p); errp != nil {
		rlog.Errorf("Could not parse gettransfers json, err %s\n", errp)
		return nil, errp
	}

	//entries := h.r.w.Show_Transfers(p.In, p.In,p.Out,p.Failed, p.Pool,p.Min_Height,p.Max_Height)
	in_entries := h.r.w.Show_Transfers(p.In, p.In, false, false, false, false, p.Min_Height, p.Max_Height)
	out_entries := h.r.w.Show_Transfers(false, false, p.Out, false, false, false, p.Min_Height, p.Max_Height)
	for j := range in_entries {
		result.In = append(result.In, structures.Transfer_Details{TXID: in_entries[j].TXID.String(),
			Payment_ID:  hex.EncodeToString(in_entries[j].PaymentID),
			Height:      in_entries[j].Height,
			Amount:      in_entries[j].Amount,
			Unlock_time: in_entries[j].Unlock_Time,
			Type:        "in",
		})

	}

	for j := range out_entries {
		result.Out = append(result.Out, structures.Transfer_Details{TXID: out_entries[j].TXID.String(),
			Payment_ID:  hex.EncodeToString(out_entries[j].PaymentID),
			Height:      out_entries[j].Height,
			Amount:      out_entries[j].Amount,
			Unlock_time: out_entries[j].Unlock_Time,
			Type:        "out",
		})

	}

	return result, nil
}

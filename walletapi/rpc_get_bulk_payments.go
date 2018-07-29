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

type Get_Bulk_Payments_Handler struct { // this has access to the wallet
	r *RPCServer
}

func (h Get_Bulk_Payments_Handler) ServeJSONRPC(c context.Context, params *fastjson.RawMessage) (interface{}, *jsonrpc.Error) {

	var p structures.Get_Bulk_Payments_Params
	var result structures.Get_Bulk_Payments_Result
	//var result structures.Transfer_Result
	//var err error

	if errp := jsonrpc.Unmarshal(params, &p); errp != nil {
		rlog.Errorf("Could not parse gettransfers json, err %s\n", errp)
		return nil, &jsonrpc.Error{Code: -2, Message: fmt.Sprintf("Could not parse gettransfers json, err %s", errp)}
	}

	// if no payment ID provided, provide all entries with payment ID
	// but this is a heavy call, // south exchange needed this compatibility

	if len(p.Payment_IDs) == 0 {

		entries := h.r.w.Show_Transfers(true, true, false, false, false, true, p.Min_block_height, 0)

		for j := range entries {
			result.Payments = append(result.Payments, structures.Transfer_Details{TXID: entries[j].TXID.String(),
				Payment_ID:  hex.EncodeToString(entries[j].PaymentID),
				Height:      entries[j].Height,
				Amount:      entries[j].Amount,
				Unlock_time: entries[j].Unlock_Time,
			})
		}

		return result, nil
		//   rlog.Errorf("0 payment ids provided")
		//return nil, &jsonrpc.Error{Code: -2, Message: fmt.Sprintf("0 payment ids provided")}
	}

	for i := range p.Payment_IDs {
		payid, err := hex.DecodeString(p.Payment_IDs[i])
		if err != nil {
			return nil, &jsonrpc.Error{Code: -2, Message: fmt.Sprintf("%s could NOT be hex decoded err %s", p.Payment_IDs[i], err)}
		}

		if !(len(payid) == 8 || len(payid) == 32) {
			return nil, &jsonrpc.Error{Code: -2, Message: fmt.Sprintf("%s not 16 or 64 hex bytes", p.Payment_IDs[i])}
		}

		// if everything is okay, fire the query and convert the result to output format
		entries := h.r.w.Get_Payments_Payment_ID(payid, p.Min_block_height)
		for j := range entries {
			result.Payments = append(result.Payments, structures.Transfer_Details{TXID: entries[j].TXID.String(),
				Payment_ID:  hex.EncodeToString(entries[j].PaymentID),
				Height:      entries[j].Height,
				Amount:      entries[j].Amount,
				Unlock_time: entries[j].Unlock_Time,
			})
		}
	}

	return result, nil

}

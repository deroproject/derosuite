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

type Get_Transfer_By_TXID_Handler struct { // this has access to the wallet
	r *RPCServer
}

func (h Get_Transfer_By_TXID_Handler) ServeJSONRPC(c context.Context, params *fastjson.RawMessage) (interface{}, *jsonrpc.Error) {

	var p structures.Get_Transfer_By_TXID_Params
	var result structures.Get_Transfer_By_TXID_Result
	//var result structures.Transfer_Result
	//var err error

	if errp := jsonrpc.Unmarshal(params, &p); errp != nil {
		rlog.Errorf("Could not parse get_transfer_by_txid json, err %s\n", errp)
		return nil, &jsonrpc.Error{Code: -2, Message: fmt.Sprintf("Could not parse get_transfer_by_txid json, err %s", errp)}
	}

	txid, err := hex.DecodeString(p.TXID)
	if err != nil {
		return nil, &jsonrpc.Error{Code: -2, Message: fmt.Sprintf("%s could NOT be hex decoded err %s", p.TXID, err)}
	}

	if len(txid) != 32 {
		return nil, &jsonrpc.Error{Code: -2, Message: fmt.Sprintf("%s not 64 hex bytes", p.TXID)}
	}

	// if everything is okay, fire the query and convert the result to output format
	entry := h.r.w.Get_Payments_TXID(txid)
	result.Transfer = structures.Transfer_Details{TXID: entry.TXID.String(),
		Payment_ID:  hex.EncodeToString(entry.PaymentID),
		Height:      entry.Height,
		Amount:      entry.Amount,
		Unlock_time: entry.Unlock_Time,

	}
	if entry.Height == 0 {
		return nil, &jsonrpc.Error{Code: -8, Message: fmt.Sprintf("Transaction not found. TXID %s", p.TXID)}
	}

	for i := range entry.Details.Daddress {
		result.Transfer.Destinations = append(result.Transfer.Destinations,
			 structures.Destination{
			 	Address:  entry.Details.Daddress[i],
			 	Amount:  entry.Details.Amount[i],
			 	})
	}

	if len(entry.Details.PaymentID) >= 1 {
		result.Transfer.Payment_ID =  entry.Details.PaymentID
	}

	if entry.Status == 0 { // if we have an amount
		result.Transfer.Type = "in"
		// send the result
		return result, nil

	}
	// setup in/out
	if entry.Status == 1 { // if we have an amount
		result.Transfer.Type = "out"
		// send the result
		return result, nil

	}

	return nil, &jsonrpc.Error{Code: -8, Message: fmt.Sprintf("Transaction not found. TXID %s", p.TXID)}
}

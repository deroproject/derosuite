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
import "encoding/json"

//import	"log"
//import 	"net/http"

import "github.com/romana/rlog"
import "github.com/intel-go/fastjson"
import "github.com/osamingo/jsonrpc"

import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/address"
import "github.com/deroproject/derosuite/structures"

type TransferSplit_Handler struct { // this has access to the wallet
	r *RPCServer
}

func (h TransferSplit_Handler) ServeJSONRPC(c context.Context, params *fastjson.RawMessage) (interface{}, *jsonrpc.Error) {

	var p structures.TransferSplit_Params
	var result structures.TransferSplit_Result
	var err error

	h.r.Lock()
	defer h.r.Unlock()
	rlog.Debugf("transfer split handler")
	defer rlog.Debugf("transfer split handler finished")

	if errp := jsonrpc.Unmarshal(params, &p); err != nil {
		rlog.Errorf("Could not parse incoming json, err %s\n", errp)
		return nil, errp
	}

	//if len(p.Destinations) < 1 || p.Mixin < 4 {
	//    return nil, jsonrpc.ErrInvalidParams()
	//}

	rlog.Debugf("Len destinations %d %+v", len(p.Destinations), p)

        unlock_time := p.Unlock_time
	payment_id := p.Payment_ID
	if len(payment_id) > 0 && (len(payment_id) == 64 || len(payment_id) == 16) != true  {
		return nil, jsonrpc.ErrInvalidParams() // we should give invalid payment ID
	}
	if _, err := hex.DecodeString(p.Payment_ID); err != nil {
		return nil, jsonrpc.ErrInvalidParams() // we should give invalid payment ID
	}
	rlog.Debugf("Payment ID %s", payment_id)

	b, err := json.Marshal(p)
	if err == nil {
		rlog.Debugf("Request can be repeated using below command")

		rlog.Debugf(`curl -X POST http://127.0.0.1:18092/json_rpc -d '{"jsonrpc":"2.0","id":"0","method":"transfer_split","params":%s}' -H 'Content-Type: application/json'`, string(b))
	}

	var address_list []address.Address
	var amount_list []uint64
	for i := range p.Destinations {
		a, err := globals.ParseValidateAddress(p.Destinations[i].Address)
		if err != nil {
			rlog.Errorf("Parsing address failed %s err %s\n", p.Destinations[i].Address, err)
			return nil, jsonrpc.ErrInvalidParams()
		}
		address_list = append(address_list, *a)
		amount_list = append(amount_list, p.Destinations[i].Amount)

	}

	fees_per_kb := uint64(0) // fees  must be calculated by walletapi
	if !h.r.w.GetMode() {    // if wallet is in online mode, use the fees, provided by the daemon, else we need to use what is provided by the user

		// TODO
	}
	tx, inputs, input_sum, change, err := h.r.w.Transfer(address_list, amount_list, unlock_time, payment_id, fees_per_kb, p.Mixin)
	_ = inputs
	if err != nil {
		rlog.Warnf("Error while building Transaction err %s\n", err)
		return nil, &jsonrpc.Error{Code: -2, Message: fmt.Sprintf("Error while building Transaction err %s", err)}

	}

	rlog.Infof("Inputs Selected for %s \n", globals.FormatMoneyPrecision(input_sum, 12))
	amount := uint64(0)
	for i := range amount_list {
		amount += amount_list[i]
	}
	rlog.Infof("Transfering total amount %s \n", globals.FormatMoneyPrecision(amount, 12))
	rlog.Infof("change amount ( will come back ) %s \n", globals.FormatMoneyPrecision(change, 12))
	rlog.Infof("fees %s \n", globals.FormatMoneyPrecision(tx.RctSignature.Get_TX_Fee(), 12))

	if input_sum != (amount + change + tx.RctSignature.Get_TX_Fee()) {
		panic(fmt.Sprintf("Inputs %d != outputs ( %d + %d + %d )", input_sum, amount, change, tx.RctSignature.Get_TX_Fee()))
	}

	if p.Do_not_relay == false { // we do not relay the tx, the user must submit it manually
		err = h.r.w.SendTransaction(tx)

		if err == nil {
			rlog.Infof("Transaction sent successfully. txid = %s", tx.GetHash())
		} else {
			rlog.Warnf("Transaction sending failed txid = %s, err %s", tx.GetHash(), err)
			return nil, &jsonrpc.Error{Code: -2, Message: fmt.Sprintf("Transaction sending failed txid = %s, err %s", tx.GetHash(), err)}
		}

	}

	result.Fee_list = append(result.Fee_list, tx.RctSignature.Get_TX_Fee())
	result.Tx_hash_list = append(result.Tx_hash_list, tx.GetHash().String())
	if p.Get_tx_hex { // request need TX blobs, give them
		result.Tx_blob_list = append(result.Tx_blob_list, hex.EncodeToString(tx.SerializeHeader()))
	}
	//
	//extract secret key and feed it in here
	if p.Get_tx_key {
		result.Tx_key_list = append(result.Tx_key_list, h.r.w.GetTXKey(tx.GetHash())) //TODO
	}

	return result, nil
}

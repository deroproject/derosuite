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
import "net/http"
import "encoding/hex"
import "encoding/json"

import "github.com/romana/rlog"

import "github.com/deroproject/derosuite/structures"
import "github.com/deroproject/derosuite/transaction"

// we definitely need to clear up the MESS that has been created by the MONERO project
// half of their APIs are json rpc and half are http
// for compatibility reasons, we are implementing theirs ( however we are also providin a json rpc implementation)
// we should DISCARD the http

//  NOTE: we have currently not implemented the decode as json parameter
//  it is however on the pending list

//type	SendRawTransaction_Handler struct{}

func SendRawTransaction_Handler(rw http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var p structures.SendRawTransaction_Params
	var result structures.SendRawTransaction_Result
	var tx transaction.Transaction

	defer func() { // provide result back whatever it is
		encoder := json.NewEncoder(rw)
		encoder.Encode(result)
	}()
	err := decoder.Decode(&p)
	if err != nil {
		logger.Warnf("err while decoding incoming sendrawtransaaction json err: %s", err)
		return
	}
	defer req.Body.Close()

	rlog.Debugf("Incoming TX from RPC Server")

	//lets decode the tx from hex
	tx_bytes, err := hex.DecodeString(p.Tx_as_hex)

	if err != nil {
		result.Status = "TX could be hex decoded"
		return
	}
	if len(tx_bytes) < 100 {
		result.Status = "TX insufficient length"
		return
	}
	// lets add tx to pool, if we can do it, so  can every one else
	err = tx.DeserializeHeader(tx_bytes)
	if err != nil {
		rlog.Debugf("Incoming TX from RPC Server could NOT be deserialized")
		result.Status = err.Error()
		return
	}

	rlog.Infof("Incoming TXID %s from RPC Server", tx.GetHash())
	// lets try to add it to pool
	success := chain.Add_TX_To_Pool(&tx)

	if success {
		result.Status = "OK"
		rlog.Infof("Incoming TXID %s from RPC Server successfully accepted by MEMPOOL", tx.GetHash())
	} else {

		result.Status = fmt.Sprintf("Transaction %s rejected by daemon, check daemon msgs", tx.GetHash())
		rlog.Warnf("Incoming TXID %s from RPC Server rejected by MEMPOOL", tx.GetHash())
	}

}

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

import "net/http"
import "strings"

//import "encoding/hex"
import "encoding/json"

import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/structures"

// we definitely need to clear up the MESS that has been created by the MONERO project
// half of their APIs are json rpc and half are http
// for compatibility reasons, we are implementing theirs ( however we are also providin a json rpc implementation)
// we should DISCARD the http

//  NOTE: we have currently not implemented the decode as json parameter
//  it is however on the pending list

type IsKeyImageSpent_Handler struct{}

func iskeyimagespent(rw http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var p structures.Is_Key_Image_Spent_Params
	var result structures.Is_Key_Image_Spent_Result
	err := decoder.Decode(&p)
	if err != nil {
		panic(err)
	}
	defer req.Body.Close()

	for i := range p.Key_images {
		hash := crypto.HashHexToHash(strings.TrimSpace(p.Key_images[i]))

		// check in blockchain
		if _, ok := chain.Read_KeyImage_Status(nil, hash); ok {
			result.Spent_Status = append(result.Spent_Status, 1) // 1 mark means spent  in blockchain
			continue
		}

		// check in pool
		if chain.Mempool.Mempool_Keyimage_Spent(hash) {
			result.Spent_Status = append(result.Spent_Status, 2) // 2 mark means spent  in pool
			continue
		}

		result.Spent_Status = append(result.Spent_Status, 0) // 0 mark means unspent
	}

	result.Status = "OK"
	//logger.Debugf("Request %+v", p)

	encoder := json.NewEncoder(rw)
	encoder.Encode(result)
}

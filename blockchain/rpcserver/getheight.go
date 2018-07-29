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
import "encoding/json"

import "github.com/deroproject/derosuite/structures"

// we definitely need to clear up the MESS that has been created by the MONERO project
// half of their APIs are json rpc and half are http
// for compatibility reasons, we are implementing theirs ( however we are also providin a json rpc implementation)
// we should DISCARD the http

func getheight(rw http.ResponseWriter, req *http.Request) {

	var result structures.Daemon_GetHeight_Result
	defer req.Body.Close()

	result.Height = uint64(chain.Get_Height())
	result.StableHeight = chain.Get_Stable_Height()
	result.TopoHeight = chain.Load_TOPO_HEIGHT(nil)
	result.Status = "OK"

	encoder := json.NewEncoder(rw)
	encoder.Encode(result)
}

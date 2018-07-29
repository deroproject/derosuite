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

package p2p

//import "fmt"
//import "net"
import "sync/atomic"
import "time"

//import "container/list"
import "github.com/romana/rlog"
import "github.com/vmihailenco/msgpack"

//import "github.com/deroproject/derosuite/crypto"
//import "github.com/deroproject/derosuite/globals"

// reads our data, length prefix blocks
func (connection *Connection) Send_TimedSync(request bool) {

	var sync Sync_Struct

	fill_common(&sync.Common) // fill common info
	sync.Command = V2_COMMAND_SYNC
	sync.Request = request

	serialized, err := msgpack.Marshal(&sync) // serialize and send
	if err != nil {
		panic(err)
	}
	if request { // queue command that we are expecting a response
		//connection.Lock()
		connection.request_time.Store(time.Now())
		//connection.Unlock()
	}
	//rlog.Tracef(2, "Timed sync sent successfully %s", connection.logid)
	connection.Send_Message(serialized)

}

// handles  incoming timed syncs
func (connection *Connection) Handle_TimedSync(buf []byte) {
	var sync Sync_Struct
	err := msgpack.Unmarshal(buf, &sync)
	if err != nil {
		rlog.Warnf("Error while decoding incoming chain request err %s %s", err, connection.logid)
		connection.Exit()
		return
	}
	//rlog.Tracef(2, "Timed sync received %s", connection.logid)
	if sync.Request {
		connection.Send_TimedSync(false) // send it as response
	} else { // this is a response for our request track latency
		//connection.Lock()
		atomic.StoreInt64(&connection.Latency, int64(time.Now().Sub(connection.request_time.Load().(time.Time))/2)) // divide by 2 is for round-trip
		//connection.Unlock()

	}
}

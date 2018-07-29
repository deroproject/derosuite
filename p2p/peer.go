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

//import "net"
//import "sync"
//import "time"
import "crypto/rand"
import "encoding/binary"

//import "path/filepath"
//import "container/list"

//import log "github.com/sirupsen/logrus"
//import "github.com/vmihailenco/msgpack"

//import "github.com/deroproject/derosuite/crypto"
//import "github.com/deroproject/derosuite/globals"

//import "github.com/deroproject/derosuite/blockchain"

// This file defines  what all needs to be responded to become a server ( handling incoming requests)

var peerid uint64
var node_tag string

// get peer id
// we make a peer id randomly at every program start
// first call to this will give you a unique peer id
func GetPeerID() uint64 {
	if peerid == 0 {
		var buf [8]byte
		rand.Read(buf[:])
		peerid = binary.LittleEndian.Uint64(buf[:])
	}
	return peerid
}

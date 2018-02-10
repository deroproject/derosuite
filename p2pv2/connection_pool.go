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

package p2pv2

/* this file implements the connection pool manager, keeping a list of active connections etc
 * this will also ensure that a single IP is connected only once
 *
 */
import "fmt"
import "net"
import "sync"
import "container/list"

import log "github.com/sirupsen/logrus"

import "github.com/deroproject/derosuite/crypto"

// any connection incoming/outgoing can only be in this state
type Conn_State string

const (
	HANDSHAKE_PENDING Conn_State = "Pending"
	IDLE                         = "Idle"
	ACTIVE                       = "Active"
)

// This structure is used to do book keeping for the connection and keeps other DATA related to peer
type Connection struct {
	Incoming bool         // is connection incoming or outgoing
	Addr     *net.TCPAddr // endpoint on the other end
	Port     uint32       // port advertised by other end as its server,if it's 0 server cannot accept connections

	Peer_Info // all peer info parameters are present here

	Peer_ID               uint64      // Remote peer id
	Last_Height           uint64      // last height sent by peer
	Top_Version           uint64      // current hard fork version supported by peer
	Exit                  bool        // Exit marker that connection needs to be killed
	State                 Conn_State  // state of the connection
	Top_ID                crypto.Hash // top block id of the connection
	Cumulative_Difficulty uint64      // cumulative difficulty of top block of peer
	logger                *log.Entry  // connection specific logger
	Requested_Objects     [][32]byte  // currently unused as we sync up with a single peer at a time
	Conn                  net.Conn    // actual object to talk
	Command_queue         *list.List  // New protocol is partly syncronous/partly asyncronous
	sync.Mutex

	HandShakeCompleted bool // whether handshake is completed

	Bytes_Sent     uint64 // total bytes sent
	Bytes_Received uint64 // total bytes received

	// TODO a bloom filter that an object has been relayed
}

var connection_map = map[string]*Connection{}
var connection_mutex sync.Mutex

func Key(ip net.IP) string {
	return string(ip.To16()) // Simple []byte => string conversion
}

// check whether an IP is in the map already
func IsConnected(ip net.IP) bool {
	connection_mutex.Lock()
	defer connection_mutex.Unlock()

	if _, ok := connection_map[Key(ip)]; ok {
		return true
	}
	return false
}

// add connection to  map
func Connection_Add(c *Connection) {
	connection_mutex.Lock()
	defer connection_mutex.Unlock()
	connection_map[Key(c.Addr.IP)] = c
}

// add connection to  map
func Connection_Delete(c *Connection) {
	connection_mutex.Lock()
	defer connection_mutex.Unlock()
	delete(connection_map, Key(c.Addr.IP))
}

// prints all the connection info to screen
func Connection_Print() {
	connection_mutex.Lock()
	defer connection_mutex.Unlock()
	fmt.Printf("Connection info for peers\n")
	fmt.Printf("%-20s %-16s %-5s %-7s %9s %3s\n", "Remote Addr", "PEER ID", "PORT", " State", "Height", "DIR")
	for _, v := range connection_map {
		dir := "OUT"
		if v.Incoming {
			dir = "INC"
		}
		fmt.Printf("%-20s %16x %5d %7s %9d %s\n", v.Addr.IP, v.Peer_ID, v.Port, v.State, v.Last_Height, dir)
	}

}

// for continuos update on command line, get the maximum height of all peers
func Best_Peer_Height() (best_height uint64) {
	connection_mutex.Lock()
	for _, v := range connection_map {
		if v.Last_Height > best_height {
			best_height = v.Last_Height
		}
	}
	connection_mutex.Unlock()
	return
}

// this function return peer count which have successful handshake
func Peer_Count() (Count uint64) {
	connection_mutex.Lock()
	for _, v := range connection_map {
		if v.State != HANDSHAKE_PENDING {
			Count++
		}
	}
	connection_mutex.Unlock()
	return
}

// this returns count of peers in both directions
func Peer_Direction_Count() (Incoming uint64, Outgoing uint64) {
	connection_mutex.Lock()
	for _, v := range connection_map {
		if v.State != HANDSHAKE_PENDING {
			if v.Incoming {
				Incoming++
			} else {
				Outgoing++
			}
		}
	}
	connection_mutex.Unlock()
	return
}

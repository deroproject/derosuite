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

import "net"
import "encoding/binary"

// This file defines the structure for the protocol which is a msgp encoded ( which is standard)
// msgp  would cause an easy rewrite of p2p layer even in c, ruby or rust etc as future may demand
// the protocol is length prefixed msgp payload
// though we can  use http2 stream features,  they may become compilcated as the project evolve
// the prefix length is 4 bytes, little endian encoded ( so a frame can be 4GB in size)
// this is Work-In-Progress
// the reason for writing it from scratch is the mess of boost serialisation
// the p2p package is currently the most complex within the entire project

// the protocol is partly syncronous, partly asyncronous , except for first handshake, so the node remain undetectable to external network scans, the detection cost is to atleast send a handshake packet*/

// these are the commands required to make it completely operational
const V2_COMMAND_HANDSHAKE = 1 // commands are syncronous and must be responded within 10 secs
const V2_COMMAND_SYNC = 2
const V2_COMMAND_CHAIN_REQUEST = 3
const V2_COMMAND_CHAIN_RESPONSE = 4
const V2_COMMAND_OBJECTS_REQUEST = 5
const V2_COMMAND_OBJECTS_RESPONSE = 6

const V2_NOTIFY_NEW_BLOCK = 0x80000000 // Notifications are asyncronous all notifications come here, such as new block, new txs
const V2_NOTIFY_NEW_TX = 0x80000001    // notify tx using this

// used to parse incoming packet for for command , so as a repective command command could be triggered
type Common struct {
	Command               uint64   `msgpack:"C"`
	Height                uint64   `msgpack:"H"`
	Cumulative_Difficulty uint64   `msgpack:"CD"`
	Top_ID                [32]byte `msgpack:"TI"` // 32 bytes of Top block
	Top_Version           uint64   `msgpack:"TV"` // this basically represents the hard fork version
}

// at start, client sends handshake and server will respond to handshake
type Handshake struct {
	Common                     // add all fields of Common
	Local_Time     int64       `msgpack:"LT"`
	Local_Port     uint32      `msgpack:"LP"`
	PeerID         uint64      `msgpack:"PID"`
	Network_ID     [16]byte    `msgpack:"NID"` // 16 bytes
	PeerList       []Peer_Info `msgpack:"PLIST"`
	Extension_List []string    `msgpack:"EXT"`
	Request        bool        `msgpack:"REQUEST"` //whether this is a request
}

type Peer_Info struct {
	IP       net.IP `msgpack:"IP"`
	Port     uint32 `msgpack:"P"`
	ID       uint64 `msgpack:"I"`
	LastSeen uint64 `msgpack:"LS"`
}

type Sync struct {
	Common // add all fields of common
}

type Chain_Request struct {
	Block_list [][32]byte `msgpack:"BLIST"`
}

type Chain_Response struct {
	Start_height uint64     `msgpack:"SH"`
	Block_list   [][32]byte `msgpack:"BLIST"`
}

type Object_Request struct {
	Block_list [][32]byte `msgpack:"BLIST"`
	Tx_list    [][32]byte `msgpack:"TXLIST"`
}

type Complete_Block struct {
	Block []byte   `msgpack:"BLOCK"`
	Txs   [][]byte `msgpack:"TXS"`
}

type Object_Response struct {
	Blocks []Complete_Block `msgpack:"CBLOCKS"`
	Txs    [][]byte         `msgpack:"TXS"`
}

type Notify_New_Objects struct {
	Block Complete_Block `msgpack:"CBLOCK"`
	Txs   [][]byte       `msgpack:"TXS"`
}

// each packet has to be parsed twice once for extracting command and then a full parsing

// the message is sent as follows
// assumingin peer lock has already been taken
func (conn *Connection) Send_Message(data_bytes []byte) {

	var length_bytes [4]byte
	binary.LittleEndian.PutUint32(length_bytes[:], uint32(len(data_bytes)))

	// send the length prefix
	conn.Conn.Write(length_bytes[:])
	conn.Conn.Write(data_bytes) // send the message itself

}

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

// This file defines the structure for the protocol which is a msgp encoded ( which is standard)
// msgp  would cause an easy rewrite of p2p layer even in c, ruby or rust etc as future may demand
// the protocol is length prefixed msgp payload
// though we can  use http2 stream features,  they may become compilcated as the project evolves
// the prefix length is 4 bytes, little endian encoded ( so a frame can be 4GB in size)
// this is Work-In-Progress
// the reason for writing it from scratch is the mess of boost serialisation
// the p2p package is currently the most complex within the entire project

// the protocol is partly syncronous, partly asyncronous , except for first handshake, so the node remain undetectable to external network scans, the detection cost is atleast send a handshake packet*/

// these are the commands required to make it completely operational

const V2_COMMAND_NULL = 0 // default value is zero  and is a null command
//const V2_COMMAND_NULL = 0
//first 40 are reserved for future use
const V2_COMMAND_HANDSHAKE = 41 // commands are syncronous and must be responded within 10 secs
const V2_COMMAND_SYNC = 42
const V2_COMMAND_CHAIN_REQUEST = 43
const V2_COMMAND_CHAIN_RESPONSE = 44
const V2_COMMAND_OBJECTS_REQUEST = 45
const V2_COMMAND_OBJECTS_RESPONSE = 46

const V2_NOTIFY_NEW_BLOCK = 0xff // Notifications are asyncronous all notifications come here, such as new block, new txs
const V2_NOTIFY_NEW_TX = 0xfe    // notify tx using this

// used to parse incoming packet for for command , so as a repective command command could be triggered
type Common_Struct struct {
	Height                int64  `msgpack:"HEIGHT"`
	TopoHeight            int64  `msgpack:"THEIGHT"`
	StableHeight          int64  `msgpack:"SHEIGHT"`
	Cumulative_Difficulty string `msgpack:"CDIFF"`
	//	Top_ID                [32]byte `msgpack:"TOP"` // 32 bytes of Top block
	Top_Version uint64 `msgpack:"HF"` // this basically represents the hard fork version
}

const FLAG_LOWCPURAM string = "LOWCPURAM"

// at start, client sends handshake and server will respond to handshake
type Handshake_Struct struct {
	Command         uint64        `msgpack:"COMMAND"`
	Common          Common_Struct `msgpack:"COMMON"`   // add all fields of Common
	ProtocolVersion string        `msgpack:"PVERSION"` // version is a sematic version string semver
	Tag             string        `msgpack:"TAG"`      // user specific tag
	DaemonVersion   string        `msgpack:"DVERSION"`
	UTC_Time        int64         `msgpack:"UTC"`
	Local_Port      uint32        `msgpack:"LP"`
	Peer_ID         uint64        `msgpack:"PID"`
	Network_ID      [16]byte      `msgpack:"NID"` // 16 bytes
	Flags           []string      `msgpack:"FLAGS"`
	PeerList        []Peer_Info   `msgpack:"PLIST"`
	Extension_List  []string      `msgpack:"EXT"`
	Request         bool          `msgpack:"REQUEST"` //whether this is a request
}

type Peer_Info struct {
	Addr  string `msgpack:"ADDR"` // ip:port pair
	Miner bool   `msgpack:"MINER"`
	//ID       uint64 `msgpack:"I"`
	//LastSeen uint64 `msgpack:"LS"`
}

type Sync_Struct struct { // sync packets are sent every 2 seconds
	Command  uint64        `msgpack:"COMMAND"`
	Common   Common_Struct `msgpack:"COMMON"`  // add all fields of Common
	PeerList []Peer_Info   `msgpack:"PLIST"`   // update peer list
	Request  bool          `msgpack:"REQUEST"` //whether this is a request
}

type Chain_Request_Struct struct { // our version of chain
	Command     uint64        `msgpack:"COMMAND"`
	Common      Common_Struct `msgpack:"COMMON"` // add all fields of Common
	Block_list  [][32]byte    `msgpack:"BLIST"`  // block list
	TopoHeights []int64       `msgpack:"TOPO"`   // topo heights of added blocks
}

type Chain_Response_Struct struct { // peers gives us point where to get the chain
	Command      uint64        `msgpack:"COMMAND"`
	Common       Common_Struct `msgpack:"COMMON"` // add all fields of Common
	Start_height int64         `msgpack:"SH"`
	Start_topoheight int64     `msgpack:"STH"`
	Block_list   [][32]byte    `msgpack:"BLIST"`
	TopBlocks    [][32]byte    `msgpack:"TOPBLOCKS"` // top blocks used for faster syncronisation of alt-tips
	// this contains all blocks hashes for the last 10 heights, heightwise ordered

}

type Object_Request_Struct struct {
	Command    uint64        `msgpack:"COMMAND"`
	Common     Common_Struct `msgpack:"COMMON"` // add all fields of Common
	Block_list [][32]byte    `msgpack:"BLIST"`
	Tx_list    [][32]byte    `msgpack:"TXLIST"`
}

type Object_Response_struct struct {
	Command uint64           `msgpack:"COMMAND"`
	Common  Common_Struct    `msgpack:"COMMON"` // add all fields of Common
	CBlocks []Complete_Block `msgpack:"CBLOCKS"`
	Txs     [][]byte         `msgpack:"TXS"`
}

type Complete_Block struct {
	Block []byte   `msgpack:"BLOCK"`
	Txs   [][]byte `msgpack:"TXS"`
}

type Notify_New_Objects_Struct struct {
	Command uint64         `msgpack:"COMMAND"`
	Common  Common_Struct  `msgpack:"COMMON"` // add all fields of Common
	CBlock  Complete_Block `msgpack:"CBLOCK"`
	Tx      []byte         `msgpack:"TX"`
}

// each packet has to be parsed twice once for extracting command and then a full parsing based on Command

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

//import "net"
//import "sync"
import "time"

//import "container/list"

//import log "github.com/sirupsen/logrus"
import "github.com/vmihailenco/msgpack"

//import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/globals"

//import "github.com/deroproject/derosuite/blockchain"

// This file defines  what all needs to be responded to become a server ( handling incoming requests)

// fill the common part from our chain
func fill_common(common *Common) {
	common.Height = chain.Get_Height() - 1
	common.Top_ID, _ = chain.Load_BL_ID_at_Height(common.Height - 1)

	common.Cumulative_Difficulty = chain.Load_Block_Cumulative_Difficulty(common.Top_ID)
	common.Top_Version = 6 // this must be taken from the hardfork

}

// send the handshake command
// when we initiate a connection, we immediately send a handshake
func (conn *Connection) Send_Handshake_Command() {
	var h Handshake
	fill_common(&h.Common) // fill common information

	conn.Lock()
	defer conn.Unlock()

	// fill other information

	h.Local_Time = time.Now().UTC().Unix()
	h.Local_Port = 0
	h.PeerID = 0
	h.Network_ID = globals.Config.Network_ID
	//PeerList       []Peer_Info `msgpack:"PLIST"`
	//Extension_List []string `msgpack:"EXT"`

	//serialize and send
	b, err := msgpack.Marshal(h)
	if err != nil {
		panic(err)
	}

	conn.Send_Message(b) // send the message to peer

}

// a handshake command has been received, make the most of it
func (conn *Connection) Handle_Handshake_Command(h *Handshake) {

	conn.Lock()
	defer conn.Unlock()

	if h.Network_ID != globals.Config.Network_ID {
		logger.Debugf("Connection represents different network %x rejecting peer err:%s\n", h.Network_ID)
		conn.Exit = true
	}

	// TODO check whether the peer represents  current hardfork, if not reject

	// TODO check if the peer time is +- 2hrs from the our time , reject him ,

	// fill other information

	h.Local_Time = time.Now().UTC().Unix()
	h.Local_Port = 0
	h.PeerID = 0
	h.Network_ID = globals.Config.Network_ID
	//PeerList       []Peer_Info `msgpack:"PLIST"`
	//Extension_List []string `msgpack:"EXT"`

	//serialize and send
	b, err := msgpack.Marshal(h)
	if err != nil {
		panic(err)
	}

	conn.Send_Message(b) // send the message to peer

}

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

import "fmt"
import "bytes"

//import "net"
import "sync/atomic"
import "time"

//import "container/list"

//import log "github.com/sirupsen/logrus"
//import "github.com/allegro/bigcache"
import "github.com/romana/rlog"
import "github.com/vmihailenco/msgpack"

import "github.com/deroproject/derosuite/config"
import "github.com/deroproject/derosuite/globals"

//import "github.com/deroproject/derosuite/blockchain"

// reads our data, length prefix blocks
func (connection *Connection) Send_Handshake(request bool) {

	var handshake Handshake_Struct

	fill_common(&handshake.Common) // fill common info
	handshake.Command = V2_COMMAND_HANDSHAKE
	handshake.Request = request

	// TODO these version strings should be setup during build
	// original protocol in c daemon should be called version 1
	// the new version is version 2
	handshake.ProtocolVersion = "1.0.0"
	handshake.DaemonVersion = config.Version.String()
	handshake.Tag = node_tag
	handshake.UTC_Time = int64(time.Now().UTC().Unix()) // send our UTC time
	handshake.Local_Port = uint32(P2P_Port)             // export requested or default port
	handshake.Peer_ID = GetPeerID()                     // give our randomly generated peer id
	if globals.Arguments["--lowcpuram"].(bool) == false {
		handshake.Flags = append(handshake.Flags, FLAG_LOWCPURAM) // add low cpu ram flag
	}

	//scan our peer list and send peers which have been recently communicated
	handshake.PeerList = get_peer_list()
	copy(handshake.Network_ID[:], globals.Config.Network_ID[:])

	// serialize and send
	serialized, err := msgpack.Marshal(&handshake)
	if err != nil {
		panic(err)
	}

	rlog.Tracef(2, "handshake sent  %s", globals.CTXString(connection.logger))
	connection.Send_Message(serialized)
}

// verify incoming handshake for number of checks such as mainnet/testnet etc etc
func (connection *Connection) Verify_Handshake(handshake *Handshake_Struct) bool {
	return bytes.Equal(handshake.Network_ID[:], globals.Config.Network_ID[:])
}

// handles both server and client connections
func (connection *Connection) Handle_Handshake(buf []byte) {

	var handshake Handshake_Struct

	err := msgpack.Unmarshal(buf, &handshake)
	if err != nil {
		rlog.Warnf("Error while decoding incoming handshake request err %s %s", err, globals.CTXString(connection.logger))
		connection.Exit()
		return
	}

	if !connection.Verify_Handshake(&handshake) { // if not same network boot off
		connection.logger.Debugf("kill connection network id mismatch peer network id %x", handshake.Network_ID)
		connection.Exit()
		return
	}

	rlog.Tracef(2, "handshake response received %+v  %s", handshake, globals.CTXString(connection.logger))
        
        // check if self connection exit
        if connection.Incoming && handshake.Peer_ID == GetPeerID(){
            rlog.Tracef(1,"Same peer ID, probably self connection, disconnecting from this client")
            connection.Exit()
            return
        }
        

	if handshake.Request {
		connection.Send_Handshake(false) // send it as response
	}
	if !connection.Incoming { // setup success
		Peer_SetSuccess(connection.Addr.String())
	}

	connection.Update(&handshake.Common) // update common information

	if atomic.LoadUint32(&connection.State) == HANDSHAKE_PENDING { // some of the fields are processed only while initial handshake
		connection.Lock()
		if len(handshake.ProtocolVersion) < 128 {
			connection.ProtocolVersion = handshake.ProtocolVersion
		}

		if len(handshake.DaemonVersion) < 128 {
			connection.DaemonVersion = handshake.DaemonVersion
		}
		connection.Port = handshake.Local_Port
		connection.Peer_ID = handshake.Peer_ID
		if len(handshake.Tag) < 128 {
			connection.Tag = handshake.Tag
		}

		// TODO we must also add the peer to our list
		// which can be distributed to other peers
		if connection.Port != 0 && connection.Port <= 65535 { // peer is saying it has an open port, handshake is success so add peer

			var p Peer
			if connection.Addr.IP.To4() != nil { // if ipv4
                            p.Address = fmt.Sprintf("%s:%d", connection.Addr.IP.String(), connection.Port)
                        }else{ // if ipv6
                            p.Address = fmt.Sprintf("[%s]:%d", connection.Addr.IP.String(), connection.Port)
                        }
			p.ID = connection.Peer_ID

			p.LastConnected = 0 // uint64(time.Now().UTC().Unix())

			/* TODO we should add any flags here if necessary, but they are not
			   required, since a peer can only be used if connected and if connected
			   we already have a truly synced view
			 for _, k := range handshake.Flags {
				switch k {
				case FLAG_MINER:
					p.Miner = true
				}
			}*/

			Peer_Add(&p)
		}

		for _, k := range handshake.Flags {
			switch k {
			case FLAG_LOWCPURAM:
				connection.Lowcpuram = true

				//connection.logger.Debugf("Miner flag \"%s\" from peer", k)
			default:
				connection.logger.Debugf("Unknown flag \"%s\" from peer, ignoring", k)

			}

		}

		// do NOT build TX cache, if we are runnin in lowcpu mode
		if globals.Arguments["--lowcpuram"].(bool) == true { // if connection is not running in low cpu mode and we are also same, activate transaction cache
			connection.TXpool_cache = nil
		} else { // we do not have any limitation, activate per peer cache
			connection.TXpool_cache = map[uint64]uint32{}

		}
		connection.Unlock()
	}

	// parse delivered peer list as grey list
	rlog.Debugf("Peer provides %d peers", len(handshake.PeerList))
	for i := range handshake.PeerList {
		Peer_Add(&Peer{Address: handshake.PeerList[i].Addr})
	}

	atomic.StoreUint32(&connection.State, ACTIVE)
	if connection.Incoming {
		Connection_Add(connection)
	}
}

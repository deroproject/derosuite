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

import "io"
import "net"

//import "sync"
import "time"
import "runtime/debug"
import "container/list"
import "encoding/binary"

import log "github.com/sirupsen/logrus"
import "github.com/romana/rlog"
import "github.com/vmihailenco/msgpack"

// this function waits for any commands from the connection and suitably responds doing sanity checks
// this is the core of the p2p package
/* this is the entire connection handler, all incoming/outgoing connections end up here  */
func Handle_Connection(conn net.Conn, remote_addr *net.TCPAddr, incoming bool) {

	var connection Connection
	connection.Incoming = incoming
	connection.Conn = conn
	var idle int

	connection.Addr = remote_addr         //  since we may be connecting via socks, get target IP
	connection.Command_queue = list.New() // init command queue
	connection.State = HANDSHAKE_PENDING
	if incoming {
		connection.logger = logger.WithFields(log.Fields{"RIP": remote_addr.String(), "DIR": "INC", "V2": "1"})
	} else {
		connection.logger = logger.WithFields(log.Fields{"RIP": remote_addr.String(), "DIR": "OUT", "V2": "1"})
	}

	defer func() {
		if r := recover(); r != nil {
			connection.logger.Warnf("Recovered while handling connection, Stack trace below", r)
			connection.logger.Warnf("Stack trace  \n%s", debug.Stack())

		}
	}()

	Connection_Add(&connection) // add connection to pool
	if !incoming {              // we initiated the connection, we must send the handshake first
		connection.Send_Handshake_Command() // send handshake
	}

	// goroutine to exit the connection if signalled
	go func() {
		ticker := time.NewTicker(1 * time.Second) // 1 second ticker
		for {
			select {
			case <-ticker.C:
				idle++
				// if idle more than 13 secs, we should send a timed sync
				if idle > 13 {
					if connection.State != HANDSHAKE_PENDING {
						connection.State = IDLE
					}
					//Send_Timed_Sync(&connection)
					//connection.logger.Debugf("We should send a timed sync")
					idle = 0
				}
			case <-Exit_Event: // p2p is shutting down, close the connection
				connection.Exit = true
				ticker.Stop() // release resources of timer
				Connection_Delete(&connection)
				conn.Close()
				return // close the connection and close the routine
			}
			if connection.Exit { // release resources of timer
				ticker.Stop()
				Connection_Delete(&connection)
				conn.Close()
				return
			}
		}
	}()

	// the infinite loop handler
	for {
		length_buf := make([]byte, 4, 4) // prefix length is 4 bytes
		if connection.Exit {
			break
		}

		read_count, err := io.ReadFull(connection.Conn, length_buf)
		if err != nil {
			rlog.Tracef(2, "Error while reading command prefix length exiting err:%s\n", err)
			connection.Exit = true
			continue
		}

		length := binary.LittleEndian.Uint32(length_buf) // convert little endian bytes 4 bytes to length

		// check safety of length, we should not allocate more than 100 MB as that is the limit of the block

		command_buf := make([]byte, length, length)
		//set_timeout(&connection) // we should not hang for hrs waiting for data to come
		read_count, err = io.ReadFull(connection.Conn, command_buf)
		if err != nil {
			rlog.Tracef(2, "Error while reading command data exiting err:%s\n", err)
			connection.Exit = true
			continue
		}

		command_buf = command_buf[:read_count]
		var dummy_command Common
		err = msgpack.Unmarshal(command_buf, &dummy_command)
		if err != nil {
			rlog.Tracef(2, "Error while parsing command data exiting err:%s\n", err)
			connection.Exit = true
			continue
		}

		// if handshake not done, donot process any command

		if !connection.HandShakeCompleted && dummy_command.Command != V2_COMMAND_HANDSHAKE {
			rlog.Tracef(2, "Peer Sending something but we are waiting for handshake command data exiting err:%s\n", err)
			connection.Exit = true
			continue
		}

		switch dummy_command.Command {
		case V2_COMMAND_HANDSHAKE:
			var handshake Handshake
			err = msgpack.Unmarshal(command_buf, &handshake)
			if err != nil {
				rlog.Tracef(2, "Error while parsing incoming handshake data exiting err:%s\n", err)
				connection.Exit = true
				continue
			}

		case V2_COMMAND_SYNC:

		case V2_COMMAND_CHAIN_REQUEST:

		case V2_COMMAND_CHAIN_RESPONSE: // this should be verified whether we are waiting for it

		case V2_COMMAND_OBJECTS_REQUEST:

		case V2_COMMAND_OBJECTS_RESPONSE: // this should be verified whether we are waiting for it

		case V2_NOTIFY_NEW_BLOCK:

		case V2_NOTIFY_NEW_TX:

		}

	}

}

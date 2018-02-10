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

import "io"
import "net"
import "fmt"
import "time"
import "testing"
import "container/list"
import "encoding/binary"
import "runtime/debug"

import "github.com/romana/rlog"
import log "github.com/sirupsen/logrus"

import "github.com/deroproject/derosuite/globals"

// all communications flow in little endian
const LEVIN_SIGNATURE = 0x0101010101012101 //Bender's nightmare
const LEVIN_SIGNATURE_DATA = 0x0102010101011101
const LEVIN_PROTOCOL_VER_0 = 0
const LEVIN_PROTOCOL_VER_1 = 1

const LEVIN_PACKET_REQUEST = 0x00000001
const LEVIN_PACKET_RESPONSE = 0x00000002

// the whole structure should be packed into 33 bytes
type Levin_Header struct {
	Signature        uint64
	CB               uint64 // this contains data size appended to buffer
	ReturnData       bool
	Command          uint32
	ReturnCode       int32
	Flags            uint32
	Protocol_Version uint32
}

// all response will have the signature in big endian form
type Levin_Data_Header struct {
	Signature uint64 // LEVIN_SIGNATURE_DATA
	//Boost_Header byte
	Data []byte
}

// sets  timeout based on connection state, so as stale connections are cleared quickly
func set_timeout(connection *Connection) {
	if connection.State == HANDSHAKE_PENDING {
		connection.Conn.SetReadDeadline(time.Now().Add(20 * time.Second)) // new connections have 20 seconds to handshake
	} else {
		connection.Conn.SetReadDeadline(time.Now().Add(300 * time.Second)) // good connections have 5 mins to communicate
	}
}

/* this is the entire connection handler, all incoming/outgoing connections end up here  */
func Handle_Connection(conn net.Conn, remote_addr *net.TCPAddr, incoming bool) {

	var connection Connection
	var levin_header Levin_Header
	connection.Incoming = incoming
	connection.Conn = conn
	var idle int

	connection.Addr = remote_addr         //  since we may be connecting via socks, get target IP
	connection.Command_queue = list.New() // init command queue
	connection.State = HANDSHAKE_PENDING
	if incoming {
		connection.logger = logger.WithFields(log.Fields{"RIP": remote_addr.String(), "DIR": "INC"})
	} else {
		connection.logger = logger.WithFields(log.Fields{"RIP": remote_addr.String(), "DIR": "OUT"})
	}

	defer func() {
		if r := recover(); r != nil {
			connection.logger.Warnf("Recovered while handling connection, Stack trace below", r)
			connection.logger.Warnf("Stack trace  \n%s", debug.Stack())

		}
	}()

	Connection_Add(&connection) // add connection to pool
	if !incoming {
		Send_Handshake(&connection) // send handshake
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
					Send_Timed_Sync(&connection)
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
				return
			}
		}
	}()

	for {
		if connection.Exit {
			connection.logger.Debugf("Connection exited")
			conn.Close()
			return
		}

		// wait and read header
		header_data := make([]byte, 33, 33) // size of levin header
		idle = 0
		rlog.Tracef(10, "waiting to read header bytes from network %s\n", globals.CTXString(connection.logger))
		set_timeout(&connection)
		read_bytes, err := io.ReadFull(conn, header_data)

		if err != nil {
			rlog.Tracef(4, "Error while reading levin header exiting err:%s\n", err)
			connection.Exit = true
			continue
		}
		rlog.Tracef(10, "Read %d bytes from network\n", read_bytes)

		if connection.State != HANDSHAKE_PENDING {
			connection.State = ACTIVE
		}

		err = levin_header.DeSerialize(header_data)

		if err != nil {
			rlog.Tracef(4, "Error while DeSerializing levin header exiting err:%s\n", err)
			connection.Exit = true
			continue
		}

		// read data as per requirement
		data := make([]byte, levin_header.CB, levin_header.CB)
		set_timeout(&connection)
		read_bytes, err = io.ReadFull(conn, data)

		rlog.Tracef(10, "Read %d bytes from network for data \n", read_bytes)
		if err != nil {
			rlog.Tracef(4, "Error while reading levin data exiting err:%s\n", err)
			connection.Exit = true
			continue
		}

		name := COMMAND_NAME[levin_header.Command]
		if name == "" {
			connection.logger.Warnf("No Such command %d exiting\n", levin_header.Command)
			connection.Exit = true
			continue
		}

		//connection.logger.WithFields(log.Fields{
		//	"command": name,
		//	"flags":   levin_header.Flags}).Debugf("Incoming Command")

		if levin_header.Flags == LEVIN_PACKET_RESPONSE {
			if connection.Command_queue.Len() < 1 {
				connection.logger.Warnf("Invalid Response ( we have not queued anything\n")
				connection.Exit = true
				continue
			}

			front_command := connection.Command_queue.Front()
			if levin_header.Command != front_command.Value.(uint32) {
				connection.logger.Warnf("Invalid Response (  we queued some other command\n")
				connection.Exit = true
				continue
			}

			connection.Lock()
			connection.Command_queue.Remove(front_command)
			connection.Unlock()

			switch levin_header.Command {
			case P2P_COMMAND_HANDSHAKE: // Parse incoming handshake response
				Handle_P2P_Handshake_Command_Response(&connection, &levin_header, data)
				// if response is OK, mark conncection as good and add it to list

			case P2P_COMMAND_TIMED_SYNC: // we never send timed response
				// connection.logger.Infof("Response for timed sync arrived")
				Handle_P2P_Timed_Sync_Response(&connection, &levin_header, data)

			case P2P_COMMAND_PING: // we never send ping packets

			case P2P_COMMAND_REQUEST_SUPPORT_FLAGS: // we never send flags packet

			}

		}

		if levin_header.Flags == LEVIN_PACKET_REQUEST {

			switch levin_header.Command {
			case P2P_COMMAND_HANDSHAKE: // send response
				connection.logger.Debugf("Incoming handshake command")
				Handle_P2P_Handshake_Command(&connection, &levin_header, data)

			case P2P_COMMAND_REQUEST_SUPPORT_FLAGS: // send reponse
				Handle_P2P_Support_Flags(&connection, &levin_header, data)

			case P2P_COMMAND_TIMED_SYNC:
				Handle_P2P_Timed_Sync(&connection, &levin_header, data)
				// crypto note core protocols commands related to blockchain
				// peer wants to syncronise his chain to ours
			case BC_NOTIFY_REQUEST_CHAIN:
				Handle_BC_Notify_Chain(&connection, &levin_header, data)
				// we want to syncronise our chain to peers
			case BC_NOTIFY_RESPONSE_CHAIN_ENTRY:
				Handle_BC_Notify_Response_Chain_Entry(&connection, &levin_header, data)

			case BC_NOTIFY_REQUEST_GET_OBJECTS: // peer requested some object
				Handle_BC_Notify_Request_GetObjects(&connection, &levin_header, data)
			case BC_NOTIFY_RESPONSE_GET_OBJECTS: // peer responded to our object requests
				Handle_BC_Notify_Response_GetObjects(&connection, &levin_header, data)
			case BC_NOTIFY_NEW_TRANSACTIONS:
				Handle_BC_Notify_New_Transactions(&connection, &levin_header, data)
			case BC_NOTIFY_NEW_BLOCK:
				Handle_BC_Notify_New_Block(&connection, &levin_header, data)

			}
		}
	}
}

/* this operation can never fail */
func SerializeLevinHeader(header Levin_Header) []byte {

	packed_buffer := make([]byte, 33, 33)

	binary.LittleEndian.PutUint64(packed_buffer[0:8], LEVIN_SIGNATURE) // packed 8 bytes
	binary.LittleEndian.PutUint64(packed_buffer[8:16], header.CB)      // packed 8 + 8 bytes
	if header.ReturnData {
		packed_buffer[16] = 1 // packed 8 + 8 + 1
	}

	binary.LittleEndian.PutUint32(packed_buffer[17:17+4], header.Command)            // packed 8+8+1+4 bytes
	binary.LittleEndian.PutUint32(packed_buffer[21:21+4], uint32(header.ReturnCode)) // packed 8+8+1+4 bytes
	binary.LittleEndian.PutUint32(packed_buffer[25:25+4], header.Flags)              // packed 8+8+1+4 bytes
	binary.LittleEndian.PutUint32(packed_buffer[29:29+4], LEVIN_PROTOCOL_VER_1)      // packed 8+8+1+4 bytes

	return packed_buffer
}

func (header Levin_Header) Serialize() ([]byte, int) {
	packed_buffer := make([]byte, 33, 33)

	binary.LittleEndian.PutUint64(packed_buffer[0:8], LEVIN_SIGNATURE) // packed 8 bytes
	binary.LittleEndian.PutUint64(packed_buffer[8:16], header.CB)      // packed 8 + 8 bytes
	if header.ReturnData {
		packed_buffer[16] = 1 // packed 8 + 8 + 1
	}

	binary.LittleEndian.PutUint32(packed_buffer[17:17+4], header.Command)            // packed 8+8+1+4 bytes
	binary.LittleEndian.PutUint32(packed_buffer[21:21+4], uint32(header.ReturnCode)) // packed 8+8+1+4 bytes
	binary.LittleEndian.PutUint32(packed_buffer[25:25+4], header.Flags)              // packed 8+8+1+4 bytes
	binary.LittleEndian.PutUint32(packed_buffer[29:29+4], LEVIN_PROTOCOL_VER_1)      // packed 8+8+1+4 bytes

	return packed_buffer, len(packed_buffer)

}

// extract structure info from hardcoded node
func (header *Levin_Header) DeSerialize(packed_buffer []byte) (err error) {

	if len(packed_buffer) != 33 {
		return fmt.Errorf("Insufficient header bytes")
	}

	header.Signature = binary.LittleEndian.Uint64(packed_buffer[0:8]) // packed 8 bytes

	if header.Signature != LEVIN_SIGNATURE {
		return fmt.Errorf("Incorrect Levin Signature")
	}
	header.CB = binary.LittleEndian.Uint64(packed_buffer[8:16]) // packed 8 + 8 bytes
	if packed_buffer[16] == 0 {
		header.ReturnData = false // packed 8 + 8 + 1
	} else {
		header.ReturnData = true // packed 8 + 8 + 1
	}

	header.Command = binary.LittleEndian.Uint32(packed_buffer[17 : 17+4])             // packed 8+8+1+4 bytes
	header.ReturnCode = (int32)(binary.LittleEndian.Uint32(packed_buffer[21 : 21+4])) // packed 8+8+1+4 bytes
	header.Flags = binary.LittleEndian.Uint32(packed_buffer[25 : 25+4])               // packed 8+8+1+4 bytes
	header.Protocol_Version = binary.LittleEndian.Uint32(packed_buffer[29 : 29+4])    // packed 8+8+1+4 bytes

	return nil

}

func (header Levin_Data_Header) Serialize() ([]byte, int) {

	var packed_buffer []byte
	// if nothing is to be placed

	if len(header.Data) == 0 {
		packed_buffer = make([]byte, 8+2, 8+2)                                  // 10 bytes minimum heade
		binary.LittleEndian.PutUint64(packed_buffer[0:8], LEVIN_SIGNATURE_DATA) // packed 8 bytes
		packed_buffer[8] = 1
		packed_buffer[9] = 0

		return packed_buffer, len(packed_buffer)

	}
	packed_buffer = make([]byte, 8+2+len(header.Data), 8+2+len(header.Data))

	binary.LittleEndian.PutUint64(packed_buffer[0:8], LEVIN_SIGNATURE_DATA) // packed 8 bytes
	packed_buffer[8] = 1
	packed_buffer[9] = 8
	copy(packed_buffer[10:], header.Data)

	return packed_buffer, len(packed_buffer)

}

// extract structure info from hardcoded node
func (header *Levin_Data_Header) DeSerialize(packed_buffer []byte) (err error) {

	if len(packed_buffer) < 10 {
		return fmt.Errorf("Insufficient header bytes")
	}

	header.Signature = binary.LittleEndian.Uint64(packed_buffer[0:8]) // packed 8 bytes

	if header.Signature != LEVIN_SIGNATURE_DATA {
		return fmt.Errorf("WRONG LEVIN_SIGNATURE_DATA")
	}

	if len(packed_buffer)-8 == 2 {
		return nil
	}
	header.Data = make([]byte, len(packed_buffer)-8+2, len(packed_buffer)-8+2)
	// ignore 2 bytes
	//	packed_buffer[8]=1 // version
	//	packed_buffer[9]=8 // boost 8 , this can be anything as per boost level
	copy(header.Data, packed_buffer[10:])

	return nil

}

func DeSerializeLevinHeader(packed_buffer []byte, header *Levin_Header) error {

	if len(packed_buffer) != 33 {
		return fmt.Errorf("Insufficient header bytes")
	}

	header.Signature = binary.LittleEndian.Uint64(packed_buffer[0:8]) // packed 8 bytes
	header.CB = binary.LittleEndian.Uint64(packed_buffer[8:16])       // packed 8 + 8 bytes
	if packed_buffer[16] == 0 {
		header.ReturnData = false // packed 8 + 8 + 1
	} else {
		header.ReturnData = true // packed 8 + 8 + 1
	}

	header.Command = binary.LittleEndian.Uint32(packed_buffer[17 : 17+4])             // packed 8+8+1+4 bytes
	header.ReturnCode = (int32)(binary.LittleEndian.Uint32(packed_buffer[21 : 21+4])) // packed 8+8+1+4 bytes
	header.Flags = binary.LittleEndian.Uint32(packed_buffer[25 : 25+4])               // packed 8+8+1+4 bytes
	header.Protocol_Version = binary.LittleEndian.Uint32(packed_buffer[29 : 29+4])    // packed 8+8+1+4 bytes

	return nil
}

func TestSerializeDeserialize(t *testing.T) {

}

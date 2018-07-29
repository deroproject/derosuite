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

//import "fmt"
import "net"

//import "sync"
//import "math/rand"
import "time"
import "math/big"
import "runtime/debug"
import "sync/atomic"
import "encoding/binary"

//import "container/list"

import "github.com/romana/rlog"
import log "github.com/sirupsen/logrus"
import "github.com/paulbellamy/ratecounter"
import "github.com/vmihailenco/msgpack"

import "github.com/deroproject/derosuite/config"

//import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/globals"

//import "github.com/deroproject/derosuite/blockchain"

// This file defines  what all needs to be responded to become a server ( handling incoming requests)

// fill the common part from our chain
func fill_common(common *Common_Struct) {
	common.Height = chain.Get_Height()
	//common.StableHeight = chain.Get_Stable_Height()
	common.TopoHeight = chain.Load_TOPO_HEIGHT(nil)
	//common.Top_ID, _ = chain.Load_BL_ID_at_Height(common.Height - 1)

	high_block, err := chain.Load_Block_Topological_order_at_index(nil, common.TopoHeight)
	if err != nil {
		common.Cumulative_Difficulty = "0"
	} else {
		common.Cumulative_Difficulty = chain.Load_Block_Cumulative_Difficulty(nil, high_block).String()
	}
	common.Top_Version = uint64(chain.Get_Current_Version_at_Height(int64(common.Height))) // this must be taken from the hardfork

}

// used while sendint TX ASAP
func fill_common_skip_topoheight(common *Common_Struct) {
	common.Height = chain.Get_Height()
	//common.StableHeight = chain.Get_Stable_Height()
	common.TopoHeight = chain.Load_TOPO_HEIGHT(nil)
	//common.Top_ID, _ = chain.Load_BL_ID_at_Height(common.Height - 1)

	high_block, err := chain.Load_Block_Topological_order_at_index(nil, common.TopoHeight)
	if err != nil {
		common.Cumulative_Difficulty = "0"
	} else {
		common.Cumulative_Difficulty = chain.Load_Block_Cumulative_Difficulty(nil, high_block).String()
	}
	common.Top_Version = uint64(chain.Get_Current_Version_at_Height(int64(common.Height))) // this must be taken from the hardfork

}

// update some common properties quickly
func (connection *Connection) Update(common *Common_Struct) {
	//connection.Lock()
	//defer connection.Unlock()
	atomic.StoreInt64(&connection.Height, common.Height)             // satify race detector GOD
	if  common.StableHeight != 0{
		atomic.StoreInt64(&connection.StableHeight, common.StableHeight) // satify race detector GOD
	}
	atomic.StoreInt64(&connection.TopoHeight, common.TopoHeight)     // satify race detector GOD

	//connection.Top_ID = common.Top_ID
	connection.Cumulative_Difficulty = common.Cumulative_Difficulty

	var x *big.Int
	x = new(big.Int)
	if _, ok := x.SetString(connection.Cumulative_Difficulty, 10); !ok { // if Cumulative_Difficulty could not be parsed, kill connection
		rlog.Warnf("Could not Parse Cumulative_Difficulty in common %s \"%s\" ", connection.Cumulative_Difficulty, globals.CTXString(connection.logger))

		connection.Exit()
	}

	connection.CDIFF.Store(x) // do it atomically

	if connection.Top_Version != common.Top_Version {
		atomic.StoreUint64(&connection.Top_Version, common.Top_Version) // satify race detector GOD
	}
}

// sets  timeout based on connection state, so as stale connections are cleared quickly
// if these timeouts are met connection is closed
func (connection *Connection) set_timeout() {
	if atomic.LoadUint32(&connection.State) == HANDSHAKE_PENDING {
		connection.Conn.SetReadDeadline(time.Now().Add(8 * time.Second)) // new connections have 8 seconds to handshake
	} else {
		// if we have queued something, time should be less than 30 sec assumming 10KB/sec BW is available
		connection.Conn.SetReadDeadline(time.Now().Add(40 * time.Second)) // good connections have 120 secs to communicate
	}

}

// marks connection as exit in progress
func (conn *Connection) Exit() {
	atomic.AddInt32(&conn.ExitCounter, 1)
}

func (conn *Connection) IsExitInProgress() bool {
	if atomic.AddInt32(&conn.ExitCounter, 0) != 0 {
		return true
	}
	return false
}

// the message is sent as follows

func (conn *Connection) Send_Message(data_bytes []byte) {

	// measure parameters
	conn.SpeedOut.Incr(int64(len(data_bytes)) + 4)
	atomic.StoreUint64(&conn.BytesOut, atomic.LoadUint64(&conn.BytesOut)+(uint64(len(data_bytes))+4))

	if atomic.LoadUint32(&conn.State) != HANDSHAKE_PENDING {
		atomic.StoreUint32(&conn.State, ACTIVE)
	}
	conn.Send_Message_prelocked(data_bytes)
}

// used when we have command queue
// assumingin peer lock has already been taken
func (conn *Connection) Send_Message_prelocked(data_bytes []byte) {

	conn.writelock.Lock()
	defer conn.writelock.Unlock()

	defer func() {
		if r := recover(); r != nil {
			conn.logger.Warnf("Recovered while handling connection, Stack trace below", r)
			conn.logger.Warnf("Stack trace  \n%s", debug.Stack())
			conn.Exit()
		}
	}()

	if conn.IsExitInProgress() {
		return
	}

	var length_bytes [4]byte
	binary.LittleEndian.PutUint32(length_bytes[:], uint32(len(data_bytes)))
        

	// each connection has a write deadline of 60 secs
    conn.Conn.SetWriteDeadline(time.Now().Add(60 * time.Second)) 
	if _, err := conn.Conn.Write(length_bytes[:]); err != nil { // send the length prefix
		conn.Exit()
		return
	}

	conn.Conn.SetWriteDeadline(time.Now().Add(60 * time.Second)) 
	if _, err := conn.Conn.Write(data_bytes[:]); err != nil { // send the message itself
		conn.Exit()
		return
	}

}

// reads our data, length prefix blocks
func (connection *Connection) Read_Data_Frame(timeout int64, max_block_size uint32) (data_buf []byte) {

	var frame_length_buf [4]byte

	connection.set_timeout()
	nbyte, err := io.ReadFull(connection.Conn, frame_length_buf[:])
	if err != nil || nbyte != 4 {
		// error while reading from connection we must disconnect it
		rlog.Warnf("Could not read length prefix err %s %s", err, globals.CTXString(connection.logger))

		connection.Exit()

		return
	}

	//  time to ban
	frame_length := binary.LittleEndian.Uint32(frame_length_buf[:])
	if frame_length == 0 || uint64(frame_length) > (15*config.CRYPTONOTE_MAX_BLOCK_SIZE/10) || uint64(frame_length) > (2*uint64(max_block_size)) {
		// most probably memory DDOS attack, kill the connection
		rlog.Warnf("Frame length is too big Expected %d Actual %d %s", (2 * uint64(max_block_size)), frame_length, globals.CTXString(connection.logger))
		connection.Exit()
		return

	}

	data_buf = make([]byte, frame_length)
	connection.set_timeout()
	data_size, err := io.ReadFull(connection.Conn, data_buf)
	if err != nil || data_size <= 0 || uint32(data_size) != frame_length {
		// error while reading from connection we must kiil it
		rlog.Warnf("Could not read data size  read %d, frame length %d err %s", data_size, frame_length, err, globals.CTXString(connection.logger))
		connection.Exit()
		return

	}
	data_buf = data_buf[:frame_length]

	// measure parameters
	connection.SpeedIn.Incr(int64(frame_length) + 4)
	connection.BytesIn += (uint64(frame_length) + 4)
	if atomic.LoadUint32(&connection.State) != HANDSHAKE_PENDING {
		atomic.StoreUint32(&connection.State, ACTIVE)
	}
	
	return data_buf

}

// handles both server and client connections
func Handle_Connection(conn net.Conn, remote_addr *net.TCPAddr, incoming bool, sync_node bool) {

	var err error

	defer func() {
		if r := recover(); r != nil { // under rare condition below defer can also raise an exception, catch it now
			// connection.logger.Warnf("Recovered while handling connection RARE", r)
			rlog.Warnf("Recovered while handling connection RARE")
			//conn.Exit()
		}
	}()

	var connection Connection
	connection.Incoming = incoming
	connection.Conn = conn
	connection.SyncNode = sync_node
	connection.Addr = remote_addr //  since we may be connecting via socks, get target IP
	//connection.Command_queue = list.New() // init command queue
	connection.Objects = make(chan Queued_Command, 2048)
	connection.CDIFF.Store(new(big.Int).SetUint64(1))
	connection.State = HANDSHAKE_PENDING

	connection.request_time.Store(time.Now())
	//connection.Exit = make(chan bool)
	connection.SpeedIn = ratecounter.NewRateCounter(60 * time.Second)
	connection.SpeedOut = ratecounter.NewRateCounter(60 * time.Second)

	if incoming {
		connection.logger = logger.WithFields(log.Fields{"RIP": remote_addr.String(), "DIR": "INC"})
	} else {
		connection.logger = logger.WithFields(log.Fields{"RIP": remote_addr.String(), "DIR": "OUT"})
	}
	// this is to kill most of the races, related to logger
	connection.logid = globals.CTXString(connection.logger)

	defer func() {
		if r := recover(); r != nil {
			connection.logger.Warnf("Recovered while handling connection, Stack trace below", r)
			connection.logger.Warnf("Stack trace  \n%s", debug.Stack())
			connection.Exit()
		}
	}()

	// goroutine to exit the connection if signalled
	go func() {

		defer func() {
			if r := recover(); r != nil { // under rare condition below defer can also raise an exception, catch it now
				// connection.logger.Warnf("Recovered while handling connection RARE", r)
				rlog.Warnf("Recovered while handling Timed Sync")
				connection.Exit()
			}
		}()

		ticker := time.NewTicker(1 * time.Second) // 1 second ticker
		idle := 0
		rehandshake := 0
		for {

			if connection.IsExitInProgress() {

				//connection.logger.Warnf("Removing connection")
				ticker.Stop() // release resources of timer
                conn.Close()
				Connection_Delete(&connection)
				return // close the connection and close the routine
			}

			select {
			case <-ticker.C:
				idle++
				rehandshake++
				if idle > 2 { // if idle more than 2 secs, we should send a timed sync
					if atomic.LoadUint32(&connection.State) != HANDSHAKE_PENDING {
						atomic.StoreUint32(&connection.State, IDLE)
						connection.Send_TimedSync(true)
						idle = 0
					}

					// if no timed sync response in 2 minute kill the connection
					if time.Now().Sub(connection.request_time.Load().(time.Time)) > 120*time.Second{
						connection.Exit()
					}

					//if !connection.Incoming { // there is not point in sending timed sync both sides simultaneously

					if rehandshake > 1800 { // rehandshake to exchange peers every half hour
						connection.Send_Handshake(true)
						rehandshake = 0

						// run cleanup process every 1800 seconds
						connection.TXpool_cache_lock.Lock()
						if connection.TXpool_cache != nil {
							current_time := uint32(time.Now().Unix())
							for k, v := range connection.TXpool_cache {
								if (v + 1800) < current_time {
									delete(connection.TXpool_cache, k)
								}
							}
						}
						connection.TXpool_cache_lock.Unlock()
					}
					//}
				}
			case <-Exit_Event:
				ticker.Stop() // release resources of timer
				conn.Close()
				Connection_Delete(&connection)
				return // close the connection and close the routine

			}
		}
	}()

	if !incoming {
		Connection_Add(&connection)     // add outgoing connection to pool, incoming are added when handshake are done
		connection.Send_Handshake(true) // send handshake request
	}

	for {

		var command Sync_Struct // decode as sync minimum

		//connection.logger.Info("Waiting for frame")
		if connection.IsExitInProgress() {
			return
		}

		//connection.logger.Infof("Waiting for data frame")
		// no one should be able to request more than necessary amount of buffer
		data_read := connection.Read_Data_Frame(0, uint32(config.CRYPTONOTE_MAX_BLOCK_SIZE*2))

		if connection.IsExitInProgress() {
			return
		}
		//connection.logger.Info("frame received")

		// connection.logger.Infof(" data frame arrived %d", len(data_read))

		// lets decode command and make sure we understand it
		err = msgpack.Unmarshal(data_read, &command)
		if err != nil {
			rlog.Warnf("Error while decoding incoming frame err %s %s", err, globals.CTXString(connection.logger))
			connection.Exit()
			return
		}

		// check version sanctity
		//connection.logger.Infof(" data frame parsed %+v", command)

		// till the time handshake is done, we donot process any commands
		if atomic.LoadUint32(&connection.State) == HANDSHAKE_PENDING && !(command.Command == V2_COMMAND_HANDSHAKE || command.Command == V2_COMMAND_SYNC) {
			// client sent something else when we were waiting for handshake, ban the peer
			rlog.Warnf("Terminating connection, we were waiting for handshake but received %d %s", command.Command, globals.CTXString(connection.logger))

			connection.Exit()
			return
		}

		//connection.logger.Debugf("v2 command incoming  %d", command.Command)

		switch command.Command {
		case V2_COMMAND_HANDSHAKE:
			connection.Update(&command.Common)
			connection.Handle_Handshake(data_read)

		case V2_COMMAND_SYNC:
			connection.Update(&command.Common)
			connection.Handle_TimedSync(data_read)
		case V2_COMMAND_CHAIN_REQUEST:
			connection.Update(&command.Common)
			connection.Handle_ChainRequest(data_read)
		case V2_COMMAND_CHAIN_RESPONSE:

			connection.Update(&command.Common)
			connection.Handle_ChainResponse(data_read)
		case V2_COMMAND_OBJECTS_REQUEST:
			connection.Update(&command.Common)
			connection.Handle_ObjectRequest(data_read)
		case V2_COMMAND_OBJECTS_RESPONSE:
			connection.Update(&command.Common)
			connection.Handle_ObjectResponse(data_read)
		case V2_NOTIFY_NEW_BLOCK: // for notification,  instead of syncing, we will process notificaton first
			connection.Handle_Notification_Block(data_read)
			connection.Update(&command.Common) // we do it a bit later so we donot staart syncing

		case V2_NOTIFY_NEW_TX:
			connection.Update(&command.Common)
			connection.Handle_Notification_Transaction(data_read)

		default:
			connection.logger.Debugf("Unhandled v2 command %d", command.Command)

		}

	}

}

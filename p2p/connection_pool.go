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

/* this file implements the connection pool manager, keeping a list of active connections etc
 * this will also ensure that a single IP is connected only once
 *
 */
import "fmt"
import "net"
import "sync"
import "sort"
import "time"
import "strings"
import "math/big"
import "math/rand"
import "sync/atomic"
import "runtime/debug"

import "encoding/binary"

//import "container/list"

import "github.com/romana/rlog"
import "github.com/vmihailenco/msgpack"
import "github.com/dustin/go-humanize"
import log "github.com/sirupsen/logrus"
import "github.com/paulbellamy/ratecounter"
import "github.com/prometheus/client_golang/prometheus"

import "github.com/deroproject/derosuite/block"
import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/transaction"

// any connection incoming/outgoing can only be in this state
//type Conn_State uint32

const (
	HANDSHAKE_PENDING uint32 = 0 // "Pending"
	IDLE                     = 1 // "Idle"
	ACTIVE                   = 2 // "Active"
)

type Queued_Command struct {
	Command uint64 // we are waiting for this response
	BLID    []crypto.Hash
	TXID    []crypto.Hash
}

// This structure is used to do book keeping for the connection and keeps other DATA related to peer
// golang restricts 64 bit uint64/int atomic on a 64 bit boundary
// therefore all atomics are on the top
type Connection struct {
	Height       int64 // last height sent by peer  ( first member alignments issues)
	StableHeight int64 // last stable height
	TopoHeight   int64 // topo height, current topo height, this is the only thing we require for syncing

	LastObjectRequestTime int64  // when was the last item placed in object list
	BytesIn               uint64 // total bytes in
	BytesOut              uint64 // total bytes out
	Latency               int64  // time.Duration            // latency to this node when sending timed sync

	Incoming          bool              // is connection incoming or outgoing
	Addr              *net.TCPAddr      // endpoint on the other end
	Port              uint32            // port advertised by other end as its server,if it's 0 server cannot accept connections
	Peer_ID           uint64            // Remote peer id
	Lowcpuram         bool              // whether the peer has low cpu ram
	SyncNode          bool              // whether the peer has been added to command line as sync node
	Top_Version       uint64            // current hard fork version supported by peer
	TXpool_cache      map[uint64]uint32 // used for ultra blocks in miner mode,cache where we keep TX which have been broadcasted to this peer
	TXpool_cache_lock sync.RWMutex
	ProtocolVersion   string
	Tag               string // tag for the other end
	DaemonVersion     string
	//Exit                  chan bool   // Exit marker that connection needs to be killed
	ExitCounter           int32
	State                 uint32       // state of the connection
	Top_ID                crypto.Hash  // top block id of the connection
	Cumulative_Difficulty string       // cumulative difficulty of top block of peer, this is NOT required
	CDIFF                 atomic.Value //*big.Int    // NOTE: this field is used internally and is the parsed from Cumulative_Difficulty

	logger            *log.Entry // connection specific logger
	logid             string     // formatted version of connection
	Requested_Objects [][32]byte // currently unused as we sync up with a single peer at a time
	Conn              net.Conn   // actual object to talk
	//	Command_queue     *list.List               // New protocol is partly syncronous
	Objects      chan Queued_Command      // contains all objects that are requested
	SpeedIn      *ratecounter.RateCounter // average speed in last 60 seconds
	SpeedOut     *ratecounter.RateCounter // average speed in last 60 secs
	request_time atomic.Value             //time.Time                // used to track latency
	writelock    sync.Mutex               // used to Serialize writes

	sync.Mutex // used only by connection go routine

}

// 300 such buckets can be used to track block propagation accuratly upto a minute
var block_propagation = prometheus.NewHistogram(prometheus.HistogramOpts{
	Name:    "block_propagation_ms",
	Help:    "Block Propagation time milliseconds as detected by daemon",
	Buckets: prometheus.LinearBuckets(0, 1000, 20), // start 0 ms, each 1000 ms,  20 such buckets.
})

// 300 such buckets can be used to track transaction propagation accurately upto a minute
var transaction_propagation = prometheus.NewHistogram(prometheus.HistogramOpts{
	Name:    "tx_propagation_ms",
	Help:    "TX Propagation time milliseconds as detected by daemon",
	Buckets: prometheus.LinearBuckets(0, 1000, 20), // start 0 ms, each 1000 ms,  20 such buckets.
})

var block_propagation_map sync.Map
var tx_propagation_map sync.Map

/* // used to debug locks
 var x = debug.Stack
func (connection *Connection )Lock() {
    connection.x.Lock()
    connection.logger.Warnf("Locking Stack trace  \n%s", debug.Stack())

}

func (connection *Connection )Unlock() {
    connection.x.Unlock()
    connection.logger.Warnf("Unlock Stack trace  \n%s", debug.Stack())

}
*/

var connection_map sync.Map                      // map[string]*Connection{}
var connection_per_ip_counter = map[string]int{} // only keeps the counter of counter of connections
//var connection_mutex sync.Mutex

// clean up propagation
func clean_up_propagation() {

	for {
		time.Sleep(time.Minute) // cleanup every minute
		current_time := time.Now()

		// track propagation upto 10 minutes
		block_propagation_map.Range(func(k, value interface{}) bool {
			first_seen := value.(time.Time)
			if current_time.Sub(first_seen).Round(time.Second) > 600 {
				block_propagation_map.Delete(k)
			}
			return true
		})

		tx_propagation_map.Range(func(k, value interface{}) bool {
			first_seen := value.(time.Time)
			if current_time.Sub(first_seen).Round(time.Second) > 600 {
				tx_propagation_map.Delete(k)
			}
			return true
		})
	}

}

// for incoming connections we use their peer id to assertain uniquenesss
// for outgoing connections, we use the tcp endpoint address, so as not more than 1 connection is done
func Key(c *Connection) string {
	if c.Incoming {
		return fmt.Sprintf("%d", c.Peer_ID)
	}
	return string(c.Addr.String()) // Simple []byte => string conversion
}

// check whether an IP is in the map already
func IsAddressConnected(address string) bool {

	if _, ok := connection_map.Load(strings.TrimSpace(address)); ok {
		return true
	}
	return false
}

// add connection to  map
// we also check for limits for incoming connections
// same ip max 8 ip ( considering NAT)
//same Peer ID   4
func Connection_Add(c *Connection) {
	//connection_mutex.Lock()
	//defer connection_mutex.Unlock()

	ip_count := 0
	peer_id_count := 0

	incoming_ip := c.Addr.IP.String()
	incoming_peer_id := c.Peer_ID

	if c.Incoming { // we need extra protection for incoming for various attacks

		connection_map.Range(func(k, value interface{}) bool {
			v := value.(*Connection)
			if v.Incoming {
				if incoming_ip == v.Addr.IP.String() {
					ip_count++
				}

				if incoming_peer_id == v.Peer_ID {
					peer_id_count++
				}
			}
			return true
		})

	}

	if ip_count >= 8 || peer_id_count >= 4 {
		rlog.Warnf("IP address %s (%d) Peer ID %d(%d) already has too many connections, exiting this connection", incoming_ip, ip_count, incoming_peer_id, peer_id_count)
		c.Exit()
		return
	}

	connection_map.Store(Key(c), c)
}

// unique connection list
// since 2 nodes may be connected in both directions, we need to deliver new blocks/tx to only one
// thereby saving NW/computing costs
// we find duplicates using peer id
func UniqueConnections() map[uint64]*Connection {
	unique_map := map[uint64]*Connection{}

	connection_map.Range(func(k, value interface{}) bool {
		v := value.(*Connection)
		if atomic.LoadUint32(&v.State) != HANDSHAKE_PENDING && GetPeerID() != v.Peer_ID { //and skip ourselves
			unique_map[v.Peer_ID] = v // map will automatically deduplicate/overwrite previous
		}
		return true
	})

	return unique_map
}

// add connection to  map
func Connection_Delete(c *Connection) {
	connection_map.Delete(Key(c))
}

// prints all the connection info to screen
func Connection_Print() {

	fmt.Printf("Connection info for peers\n")

	if globals.Arguments["--debug"].(bool) == true {
		fmt.Printf("%-20s %-16s %-5s %-7s %-7s %23s %3s %5s %s %s %s %s %10s\n", "Remote Addr", "PEER ID", "PORT", " State", "Latency", "S/H/T", "DIR", "QUEUE", "     IN", "    OUT", " IN SPEED", " OUT SPEED", "Version")
	} else {
		fmt.Printf("%-20s %-16s %-5s %-7s %-7s %17s %3s %5s %s %s %s %s %10s\n", "Remote Addr", "PEER ID", "PORT", " State", "Latency", "H/T", "DIR", "QUEUE", "     IN", "    OUT", " IN SPEED", " OUT SPEED", "Version")

	}

	var clist []*Connection

	connection_map.Range(func(k, value interface{}) bool {
		v := value.(*Connection)
		clist = append(clist, v)
		return true
	})

	rlog.Infof("Obtained %d for connections printing", len(clist))

	// sort the list
	sort.Slice(clist, func(i, j int) bool { return clist[i].Addr.String() < clist[j].Addr.String() })

	our_topo_height := chain.Load_TOPO_HEIGHT(nil)

	for i := range clist {

		// skip pending  handshakes and skip ourselves
		if atomic.LoadUint32(&clist[i].State) == HANDSHAKE_PENDING || GetPeerID() == clist[i].Peer_ID {
			continue
		}

		dir := "OUT"
		if clist[i].Incoming {
			dir = "INC"
		}
		state := "PENDING"
		if atomic.LoadUint32(&clist[i].State) == IDLE {
			state = "IDLE"
		} else if atomic.LoadUint32(&clist[i].State) == ACTIVE {
			state = "ACTIVE"
		}

		version := clist[i].DaemonVersion

		if len(version) > 20 {
			version = version[:20]
		}

		tag := clist[i].Tag
		if len(tag) > 20 {
			tag = tag[:20]
		}

		var color_yellow = "\033[33m"
		var color_normal = "\033[0m"

		//if our_height is more than
		if our_topo_height > clist[i].TopoHeight {
			fmt.Print(color_yellow)
		}

		if globals.Arguments["--debug"].(bool) == true {
			hstring := fmt.Sprintf("%d/%d/%d", clist[i].StableHeight, clist[i].Height, clist[i].TopoHeight)
			fmt.Printf("%-20s %16x %5d %7s %7s %23s %s %5d %7s %7s %8s %9s     %10s %s\n", clist[i].Addr.IP, clist[i].Peer_ID, clist[i].Port, state, time.Duration(atomic.LoadInt64(&clist[i].Latency)).Round(time.Millisecond).String(), hstring, dir, clist[i].IsConnectionSyncing(), humanize.Bytes(atomic.LoadUint64(&clist[i].BytesIn)), humanize.Bytes(atomic.LoadUint64(&clist[i].BytesOut)), humanize.Bytes(uint64(clist[i].SpeedIn.Rate()/60)), humanize.Bytes(uint64(clist[i].SpeedOut.Rate()/60)), version, tag)

		} else {
			hstring := fmt.Sprintf("%d/%d", clist[i].Height, clist[i].TopoHeight)
			fmt.Printf("%-20s %16x %5d %7s %7s %17s %s %5d %7s %7s %8s %9s     %10s %s\n", clist[i].Addr.IP, clist[i].Peer_ID, clist[i].Port, state, time.Duration(atomic.LoadInt64(&clist[i].Latency)).Round(time.Millisecond).String(), hstring, dir, clist[i].IsConnectionSyncing(), humanize.Bytes(atomic.LoadUint64(&clist[i].BytesIn)), humanize.Bytes(atomic.LoadUint64(&clist[i].BytesOut)), humanize.Bytes(uint64(clist[i].SpeedIn.Rate()/60)), humanize.Bytes(uint64(clist[i].SpeedOut.Rate()/60)), version, tag)

		}

		fmt.Print(color_normal)
	}

}

// for continuos update on command line, get the maximum height of all peers
// show the average network status
func Best_Peer_Height() (best_height, best_topo_height int64) {

	var heights []uint64
	var topoheights []uint64

	connection_map.Range(func(k, value interface{}) bool {
		v := value.(*Connection)
		if atomic.LoadUint32(&v.State) != HANDSHAKE_PENDING {
			height := atomic.LoadInt64(&v.Height)
			heights = append(heights, uint64(height))
			topoheights = append(topoheights, uint64(atomic.LoadInt64(&v.TopoHeight)))
		}
		return true
	})

	best_height = int64(Median(heights))
	best_topo_height = int64(Median(topoheights))

	return
}

// this function return peer count which have successful handshake
func Disconnect_All() (Count uint64) {
	return
	/*
		connection_mutex.Lock()
		for _, v := range connection_map {
			// v.Lock()
			close(v.Exit) // close the connection
			//v.Unlock()
		}
		connection_mutex.Unlock()
		return
	*/
}

// this function return peer count which have successful handshake
func Peer_Count() (Count uint64) {

	connection_map.Range(func(k, value interface{}) bool {
		v := value.(*Connection)
		if atomic.LoadUint32(&v.State) != HANDSHAKE_PENDING && GetPeerID() != v.Peer_ID {
			Count++
		}
		return true
	})

	return
}

// this function returnw random connection which have successful handshake
func Random_Connection(height int64) (c *Connection) {

	var clist []*Connection

	connection_map.Range(func(k, value interface{}) bool {
		v := value.(*Connection)
		if atomic.LoadInt64(&v.Height) >= height {
			clist = append(clist, v)
		}
		return true
	})

	if len(clist) > 0 {
		return clist[rand.Int()%len(clist)]
	}

	return nil
}

// this returns count of peers in both directions
func Peer_Direction_Count() (Incoming uint64, Outgoing uint64) {

	connection_map.Range(func(k, value interface{}) bool {
		v := value.(*Connection)
		if atomic.LoadUint32(&v.State) != HANDSHAKE_PENDING && GetPeerID() != v.Peer_ID {
			if v.Incoming {
				Incoming++
			} else {
				Outgoing++
			}
		}
		return true
	})

	return
}

// broad cast a block to all connected peers
// we can only broadcast a block which is in our db
// this function is trigger from 2 points, one when we receive a unknown block which can be successfully added to chain
// second from the blockchain which has to relay locally  mined blocks as soon as possible
func Broadcast_Block(cbl *block.Complete_Block, PeerID uint64) { // if peerid is provided it is skipped
	var request Notify_New_Objects_Struct

	defer func() {
		if r := recover(); r != nil {
			logger.Warnf("Recovered while broadcasting Block, Stack trace below", r)
			logger.Warnf("Stack trace  \n%s", debug.Stack())
		}
	}()



	/*if IsSyncing() { // if we are syncing, do NOT broadcast the block
		return
	}*/

	fill_common(&request.Common) // fill common info
	request.Command = V2_NOTIFY_NEW_BLOCK
	request.CBlock.Block = cbl.Bl.Serialize()

	for i := range cbl.Txs {
		request.CBlock.Txs = append(request.CBlock.Txs, cbl.Txs[i].Serialize())
	}

	serialized, err := msgpack.Marshal(&request)
	if err != nil {
		panic(err)
	}

	our_height := chain.Get_Height()
	// build the request once and dispatch it to all possible peers
	count := 0
	unique_map := UniqueConnections()

	for _, v := range unique_map {
		select {
		case <-Exit_Event:
			return
		default:
		}
		if atomic.LoadUint32(&v.State) != HANDSHAKE_PENDING && PeerID != v.Peer_ID { // skip pre-handshake connections

			// if the other end is > 50 blocks behind, do not broadcast block to hime
			// this is an optimisation, since if the other end is syncing
			// every peer will keep on broadcasting and thus making it more lagging
			// due to overheads
			peer_height := atomic.LoadInt64(&v.Height)
			if (our_height - peer_height) > 25 {
				continue
			}

			count++
			go func(connection *Connection) {
				defer func() {
					if r := recover(); r != nil {
						rlog.Warnf("Recovered while handling connection, Stack trace below", r)
						rlog.Warnf("Stack trace  \n%s", debug.Stack())
					}
				}()
				if globals.Arguments["--lowcpuram"].(bool) == false && connection.TXpool_cache != nil { // everyone needs ultrac compact block if possible
					var miner_specific_request Notify_New_Objects_Struct
					miner_specific_request.Common = request.Common
					miner_specific_request.Command = V2_NOTIFY_NEW_BLOCK
					miner_specific_request.CBlock.Block = request.CBlock.Block

					sent := 0
					skipped := 0
					connection.TXpool_cache_lock.RLock()
					for i := range cbl.Bl.Tx_hashes {
						// in ultra compact mode send a transaction only if we know that we have not sent that transaction earlier
						// send only tx not found in cache

						if _, ok := connection.TXpool_cache[binary.LittleEndian.Uint64(cbl.Bl.Tx_hashes[i][:])]; !ok {
							miner_specific_request.CBlock.Txs = append(miner_specific_request.CBlock.Txs, request.CBlock.Txs[i])
							sent++

						} else {
							skipped++
						}

					}
					connection.TXpool_cache_lock.RUnlock()

					connection.logger.Debugf("Sending ultra block to peer total %d tx skipped %d sent %d", len(cbl.Bl.Tx_hashes), skipped, sent)

					serialized_miner_specific, err := msgpack.Marshal(&miner_specific_request)
					if err != nil {
						panic(err)
					}
					connection.Send_Message(serialized_miner_specific) // miners need ultra compact blocks

				} else {
					connection.logger.Debugf("Sending full block to peer")

					connection.Send_Message(serialized) // non miners need full blocks
				}
			}(v)
		}

	}

	rlog.Infof("Broadcasted block %s to %d peers", cbl.Bl.GetHash(), count)

}

// broadcast a new transaction, return to how many peers the transaction has been broadcasted
// this function is trigger from 2 points, one when we receive a unknown tx
// second from the mempool which may want to relay local ot soon going to expire transactions
func Broadcast_Tx(tx *transaction.Transaction, PeerID uint64) (relayed_count int) {

	defer func() {
		if r := recover(); r != nil {
			logger.Warnf("Recovered while broadcasting TX, Stack trace below", r)
			logger.Warnf("Stack trace  \n%s", debug.Stack())
		}
	}()

	var request Notify_New_Objects_Struct

	fill_common_skip_topoheight(&request.Common) // fill common info, but skip topo height
	request.Command = V2_NOTIFY_NEW_TX
	request.Tx = tx.Serialize()

	serialized, err := msgpack.Marshal(&request)
	if err != nil {
		panic(err)
	}

	txhash := tx.GetHash()
	our_height := chain.Get_Height()

	unique_map := UniqueConnections()

	for _, v := range unique_map {
		select {
		case <-Exit_Event:
			return
		default:
		}
		if atomic.LoadUint32(&v.State) != HANDSHAKE_PENDING && PeerID != v.Peer_ID { // skip pre-handshake connections

			// if the other end is > 50 blocks behind, do not broadcast block to hime
			// this is an optimisation, since if the other end is syncing
			// every peer will keep on broadcasting and thus making it more lagging
			// due to overheads
			// if we are lagging or peer is lagging, do not brodcast transactions
			peer_height := atomic.LoadInt64(&v.Height)
			if (our_height - peer_height) > 25 ||  (our_height+5) < peer_height   {
				continue
			}

			relayed_count++
			go func(connection *Connection) {
				defer func() {
					if r := recover(); r != nil {
						rlog.Warnf("Recovered while handling connection, Stack trace below", r)
						rlog.Warnf("Stack trace  \n%s", debug.Stack())
					}
				}()

				connection.Send_Message(serialized) // send the bytes

				// mark the TX  as having sent to this node
				// when this tx is
				connection.TXpool_cache_lock.Lock()
				// disable cache if not possible due to options
				if globals.Arguments["--lowcpuram"].(bool) == false && connection.TXpool_cache != nil {
					connection.TXpool_cache[binary.LittleEndian.Uint64(txhash[:])] = uint32(time.Now().Unix())
				}
				connection.TXpool_cache_lock.Unlock()

			}(v)
		}

	}
	//rlog.Infof("Broadcasted tx %s to %d peers", txhash, relayed_count)
	return
}

//var sync_in_progress bool

// we can tell whether we are syncing by seeing the pending queue of expected response
// if objects response are queued, we are syncing
// if even one of the connection is syncing, then we are syncronising
// returns a number how many blocks are queued
func (connection *Connection) IsConnectionSyncing() (count int) {
	//connection.Lock()
	//defer connection.Unlock()

	if atomic.LoadUint32(&connection.State) == HANDSHAKE_PENDING { // skip pre-handshake connections
		return 0
	}

	// check whether 15 secs have passed, if yes close the connection
	// so we can try some other connection
	if len(connection.Objects) > 0 {
		if time.Now().Unix() >= (13 + atomic.LoadInt64(&connection.LastObjectRequestTime)) {
			connection.Exit()
			return 0
		}
	}

	return len(connection.Objects)

}

// trigger a sync with a random peer
func trigger_sync() {

	defer func() {
		if r := recover(); r != nil {
			logger.Warnf("Recovered while triggering sync, Stack trace below", r)
			logger.Warnf("Stack trace  \n%s", debug.Stack())
		}
	}()

	_, topoheight := Best_Peer_Height()

	unique_map := UniqueConnections()


	var clist []*Connection

	for _,value := range unique_map {
		clist = append(clist, value)

	}

	
	// sort the list random
	// do random shuffling, can we get away with len/2 random shuffling
	globals.Global_Random.Shuffle(len(clist), func(i, j int) {
		clist[i], clist[j] = clist[j], clist[i]
	})

	

	for _, connection := range clist {

		//connection.Lock()   recursive mutex are not suported
		// only choose highest available peers for syncing
		if atomic.LoadUint32(&connection.State) != HANDSHAKE_PENDING && topoheight <= atomic.LoadInt64(&connection.TopoHeight) { // skip pre-handshake connections
			// check whether we are lagging with this connection
			//connection.Lock()
			islagging := chain.IsLagging(connection.CDIFF.Load().(*big.Int)) // we only use cdiff to see if we need to resync
			// islagging := true
			//connection.Unlock()
			if islagging {

				//connection.Lock()
				connection.logger.Debugf("We need to resync with the peer height %d", connection.Height)
				//connection.Unlock()
				// set mode to syncronising
				connection.Send_ChainRequest()
				break
			}
		}

	}

}

//detect if something is queued to any of the peer
// is something is queue we are syncing
func IsSyncing() (result bool) {

	syncing := false
	connection_map.Range(func(k, value interface{}) bool {
		v := value.(*Connection)
		if v.IsConnectionSyncing() != 0 {
			syncing = true
			return false
		}
		return true
	})
	return syncing
}

// detect whether we are behind any of the connected peers and trigger sync ASAP
// randomly with one of the peers
func syncroniser() {
	for {
		select {
		case <-Exit_Event:
			return
		case <-time.After(1000 * time.Millisecond):
		}

		if !IsSyncing() {
			trigger_sync() // check whether we are out of sync
		}

	}

}

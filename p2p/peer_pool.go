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

/* this file implements the peer manager, keeping a list of peers which can be tried for connection etc
 *
 */
import "os"
import "fmt"

//import "net"
import "sync"
import "time"
import "sort"
import "path/filepath"
import "encoding/json"

//import "encoding/binary"
//import "container/list"

//import log "github.com/sirupsen/logrus"

import "github.com/deroproject/derosuite/globals"

//import "github.com/deroproject/derosuite/crypto"

// This structure is used to do book keeping for the peer list and keeps other DATA related to peer
// all peers are servers, means they have exposed a port for connections
// all peers are identified by their endpoint tcp address
// all clients are identified by ther peer id ( however ip-address is used to control amount )
// the following daemon commands interact with the list
// peer_list := print the peer list
// ban address  time  // ban this address for spcific time
// unban address
// enableban address  // by default all addresses are bannable
// disableban address  // this address will never be banned
type Peer struct {
	Address string `json:"address"` // pairs in the ip:port or dns:port, basically  endpoint
	ID      uint64 `json:"peerid"`  // peer id
	Miner   bool   `json:"miner"`   // miner
	//NeverBlacklist    bool    // this address will never be blacklisted
	LastConnected   uint64 `json:"lastconnected"`   // epoch time when it was connected , 0 if never connected
	FailCount       uint64 `json:"failcount"`       // how many times have we failed  (tcp errors)
	ConnectAfter    uint64 `json:"connectafter"`    // we should connect when the following timestamp passes
	BlacklistBefore uint64 `json:"blacklistbefore"` // peer blacklisted till epoch , priority nodes are never blacklisted, 0 if not blacklist
	GoodCount       uint64 `json:"goodcount"`       // how many times peer has been shared with us
	Version         int    `json:"version"`         // version 1 is original C daemon peer, version 2 is golang p2p version
	Whitelist       bool   `json:"whitelist"`
	sync.Mutex
}

var peer_map = map[string]*Peer{}
var peer_mutex sync.Mutex

// loads peers list from disk
func load_peer_list() {
	defer clean_up()
	peer_mutex.Lock()
	defer peer_mutex.Unlock()

	peer_file := filepath.Join(globals.GetDataDirectory(), "peers.json")
	file, err := os.Open(peer_file)
	if err != nil {
		logger.Warnf("Error opening peer data file %s err %s", peer_file, err)
	} else {
		defer file.Close()
		decoder := json.NewDecoder(file)
		err = decoder.Decode(&peer_map)
		if err != nil {
			logger.Warnf("Error unmarshalling p2p data err %s", err)
		} else { // successfully unmarshalled data
			logger.Debugf("Successfully loaded %d peers from  file", (len(peer_map)))
		}
	}

}

//save peer list to disk
func save_peer_list() {

	clean_up()
	peer_mutex.Lock()
	defer peer_mutex.Unlock()

	peer_file := filepath.Join(globals.GetDataDirectory(), "peers.json")
	file, err := os.Create(peer_file)
	if err != nil {
		logger.Warnf("Error creating peer data file %s err %s", peer_file, err)
	} else {
		defer file.Close()
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "\t")
		err = encoder.Encode(&peer_map)
		if err != nil {
			logger.Warnf("Error marshalling p2p data err %s", err)
		} else { // successfully unmarshalled data
			logger.Debugf("Successfully saved %d peers to file", (len(peer_map)))
		}
	}
}

// clean up by discarding entries which are too much into future
func clean_up() {
	peer_mutex.Lock()
	defer peer_mutex.Unlock()
	for k, v := range peer_map {
		if v.FailCount >= 16 { // roughly 16 tries, 18 hrs before we discard the peer
			delete(peer_map, k)
		}
	}

}

// check whether an IP is in the map already
func IsPeerInList(address string) bool {
	peer_mutex.Lock()
	defer peer_mutex.Unlock()

	if _, ok := peer_map[address]; ok {
		return true
	}
	return false
}
func GetPeerInList(address string) *Peer {
	peer_mutex.Lock()
	defer peer_mutex.Unlock()

	if v, ok := peer_map[address]; ok {
		return v
	}
	return nil
}

// add connection to  map
func Peer_Add(p *Peer) {
	peer_mutex.Lock()
	defer peer_mutex.Unlock()

	if p.ID == GetPeerID() { // if peer is self do not connect
		// logger.Infof("Peer is ourselves, discard")
		return

	}

	if v, ok := peer_map[p.Address]; ok {
		v.Lock()
		// logger.Infof("Peer already in list adding good count")
		v.GoodCount++
		v.Unlock()
	} else {
		// logger.Infof("Peer adding to list")
		peer_map[p.Address] = p
	}
}

// a peer marked as fail, will only be connected  based on exponential back-off based on powers of 2
func Peer_SetFail(address string) {
	p := GetPeerInList(address)
	if p == nil {
		return
	}
	peer_mutex.Lock()
	defer peer_mutex.Unlock()
	p.FailCount++ //  increase fail count, and mark for delayed connect
	p.ConnectAfter = uint64(time.Now().UTC().Unix()) + 1<<(p.FailCount-1)
}

// set peer as successfully connected
// we will only distribute peers which have been successfully connected by us
func Peer_SetSuccess(address string) {
	//logger.Infof("Setting peer as success")
	p := GetPeerInList(address)
	if p == nil {
		return
	}
	peer_mutex.Lock()
	defer peer_mutex.Unlock()
	p.FailCount = 0 //  fail count is zero again
	p.ConnectAfter = 0
	p.Whitelist = true
	p.LastConnected = uint64(time.Now().UTC().Unix()) // set time when last connected
	// logger.Infof("Setting peer as white listed")
}

/*
 //TODO do we need a functionality so some peers are never banned
func Peer_DisableBan(address string) (err error){
    p := GetPeerInList(address)
    if p == nil {
     return fmt.Errorf("Peer \"%s\" not found in list")
    }
    p.Lock()
    defer p.Unlock()
    p.NeverBlacklist = true
}

func Peer_EnableBan(address string) (err error){
    p := GetPeerInList(address)
    if p == nil {
     return fmt.Errorf("Peer \"%s\" not found in list")
    }
    p.Lock()
    defer p.Unlock()
    p.NeverBlacklist = false
}
*/

// add connection to  map
func Peer_Delete(p *Peer) {
	peer_mutex.Lock()
	defer peer_mutex.Unlock()
	delete(peer_map, p.Address)
}

// prints all the connection info to screen
func PeerList_Print() {
	peer_mutex.Lock()
	defer peer_mutex.Unlock()
	fmt.Printf("Peer List\n")
	fmt.Printf("%-22s %-6s %-4s   %-5s %-7s %9s %3s\n", "Remote Addr", "Active", "Good", "Fail", " State", "Height", "DIR")

	var list []*Peer
	greycount := 0
	for _, v := range peer_map {
		if v.Whitelist { // only display white listed peer
			list = append(list, v)
		} else {
			greycount++
		}
	}

	// sort the list
	sort.Slice(list, func(i, j int) bool { return list[i].Address < list[j].Address })

	for i := range list {
		connected := ""
		if IsAddressConnected(list[i].Address) {
			connected = "ACTIVE"
		}
		fmt.Printf("%-22s %-6s %4d %5d \n", list[i].Address, connected, list[i].GoodCount, list[i].FailCount)
	}

	fmt.Printf("\nWhitelist size %d\n", len(peer_map)-greycount)
	fmt.Printf("Greylist size %d\n", greycount)

}

// this function return peer count which have successful handshake
func Peer_Counts() (Count uint64) {
	peer_mutex.Lock()
	defer peer_mutex.Unlock()
	return uint64(len(peer_map))
}

// this function finds a possible peer to connect to keeping blacklist and already existing connections into picture
// it must not be already connected using outgoing connection
// we do allow loops such as both  incoming/outgoing simultaneously
// this will return atmost 1 address, empty address if peer list is empty
func find_peer_to_connect(version int) *Peer {
	defer clean_up()
	peer_mutex.Lock()
	defer peer_mutex.Unlock()

	// first search the whitelisted ones
	for _, v := range peer_map {
		if uint64(time.Now().Unix()) > v.BlacklistBefore && //  if ip is blacklisted skip it
			uint64(time.Now().Unix()) > v.ConnectAfter &&
			!IsAddressConnected(v.Address) && v.Whitelist && !IsAddressInBanList(v.Address) {
			v.ConnectAfter = uint64(time.Now().UTC().Unix()) + 10 // minimum 10 secs gap
			return v
		}
	}
	// if we donot have any white listed, choose from the greylist
	for _, v := range peer_map {
		if uint64(time.Now().Unix()) > v.BlacklistBefore && //  if ip is blacklisted skip it
			uint64(time.Now().Unix()) > v.ConnectAfter &&
			!IsAddressConnected(v.Address) && !v.Whitelist && !IsAddressInBanList(v.Address) {
			v.ConnectAfter = uint64(time.Now().UTC().Unix()) + 10 // minimum 10 secs gap
			return v
		}
	}

	return nil // if no peer found, return nil
}

// return white listed peer list
// for use in handshake
func get_peer_list() (peers []Peer_Info) {
	peer_mutex.Lock()
	defer peer_mutex.Unlock()

	for _, v := range peer_map {
		if v.Whitelist {
			peers = append(peers, Peer_Info{Addr: v.Address})
		}
	}
	return
}

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

//import "fmt"
//import "net"
import "sync"
import "time"

//import "container/list"

//import log "github.com/sirupsen/logrus"
//import "github.com/vmihailenco/msgpack"

import "github.com/deroproject/derosuite/crypto"

//import "github.com/deroproject/derosuite/globals"

import "github.com/deroproject/derosuite/block"

// if block request pool is empty, we are syncronised otherwise we are syncronising
var block_request_pool = map[crypto.Hash]uint64{} // these are received, we must attach a connection to blacklist peers

var block_received = map[crypto.Hash]*block.Complete_Block{} // once blocks are received, they are placed here
var block_request_pool_mutex sync.Mutex

func queue_block(blid crypto.Hash) {
	block_request_pool_mutex.Lock()
	defer block_request_pool_mutex.Unlock()

	// if object has already been received, it no longer should be request
	if _, ok := block_received[blid]; ok { // object is already in pool
		return
	}

	// if object has already been requested, skip it , else add it to queue
	if _, ok := block_request_pool[blid]; !ok {
		block_request_pool[blid] = 0
	}
}

// a block hash been received, make it ready for consumption
func queue_block_received(blid crypto.Hash, cbl *block.Complete_Block) {
	block_request_pool_mutex.Lock()
	defer block_request_pool_mutex.Unlock()

	if _, ok := block_request_pool[blid]; !ok {
		// unknown object received discard it
		return
	}
	delete(block_request_pool, blid)

	block_received[blid] = cbl
}

// continusly retrieve_objects
func retrieve_objects() {

	for {
		select {
		case <-Exit_Event:
			return
		case <-time.After(5 * time.Second):
		}

		block_request_pool_mutex.Lock()

		for k, _ := range block_request_pool {

			connection := Random_Connection(0)
			if connection != nil {
				connection.Send_ObjectRequest([]crypto.Hash{k}, []crypto.Hash{})
			}
		}
		block_request_pool_mutex.Unlock()

	}

}

// this goroutine will keep searching the queue for any blocks and see if they are can be attached somewhere, if yes they will be attached and cleaned up
func sync_retrieved_blocks() {
	// success := make(chan bool)
	for {

		select {
		case <-Exit_Event:
			return
		//case <- success:
		case <-time.After(1 * time.Second):
		}

		block_request_pool_mutex.Lock()

		for blid, cbl := range block_received {
			if chain.Block_Exists(nil, blid) { /// block is already in chain, discard it
				delete(block_received, blid)
				continue
			}
			_ = cbl
			/*if cbl == nil {
			    delete(block_received,blid)
			     block_request_pool[blid]=0 // need to request object again
			    continue
			}*/

			/*
				// attach it to parent, else skip for now
				if chain.Block_Exists(cbl.Bl.Prev_Hash) {
					if chain.Add_Complete_Block(cbl) {
						// if block successfully added , delete it now
						delete(block_received, blid)

						continue
					} else { // checksum of something failed, request it randomly from another peer
						block_request_pool[blid] = 0
					}
				}*/
		}
		block_request_pool_mutex.Unlock()
	}
}

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

package mempool

import "sync"
import "time"
import "sync/atomic"

import log "github.com/sirupsen/logrus"

import "github.com/deroproject/derosuite/transaction"
import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/crypto"

// NOTE: do NOT consider this code as useless, as it is used to avooid double spending attacks within the block
// let me explain, since we are a state machine, we add block to our blockchain
// so, if a double spending attack comes, 2 transactions with same inputs, we reject one of them
// the algo is documented somewhere else  which explains the entire process

// at this point in time, this is an ultrafast written mempool,
// it will not scale for more than 10000 transactions  but is good enough for now
// we can always come back and rewrite it
// NOTE: the pool is not persistant , means closing the daemon will make the mempool empty next restart
// TODO:  make the pool persistant
type Mempool struct {
	txs        map[crypto.Hash]mempool_object
	key_images map[crypto.Hash]bool // contains key images of all txs
	modified   bool                 // used to monitor whethel mem pool contents have changed,

	// global variable , but don't see it utilisation here except fot tx verification
	//chain *Blockchain

	sync.Mutex
}

type mempool_object struct {
	Tx     *transaction.Transaction
	Added  uint64 // time in epoch format
	Reason int    //  why is the tx in the mempool
}

var loggerpool *log.Entry

func Init_Mempool(params map[string]interface{}) (*Mempool, error) {
	var mempool Mempool
	//mempool.chain = params["chain"].(*Blockchain)

	loggerpool = globals.Logger.WithFields(log.Fields{"com": "POOL"}) // all components must use this logger
	loggerpool.Infof("Mempool started")
	atomic.AddUint32(&globals.Subsystem_Active, 1) // increment subsystem

	// initialize maps
	mempool.txs = map[crypto.Hash]mempool_object{}
	mempool.key_images = map[crypto.Hash]bool{}
	//TODO load any trasactions saved at previous exit

	return &mempool, nil
}

// this is created per incoming block and then discarded
// This does not require shutting down and will be garbage collected automatically
func Init_Block_Mempool(params map[string]interface{}) (*Mempool, error) {
	var mempool Mempool

	// initialize maps
	mempool.txs = map[crypto.Hash]mempool_object{}
	mempool.key_images = map[crypto.Hash]bool{}
	//TODO load any trasactions saved at previous exit

	return &mempool, nil
}

func (pool *Mempool) Shutdown() {
	//TODO save mempool tx somewhere
	loggerpool.Infof("Mempool stopped")
	atomic.AddUint32(&globals.Subsystem_Active, ^uint32(0)) // this decrement 1 fom subsystem

}

// start pool monitoring for changes for some specific time
// this is required so as we can add or discard transactions while selecting work for mining
func (pool *Mempool) Monitor() {
	pool.Lock()
	pool.modified = false
	pool.Unlock()
}

// return whether pool contents have changed
func (pool *Mempool) HasChanged() (result bool) {
	pool.Lock()
	result = pool.modified
	pool.Unlock()
	return
}

// a tx should only be added to pool after verification is complete
func (pool *Mempool) Mempool_Add_TX(tx *transaction.Transaction, Reason int) (result bool) {
	result = false
	pool.Lock()
	defer pool.Unlock()

	var object mempool_object

	tx_hash := crypto.Hash(tx.GetHash())

	// check if tx already exists, skip it
	if _, ok := pool.txs[tx_hash]; ok {
		loggerpool.Infof("Pool already contains %s, skipping \n", tx_hash)
		return false
	}

	// we should also extract all key images and add them to have multiple pending
	for i := 0; i < len(tx.Vin); i++ {
		if _, ok := pool.key_images[tx.Vin[i].(transaction.Txin_to_key).K_image]; ok {
			loggerpool.WithFields(log.Fields{
				"txid":   tx_hash,
				"kimage": tx.Vin[i].(transaction.Txin_to_key).K_image,
			}).Warnf("TX using inputs  which have already been used, Possible Double spend attack rejected")
			return false
		}
	}

	// add all the key images to check double spend attack within the pool
	for i := 0; i < len(tx.Vin); i++ {
		pool.key_images[tx.Vin[i].(transaction.Txin_to_key).K_image] = true // add element to map for next check
	}

	// we are here means we can add it to pool
	object.Tx = tx
	object.Reason = Reason
	object.Added = uint64(time.Now().Unix())

	pool.txs[tx_hash] = object
	pool.modified = true // pool has been modified

	return true
}

// check whether a tx exists in the pool
func (pool *Mempool) Mempool_TX_Exist(txid crypto.Hash) (result bool) {
	pool.Lock()
	defer pool.Unlock()

	if _, ok := pool.txs[txid]; ok {
		return true
	}
	return false
}

// delete specific tx from pool and return it
// if nil is returned Tx was not found in pool
func (pool *Mempool) Mempool_Delete_TX(txid crypto.Hash) (tx *transaction.Transaction) {
	pool.Lock()
	defer pool.Unlock()

	// check if tx already exists, skip it
	if _, ok := pool.txs[txid]; !ok {
		loggerpool.Warnf("Pool does NOT contain %s, returning nil \n", txid)
		return nil
	}

	// we reached here means, we have the tx remove it from our list, do maintainance cleapup and discard it
	object := pool.txs[txid]
	delete(pool.txs, txid)

	// remove all the key images
	for i := 0; i < len(object.Tx.Vin); i++ {
		delete(pool.key_images, object.Tx.Vin[i].(transaction.Txin_to_key).K_image)
	}

	pool.modified = true // pool has been modified
	return object.Tx     // return the tx
}

// get specific tx from mem pool without removing it
func (pool *Mempool) Mempool_Get_TX(txid crypto.Hash) (tx *transaction.Transaction) {
	pool.Lock()
	defer pool.Unlock()

	if _, ok := pool.txs[txid]; !ok {
		//loggerpool.Warnf("Pool does NOT contain %s, returning nil \n", txid)
		return nil
	}

	// we reached here means, we have the tx, return the pointer back
	object := pool.txs[txid]

	return object.Tx
}

// return list of all txs in pool
func (pool *Mempool) Mempool_List_TX() []crypto.Hash {
	pool.Lock()
	defer pool.Unlock()

	var list []crypto.Hash
	for k, _ := range pool.txs {
		list = append(list, k)
	}

	return list
}

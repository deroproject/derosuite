package blockchain


import "sync"
import "time"
import "sync/atomic"

import log "github.com/sirupsen/logrus"

import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/crypto"



// at this point in time, this is an ultrafast written mempool,
// it will not scale for more than 10000 transactions  but is good enough for now
// we can always come back and rewrite it
type Mempool struct {
	txs map[crypto.Hash]mempool_object

	// global variable , but don't see it utilisation here except fot tx verification
	chain *Blockchain

	sync.Mutex
}

type mempool_object struct {
	Tx     *Transaction
	Added  uint64
	Reason int //  why is the tx in the mempool
}

var loggerpool *log.Entry

func Init_Mempool(params map[string]interface{}) (*Mempool, error) {
	var mempool Mempool
	//mempool.chain = params["chain"].(*Blockchain)

	loggerpool = globals.Logger.WithFields(log.Fields{"com": "POOL"}) // all components must use this logger
	loggerpool.Infof("Mempool started")
	atomic.AddUint32(&globals.Subsystem_Active, 1) // increment subsystem

	//TODO load any trasactions saved at previous exit

	return &mempool, nil
}

func (pool *Mempool) Shutdown() {
        //TODO save mempool tx somewhere
	loggerpool.Infof("Mempool stopped")
	atomic.AddUint32(&globals.Subsystem_Active, ^uint32(0)) // this decrement 1 fom subsystem

}

func (pool *Mempool) Mempool_Add_TX(tx *Transaction, Reason int) {
	pool.Lock()
	defer pool.Unlock()

	var object mempool_object

	hash := crypto.Hash(tx.GetHash())

	// check if tx already exists, skip it
	if _, ok := pool.txs[hash]; ok {
		loggerpool.Infof("Pool already contains %x, skipping \n", hash)
		return
	}
	object.Tx = tx
	object.Reason = Reason
	object.Added = uint64(time.Now().Unix())

	pool.txs[hash] = object

}

// delete specific tx from pool
func (pool *Mempool) Mempool_Delete_TX(crypto.Hash) {

}

// get specific tx from mem pool
func (pool *Mempool) Mempool_Get_TX(txid crypto.Hash) ([]byte, error) {
	return nil, nil
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

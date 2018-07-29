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

import "os"
import "fmt"
import "sync"
import "sort"
import "time"
import "sync/atomic"
import "path/filepath"
import "encoding/hex"
import "encoding/json"

import "github.com/romana/rlog"
import log "github.com/sirupsen/logrus"

import "github.com/deroproject/derosuite/transaction"
import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/crypto"

// this is only used for sorting and nothing else
type TX_Sorting_struct struct {
	FeesPerByte uint64      // this is fees per byte
	Hash        crypto.Hash // transaction hash
	Size        uint64      // transaction size
}

// NOTE: do NOT consider this code as useless, as it is used to avooid double spending attacks within the block and within the pool
// let me explain, since we are a state machine, we add block to our blockchain
// so, if a double spending attack comes, 2 transactions with same inputs, we reject one of them
// the algo is documented somewhere else  which explains the entire process

// at this point in time, this is an ultrafast written mempool,
// it will not scale for more than 10000 transactions  but is good enough for now
// we can always come back and rewrite it
// NOTE: the pool is now persistant
type Mempool struct {
	txs           sync.Map //map[crypto.Hash]*mempool_object
	key_images    sync.Map //map[crypto.Hash]bool // contains key images of all txs
	sorted_by_fee []crypto.Hash        // contains txids sorted by fees
	sorted        []TX_Sorting_struct  // contains TX sorting information, so as new block can be forged easily
	modified      bool                 // used to monitor whethel mem pool contents have changed,
	height        uint64               // track blockchain height

	P2P_TX_Relayer p2p_TX_Relayer // actual pointer, setup by the dero daemon during runtime

	// global variable , but don't see it utilisation here except fot tx verification
	//chain *Blockchain
	Exit_Mutex chan bool

	sync.Mutex
}

// this object is serialized  and deserialized
type mempool_object struct {
	Tx         *transaction.Transaction
	Added      uint64 // time in epoch format
	Height     uint64 //  at which height the tx unlocks in the mempool
	Relayed    int    // relayed count
	RelayedAt  int64  // when was tx last relayed
	Size       uint64 // size in bytes of the TX
	FEEperBYTE uint64 // fee per byte
}

var loggerpool *log.Entry

// marshal object as json
func (obj *mempool_object) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Tx        string `json:"tx"` // hex encoding
		Added     uint64 `json:"added"`
		Height    uint64 `json:"height"`
		Relayed   int    `json:"relayed"`
		RelayedAt int64  `json:"relayedat"`
	}{
		Tx:        hex.EncodeToString(obj.Tx.Serialize()),
		Added:     obj.Added,
		Height:    obj.Height,
		Relayed:   obj.Relayed,
		RelayedAt: obj.RelayedAt,
	})
}

// unmarshal object from json encoding
func (obj *mempool_object) UnmarshalJSON(data []byte) error {
	aux := &struct {
		Tx        string `json:"tx"`
		Added     uint64 `json:"added"`
		Height    uint64 `json:"height"`
		Relayed   int    `json:"relayed"`
		RelayedAt int64  `json:"relayedat"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	obj.Added = aux.Added
	obj.Height = aux.Height
	obj.Relayed = aux.Relayed
	obj.RelayedAt = aux.RelayedAt

	tx_bytes, err := hex.DecodeString(aux.Tx)
	if err != nil {
		return err
	}
	obj.Size = uint64(len(tx_bytes))

	obj.Tx = &transaction.Transaction{}
	err = obj.Tx.DeserializeHeader(tx_bytes)

	if err == nil {
		obj.FEEperBYTE = obj.Tx.RctSignature.Get_TX_Fee() / obj.Size
	}
	return err
}

func Init_Mempool(params map[string]interface{}) (*Mempool, error) {
	var mempool Mempool
	//mempool.chain = params["chain"].(*Blockchain)

	loggerpool = globals.Logger.WithFields(log.Fields{"com": "POOL"}) // all components must use this logger
	loggerpool.Infof("Mempool started")
	atomic.AddUint32(&globals.Subsystem_Active, 1) // increment subsystem

	mempool.Exit_Mutex = make(chan bool)

	// initialize maps
	//mempool.txs = map[crypto.Hash]*mempool_object{}
	//mempool.key_images = map[crypto.Hash]bool{}

	//TODO load any trasactions saved at previous exit

	mempool_file := filepath.Join(globals.GetDataDirectory(), "mempool.json")

	file, err := os.Open(mempool_file)
	if err != nil {
		loggerpool.Warnf("Error opening mempool data file %s err %s", mempool_file, err)
	} else {
		defer file.Close()

		var objects []mempool_object
		decoder := json.NewDecoder(file)
		err = decoder.Decode(&objects)
		if err != nil {
			loggerpool.Warnf("Error unmarshalling mempool data err %s", err)
		} else { // successfully unmarshalled data, add it to mempool
			loggerpool.Debugf("Will try to load %d txs from mempool file", (len(objects)))
			for i := range objects {
				result := mempool.Mempool_Add_TX(objects[i].Tx, 0)
				if result { // setup time
					//mempool.txs[objects[i].Tx.GetHash()] = &objects[i] // setup time and other artifacts
					mempool.txs.Store(objects[i].Tx.GetHash(),&objects[i] )
				}
			}
		}
	}

	go mempool.Relayer_and_Cleaner()

	return &mempool, nil
}

// this is created per incoming block and then discarded
// This does not require shutting down and will be garbage collected automatically
func Init_Block_Mempool(params map[string]interface{}) (*Mempool, error) {
	var mempool Mempool

	// initialize maps
	//mempool.txs = map[crypto.Hash]*mempool_object{}
	//mempool.key_images = map[crypto.Hash]bool{}

	return &mempool, nil
}

func (pool *Mempool) HouseKeeping(height uint64, Verifier func(*transaction.Transaction) bool) {
	pool.height = height

	// this code is executed in rare conditions which are as follows
	// chain has a tx which has spent most recent input possible (10 block)
	// chain height = say 1000 , tx consumes outputs from 9990
	// then a major soft fork occurs roughly at say 9997
	// under this case of reorganisation, tx would have failed, since are still not mature
	// so what we do is, when we push tx during reorganisation,we verify and make them available at point when they were mined
	// thereby avoiding problem of input maturity
	// carry over book keeping activities, such as keep transactions hidden till specific height is reached
	// this is done to avoid a soft-fork issue, related to input maturity
	var delete_list []crypto.Hash

	pool.txs.Range(func(k, value interface{}) bool {
		txhash :=  k.(crypto.Hash)
		v := value.(*mempool_object)

		if height > v.Height { // verify all tx in pool for double spending
			if !Verifier(v.Tx) { // this tx could not verified against specific height
				rlog.Warnf("!!MEMPOOL Deleting TX since atleast one of the key images is already consumed  txid=%s chain height %d tx height %d", txhash, height, v.Height)
				delete_list = append(delete_list, txhash)
			}
		}

		// delete transactions whose keyimages have been mined already
		if v.Height < height {
			if !Verifier(v.Tx) { // this tx is already spent and thus can never be mined
				delete_list = append(delete_list, txhash)
			}
		}
		// expire TX which could not be mined in 1 day 86400 hrs
		if v.Height == 0 && (v.Added+86400) <= uint64(time.Now().UTC().Unix()) {
			rlog.Warnf("!! MEMPOOL Deleting TX since TX could not be mined in 1 day txid=%s chain height %d tx height %d", txhash, height, v.Height)
			delete_list = append(delete_list, txhash)
		}
		// give 7 days time to alt chain txs
		if v.Height != 0 && (v.Added+(86400*7)) <= uint64(time.Now().UTC().Unix()) {
			rlog.Warnf("!! MEMPOOL Deleting alt-chain TX since TX could not be mined in 7 days txid=%s chain height %d tx height %d", txhash, height, v.Height)
			delete_list = append(delete_list, txhash)
		}
		return true
	})


	for i := range delete_list {
		pool.Mempool_Delete_TX(delete_list[i])
	}
}

func (pool *Mempool) Shutdown() {
	//TODO save mempool tx somewhere

	close(pool.Exit_Mutex) // stop relaying

	pool.Lock()
	defer pool.Unlock()

	mempool_file := filepath.Join(globals.GetDataDirectory(), "mempool.json")

	// collect all txs in pool and serialize them and store them
	var objects []mempool_object


	pool.txs.Range(func(k, value interface{}) bool {
		v := value.(*mempool_object)
		objects = append(objects, *v)
		return true
	})

	/*for _, v := range pool.txs {
		objects = append(objects, *v)
	}*/

	var file, err = os.Create(mempool_file)
	if err == nil {
		defer file.Close()
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "\t")
		err = encoder.Encode(objects)

		if err != nil {
			loggerpool.Warnf("Error marshaling mempool data err %s", err)
		}

	} else {
		loggerpool.Warnf("Error creating new file to store mempool data file %s err %s", mempool_file, err)
	}

	loggerpool.Infof("Succesfully saved %d txs to file", (len(objects)))

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
func (pool *Mempool) Mempool_Add_TX(tx *transaction.Transaction, Height uint64) (result bool) {
	result = false
	pool.Lock()
	defer pool.Unlock()

	var object mempool_object

	tx_hash := crypto.Hash(tx.GetHash())

	// check if tx already exists, skip it
	if _, ok := pool.txs.Load(tx_hash); ok {
		//rlog.Debugf("Pool already contains %s, skipping", tx_hash)
		return false
	}

	// we should also extract all key images and add them to have multiple pending
	for i := 0; i < len(tx.Vin); i++ {
		if _, ok := pool.key_images.Load(tx.Vin[i].(transaction.Txin_to_key).K_image); ok {
			rlog.Warnf("TX using inputs  which have already been used, Possible Double spend attack rejected txid %s kimage %s",  tx_hash,
				 tx.Vin[i].(transaction.Txin_to_key).K_image)
			return false
		}
	}

	// add all the key images to check double spend attack within the pool
	for i := 0; i < len(tx.Vin); i++ {
		pool.key_images.Store(tx.Vin[i].(transaction.Txin_to_key).K_image,true) // add element to map for next check
	}

	// we are here means we can add it to pool
	object.Tx = tx
	object.Height = Height
	object.Added = uint64(time.Now().UTC().Unix())

	object.Size = uint64(len(tx.Serialize()))
	object.FEEperBYTE = tx.RctSignature.Get_TX_Fee() / object.Size

	pool.txs.Store(tx_hash,&object)
	pool.modified = true // pool has been modified

	//pool.sort_list() // sort and update pool list

	return true
}

// check whether a tx exists in the pool
func (pool *Mempool) Mempool_TX_Exist(txid crypto.Hash) (result bool) {
	//pool.Lock()
	//defer pool.Unlock()

	if _, ok := pool.txs.Load(txid); ok {
		return true
	}
	return false
}

// check whether a keyimage exists in the pool
func (pool *Mempool) Mempool_Keyimage_Spent(ki crypto.Hash) (result bool) {
	//pool.Lock()
	//defer pool.Unlock()

	if _, ok := pool.key_images.Load(ki); ok {
		return true
	}
	return false
}

// delete specific tx from pool and return it
// if nil is returned Tx was not found in pool
func (pool *Mempool) Mempool_Delete_TX(txid crypto.Hash) (tx *transaction.Transaction) {
	//pool.Lock()
	//defer pool.Unlock()

	var ok bool
	var objecti interface{}

	// check if tx already exists, skip it
	if objecti, ok = pool.txs.Load(txid); !ok {
		rlog.Warnf("Pool does NOT contain %s, returning nil", txid)
		return nil
	}



	// we reached here means, we have the tx remove it from our list, do maintainance cleapup and discard it
	object := objecti.(*mempool_object)
	pool.txs.Delete(txid)

	// remove all the key images
	for i := 0; i < len(object.Tx.Vin); i++ {
		pool.key_images.Delete(object.Tx.Vin[i].(transaction.Txin_to_key).K_image)
	}

	//pool.sort_list()     // sort and update pool list
	pool.modified = true // pool has been modified
	return object.Tx     // return the tx
}

// get specific tx from mem pool without removing it
func (pool *Mempool) Mempool_Get_TX(txid crypto.Hash) (tx *transaction.Transaction) {
//	pool.Lock()
//	defer pool.Unlock()

	var ok bool
	var objecti interface{}

	if objecti, ok = pool.txs.Load(txid); !ok {
		//loggerpool.Warnf("Pool does NOT contain %s, returning nil", txid)
		return nil
	}

	// we reached here means, we have the tx, return the pointer back
	//object := pool.txs[txid]
	object := objecti.(*mempool_object)

	return object.Tx
}

// return list of all txs in pool
func (pool *Mempool) Mempool_List_TX() []crypto.Hash {
//	pool.Lock()
//	defer pool.Unlock()

	var list []crypto.Hash

	pool.txs.Range(func(k, value interface{}) bool {
		 txhash  := k.(crypto.Hash)
		//v := value.(*mempool_object)
		//objects = append(objects, *v)
		 list = append(list,txhash)
		return true
	})

	//pool.sort_list() // sort and update pool list

	// list should be as big as spurce list
	//list := make([]crypto.Hash, len(pool.sorted_by_fee), len(pool.sorted_by_fee))
	//copy(list, pool.sorted_by_fee) // return list sorted by fees

	return list
}

// passes back sorting information and length information for easier new block forging
func (pool *Mempool) Mempool_List_TX_SortedInfo() []TX_Sorting_struct {
//	pool.Lock()
//	defer pool.Unlock()

	_, data := pool.sort_list() // sort and update pool list
	return data

/*	// list should be as big as spurce list
	list := make([]TX_Sorting_struct, len(pool.sorted), len(pool.sorted))
	copy(list, pool.sorted) // return list sorted by fees

	return list
	*/
}

// print current mempool txs
// TODO add sorting
func (pool *Mempool) Mempool_Print() {
	pool.Lock()
	defer pool.Unlock()

	var klist []crypto.Hash
	var vlist []*mempool_object

	pool.txs.Range(func(k, value interface{}) bool {
		 txhash  := k.(crypto.Hash)
		v := value.(*mempool_object)
		//objects = append(objects, *v)
		 klist = append(klist,txhash)
		 vlist = append(vlist,v)

		return true
	})


	fmt.Printf("Total TX in mempool = %d\n", len(klist))
	fmt.Printf("%20s  %14s %7s %7s %6s %32s\n", "Added", "Last Relayed", "Relayed", "Size", "Height", "TXID")

	for i := range klist {
		k := klist[i]
		v := vlist[i]
		fmt.Printf("%20s  %14s %7d %7d %6d %32s\n", time.Unix(int64(v.Added), 0).UTC().Format(time.RFC3339), time.Duration(v.RelayedAt)*time.Second, v.Relayed,
			len(v.Tx.Serialize()), v.Height, k)
	}
}

// flush mempool
func (pool *Mempool) Mempool_flush() {
	var list []crypto.Hash

	pool.txs.Range(func(k, value interface{}) bool {
		 txhash  := k.(crypto.Hash)
		//v := value.(*mempool_object)
		//objects = append(objects, *v)
		 list = append(list,txhash)
		return true
	})

	fmt.Printf("Total TX in mempool = %d \n", len(list))
	fmt.Printf("Flushing mempool \n")


	for i := range list {
		pool.Mempool_Delete_TX(list[i])
	}
}

// sorts the pool internally
// this function assummes lock is already taken
// ??? if we  selecting transactions randomly, why to keep them sorted
func (pool *Mempool) sort_list() ([]crypto.Hash, []TX_Sorting_struct){

	data := make([]TX_Sorting_struct,0,512) // we are rarely expectingmore than this entries in mempool
	// collect data from pool for sorting

	pool.txs.Range(func(k, value interface{}) bool {
		 txhash  := k.(crypto.Hash)
		 v := value.(*mempool_object)
		 if v.Height <= pool.height {
			data = append(data, TX_Sorting_struct{Hash: txhash, FeesPerByte: v.FEEperBYTE, Size: v.Size})
		}
		return true
	})


	// inverted comparision sort to do descending sort
	sort.SliceStable(data, func(i, j int) bool { return data[i].FeesPerByte > data[j].FeesPerByte })

	sorted_list := make([]crypto.Hash,0,len(data))
	//pool.sorted_by_fee = pool.sorted_by_fee[:0] // empty old slice

	for i := range data {
		sorted_list = append(sorted_list, data[i].Hash)
	}
	//pool.sorted = data
	return sorted_list, data

}

type p2p_TX_Relayer func(*transaction.Transaction, uint64) int // function type, exported in p2p but cannot use due to cyclic dependency

// this tx relayer keeps on relaying tx and cleaning mempool
// if a tx has been relayed less than 10 peers, tx relaying is agressive
// otherwise the tx are relayed every 30 minutes, till it has been relayed to 20
// then the tx is relayed every 3 hours, just in case
func (pool *Mempool) Relayer_and_Cleaner() {

	for {

		select {
		case <-pool.Exit_Mutex:
			return
		case <-time.After(4000 * time.Millisecond):

		}

		sent_count := 0

		//pool.Lock()

		//loggerpool.Warnf("send Pool lock taken")


		pool.txs.Range(func(ktmp, value interface{}) bool {
		 k  := ktmp.(crypto.Hash)
		 v := value.(*mempool_object)
	
			select { // exit fast of possible
			case <-pool.Exit_Mutex:
				pool.Unlock()
				return false
			default:
			}

			if sent_count > 200 { // send a burst of 200 txs max in 1 go
				return false;
			}

			if v.Height <= pool.height { // only carry out activities for valid txs

				if (v.Relayed < 3 && (time.Now().Unix()-v.RelayedAt) > 60) || // relay it now
					(v.Relayed >= 4 && v.Relayed <= 20 && (time.Now().Unix()-v.RelayedAt) > 1200) || // relay it now
					((time.Now().Unix() - v.RelayedAt) > (3600 * 3)) {
					if pool.P2P_TX_Relayer != nil {

						relayed_count := pool.P2P_TX_Relayer(v.Tx, 0)
						//relayed_count := 0
						if relayed_count > 0 {
							v.Relayed += relayed_count

							sent_count++

							//loggerpool.Debugf("%d  %d\n",time.Now().Unix(), v.RelayedAt)
							rlog.Tracef(1,"Relayed %s to %d peers (%d %d)", k, relayed_count, v.Relayed, (time.Now().Unix() - v.RelayedAt))
							v.RelayedAt = time.Now().Unix()
							//loggerpool.Debugf("%d  %d",time.Now().Unix(), v.RelayedAt)
						}
					}
				}
			}
			 
		return true
	})


		// loggerpool.Warnf("send Pool lock released")
		//pool.Unlock()
	}
}

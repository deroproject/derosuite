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

package blockchain

// This file runs the core consensus protocol
// please think before randomly editing for after effects
// We must not call any packages that can call panic
// NO Panics or FATALs please

import "os"
import "fmt"
import "sort"
import "sync"
import "bufio"
import "time"
import "bytes"
import "runtime"
import "math/big"
import "sync/atomic"
import "runtime/debug"

import "golang.org/x/crypto/sha3"
import "github.com/romana/rlog"
import log "github.com/sirupsen/logrus"
import "github.com/golang/groupcache/lru"
import hashicorp_lru "github.com/hashicorp/golang-lru"

import "github.com/deroproject/derosuite/config"
import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/errormsg"
import "github.com/prometheus/client_golang/prometheus"

//import "github.com/deroproject/derosuite/address"
import "github.com/deroproject/derosuite/emission"
import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/storage"
import "github.com/deroproject/derosuite/crypto/ringct"
import "github.com/deroproject/derosuite/block"
import "github.com/deroproject/derosuite/transaction"
import "github.com/deroproject/derosuite/checkpoints"
import "github.com/deroproject/derosuite/metrics"
import "github.com/deroproject/derosuite/blockchain/mempool"
import "github.com/deroproject/derosuite/blockchain/inputmaturity"

// all components requiring access to blockchain must use , this struct to communicate
// this structure must be update while mutex
type Blockchain struct {
	store       storage.Store // interface to storage layer
	Height      int64         // chain height is always 1 more than block
	height_seen int64         // height seen on peers
	Top_ID      crypto.Hash   // id of the top block
	//Tips              map[crypto.Hash]crypto.Hash // current tips
	dag_unsettled              map[crypto.Hash]bool // current unsettled dag
	dag_past_unsettled_cache   *lru.Cache
	dag_future_unsettled_cache *lru.Cache
	lrucache_workscore         *lru.Cache
	lrucache_fullorder         *lru.Cache // keeps full order for  tips upto a certain height

	MINING_BLOCK bool // used to pause mining

	Difficulty        uint64 // current cumulative difficulty
	Median_Block_Size uint64 // current median block size
	Mempool           *mempool.Mempool
	Exit_Event        chan bool // blockchain is shutting down and we must quit ASAP

	Top_Block_Median_Size uint64 // median block size of current top block
	Top_Block_Base_Reward uint64 // top block base reward

	checkpints_disabled bool // are checkpoints disabled
	simulator           bool // is simulator mode

	P2P_Block_Relayer func(*block.Complete_Block, uint64) // tell p2p to broadcast any block this daemon hash found

	sync.RWMutex
}

var logger *log.Entry

//var Exit_Event = make(chan bool) // causes all threads to exit

// All blockchain activity is store in a single

/* do initialisation , setup storage, put genesis block and chain in store
   This is the first component to get up
   Global parameters are picked up  from the config package
*/

func Blockchain_Start(params map[string]interface{}) (*Blockchain, error) {

	var err error
	var chain Blockchain

	logger = globals.Logger.WithFields(log.Fields{"com": "BLKCHAIN"})
	logger.Infof("Initialising blockchain")
	//init_static_checkpoints()           // init some hard coded checkpoints
	checkpoints.LoadCheckPoints(logger) // load checkpoints from file if provided

	if params["--simulator"] == true { // simulator always uses boltdb backend
		chain.store = storage.Bolt_backend // setup backend
		chain.store.Init(params)           // init backend

	} else {
		if (runtime.GOARCH == "amd64" && !globals.Arguments["--badgerdb"].(bool)) || globals.Arguments["--boltdb"].(bool) {
			chain.store = storage.Bolt_backend // setup backend
			chain.store.Init(params)           // init backend
		} else {
			chain.store = storage.Badger_backend // setup backend
			chain.store.Init(params)             // init backend
		}
	}

	/*

		   logger.Infof("%+v", *storage.MySQL_backend)
		   chain.store = storage.MySQL_backend
		   chain.store.Init(params)
		   logger.Infof("%+v", *storage.MySQL_backend)

		//       xyz := storage.MySQL_backend
		//xyz.Init(params)

		if err != nil {
			logger.Infof("Error Initialising blockchain mysql backend , err %s", err)
			return nil, err
		}

	*/
	//chain.Tips = map[crypto.Hash]crypto.Hash{} // initialize Tips map
	chain.lrucache_workscore = lru.New(8191)  // temporary cache for work caclculation
	chain.lrucache_fullorder = lru.New(20480) // temporary cache for fullorder caclculation

	if globals.Arguments["--disable-checkpoints"] != nil {
		chain.checkpints_disabled = globals.Arguments["--disable-checkpoints"].(bool)
	}

	if params["--simulator"] == true {
		chain.simulator = true // enable simulator mode, this will set hard coded difficulty to 1
	}

	chain.Exit_Event = make(chan bool) // init exit channel

	// init mempool before chain starts
	chain.Mempool, err = mempool.Init_Mempool(params)

	// we need to check mainnet/testnet check whether the genesis block matches the testnet/mainet
	// mean whether the user is trying to use mainnet db with testnet option or vice-versa
	if chain.Block_Exists(nil, config.Mainnet.Genesis_Block_Hash) || chain.Block_Exists(nil, config.Testnet.Genesis_Block_Hash) {

		if globals.IsMainnet() && !chain.Block_Exists(nil, config.Mainnet.Genesis_Block_Hash) {
			logger.Fatalf("Tryng to use a testnet database with mainnet, please add --testnet option")
		}

		if !globals.IsMainnet() && !chain.Block_Exists(nil, config.Testnet.Genesis_Block_Hash) {
			logger.Fatalf("Tryng to use a mainnet database with testnet, please remove --testnet option")
		}

		// check if user is trying to load previous testnet DB with , reject
		if !globals.IsMainnet() && chain.Block_Exists(nil,crypto.HashHexToHash("4dfc6daa5e104250125e0a14b74eca04730fd5bec4e826fa54f791245aa924f2")) { 
			logger.Warnf("Please delete existing testnet DB as testnet has boostrapped")
			return nil, fmt.Errorf("Please delete existing testnet DB.")
		}

	}

	// genesis block not in chain, add it to chain, together with its miner tx
	// make sure genesis is in the store
	bl := Generate_Genesis_Block()
	//if !chain.Block_Exists(globals.Config.Genesis_Block_Hash) {
	if !chain.Block_Exists(nil, bl.GetHash()) {
		//chain.Store_TOP_ID(globals.Config.Genesis_Block_Hash) // store top id , exception of genesis block
		logger.Debugf("Genesis block not in store, add it now")
		var complete_block block.Complete_Block
		//bl := Generate_Genesis_Block()
		complete_block.Bl = &bl

		/*if !chain.Add_Complete_Block(&complete_block) {
			logger.Fatalf("Failed to add genesis block, we can no longer continue")
		}*/

		logger.Infof("Added block successfully")

		//chain.store_Block_Settled(bl.GetHash(),true) // genesis block is always settled

		dbtx, err := chain.store.BeginTX(true)
		if err != nil {
			logger.Warnf("Could NOT add block to chain. Error opening writable TX, err %s", err)
			// return
		}
		chain.Store_BL(dbtx, &bl)

		bl_current_hash := bl.GetHash()
		// store total  reward
		dbtx.StoreUint64(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, bl_current_hash[:], PLANET_MINERTX_REWARD, bl.Miner_TX.Vout[0].Amount)

		// store base reward
		dbtx.StoreUint64(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, bl_current_hash[:], PLANET_BASEREWARD, bl.Miner_TX.Vout[0].Amount)

		// store total generated coins
		// this is hardcoded at initial chain import, keeping original emission schedule
		if globals.IsMainnet(){
				dbtx.StoreUint64(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, bl_current_hash[:], PLANET_ALREADY_GENERATED_COINS, config.MAINNET_HARDFORK_1_TOTAL_SUPPLY)		
			}else{
				dbtx.StoreUint64(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, bl_current_hash[:], PLANET_ALREADY_GENERATED_COINS, config.TESTNET_HARDFORK_1_TOTAL_SUPPLY)		
			}
		

		chain.Store_Block_Topological_order(dbtx, bl.GetHash(), 0) // genesis block is the lowest
		chain.Store_TOPO_HEIGHT(dbtx, 0)                           //
		chain.Store_TOP_HEIGHT(dbtx, 0)

		chain.store_TIPS(dbtx, []crypto.Hash{bl.GetHash()})

		dbtx.Commit()

	}

	//fmt.Printf("Genesis Block should be present at height 0\n")
	/*blocks := chain.Get_Blocks_At_Height(0)
	  fmt.Printf("blocks at height 0 %+v\n", blocks)

	  fmt.Printf("Past of  genesis %+v\n", chain.Get_Block_Past(bl.GetHash()))
	  fmt.Printf("Future of  genesis %+v\n", chain.Get_Block_Future(bl.GetHash()))

	  fmt.Printf("Future of  zero block  %+v\n", chain.Get_Block_Future(ZERO_HASH))
	*/

	// load the chain from the disk
	chain.Initialise_Chain_From_DB()

	//   logger.Fatalf("Testing complete quitting")

	// hard forks must be initialized after chain is up
	init_hard_forks(params)

	go clean_up_valid_cache() // clean up valid cache

	/*  txlist := chain.Mempool.Mempool_List_TX()
	    for i := range txlist {
	       // if fmt.Sprintf("%s", txlist[i]) == "0fe0e7270ba911956e91d9ea099e4d12aa1bce2473d4064e239731bc37acfd86"{
	        logger.Infof("Verifying tx %s %+v", txlist[i], chain.Verify_Transaction_NonCoinbase(chain.Mempool.Mempool_Get_TX(txlist[i])))

	        //}
	        //p2p.Broadcast_Tx(chain.Mempool.Mempool_Get_TX(txlist[i]))
	    }
	*/

	if chain.checkpints_disabled {
		logger.Infof("Internal Checkpoints are disabled")
	} else {
		logger.Debugf("Internal Checkpoints are enabled")
	}

	_ = err

	atomic.AddUint32(&globals.Subsystem_Active, 1) // increment subsystem

	// register the metrics with the metrics registry
	metrics.Registry.MustRegister(blockchain_tx_counter)
	metrics.Registry.MustRegister(mempool_tx_counter)
	metrics.Registry.MustRegister(mempool_tx_count)
	metrics.Registry.MustRegister(block_size)
	metrics.Registry.MustRegister(transaction_size)
	metrics.Registry.MustRegister(block_tx_count)
	metrics.Registry.MustRegister(block_processing_time)

	return &chain, nil
}

// this function is called to read blockchain state from DB
// It is callable at any point in time
func (chain *Blockchain) Initialise_Chain_From_DB() {
	chain.Lock()
	defer chain.Unlock()

	// locate top block
	/*
		chain.Top_ID = chain.Load_TOP_ID()
		//chain.Height = (chain.Load_Height_for_BL_ID(chain.Top_ID) + 1)
		chain.Difficulty = chain.Get_Difficulty()
		chain.Top_Block_Median_Size = chain.Get_Median_BlockSize_At_Block(chain.Top_ID)
		chain.Top_Block_Base_Reward = chain.Load_Block_Reward(chain.Top_ID)
		// set it so it is not required to be calculated frequently
		chain.Median_Block_Size = chain.Get_Median_BlockSize_At_Block(chain.Get_Top_ID())
		if chain.Median_Block_Size < config.CRYPTONOTE_BLOCK_GRANTED_FULL_REWARD_ZONE {
			chain.Median_Block_Size = config.CRYPTONOTE_BLOCK_GRANTED_FULL_REWARD_ZONE
		}
	*/
	// find the tips from the chain , first by reaching top height
	// then downgrading to top-10 height
	// then reworking the chain to get the tip
	best_height := chain.Load_TOP_HEIGHT(nil)
	chain.Height = best_height

	// reload tips from disk
	tips := chain.load_TIPS(nil)

	// get dag unsettled, it's only possible when we have the tips
	// chain.dag_unsettled = chain.Get_DAG_Unsettled() // directly off the disk

	logger.Infof("Chain Tips  %+v Height %d", tips, chain.Height)

}

// before shutdown , make sure p2p is confirmed stopped
func (chain *Blockchain) Shutdown() {

	chain.Lock()            // take the lock as chain is no longer in unsafe mode
	close(chain.Exit_Event) // send signal to everyone we are shutting down

	chain.Mempool.Shutdown() // shutdown mempool first

	logger.Infof("Stopping Blockchain")
	chain.store.Shutdown()
	atomic.AddUint32(&globals.Subsystem_Active, ^uint32(0)) // this decrement 1 fom subsystem
}

// get top unstable height
// this is obtained by  getting the highest topo block and getting its height
func (chain *Blockchain) Get_Height() int64 {

	topo_height := chain.Load_TOPO_HEIGHT(nil)

	blid, err := chain.Load_Block_Topological_order_at_index(nil, topo_height)
	if err != nil {
		logger.Warnf("Cannot get block  at topoheight %d err: %s", topo_height, err)
		return 0
	}

	height := chain.Load_Height_for_BL_ID(nil, blid)

	//return atomic.LoadUint64(&chain.Height)
	return height
}

// get height where chain is now stable
func (chain *Blockchain) Get_Stable_Height() int64 {

	dbtx, err := chain.store.BeginTX(false)
	if err != nil {
		logger.Warnf("Could NOT add block to chain. Error opening writable TX, err %s", err)
		return -1
	}

	defer dbtx.Rollback()

	tips := chain.Get_TIPS()
	base, base_height := chain.find_common_base(dbtx, tips)
	_ = base

	return int64(base_height)
}

// we should be holding lock at this time, atleast read only
func (chain *Blockchain) Get_TIPS() (tips []crypto.Hash) {
	return chain.load_TIPS(nil)
}

func (chain *Blockchain) Get_Top_ID() crypto.Hash {
	topo_height := chain.Load_TOPO_HEIGHT(nil)

	blid, err := chain.Load_Block_Topological_order_at_index(nil, topo_height)
	if err != nil {
		logger.Warnf("Cannot get block  at topoheight %d err: %s", topo_height, err)
		return blid
	}

	return blid
}

func (chain *Blockchain) Get_Difficulty() uint64 {
	return chain.Get_Difficulty_At_Tips(nil, chain.Get_TIPS()).Uint64()
}

func (chain *Blockchain) Get_Cumulative_Difficulty() uint64 {

	/*
		topo_height := chain.Load_TOPO_HEIGHT(nil)

		blid, err := chain.Load_Block_Topological_order_at_index(nil, topo_height)
		if err != nil {
			logger.Warnf("Cannot get block  at topoheight %d err: %s",topo_height,err)
			return 0
		}

		past := chain.Get_Block_Past(nil,blid)
		return  chain.Get_Difficulty_At_Tips(nil, past,uint64(uint64(time.Now().UTC().Unix())+config.BLOCK_TIME)).Uint64()

	*/

	return 0 //chain.Load_Block_Cumulative_Difficulty(chain.Top_ID)
}

func (chain *Blockchain) Get_Median_Block_Size() uint64 { // get current cached median size
	return chain.Median_Block_Size
}

func (chain *Blockchain) Get_Network_HashRate() uint64 {
	return chain.Get_Difficulty() / chain.Get_Current_BlockTime()
}

// confirm whether the block exist in the data
// this only confirms whether the block has been downloaded
// a separate check is required, whether the block is valid ( satifies PoW and other conditions)
// we will not add a block to store, until it satisfies PoW
func (chain *Blockchain) Block_Exists(dbtx storage.DBTX, h crypto.Hash) bool {
	_, err := chain.Load_BL_FROM_ID(dbtx, h)
	if err == nil {
		return true
	}
	return false
}

// various counters/gauges which track a numer of metrics
// such as number of txs, number of inputs, number of outputs
// mempool total addition, current mempool size
// block processing time etcs

// Try it once more, this time with a help string.
var blockchain_tx_counter = prometheus.NewCounter(prometheus.CounterOpts{
	Name: "blockchain_tx_counter",
	Help: "Number of tx mined",
})

var mempool_tx_counter = prometheus.NewCounter(prometheus.CounterOpts{
	Name: "mempool_tx_counter",
	Help: "Total number of tx added in mempool",
})
var mempool_tx_count = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "mempool_tx_count",
	Help: "Number of tx in mempool at this point",
})

//  track block size about 2 MB
var block_size = prometheus.NewHistogram(prometheus.HistogramOpts{
	Name:    "block_size_byte",
	Help:    "Block size in byte (complete)",
	Buckets: prometheus.LinearBuckets(0, 102400, 10), // start block size 0, each 1 KB step,  2048 such buckets .
})

//  track transaction size upto 500 KB
var transaction_size = prometheus.NewHistogram(prometheus.HistogramOpts{
	Name:    "tx_size_byte",
	Help:    "TX size in byte",
	Buckets: prometheus.LinearBuckets(0, 10240, 16), // start 0  byte, each 1024 byte,  512 such buckets.
})

//  number of tx per block
var block_tx_count = prometheus.NewHistogram(prometheus.HistogramOpts{
	Name:    "block_tx_count",
	Help:    "Number of TX in the block",
	Buckets: prometheus.LinearBuckets(0, 20, 25), // start 0  byte, each 1024 byte,  1024 such buckets.
})

//
var block_processing_time = prometheus.NewHistogram(prometheus.HistogramOpts{
	Name:    "block_processing_time_ms",
	Help:    "Block processing time milliseconds",
	Buckets: prometheus.LinearBuckets(0, 100, 20), // start 0  ms, each 100 ms,  200 such buckets.
})

// this is the only entrypoint for new txs in the chain
// add a transaction to MEMPOOL,
// verifying everything  means everything possible
// this only change mempool, no DB changes
func (chain *Blockchain) Add_TX_To_Pool(tx *transaction.Transaction) (result bool) {

	// chain lock is no longer required as we only do readonly processing
	//	chain.Lock()
	//	defer chain.Unlock()

	dbtx, err := chain.store.BeginTX(false)
	if err != nil {
		logger.Warnf("Could NOT create DB transaction  err %s", err)
		return true // just make it rebroadcast
	}

	// track counter for the amount of mempool tx
	defer mempool_tx_count.Set(float64(len(chain.Mempool.Mempool_List_TX())))

	defer dbtx.Rollback()

	txhash := tx.GetHash()

	// Coin base TX can not come through this path
	if tx.IsCoinbase() {
		logger.WithFields(log.Fields{"txid": txhash}).Warnf("TX rejected  coinbase tx cannot appear in mempool")
		return false
	}

	// quick check without calculating everything whether tx is in pool, if yes we do nothing
	if chain.Mempool.Mempool_TX_Exist(txhash) {
		rlog.Tracef(2,"TX %s rejected Already in MEMPOOL", txhash)
		return true
	}

	hf_version := chain.Get_Current_Version_at_Height(chain.Get_Height())

	// TODO if someone is relaying existing tx again and again, we need to quickly figure it and avoid expensive verification
	// a simple technique seems to be to do key image verification for double spend, if it's reject
	// this test is placed to avoid ring signature verification cost for faulty tx as early as possible
	if !chain.Verify_Transaction_NonCoinbase_DoubleSpend_Check(dbtx, tx) { // BUG BUG BUG we must use dbtx to confirm
		rlog.Tracef(2,"TX %s rejected due to double spending", txhash)
		return false
	}

	// if TX is too big, then it cannot be mined due to fixed block size, reject such TXs here
	// currently, limits are  as per consensus
	if uint64(len(tx.Serialize())) > config.CRYPTONOTE_MAX_TX_SIZE {
		logger.WithFields(log.Fields{"txid": txhash}).Warnf("TX rejected  Size %d byte Max possible %d", len(tx.Serialize()), config.CRYPTONOTE_MAX_TX_SIZE)
		return false
	}

	// check whether enough fees is provided in the transaction
	calculated_fee := chain.Calculate_TX_fee(hf_version, uint64(len(tx.Serialize())))
	provided_fee := tx.RctSignature.Get_TX_Fee() // get fee from tx

	if calculated_fee > provided_fee { // 2 % margin see blockchain.cpp L 2913
		logger.WithFields(log.Fields{"txid": txhash}).Warnf("TX rejected due to low fees  provided fee %d calculated fee %d", provided_fee, calculated_fee)

		rlog.Warnf("TX  %s rejected due to low fees  provided fee %d calculated fee %d", txhash, provided_fee, calculated_fee)
		return false
	}

	if chain.Verify_Transaction_NonCoinbase(dbtx, hf_version, tx) && chain.Verify_Transaction_NonCoinbase_DoubleSpend_Check(dbtx, tx) {
		if chain.Mempool.Mempool_Add_TX(tx, 0) { // new tx come with 0 marker
			rlog.Tracef(2,"Successfully added tx %s to pool", txhash)

			mempool_tx_counter.Inc()
			return true
		} else {
			rlog.Tracef(2,"TX %s rejected by pool", txhash)
			return false
		}
	}

	rlog.Warnf("Incoming TX %s could not be verified", txhash)
	return false

}

// structure used to rank/sort  blocks on a number of factors
type BlockScore struct {
	BLID crypto.Hash
	// Weight uint64
	Height                int64    // block height
	Cumulative_Difficulty *big.Int // used to score blocks on cumulative difficulty
}

// Heighest node weight is ordered first,  the condition is reverted see eg. at https://golang.org/pkg/sort/#Slice
//  if weights are equal, nodes are sorted by their block ids which will never collide , hopefullly
// block ids are sorted by lowest byte first diff
func sort_descending_by_cumulative_difficulty(tips_scores []BlockScore) {

	sort.Slice(tips_scores, func(i, j int) bool {
		if tips_scores[i].Cumulative_Difficulty.Cmp(tips_scores[j].Cumulative_Difficulty) != 0 { // if diffculty mismatch use them

			if tips_scores[i].Cumulative_Difficulty.Cmp(tips_scores[j].Cumulative_Difficulty) > 0 { // if i diff >  j diff
				return true
			} else {
				return false
			}

		} else {
			return bytes.Compare(tips_scores[i].BLID[:], tips_scores[j].BLID[:]) == -1
		}
	})
}

func sort_ascending_by_height(tips_scores []BlockScore) {

	// base is the lowest height
	sort.Slice(tips_scores, func(i, j int) bool { return tips_scores[i].Height < tips_scores[j].Height })

}

// this will sort the tips based on cumulative difficulty and/or block ids
// the tips will sorted in descending order
func (chain *Blockchain) SortTips(dbtx storage.DBTX, tips []crypto.Hash) (sorted []crypto.Hash) {
	if len(tips) == 0 {
		panic("tips cannot be 0")
	}
	if len(tips) == 1 {
		sorted = []crypto.Hash{tips[0]}
		return
	}

	tips_scores := make([]BlockScore, len(tips), len(tips))
	for i := range tips {
		tips_scores[i].BLID = tips[i]
		tips_scores[i].Cumulative_Difficulty = chain.Load_Block_Cumulative_Difficulty(dbtx, tips[i])
	}

	sort_descending_by_cumulative_difficulty(tips_scores)

	for i := range tips_scores {
		sorted = append(sorted, tips_scores[i].BLID)
	}
	return
}


// side blocks are blocks which lost the race the to become part
// of main chain, but there transactions are honoured,
// they are given 67 % reward
// a block is a side block if it satisfies the following condition
// if  block height   is less than or equal to height of past 8 topographical blocks
// this is part of consensus rule
// this is the topoheight of this block itself
func (chain *Blockchain) isblock_SideBlock(dbtx storage.DBTX, blid crypto.Hash, block_topoheight int64) (result bool) {

	if block_topoheight <= 2 {
		return false
	}
	// lower reward for byzantine behaviour
	// for as many block as added
	block_height := chain.Load_Height_for_BL_ID(dbtx, blid)

	counter := int64(0)
	for i := block_topoheight - 1; i >= 0 && counter < config.STABLE_LIMIT; i-- {
		counter++

		previous_blid, err := chain.Load_Block_Topological_order_at_index(dbtx, i)
		if err != nil {
			panic("Could not load block from previous order")
		}
		// height of previous topo ordered block
		previous_height := chain.Load_Height_for_BL_ID(dbtx, previous_blid)

		if block_height <= previous_height { // lost race (or byzantine behaviour)
			return true // give only 67 % reward
		}

	}

	return false
}

// this is the only entrypoint for new / old blocks even for genesis block
// this will add the entire block atomically to the chain
// this is the only function which can add blocks to the chain
// this is exported, so ii can be fed new blocks by p2p layer
// genesis block is no different
// TODO: we should stop mining while adding the new block
func (chain *Blockchain) Add_Complete_Block(cbl *block.Complete_Block) (err error, result bool) {

	var block_hash crypto.Hash
	chain.Lock()
	defer chain.Unlock()
	result = false

	dbtx, err := chain.store.BeginTX(true)
	if err != nil {
		logger.Warnf("Could NOT add block to chain. Error opening writable TX, err %s", err)
		return errormsg.ErrInvalidStorageTX, false
	}

	chain.MINING_BLOCK = true

	processing_start := time.Now()

	//old_top := chain.Load_TOP_ID() // store top as it may change
	defer func() {

		// safety so if anything wrong happens, verification fails
		if r := recover(); r != nil {
			logger.Warnf("Recovered while adding new block, Stack trace below block_hash %s", block_hash)
			logger.Warnf("Stack trace  \n%s", debug.Stack())
			result = false
			err = errormsg.ErrPanic
		}

		chain.MINING_BLOCK = false

		if result == true { // block was successfully added, commit it atomically
			dbtx.Commit()

			rlog.Infof("Block successfully acceppted by chain %s", block_hash)

			// gracefully try to instrument
			func() {
				defer func() {
					if r := recover(); r != nil {
						rlog.Warnf("Recovered while instrumenting")
						rlog.Warnf("Stack trace \n%s", debug.Stack())

					}
				}()
				blockchain_tx_counter.Add(float64(len(cbl.Bl.Tx_hashes)))
				block_tx_count.Observe(float64(len(cbl.Bl.Tx_hashes)))
				block_processing_time.Observe(float64(time.Now().Sub(processing_start).Round(time.Millisecond) / 1000000))

				// tracks counters for tx_size

				{
					complete_block_size := 0
					for i := 0; i < len(cbl.Txs); i++ {
						tx_size := len(cbl.Txs[i].Serialize())
						complete_block_size += tx_size
						transaction_size.Observe(float64(tx_size))
					}
					block_size.Observe(float64(complete_block_size))
				}
			}()

			//dbtx.Sync() // sync the DB to disk after every execution of this function

			//if old_top != chain.Load_TOP_ID() { // if top has changed, discard mining templates and start afresh
			// TODO discard mining templates or something else, if top chnages requires some action

			//}
		} else {
			dbtx.Rollback() // if block could not be added, rollback all changes to previous block
			rlog.Infof("Block rejected by chain %s err %s", block_hash, err)
		}
	}()

	bl := cbl.Bl // small pointer to block

	// first of all lets do some quick checks
	// before doing extensive checks
	result = false

	block_hash = bl.GetHash()
	block_logger := logger.WithFields(log.Fields{"blid": block_hash})

	// check if block already exist skip it
	if chain.Block_Exists(dbtx, block_hash) {
		block_logger.Debugf("block already in chain skipping it ")
		return errormsg.ErrAlreadyExists, false
	}

	// only 3 tips allowed in block
	if len(bl.Tips) >= 4 {
		rlog.Warnf("More than 3 tips present in block %s rejecting", block_hash)
		return errormsg.ErrPastMissing, false
	}

	// check whether the tips exist in our chain, if not reject
	if chain.Get_Height() > 0 {
		for i := range bl.Tips {
			if !chain.Block_Exists(dbtx, bl.Tips[i]) {
				rlog.Warnf("Tip  %s  is NOT present in chain current block %s, skipping it till we get a parent", bl.Tips[i], block_hash)
				return errormsg.ErrPastMissing, false
			}
		}
	}

	

	block_height := chain.Calculate_Height_At_Tips(dbtx, bl.Tips)

	if block_height == 0 && bl.GetHash() != globals.Config.Genesis_Block_Hash {
		block_logger.Warnf("There can can be only one genesis block, reject it, len of tips(%d)", len(bl.Tips))
		return errormsg.ErrInvalidBlock, false
	}
	if block_height < chain.Get_Stable_Height() {
		rlog.Warnf("Block %s rejected since it is stale stable height %d  block height %d", bl.GetHash(), chain.Get_Stable_Height(), block_height)
		return errormsg.ErrInvalidBlock, false
	}

	// use checksum to quick jump
	if chain.checkpints_disabled == false && checkpoints.IsCheckSumKnown(chain.BlockCheckSum(cbl)) {
		rlog.Debugf("Skipping Deep Checks for block %s ", block_hash)
		goto skip_checks
	} else {
		rlog.Debugf("Deep Checks for block %s ", block_hash)
	}

	// version 1 blocks ( old chain) should NOT be mined by used
	// they should use hard coded checkpoints
	if chain.checkpints_disabled == false && chain.Get_Current_Version_at_Height(block_height) == 1 {
		logger.Warnf("v1 blocks cannot be mined (these are imported blocks), rejecting")
		return errormsg.ErrInvalidBlock, false
	}

	/*

		// check  a small list 100 hashes whether they have been reached
		if IsCheckPointKnown_Static(block_hash, chain.Load_Height_for_BL_ID(bl.Prev_Hash)+1) {
			logger.Infof("Static Checkpoint reached at height %d", chain.Load_Height_for_BL_ID(bl.Prev_Hash)+1)
		}

		rlog.Tracef(1, "Checking Known checkpoint %s at height %d", block_hash, chain.Load_Height_for_BL_ID(bl.Prev_Hash)+1)

		//if we have checkpoints embedded, they must match
		// until user disables this check
		// skip checkpoint check for genesis block
		if block_hash != globals.Config.Genesis_Block_Hash {
			if chain.checkpints_disabled == false && checkpoints.Length() > chain.Load_Height_for_BL_ID(bl.Prev_Hash)+1 && !checkpoints.IsCheckPointKnown(block_hash, chain.Load_Height_for_BL_ID(bl.Prev_Hash)+1) {
				block_logger.Warnf("Block hash mismatch with checkpoint height %d", chain.Load_Height_for_BL_ID(bl.Prev_Hash)+1)
				return
			}



		}
	*/

	// make sure time is NOT too much into future, we have 2 seconds of margin here
	// some OS have trouble syncing with more than 1 sec granularity
	// if clock diff is more than   2 secs, reject the block
	if bl.Timestamp > (uint64(time.Now().UTC().Unix()) + config.CRYPTONOTE_FUTURE_TIME_LIMIT) {
		block_logger.Warnf("Rejecting Block, timestamp is too much into future, make sure that system clock is correct")
		return errormsg.ErrFutureTimestamp, false
	}

	// verify that the clock is not being run in reverse
	// the block timestamp cannot be less than any of the parents
	for i := range bl.Tips {
		if uint64(chain.Load_Block_Timestamp(dbtx, bl.Tips[i])) > bl.Timestamp {
			block_logger.Warnf("Block timestamp is  less than its parent, rejecting block")
			return errormsg.ErrInvalidTimestamp, false
		}
	}

	//logger.Infof("current version %d  height %d", chain.Get_Current_Version_at_Height( 2500), chain.Calculate_Height_At_Tips(dbtx, bl.Tips))
	// check whether the major version ( hard fork) is valid
	if !chain.Check_Block_Version(dbtx, bl) {
		block_logger.Warnf("Rejecting !! Block has invalid fork version actual %d expected %d", bl.Major_Version, chain.Get_Current_Version_at_Height(chain.Calculate_Height_At_Tips(dbtx, bl.Tips)))
		return errormsg.ErrInvalidBlock, false
	}

	// verify whether the tips are unreachable from one another
	if !chain.VerifyNonReachability(dbtx, bl) {
		block_logger.Warnf("Rejecting !! Block has invalid reachability")
		return errormsg.ErrInvalidBlock, false

	}

	// if the block is referencing any past tip too distant into main chain discard now
	// TODO FIXME this need to computed
	for i := range bl.Tips {
		rusty_tip_base_distance := chain.calculate_mainchain_distance(dbtx, bl.Tips[i])

		// tips of deviation >= 8 will rejected
		if (int64(chain.Get_Height()) - rusty_tip_base_distance) >= config.STABLE_LIMIT {
			block_logger.Warnf("Rusty TIP  mined by ROGUE miner discarding block %s  best height %d deviation %d rusty_tip %d", bl.Tips[i], chain.Get_Height(), (int64(chain.Get_Height()) - rusty_tip_base_distance), rusty_tip_base_distance)
			return errormsg.ErrInvalidBlock, false
		}
	}

	// verify difficulty of tips provided
	if len(bl.Tips) > 1 {
		best_tip := chain.find_best_tip_cumulative_difficulty(dbtx, bl.Tips)
		for i := range bl.Tips {
			if best_tip != bl.Tips[i] {
				if !chain.validate_tips(dbtx, best_tip, bl.Tips[i]) { // reference is first
					block_logger.Warnf("Rusty tip mined by ROGUE miner, discarding block")
					return errormsg.ErrInvalidBlock, false
				}
			}
		}
	}

	// check whether the block crosses the size limit
	// block size is calculate by adding all the txs
	// block header/miner tx is excluded, only tx size if calculated
	{
		block_size := 0
		for i := 0; i < len(cbl.Txs); i++ {
			block_size += len(cbl.Txs[i].Serialize())
			if uint64(block_size) >= config.CRYPTONOTE_MAX_BLOCK_SIZE {
				block_logger.Warnf("Block is bigger than max permitted, Rejecting it Actual %d MAX %d ", block_size, config.CRYPTONOTE_MAX_BLOCK_SIZE)
				return errormsg.ErrInvalidSize, false
			}
		}
	}

	//logger.Infof("pow hash %s height %d", bl.GetPoWHash(), block_height)

	// Verify Blocks Proof-Of-Work
	// check if the PoW is satisfied
	if !chain.VerifyPoW(dbtx, bl) { // if invalid Pow, reject the bloc
		block_logger.Warnf("Block has invalid PoW, rejecting it %x", bl.Serialize())
		return errormsg.ErrInvalidPoW, false
	}

	// verify coinbase tx
	if !chain.Verify_Transaction_Coinbase(dbtx, cbl, &bl.Miner_TX) {
		block_logger.Warnf("Miner tx failed verification  rejecting ")
		return errormsg.ErrInvalidBlock, false
	}

	// from version 2, minertx should contain the 0 as base reward as it is calculated by client protocol
	if chain.Get_Current_Version_at_Height(block_height) >= 2 && bl.Miner_TX.Vout[0].Amount != 0 {
		block_logger.Warnf("Miner tx failed should have block reward as zero,rejecting block")
		return errormsg.ErrInvalidBlock, false
	}

	{
		// now we need to verify each and every tx in detail
		// we need to verify each and every tx contained in the block, sanity check everything
		// first of all check, whether all the tx contained in the block, match their hashes
		{
			if len(bl.Tx_hashes) != len(cbl.Txs) {
				block_logger.Warnf("Block says it has %d txs , however complete block contained %d txs", len(bl.Tx_hashes), len(cbl.Txs))
				return errormsg.ErrInvalidBlock, false
			}

			// first check whether the complete block contains any diplicate hashes
			tx_checklist := map[crypto.Hash]bool{}
			for i := 0; i < len(bl.Tx_hashes); i++ {
				tx_checklist[bl.Tx_hashes[i]] = true
			}

			if len(tx_checklist) != len(bl.Tx_hashes) { // block has duplicate tx, reject
				block_logger.Warnf("Block has %d  duplicate txs, reject it", len(bl.Tx_hashes)-len(tx_checklist))
				return errormsg.ErrInvalidBlock, false

			}
			// now lets loop through complete block, matching each tx
			// detecting any duplicates using txid hash
			for i := 0; i < len(cbl.Txs); i++ {
				tx_hash := cbl.Txs[i].GetHash()
				if _, ok := tx_checklist[tx_hash]; !ok {
					// tx is NOT found in map, RED alert reject the block
					block_logger.Warnf("Block says it has tx %s, but complete block does not have it", tx_hash)
					return errormsg.ErrInvalidBlock, false
				}
			}
		}

		// another check, whether the tx contains any duplicate key images within the block
		// block wide duplicate input detector
		// TODO FIXME replace with a simple map
		{
			key_image_map := map[crypto.Hash]bool{}
			for i := 0; i < len(cbl.Txs); i++ {
				for j := 0; j < len(cbl.Txs[i].Vin); j++ {
					if _, ok := key_image_map[cbl.Txs[i].Vin[j].(transaction.Txin_to_key).K_image]; ok {
						block_logger.Warnf("Double Spend attack within block %s", cbl.Txs[i].GetHash())
						return errormsg.ErrTXDoubleSpend, false
					}
					key_image_map[cbl.Txs[i].Vin[j].(transaction.Txin_to_key).K_image] = true
				}
			}
		}

		// TODO FIXME
		// we need to check whether the dishonest miner is trying to include junk transactions which have been already mined and confirmed
		// for these purposes we track keyimages with height where they have been spent
		// so if a block contains any key images from earlier the stable point, reject the block even if PoW was good

		for i := 0; i < len(cbl.Txs); i++ { // loop through all the TXs
			for j := 0; j < len(cbl.Txs[i].Vin); j++ {
				keyimage_height, ok := chain.Read_KeyImage_Status(dbtx, cbl.Txs[i].Vin[j].(transaction.Txin_to_key).K_image)
				if ok && block_height-keyimage_height > 13 { // why 13, because reachability checks for 15
					block_logger.Warnf("Dead TX attack tx %s contains DEAD transactions, rejecting ", cbl.Txs[i].GetHash())
					return errormsg.ErrTXDead, false
				}
			}
		}

		// we also need to reject if the the immediately reachable history, has spent the keyimage
		// both the checks works on the basis of keyimages and not on the basis of txhash
		reachable_key_images := chain.BuildReachabilityKeyImages(dbtx, bl)

		//block_logger.Infof("len of reachable keyimages %d", len(reachable_key_images))
		for i := 0; i < len(cbl.Txs); i++ { // loop through all the TXs
			for j := 0; j < len(cbl.Txs[i].Vin); j++ {
				if _, ok := reachable_key_images[cbl.Txs[i].Vin[j].(transaction.Txin_to_key).K_image]; ok {
					block_logger.Warnf("Double spend attack tx %s is already mined, rejecting ", cbl.Txs[i].GetHash())
					return errormsg.ErrTXDead, false
				}
			}
		}

		// verify all non coinbase tx, single threaded, we have a multithreaded version below

		/*
			hf_version := chain.Get_Current_Version_at_Height(chain.Calculate_Height_At_Tips(dbtx, bl.Tips))

			for i := 0 ; i < len(cbl.Txs); i++ {
				rlog.Debugf("addcomplete block tx %s hf_version %d  height %d", cbl.Txs[i].GetHash(), hf_version, chain.Calculate_Height_At_Tips(dbtx, bl.Tips)  )

				if !chain.Verify_Transaction_NonCoinbase(dbtx,hf_version,cbl.Txs[i]){
					logger.Warnf("Non Coinbase tx failed verification  rejecting " )
				 	return errormsg.ErrInvalidTX, false
				}
			}
		*/

		// we need to anyways verify the TXS since RCT signatures are not covered by checksum
		fail_count := int32(0)
		wg := sync.WaitGroup{}
		wg.Add(len(cbl.Txs)) // add total number of tx as work

		hf_version := chain.Get_Current_Version_at_Height(chain.Calculate_Height_At_Tips(dbtx, bl.Tips))
		for i := 0; i < len(cbl.Txs); i++ {
			go func(j int) {

				// NOTE : do NOT skip verification of Ring Signatures, even if the TX is already stored
				//        as change of conditions might cause the signature to be invalid
				if !chain.Verify_Transaction_NonCoinbase(dbtx, hf_version, cbl.Txs[j]) { // transaction verification failed
					atomic.AddInt32(&fail_count, 1) // increase fail count by 1
					block_logger.Warnf("Block verification failed rejecting since TX  %s verification failed", cbl.Txs[j].GetHash())
				}
				wg.Done()
			}(i)
		}

		wg.Wait()           // wait for verifications to finish
		if fail_count > 0 { // check the result
			block_logger.Warnf("Block verification failed  rejecting since TX verification failed ")
			return errormsg.ErrInvalidTX, false
		}

	}

	// we are here means everything looks good, proceed and save to chain
skip_checks:

	// save all the txs
	// and then save the block
	{ // first lets save all the txs, together with their link to this block as height
		for i := 0; i < len(cbl.Txs); i++ {
			chain.Store_TX(dbtx, cbl.Txs[i])
		}
	}

	chain.Store_BL(dbtx, bl)

	// if the block is on a lower height tip, the block will not increase chain height
	height := chain.Load_Height_for_BL_ID(dbtx, block_hash)
	if height > chain.Get_Height() || height == 0 { // exception for genesis block
		atomic.StoreInt64(&chain.Height, height)
		chain.Store_TOP_HEIGHT(dbtx, height)
		rlog.Infof("Chain extended new height %d blid %s", chain.Height, block_hash)

	} else {
		rlog.Infof("Chain extended but height is same %d blid %s", chain.Height, block_hash)

	}

	// calculate new set of tips
	// this is done by removing all known tips which are in the past
	// and add this block as tip

	past := chain.Get_Block_Past(dbtx, bl.GetHash())

	old_tips := chain.load_TIPS(dbtx)
	tips_map := map[crypto.Hash]bool{bl.GetHash(): true} // add this new block as tip
	for i := range old_tips {
		tips_map[old_tips[i]] = true
	}
	for i := range past {
		delete(tips_map, past[i])
	}

	tips := []crypto.Hash{}
	for k, _ := range tips_map {
		tips = append(tips, k)
	}
	chain.store_TIPS(dbtx, tips)

	// find the biggest tip  in terms of work
	{
		tips := chain.load_TIPS(dbtx)
		base, base_height := chain.find_common_base(dbtx, tips)
		best := chain.find_best_tip(dbtx, tips, base, base_height)

		//logger.Infof("tips %+v  base %s ",tips, base)

		// we  only generate full order for the biggest tip

		//gbl := Generate_Genesis_Block()
		// full_order := chain.Generate_Full_Order( bl.GetHash(), gbl.GetHash(), 0,0)
		//base_topo_index := chain.Load_Block_Topological_order(gbl.GetHash())

		full_order := chain.Generate_Full_Order(dbtx, best, base, base_height, 0)
		base_topo_index := chain.Load_Block_Topological_order(dbtx, base)

		highest_topo := int64(0)

		// we must also run the client protocol in reverse to undo changes in already existing  order

		// reverse the order
		// for i, j := 0, len(full_order)-1; i < j; i, j = i+1, j-1 {
		//     full_order[i], full_order[j] = full_order[j], full_order[i]
		// }

		rlog.Infof("Full order %+v base %s base topo pos %d", full_order, base, base_topo_index)

		last_topo_height := chain.Load_TOPO_HEIGHT(dbtx)

		if len(bl.Tips) == 0 {
			base_topo_index = 0
		}

		// run the client_protocol_reverse , till we reach the base block
		for last_topo_height > 0 {
			last_topo_block, err := chain.Load_Block_Topological_order_at_index(dbtx, last_topo_height)

			if err != nil {
				logger.Warnf("Block not found while running client protocol in reverse %s, probably DB corruption", last_topo_block)
				return errormsg.ErrInvalidBlock, false
			}

			bl_current, err := chain.Load_BL_FROM_ID(dbtx, last_topo_block)
			if err != nil {
				block_logger.Debugf("Cannot load block %s for client protocol reverse,probably DB corruption ", last_topo_block)
				return errormsg.ErrInvalidBlock, false
			}

			rlog.Debugf("running client protocol in reverse for %s", last_topo_block)

			chain.client_protocol_reverse(dbtx, bl_current, last_topo_block)

			// run client protocol in reverse till we reach base
			if last_topo_block != full_order[0] {
				last_topo_height--
			} else {
				break
			}
		}

		// TODO FIXME we must avoid reprocessing  base block and or duplicate blocks, no point in reprocessing it
		for i := int64(0); i < int64(len(full_order)); i++ {

			chain.Store_Block_Topological_order(dbtx, full_order[i], i+base_topo_index)
			highest_topo = base_topo_index + i

			rlog.Debugf("%d %s   topo_index %d  base topo %d", i, full_order[i], i+base_topo_index, base_topo_index)

			// TODO we must run smart contracts and TXs in this order
			// basically client protocol must run here
			// even if the HF has triggered we may still accept, old blocks for some time
			// so hf is detected block-wise and processed as such

			bl_current_hash := full_order[i]
			bl_current, err1 := chain.Load_BL_FROM_ID(dbtx, bl_current_hash)
			if err1 != nil {
				block_logger.Debugf("Cannot load block %s for client protocol,probably DB corruption", bl_current_hash)
				return errormsg.ErrInvalidBlock, false
			}

			height_current := chain.Calculate_Height_At_Tips(dbtx, bl_current.Tips)
			hard_fork_version_current := chain.Get_Current_Version_at_Height(height_current)

			//  run full client protocol and find valid transactions
			// find all transactions within this block which are NOT double-spend
			// if any double-SPEND are found ignore them, else collect their fees to give to miner
			total_fees := chain.client_protocol(dbtx, bl_current, bl_current_hash, height_current, highest_topo)

			rlog.Debugf("running client protocol for %s minertx %s  topo %d", bl_current_hash, bl_current.Miner_TX.GetHash(), highest_topo)

			// store and parse miner tx
			chain.Store_TX(dbtx, &bl_current.Miner_TX)
			chain.Store_TX_Height(dbtx, bl_current.Miner_TX.GetHash(), highest_topo)

			// mark TX found in this block also  for explorer
			chain.store_TX_in_Block(dbtx, bl_current_hash, bl_current.Miner_TX.GetHash())

			//mark tx found in this block is valid
			chain.mark_TX(dbtx, bl_current_hash, bl_current.Miner_TX.GetHash(), true)

			// hard fork version is used to import transactions from earlier version of DERO chain
			// in order to keep things simple, the earlier emission/fees calculation/dynamic block size has been discarded
			// due to above reasons miner TX from the earlier could NOT be verified
			// emission calculations/ total supply should NOT change when importing earlier chain
			if hard_fork_version_current == 1 {

				// store total  reward
				dbtx.StoreUint64(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, bl_current_hash[:], PLANET_MINERTX_REWARD, bl_current.Miner_TX.Vout[0].Amount)

				// store base reward
				dbtx.StoreUint64(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, bl_current_hash[:], PLANET_BASEREWARD, bl_current.Miner_TX.Vout[0].Amount)

				// store total generated coins
				// this is hardcoded at initial chain import, keeping original emission schedule
				if globals.IsMainnet(){
						dbtx.StoreUint64(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, bl_current_hash[:], PLANET_ALREADY_GENERATED_COINS, config.MAINNET_HARDFORK_1_TOTAL_SUPPLY)
					}else{
						dbtx.StoreUint64(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, bl_current_hash[:], PLANET_ALREADY_GENERATED_COINS, config.TESTNET_HARDFORK_1_TOTAL_SUPPLY)
					}
				

			} else { //  hf 2 or later generate miner TX rewards as per client protocol

				past_coins_generated := chain.Load_Already_Generated_Coins_for_Topo_Index(dbtx, highest_topo-1)

				base_reward := emission.GetBlockReward_Atlantis(hard_fork_version_current, past_coins_generated)

				// base reward is only 90%, rest 10 % is pushed back
				if globals.IsMainnet(){
					base_reward = (base_reward * 9) / 10
				}

				// lower reward for byzantine behaviour
				// for as many block as added
				if chain.isblock_SideBlock(dbtx, bl_current_hash, highest_topo) { // lost race (or byzantine behaviour)
                                    if  hard_fork_version_current == 2 {
					base_reward = (base_reward * 67) / 100 // give only 67 % reward
                                    }else{
                                        base_reward = (base_reward * 8) / 100 // give only 8 % reward
                                    }
				}

				// logger.Infof("past coins generated %d base reward %d", past_coins_generated, base_reward)

				// the total reward must be given to the miner TX, since it contains 0, we patch only the output
				// and leave the original TX untouched
				total_reward := base_reward + total_fees

				// store total  reward
				dbtx.StoreUint64(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, bl_current_hash[:], PLANET_MINERTX_REWARD, total_reward)

				// store base reward
				dbtx.StoreUint64(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, bl_current_hash[:], PLANET_BASEREWARD, base_reward)

				// store total generated coins
				dbtx.StoreUint64(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, bl_current_hash[:], PLANET_ALREADY_GENERATED_COINS, past_coins_generated+base_reward)

				//logger.Infof("base reward %s  total generated %s",globals.FormatMoney12(base_reward), globals.FormatMoney12(past_coins_generated+base_reward))

			}

			// TODO FIXME valid transactions must be found and thier fees should be added as reward

			// output index starts from the ending of the previous block

			// get previous block
			output_index_start := int64(0)
			if (highest_topo - 1) >= 0 {
				previous_block, err1 := chain.Load_Block_Topological_order_at_index(dbtx, highest_topo-1)
				if err1 != nil {
					logger.Warnf("Errr could not find topo index of previous block")
					return errormsg.ErrInvalidBlock, false
				}
				// we will start where the previous block vouts ended
				_, output_index_start = chain.Get_Block_Output_Index(dbtx, previous_block)
			}
			if !chain.write_output_index(dbtx, bl_current_hash, output_index_start, hard_fork_version_current) {
				logger.Warnf("Since output index data cannot be wrritten, skipping block")
				return errormsg.ErrInvalidBlock, false
			}

			// this tx must be stored, linked with this block

		}

		chain.Store_TOPO_HEIGHT(dbtx, int64(highest_topo))

		// set main chain as new topo order
		// we must discard any rusty tips after they go stale
		best_height := int64(chain.Load_Height_for_BL_ID(dbtx, best))

		new_tips := []crypto.Hash{}
		for i := range tips {
			rusty_tip_base_distance := chain.calculate_mainchain_distance(dbtx, tips[i])
			// tips of deviation > 6 will be rejected
			if (best_height - rusty_tip_base_distance) < (config.STABLE_LIMIT - 1) {
				new_tips = append(new_tips, tips[i])

			} else { // this should be a rarest event, probably should never occur, until the network is under sever attack
				logger.Warnf("Rusty TIP declared stale %s  best height %d deviation %d rusty_tip %d", tips[i], best_height, (best_height - rusty_tip_base_distance), rusty_tip_base_distance)
				chain.transaction_scavenger(dbtx, tips[i]) // scavenge tx if possible
				// TODO we must include any TX from the orphan blocks back to the mempool to avoid losing any TX
			}
		}

		// do more cleanup of tips for byzantine behaviour
		// this copy is necessary, otherwise data corruption occurs
		tips = append([]crypto.Hash{},new_tips...) 
		new_tips = new_tips[:0]
		best_tip := chain.find_best_tip_cumulative_difficulty(dbtx, tips)
		
		new_tips = append(new_tips, best_tip)
		for i := range tips {
			if best_tip != tips[i] {
				if !chain.validate_tips(dbtx, best_tip, tips[i]) { // reference is first
					logger.Warnf("Rusty tip %s declaring stale", tips[i])
					chain.transaction_scavenger(dbtx, tips[i]) // scavenge tx if possible
				} else {
					new_tips = append(new_tips, tips[i])
				}
			}
		}

		rlog.Infof("New tips(after adding %s) %+v", bl.GetHash(), new_tips)
		chain.store_TIPS(dbtx, new_tips)

	}

	//chain.store_TIPS(chain.)

	//chain.Top_ID = block_hash // set new top block id

	// every 200 block print a line
	if chain.Get_Height()%200 == 0 {
		block_logger.Infof("Chain Height %d", chain.Height)
	}

	result = true

	// TODO fix hard fork
	// maintain hard fork votes to keep them SANE
	//chain.Recount_Votes() // does not return anything

	// enable mempool book keeping

	func() {
		if r := recover(); r != nil {
			logger.Warnf("Mempool House Keeping triggered panic height = %d", block_height)
		}

		// discard the transactions from mempool if they are present there
		chain.Mempool.Monitor()

		for i := 0; i < len(cbl.Txs); i++ {
			txid := cbl.Txs[i].GetHash()
			if chain.Mempool.Mempool_TX_Exist(txid) {
				rlog.Tracef(1, "Deleting TX from pool txid=%s", txid)
				chain.Mempool.Mempool_Delete_TX(txid)
			}
		}

		// give mempool an oppurtunity to clean up tx, but only if they are not mined
		// but only check for double spending
		chain.Mempool.HouseKeeping(uint64(block_height), func(tx *transaction.Transaction) bool {
			return chain.Verify_Transaction_NonCoinbase_DoubleSpend_Check(dbtx, tx)
		})
	}()

	return // run any handlers necesary to atomically
}

// runs the client protocol which includes the following operations
// if any TX are being duplicate or double-spend ignore them
// mark all the valid transactions as valid
// mark all invalid transactions  as invalid
// calculate total fees based on valid TX
// we need NOT check ranges/ring signatures here, as they have been done already by earlier steps
func (chain *Blockchain) client_protocol(dbtx storage.DBTX, bl *block.Block, blid crypto.Hash, height int64, topoheight int64) (total_fees uint64) {
	// run client protocol for all TXs
	for i := range bl.Tx_hashes {
		tx, err := chain.Load_TX_FROM_ID(dbtx, bl.Tx_hashes[i])
		if err != nil {
			panic(fmt.Errorf("Cannot load  tx for %x err %s ", bl.Tx_hashes[i], err))
		}
		// mark TX found in this block also  for explorer
		chain.store_TX_in_Block(dbtx, blid, bl.Tx_hashes[i])

		// check all key images as double spend, if double-spend detected mark invalid, else consider valid
		if chain.Verify_Transaction_NonCoinbase_DoubleSpend_Check(dbtx, tx) {

			chain.consume_keyimages(dbtx, tx, height) // mark key images as consumed
			total_fees += tx.RctSignature.Get_TX_Fee()

			chain.Store_TX_Height(dbtx, bl.Tx_hashes[i], topoheight) // link the tx with the topo height

			//mark tx found in this block is valid
			chain.mark_TX(dbtx, blid, bl.Tx_hashes[i], true)

		} else { // TX is double spend or reincluded by 2 blocks simultaneously
			rlog.Tracef(1,"Double spend TX is being ignored %s %s", blid, bl.Tx_hashes[i])
			chain.mark_TX(dbtx, blid, bl.Tx_hashes[i], false)
		}
	}

	return total_fees
}

// this undoes everything that is done by client protocol
// NOTE: this will have any effect, only if client protocol has been run on this block earlier
func (chain *Blockchain) client_protocol_reverse(dbtx storage.DBTX, bl *block.Block, blid crypto.Hash) {
	// run client protocol for all TXs
	for i := range bl.Tx_hashes {
		tx, err := chain.Load_TX_FROM_ID(dbtx, bl.Tx_hashes[i])
		if err != nil {
			panic(fmt.Errorf("Cannot load  tx for %x err %s ", bl.Tx_hashes[i], err))
		}
		// only the  valid TX must be revoked
		if chain.IS_TX_Valid(dbtx, blid, bl.Tx_hashes[i]) {
			chain.revoke_keyimages(dbtx, tx) // mark key images as not used

			chain.Store_TX_Height(dbtx, bl.Tx_hashes[i], -1) // unlink the tx with the topo height

			//mark tx found in this block is invalid
			chain.mark_TX(dbtx, blid, bl.Tx_hashes[i], false)

		} else { // TX is double spend or reincluded by 2 blocks simultaneously
			// invalid tx is related
		}
	}

	return
}

// scavanger for transactions from rusty/stale tips to reinsert them into pool
func (chain *Blockchain) transaction_scavenger(dbtx storage.DBTX, blid crypto.Hash) {
	defer func() {
		if r := recover(); r != nil {
			logger.Warnf("Recovered while transaction scavenging, Stack trace below ")
			logger.Warnf("Stack trace  \n%s", debug.Stack())
		}
	}()

	logger.Debugf("scavenging transactions from blid %s", blid)
	reachable_blocks := chain.BuildReachableBlocks(dbtx, []crypto.Hash{blid})
	reachable_blocks[blid] = true // add self
	for k, _ := range reachable_blocks {
		if chain.Is_Block_Orphan(k) {
			bl, err := chain.Load_BL_FROM_ID(dbtx, k)
			if err == nil {
				for i := range bl.Tx_hashes {
					tx, err := chain.Load_TX_FROM_ID(dbtx, bl.Tx_hashes[i])
					if err != nil {
						rlog.Warnf("err while scavenging blid %s  txid %s err %s", k, bl.Tx_hashes[i], err)
					} else {
						// add tx to pool, it will do whatever is necessarry
						chain.Add_TX_To_Pool(tx)
					}
				}
			} else {
				rlog.Warnf("err while scavenging blid %s err %s", k, err)
			}
		}
	}
}

// Finds whether a  block is orphan
// since we donot store any fields, we need to calculate/find the block as orphan
// using an algorithm
// if the block is NOT topo ordered , it is orphan/stale
func (chain *Blockchain) Is_Block_Orphan(hash crypto.Hash) bool {
	return !chain.Is_Block_Topological_order(nil, hash)
}

// this is used to find if a tx is orphan, YES orphan TX
// these can occur during  when they are detect to be double-spended on
// so the TX becomes orphan ( chances are less may be less that .000001 % but they are there)
// if a tx is not valid in any of the blocks, it has been mined it is orphan
func (chain *Blockchain) Is_TX_Orphan(hash crypto.Hash) (result bool) {
	blocks := chain.Load_TX_blocks(nil, hash)
	for i := range blocks {
		if chain.IS_TX_Valid(nil, blocks[i], hash) && chain.Is_Block_Topological_order(nil, blocks[i]) {
			return false
		}

	}

	return true

}

// this is used to for quick syncs as entire blocks as SHA1,
// entires block can skipped for verification, if checksum matches what the devs have stored
func (chain *Blockchain) BlockCheckSum(cbl *block.Complete_Block) []byte {
	h := sha3.New256()
	h.Write(cbl.Bl.Serialize())
	for i := range cbl.Txs {
		h.Write(cbl.Txs[i].Serialize())
	}
	return h.Sum(nil)
}

// this function will mark all the key images present in the tx as requested
// this is done so as they cannot be respent
// mark is int64  height
func (chain *Blockchain) mark_keyimages(dbtx storage.DBTX, tx *transaction.Transaction, height int64) bool {
	// mark keyimage as spent
	for i := 0; i < len(tx.Vin); i++ {
		k_image := tx.Vin[i].(transaction.Txin_to_key).K_image
		chain.Store_KeyImage(dbtx, crypto.Hash(k_image), height)
	}
	return true
}

//this will mark all the keyimages present in this TX as spent
//this is done so as an input cannot be spent twice
func (chain *Blockchain) consume_keyimages(dbtx storage.DBTX, tx *transaction.Transaction, height int64) bool {
	return chain.mark_keyimages(dbtx, tx, height)
}

//this will mark all the keyimages present in this TX as unspent
//this is required during  client protocol runs
func (chain *Blockchain) revoke_keyimages(dbtx storage.DBTX, tx *transaction.Transaction) bool {
	return chain.mark_keyimages(dbtx, tx, -1) // -1 is a marker that stored key-image is not valid
}

/* this will only give you access to transactions which have been mined
 */
func (chain *Blockchain) Get_TX(dbtx storage.DBTX, hash crypto.Hash) (*transaction.Transaction, error) {
	tx, err := chain.Load_TX_FROM_ID(dbtx, hash)

	return tx, err
}

// This will get the biggest height of tip for hardfork version and other calculations
// get biggest height of parent, add 1
func (chain *Blockchain) Calculate_Height_At_Tips(dbtx storage.DBTX, tips []crypto.Hash) int64 {
	height := int64(0)
	if len(tips) == 0 { // genesis block has no parent

	} else { // find the best height of past
		for i := range tips {
			past_height := chain.Load_Height_for_BL_ID(dbtx, tips[i])
			if height <= past_height {
				height = past_height
			}
		}
		height++
	}
	return height

}

// this function return the current top block, if we start at specific block
// this works for any blocks which were added
/*
func (chain *Blockchain) Get_Top_Block(block_id crypto.Hash) crypto.Hash {
	for {
		// check if the block has child, if not , we are the top
		if !chain.Does_Block_Have_Child(block_id) {
			return block_id
		}
		block_id = chain.Load_Block_Child(block_id) // continue searching the new top
	}
	//	panic("We can never reach this point")
	//	return block_id // we will never reach here
}
*/

// verifies whether we are lagging
// return true if we need resync
// returns false if we are good and resync is not required
func (chain *Blockchain) IsLagging(peer_cdiff *big.Int) bool {

	our_diff := new(big.Int).SetInt64(0)

	high_block, err := chain.Load_Block_Topological_order_at_index(nil, chain.Load_TOPO_HEIGHT(nil))
	if err != nil {
		return false
	} else {
		our_diff = chain.Load_Block_Cumulative_Difficulty(nil, high_block)
	}
	rlog.Tracef(2, "P_cdiff %s cdiff %d  our top block %s", peer_cdiff.String(), our_diff.String(), high_block)

	if our_diff.Cmp(peer_cdiff) < 0 {
		return true // peer's cumulative difficulty is more than ours , active resync
	}
	return false
}

// This function will expand a transaction with all the missing info being reconstitued from the blockchain
// this also increases security since data is coming from the chain or being calculated
// basically this places data for ring signature verification
// REMEMBER to expand key images from the blockchain
// TODO we must enforce that the keyimages used are valid and specific outputs are unlocked
func (chain *Blockchain) Expand_Transaction_v2(dbtx storage.DBTX, hf_version int64, tx *transaction.Transaction) (result bool) {

	result = false
	if tx.Version != 2 {
		panic("TX not version 2")
	}

	chain_height_cached := uint64(chain.Get_Height()) // do it once instead of in for loop

	//if rctsignature is null

	// fill up the message hash first
	tx.RctSignature.Message = crypto.Key(tx.GetPrefixHash())

	// fill up the key images from the blockchain
	for i := 0; i < len(tx.Vin); i++ {
		tx.RctSignature.MlsagSigs[i].II = tx.RctSignature.MlsagSigs[i].II[:0] // zero it out
		tx.RctSignature.MlsagSigs[i].II = make([]crypto.Key, 1, 1)
		tx.RctSignature.MlsagSigs[i].II[0] = crypto.Key(tx.Vin[i].(transaction.Txin_to_key).K_image)
	}

	// now we need to fill up the mixring ctkey
	// one part is the destination address, second is the commitment mask from the outpk
	// mixring is stored in different ways for rctfull and simple

	switch tx.RctSignature.Get_Sig_Type() {

	case ringct.RCTTypeFull:
		// TODO, we need to make sure all ring are of same size

		if len(tx.Vin) > 1 {
			panic("unsupported ringct full case please investigate")
		}

		// make a matrix of mixin x 1 elements
		mixin := len(tx.Vin[0].(transaction.Txin_to_key).Key_offsets)
		tx.RctSignature.MixRing = make([][]ringct.CtKey, mixin, mixin)
		for n := 0; n < len(tx.Vin); n++ {
			offset := uint64(0)
			for m := 0; m < len(tx.Vin[n].(transaction.Txin_to_key).Key_offsets); m++ {
				tx.RctSignature.MixRing[m] = make([]ringct.CtKey, len(tx.Vin), len(tx.Vin))

				offset += tx.Vin[n].(transaction.Txin_to_key).Key_offsets[m]
				// extract the keys from specific offset
				offset_data, success := chain.load_output_index(dbtx, offset)

				if !success {
					return false
				}

				// check maturity of inputs
				if hf_version >= 2 && !inputmaturity.Is_Input_Mature(chain_height_cached, offset_data.Height, offset_data.Unlock_Height, offset_data.SigType) {
					rlog.Tracef(1, "transaction using immature inputs from block %d chain height %d", offset_data.Height, chain_height_cached)
					return false
				}

				tx.RctSignature.MixRing[m][n].Destination = offset_data.InKey.Destination
				tx.RctSignature.MixRing[m][n].Mask = offset_data.InKey.Mask
				//	fmt.Printf("%d %d dest %s\n",n,m, offset_data.InKey.Destination)
				//	fmt.Printf("%d %d mask %s\n",n,m, offset_data.InKey.Mask)

			}
		}

	case ringct.RCTTypeSimple, ringct.RCTTypeSimpleBulletproof:
		mixin := len(tx.Vin[0].(transaction.Txin_to_key).Key_offsets)
		_ = mixin
		tx.RctSignature.MixRing = make([][]ringct.CtKey, len(tx.Vin), len(tx.Vin))

		for n := 0; n < len(tx.Vin); n++ {

			tx.RctSignature.MixRing[n] = make([]ringct.CtKey, len(tx.Vin[n].(transaction.Txin_to_key).Key_offsets),
				len(tx.Vin[n].(transaction.Txin_to_key).Key_offsets))
			offset := uint64(0)

			// this test is being done keeping future in mind
			duplicate_destination_mask := map[crypto.Key]bool{}
			for m := 0; m < len(tx.Vin[n].(transaction.Txin_to_key).Key_offsets); m++ {

				offset += tx.Vin[n].(transaction.Txin_to_key).Key_offsets[m]

				// extract the keys from specific offset
				offset_data, success := chain.load_output_index(dbtx, offset)

				if !success {
					return false
				}

				//logger.Infof("Ring member %+v", offset_data)

				//logger.Infof("cheight %d  ring member height %d  locked height %d sigtype %d", chain.Get_Height(), offset_data.Height, offset_data.Unlock_Height, 1 )

				//logger.Infof("mature %+v  tx %s hf_version %d", inputmaturity.Is_Input_Mature(uint64(chain.Get_Height()), offset_data.Height, offset_data.Unlock_Height, 1), tx.GetHash(), hf_version)

				// check maturity of inputs
				if hf_version >= 2 && !inputmaturity.Is_Input_Mature(chain_height_cached, offset_data.Height, offset_data.Unlock_Height, offset_data.SigType) {
					rlog.Tracef(1, "transaction using immature inputs from block %d chain height %d", offset_data.Height, chain_height_cached)
					return false
				}
				if _, ok := duplicate_destination_mask[offset_data.InKey.Destination]; ok {
					rlog.Warnf("Duplicate Keys tx hash %s", tx.GetHash())
					return false
				}
				if _, ok := duplicate_destination_mask[offset_data.InKey.Mask]; ok {
					rlog.Warnf("Duplicate Masks %s", tx.GetHash())
					return false
				}

				duplicate_destination_mask[offset_data.InKey.Destination] = true
				duplicate_destination_mask[offset_data.InKey.Mask] = true

				tx.RctSignature.MixRing[n][m].Destination = offset_data.InKey.Destination
				tx.RctSignature.MixRing[n][m].Mask = offset_data.InKey.Mask
				//	fmt.Printf("%d %d dest %s\n",n,m, offset_data.InKey.Destination)
				//	fmt.Printf("%d %d mask %s\n",n,m, offset_data.InKey.Mask)

			}
		}

	default:
		logger.Warnf("unknown ringct transaction")
		return false
	}

	return true
}

// this function count all the vouts of the block,
// this function exists here because  only the chain knows the tx
func (chain *Blockchain) Block_Count_Vout(dbtx storage.DBTX, block_hash crypto.Hash) (count uint64) {
	count = 1 // miner tx is always present

	bl, err := chain.Load_BL_FROM_ID(dbtx, block_hash)

	if err != nil {
		panic(fmt.Errorf("Cannot load  block for %s err %s", block_hash, err))
	}

	for i := 0; i < len(bl.Tx_hashes); i++ { // load all tx one by one
		tx, err := chain.Load_TX_FROM_ID(dbtx, bl.Tx_hashes[i])
		if err != nil {
			panic(fmt.Errorf("Cannot load  tx for %s err %s", bl.Tx_hashes[i], err))
		}

		// tx has been loaded, now lets get the vout
		vout_count := uint64(len(tx.Vout))
		count += vout_count
	}
	return count
}

// tells whether the hash already exists in slice
func sliceExists(slice []crypto.Hash, hash crypto.Hash) bool {
	for i := range slice {
		if slice[i] == hash {
			return true
		}
	}
	return false
}

// this function will rewind the chain from the topo height one block at a time
// this function also runs the client protocol in reverse and also deletes the block from the storage
func (chain *Blockchain) Rewind_Chain(rewind_count int) (result bool) {
	chain.Lock()
	defer chain.Unlock()

	dbtx, err := chain.store.BeginTX(true)
	if err != nil {
		logger.Warnf("Could NOT rewind chain. Error opening writable TX, err %s", err)
		return false
	}

	defer func() {
		// safety so if anything wrong happens, verification fails
		if r := recover(); r != nil {
			logger.Warnf("Recovered while rewinding chain, Stack trace below block_hash ")
			logger.Warnf("Stack trace  \n%s", debug.Stack())
			result = false
		}

		if result == true { // block was successfully added, commit it atomically
			dbtx.Commit()
			dbtx.Sync() // sync the DB to disk after every execution of this function
		} else {
			dbtx.Rollback() // if block could not be added, rollback all changes to previous block
		}
	}()

	// we must always rewind till a safety point is found
	stable_points := map[int64]bool{}

	// we must till we reach a safe point
	// safe point is point where a single block exists at specific height
	// this may lead us to rewinding a it more
	//safe := false

	// TODO we must fix safeness using the stable calculation

	// keep rewinding till safe point is reached
	for done := 0; ; done++ {
		top_block_topo_index := chain.Load_TOPO_HEIGHT(dbtx)

		//logger.Infof("stable points %d", len(stable_points))
		// keep rewinding till a safe point is not found
		if done >= rewind_count {
			if _, ok := stable_points[top_block_topo_index]; ok {
				break
			}
		}

		if top_block_topo_index < 1 {
			logger.Warnf("Cannot rewind  genesis block  topoheight %d", top_block_topo_index)
			return false
		}

		// check last 100 blocks for safety
		for i := int64(0); i < 50 && (top_block_topo_index-i >= 5); i++ {
			hash, err := chain.Load_Block_Topological_order_at_index(dbtx, top_block_topo_index-i)
			if err != nil {
				logger.Warnf("Cannot rewind chain at topoheight %d err: %s", top_block_topo_index, err)
				return false
			}

			// TODO add a check whether a single block exists at this height,
			// if yes consider it as a sync block
			h := chain.Load_Height_for_BL_ID(dbtx, hash)

			if len(chain.Get_Blocks_At_Height(dbtx, h)) != 1 { // we should have exactly 1 block at this height
				continue
			}

			if chain.IsBlockSyncBlockHeight(dbtx, hash) {
				stable_points[top_block_topo_index-i] = true
			}
		}

		blid, err := chain.Load_Block_Topological_order_at_index(dbtx, top_block_topo_index)
		if err != nil {
			logger.Warnf("Cannot rewind chain at topoheight %d err: %s", top_block_topo_index, err)
			return false
		}

		blocks_at_height := chain.Get_Blocks_At_Height(dbtx, chain.Load_Height_for_BL_ID(dbtx, blid))

		for _, blid := range blocks_at_height {
			// run the client protocol in reverse to undo keyimages

			bl_current, err := chain.Load_BL_FROM_ID(dbtx, blid)
			if err != nil {
				logger.Warnf("Cannot load block %s for client protocol reverse ", blid)
				return false
			}

			logger.Debugf("running client protocol in reverse for %s", blid)
			// run client protocol in reverse
			chain.client_protocol_reverse(dbtx, bl_current, blid)

			// delete the tip
			tips := chain.load_TIPS(dbtx)
			new_tips := []crypto.Hash{}

			for i := range tips {
				if tips[i] != blid {
					new_tips = append(new_tips, tips[i])
				}
			}

			// all the tips consumed by this block become the new tips
			for i := range bl_current.Tips {
				new_tips = append(new_tips, bl_current.Tips[i])
			}

			chain.store_TIPS(dbtx, new_tips) // store updated tips, we should rank and store them

			dbtx.StoreObject(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, blid[:], PLANET_BLOB, []byte(""))
			chain.Store_Block_Topological_order(dbtx, blid, -1)

		}

		// height of previous block becomes new height
		old_height := chain.Load_Height_for_BL_ID(dbtx, blid)
		chain.Store_TOP_HEIGHT(dbtx, old_height-1)

		tmp_blocks_at_height := chain.Get_Blocks_At_Height(dbtx, old_height-1)

		/*
			// we must unwind till the safe point is reached
			if len(tmp_blocks_at_height) == 1 && done >= rewind_count  && safe == false{
				rlog.Infof("Safety reached")
				safe = true
			}

			if len(tmp_blocks_at_height) != 1 && done >= rewind_count  && safe == false{
				rlog.Infof("Safety not reached rewinding more")
			}
		*/

		// this avoid possible database corruption by multiple blocks at same height

		lowest_positive_topo := int64(0)
		for _, blid := range tmp_blocks_at_height {
			if chain.Is_Block_Topological_order(dbtx, blid) {
				lowest_positive_topo = chain.Load_Block_Topological_order(dbtx, blid)
			}
		}
		chain.Store_TOPO_HEIGHT(dbtx, lowest_positive_topo)

		// no more blocks are stored at this height clean them
		dbtx.StoreObject(BLOCKCHAIN_UNIVERSE, GALAXY_HEIGHT, PLANET_HEIGHT, itob(uint64(old_height)), []byte{})

	}
	rlog.Infof("height after rewind %d", chain.Load_TOPO_HEIGHT(dbtx))

	return true
}


// build reachability graph upto 2*config deeps to answer reachability queries
func (chain *Blockchain) buildReachability_internal(dbtx storage.DBTX, reachmap map[crypto.Hash]bool, blid crypto.Hash, level int) {
	past := chain.Get_Block_Past(dbtx, blid)
	reachmap[blid] = true // add self to reach map

	if level >= int(2*config.STABLE_LIMIT) { // stop recursion must be more than  checks in add complete block
		return
	}
	for i := range past { // if no past == genesis return
		if _, ok := reachmap[past[i]]; !ok { // process a node, only if has not been processed earlier
			chain.buildReachability_internal(dbtx, reachmap, past[i], level+1)
		}
	}

}

// build reachability graph upto 2*limit  deeps to answer reachability queries
func (chain *Blockchain) buildReachability(dbtx storage.DBTX, blid crypto.Hash) map[crypto.Hash]bool {
	reachmap := map[crypto.Hash]bool{}
	chain.buildReachability_internal(dbtx, reachmap, blid, 0)
	return reachmap
}

// this is part of consensus rule, 2 tips cannot refer to their common parent
func (chain *Blockchain) VerifyNonReachability(dbtx storage.DBTX, bl *block.Block) bool {

	reachmaps := make([]map[crypto.Hash]bool, len(bl.Tips), len(bl.Tips))
	for i := range bl.Tips {
		reachmaps[i] = chain.buildReachability(dbtx, bl.Tips[i])
	}

	// bruteforce all reachability combinations, max possible 3x3 = 9 combinations
	for i := range bl.Tips {
		for j := range bl.Tips {
			if i == j { // avoid self test
				continue
			}

			if _, ok := reachmaps[j][bl.Tips[i]]; ok { // if a tip can be referenced as another's past, this is not a tip , probably malicious, discard block
				return false
			}

		}
	}

	return true
}

// used in the difficulty calculation for consensus and while scavenging
func (chain *Blockchain) BuildReachableBlocks(dbtx storage.DBTX, tips []crypto.Hash) map[crypto.Hash]bool {
	reachblocks := map[crypto.Hash]bool{} // contains a list of all reachable blocks
	for i := range tips {
		reachmap := chain.buildReachability(dbtx, tips[i])
		for k, _ := range reachmap {
			reachblocks[k] = true // build unique block list
		}
	}
	return reachblocks
}

// this is part of consensus rule, reachable blocks cannot have keyimages collision with new blocks
// this is to avoid dishonest miners including dead transactions
//
func (chain *Blockchain) BuildReachabilityKeyImages(dbtx storage.DBTX, bl *block.Block) map[crypto.Hash]bool {

	Keyimage_reach_map := map[crypto.Hash]bool{}

	reachblocks := map[crypto.Hash]bool{} // contains a list of all reachable blocks
	for i := range bl.Tips {
		reachmap := chain.buildReachability(dbtx, bl.Tips[i])
		for k, _ := range reachmap {
			reachblocks[k] = true // build unique block list
		}
	}

	// load all blocks and process their TX as per client protocol
	for blid, _ := range reachblocks {

		bl, err := chain.Load_BL_FROM_ID(dbtx, blid)
		if err != nil {
			panic(fmt.Errorf("Cannot load  block for %s err %s", blid, err))
		}

		for i := 0; i < len(bl.Tx_hashes); i++ { // load all tx one by one, skipping as per client_protocol

			/*
				if !chain.IS_TX_Valid(dbtx, blid, bl.Tx_hashes[i]) { // skip invalid TX
					rlog.Tracef(1, "bl %s tx %s ignored while building key image reachability as per client protocol")
					continue
				}
			*/

			tx, err := chain.Load_TX_FROM_ID(dbtx, bl.Tx_hashes[i])
			if err != nil {
				panic(fmt.Errorf("Cannot load  tx for %s err %s", bl.Tx_hashes[i], err))
			}

			// tx has been loaded, now lets get all the key images
			for i := 0; i < len(tx.Vin); i++ {
				Keyimage_reach_map[tx.Vin[i].(transaction.Txin_to_key).K_image] = true // add element to map for next check
			}
		}
	}
	return Keyimage_reach_map
}

// sync blocks have the following specific property
// 1) the block is singleton at this height
// basically the condition allow us to confirm weight of future blocks with reference to sync blocks
// these are the one who settle the chain and guarantee it
func (chain *Blockchain) IsBlockSyncBlockHeight(dbtx storage.DBTX, blid crypto.Hash) bool {

	// TODO make sure that block exist
	height := chain.Load_Height_for_BL_ID(dbtx, blid)
	if height == 0 { // genesis is always a sync block
		return true
	}

	//  top blocks are always considered unstable
	if (height + config.STABLE_LIMIT) > chain.Get_Height() {
		return false
	}

	// if block is not ordered, it can never be sync block
	if !chain.Is_Block_Topological_order(dbtx, blid) {
		return false
	}

	blocks := chain.Get_Blocks_At_Height(dbtx, height)

	if len(blocks) == 0 && height != 0 { // this  should NOT occur
		panic("No block exists at this height, not possible")
	}

	
	//   if len(blocks) == 1 { //  ideal blockchain case, it is a sync block
	//       return true
	//   }
	

	// check whether single block exists in the TOPO order index, if no we are NOT a sync block

	// we are here means we have one oor more block
	blocks_in_main_chain := 0
	for i := range blocks {
		if chain.Is_Block_Topological_order(dbtx, blocks[i]) {
			blocks_in_main_chain++
			if blocks_in_main_chain >= 2 {
				return false
			}
		}
	}

	// we are here if we only have one block in topological order, others are  dumped/rejected blocks

	// collect all blocks of past LIMIT heights
	var preblocks []crypto.Hash
	for i := height - 1; i >= (height-config.STABLE_LIMIT) && i != 0; i-- {
		blocks := chain.Get_Blocks_At_Height(dbtx, i)
		for j := range blocks { //TODO BUG BUG BUG we need to make sure only main chain blocks are considered
			preblocks = append(preblocks, blocks[j])
		}
	}

	// we need to find a common base to compare them, otherwise comparision is futile  due to duplication
	sync_block_cumulative_difficulty := chain.Load_Block_Cumulative_Difficulty(dbtx, blid) //+ chain.Load_Block_Difficulty(blid)

	// if any of the blocks  has a cumulative difficulty  more than  sync block, this situation affects  consensus, so mitigate it
	for i := range preblocks {
		cumulative_difficulty := chain.Load_Block_Cumulative_Difficulty(dbtx, preblocks[i]) // + chain.Load_Block_Difficulty(preblocks[i])

		//if cumulative_difficulty >= sync_block_cumulative_difficulty {
		if cumulative_difficulty.Cmp(sync_block_cumulative_difficulty) >= 0 {
			rlog.Warnf("Mitigating CONSENSUS issue on block %s height %d  child %s cdiff %d sync block cdiff %d", blid, height, preblocks[i], cumulative_difficulty, sync_block_cumulative_difficulty)
			return false
		}

	}

	return true
}

func (chain *Blockchain) IsBlockSyncBlockHeightSpecific(dbtx storage.DBTX, blid crypto.Hash, chain_height int64) bool {

	// TODO make sure that block exist
	height := chain.Load_Height_for_BL_ID(dbtx, blid)
	if height == 0 { // genesis is always a sync block
		return true
	}

	//  top blocks are always considered unstable
	if (height + config.STABLE_LIMIT) > chain_height {
		return false
	}

	// if block is not ordered, it can never be sync block
	if !chain.Is_Block_Topological_order(dbtx, blid) {
		return false
	}

	blocks := chain.Get_Blocks_At_Height(dbtx, height)

	if len(blocks) == 0 && height != 0 { // this  should NOT occur
		panic("No block exists at this height, not possible")
	}

	
	//   if len(blocks) == 1 { //  ideal blockchain case, it is a sync block
	//       return true
	//   }
	

	// check whether single block exists in the TOPO order index, if no we are NOT a sync block

	// we are here means we have one oor more block
	blocks_in_main_chain := 0
	for i := range blocks {
		if chain.Is_Block_Topological_order(dbtx, blocks[i]) {
			blocks_in_main_chain++
			if blocks_in_main_chain >= 2 {
				return false
			}
		}
	}

	// we are here if we only have one block in topological order, others are  dumped/rejected blocks

	// collect all blocks of past LIMIT heights
	var preblocks []crypto.Hash
	for i := height - 1; i >= (height-config.STABLE_LIMIT) && i != 0; i-- {
		blocks := chain.Get_Blocks_At_Height(dbtx, i)
		for j := range blocks { //TODO BUG BUG BUG we need to make sure only main chain blocks are considered
			preblocks = append(preblocks, blocks[j])
		}
	}

	// we need to find a common base to compare them, otherwise comparision is futile  due to duplication
	sync_block_cumulative_difficulty := chain.Load_Block_Cumulative_Difficulty(dbtx, blid) //+ chain.Load_Block_Difficulty(blid)

	// if any of the blocks  has a cumulative difficulty  more than  sync block, this situation affects  consensus, so mitigate it
	for i := range preblocks {
		cumulative_difficulty := chain.Load_Block_Cumulative_Difficulty(dbtx, preblocks[i]) // + chain.Load_Block_Difficulty(preblocks[i])

		//if cumulative_difficulty >= sync_block_cumulative_difficulty {
		if cumulative_difficulty.Cmp(sync_block_cumulative_difficulty) >= 0 {
			rlog.Warnf("Mitigating CONSENSUS issue on block %s height %d  child %s cdiff %d sync block cdiff %d", blid, height, preblocks[i], cumulative_difficulty, sync_block_cumulative_difficulty)
			return false
		}

	}

	return true
}


// key is string of blid and appendded chain height
var tipbase_cache,_ = hashicorp_lru.New(10240)

// base of a tip is last known sync point
// weight of bases in mentioned in term of height
// this must not employ any cache
func (chain *Blockchain) FindTipBase(dbtx storage.DBTX, blid crypto.Hash, chain_height int64) (bs BlockScore) {

	// see if cache contains it
	if bsi,ok := tipbase_cache.Get(fmt.Sprintf("%s%d", blid,chain_height));ok{
		bs = bsi.(BlockScore)
		return bs
	}

	defer func(){ // capture return value of bs to cache
		z := bs
		tipbase_cache.Add(fmt.Sprintf("%s%d", blid,chain_height),z)
	}()


	// if we are genesis return genesis block as base

	/* bl,err := chain.Load_BL_FROM_ID(blid)

	  if err != nil {
	   panic(fmt.Sprintf("Block NOT found %s", blid))
	  }
	  if len(bl.Tips) == 0 {
	       gbl := Generate_Genesis_Block()

	//      logger.Infof("Return genesis block as base")
	      return BlockScore{gbl.GetHash(),0}
	  }

	  bases := make([]BlockScore,len(bl.Tips),len(bl.Tips))
	  for i := range bl.Tips{
	      if chain.IsBlockSyncBlockHeight(bl.Tips[i]){
	        return BlockScore{bl.Tips[i], chain.Load_Height_for_BL_ID(bl.Tips[i])}
	      }
	       bases[i] = chain.FindTipBase(bl.Tips[i])
	  }*/

	tips := chain.Get_Block_Past(dbtx, blid)
	if len(tips) == 0 {
		gbl := Generate_Genesis_Block()
		bs = BlockScore{gbl.GetHash(), 0, nil}
		return
	}

	bases := make([]BlockScore, len(tips), len(tips))
	for i := range tips {
		if chain.IsBlockSyncBlockHeightSpecific(dbtx, tips[i], chain_height) {
			rlog.Tracef(2, "SYNC block %s", tips[i])
			bs = BlockScore{tips[i], chain.Load_Height_for_BL_ID(dbtx, tips[i]), nil}
			return
		}
		bases[i] = chain.FindTipBase(dbtx, tips[i], chain_height)
	}

	sort_ascending_by_height(bases)

	//   logger.Infof("return BASE %s",bases[0])
	bs = bases[0]
	return bs
}

// this will find the sum of  work done ( skipping any repetive nodes )
// all the information is privided in unique_map
func (chain *Blockchain) FindTipWorkScore_internal(dbtx storage.DBTX, unique_map map[crypto.Hash]*big.Int, blid crypto.Hash, base crypto.Hash, base_height int64) {
	/*bl,err := chain.Load_BL_FROM_ID(blid)
	  if err != nil {
	   panic(fmt.Sprintf("Block NOT found %s", blid))
	  }



	  for i := range bl.Tips{
	      if _,ok := unique_map[bl.Tips[i]];!ok{

	          ordered := chain.Is_Block_Topological_order(bl.Tips[i])
	          if !ordered ||
	              ordered && chain.Load_Block_Topological_order(bl.Tips[i]) >= chain.Load_Block_Topological_order(base){
	                 chain.FindTipWorkScore_internal(unique_map,bl.Tips[i],base,base_height) // recursively process any nodes
	              }
	      }
	  }*/

	tips := chain.Get_Block_Past(dbtx, blid)

	for i := range tips {
		if _, ok := unique_map[tips[i]]; !ok {

			ordered := chain.Is_Block_Topological_order(dbtx, tips[i])

			if !ordered {
				chain.FindTipWorkScore_internal(dbtx, unique_map, tips[i], base, base_height) // recursively process any nodes
				//logger.Infof("IBlock is not ordered %s", tips[i])
			} else if ordered && chain.Load_Block_Topological_order(dbtx, tips[i]) >= chain.Load_Block_Topological_order(dbtx, base) {
				chain.FindTipWorkScore_internal(dbtx, unique_map, tips[i], base, base_height) // recursively process any nodes

				//logger.Infof("IBlock ordered %s %d %d", tips[i],chain.Load_Block_Topological_order(tips[i]), chain.Load_Block_Topological_order(base) )
			}
		}
	}

	unique_map[blid] = chain.Load_Block_Difficulty(dbtx, blid)

}

type cachekey struct {
	blid        crypto.Hash
	base        crypto.Hash
	base_height int64
}

// find the score of the tip  in reference to  a base (NOTE: base is always a sync block otherwise results will be wrong )
func (chain *Blockchain) FindTipWorkScore(dbtx storage.DBTX, blid crypto.Hash, base crypto.Hash, base_height int64) (map[crypto.Hash]*big.Int, *big.Int) {

	//logger.Infof("BASE %s",base)
	if tmp_map_i, ok := chain.lrucache_workscore.Get(cachekey{blid, base, base_height}); ok {
		work_score := tmp_map_i.(map[crypto.Hash]*big.Int)

		map_copy := map[crypto.Hash]*big.Int{}
		score := new(big.Int).SetInt64(0)
		for k, v := range work_score {
			map_copy[k] = v
			score.Add(score, v)
		}
		return map_copy, score
	}

	bl, err := chain.Load_BL_FROM_ID(dbtx, blid)
	if err != nil {
		panic(fmt.Sprintf("Block NOT found %s", blid))
	}
	unique_map := map[crypto.Hash]*big.Int{}

	for i := range bl.Tips {
		if _, ok := unique_map[bl.Tips[i]]; !ok {
			//if chain.Load_Height_for_BL_ID(bl.Tips[i]) >  base_height {
			//    chain.FindTipWorkScore_internal(unique_map,bl.Tips[i],base,base_height) // recursively process any nodes
			//}

			ordered := chain.Is_Block_Topological_order(dbtx, bl.Tips[i])
			if !ordered {
				chain.FindTipWorkScore_internal(dbtx, unique_map, bl.Tips[i], base, base_height) // recursively process any nodes
				//   logger.Infof("Block is not ordered %s", bl.Tips[i])
			} else if ordered && chain.Load_Block_Topological_order(dbtx, bl.Tips[i]) >= chain.Load_Block_Topological_order(dbtx, base) {
				chain.FindTipWorkScore_internal(dbtx, unique_map, bl.Tips[i], base, base_height) // recursively process any nodes

				// logger.Infof("Block ordered %s %d %d", bl.Tips[i],chain.Load_Block_Topological_order(bl.Tips[i]), chain.Load_Block_Topological_order(base) )
			}
		}
	}

	if base != blid {
		unique_map[base] = chain.Load_Block_Cumulative_Difficulty(dbtx, base)
		// add base cumulative score
		// base_work := chain.Load_Block_Cumulative_Difficulty(base)
		// gbl:=Generate_Genesis_Block()
		// _, base_work  := chain.FindTipWorkScore(base, gbl.GetHash(),0)
		//unique_map[base]= base_work
		//unique_map[base] = new(big.Int).SetUint64(base_work)
	}

	/* if base_work == 0 {
	    logger.Infof("base Work done is zero %s", base)
	}*/

	unique_map[blid] = chain.Load_Block_Difficulty(dbtx, blid)
	//unique_map[blid]= work_done

	//if work_done == 0 {
	//    logger.Infof("Work done is zero")
	//}

	score := new(big.Int).SetInt64(0)
	for _, v := range unique_map {
		score.Add(score, v)
	}

	//set in cache, save a copy in cache
	{
		map_copy := map[crypto.Hash]*big.Int{}
		for k, v := range unique_map {
			map_copy[k] = v
		}
		chain.lrucache_workscore.Add(cachekey{blid, base, base_height}, map_copy)
	}

	return unique_map, score

}

// this function finds a common base  which can be used to compare tips
// weight is replace by height
func (chain *Blockchain) find_common_base(dbtx storage.DBTX, tips []crypto.Hash) (base crypto.Hash, base_height int64) {


	scores := make([]BlockScore, len(tips), len(tips))

	// var base crypto.Hash
	var best_height int64
	for i := range tips {
		tip_height := chain.Load_Height_for_BL_ID(dbtx, tips[i])
		if tip_height > best_height{
			best_height = tip_height
		}
	}


	for i := range tips {
		scores[i] = chain.FindTipBase(dbtx, tips[i],best_height) // we should chose the lowest weight
		scores[i].Height = chain.Load_Height_for_BL_ID(dbtx, scores[i].BLID)
	}
	// base is the lowest height
	sort_ascending_by_height(scores)

	base = scores[0].BLID
	base_height = scores[0].Height

	return

}

// this function finds a common base  which can be used to compare tips based on cumulative difficulty
func (chain *Blockchain) find_best_tip(dbtx storage.DBTX, tips []crypto.Hash, base crypto.Hash, base_height int64) (best crypto.Hash) {

	tips_scores := make([]BlockScore, len(tips), len(tips))

	for i := range tips {
		tips_scores[i].BLID = tips[i] // we should chose the lowest weight
		_, tips_scores[i].Cumulative_Difficulty = chain.FindTipWorkScore(dbtx, tips[i], base, base_height)
	}

	sort_descending_by_cumulative_difficulty(tips_scores)

	best = tips_scores[0].BLID
	//   base_height = scores[0].Weight

	return best

}

func (chain *Blockchain) calculate_mainchain_distance_internal_recursive(dbtx storage.DBTX, unique_map map[crypto.Hash]int64, blid crypto.Hash) {
	tips := chain.Get_Block_Past(dbtx, blid)
	for i := range tips {
		ordered := chain.Is_Block_Topological_order(dbtx, tips[i])
		if ordered {
			unique_map[tips[i]] = chain.Load_Height_for_BL_ID(dbtx, tips[i])
		} else {
			chain.calculate_mainchain_distance_internal_recursive(dbtx, unique_map, tips[i]) // recursively process any nodes
		}
	}
	return
}

// NOTE: some of the past may not be in the main chain  right now and need to be travelled recursively
// distance is number of hops to find a node, which is itself
func (chain *Blockchain) calculate_mainchain_distance(dbtx storage.DBTX, blid crypto.Hash) int64 {

	unique_map := map[crypto.Hash]int64{}
	//tips := chain.Get_Block_Past(dbtx, blid)

	//fmt.Printf("tips  %+v \n", tips)

	// if the block is already in order, no need to look back

	ordered := chain.Is_Block_Topological_order(dbtx, blid)
	if ordered {
		unique_map[blid] = chain.Load_Height_for_BL_ID(dbtx, blid)
	} else {
		chain.calculate_mainchain_distance_internal_recursive(dbtx, unique_map, blid)
	}

	//for i := range tips {
	//}

	//fmt.Printf("unique_map %+v \n", unique_map)

	lowest_height := int64(0x7FFFFFFFFFFFFFFF) // max possible
	// now we need to find the lowest height
	for k, v := range unique_map {
		_ = k
		if lowest_height >= v {
			lowest_height = v
		}
	}

	return int64(lowest_height)
}

// converts a DAG's partial order into a full order, this function is recursive
// generate full order should be only callled on the basis of base blocks which satisfy sync block properties as follows
// generate full order is called on maximum weight tip at every tip change
// blocks are ordered recursively, till we find a find a block  which is already in the chain
func (chain *Blockchain) Generate_Full_Order(dbtx storage.DBTX, blid crypto.Hash, base crypto.Hash, base_height int64, level int) (order_bucket []crypto.Hash) {

	// return from cache if possible
	if tmp_order, ok := chain.lrucache_fullorder.Get(cachekey{blid, base, base_height}); ok {
		order := tmp_order.([]crypto.Hash)
		order_bucket = make([]crypto.Hash, len(order), len(order))
		copy(order_bucket, order[0:])
		return
	}

	bl, err := chain.Load_BL_FROM_ID(dbtx, blid)
	if err != nil {
		panic(fmt.Sprintf("Block NOT found %s", blid))
	}

	if len(bl.Tips) == 0 {
		gbl := Generate_Genesis_Block()
		order_bucket = append(order_bucket, gbl.GetHash())
		return
	}

	// if the block has been previously ordered,  stop the recursion and return it as base
	//if chain.Is_Block_Topological_order(blid){
	if blid == base {
		order_bucket = append(order_bucket, blid)
		// logger.Infof("Generate order base reached  base %s", base)
		return
	}

	// we need to order previous tips first
	var tips_scores []BlockScore
	//tips_scores := make([]BlockScore,len(bl.Tips),len(bl.Tips))

	node_maps := map[crypto.Hash]map[crypto.Hash]*big.Int{}
	_ = node_maps
	for i := range bl.Tips {

		ordered := chain.Is_Block_Topological_order(dbtx, bl.Tips[i])

		if !ordered {
			var score BlockScore
			score.BLID = bl.Tips[i]
			//node_maps[bl.Tips[i]], score.Weight = chain.FindTipWorkScore(bl.Tips[i],base,base_height)
			score.Cumulative_Difficulty = chain.Load_Block_Cumulative_Difficulty(dbtx, bl.Tips[i])

			tips_scores = append(tips_scores, score)

		} else if ordered && chain.Load_Block_Topological_order(dbtx, bl.Tips[i]) >= chain.Load_Block_Topological_order(dbtx, base) {

			//  logger.Infof("Generate order topo order wrt base %d %d", chain.Load_Block_Topological_order(dbtx,bl.Tips[i]), chain.Load_Block_Topological_order(dbtx,base))
			var score BlockScore
			score.BLID = bl.Tips[i]

			//score.Weight = chain.Load_Block_Cumulative_Difficulty(bl.Tips[i])
			score.Cumulative_Difficulty = chain.Load_Block_Cumulative_Difficulty(dbtx, bl.Tips[i])

			tips_scores = append(tips_scores, score)
		}

	}

	sort_descending_by_cumulative_difficulty(tips_scores)

	// now we must add the nodes in the topographical order

	for i := range tips_scores {
		tmp_bucket := chain.Generate_Full_Order(dbtx, tips_scores[i].BLID, base, base_height, level+1)
		for j := range tmp_bucket {
			//only process if  this block is unsettled
			//if !chain.IsBlockSettled(tmp_bucket[j]) {
			// if order is already decided, do not order it again
			if !sliceExists(order_bucket, tmp_bucket[j]) {
				order_bucket = append(order_bucket, tmp_bucket[j])
			}
			//}
		}
	}
	// add self to the end, since all past nodes have been ordered
	order_bucket = append(order_bucket, blid)

	//  logger.Infof("Generate Order %s %+v  %+v", blid , order_bucket, tips_scores)

	//set in cache, save a copy in cache
	{
		order_copy := make([]crypto.Hash, len(order_bucket), len(order_bucket))
		copy(order_copy, order_bucket[0:])

		chain.lrucache_fullorder.Add(cachekey{blid, base, base_height}, order_copy)
	}

	if level == 0 {
		//logger.Warnf("generating full order for block %s %d", blid, level)
		/*
		   for i := range order_bucket{
		            logger.Infof("%2d  %s", i, order_bucket[i])
		   }
		*/
		//logger.Warnf("generating full order finished")
	}
	return
}

// this function finds a block at specific height whether it is a sync block
// if yes we generate a full order and settle the chain upto that level
/*
func (chain *Blockchain) SettleChainAtHeight(height uint64) {
    blocks := chain.Get_Blocks_At_Height(height)
     for i := range blocks {
         if !chain.IsBlockSettled(blocks[i]) && chain.IsBlockSyncBlock(blocks[i]){ // only unsettled blocks must be settled
             order := chain.Generate_Full_Order(blocks[i],0)
             logger.Warnf("Chain hash been settled at height %d order %+v", height,order)
         }

     }




}
*/
var node_map = map[crypto.Hash]bool{}

func collect_nodes(chain *Blockchain, dbtx storage.DBTX, blid crypto.Hash) {
	future := chain.Get_Block_Future(dbtx, blid)
	for i := range future {
		//node_map[future[i]]=true

		if _, ok := node_map[future[i]]; !ok {
			collect_nodes(chain, dbtx, future[i]) // recursive add node
		}
	}

	node_map[blid] = true

}

func writenode(chain *Blockchain, dbtx storage.DBTX, w *bufio.Writer, blid crypto.Hash) { // process a node, recursively

	collect_nodes(chain, dbtx, blid)

	sync_blocks := map[crypto.Hash]uint64{}

	for k, _ := range node_map {
		if chain.IsBlockSyncBlockHeight(dbtx, k) {
			// sync_blocks = append(sync_blocks,
			sync_blocks[k] = uint64(chain.Load_Height_for_BL_ID(dbtx, k))
		}
	}

	w.WriteString(fmt.Sprintf("node [ fontsize=12 style=filled ]\n{\n"))
	for k := range node_map {

		//anticone := chain.Get_AntiCone_Unsettled(k)

		color := "white"

		if chain.IsBlockSyncBlockHeight(dbtx, k) {
			color = "green"
		}

		/*
		   logger.Infof("Scores for %s",k)

		   height := chain.Load_Height_for_BL_ID(k)

		   for base,base_height := range  sync_blocks {
		       if height > base_height {
		           work_data, work_score := chain.FindTipWorkScore(k,base,base_height)
		        logger.Infof("Scores base %s height %5d  base_height %5d work %d  fff",base,height, base_height,work_score)

		        _ = work_data
		        for k1,v1 := range work_data{
		            _ = k1
		            _ = v1
		           // logger.Infof("score consists of %s %d  topo index %d",k1,v1, chain.Load_Block_Topological_order(k1))
		        }

		        full_order := chain.Generate_Full_Order(k, base, base_height,0)
		        for j := range full_order {
		            _ = j
		        // logger.Infof("full order %d %s", j, full_order[j])
		        }

		       }
		   }
		*/

		/*
		   if  len(anticone) >=4{
		       color = "red"
		   }

		   if chain.IsBlockSyncBlock(k) &&  len(anticone) >=4{
		       color = "orange"
		   }
		*/

		/*gbl := Generate_Genesis_Block()
		gbl_map, cumulative_difficulty := chain.FindTipWorkScore(dbtx, k, gbl.GetHash(), 0)

		if cumulative_difficulty.Cmp(chain.Load_Block_Cumulative_Difficulty(dbtx, k)) != 0 {
			logger.Infof("workmap from genesis  MISMATCH  blid %s height %d", k, chain.Load_Height_for_BL_ID(dbtx, k))
			for k, v := range gbl_map {
				logger.Infof("%s %d", k, v)
			}

		}*/

		//w.WriteString(fmt.Sprintf("L%s  [ fillcolor=%s label = \"%s %d height %d score %d stored %d order %d\"  ];\n", k.String(), color, k.String(), 0, chain.Load_Height_for_BL_ID(dbtx, k), cumulative_difficulty, chain.Load_Block_Cumulative_Difficulty(dbtx, k), chain.Load_Block_Topological_order(dbtx, k)))
		w.WriteString(fmt.Sprintf("L%s  [ fillcolor=%s label = \"%s %d height %d score %d stored %d order %d\"  ];\n", k.String(), color, k.String(), 0, chain.Load_Height_for_BL_ID(dbtx, k), 0, chain.Load_Block_Cumulative_Difficulty(dbtx, k), chain.Load_Block_Topological_order(dbtx, k)))
	}
	w.WriteString(fmt.Sprintf("}\n"))

	// now dump the interconnections
	for k := range node_map {
		future := chain.Get_Block_Future(dbtx, k)
		for i := range future {
			w.WriteString(fmt.Sprintf("L%s -> L%s ;\n", k.String(), future[i].String()))
		}

	}
}

func WriteBlockChainTree(chain *Blockchain, filename string) (err error) {

	dbtx, err := chain.store.BeginTX(false)
	if err != nil {
		logger.Warnf("Could NOT add block to chain. Error opening writable TX, err %s", err)
		return
	}

	defer dbtx.Rollback()

	f, err := os.Create(filename)
	if err != nil {
		return
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()
	w.WriteString("digraph dero_blockchain_graph { \n")

	blid, err := chain.Load_Block_Topological_order_at_index(nil, 158800)
	if err != nil {
		logger.Warnf("Cannot get block  at topoheight %d err: %s", 158800, err)
		return
	}

	writenode(chain, dbtx, w, blid)
	/*g := Generate_Genesis_Block()
	writenode(chain, dbtx, w, g.GetHash())
	*/
	w.WriteString("}\n")

	return
}

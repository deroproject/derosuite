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

//import "os"
import "fmt"
import "sort"
import "sync"
import "time"
import "bytes"
import "sync/atomic"
import "runtime/debug"

import log "github.com/sirupsen/logrus"
import "github.com/romana/rlog"

import "github.com/deroproject/derosuite/config"
import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/storage"
import "github.com/deroproject/derosuite/difficulty"
import "github.com/deroproject/derosuite/crypto/ringct"
import "github.com/deroproject/derosuite/block"
import "github.com/deroproject/derosuite/transaction"
import "github.com/deroproject/derosuite/checkpoints"
import "github.com/deroproject/derosuite/blockchain/mempool"
import "github.com/deroproject/derosuite/blockchain/inputmaturity"

// all components requiring access to blockchain must use , this struct to communicate
// this structure must be update while mutex
type Blockchain struct {
	store       storage.Store // interface to storage layer
	Height      uint64        // chain height is always 1 more than block
	height_seen uint64        // height seen on peers
	Top_ID      crypto.Hash   // id of the top block
	Difficulty  uint64        // current cumulative difficulty
	Mempool     *mempool.Mempool
	Exit_Event  chan bool // blockchain is shutting down and we must quit ASAP

	Top_Block_Median_Size uint64 // median block size of current top block
	Top_Block_Base_Reward uint64 // top block base reward

	checkpints_disabled bool // are checkpoints disabled

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
	init_static_checkpoints()          // init some hard coded checkpoints
	chain.store = storage.Bolt_backend // setup backend
	chain.store.Init(params)           // init backend

	chain.checkpints_disabled = params["--disable-checkpoints"].(bool)

	chain.Exit_Event = make(chan bool) // init exit channel

	// init mempool before chain starts
	chain.Mempool, err = mempool.Init_Mempool(params)

	// we need to check mainnet/testnet check whether the genesis block matches the testnet/mainet
	// mean whether the user is trying to use mainnet db with testnet option or vice-versa
	if chain.Block_Exists(config.Mainnet.Genesis_Block_Hash) || chain.Block_Exists(config.Testnet.Genesis_Block_Hash) {

		if globals.IsMainnet() && !chain.Block_Exists(config.Mainnet.Genesis_Block_Hash) {
			logger.Fatalf("Tryng to use a testnet database with mainnet, please add --testnet option")
		}

		if !globals.IsMainnet() && chain.Block_Exists(config.Testnet.Genesis_Block_Hash) {
			logger.Fatalf("Tryng to use a mainnet database with testnet, please remove --testnet option")
		}

	}

	// genesis block not in chain, add it to chain, together with its miner tx
	// make sure genesis is in the store
	if !chain.Block_Exists(globals.Config.Genesis_Block_Hash) {
		logger.Debugf("Genesis block not in store, add it now")
		var complete_block block.Complete_Block
		bl := Generate_Genesis_Block()
		complete_block.Bl = &bl

		if !chain.Add_Complete_Block(&complete_block) {
			logger.Fatalf("Failed to add genesis block, we can no longer continue")
		}

	}

	// load the chain from the disk
	chain.Initialise_Chain_From_DB()

	if chain.checkpints_disabled {
		logger.Infof("Internal Checkpoints are disabled")
	} else {
		logger.Debugf("Internal Checkpoints are enabled")
	}

	_ = err

	atomic.AddUint32(&globals.Subsystem_Active, 1) // increment subsystem

	//go chain.Handle_Block_Event_Loop()
	//go chain.Handle_Transaction_Event_Loop()

	/*for i := uint64(0); i < 100;i++{
		block_id,_ := chain.Load_BL_ID_at_Height(i)
		chain.write_output_index(block_id)
	}*/

	//	chain.Inject_Alt_Chain()

	return &chain, nil
}

// this function is called to read blockchain state from DB
// It is callable at any point in time
func (chain *Blockchain) Initialise_Chain_From_DB() {
	chain.Lock()
	defer chain.Unlock()

	// locate top block

	chain.Top_ID = chain.Load_TOP_ID()
	chain.Height = (chain.Load_Height_for_BL_ID(chain.Top_ID) + 1)
	chain.Difficulty = chain.Get_Difficulty()
	chain.Top_Block_Median_Size = chain.Get_Median_BlockSize_At_Block(chain.Top_ID)
	chain.Top_Block_Base_Reward = chain.Load_Block_Reward(chain.Top_ID)

	logger.Infof("Chain Top Block %s Height %d", chain.Top_ID, chain.Height)

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

func (chain *Blockchain) Get_Height() uint64 {
	return chain.Height
}

func (chain *Blockchain) Get_Top_ID() crypto.Hash {
	return chain.Top_ID
}

func (chain *Blockchain) Get_Difficulty() uint64 {
	return chain.Get_Difficulty_At_Block(chain.Top_ID)
}

func (chain *Blockchain) Get_Network_HashRate() uint64 {
	return chain.Get_Difficulty_At_Block(chain.Top_ID) / config.BLOCK_TIME
}

// confirm whether the block exist in the data
// this only confirms whether the block has been downloaded
// a separate check is required, whether the block is valid ( satifies PoW and other conditions)
// we will not add a block to store, until it satisfies PoW
func (chain *Blockchain) Block_Exists(h crypto.Hash) bool {
	_, err := chain.Load_BL_FROM_ID(h)
	if err == nil {
		return true
	}
	return false
}

// this is the only entrypoint for new txs in the chain
// add a transaction to MEMPOOL,
// verifying everything  means everything possible
// TODO: currently we are not verifying fees, its on TODO list
// this only change mempool, no DB changes
func (chain *Blockchain) Add_TX_To_Pool(tx *transaction.Transaction) (result bool) {

	// Coin base TX can not come through this path
	if tx.IsCoinbase() {
		logger.WithFields(log.Fields{"txid": tx.GetHash()}).Warnf("TX rejected  coinbase tx cannot appear in mempool")
		return false
	}

	// check whether enough fees is provided in the transaction
	// calculate dynamic_fees_per_kb
	dynamic_fees_per_kb := uint64(0)
	previous_height := chain.Load_Height_for_BL_ID(chain.Get_Top_ID())
	if previous_height >= 2 {
		dynamic_fees_per_kb = chain.Get_Dynamic_Fee_Rate(previous_height)
	}

	// check whether the fee provided is enoough
	calculated_fee := chain.Calculate_TX_fee(dynamic_fees_per_kb, uint64(len(tx.Serialize())))
	provided_fee := tx.RctSignature.Get_TX_Fee() // get fee from tx

	if ((calculated_fee * 98) / 100) > provided_fee { // 2 % margin see blockchain.cpp L 2913
		logger.WithFields(log.Fields{"txid": tx.GetHash()}).Warnf("TX rejected due to low fees  provided fee %d calculated fee %d", provided_fee, calculated_fee)
		return false
	}

	if chain.Verify_Transaction_NonCoinbase(tx) {
		if chain.Mempool.Mempool_Add_TX(tx, 0) {
			logger.Debugf("successfully added tx to pool")
			return true
		} else {
			logger.Debugf("TX rejected by pool")
			return false
		}
	}

	logger.Warnf("Incoming TX could not be verified")
	return false

}

// this is the only entrypoint for new / old blocks even for genesis block
// this will add the entire block atomically to the chain
// this is the only function which can add blocks to the chain
// this is exported, so ii can be fed new blocks by p2p layer
// genesis block is no different
// TODO: we should stop mining while adding the new block
func (chain *Blockchain) Add_Complete_Block(cbl *block.Complete_Block) (result bool) {

	var block_hash crypto.Hash
	chain.Lock()
	defer chain.Unlock()
	defer func() {

		// safety so if anything wrong happens, verification fails
		if r := recover(); r != nil {
			logger.Warnf("Recovered while adding new block, Stack trace below block_hash %s", block_hash)
			logger.Warnf("Stack trace  \n%s", debug.Stack())
			result = false
		}

		if result == true { // block was successfully added, commit it atomically
			chain.store.Commit()
			chain.store.Sync() // sync the DB to disk after every execution of this function
		} else {
			chain.store.Rollback() // if block could not be added, rollback all changes to previous block
		}
	}()

	bl := cbl.Bl // small pointer to block

	// first of all lets do some quick checks
	// before doing extensive checks
	result = false

	block_hash = bl.GetHash()
	block_logger := logger.WithFields(log.Fields{"blid": block_hash})

	// check if block already exist skip it
	if chain.Block_Exists(block_hash) {
		block_logger.Debugf("block already in chain skipping it ")
		return
	}

	// make sure prev_hash refers to some point in our our chain
	// there is an edge case, where we know child but still donot know parent
	// this might be some some corrupted miner or initial sync
	if block_hash != globals.Config.Genesis_Block_Hash && !chain.Block_Exists(bl.Prev_Hash) {
		// TODO we must queue this block for say 60 minutes, if parents donot appear it, discard it
		block_logger.Warnf("Prev_Hash  no where in the chain, skipping it till we get a parent ")
		return
	}

	// make sure time is NOT too much into future
	// if clock diff is more than 2 hrs, reject the block
	if bl.Timestamp > (uint64(time.Now().Unix()) + config.CRYPTONOTE_BLOCK_FUTURE_TIME_LIMIT) {
		block_logger.Warnf("Block timestamp is too much into future, make sure that system clock is correct")
		return
	}

	// verify that the clock is not being run in reverse
	median_timestamp := chain.Get_Median_Timestamp_At_Block(bl.Prev_Hash)
	if bl.Timestamp < median_timestamp {
		block_logger.Warnf("Block timestamp  %d is less than median timestamp (%d) of %d blocks", bl.Timestamp, median_timestamp, config.BLOCKCHAIN_TIMESTAMP_CHECK_WINDOW)
		return
	}

	// check  a small list 100 hashes whether they have been reached
	if IsCheckPointKnown_Static(block_hash, chain.Load_Height_for_BL_ID(bl.Prev_Hash)+1) {
		rlog.Tracef(1, "Static Checkpoint reached at height %d", chain.Load_Height_for_BL_ID(bl.Prev_Hash)+1)
	}

	rlog.Tracef(1, "Checking Known checkpoint %s at height %d", block_hash, chain.Load_Height_for_BL_ID(bl.Prev_Hash)+1)
	// disable security checks if checkpoint is already known
	if chain.checkpints_disabled || !checkpoints.IsCheckPointKnown(block_hash, chain.Load_Height_for_BL_ID(bl.Prev_Hash)+1) {
		rlog.Tracef(1, "Unknown checkpoint %s at height %d, verifying throughly", block_hash, chain.Load_Height_for_BL_ID(bl.Prev_Hash)+1)
		// Verify Blocks Proof-Of-Work
		PoW := bl.GetPoWHash()
		current_difficulty := chain.Get_Difficulty_At_Block(bl.Prev_Hash)

		if block_hash != globals.Config.Genesis_Block_Hash {
			logger.Debugf("Difficulty at height %d    is %d", chain.Load_Height_for_BL_ID(bl.Prev_Hash), current_difficulty)
		}
		// check if the PoW is satisfied
		if !difficulty.CheckPowHash(PoW, current_difficulty) { // if invalid Pow, reject the bloc
			block_logger.Warnf("Block has invalid PoW, rejecting it")
			return false
		}

		// TODO we need to verify block size whether it crosses the limits

		// we need to verify each and every tx contained in the block, sanity check everything
		// first of all check, whether all the tx contained in the block, match their hashes
		{
			if len(bl.Tx_hashes) != len(cbl.Txs) {
				block_logger.Warnf("Block says it has %d txs , however complete block contained %d txs", len(bl.Tx_hashes), len(cbl.Txs))
				return false
			}

			// first check whether the complete block contains any diplicate hashes
			tx_checklist := map[crypto.Hash]bool{}
			for i := 0; i < len(bl.Tx_hashes); i++ {
				tx_checklist[bl.Tx_hashes[i]] = true
			}

			if len(tx_checklist) != len(bl.Tx_hashes) { // block has duplicate tx, reject
				block_logger.Warnf("Block has %d  duplicate txs, reject it", len(bl.Tx_hashes)-len(tx_checklist))

			}
			// now lets loop through complete block, matching each tx
			// detecting any duplicates using txid hash
			for i := 0; i < len(cbl.Txs); i++ {
				tx_hash := cbl.Txs[i].GetHash()
				if _, ok := tx_checklist[tx_hash]; !ok {
					// tx is NOT found in map, RED alert reject the block
					block_logger.Warnf("Block says it has tx %s, but complete block does not have it", tx_hash)
					return false
				}
			}
		}

		// another check, whether the tx contains any duplicate key images within the block
		// block wide duplicate input detector
		{
			block_pool, _ := mempool.Init_Block_Mempool(nil)
			for i := 0; i < len(cbl.Txs); i++ {
				if !block_pool.Mempool_Add_TX(cbl.Txs[i], 0) { // block pool will reject any tx which are duplicates or double spend attacks
					block_logger.Warnf("Double spend attack  %s, rejecting ", cbl.Txs[i].GetHash())
					return false
				}
			}
		}

		// now we need to verify each and every tx in detail
		// verify coinbase tx
		if chain.Get_Height() > 5 { // skip checks for first 5 blocks
			if !chain.Verify_Transaction_Coinbase(cbl, &bl.Miner_tx) {
				block_logger.Warnf("Miner tx failed verification  rejecting ")
				return false
			}
		}

		/*
			// verify all non coinbase tx, single threaded, we have a multithreaded version below
			for i := 0 ; i < len(cbl.Txs); i++ {
				if !chain.Verify_Transaction_NonCoinbase(cbl.Txs[i]){
					logger.Warnf("Non Coinbase tx failed verification  rejecting " )
				 	return false
				}
			}
		*/

		fail_count := uint64(0)
		wg := sync.WaitGroup{}
		wg.Add(len(cbl.Txs)) // add total number of tx as work
		for i := 0; i < len(cbl.Txs); i++ {
			go func(j int) {
				if !chain.Verify_Transaction_NonCoinbase(cbl.Txs[j]) { // transaction verification failed
					atomic.AddUint64(&fail_count, 1) // increase fail count by 1
				}
				wg.Done()
			}(i)
		}

		wg.Wait()           // wait for verifications to finish
		if fail_count > 0 { // check the result
			block_logger.Warnf("Block verification failed  rejecting ")
			return false
		}

	} // checkpoint based validation completed here
	// if checkpoint is found, we land here

	// we are here means everything looks good, proceed and save to chain

	// discard the transactions from mempool if they are present there
	chain.Mempool.Monitor()

	for i := 0; i < len(cbl.Txs); i++ {
		txid := cbl.Txs[i].GetHash()
		if chain.Mempool.Mempool_TX_Exist(txid) {
			rlog.Tracef(1, "Deleting TX from pool txid=%s", txid)
			chain.Mempool.Mempool_Delete_TX(txid)
		}
	}

	// save all the txs
	// and then save the block
	{ // first lets save all the txs, together with their link to this block as height
		height := uint64(0)
		if block_hash != globals.Config.Genesis_Block_Hash {
			// get height from parent block
			height = chain.Load_Height_for_BL_ID(bl.Prev_Hash)
			height++
		}
		for i := 0; i < len(cbl.Txs); i++ {
			chain.Store_TX(cbl.Txs[i], height)
		}
	}

	// check we need to  extend the chain or do a soft fork
	// this condition is automatically satisfied by the genesis block ( since golang gives everything a zero value)
	if bl.Prev_Hash == chain.Top_ID /* lock_hash == globals.Config.Genesis_Block_Hash */ {
		// we need to extend the chain
		//log.Debugf("Extendin chain using block %x", block_hash )

		chain.Store_BL(bl)
		chain.consume_keyimages(block_hash) // consume all keyimages as spent

		chain.write_output_index(block_hash) // extract and store keys
		chain.Store_TOP_ID(block_hash)       // make new block top block
		//chain.Add_Child(bl.Prev_Hash, block_hash) // add the new block as chil
		chain.Store_Block_Child(bl.Prev_Hash, block_hash)

		chain.Store_BL_ID_at_Height(chain.Height, block_hash) // store height to block id mapping

		// lower the window, where top_id and  chain height are different
		chain.Height = chain.Height + 1 // increment height
		chain.Top_ID = block_hash       // set new top block id

		block_logger.Debugf("Chain extended new height %d", chain.Height)

		// every 20 block print a line
		if chain.Height%20 == 0 {
			block_logger.Infof("Chain Height %d", chain.Height)
		}

	} else { // a soft fork is in progress
		block_logger.Debugf("Soft Fork is in progress")
		chain.Chain_Add_And_Reorganise(bl)
	}

	result = true
	return // run any handlers necesary to atomically
}

/* the block we have is NOT at the top, it either belongs to an altchain or is an alternative */
func (chain *Blockchain) Chain_Add_And_Reorganise(bl *block.Block) (result bool) {
	block_hash := bl.GetHash()

	// check whether the parent already has a child
	parent_has_child := chain.Does_Block_Have_Child(bl.Prev_Hash)
	// first lets add ourselves to the chain
	chain.Store_BL(bl)
	chain.consume_keyimages(block_hash)
	if !parent_has_child {
		chain.Store_Block_Child(bl.Prev_Hash, block_hash)
		logger.Infof("Adding alternative block %s to alt chain top", block_hash)
	} else {
		logger.Infof("Adding alternative block %s", block_hash)

		// load existing children, there can be more than 1 in extremely rare case or unknown attacks
		children_list := chain.Load_Block_Children(bl.Prev_Hash)
		children_list = append(children_list, block_hash) // add ourselves to children list
		// store children excluding main child of prev block
		chain.Store_Block_Children(bl.Prev_Hash, children_list, chain.Load_Block_Child(bl.Prev_Hash))
	}

	// now we must trigger the recursive reorganise process from the parent block,
	// the recursion should always end at the genesis block
	// adding a block can cause chain reorganisation 1 time in 99.99% cases
	// but we are prepared for the case, which might occur due to alt-alt-chains

	chain.reorganise(block_hash)

	return true
}

type chain_data struct {
	hash        crypto.Hash
	cdifficulty uint64
	foundat     uint64 // when block was found
}

// NOTE: below algorithm is the core and and is used to achieve network consensus
// the best chain is found using the following algorithm
// cryptonote protocol algo is below
// compare cdiff, chain with higher diff wins, if diff is same, no reorg, this cause frequent splits

// new algo is this
// compare cdiff, chain with higher diff wins, if diff is same, go below
// compare time stamp, block with lower timestamp wins (since it has probable spread more than other blocks)
// if timestamps are same, block with lower block hash (No PoW involved) wins
//  block hash cannot be same

type bestChain []chain_data

func (s bestChain) Len() int      { return len(s) }
func (s bestChain) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s bestChain) Less(i, j int) bool {
	if s[i].cdifficulty > s[j].cdifficulty {
		return true
	}
	if s[i].cdifficulty < s[j].cdifficulty {
		return false
	}
	// we are here of if difficulty are same
	if s[i].foundat < s[j].foundat { // check if timestamps are diff
		return true
	}
	if s[i].foundat > s[j].foundat { // check if timestamps are diff
		return false
	}

	if bytes.Compare(s[i].hash[:], s[j].hash[:]) < 0 {
		return true
	}
	return false
}

// this function will recursive reorganise the chain, till the genesis block if required
// we are doing it this way as we can do away with too much book keeping
// this is processor and IO intensive in normal cases
func (chain *Blockchain) reorganise(block_hash crypto.Hash) {

	var children_data bestChain
	if block_hash == globals.Config.Genesis_Block_Hash {
		logger.Infof("Reorganisation completed successfully, we reached genesis block")
		return
	}

	// check if  the block mentioned has more than 1 child
	block_hash, found := chain.find_parent_with_children(block_hash)
	logger.Debugf("block with children (in reverse) %s  found %d", block_hash, found)
	if found {
		// reorganise chain at this block
		children := chain.Load_Block_Children(block_hash)
		if len(children) < 2 {
			panic(fmt.Sprintf("Children disappeared for block %s", block_hash))
		}
		main_chain := chain.Load_Block_Child(block_hash)

		// choose the best chain and make it parent
		for i := range children {
			top_hash := chain.Get_Top_Block(children[i])
			top_cdiff := chain.Load_Block_Cumulative_Difficulty(top_hash)
			timestamp := chain.Load_Block_Timestamp(children[i])
			children_data = append(children_data, chain_data{hash: children[i], cdifficulty: top_cdiff, foundat: timestamp})
		}

		sort.Sort(children_data)
		logger.Infof("Choosing best chain")
		for i := range children {
			logger.Infof("%d %+v\n", i, children_data[i])
		}

		best_chain := children_data[0].hash
		if main_chain == best_chain {
			logger.Infof("Main chain is already best, nothing to do")
			return
		} else {
			logger.Infof("Making alt chain -> main chain and vice-versa")
			// first lets fix up the connection
			chain.Store_Block_Child(block_hash, best_chain)              // store main connection
			chain.Store_Block_Children(block_hash, children, best_chain) // store remaining child

			// setup new height
			new_height := chain.Load_Height_for_BL_ID(chain.Get_Top_Block(best_chain)) + 1

			//also walk through all the new main chain till the top, setting output keys, first of all
			chain.write_output_index(block_hash) // extract and store keys
			loop_block_hash := block_hash
			for {
				chain.write_output_index(loop_block_hash) // extract and store keys
				chain.consume_keyimages(loop_block_hash)  // consume all keyimages

				// fix up height to block id mapping, which is used to find orphans later on
				height := chain.Load_Height_for_BL_ID(loop_block_hash)
				chain.Store_BL_ID_at_Height(height, loop_block_hash)

				// check if the block has child, if not , we are the top
				if !chain.Does_Block_Have_Child(loop_block_hash) {
					break
				}
				loop_block_hash = chain.Load_Block_Child(loop_block_hash) // continue searching the new top
			}

			// invalidate all transactionw contained within old main chain
			// validate all transactions in new main chain
			logger.Debugf("Invalidating all transactions with old main chain")
			logger.Debugf("Validating all transactions with old alt chain")

			// pushing alt_chain txs to mempool after verification
			loop_block_hash = main_chain // main chain at this point is the old chain
			for {
				// load the block
				bl, err := chain.Load_BL_FROM_ID(loop_block_hash)
				if err != nil {
					chain.revoke_keyimages(bl.GetHash()) // revoke all keyimages

					for i := 0; i < len(bl.Tx_hashes); i++ {
						tx, err := chain.Load_TX_FROM_ID(bl.Tx_hashes[i])
						if err != nil {
							if !chain.Verify_Transaction_NonCoinbase(tx) {
								logger.Warnf("Non Coinbase tx failed verification  rejecting ")

							} else { // tx passed verification add to mempool
								// TODO check whether the additiontion was successfull
								chain.Mempool.Mempool_Add_TX(tx, 0)

							}
						}
					}
				} else {
					logger.Debugf("error during chain reorganisation, failed to push alt chain TX to pool")
				}

				// check if the block has child, if not , we are the top
				if !chain.Does_Block_Have_Child(loop_block_hash) {
					break
				}
				loop_block_hash = chain.Load_Block_Child(loop_block_hash) // continue searching the new top
			}

			logger.Infof("Reorganise old height %d,  new height %d", chain.Get_Height(), new_height)

			chain.Top_ID = chain.Get_Top_Block(best_chain)
			chain.Height = new_height

			chain.Store_TOP_ID(chain.Top_ID) // make new block top block

			logger.Infof("Reorganise success")
		}

		// TODO if we need to support alt-alt chains, uncomment the code below
		//chain.reorganise(chain.Load_Block_Parent_ID(block_hash))
	}
}

/*
func (chain *Blockchain)find_best_chain(list []crypto.Hash)  best_child crypto.Hash {
	if len(list) < 2 {
		panic("Cannot find best child, when child_count = 1")
	}
}
*/

// find a block with 2 or more child,
// returns false, if we reach genesis block
func (chain *Blockchain) find_parent_with_children(block_hash crypto.Hash) (hash crypto.Hash, found bool) {
	// TODO we can also stop on the heighest checkpointed state, to save computing resources and time
	if block_hash == globals.Config.Genesis_Block_Hash {
		return hash, false // we do not have parent of  genesis block
	}
	for {
		// load children
		children := chain.Load_Block_Children(block_hash)
		if len(children) >= 2 {
			return block_hash, true
		}
		block_hash = chain.Load_Block_Parent_ID(block_hash)
		if block_hash == globals.Config.Genesis_Block_Hash {
			return hash, false // we do not have parent of  genesis block
		}
	}
}

// Finds whether a  block is orphan
// since we donot store any fields, we need to calculate/find the block as orphan
// using an algorithm
// find the block height and then relook up block using height
// if both are same, the block is good otherwise we treat it as orphan
func (chain *Blockchain) Is_Block_Orphan(hash crypto.Hash) bool {
	height := chain.Load_Height_for_BL_ID(hash)
	block_hash_at_height, _ := chain.Load_BL_ID_at_Height(height)
	if hash == block_hash_at_height {
		return false
	}
	return true
}

// this function will mark all the key images present in the block as requested
// this is done so as they cannot be respent
// mark is bool
func (chain *Blockchain) mark_keyimages(block_hash crypto.Hash, mark bool) bool {
	bl, err := chain.Load_BL_FROM_ID(block_hash)
	if err == nil {
		for i := 0; i < len(bl.Tx_hashes); i++ {
			tx, err := chain.Load_TX_FROM_ID(bl.Tx_hashes[i])
			if err != nil {
				logger.Debugf("TX loading error while marking keyimages as spent blid %s txid %s", block_hash, bl.Tx_hashes[i])
				return false
			} else {
				// mark keyimage as spent
				for i := 0; i < len(tx.Vin); i++ {
					k_image := tx.Vin[i].(transaction.Txin_to_key).K_image
					chain.Store_KeyImage(crypto.Hash(k_image), mark)
				}

			}
		}
	} else {
		logger.Debugf("BL loading error while marking keyimages as spent blid %s err %s", block_hash, err)
		return false
	}
	return true
}

//this will mark all the keyimages present in this block as spent
//this is done so as an input cannot be spent twice
func (chain *Blockchain) consume_keyimages(block_hash crypto.Hash) bool {
	return chain.mark_keyimages(block_hash, true)
}

//this will mark all the keyimages present in this block as unspent
//this is required during chain reorganisation
// when altchain becomes mainchain or viceversa,
// one of the chains needs to be markek unconsumed, so they can be consumed again
func (chain *Blockchain) revoke_keyimages(block_hash crypto.Hash) bool {
	return chain.mark_keyimages(block_hash, false)
}

/* this will only give you access to transactions which have been mined
 */
func (chain *Blockchain) Get_TX(hash crypto.Hash) (*transaction.Transaction, error) {
	tx, err := chain.Load_TX_FROM_ID(hash)

	return tx, err
}

// get difficulty at specific height  but height must be <= than current block chain height
func (chain *Blockchain) Get_Difficulty_At_Height(Height uint64) uint64 {

	if Height > chain.Get_Height() {
		logger.Warnf("Difficulty Requested for invalid Height Chain Height %d requested Height %d", chain.Get_Height(), Height)
		panic("Difficulty Requested for invalid Height")
	}
	// get block id at that height
	block_id, err := chain.Load_BL_ID_at_Height(Height)

	if err != nil {
		logger.Warnf("No Block at Height %d  , chain height %d", Height, chain.Get_Height())
		panic("No Block at Height")
	}

	// we have a block id, now Lets get the difficulty

	return chain.Get_Difficulty_At_Block(block_id)
}

// get difficulty at specific block_id, only condition is block must exist and must be connected
func (chain *Blockchain) Get_Difficulty_At_Block(block_id crypto.Hash) uint64 {

	var cumulative_difficulties []uint64
	var timestamps []uint64
	var zero_block crypto.Hash

	current_block_id := block_id
	// traverse chain from the block referenced, to max 30 blocks ot till genesis block is reached
	for i := 0; i < config.DIFFICULTY_BLOCKS_COUNT_V2; i++ {
		if current_block_id == globals.Config.Genesis_Block_Hash || current_block_id == zero_block {
			rlog.Tracef(2, "Reached genesis block for difficulty calculation %s", block_id)
			break // break we have reached genesis block
		}
		// read timestamp of block and cumulative difficulty at that block
		timestamp := chain.Load_Block_Timestamp(current_block_id)
		cdifficulty := chain.Load_Block_Cumulative_Difficulty(current_block_id)

		timestamps = append([]uint64{timestamp}, timestamps...)                             // prepend timestamp
		cumulative_difficulties = append([]uint64{cdifficulty}, cumulative_difficulties...) // prepend timestamp

		current_block_id = chain.Load_Block_Parent_ID(current_block_id)

	}
	return difficulty.Next_Difficulty(timestamps, cumulative_difficulties, config.BLOCK_TIME)
}

// get median time stamp at specific block_id, only condition is block must exist and must be connected
func (chain *Blockchain) Get_Median_Timestamp_At_Block(block_id crypto.Hash) uint64 {

	var timestamps []uint64
	var zero_block crypto.Hash

	current_block_id := block_id
	// traverse chain from the block referenced, to max 30 blocks ot till genesis block is researched
	for i := 0; i < config.BLOCKCHAIN_TIMESTAMP_CHECK_WINDOW; i++ {
		if current_block_id == globals.Config.Genesis_Block_Hash || current_block_id == zero_block {
			rlog.Tracef(4, "Reached genesis block for median calculation %s", block_id)
			break // break we have reached genesis block
		}
		// read timestamp of block and cumulative difficulty at that block
		timestamp := chain.Load_Block_Timestamp(current_block_id)
		timestamps = append(timestamps, timestamp) // append timestamp
		current_block_id = chain.Load_Block_Parent_ID(current_block_id)

	}

	return Median(timestamps)
}

// get median blocksize  at specific block_id, only condition is block must exist and must be connected
func (chain *Blockchain) Get_Median_BlockSize_At_Block(block_id crypto.Hash) uint64 {

	var block_sizes []uint64
	var zero_block crypto.Hash

	current_block_id := block_id
	// traverse chain from the block referenced, to max 30 blocks ot till genesis block is researched
	for i := uint64(0); i < config.CRYPTONOTE_REWARD_BLOCKS_WINDOW; i++ {
		if current_block_id == globals.Config.Genesis_Block_Hash || current_block_id == zero_block {
			rlog.Tracef(4, "Reached genesis block for median calculation %s", block_id)
			break // break we have reached genesis block
		}
		// read timestamp of block and cumulative difficulty at that block
		block_size := chain.Load_Block_Size(current_block_id)
		block_sizes = append(block_sizes, block_size) // append size
		current_block_id = chain.Load_Block_Parent_ID(current_block_id)

	}

	return Median(block_sizes)
}

// this function return the current top block, if we start at specific block
// this works for any blocks which were added
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

// verifies whether we are lagging
// return true if we need resync
// returns false if we are good and resync is not required
func (chain *Blockchain) IsLagging(peer_cdifficulty, peer_height uint64, peer_top_id crypto.Hash) bool {
	top_id := chain.Get_Top_ID()
	cdifficulty := chain.Load_Block_Cumulative_Difficulty(top_id)
	height := chain.Load_Height_for_BL_ID(top_id) + 1
	rlog.Tracef(3, "P_cdiff %d cdiff %d , P_BH %d BH %d, p_top %s top %s",
		peer_cdifficulty, cdifficulty,
		peer_height, height,
		peer_top_id, top_id)

	if peer_cdifficulty > cdifficulty {
		return true // peer's cumulative difficulty is more than ours , active resync
	}

	if peer_cdifficulty == cdifficulty && peer_top_id != top_id {
		return true // cumulative difficulty is  same but tops are different , active resync
	}

	return false
}

// This function will expand a transaction with all the missing info being reconstitued from the blockchain
// this also increases security since data is coming from the chain or being calculated
// basically this places data for ring signature verification
// REMEMBER to expand key images from the blockchain
// TODO we must enforce that the keyimages used are valid and specific outputs are unlocked
func (chain *Blockchain) Expand_Transaction_v2(tx *transaction.Transaction) (result bool) {

	result = false
	if tx.Version != 2 {
		panic("TX not version 2")
	}

	//if rctsignature is null

	// fill up the message hash first
	tx.RctSignature.Message = ringct.Key(tx.GetPrefixHash())

	// fill up the key images from the blockchain
	for i := 0; i < len(tx.Vin); i++ {
		tx.RctSignature.MlsagSigs[i].II = tx.RctSignature.MlsagSigs[i].II[:0] // zero it out
		tx.RctSignature.MlsagSigs[i].II = make([]ringct.Key, 1, 1)
		tx.RctSignature.MlsagSigs[i].II[0] = ringct.Key(tx.Vin[i].(transaction.Txin_to_key).K_image)
	}

	// now we need to fill up the mixring ctkey
	// one part is the destination address, second is the commitment mask from the outpk
	// mixring is stored in different ways for rctfull and simple

	switch tx.RctSignature.Get_Sig_Type() {

	case ringct.RCTTypeFull:
		// TODO, we need to make sure all ring are of same size

		if len(tx.Vin) > 1 {
			panic("unsipported rcctt full case please investigate")
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
				offset_data := chain.load_output_index(offset)

				// check maturity of inputs
				if !inputmaturity.Is_Input_Mature(chain.Get_Height(), offset_data.Height, offset_data.Unlock_Height, 1) {
					logger.Warnf("transaction using immature inputs from block %d chain height %d", offset_data.Height, chain.Get_Height())
					return false
				}

				tx.RctSignature.MixRing[m][n].Destination = offset_data.InKey.Destination
				tx.RctSignature.MixRing[m][n].Mask = offset_data.InKey.Mask
				//	fmt.Printf("%d %d dest %s\n",n,m, offset_data.InKey.Destination)
				//	fmt.Printf("%d %d mask %s\n",n,m, offset_data.InKey.Mask)

			}
		}

	case ringct.RCTTypeSimple:
		mixin := len(tx.Vin[0].(transaction.Txin_to_key).Key_offsets)
		_ = mixin
		tx.RctSignature.MixRing = make([][]ringct.CtKey, len(tx.Vin), len(tx.Vin))

		for n := 0; n < len(tx.Vin); n++ {

			tx.RctSignature.MixRing[n] = make([]ringct.CtKey, len(tx.Vin[n].(transaction.Txin_to_key).Key_offsets),
				len(tx.Vin[n].(transaction.Txin_to_key).Key_offsets))
			offset := uint64(0)
			for m := 0; m < len(tx.Vin[n].(transaction.Txin_to_key).Key_offsets); m++ {

				offset += tx.Vin[n].(transaction.Txin_to_key).Key_offsets[m]
				// extract the keys from specific offset
				offset_data := chain.load_output_index(offset)

				// check maturity of inputs
				if !inputmaturity.Is_Input_Mature(chain.Get_Height(), offset_data.Height, offset_data.Unlock_Height, 1) {
					logger.Warnf("transaction using immature inputs from block %d chain height %d", offset_data.Height, chain.Get_Height())
					return false
				}

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
// this function exists here because  only the chain knws the tx
//
func (chain *Blockchain) Block_Count_Vout(block_hash crypto.Hash) (count uint64) {
	count = 1 // miner tx is always present

	bl, err := chain.Load_BL_FROM_ID(block_hash)

	if err != nil {
		panic(fmt.Errorf("Cannot load  block for %s err %s", block_hash, err))
	}

	for i := 0; i < len(bl.Tx_hashes); i++ { // load all tx one by one
		tx, err := chain.Load_TX_FROM_ID(bl.Tx_hashes[i])
		if err != nil {
			panic(fmt.Errorf("Cannot load  tx for %s err %s", bl.Tx_hashes[i], err))
		}

		// tx has been loaded, now lets get the vout
		vout_count := uint64(len(tx.Vout))
		count += vout_count
	}
	return count
}

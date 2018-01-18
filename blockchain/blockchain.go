package blockchain


import "os"
import "fmt"
import "sort"
import "sync"
import "bytes"
import "sync/atomic"

import log "github.com/sirupsen/logrus"
import "github.com/romana/rlog"

import "github.com/deroproject/derosuite/config"
import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/storage"
import "github.com/deroproject/derosuite/difficulty"
import "github.com/deroproject/derosuite/crypto/ringct"



// all components requiring access to blockchain must use , this struct to communicate
// this structure must be update while mutex
type Blockchain struct {
	store       storage.Store // interface to storage layer
	Height      uint64        // chain height is always 1 more than block
	height_seen uint64        // height seen on peers
	Top_ID      crypto.Hash   // id of the top block
	Difficulty  uint64        // current cumulative difficulty
	Mempool     *Mempool

	sync.RWMutex
}

var logger *log.Entry

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
	init_checkpoints()                 // init some hard coded checkpoints
	chain.store = storage.Bolt_backend // setup backend
	chain.store.Init(params)           // init backend

	// genesis block not in chain, add it to chain, together with its miner tx
	// make sure genesis is in the store
	if !chain.Block_Exists(globals.Config.Genesis_Block_Hash) {
		logger.Debugf("Genesis block not in store, add it now")
		bl := Generate_Genesis_Block()
		chain.Store_BL(&bl)
		chain.Store_TOP_ID(globals.Config.Genesis_Block_Hash)
		// store height mapping, genesis block is at id
		chain.Store_BL_ID_at_Height(0, globals.Config.Genesis_Block_Hash)
	}

	// load the chain from the disk
	chain.Initialise_Chain_From_DB()

	// init mempool
	chain.Mempool, err = Init_Mempool(params)

	_ = err

	atomic.AddUint32(&globals.Subsystem_Active, 1) // increment subsystem

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

	logger.Infof("Chain Top Block %x Height %d", chain.Top_ID, chain.Height)

}

// before shutdown , make sure p2p is confirmed stopped
func (chain *Blockchain) Shutdown() {

	chain.Mempool.Shutdown() // shutdown mempool first

	chain.Lock() // take the lock as chain is no longer in unsafe mode
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

/* this function will extend the chain and increase the height,
   this function trigger checks  for the block and transactions for validity, recursively
   this is the only function which change chain height and top id
*/

func (chain *Blockchain) Chain_Add(bl *Block) (result bool) {

	chain.Lock()
	defer chain.Unlock()

	result = false

	block_hash := bl.GetHash()
	// make sure that the  block refers to some where in the chain
	// and also make sure block is not the  genesis block

	if block_hash == globals.Config.Genesis_Block_Hash {
		logger.Debugf("Genesis block already in chain skipping it")
		return
	}

	// check if block already exist skip it

	if chain.Block_Exists(block_hash) {
		logger.Debugf("block already in chain skipping it %x", block_hash)
		return
	}

	if chain.Height > 16996 {
		//	os.Exit(0)
	}

	if chain.Height > 17000 {
		//os.Exit(0)
	}

	if chain.Height > 17010 {
		//	os.Exit(0)
	}

	if chain.Height > 16996000 {
		os.Exit(0)
	}

	// make sure prev_hash refers to some point in our our chain
	// there is an edge case, where we know child but still donot know parent
	// this might be some some corrupted miner or initial sync
	if !chain.Block_Exists(bl.Prev_Hash) {
		// TODO we must queue this block for say 60 minutes, if parents donot appear it, discard it
		logger.Warnf("Prev_Hash  no where in the chain, skipping it till we get a parent %x", block_hash)
		return

	}

	PoW := bl.GetPoWHash()

	current_difficulty := chain.Get_Difficulty_At_Block(bl.Prev_Hash)
	logger.Debugf("Difficulty at height %d    is %d", chain.Load_Height_for_BL_ID(bl.Prev_Hash), current_difficulty)
	// check if the PoW is satisfied
	if !difficulty.CheckPowHash(PoW, current_difficulty) { // if invalid Pow, reject the bloc
		logger.Warnf("Block %x has invalid PoW, ignoring it", block_hash)
		return false
	}

	// check we need to  extend the chain or do a soft fork
	if bl.Prev_Hash == chain.Top_ID {
		// we need to extend the chain
		//log.Debugf("Extendin chain using block %x", block_hash )

		chain.Store_BL(bl)
		chain.Store_TOP_ID(block_hash) // make new block top block
		//chain.Add_Child(bl.Prev_Hash, block_hash) // add the new block as chil
		chain.Store_Block_Child(bl.Prev_Hash, block_hash)

		chain.Store_BL_ID_at_Height(chain.Height, block_hash)

		// lower the window, where top_id and  chain height are different
		chain.Height = chain.Height + 1 // increment height
		chain.Top_ID = block_hash       // set new top block id

		logger.Debugf("Chain extended using block %x new height %d", block_hash[:], chain.Height)

		// every 10 block print a line
		if chain.Height%20 == 0 {
			logger.Infof("Chain Height %d using block %x", chain.Height, block_hash[:])
		}

	} else { // a soft fork is in progress
		logger.Debugf("Soft Fork is in progress block due to %x", block_hash)
		chain.Chain_Add_And_Reorganise(bl)
	}

	//	fmt.Printf("We should add the block to DB, reorganise if required\n")
	return false
}

/* the block we have is NOT at the top, it either belongs to an altchain or is an alternative */
func (chain *Blockchain) Chain_Add_And_Reorganise(bl *Block) (result bool) {
	block_hash := bl.GetHash()

	// check whether the parent already has a child
	parent_has_child := chain.Does_Block_Have_Child(bl.Prev_Hash)
	// first lets add ourselves to the chain
	chain.Store_BL(bl)
	if !parent_has_child {
		chain.Store_Block_Child(bl.Prev_Hash, block_hash)
		logger.Infof("Adding alternative block %x to alt chain top\n", block_hash)
	} else {
		logger.Infof("Adding alternative block %x\n", block_hash)

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

// NOTE: below algorithm is the core and and is used to network consensus
// the best chain found using the following algorithm
// cryptonote protocol algo is below
// compare cdiff, chain with higher diff wins, if diff is same, no reorg, this cause frequent splits

// new algo is this
// compare cdiff, chain with higher diff wins, if diff is same, go below
// compare time stamp, block with lower timestamp wins (since it has probable spread more than other blocks)
// if timestamps are same, block with lower block has (No PoW involved) wins
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
	if found {
		// reorganise chain at this block
		children := chain.Load_Block_Children(block_hash)
		if len(children) < 2 {
			panic(fmt.Sprintf("Children disappeared for block %x", block_hash))
		}
		main_chain := chain.Load_Block_Child(block_hash)
		_ = main_chain

		// choose the best chain and make it parent
		for i := range children {
			top_hash := chain.Get_Top_Block(children[i])
			top_cdiff := chain.Load_Block_Cumulative_Difficulty(top_hash)
			timestamp := chain.Load_Block_Timestamp(children[i])
			children_data = append(children_data, chain_data{hash: children[i], cdifficulty: top_cdiff, foundat: timestamp})
		}

		sort.Sort(children_data)
		logger.Infof("Choosing best chain\n")
		for i := range children {
			fmt.Printf("%d %+v\n", i, children_data[i])
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

			// invalidate all transactionw contained within old main chain
			// validate all transactions in new main chain
			logger.Debugf("Invalidating all transactions with old main chain")

			logger.Debugf("Validating all transactions with old alt chain")

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

// make sure the block is valid before we even attempt to add it
func (chain *Blockchain) Is_Block_Valid(height uint64, bl *Block) bool {

	return true
}

/*

//  Parent's list is appended to add child
func (chain *Blockchain) Add_Child( Parent_Hash, Child_Hash crypto.Hash ){

fmt.Printf("caller %s\n", CallerName())
// child list is only appended and never truncated
// fetch old list
children_list := chain.Load_Chain(Parent_Hash)

if len(children_list) % 32 != 0 {
	log.Fatalf("Database corruption has occurred at this hash %x", Parent_Hash[:])
}


if len(children_list) == 0 {
	    chain.Store_Chain(Parent_Hash, Child_Hash[:])
	   log.Debugf("%x is a child of %x", Child_Hash, Parent_Hash)
	}else{ // we need to sort the children based on Pow
		panic("Chain need reorganisation, Sorting on PoW NOT implemented")

	}


return
}

*/

/* add a transaction to chain,, we are currently not verifying the TX,
its a BUG  and disaster, implement it ASAP
*/
func (chain *Blockchain) Add_TX(tx *Transaction) {

	chain.Store_TX(tx)

}

/* this will only give you access to transactions which have been mined
 */
func (chain *Blockchain) Get_TX(hash crypto.Hash) (*Transaction, error) {
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
		logger.Warnf("No Block at Height %d  , chain height %d\n", Height, chain.Get_Height())
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
	// traverse chain from the block referenced, to max 30 blocks ot till genesis block is researched
	for i := 0; i < config.DIFFICULTY_BLOCKS_COUNT_V2; i++ {
		if current_block_id == globals.Config.Genesis_Block_Hash || current_block_id == zero_block {
			rlog.Tracef(2, "Reached genesis block for difficulty calculation %x", block_id[:])
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
	panic("We can never reach this point")
	return block_id // we will never reach here
}

// verifies whether we are lagging
// return true if we need resync
// returns false if we are good
func (chain *Blockchain) IsLagging(peer_cdifficulty, peer_height uint64, peer_top_id crypto.Hash) bool {
	top_id := chain.Get_Top_ID()
	cdifficulty := chain.Load_Block_Cumulative_Difficulty(top_id)
	height := chain.Load_Height_for_BL_ID(top_id) + 1
	rlog.Tracef(3, "P_cdiff %d cdiff %d , P_BH %d BH %d, p_top %x top %x",
		peer_cdifficulty, cdifficulty,
		peer_height, height,
		peer_top_id, top_id)

	if peer_cdifficulty > cdifficulty{
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
func (chain *Blockchain)  Expand_Transaction_v2 (tx *Transaction){

	if tx.Version != 2 {
		panic("TX not version 2")
	}

	//if rctsignature is null

	// fill up the message first
	tx.RctSignature.Message = ringct.Key(tx.GetPrefixHash())


	// fill up the key images
	for i := 0; i < len(tx.Vin);i++{
		tx.RctSignature.MlsagSigs[i].II[0]= ringct.Key(tx.Vin[i].(Txin_to_key).K_image)
	}

	// now we need to fill up the mixring ctkey

}


// this function count all the vouts of the block,
// this function exists here because  only the chain knws the tx
//
func (chain *Blockchain) Block_Count_Vout(block_hash crypto.Hash) (count uint64){
	count = 1 // miner tx is always present

	bl, err := chain.Load_BL_FROM_ID(block_hash)

	if err != nil {
		panic(fmt.Errorf("Cannot load  block for %x err %s", block_hash, err))
	}

	for i := 0 ; i < len(bl.Tx_hashes);i++{ // load all tx one by one
		tx, err := chain.Load_TX_FROM_ID(bl.Tx_hashes[i])
		if err != nil{
			panic(fmt.Errorf("Cannot load  tx for %x err %s", bl.Tx_hashes[i], err))
		}

		// tx has been loaded, now lets get the vout
		 vout_count := uint64(len(tx.Vout))
		 count +=  vout_count
	}
	return count
}
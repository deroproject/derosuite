package blockchain

import "fmt"
import "encoding/binary"

//import log "github.com/sirupsen/logrus"

import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/globals"



/* this file implements the only interface which translates comands  to/from blockchain to storage layer *
 *
 *
 */

var TOP_ID = []byte("TOP_ID")  // stores current TOP, only stores single value
var TX_ID = []byte("TX")       // stores transactions
var BLOCK_ID = []byte("BLOCK") // stores blocks
var CHAIN = []byte("CHAIN")    // this stores the actual chain, parents keeps child list, starts from genesis block

var HEIGHT_TO_BLOCK_ID = []byte("HEIGHT_TO_BLOCK_ID") // stores block height to block id mapping
var BLOCK_ID_TO_HEIGHT = []byte("BLOCK_ID_TO_HEIGHT") // stores block id to height mapping

var BLOCK_ID_POW = []byte("BLOCK_ID_POW") // stores block_id to pow, this is slow to calculate, do we even need to store it

// all orphan ids are stored here, it id exists its orphan,
// once a block is adden to orphan list it cannot be removed, since we have a better block
var ORPHAN = []byte("ORPHAN")

var ORPHAN_HEIGHT = []byte("ORPHAN_HEIGHT") // height wise orphans are store here

var OO_ID = []byte{0x4} // mapping of incremental numbers to respective transaction ID

var ALTERNATIVE_BLOCKS_ID = []byte{0x5} // each block contains lists of alternative child blocks

var BLOCKCHAIN_UNIVERSE = []byte("BLOCKCHAIN_UNIVERSE") // all block chain data is store in this BLOCKCHAIN_UNIVERSE

// there are only 3 galaxies
var GALAXY_BLOCK = []byte("BLOCK")
var GALAXY_TRANSACTION = []byte("TRANSACTION")
var GALAXY_KEYIMAGE = []byte("KEYIMAGE")

//2 galaxies store inverse mapping
var GALAXY_HEIGHT = []byte("HEIGHT")             // height to block id mapping
var GALAXY_OUTPUT_INDEX = []byte("OUTPUT_INDEX") // incremental index over  output index

// the unique TXID or block ID becomes the solar system , which is common and saves lot of space

// individual attributes becomes the planets
// individual attributes should be max  1 or 2 chars long, as they will be repeated millions of times and storing a static string millions of times shows foolishness
var PLANET_BLOB = []byte("BLOB")     //it shows serialised block
var PLANET_HEIGHT = []byte("HEIGHT") // contains height
var PLANET_PARENT = []byte("PARENT") // parent of block
var PLANET_SIZE = []byte("SIZE")                      // sum of block + all txs
var PLANET_ALREADY_GENERATED_COINS = []byte("CCOINS") // all coins generated till this block
var PLANET_OUTPUT_INDEX = []byte("OUTPUT_INDEX") // tx outputs indexing starts from here for this block
var PLANET_CUMULATIVE_DIFFICULTY = []byte("CDIFFICULTY")
var PLANET_CHILD = []byte("CHILD")

//var PLANET_ORPHAN = []byte("ORPHAN")
var PLANET_TIMESTAMP = []byte("TIMESTAMP")

// this ill only be present if more tahn 1 child exists
var PLANET_CHILDREN = []byte("CHILREN") // children list excludes the main child, so its a multiple of 32

// the TX has the following attributes
var PLANET_TX_BLOB = []byte("BLOB")                 // contains serialised  TX , this attribute is also found in BLOCK where
var PLANET_TX_MINED_IN_BLOCK = []byte("MINERBLOCK") // which block mined this tx
var PLANET_TX_SIZE = []byte("SIZE")

// the universe concept is there, as we bring in smart contracts, we will give each of them a universe to play within
// while communicating with external universe

/*
func (chain *Blockchain) Store_Main_Chain(parent_id  crypto.Hash, child_id crypto.Hash){
   err := chain.store.StoreObject(BLOCKCHAIN_UNIVERSE,GALAXY_BLOCK,parent_id[:],PLANET_CHILD, child_id[:] )
     _ = err
 }

func (chain *Blockchain) Load_Main_Chain(parent_id  crypto.Hash) (child_id crypto.Hash ){
	var err error
    // store OO to TXID automatically
     object_data,err = chain.store.LoadObject(BLOCKCHAIN_UNIVERSE,GALAXY_BLOCK,parent_id[:],PLANET_CHILD )

    if err != nil {
    	return child_id,err
    }

    if len(object_data) == 0 {
    	return child_id, fmt.Errorf("No Block at such Height %d", Height)
    }

    if len(object_data) != 32 {
    	panic("Database corruption, invalid block hash ")
    }

    copy(child_id[:],object_data[:32])

_ = err

  return child_id

}

*/

// check whether the block has a child
func (chain *Blockchain) Does_Block_Have_Child(block_id crypto.Hash) bool {
	var err error
	object_data, err := chain.store.LoadObject(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, block_id[:], PLANET_CHILD)

	if err != nil || len(object_data) == 0 {
		return false
	}
	if len(object_data) != 32 {
		panic("Database corruption, invalid block hash ")
	}
	return true
}

// load the main child
func (chain *Blockchain) Load_Block_Child(parent_id crypto.Hash) (child_id crypto.Hash) {
	if !chain.Does_Block_Have_Child(parent_id) {
		panic("Block does not have a child")
	}
	object_data, _ := chain.store.LoadObject(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, parent_id[:], PLANET_CHILD)

	copy(child_id[:], object_data)
	return
}

// changes  or set child block of a parent
// there can be only 1 child, rest all are alternatives and stored as
func (chain *Blockchain) Store_Block_Child(parent_id crypto.Hash, child_id crypto.Hash) {
	err := chain.store.StoreObject(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, parent_id[:], PLANET_CHILD, child_id[:])

	// load block children
	_ = err
}

// while store children
func (chain *Blockchain) Store_Block_Children(parent_id crypto.Hash, children []crypto.Hash, exclude_child crypto.Hash) {
	var children_bytes []byte
	for i := range children {
		if children[i] != exclude_child { // exclude main child
			children_bytes = append(children_bytes, children[i][:]...)
		}
	}
	err := chain.store.StoreObject(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, parent_id[:], PLANET_CHILDREN, children_bytes)
	_ = err
}

func (chain *Blockchain) Load_Block_Children(parent_id crypto.Hash) (children []crypto.Hash) {
	var child_hash crypto.Hash
	if !chain.Does_Block_Have_Child(parent_id) { // block doesnot have a child, so it cannot have children
		return
	}
	// we are here means parent does have child
	children = append(children, chain.Load_Block_Child(parent_id))

	// check for children
	children_bytes, _ := chain.store.LoadObject(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, parent_id[:], PLANET_CHILDREN)

	if len(children_bytes)%32 != 0 {
		panic(fmt.Sprintf("parent does not have child hash in multiples of 32, block_hash %x", parent_id))
	}
	for i := 0; i < len(children_bytes); i = i + 32 {
		copy(child_hash[:], children_bytes[i:i+32])
		children = append(children, child_hash)
	}
	return children
}

// store a tx
// this only occurs when a tx has been mined
func (chain *Blockchain) Store_TX(tx *Transaction) {
	hash := tx.GetHash()
	serialized := tx.Serialize()
	err := chain.store.StoreObject(BLOCKCHAIN_UNIVERSE, GALAXY_TRANSACTION, hash[:], PLANET_TX_BLOB, serialized)
	// store size of tx
	chain.store.StoreUint64(BLOCKCHAIN_UNIVERSE, GALAXY_TRANSACTION, hash[:], PLANET_TX_SIZE, uint64(len(serialized)))

	_ = err
}

func (chain *Blockchain) Store_TX_Miner(txhash crypto.Hash, block_id crypto.Hash) {
	// store block id  which mined this tx
	err := chain.store.StoreObject(BLOCKCHAIN_UNIVERSE, GALAXY_TRANSACTION, txhash[:], PLANET_TX_MINED_IN_BLOCK, block_id[:])
	_ = err
}

func (chain *Blockchain) Load_TX_Size(txhash crypto.Hash) uint64 {
	// store block id  which mined this tx
	size, err := chain.store.LoadUint64(BLOCKCHAIN_UNIVERSE, GALAXY_TRANSACTION, txhash[:], PLANET_TX_SIZE)

	if err != nil {
		logger.Warnf("Size not stored for tx %x", txhash)
	}
	return size
}

// BUG we should be able to delete any arbitrary key
// since a tx mined by one block, can be back in pool after chain reorganises

// TODO the miner tx should be extracted ands stored from somewhere else
// NOTE: before storing a block, its transactions must be stored
func (chain *Blockchain) Store_BL(bl *Block) {

	// store block height BHID automatically

	hash := bl.GetHash()

	// we should deserialize the block here
	serialized_bytes := bl.Serialize() // we are storing the miner transactions within
	err := chain.store.StoreObject(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, hash[:], PLANET_BLOB, serialized_bytes)

	// get height of parent, add 1 and store it
	height := uint64(0)
	if hash != globals.Config.Genesis_Block_Hash { // genesis block has no parent
		height = chain.Load_Height_for_BL_ID(bl.Prev_Hash)
		height++
	}

	// store new height
	chain.store.StoreUint64(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, hash[:], PLANET_HEIGHT, height)

	// store timestamp
	chain.store.StoreUint64(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, hash[:], PLANET_TIMESTAMP, bl.Timestamp)

	// store parent
	chain.store.StoreObject(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, hash[:], PLANET_PARENT, bl.Prev_Hash[:])

	// calculate cumulative difficulty at last block
	difficulty_of_current_block := uint64(0)
	cumulative_difficulty := uint64(0)
	if hash != globals.Config.Genesis_Block_Hash { // genesis block has no parent
		cumulative_difficulty = chain.Load_Block_Cumulative_Difficulty(bl.Prev_Hash)
		difficulty_of_current_block = chain.Get_Difficulty_At_Block(bl.Prev_Hash)
	} else {
		cumulative_difficulty = 1 // genesis block cumulative difficulty is 1
	}

	total_difficulty := cumulative_difficulty + difficulty_of_current_block
	chain.store.StoreUint64(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, hash[:], PLANET_CUMULATIVE_DIFFICULTY, total_difficulty)

	// total size of block = size of block + size of all transactions in block ( excludind miner tx)
	size_of_block := uint64(len(serialized_bytes))
	for i := 0; i < len(bl.Tx_hashes); i++ {
		size_of_tx := chain.Load_TX_Size(bl.Tx_hashes[i])
		size_of_block += size_of_tx
	}
	chain.store.StoreUint64(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, hash[:], PLANET_SIZE, size_of_block)

        
        // calculated position of vouts in global index
        /* TODO below code has been disabled and should be enabled for extensive testing
        index_pos := uint64(0)
        if hash != globals.Config.Genesis_Block_Hash {
           // load index pos from last block + add count of vouts from last block
            index_pos = chain.Get_Block_Output_Index(bl.Prev_Hash)
            vout_count_prev_block := chain.Block_Count_Vout(bl.Prev_Hash)
            index_pos += vout_count_prev_block
        }
        chain.store.StoreUint64(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, hash[:], PLANET_OUTPUT_INDEX, index_pos)
        logger.Debugf("height %d   output index %d\n",height, index_pos)
        */
        
	// TODO calculate total coins emitted till this block

	// also extract and store the miner tx separetly, fr direct querying purpose
	chain.Store_TX(&bl.Miner_tx)

	_ = err
}

func (chain *Blockchain) Load_TX_FROM_ID(hash [32]byte) (*Transaction, error) {
	var tx Transaction
	tx_data, err := chain.store.LoadObject(BLOCKCHAIN_UNIVERSE, GALAXY_TRANSACTION, hash[:], PLANET_TX_BLOB)

	if err != nil {
		return nil, err
	}

	// we should deserialize the block here
	err = tx.DeserializeHeader(tx_data)

	if err != nil {
		logger.Printf("fError deserialiing tx, block id %x len(data) %d data %x\n", hash[:], len(tx_data), tx_data)
		return nil, err
	}
	return &tx, nil

}

func (chain *Blockchain) Load_TX_FROM_OO(Offset uint64) {

}

func (chain *Blockchain) Load_BL_FROM_ID(hash [32]byte) (*Block, error) {
	var bl Block
	block_data, err := chain.store.LoadObject(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, hash[:], PLANET_BLOB)

	if err != nil {
		return nil, err
	}

	if len(block_data) == 0 {
		return nil, fmt.Errorf("Block not found in DB")
	}

	// we should deserialize the block here
	err = bl.Deserialize(block_data)

	if err != nil {
		logger.Warnf("fError deserialiing block, block id %x len(data) %d data %x\n", hash[:], len(block_data), block_data)
		return nil, err
	}
	return &bl, nil
}

// this will give you a block id at a specific height
func (chain *Blockchain) Load_BL_ID_at_Height(Height uint64) (hash crypto.Hash, err error) {
	object_data, err := chain.store.LoadObject(BLOCKCHAIN_UNIVERSE, GALAXY_HEIGHT, PLANET_HEIGHT, itob(Height))

	if err != nil {
		return hash, err
	}

	if len(object_data) == 0 {
		return hash, fmt.Errorf("No Block at such Height %d", Height)
	}

	if len(object_data) != 32 {
		panic("Database corruption, invalid block hash ")
	}
	copy(hash[:], object_data[:32])
	return hash, nil

}

// this will give you a block id at a specific height
func (chain *Blockchain) Store_BL_ID_at_Height(Height uint64, hash crypto.Hash) {
	// store height to block id mapping
	chain.store.StoreObject(BLOCKCHAIN_UNIVERSE, GALAXY_HEIGHT, PLANET_HEIGHT, itob(Height), hash[:])

}

func (chain *Blockchain) Load_Height_for_BL_ID(hash crypto.Hash) (Height uint64) {
	object_data, err := chain.store.LoadObject(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, hash[:], PLANET_HEIGHT)

	if err != nil {
		logger.Warnf("Error while querying height for block %x\n", hash)
		return
	}

	if len(object_data) == 0 {
		//return hash, fmt.Errorf("No Height for block %x", hash[:])
		return
	}

	if len(object_data) != 8 {
		panic("Database corruption, invalid block hash ")
	}

	Height = binary.BigEndian.Uint64(object_data)

	return Height

}

func (chain *Blockchain) Load_Block_Timestamp(hash crypto.Hash) uint64 {
	timestamp, err := chain.store.LoadUint64(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, hash[:], PLANET_TIMESTAMP)

	if err != nil {
		logger.Fatalf("Error while querying timestamp for block %x\n", hash)
		panic("Error while querying timestamp for block")
	}

	return timestamp
}

func (chain *Blockchain) Load_Block_Cumulative_Difficulty(hash crypto.Hash) uint64 {
	cdifficulty, err := chain.store.LoadUint64(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, hash[:], PLANET_CUMULATIVE_DIFFICULTY)

	if err != nil {
		logger.Panicf("Error while querying cumulative difficulty for block %x\n", hash)
		
	}

	return cdifficulty
}

func (chain *Blockchain) Load_Block_Parent_ID(hash crypto.Hash) crypto.Hash {
	var parent_id crypto.Hash
	object_data, err := chain.store.LoadObject(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, hash[:], PLANET_PARENT)

	if err != nil || len(object_data) != 32 {
		logger.Panicf("Error while querying parent id for block %x\n", hash)
		}
	copy(parent_id[:], object_data)

	return parent_id
}

// store current top id
func (chain *Blockchain) Store_TOP_ID(hash crypto.Hash) {
	chain.store.StoreObject(BLOCKCHAIN_UNIVERSE, TOP_ID, TOP_ID, TOP_ID, hash[:])
}

func (chain *Blockchain) Load_TOP_ID() (hash crypto.Hash) {
	object_data, err := chain.store.LoadObject(BLOCKCHAIN_UNIVERSE, TOP_ID, TOP_ID, TOP_ID)

	if err != nil {
		panic("Backend failure")
	}

	if len(object_data) == 0 {
		panic(fmt.Errorf("most probably Database corruption, No TOP_ID stored "))
	}

	if len(object_data) != 32 {
		panic("Database corruption, invalid block hash ")
	}
	copy(hash[:], object_data[:32])
	return hash
}

// itob returns an 8-byte big endian representation of v.
func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}



// get the position from where indexing must start for this block
// indexing mean vout based index
// cryptonote works by giving each vout a unique index
func (chain *Blockchain)Get_Block_Output_Index(block_id crypto.Hash) uint64 {
    if block_id == globals.Config.Genesis_Block_Hash { // genesis block has no output index
		return 0 ; // counting starts from zero
	}
    
    index, err := chain.store.LoadUint64(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, block_id[:], PLANET_OUTPUT_INDEX)
    if err != nil {
        // TODO  this panic must be enabled to catch some bugs
        panic(fmt.Errorf("Cannot load output index for %x err %s", block_id, err))
        return 0
    }
    
    return index
}

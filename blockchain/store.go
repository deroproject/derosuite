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

import "fmt"
import "encoding/binary"

//import log "github.com/sirupsen/logrus"

import "github.com/deroproject/derosuite/block"
import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/transaction"

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
var PLANET_BLOB = []byte("BLOB")                      //it shows serialised block
var PLANET_HEIGHT = []byte("HEIGHT")                  // contains height
var PLANET_PARENT = []byte("PARENT")                  // parent of block
var PLANET_SIZE = []byte("SIZE")                      // sum of block + all txs
var PLANET_ALREADY_GENERATED_COINS = []byte("CCOINS") // all coins generated till this block
var PLANET_OUTPUT_INDEX = []byte("OUTPUT_INDEX")      // tx outputs indexing starts from here for this block
var PLANET_CUMULATIVE_DIFFICULTY = []byte("CDIFFICULTY")
var PLANET_CHILD = []byte("CHILD")
var PLANET_BASEREWARD = []byte("BASEREWARD") // base reward of a block
var PLANET_TIMESTAMP = []byte("TIMESTAMP")

// this ill only be present if more tahn 1 child exists
var PLANET_CHILDREN = []byte("CHILREN") // children list excludes the main child, so its a multiple of 32

// the TX has the following attributes
var PLANET_TX_BLOB = []byte("BLOB")                 // contains serialised  TX , this attribute is also found in BLOCK where
var PLANET_TX_MINED_IN_BLOCK = []byte("MINERBLOCK") // which block mined this tx, height
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
		panic(fmt.Sprintf("parent does not have child hash in multiples of 32, block_hash %s", parent_id))
	}
	for i := 0; i < len(children_bytes); i = i + 32 {
		copy(child_hash[:], children_bytes[i:i+32])
		children = append(children, child_hash)
	}
	return children
}

// store a tx
// this only occurs when a tx has been mined
// stores a height to show at what height it has been mined
func (chain *Blockchain) Store_TX(tx *transaction.Transaction, Height uint64) {
	hash := tx.GetHash()
	serialized := tx.Serialize()
	err := chain.store.StoreObject(BLOCKCHAIN_UNIVERSE, GALAXY_TRANSACTION, hash[:], PLANET_TX_BLOB, serialized)
	// store size of tx
	chain.store.StoreUint64(BLOCKCHAIN_UNIVERSE, GALAXY_TRANSACTION, hash[:], PLANET_TX_SIZE, uint64(len(serialized)))

	chain.store.StoreUint64(BLOCKCHAIN_UNIVERSE, GALAXY_TRANSACTION, hash[:], PLANET_TX_MINED_IN_BLOCK, Height)

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
		logger.Warnf("Size not stored for tx %s", txhash)
	}
	return size
}

// load height at which a specific tx was mined
func (chain *Blockchain) Load_TX_Height(txhash crypto.Hash) uint64 {
	height, err := chain.store.LoadUint64(BLOCKCHAIN_UNIVERSE, GALAXY_TRANSACTION, txhash[:], PLANET_TX_MINED_IN_BLOCK)
	if err != nil {
		logger.Warnf("Error while querying height for tx %s\n", txhash)
	}
	return height
}

// BUG we should be able to delete any arbitrary key
// since a tx mined by one block, can be back in pool after chain reorganises

// TODO the miner tx should be extracted ands stored from somewhere else
// NOTE: before storing a block, its transactions must be stored
func (chain *Blockchain) Store_BL(bl *block.Block) {

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

	// total size of block = size of miner_tx + size of all transactions in block ( excludind miner tx)
	size_of_block := uint64(len(bl.Miner_tx.Serialize()))
	for i := 0; i < len(bl.Tx_hashes); i++ {
		size_of_tx := chain.Load_TX_Size(bl.Tx_hashes[i])
		size_of_block += size_of_tx
	}
	chain.store.StoreUint64(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, hash[:], PLANET_SIZE, size_of_block)

	// calculated position of vouts in global indexs
	index_pos := uint64(0)
	if hash != globals.Config.Genesis_Block_Hash {
		// load index pos from last block + add count of vouts from last block
		index_pos = chain.Get_Block_Output_Index(bl.Prev_Hash)
		vout_count_prev_block := chain.Block_Count_Vout(bl.Prev_Hash)
		index_pos += vout_count_prev_block
	}
	chain.store.StoreUint64(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, hash[:], PLANET_OUTPUT_INDEX, index_pos)
	//logger.Debugf("height %d   output index %d",height, index_pos)

	total_fees := uint64(0)
	for i := 0; i < len(bl.Tx_hashes); i++ {
		tx, _ := chain.Load_TX_FROM_ID(bl.Tx_hashes[i])
		total_fees += tx.RctSignature.Get_TX_Fee()
	}

	total_reward := bl.Miner_tx.Vout[0].Amount
	base_reward := total_reward - total_fees
	chain.store.StoreUint64(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, hash[:], PLANET_BASEREWARD, base_reward)

	already_generated_coins := uint64(0)
	if hash != globals.Config.Genesis_Block_Hash { // genesis block has no parent
		already_generated_coins = chain.Load_Already_Generated_Coins_for_BL_ID(bl.Prev_Hash)
	} else {
		base_reward = 1000000000000 // trigger the bug to fix coin calculation, see comments in emission
	}
	already_generated_coins += base_reward
	chain.store.StoreUint64(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, hash[:], PLANET_ALREADY_GENERATED_COINS, already_generated_coins)

	// also extract and store the miner tx separetly, fr direct querying purpose
	chain.Store_TX(&bl.Miner_tx, height)

	_ = err
}

func (chain *Blockchain) Load_TX_FROM_ID(hash [32]byte) (*transaction.Transaction, error) {
	var tx transaction.Transaction
	tx_data, err := chain.store.LoadObject(BLOCKCHAIN_UNIVERSE, GALAXY_TRANSACTION, hash[:], PLANET_TX_BLOB)

	if err != nil {
		return nil, err
	}

	// we should deserialize the block here
	err = tx.DeserializeHeader(tx_data)

	if err != nil {
		logger.Printf("fError deserialiing tx, block id %s len(data) %d data %x\n", hash[:], len(tx_data), tx_data)
		return nil, err
	}
	return &tx, nil

}

func (chain *Blockchain) Load_BL_FROM_ID(hash [32]byte) (*block.Block, error) {
	var bl block.Block
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
		logger.Warnf("fError deserialiing block, block id %s len(data) %d data %x\n", hash[:], len(block_data), block_data)
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

	if hash == ZERO_HASH { // handle special case for  genesis
		return 0
	}

	object_data, err := chain.store.LoadObject(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, hash[:], PLANET_HEIGHT)

	if err != nil {
		logger.Warnf("Error while querying height for block %s\n", hash)
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
		logger.Warnf("Error while querying timestamp for block %s\n", hash)

	}

	return timestamp
}

func (chain *Blockchain) Load_Block_Cumulative_Difficulty(hash crypto.Hash) uint64 {
	cdifficulty, err := chain.store.LoadUint64(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, hash[:], PLANET_CUMULATIVE_DIFFICULTY)

	if err != nil {
		logger.Panicf("Error while querying cumulative difficulty for block %s\n", hash)

	}

	return cdifficulty
}

func (chain *Blockchain) Load_Block_Reward(hash crypto.Hash) uint64 {
	timestamp, err := chain.store.LoadUint64(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, hash[:], PLANET_BASEREWARD)
	if err != nil {
		logger.Warnf("Error while querying base_reward for block %s\n", hash)
	}

	return timestamp
}

func (chain *Blockchain) Load_Already_Generated_Coins_for_BL_ID(hash crypto.Hash) uint64 {

	if hash == ZERO_HASH {
		return 0
	}
	timestamp, err := chain.store.LoadUint64(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, hash[:], PLANET_ALREADY_GENERATED_COINS)
	if err != nil {
		logger.Warnf("Error while querying alreadt generated coins for block %s\n", hash)

	}

	return timestamp
}

func (chain *Blockchain) Load_Block_Size(hash crypto.Hash) uint64 {
	timestamp, err := chain.store.LoadUint64(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, hash[:], PLANET_SIZE)
	if err != nil {
		logger.Warnf("Error while querying size for block %s\n", hash)
	}

	return timestamp
}

func (chain *Blockchain) Load_Block_Parent_ID(hash crypto.Hash) crypto.Hash {
	var parent_id crypto.Hash
	object_data, err := chain.store.LoadObject(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, hash[:], PLANET_PARENT)

	if err != nil || len(object_data) != 32 {
		logger.Warnf("Error while querying parent id for block %s\n", hash)
	}
	copy(parent_id[:], object_data)

	return parent_id
}

// store current top id
func (chain *Blockchain) Store_TOP_ID(hash crypto.Hash) {
	chain.store.StoreObject(BLOCKCHAIN_UNIVERSE, TOP_ID, TOP_ID, TOP_ID, hash[:])
}

// crash if something is not correct
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
func (chain *Blockchain) Get_Block_Output_Index(block_id crypto.Hash) uint64 {
	if block_id == globals.Config.Genesis_Block_Hash { // genesis block has no output index
		return 0 // counting starts from zero
	}

	index, err := chain.store.LoadUint64(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, block_id[:], PLANET_OUTPUT_INDEX)
	if err != nil {
		// TODO  this panic must be enabled to catch some bugs
		logger.Warnf("Cannot load output index for %s err %s", block_id, err)
		return 0
	}

	return index
}

// store key image to its own galaxy
// a keyimage stored with value 1  == it has been consumed
// a keyimage stored with value 0  == it has not been consumed
// a key image not found in store == it has NOT been consumed
func (chain *Blockchain) Store_KeyImage(hash crypto.Hash, mark bool) {
	store_value := byte(0)
	if mark {
		store_value = byte(1)
	}
	chain.store.StoreObject(BLOCKCHAIN_UNIVERSE, GALAXY_KEYIMAGE, GALAXY_KEYIMAGE, hash[:], []byte{store_value})
}

// read a key image, whether it's stored with value 1
// a keyimage stored with value 1  == it has been consumed
// a keyimage stored with value 0  == it has not been consumed
// a key image not found in store == it has NOT been consumed
func (chain *Blockchain) Read_KeyImage_Status(hash crypto.Hash) bool {
	object_data, err := chain.store.LoadObject(BLOCKCHAIN_UNIVERSE, GALAXY_KEYIMAGE, GALAXY_KEYIMAGE, hash[:])

	if err != nil {
		return false
	}

	if len(object_data) == 0 {
		return false
	}

	if len(object_data) != 1 {
		panic(fmt.Errorf("probably Database corruption, Wrong data stored in keyimage, expected size 1, actual size %d", len(object_data)))
	}

	if object_data[0] == 1 {
		return true
	}

	// anything other than value 1 is considered wrong keyimage

	return false
}

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
import "time"
import "sync"
import "strings"
import "math/rand"
import "runtime/debug"

//import "encoding/binary"
import "golang.org/x/time/rate"

//import "encoding/hex"

import "github.com/romana/rlog"

// this file creates the blobs which can be used to mine new blocks

import "github.com/deroproject/derosuite/block"
import "github.com/deroproject/derosuite/config"
import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/address"

//import "github.com/deroproject/derosuite/emission"
import "github.com/deroproject/derosuite/transaction"

//NOTE: this function is quite naughty due to chicken and egg problem
// we cannot calculate reward until we known blocksize ( exactly upto byte size )
// also note that we cannot calculate blocksize  until we know reward
// how we do it
// reward field is a varint, ( can be 8 bytes )
// so max deviation can be 8 bytes, due to reward
// so we do a bruterforce till the reward is obtained but lets try to KISS
// the top hash over which to do mining now ( it should already be in the chain)
// this is work in progress
// TODO we need to rework fees algorithm, to improve its quality and lower fees
func (chain *Blockchain) Create_new_miner_block(miner_address address.Address) (cbl *block.Complete_Block, bl block.Block) {
	//chain.Lock()
	//defer chain.Unlock()

	cbl = &block.Complete_Block{}

	dbtx, err := chain.store.BeginTX(false)
	if err != nil {
		logger.Warnf(" Error opening readonly TX, err %s", err)
		return
	}

	defer dbtx.Rollback()

	// let give is a chance to mempool to do house keeping
	// we can skip house-keeping
        /*
	func() {
		if r := recover(); r != nil {
			logger.Warnf("Mempool House Keeping triggered panic height = %d", chain.Get_Height())
		}

		// give mempool an oppurtunity to verify alt-chain tx, but only if they are not mined
		// but only check for double spending
		chain.Mempool.HouseKeeping(uint64(chain.Get_Height()), func(tx *transaction.Transaction) bool {
			return chain.Verify_Transaction_NonCoinbase_DoubleSpend_Check(dbtx, tx)
		})
	}() */
	

	// first of lets find the tx fees collected by consuming txs from mempool
	tx_hash_list_sorted := chain.Mempool.Mempool_List_TX_SortedInfo() // hash of all tx expected to be included within this block , sorted by fees
	var tx_hash_list_included []crypto.Hash                           // these tx will be included ( due to  block size limit )

	// use best 3 tips
	tips := chain.SortTips(dbtx, chain.Load_TIPS_ATOMIC(dbtx))
	for i := range tips {
		if len(bl.Tips) >= 3 {
			break
		}
		bl.Tips = append(bl.Tips, tips[i])
	}
	height := chain.Calculate_Height_At_Tips(dbtx, bl.Tips) // we are 1 higher than previous highest tip

	sizeoftxs := uint64(0) // size of all non coinbase tx included within this block
	//fees_collected := uint64(0)

	rlog.Infof("Total tx in pool %d", len(tx_hash_list_sorted))

	reachable_key_images := chain.BuildReachabilityKeyImages(dbtx, &bl) // this requires only bl.Tips

	// select 10%  tx based on fees
	// select 90%  tx randomly
	// random selection helps us to easily reach 80 TPS
	i := 0
	for ; i < len(tx_hash_list_sorted); i++ {

		if (sizeoftxs + tx_hash_list_sorted[i].Size) > (10*config.CRYPTONOTE_MAX_BLOCK_SIZE)/100 { // limit block to max possible
			break
		}

		tx := chain.Mempool.Mempool_Get_TX(tx_hash_list_sorted[i].Hash)
		if tx != nil {
			// skip and delete any mempool tx
			if chain.Verify_Transaction_NonCoinbase_DoubleSpend_Check(dbtx, tx) == false {
				chain.Mempool.Mempool_Delete_TX(tx_hash_list_sorted[i].Hash)
				continue
			}

			failed := false
			for j := 0; j < len(tx.Vin); j++ {
				if _, ok := reachable_key_images[tx.Vin[j].(transaction.Txin_to_key).K_image]; ok {
					rlog.Warnf("TX already in history, but tx %s  is still in mempool HOW ?? skipping it", tx_hash_list_sorted[i].Hash)
					failed = true
					break
				}
			}
			if failed {
				continue
			}

			rlog.Tracef(1,"Adding Top  Sorted tx %s to Complete_Block current size %.2f KB max possible %.2f KB\n", tx_hash_list_sorted[i].Hash, float32(sizeoftxs+tx_hash_list_sorted[i].Size)/1024.0, float32(config.CRYPTONOTE_MAX_BLOCK_SIZE)/1024.0)
			sizeoftxs += tx_hash_list_sorted[i].Size
			cbl.Txs = append(cbl.Txs, tx)
			tx_hash_list_included = append(tx_hash_list_included, tx_hash_list_sorted[i].Hash)
		}
	}
	// any left over transactions, should be randomly selected
	tx_hash_list_sorted = tx_hash_list_sorted[i:]

	// do random shuffling, can we get away with len/2 random shuffling
	rand.Shuffle(len(tx_hash_list_sorted), func(i, j int) {
		tx_hash_list_sorted[i], tx_hash_list_sorted[j] = tx_hash_list_sorted[j], tx_hash_list_sorted[i]
	})

	// if we were crossing limit, transactions would be randomly selected
	// otherwise they will sorted by fees

	// now select as much as possible
	for i := range tx_hash_list_sorted {
		if (sizeoftxs + tx_hash_list_sorted[i].Size) > (config.CRYPTONOTE_MAX_BLOCK_SIZE) { // limit block to max possible
			break
		}

		tx := chain.Mempool.Mempool_Get_TX(tx_hash_list_sorted[i].Hash)
		if tx != nil {

			// skip and delete any mempool tx
			if chain.Verify_Transaction_NonCoinbase_DoubleSpend_Check(dbtx, tx) == false {
				chain.Mempool.Mempool_Delete_TX(tx_hash_list_sorted[i].Hash)
				continue
			}
			failed := false
			for j := 0; j < len(tx.Vin); j++ {
				if _, ok := reachable_key_images[tx.Vin[j].(transaction.Txin_to_key).K_image]; ok {
					rlog.Warnf("TX already in history, but tx %s  is still in mempool HOW ?? skipping it", tx_hash_list_sorted[i].Hash)
					failed = true
					break
				}
			}
			if failed {
				continue
			}

			rlog.Tracef(1,"Adding Random tx %s to Complete_Block current size %.2f KB max possible %.2f KB\n", tx_hash_list_sorted[i].Hash, float32(sizeoftxs+tx_hash_list_sorted[i].Size)/1024.0, float32(config.CRYPTONOTE_MAX_BLOCK_SIZE)/1024.0)
			sizeoftxs += tx_hash_list_sorted[i].Size
			cbl.Txs = append(cbl.Txs, tx)
			tx_hash_list_included = append(tx_hash_list_included, tx_hash_list_sorted[i].Hash)
		}
	}

	// collect tx list + their fees

	// now we have all major parts of block, assemble the block
	bl.Major_Version = uint64(chain.Get_Current_Version_at_Height(height))
	bl.Minor_Version = uint64(chain.Get_Ideal_Version_at_Height(height)) // This is used for hard fork voting,
	bl.Timestamp = uint64(uint64(time.Now().UTC().Unix()))
	bl.Miner_TX, err = Create_Miner_TX2(int64(bl.Major_Version), height, miner_address)
	if err != nil {
		logger.Warnf("Error while creating miner block, err %s", err)
	}

	//bl.Prev_Hash = top_hash
	bl.Nonce = rand.New(globals.NewCryptoRandSource()).Uint32() // nonce can be changed by the template header

	for i := range tx_hash_list_included {
		bl.Tx_hashes = append(bl.Tx_hashes, tx_hash_list_included[i])
	}
	cbl.Bl = &bl

	//logger.Infof("miner block %+v", bl)
	return
}

// returns a new block template ready for mining
// block template has the following format
// miner block header in hex  +
// miner tx in hex +
// 2 bytes ( inhex 4 bytes for number of tx )
// tx hashes that follow

func (chain *Blockchain) Create_new_block_template_mining(top_hash crypto.Hash, miner_address address.Address, reserve_space int) (bl block.Block, blockhashing_blob string, block_template_blob string, reserved_pos int) {

	rlog.Debugf("Mining block will give reward to %s", miner_address)
	_, bl = chain.Create_new_miner_block(miner_address)

	blockhashing_blob = fmt.Sprintf("%x", bl.GetBlockWork())

	// block template is all the parts of the block in dismantled form
	// first is the block header
	// then comes the miner tx
	// then comes all the tx headers
	block_template_blob = fmt.Sprintf("%x", bl.Serialize())

	// lets locate  extra nonce
	pos := strings.Index(blockhashing_blob, "0000000000000000000000000000000000000000000000000000000000000000")

	pos = pos / 2 // we searched in hexadecimal form but we need to give position in byte form
	reserved_pos = pos

	return
}

// rate limiter is deployed, in case RPC is exposed over internet
// someone should not be just giving fake inputs and delay chain syncing
var accept_limiter = rate.NewLimiter(1.0, 4) // 1 block per sec, burst of 4 blocks is okay

var accept_lock = sync.Mutex{}

var duplicate_check = map[crypto.Hash]bool{}
var duplicate_height_check = map[uint64]bool{}

// accept work given by us
// we should verify that the transaction list supplied back by the miner exists in the mempool
// otherwise the miner is trying to attack the network
func (chain *Blockchain) Accept_new_block(block_template []byte, blockhashing_blob []byte) (blid crypto.Hash,result bool, err error) {

	// check whether  we are in lowcpu mode, if yes reject the block
	if globals.Arguments["--lowcpuram"].(bool) {
		globals.Logger.Warnf("Mining is deactivated since daemon is running in low cpu mode, please check program options.")
		return blid,false, fmt.Errorf("Please decativate lowcpuram mode")
	}
	if globals.Arguments["--sync-node"].(bool) {
		globals.Logger.Warnf("Mining is deactivated since daemon is running with --sync-mode, please check program options.")
		return blid,false, fmt.Errorf("Please deactivate --sync-node option before mining")
	}

	accept_lock.Lock()
	defer accept_lock.Unlock()

	cbl := &block.Complete_Block{}
	bl := block.Block{}

	//logger.Infof("Incoming block for accepting %x", block_template)
	// safety so if anything wrong happens, verification fails
	defer func() {
		if r := recover(); r != nil {
			logger.Warnf("Recovered while accepting new block, Stack trace below ")
			logger.Warnf("Stack trace  \n%s", debug.Stack())
			err = fmt.Errorf("Error while parsing block")
		}
	}()

	err = bl.Deserialize(block_template)
	if err != nil {
		logger.Warnf("Error parsing submitted work block template err %s", err)
		return 
	}

	length_of_block_header := len(bl.Serialize())
	template_data := block_template[length_of_block_header:]

	if len(blockhashing_blob) >= 2{
	err = bl.CopyNonceFromBlockWork(blockhashing_blob)
	if err != nil {
		logger.Warnf("Submitted block has been rejected, since blockhashing_blob is invalid")
		return
	}
        }

	if len(template_data) != 0 {
		logger.Warnf("Extra bytes (%d) left over while accepting block from mining pool %x", len(template_data), template_data)
	}

	// if we reach here, everything looks ok
	// try to craft a complete block by grabbing entire tx from the mempool
	//logger.Debugf("Block parsed successfully")
	
	blid = bl.GetHash()
	
        // if a duplicate block is being sent, reject the block
	if _, ok := duplicate_check[bl.Miner_TX.GetHash()]; ok {
		logger.Warnf("Block %s rejected by chain due to duplicate work.", bl.GetHash())
                err = fmt.Errorf("Error duplicate work")
		return
	}

	if _,ok := duplicate_height_check[bl.Miner_TX.Vin[0].(transaction.Txin_gen).Height];ok{
		logger.Warnf("Block %s rejected by chain due to duplicate hwork.", bl.GetHash())
                err = fmt.Errorf("Error duplicate work")
		return
	}

	// lets build up the complete block

	// collect tx list + their fees
	var tx *transaction.Transaction
	for i := range bl.Tx_hashes {
		tx = chain.Mempool.Mempool_Get_TX(bl.Tx_hashes[i])
		if tx != nil {
			cbl.Txs = append(cbl.Txs, tx)
		} else {
			tx, err = chain.Load_TX_FROM_ID(nil, bl.Tx_hashes[i])
			if err != nil {
				logger.Warnf("Tx %s not found in pool or DB, rejecting submitted block", bl.Tx_hashes[i])
				return
			}
			cbl.Txs = append(cbl.Txs, tx)
		}
	}
	cbl.Bl = &bl // the block is now complete, lets try to add it to chain



	if !accept_limiter.Allow() { // if rate limiter allows, then add block to chain
		logger.Warnf("Block %s rejected by chain.", bl.GetHash())
		return
	}

	err, result = chain.Add_Complete_Block(cbl)

	if result {

		duplicate_check[bl.Miner_TX.GetHash()] = true
		duplicate_height_check[bl.Miner_TX.Vin[0].(transaction.Txin_gen).Height] = true

		logger.Infof("Block %s successfully accepted at height %d, Notifying Network", bl.GetHash(), chain.Load_Height_for_BL_ID(nil, bl.GetHash()))
		if !chain.simulator { // if not in simulator mode, relay block to the chain
			chain.P2P_Block_Relayer(cbl, 0) // lets relay the block to network
		}
	} else {
		logger.Warnf("Block Rejected %s error %s", bl.GetHash(), err)
	}
	return
}

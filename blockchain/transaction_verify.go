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

//import "fmt"
//import "time"
/*import "bytes"
import "encoding/binary"

import "github.com/romana/rlog"

*/
//import "github.com/romana/rlog"

import "runtime/debug"

import log "github.com/sirupsen/logrus"

import "github.com/deroproject/derosuite/block"
import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/crypto/ringct"
import "github.com/deroproject/derosuite/transaction"
import "github.com/deroproject/derosuite/emission"

/* This function verifies tx fully, means all checks,
 * if the transaction has passed the check it can be added to mempool, relayed or added to blockchain
 * the transaction has already been deserialized thats it
 * */
func (chain *Blockchain) Verify_Transaction(tx *transaction.Transaction) (result bool) {

	return false
}

/* Coinbase transactions need to verify the amount of coins
 * */
func (chain *Blockchain) Verify_Transaction_Coinbase(cbl *block.Complete_Block, minertx *transaction.Transaction) (result bool) {

	if !minertx.IsCoinbase() { // transaction is not coinbase, return failed
		return false
	}

	// coinbase transactions only have 1 vin, 1 vout
	if len(minertx.Vin) != 1 {
		return false
	}

	if len(minertx.Vout) != 1 {
		return false
	}

	// check whether the height mentioned in tx.Vin is equal to block height
	// this does NOT hold for genesis block so test it differently

	expected_height := chain.Load_Height_for_BL_ID(cbl.Bl.Prev_Hash)
	expected_height++
	if cbl.Bl.GetHash() == globals.Config.Genesis_Block_Hash {
		expected_height = 0
	}
	if minertx.Vin[0].(transaction.Txin_gen).Height != expected_height {
		logger.Warnf(" Rejected Height %d   should be %d", minertx.Vin[0].(transaction.Txin_gen).Height, chain.Load_Height_for_BL_ID(cbl.Bl.Prev_Hash))
		return false

	}

	// verify coins amount ( minied amount ) + the fees colllected which is sum of all tx included in this block
	// and whether it has been calculated correctly
	total_fees := uint64(0)
	for i := 0; i < len(cbl.Txs); i++ {
		total_fees += cbl.Txs[i].RctSignature.Get_TX_Fee()
	}

	total_reward := minertx.Vout[0].Amount
	base_reward := total_reward - total_fees

	// size of block = size of miner_tx + size of all non coinbase tx
	sizeofblock := uint64(0)
	sizeofblock += uint64(len(cbl.Bl.Miner_tx.Serialize()))
	//    logger.Infof("size of block %d  sizeof miner tx %d", sizeofblock, len(cbl.Bl.Miner_tx.Serialize()))
	for i := 0; i < len(cbl.Txs); i++ {
		sizeofblock += uint64(len(cbl.Txs[i].Serialize()))
		//        logger.Infof("size of tx i %d   tx size %d total size %d", i, uint64(len(cbl.Txs[i].Serialize())) ,sizeofblock)
	}

	median_block_size := chain.Get_Median_BlockSize_At_Block(cbl.Bl.Prev_Hash)

	already_generated_coins := chain.Load_Already_Generated_Coins_for_BL_ID(cbl.Bl.Prev_Hash)

	base_reward_calculated := emission.GetBlockReward(median_block_size, sizeofblock, already_generated_coins, 6, 0)
	if base_reward != base_reward_calculated {
		logger.Warnf("Base reward %d   should be %d", base_reward, base_reward_calculated)
		logger.Warnf("median_block_size %d   block_size %d already already_generated_coins %d", median_block_size, sizeofblock,
			already_generated_coins)
		return false
	}

	/*
	   for i := sizeofblock-4000; i < (sizeofblock+4000);i++{
	   // now we need to verify whether base reward is okay
	   //base_reward_calculated := emission.GetBlockReward(median_block_size,sizeofblock,already_generated_coins,6,0)

	       base_reward_calculated := emission.GetBlockReward(median_block_size,i,already_generated_coins,6,0)

	    if base_reward == base_reward_calculated {
	        logger.Warnf("Base reward %d   should be %d  size %d", base_reward, base_reward_calculated,i)

	    }

	   }
	*/

	return true
}

// all non miner tx must be non-coinbase tx
// each check is placed in a separate  block of code, to avoid ambigous code or faulty checks
// all check are placed and not within individual functions ( so as we cannot skip a check )
func (chain *Blockchain) Verify_Transaction_NonCoinbase(tx *transaction.Transaction) (result bool) {
	result = false

	var tx_hash crypto.Hash
	defer func() { // safety so if anything wrong happens, verification fails
		if r := recover(); r != nil {
			logger.WithFields(log.Fields{"txid": tx_hash}).Warnf("Recovered while Verifying transaction, failed verification, Stack trace below")
			logger.Warnf("Stack trace  \n%s", debug.Stack())
			result = false
		}
	}()

	tx_hash = tx.GetHash()

	if tx.Version != 2 {
		return false
	}

	// make sure atleast 1 vin and 1 vout are there
	if len(tx.Vin) < 1 || len(tx.Vout) < 1 {
		logger.WithFields(log.Fields{"txid": tx_hash}).Warnf("Incoming TX does NOT have atleast 1 vin and 1 vout")
		return false
	}

	// this means some other checks have failed somewhere else
	if tx.IsCoinbase() { // transaction coinbase must never come here
		logger.WithFields(log.Fields{"txid": tx_hash}).Warnf("Coinbase tx in non coinbase path, Please investigate")
		return false
	}

	// Vin can be only specific type rest all make the fail case
	for i := 0; i < len(tx.Vin); i++ {
		switch tx.Vin[i].(type) {
		case transaction.Txin_gen:
			return false // this is for coinbase so fail it
		case transaction.Txin_to_key: // pass
		default:
			return false
		}
	}

	// Vout can be only specific type rest all make th fail case
	for i := 0; i < len(tx.Vout); i++ {
		switch tx.Vout[i].Target.(type) {
		case transaction.Txout_to_key: // pass
		default:
			return false
		}
	}

	// Vout should have amount 0
	for i := 0; i < len(tx.Vout); i++ {
		if tx.Vout[i].Amount != 0 {
			logger.WithFields(log.Fields{"txid": tx_hash, "Amount": tx.Vout[i].Amount}).Warnf("Amount must be zero in ringCT world")
			return false
		}
	}

	// just some extra logs for testing purposes
	if len(tx.Vin) >= 3 {
		logger.WithFields(log.Fields{"txid": tx_hash}).Warnf("tx has more than 3 inputs")
	}

	if len(tx.Vout) >= 3 {
		logger.WithFields(log.Fields{"txid": tx_hash}).Warnf("tx has more than 3 outputs")
	}

	// check the mixin , it should be atleast 4 and should be same through out the tx ( all other inputs)
	// someone did send a mixin of 3 in 12006 block height
	mixin := len(tx.Vin[0].(transaction.Txin_to_key).Key_offsets)
	if mixin < 3 {
		logger.WithFields(log.Fields{"txid": tx_hash, "Mixin": mixin}).Warnf("Mixin must be atleast 3 in ringCT world")
		return false
	}
	for i := 0; i < len(tx.Vin); i++ {
		if mixin != len(tx.Vin[i].(transaction.Txin_to_key).Key_offsets) {
			logger.WithFields(log.Fields{"txid": tx_hash, "Mixin": mixin}).Warnf("Mixin must be same for entire TX in ringCT world")
			return false
		}
	}

	// duplicate ringmembers are not allowed, check them here
	// just in case protect ourselves as much as we can
	for i := 0; i < len(tx.Vin); i++ {
		ring_members := map[uint64]bool{} // create a separate map for each input
		ring_member := uint64(0)
		for j := 0; j < len(tx.Vin[i].(transaction.Txin_to_key).Key_offsets); j++ {
			ring_member += tx.Vin[i].(transaction.Txin_to_key).Key_offsets[j]
			if _, ok := ring_members[ring_member]; ok {
				logger.WithFields(log.Fields{"txid": tx_hash, "input_index": i}).Warnf("Duplicate ring member within the TX")
				return false
			}
			ring_members[ring_member] = true // add member to ring member
		}
	}

	// check whether the key image is duplicate within the inputs
	// NOTE: a block wide key_image duplication is done during block testing but we are still keeping it
	{
		kimages := map[crypto.Hash]bool{}
		for i := 0; i < len(tx.Vin); i++ {
			if _, ok := kimages[tx.Vin[i].(transaction.Txin_to_key).K_image]; ok {
				logger.WithFields(log.Fields{
					"txid":   tx_hash,
					"kimage": tx.Vin[i].(transaction.Txin_to_key).K_image,
				}).Warnf("TX using duplicate inputs within the TX")
				return false
			}
			kimages[tx.Vin[i].(transaction.Txin_to_key).K_image] = true // add element to map for next check
		}
	}

	// check whether the key image is low order attack, if yes reject it right now
	for i := 0; i < len(tx.Vin); i++ {
		k_image := ringct.Key(tx.Vin[i].(transaction.Txin_to_key).K_image)
		curve_order := ringct.CurveOrder()
		mult_result := ringct.ScalarMultKey(&k_image, &curve_order)
		if *mult_result != ringct.Identity {
			logger.WithFields(log.Fields{
				"txid":        tx_hash,
				"kimage":      tx.Vin[i].(transaction.Txin_to_key).K_image,
				"curve_order": curve_order,
				"mult_result": *mult_result,
				"identity":    ringct.Identity,
			}).Warnf("TX contains a low order key image attack, but we are already safeguarded")
			return false
		}
	}

	// a similiar block level check is done for double spending attacks within the block itself
	// check whether the key image is already used or spent earlier ( in blockchain )
/*
	for i := 0; i < len(tx.Vin); i++ {
		k_image := ringct.Key(tx.Vin[i].(transaction.Txin_to_key).K_image)
		if chain.Read_KeyImage_Status(crypto.Hash(k_image)) {
			logger.WithFields(log.Fields{
				"txid":   tx_hash,
				"kimage": k_image,
			}).Warnf("Key image is already spent, attempt to double spend ")
			return false
		}
	}
*/
	// check whether the TX contains a signature or NOT
	switch tx.RctSignature.Get_Sig_Type() {
	case ringct.RCTTypeSimple, ringct.RCTTypeFull: // default case, pass through
	default:
		logger.WithFields(log.Fields{"txid": tx_hash}).Warnf("TX does NOT contain a ringct signature. It is NOT possible")
		return false
	}

	// expand the signature first
	// whether the inputs are mature and can be used at time is verified while expanding the inputs
	if !chain.Expand_Transaction_v2(tx) {
		logger.WithFields(log.Fields{"txid": tx_hash}).Warnf("TX inputs could not be expanded or inputs are NOT mature")
		return false
	}

	// check the ring signature
	if !tx.RctSignature.Verify() {
		logger.WithFields(log.Fields{"txid": tx_hash}).Warnf("TX RCT Signature failed")
		return false

	}

	logger.WithFields(log.Fields{"txid": tx_hash}).Debugf("TX successfully verified")

	return true
}

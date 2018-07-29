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
import "math/big"

import "github.com/romana/rlog"

import "github.com/deroproject/derosuite/block"
import "github.com/deroproject/derosuite/config"
import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/storage"

var (
	// bigZero is 0 represented as a big.Int.  It is defined here to avoid
	// the overhead of creating it multiple times.
	bigZero = big.NewInt(0)

	// bigOne is 1 represented as a big.Int.  It is defined here to avoid
	// the overhead of creating it multiple times.
	bigOne = big.NewInt(1)

	// oneLsh256 is 1 shifted left 256 bits.  It is defined here to avoid
	// the overhead of creating it multiple times.
	oneLsh256 = new(big.Int).Lsh(bigOne, 256)

	// enabling this will simulation mode with hard coded difficulty set to 1
	// the variable is knowingly not exported, so no one can tinker with it
	//simulation = false // simulation mode is disabled
)

// HashToBig converts a PoW has into a big.Int that can be used to
// perform math comparisons.
func HashToBig(buf crypto.Hash) *big.Int {
	// A Hash is in little-endian, but the big package wants the bytes in
	// big-endian, so reverse them.
	blen := len(buf) // its hardcoded 32 bytes, so why do len but lets do it
	for i := 0; i < blen/2; i++ {
		buf[i], buf[blen-1-i] = buf[blen-1-i], buf[i]
	}

	return new(big.Int).SetBytes(buf[:])
}

// this function calculates the difficulty in big num form
func ConvertDifficultyToBig(difficultyi uint64) *big.Int {
	if difficultyi == 0 {
		panic("difficulty can never be zero")
	}
	// (1 << 256) / (difficultyNum )
	difficulty := new(big.Int).SetUint64(difficultyi)
	denominator := new(big.Int).Add(difficulty, bigZero) // above 2 lines can be merged
	return new(big.Int).Div(oneLsh256, denominator)
}

func ConvertIntegerDifficultyToBig(difficultyi *big.Int) *big.Int {

	if difficultyi.Cmp(bigZero) == 0 { // if work_pow is less than difficulty
		panic("difficulty can never be zero")
	}

	return new(big.Int).Div(oneLsh256, difficultyi)
}

// this function check whether the pow hash meets difficulty criteria
func CheckPowHash(pow_hash crypto.Hash, difficulty uint64) bool {
	big_difficulty := ConvertDifficultyToBig(difficulty)
	big_pow_hash := HashToBig(pow_hash)

	if big_pow_hash.Cmp(big_difficulty) <= 0 { // if work_pow is less than difficulty
		return true
	}
	return false
}

// this function check whether the pow hash meets difficulty criteria
// however, it take diff in  bigint format
func CheckPowHashBig(pow_hash crypto.Hash, big_difficulty_integer *big.Int) bool {
	big_pow_hash := HashToBig(pow_hash)

	big_difficulty := ConvertIntegerDifficultyToBig(big_difficulty_integer)
	if big_pow_hash.Cmp(big_difficulty) <= 0 { // if work_pow is less than difficulty
		return true
	}
	return false
}

// this function finds a common base  which can be used to compare tips based on cumulative difficulty
func (chain *Blockchain) find_best_tip_cumulative_difficulty(dbtx storage.DBTX, tips []crypto.Hash) (best crypto.Hash) {

	tips_scores := make([]BlockScore, len(tips), len(tips))

	for i := range tips {
		tips_scores[i].BLID = tips[i] // we should chose the lowest weight
		tips_scores[i].Cumulative_Difficulty = chain.Load_Block_Cumulative_Difficulty(dbtx, tips[i])
	}

	sort_descending_by_cumulative_difficulty(tips_scores)

	best = tips_scores[0].BLID
	//   base_height = scores[0].Weight

	return best

}

// confirms whether the actual tip difficulty is withing 9% deviation with reference
// actual tip cannot be less than 91% of main tip
// if yes tip is okay, else tip should be declared stale
// both the tips should be in the store
func (chain *Blockchain) validate_tips(dbtx storage.DBTX, reference, actual crypto.Hash) (result bool) {

	reference_diff := chain.Load_Block_Difficulty(dbtx, reference)
	actual_diff := chain.Load_Block_Difficulty(dbtx, actual)

	// multiply by 91
	reference91 := new(big.Int).Mul(reference_diff, new(big.Int).SetUint64(91))
	// divide by 100
	reference91.Div(reference91, new(big.Int).SetUint64(100))

	if reference91.Cmp(actual_diff) < 0 {
		return true
	} else {
		return false
	}

}

// when creating a new block, current_time in utc + chain_block_time must be added
// while verifying the block, expected time stamp should be replaced from what is in blocks header
// in DERO atlantis difficulty is based on previous tips
// get difficulty at specific  tips,
// algorithm is as follows choose biggest difficulty tip (// division is integer and not floating point)
// diff = (parent_diff +   (parent_diff / 100 * max(1 - (parent_timestamp - parent_parent_timestamp) // (chain_block_time*2//3), -1))
// this should be more thoroughly evaluated

// NOTE: we need to evaluate if the mining adversary gains something, if the they set the time diff to 1
// we need to do more simulations and evaluations
func (chain *Blockchain) Get_Difficulty_At_Tips(dbtx storage.DBTX, tips []crypto.Hash) *big.Int {

 	var MinimumDifficulty *big.Int

 	if globals.IsMainnet() {
 		MinimumDifficulty = new(big.Int).SetUint64(config.MAINNET_MINIMUM_DIFFICULTY) // this must be controllable parameter

 		}else{
 			MinimumDifficulty = new(big.Int).SetUint64(config.TESTNET_MINIMUM_DIFFICULTY) // this must be controllable parameter
 		}
	//MinimumDifficulty := new(big.Int).SetUint64(131072) // TODO  tthis must be controllable parameter
	GenesisDifficulty := new(big.Int).SetUint64(1)

	if chain.simulator == true {
		return GenesisDifficulty
	}

	if len(tips) == 0 { // genesis block difficulty is 1
		return GenesisDifficulty // it should be configurable via params
	}

	height := chain.Calculate_Height_At_Tips(dbtx, tips)

	// hard fork version 1 has difficulty set to  1
	if 1 == chain.Get_Current_Version_at_Height(height) {
		return new(big.Int).SetUint64(1)
	}

	// if we are hardforking from 1 to 2
	// we can start from high difficulty to find the right point
	if height >= 1 && chain.Get_Current_Version_at_Height(height-1) == 1 && chain.Get_Current_Version_at_Height(height) == 2 {
		if globals.IsMainnet() {
			bootstrap_difficulty := new(big.Int).SetUint64(config.MAINNET_BOOTSTRAP_DIFFICULTY) // return bootstrap mainnet difficulty
			rlog.Infof("Returning bootstrap difficulty %s at height %d", bootstrap_difficulty.String(), height)
			return bootstrap_difficulty
		} else {
			bootstrap_difficulty := new(big.Int).SetUint64(config.TESTNET_BOOTSTRAP_DIFFICULTY)
			rlog.Infof("Returning bootstrap difficulty %s at height %d", bootstrap_difficulty.String(), height)
			return bootstrap_difficulty // return bootstrap difficulty for testnet
		}
	}

	// for testing purposes, not possible on mainchain
	if height < 4 && chain.Get_Current_Version_at_Height(height) >= 2 {
		return MinimumDifficulty
	}

	/*
		// build all blocks whivh are reachale
		// process only which are close to the chain
		reachable_blocks := chain.BuildReachableBlocks(dbtx,tips)
		var difficulty_sum big.Int // used to calculate average difficulty
		var average_difficulty big.Int
		var lowest_average_difficulty big.Int
		var block_count int64
		for k,_ := range reachable_blocks{
			height_of_k := chain.Load_Height_for_BL_ID(dbtx,k)
			if (height - height_of_k) <=  ((config.STABLE_LIMIT*3)/4) {
				block_count++
				difficulty_of_k := chain.Load_Block_Difficulty(dbtx, k)
				difficulty_sum.Add(&difficulty_sum, difficulty_of_k)
			}
		}

		// used to rate limit maximum drop over a certain number of blocks
		average_difficulty.Div(&difficulty_sum,new(big.Int).SetInt64(block_count))
		average_difficulty.Mul(&average_difficulty,new(big.Int).SetUint64(92))  //max 10 % drop
		average_difficulty.Div(&average_difficulty,new(big.Int).SetUint64(100))

		lowest_average_difficulty.Set(&average_difficulty) // difficulty can never drop less than this


	*/

	biggest_tip := chain.find_best_tip_cumulative_difficulty(dbtx, tips)
	biggest_difficulty := chain.Load_Block_Difficulty(dbtx, biggest_tip)

	//  take the time from the most heavy block
	parent_highest_time := chain.Load_Block_Timestamp(dbtx, biggest_tip)

	// find parents parents tip which hash highest tip
	parent_past := chain.Get_Block_Past(dbtx, biggest_tip)

	past_biggest_tip := chain.find_best_tip_cumulative_difficulty(dbtx, parent_past)
	parent_parent_highest_time := chain.Load_Block_Timestamp(dbtx, past_biggest_tip)

	if biggest_difficulty.Cmp(MinimumDifficulty) < 0 {
		biggest_difficulty.Set(MinimumDifficulty)
	}

	// create 3 ranges, used for physical verification
	/*
		switch {
			case (parent_highest_time - parent_parent_highest_time) <= 6: // increase diff
					logger.Infof(" increase diff")
			case (parent_highest_time - parent_parent_highest_time) >= 12: // decrease diff
					logger.Infof(" decrease diff")

			default ://  between 6 to 12, 7,8,9,10,11 do nothing, return previous difficulty
			logger.Infof("stable diff diff")
		}*/

	bigTime := new(big.Int).SetInt64(parent_highest_time)
	bigParentTime := new(big.Int).SetInt64(parent_parent_highest_time)

	// holds intermediate values to make the algo easier to read & audit
	x := new(big.Int)
	y := new(big.Int)

	// 1 - (block_timestamp - parent_timestamp) // ((config.BLOCK_TIME*2)/3)
	// the above creates the following ranges  0-5 , increase diff 6-11 keep it constant,  above 12 and above decrease
	big1 := new(big.Int).SetUint64(1)
	big_block_chain_time_range := new(big.Int).SetUint64((config.BLOCK_TIME * 2) / 3)
	DifficultyBoundDivisor := new(big.Int).SetUint64(100) // granularity of 100 steps to increase or decrease difficulty
	bigmaxdifficulydrop := new(big.Int).SetInt64(-2)      // this should ideally be .05% of difficuly bound divisor, but currentlt its 0.5 %

	x.Sub(bigTime, bigParentTime)

	x.Div(x, big_block_chain_time_range)
	//logger.Infof(" block time -  parent time %d %s / 6",parent_highest_time - parent_parent_highest_time,  x.String())

	x.Sub(big1, x)

	//logger.Infof("x  %s biggest %s lowest average %s ", x.String(), biggest_difficulty, lowest_average_difficulty.String())

	// max(1 - (block_timestamp - parent_timestamp) // chain_block_time, -99)
	if x.Cmp(bigmaxdifficulydrop) < 0 {
		x.Set(bigmaxdifficulydrop)
	}
	// logger.Infof("x  %s biggest %s ", x.String(), biggest_difficulty)

	// (parent_diff + parent_diff // 2048 * max(1 - (block_timestamp - parent_timestamp) // 10, -99))
	y.Div(biggest_difficulty, DifficultyBoundDivisor)

	// decreases are 1/2 of increases
	// this will cause the network to adjust slower to big difficulty drops
	// but has more benefits
	/*if x.Sign() < 0 {
		logger.Infof("decrease will be  1//2 ")

		y.Div(y, new(big.Int).SetUint64(2))
	}*/

	//logger.Infof("max increase/decrease %s   x %s", y.String(), x.String())

	x.Mul(y, x)
	x.Add(biggest_difficulty, x)

	/*
		// if difficulty drop is more than X% than the average, limit it here
		if x.Cmp(&lowest_average_difficulty) < 0{
			x.Set(&lowest_average_difficulty)
		}

	*/
	//
	// minimum difficulty can ever be
	if x.Cmp(MinimumDifficulty) < 0 {
		x.Set(MinimumDifficulty)
	}
	// logger.Infof("Final diff  %s biggest %s lowest average %s ", x.String(), biggest_difficulty, lowest_average_difficulty.String())

	return x
}

func (chain *Blockchain) VerifyPoW(dbtx storage.DBTX, bl *block.Block) (verified bool) {

	verified = false
	//block_work := bl.GetBlockWork()

	//PoW := crypto.Scrypt_1024_1_1_256(block_work)
	//PoW := crypto.Keccak256(block_work)
	PoW := bl.GetPoWHash()

	block_difficulty := chain.Get_Difficulty_At_Tips(dbtx, bl.Tips)

	// test new difficulty checksm whether they are equivalent to integer math

	/*if CheckPowHash(PoW, block_difficulty.Uint64()) != CheckPowHashBig(PoW, block_difficulty) {
		logger.Panicf("Difficuly mismatch between big and uint64 diff ")
	}*/

	if CheckPowHashBig(PoW, block_difficulty) == true {
		return true
	}
	/* *
	if CheckPowHash(PoW, block_difficulty.Uint64()) == true {
		return true
	}*/

	return false
}

// this function calculates difficulty on the basis of previous difficulty  and number of blocks
// THIS is the ideal algorithm for us as it will be optimal based on the number of orphan blocks
// we may deploy it when the  block reward becomes insignificant in comparision to fees
//  basically tail emission kicks in or we need to optimally increase number of blocks
// the algorithm does NOT work if the network has a single miner  !!!
// this algorithm will work without the concept of time

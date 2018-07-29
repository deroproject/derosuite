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

package emission

import "fmt"
import "math/big"
import "github.com/deroproject/derosuite/config"

//TODO trickling code  is note implemented still, but we do NOT require it atleast for another 7-8 years

// the logic is same as cryptonote_basic_impl.cpp

// this file controls the logic for emission of coins at each height
// calculates block reward

func GetBlockReward(bl_median_size uint64,
	bl_current_size uint64,
	already_generated_coins uint64,
	hard_fork_version uint64,
	fee uint64) (reward uint64) {

	target := config.COIN_DIFFICULTY_TARGET
	target_minutes := target / 60
	emission_speed_factor := config.COIN_EMISSION_SPEED_FACTOR - (target_minutes - 1)
	// handle special cases
	switch already_generated_coins {
	case 0:
		reward = 1000000000000 // give 1 DERO to genesis, but we gave 35 due to a silly typo, so continue as is
		return reward
	case 1000000000000:
		reward = 2000000 * 1000000000000 // give the developers initial premine, while keeping the, into account and respecting transparancy
		return reward
	}

	base_reward := (config.COIN_MONEY_SUPPLY - already_generated_coins) >> emission_speed_factor
	if base_reward < (config.COIN_FINAL_SUBSIDY_PER_MINUTE * target_minutes) {
		base_reward = config.COIN_FINAL_SUBSIDY_PER_MINUTE * target_minutes
	}

	//full_reward_zone = get_min_block_size(version);
	full_reward_zone := config.CRYPTONOTE_BLOCK_GRANTED_FULL_REWARD_ZONE

	if bl_median_size < full_reward_zone {
		bl_median_size = full_reward_zone
	}

	if bl_current_size <= bl_median_size {
		reward = base_reward
		return reward
	}

	// block is bigger than median size , we must calculate it
	if bl_current_size > 2*bl_median_size {
		//MERROR("Block cumulative size is too big: " << current_block_size << ", expected less than " << 2 * median_size);
		panic(fmt.Sprintf("Block size is too big current size %d  max possible size %d", bl_current_size, 2*bl_median_size))
	}

	multiplicand := (2 * bl_median_size) - bl_current_size
	multiplicand = multiplicand * bl_current_size

	var big_base_reward, big_multiplicand, big_product, big_reward, big_median_size big.Int

	big_median_size.SetUint64(bl_median_size)
	big_base_reward.SetUint64(base_reward)
	big_multiplicand.SetUint64(multiplicand)

	big_product.Mul(&big_base_reward, &big_multiplicand)

	big_reward.Div(&big_product, &big_median_size)
	big_reward.Div(&big_reward, &big_median_size)
	// lower 64 bits contains the reward
	if !big_reward.IsUint64() {
		panic("GetBlockReward has issues\n")
	}

	reward_lo := big_reward.Uint64()

	if reward_lo > base_reward {
		panic("Reward must be less than base reward\n")
	}
	return reward_lo
}

// atlantis has very simple block reward
// since our chain has already bootstrapped
//  FIXME this will not workaround , when already already_generated_coins wraps around
// but we have few years, atleast 6-7 to fix it
func GetBlockReward_Atlantis(hard_fork_version int64, already_generated_coins uint64) (reward uint64) {

	target := uint64(120) // initial target was 120 secs however difficult targeted 180 secs
	target_minutes := target / 60
	emission_speed_factor := config.COIN_EMISSION_SPEED_FACTOR - (target_minutes - 1)

	base_reward := (config.COIN_MONEY_SUPPLY - already_generated_coins) >> emission_speed_factor
	if base_reward < (config.COIN_FINAL_SUBSIDY_PER_MINUTE * target_minutes) {
		base_reward = config.COIN_FINAL_SUBSIDY_PER_MINUTE * target_minutes
	}

	// however the new target is less than 10 secs, so divide the reward into equal parts
	//base_reward = (base_reward * config.BLOCK_TIME)/ target
	base_reward = (base_reward * config.BLOCK_TIME) / 180 // original daemon emission schedule

	return base_reward
}

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

import "math/big"
import "github.com/deroproject/derosuite/config"

// this file implements the logic to calculate fees dynamicallly

//  get mininum size of block
func Get_Block_Minimum_Size(version uint64) uint64 {
	return config.CRYPTONOTE_BLOCK_GRANTED_FULL_REWARD_ZONE
}

// get maximum size of TX
func Get_Transaction_Maximum_Size() uint64 {
	return config.CRYPTONOTE_MAX_TX_SIZE
}

// get per KB fees which is dependent  of block reward and median_block_size
//
func (chain *Blockchain) Get_Dynamic_per_kb_fee(block_reward, median_block_size, version uint64) uint64 {
	fee_per_kb_base := config.DYNAMIC_FEE_PER_KB_BASE_FEE_V5

	min_block_size := Get_Block_Minimum_Size(version)
	if median_block_size < min_block_size {
		median_block_size = min_block_size
	}

	var unscaled_fee_per_kb big.Int

	unscaled_fee_per_kb.SetUint64((fee_per_kb_base * min_block_size) / median_block_size)
	unscaled_fee_per_kb.Mul(&unscaled_fee_per_kb, big.NewInt(int64(block_reward))) // block reward is always positive
	unscaled_fee_per_kb.Div(&unscaled_fee_per_kb, big.NewInt(int64(config.DYNAMIC_FEE_PER_KB_BASE_BLOCK_REWARD)))

	if !unscaled_fee_per_kb.IsUint64() {
		panic("Get_Dynamic_per_kb_fee has issues, Need to fix it urgently\n")
	}

	lo := unscaled_fee_per_kb.Uint64()
	// we need to quantise it 8 decimal // make 4 lower digits 0
	mask := uint64(10000)

	qlo := (lo + mask - 1) / mask * mask

	//logger.Infof("dynamic fee lo= %d  qlo=%d", lo, qlo)

	return qlo
}

// this will give the dynamic fee at any specfic height
func (chain *Blockchain) Get_Dynamic_Fee_Rate(height uint64) uint64 {
	var base_reward, median_block_size uint64

	// sanitize height
	if height >= chain.Get_Height() {
		height = chain.Load_Height_for_BL_ID(chain.Get_Top_ID())
	}

	block_id_at_height, _ := chain.Load_BL_ID_at_Height(height)

	median_block_size = chain.Get_Median_BlockSize_At_Block(block_id_at_height)
	base_reward = chain.Load_Block_Reward(block_id_at_height)

	return chain.Get_Dynamic_per_kb_fee(base_reward, median_block_size, 0)
}

// get the tx fee
// this function assumes that dynamic fees has been calculated as per requirement
// for every part of 1KB multiply by dynamic_fee_per_kb
// the dynamic fee should be calculated on the basis of the current top
func (chain *Blockchain) Calculate_TX_fee(dynamic_fee_per_kb uint64, tx_size uint64) uint64 {
	size_in_kb := tx_size / 1024

	if (tx_size % 1024) != 0 { // for any part there of, use a full KB fee
		size_in_kb += 1
	}

	needed_fee := size_in_kb * dynamic_fee_per_kb
	return needed_fee

}

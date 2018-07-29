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

//import "math/big"

import "github.com/deroproject/derosuite/config"

//import "github.com/deroproject/derosuite/emission"

// this file implements the logic to calculate fees dynamicallly

//  get mininum size of block
/*func Get_Block_Minimum_Size(version uint64) uint64 {
	return config.CRYPTONOTE_BLOCK_GRANTED_FULL_REWARD_ZONE
}*/

// get maximum size of TX
func Get_Transaction_Maximum_Size() uint64 {
	return config.CRYPTONOTE_MAX_TX_SIZE
}

// get the tx fee
// this function assumes that  fees are per KB
// for every part of 1KB multiply by fee_per_kb
func (chain *Blockchain) Calculate_TX_fee(hard_fork_version int64, tx_size uint64) uint64 {
	size_in_kb := tx_size / 1024

	if (tx_size % 1024) != 0 { // for any part there of, use a full KB fee
		size_in_kb += 1
	}

	needed_fee := size_in_kb * config.FEE_PER_KB
	return needed_fee

}

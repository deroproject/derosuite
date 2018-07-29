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

package inputmaturity

import "time"
import "github.com/deroproject/derosuite/config"

//this file implements the logic to detect whether an input is mature or still locked
//This function is crucial as any bugs can have catastrophic effects
//this function is used both by the core blockchain and wallet
//TODO we need to check the edge cases
func Is_Input_Mature(current_chain_height uint64, input_block_height uint64, locked_to_height uint64, sigtype uint64) bool {

	// current_chain_height must be greater or equal to input height
	if current_chain_height < input_block_height { // wrong cases reject
		return false
	}

	// all miner txs end up here
	if sigtype == 0 { // miner tx will also always unlock 60 blocks in future
		if current_chain_height >= (input_block_height + config.MINER_TX_AMOUNT_UNLOCK) {
			return true
		}
		return false
	}

	// 99.99 % normal tx cases come here
	if locked_to_height == 0 { // input is not locked, so it must be unlocked in 8 blocks
		if current_chain_height >= (input_block_height + (config.NORMAL_TX_AMOUNT_UNLOCK)) {
			return true
		}
		return false
	}

	// input_block_height is no longer useful

	//  if we are here input was locked to specific height or time
	if locked_to_height < config.CRYPTONOTE_MAX_BLOCK_NUMBER { // input was locked to specific block height
		if current_chain_height >= locked_to_height {
			return true
		}
		return false
	}

	// anything above LIMIT is time based lock
	// input was locked to specific time in epoch
	if locked_to_height < (uint64(time.Now().UTC().Unix())) {
		return true
	}

	return false // reject time based locks if not mature
}

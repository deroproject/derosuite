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

import "testing"

import "github.com/deroproject/derosuite/config"

//this file implements the logic to detect whether an input is mature or still locked
//This function is crucial as any bugs can have catastrophic effects
//this function is used both by the core blockchain and wallet
//TODO we need to check the edge cases
func Test_Input_Maturity(t *testing.T) {
	tests := []struct {
		name                 string
		current_chain_height uint64
		input_block_height   uint64
		locked_to_height     uint64
		sigtype              uint64
		expected             bool
	}{
		{
			name:                 "simple test",
			current_chain_height: 0,
			input_block_height:   1,
			locked_to_height:     0,
			sigtype:              0,
			expected:             false,
		},
		// miner tx blocks mature in 60 blocks
		{
			name:                 "miner test 59",
			current_chain_height: 59,
			input_block_height:   0,
			locked_to_height:     0,
			sigtype:              0,
			expected:             false,
		},
		{
			name:                 "miner test 60",
			current_chain_height: 60,
			input_block_height:   0,
			locked_to_height:     0,
			sigtype:              0,
			expected:             true,
		},

		{
			name:                 "miner test 60", // genesis block reward should mature at block 61
			current_chain_height: 61,
			input_block_height:   0,
			locked_to_height:     0,
			sigtype:              0,
			expected:             true,
		},

		// normal tx output matures in 10 blocks
		{
			name:                 "normal test 9", // reward should mature at block 10
			current_chain_height: 9,
			input_block_height:   0,
			locked_to_height:     0,
			sigtype:              1,
			expected:             false,
		},
		{
			name:                 "normal test 10", //reward should mature at block 10
			current_chain_height: 10,
			input_block_height:   0,
			locked_to_height:     0,
			sigtype:              1,
			expected:             false,
		},
		{
			name:                 "normal test 11", //  reward should mature at block 11
			current_chain_height: 11,
			input_block_height:   0,
			locked_to_height:     0,
			sigtype:              1,
			expected:             true,
		},

		// height based lock
		{
			name:                 "locked_to_height ", //  reward should mature at specific block
			current_chain_height: config.CRYPTONOTE_MAX_BLOCK_NUMBER - 1,
			input_block_height:   0,
			locked_to_height:     config.CRYPTONOTE_MAX_BLOCK_NUMBER - 1,
			sigtype:              1,
			expected:             true,
		},

		{
			name:                 "locked_to_height false", //  reward should mature at block 11
			current_chain_height: config.CRYPTONOTE_MAX_BLOCK_NUMBER - 2,
			input_block_height:   0,
			locked_to_height:     config.CRYPTONOTE_MAX_BLOCK_NUMBER - 1,
			sigtype:              1,
			expected:             false,
		},

		// time based locked

		{
			name:                 "locked_to_time false",
			current_chain_height: config.CRYPTONOTE_MAX_BLOCK_NUMBER,
			input_block_height:   0,
			locked_to_height:     15174219710,
			sigtype:              1,
			expected:             false,
		},

		{
			name:                 "locked_to_time true",
			current_chain_height: config.CRYPTONOTE_MAX_BLOCK_NUMBER,
			input_block_height:   0,
			locked_to_height:     1517421971,
			sigtype:              1,
			expected:             true,
		},
	}

	for _, test := range tests {
		actual := Is_Input_Mature(test.current_chain_height, test.input_block_height, test.locked_to_height, test.sigtype)
		if actual != test.expected {
			t.Fatalf("Input Maturity testing failed name %s  actual %v  expected %v", test.name, actual, test.expected)
		}
	}

}

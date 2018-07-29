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

//// + build ignore

package emission

import "testing"

// Rough estimates of block rewards at specific heights
func Test_Fees_Data(t *testing.T) {

	already_generated_coins := uint64(0)
	reward := uint64(0)

	//for i:=0 ; i < 32620;i++ { 2 months
	for i := 0; i < 10*6; i++ { // 1 year
		reward = GetBlockReward(0, 0, already_generated_coins, 0, 0)
		already_generated_coins += reward
	}

	for i := uint64(280); i < 585; i += 10 {
		reward = GetBlockReward(0, 0, already_generated_coins, 0, 0)
		reward_with_penalty := GetBlockReward(0, i*1024, already_generated_coins, 0, 0)
		t.Logf("block reward %.2f penalty %.2f   size %d KB ", float64(reward)/1000000000000.0, (float64(reward-reward_with_penalty) / 1000000000000.0), i)
	}

}

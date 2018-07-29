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

//import "encoding/hex"
//import "bytes"
import "testing"

func Test_Emission_Rewards(t *testing.T) {

	bl_median_size := uint64(0)
	bl_current_size := uint64(0)
	already_generated_coins := uint64(0)
	hard_fork_version := uint64(6)
	fee := uint64(0)
	if GetBlockReward(bl_median_size, bl_current_size, already_generated_coins, hard_fork_version, fee) != uint64(1000000000000) {
		t.Error("Block reward failed for genesis block\n")
	}

	already_generated_coins = uint64(1000000000000)

	if GetBlockReward(bl_median_size, bl_current_size, already_generated_coins, hard_fork_version, fee) != uint64(2000000000000000000) {
		t.Error("Block reward failed for pre-mine\n")
	}

	already_generated_coins += uint64(2000000000000000000)

	calculated := GetBlockReward(bl_median_size, bl_current_size, already_generated_coins, hard_fork_version, fee)
	expected := uint64(31369672915858)
	if calculated != expected {
		t.Errorf("Block reward failed for 2nd block estm %d calculated %d\n", expected, calculated)
	}

	already_generated_coins += uint64(31369672915858)

	// calculated block reward for 3rd block
	calculated = GetBlockReward(bl_median_size, bl_current_size, already_generated_coins, hard_fork_version, fee)
	expected = uint64(31369613082955)

	if calculated != expected {
		t.Errorf("Block reward failed for 3rd block estm %d calculated %d\n", expected, calculated)
	}

	// 4 consecutive blocks were filled so test them
	// tests if the block size is greater than median block ( penalty )
	// test from block
	bl_median_size = 300000
	bl_current_size = 301174
	already_generated_coins = 3160947266765937256
	calculated = GetBlockReward(bl_median_size, bl_current_size, already_generated_coins, hard_fork_version, fee)
	expected = uint64(29154897151057)
	if calculated != expected {
		t.Errorf("Block reward failed for 38383 block estm %d calculated %d\n", expected, calculated)
	}

	bl_median_size = 300000
	bl_current_size = 301237
	already_generated_coins = 3160976421663088313
	calculated = GetBlockReward(bl_median_size, bl_current_size, already_generated_coins, hard_fork_version, fee)
	expected = uint64(29154792337976)
	if calculated != expected {
		t.Errorf("Block reward failed for 38384 block estm %d calculated %d\n", expected, calculated)
	}

	bl_median_size = 300000
	bl_current_size = 301745
	already_generated_coins = 3161005576455426289
	calculated = GetBlockReward(bl_median_size, bl_current_size, already_generated_coins, hard_fork_version, fee)
	expected = uint64(29154245997471)
	if calculated != expected {
		t.Errorf("Block reward failed for 38385 block estm %d calculated %d\n", expected, calculated)
	}

	bl_median_size = 300000
	bl_current_size = 302406
	already_generated_coins = 3161034730701423760
	calculated = GetBlockReward(bl_median_size, bl_current_size, already_generated_coins, hard_fork_version, fee)
	expected = uint64(29153301544320)
	if calculated != expected {
		t.Errorf("Block reward failed for 38386 block estm %d calculated %d \n", expected, calculated)
	}

	// TODO test trickle emission
	bl_median_size = 0
	bl_current_size = 0
	already_generated_coins = 0xffffffffffffffff
	calculated = GetBlockReward(bl_median_size, bl_current_size, already_generated_coins, hard_fork_version, fee)
	expected = uint64(600000000000)
	if calculated != expected {
		t.Errorf("Block reward failed for trickle block estm %d calculated %d \n", expected, calculated)
	}

}

func Test_Emission_Panic_For_Huge_Block(t *testing.T) {

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Emission did not panic on huge blocks")
		}
	}()

	bl_median_size := uint64(9)
	bl_current_size := uint64(7 * 100 * 1024) // default median size is 300KB
	already_generated_coins := uint64(3161034730701423760)
	hard_fork_version := uint64(6)
	fee := uint64(0)
	expected := GetBlockReward(bl_median_size, bl_current_size, already_generated_coins, hard_fork_version, fee) // this should panic

	if expected >= 0 {
		t.Errorf("We should never land here %d\n", expected)
	}

}

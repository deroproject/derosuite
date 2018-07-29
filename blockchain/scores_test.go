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

import "testing"
import "math/big"

import "github.com/deroproject/derosuite/crypto"

func Test_Scores_Sorting(t *testing.T) {

	tips_scores := []BlockScore{
		{
			BLID:                  crypto.HashHexToHash("0000000000000000000000000000000000000000000000000000000000000000"),
			Height:                76,
			Cumulative_Difficulty: new(big.Int).SetUint64(95),
		},
		{
			BLID:                  crypto.HashHexToHash("0000000000000000000000000000000000000000000000000000000000000099"),
			Height:                7,
			Cumulative_Difficulty: new(big.Int).SetUint64(99),
		},
		{
			BLID:                  crypto.HashHexToHash("0000000000000000000000000000000000000000000000000000000000000011"),
			Height:                45,
			Cumulative_Difficulty: new(big.Int).SetUint64(99),
		},
	}

	sort_descending_by_cumulative_difficulty(tips_scores)

	if tips_scores[0].BLID != crypto.HashHexToHash("0000000000000000000000000000000000000000000000000000000000000011") {
		t.Fatalf("core sorting test failed")
	}
	if tips_scores[1].BLID != crypto.HashHexToHash("0000000000000000000000000000000000000000000000000000000000000099") {
		t.Fatalf("core sorting test 2 failed")
	}

	sort_ascending_by_height(tips_scores)

	if tips_scores[0].BLID != crypto.HashHexToHash("0000000000000000000000000000000000000000000000000000000000000099") {
		t.Fatalf("core sorting test by height failed")
	}
	if tips_scores[1].BLID != crypto.HashHexToHash("0000000000000000000000000000000000000000000000000000000000000011") {
		t.Fatalf("core sorting test 2 by height failed")
	}
}

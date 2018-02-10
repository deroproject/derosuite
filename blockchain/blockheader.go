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
import "github.com/deroproject/derosuite/crypto"

// this is used to print blockheader for the rpc and the daemon
type BlockHeader_Print struct {
	Depth         uint64 `json:"depth"`
	Difficulty    uint64 `json:"difficulty"`
	Hash          string `json:"hash"`
	Height        uint64 `json:"height"`
	Major_Version uint64 `json:"major_version"`
	Minor_Version uint64 `json:"minor_version"`
	Nonce         uint64 `json:"nonce"`
	Orphan_Status bool   `json:"orphan_status"`
	Reward        uint64 `json:"reward"`
	Prev_Hash     string `json:"prev_hash"`
	Timestamp     uint64 `json:"timestamp"`
}

/* fill up the above structure from the blockchain */
func (chain *Blockchain) GetBlockHeader(hash crypto.Hash) (result BlockHeader_Print, err error) {

	bl, err := chain.Load_BL_FROM_ID(hash)
	if err != nil {
		return
	}

	result.Height = chain.Load_Height_for_BL_ID(hash)
	result.Depth = chain.Get_Height() - result.Height - 1
	result.Difficulty = chain.Get_Difficulty_At_Block(hash)
	result.Hash = hash.String()
	result.Height = chain.Load_Height_for_BL_ID(hash)
	result.Major_Version = uint64(bl.Major_Version)
	result.Minor_Version = uint64(bl.Minor_Version)
	result.Nonce = uint64(bl.Nonce)
	result.Orphan_Status = chain.Is_Block_Orphan(hash)
	result.Reward = chain.Load_Block_Reward(hash)

	result.Prev_Hash = bl.Prev_Hash.String()
	result.Timestamp = bl.Timestamp

	return
}

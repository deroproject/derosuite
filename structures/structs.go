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

// this package contains only struct definitions
// in order to avoid the dependency on block chain by any package requiring access to rpc
// and other structures
// having the structures was causing the build times of explorer/wallet to be more than couple of secs
// so separated the logic from the structures

package structures

// this is used to print blockheader for the rpc and the daemon
type BlockHeader_Print struct {
	Depth         int64  `json:"depth"`
	Difficulty    string `json:"difficulty"`
	Hash          string `json:"hash"`
	Height        int64  `json:"height"`
	TopoHeight    int64  `json:"topoheight"`
	Major_Version uint64 `json:"major_version"`
	Minor_Version uint64 `json:"minor_version"`
	Nonce         uint64 `json:"nonce"`
	Orphan_Status bool   `json:"orphan_status"`
	SyncBlock     bool   `json:"syncblock"`
	TXCount       int64  `json:"txcount"`

	Reward    uint64   `json:"reward"`
	Tips      []string `json:"tips"`
	Timestamp uint64   `json:"timestamp"`
}

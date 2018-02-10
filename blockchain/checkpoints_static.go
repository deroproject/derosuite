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

import "fmt"
import "encoding/hex"

//import log "github.com/sirupsen/logrus"

import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/globals"

var mainnet_static_checkpoints = map[uint64]crypto.Hash{}
var testnet_static_checkpoints = map[uint64]crypto.Hash{}

// initialize the checkpoints only few manually
// just in case we decide to skip adding  blocklist in the future
func init_static_checkpoints() {

	switch globals.Config.Name {
	case "mainnet":
		ADD_CHECKPOINT(mainnet_static_checkpoints, 1, "cea3eb82889a1d063aa6205011a27a108e727f2ac66721a46bec8b7ee9e83d7e")
		ADD_CHECKPOINT(mainnet_static_checkpoints, 10, "a8feb533ad2ad021f356b964957f4880445f89d6c6658a9407d63ce5144fe8ea")
		ADD_CHECKPOINT(mainnet_static_checkpoints, 100, "462ef1347bd00511ad6e7be463cba4d44a69dbf8b7d1d478ff1fd68507dfc9e2")

		logger.Debugf("Added %d static checkpoints to mainnet", len(mainnet_static_checkpoints))
	case "testnet":
		logger.Debugf("Added %d static checkpoints to testnet", len(testnet_static_checkpoints))
	default:
		panic(fmt.Sprintf("Unknown Network \"%s\"", globals.Config.Name))
	}

}

// add a checkpoint to specific network
func ADD_CHECKPOINT(x map[uint64]crypto.Hash, Height uint64, hash_hex string) {

	var hash crypto.Hash
	hash_raw, err := hex.DecodeString(hash_hex)

	if err != nil {
		panic(fmt.Sprintf("Cannot hex decode checkpint hash \"%s\"", hash_hex))
	}

	if len(hash_raw) != 32 {
		panic(fmt.Sprintf(" hash not 32 byte size Cannot hex decode checkpint hash \"%s\"", hash_hex))
	}

	copy(hash[:], hash_raw)
	x[Height] = hash
}

// verify whether the hash at the specific height matches
func IsCheckPointKnown_Static(actual_hash crypto.Hash, height uint64) (result bool) {
	switch globals.Config.Name {
	case "mainnet":
		if expected_hash, ok := mainnet_static_checkpoints[height]; ok {
			if actual_hash == expected_hash {
				result = true
				return
			}
		}

	case "testnet":
		if expected_hash, ok := testnet_static_checkpoints[height]; ok {

			if actual_hash == expected_hash {
				result = true
				return
			}
		}

	default:
		panic(fmt.Sprintf("Unknown Network \"%s\"", globals.Config.Name))
	}
	return
}

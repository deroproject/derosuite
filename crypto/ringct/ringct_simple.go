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

package ringct

import "fmt"

/* this files handles the generation and verification in ringct full */

// NOTE the transaction must have been expanded earlier and must have a key image, mixring etc
// this is implementation of verRctMG from rctSigs.cpp file
func (r *RctSig) VerifyRCTSimple_Core() (result bool) {
	result = false
	if r.sigType != RCTTypeSimple {
		if DEBUGGING_MODE {
			fmt.Printf("Signature NOT RingCT Simple  type, verification failed\n")
		}
		result = false
		return
	}

	pre_mlsag_hash := Key(Get_pre_mlsag_hash(r))

	// loop through all the inputs
	for inputi := 0; inputi < len(r.pseudoOuts); inputi++ {

		rows := 1
		cols := len(r.MixRing[inputi])

		if cols <= 2 {
			result = false
		}

		M := make([][]Key, cols) // lets create the double dimensional array
		for i := 0; i < cols; i++ {
			M[i] = make([]Key, rows+1, rows+1)
		}

		//create the matrix to mg sig
		for i := 0; i < cols; i++ {
			M[i][0] = r.MixRing[inputi][i].Destination
			SubKeys(&M[i][1], &r.MixRing[inputi][i].Mask, &r.pseudoOuts[inputi])
		}

		// do the mlsag verification
		result = MLSAG_Ver(pre_mlsag_hash, M, &r.MlsagSigs[inputi], rows, r)

		if result == false { // verification of 1 one vin failed mark, entire TX as failed
			if DEBUGGING_MODE {
				fmt.Printf("RCT Simple  signature verification failed for input %d\n", inputi)
			}

			return
		}

	}

	// we are here means everything went smoothly
	if DEBUGGING_MODE {
		fmt.Printf(" RCT Simple Signature successfully verified\n")
	}
	// result is already true so

	return
}

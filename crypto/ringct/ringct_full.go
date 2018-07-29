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
import "github.com/deroproject/derosuite/crypto"

/* this files handles the generation and verification in ringct full */

// NOTE the transaction must have been expanded earlier and must have a key image, mixring etc
// this is implementation of verRctMG from rctSigs.cpp file
func (r *RctSig) VerifyRCTFull_Core() (result bool) {
	result = false
	if r.sigType != RCTTypeFull {
		if DEBUGGING_MODE {
			fmt.Printf("Signature NOT RingCT MG type, verification failed\n")
		}
		result = false
		return
	}

	// some sanity checking
	/* if len(r.MixRing) != 1 { // this is hard code 1 for rct mg
	       if DEBUGGING_MODE {
	           fmt.Printf("RingCT MG  must have mixring rows 1\n")
	       }
	       result= false
	       return
	   }
	   if len(r.MixRing[0]) <= 1 { // mixing should be more than 1
	       if DEBUGGING_MODE {
	           fmt.Printf("RingCT MG  mixring  cannot be 1 or less\n")
	       }
	       result= false
	       return
	   }*/

	pre_mlsag_hash := crypto.Key(Get_pre_mlsag_hash(r))
	txfeekey := Commitment_From_Amount(r.txFee)

	cols := len(r.MixRing)
	rows := len(r.MixRing[0])

	//  fmt.Printf("cols %d rows %d \n", cols, rows)

	// if cols = 1 ,  if mixin = 5 , rows = 5
	// create a matrix of the form
	// 0  0
	// 1  1
	// 2  2
	// 3  3
	// 4  4
	// 5  5   // yes there is an extra row

	M := make([][]crypto.Key, cols)
	for i := 0; i < (cols); i++ {
		M[i] = make([]crypto.Key, rows+1, rows+1)
		for j := 0; j < (rows + 1); j++ { // yes there is an extra column
			M[i][j] = Identity // fill it with identity
			// fmt.Printf("M[%d][%d] %s\n",i,j, M[i][j])
		}
	}

	for j := 0; j < rows; j++ {
		for i := 0; i < cols; i++ {
			//fmt.Printf("j %d i %d \n", j,i)
			//   fmt.Printf("f j %d i %d  %s\n", j,i, M[i][j])
			//fmt.Printf("i %d rows %d \n", i, rows)
			M[i][j] = r.MixRing[i][j].Destination

			//    fmt.Printf("f M[i][rows] == %s\n",M[i][rows]);
			crypto.AddKeys(&M[i][rows], &M[i][rows], &r.MixRing[i][j].Mask) //add Ci in last row
			//    fmt.Printf("f M[i][rows] =  %s\n",M[i][rows]);
		}
	}

	for i := 0; i < cols; i++ {
		for j := 0; j < len(r.OutPk); j++ {
			crypto.SubKeys(&M[i][rows], &M[i][rows], &r.OutPk[j].Mask) //subtract output Ci's in last row
			//    fmt.Printf("s i %d j %d  %s \n",i,j,M[i][rows]);
		}
		//subtract txn fee output in last row
		crypto.SubKeys(&M[i][rows], &M[i][rows], &txfeekey)

		//  fmt.Printf("s M[i][rows] = %s\n",M[i][rows])
	}

	// do the mlsag verification

	result = MLSAG_Ver(pre_mlsag_hash, M, &r.MlsagSigs[0], rows, r)

	if DEBUGGING_MODE {
		if result {
			fmt.Printf("Signature Full successfully verified\n")
		} else {
			fmt.Printf("RCT MG  signarure verification failed\n")
		}

	}

	return
}

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

//import "io"
//import "fmt"
//import "crypto/rand"

import "github.com/deroproject/derosuite/crypto"

// bothe the function resturn identity of the ed25519 curve
func identity() (result *crypto.Key) {
	result = new(crypto.Key)
	result[0] = 1
	return
}

// convert a uint64 to a scalar
func d2h(val uint64) (result *crypto.Key) {
	result = new(crypto.Key)
	for i := 0; val > 0; i++ {
		result[i] = byte(val & 0xFF)
		val /= 256
	}
	return
}

//32 byte key to uint long long
// if the key holds a value > 2^64
// then the value in the first 8 bytes is returned
func h2d(input crypto.Key) (value uint64) {
	for j := 7; j >= 0; j-- {
		value = (value*256 + uint64(input[j]))
	}
	return value
}

// this gives you a commitment from an amount
// this is used to convert tx fee or miner tx amount to commitment
func Commitment_From_Amount(amount uint64) crypto.Key {
	return *(crypto.ScalarMultH(d2h(amount)))
}

// this is used to convert miner tx commitment to  mask
// equivalent to rctOps.cpp zeroCommit
func ZeroCommitment_From_Amount(amount uint64) crypto.Key {
	mask := *(identity())
	mask = crypto.ScalarmultBase(mask)
	am := d2h(amount)
	bH := crypto.ScalarMultH(am)
	crypto.AddKeys(&mask, &mask, bH)
	return mask
}

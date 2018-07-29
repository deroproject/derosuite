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

import "encoding/hex"
import "github.com/deroproject/derosuite/crypto"

// convert a hex string to a key
// a copy of these functions exists in the crypto package also
func HexToKey(h string) (result crypto.Key) {
	byteSlice, _ := hex.DecodeString(h)
	if len(byteSlice) != 32 {
		panic("Incorrect key size")
	}
	copy(result[:], byteSlice)
	return
}

func HexToHash(h string) (result crypto.Hash) {
	byteSlice, _ := hex.DecodeString(h)
	if len(byteSlice) != 32 {
		panic("Incorrect key size")
	}
	copy(result[:], byteSlice)
	return
}

// zero fill the key
func Sc_0(k *crypto.Key) {
	for i := 0; i < 32; i++ {
		k[i] = 0
	}
}

// RandomPubKey takes a random scalar, interprets it as a point on the curve
// and then multiplies by 8 to make it a point in the Group
//  remember the low order bug and do more auditing of the entire thing
func RandomPubKey() (result *crypto.Key) {
	result = new(crypto.Key)
	p3 := new(crypto.ExtendedGroupElement)
	var p1 crypto.ProjectiveGroupElement
	var p2 crypto.CompletedGroupElement
	h := crypto.RandomScalar()
	p1.FromBytes(h)
	crypto.GeMul8(&p2, &p1)
	p2.ToExtended(p3)
	p3.ToBytes(result)
	return
}

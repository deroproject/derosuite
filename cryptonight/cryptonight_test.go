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

package cryptonight

import "testing"
import "encoding/hex"

func Test_Cryptonight_Hash(t *testing.T) {

	// there are 4 sub algorithms, basically blake , groestl , jh , skein
	// other things are common, these tect vectors have been manually pulled from c codes

	blake_hash := cryptonight([]byte("This is a testi" + "\x01"))

	if hex.EncodeToString(blake_hash) != "7958d1afe0c46670642c0341f92e89bf6de6a2573ef89742237162e66ea4a121" {
		t.Error("Cryptonight blake_hash testing Failed\n")
		return

	}

	// this is from cryptonote whitepaper
	groestl_hash := cryptonight([]byte("This is a test" + "\x01"))

	if hex.EncodeToString(groestl_hash) != "a084f01d1437a09c6985401b60d43554ae105802c5f5d8a9b3253649c0be6605" {
		t.Error("Cryptonight testing Failed\n")
		return

	}

	jh_hash := cryptonight([]byte("This is a test2" + "\x01"))

	if hex.EncodeToString(jh_hash) != "6f93b51852d1a47277c62e720bf0e10bf90e92123847be246f67e3fd2639f4b4" {
		t.Error("Cryptonight testing Failed\n")
		return

	}

	skein_hash := cryptonight([]byte("This is a testw" + "\x01"))

	if hex.EncodeToString(skein_hash) != "3174ef437b24fd30e81d307d9b7d02ba21eb6f627cafc9d8134ea63adc4985b0" {
		t.Error("Cryptonight testing Failed\n")
		return

	}

}

// from fib_test.go
func BenchmarkCryptonight(b *testing.B) {
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		cryptonight([]byte("This is a testi" + "\x01"))
	}
}

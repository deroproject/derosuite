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

package crypto

import "testing"
import "encoding/hex"

func TestKeccak256(t *testing.T) {
	tests := []struct {
		name       string
		messageHex string
		wantHex    string
	}{
		{
			name:       "from monero 1",
			messageHex: "c8fedd380dbae40ffb52",
			wantHex:    "8e41962058b7422e7404253121489a3e63d186ed115086919a75105661483ba9",
		},
		{
			name:       "from monero 2",
			messageHex: "5020c4d530b6ec6cb4d9",
			wantHex:    "8a597f11961935e32e0adeab2ce48b3df2d907c9b26619dad22f42ff65ab7593",
		},
		{
			name:       "hello",
			messageHex: "68656c6c6f",
			wantHex:    "1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8",
		},
		{
			name:       "from monero cryptotest.pl",
			messageHex: "0f3fe9c20b24a11bf4d6d1acd335c6a80543f1f0380590d7323caf1390c78e88",
			wantHex:    "73b7a236f2a97c4e1805f7a319f1283e3276598567757186c526caf9a49e0a92",
		},
	}
	for _, test := range tests {
		message, _ := hex.DecodeString(test.messageHex)
		got := Keccak256(message)
		want := HexToHash(test.wantHex)
		if want != got {
			t.Errorf("want %x, got %x", want, got)
		}
	}
}

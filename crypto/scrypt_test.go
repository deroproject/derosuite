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

import "fmt"
import "testing"
import "encoding/hex"

// test cases from  litecoin for guranteed compatibility
func Test_Scrypt_1024_1_1_256(t *testing.T) {
	tests := []struct {
		data     string
		expected string
	}{
		{
			data:     "020000004c1271c211717198227392b029a64a7971931d351b387bb80db027f270411e398a07046f7d4a08dd815412a8712f874a7ebf0507e3878bd24e20a3b73fd750a667d2f451eac7471b00de6659",
			expected: "00000000002bef4107f882f6115e0b01f348d21195dacd3582aa2dabd7985806",
		},
		{
			data:     "0200000011503ee6a855e900c00cfdd98f5f55fffeaee9b6bf55bea9b852d9de2ce35828e204eef76acfd36949ae56d1fbe81c1ac9c0209e6331ad56414f9072506a77f8c6faf551eac7471b00389d01",
			expected: "00000000003a0d11bdd5eb634e08b7feddcfbbf228ed35d250daf19f1c88fc94",
		},
		{
			data:     "02000000a72c8a177f523946f42f22c3e86b8023221b4105e8007e59e81f6beb013e29aaf635295cb9ac966213fb56e046dc71df5b3f7f67ceaeab24038e743f883aff1aaafaf551eac7471b0166249b",
			expected: "00000000000b40f895f288e13244728a6c2d9d59d8aff29c65f8dd5114a8ca81",
		},
		{
			data:     "010000007824bc3a8a1b4628485eee3024abd8626721f7f870f8ad4d2f33a27155167f6a4009d1285049603888fe85a84b6c803a53305a8d497965a5e896e1a00568359589faf551eac7471b0065434e",
			expected: "00000000003007005891cd4923031e99d8e8d72f6e8e7edc6a86181897e105fe",
		},
		{
			data:     "0200000050bfd4e4a307a8cb6ef4aef69abc5c0f2d579648bd80d7733e1ccc3fbc90ed664a7f74006cb11bde87785f229ecd366c2d4e44432832580e0608c579e4cb76f383f7f551eac7471b00c36982",
			expected: "000000000018f0b426a4afc7130ccb47fa02af730d345b4fe7c7724d3800ec8c",
		},
	}

	for _, test := range tests {
		data, err := hex.DecodeString(test.data)
		if err != nil {
			t.Fatalf("Could NOT decode test")
		}

		actual := Scrypt_1024_1_1_256(data)

		// reverse it
		blen := len(actual) // its hardcoded 32 bytes, so why do len but lets do it
		for i := 0; i < blen/2; i++ {
			actual[i], actual[blen-1-i] = actual[blen-1-i], actual[i]
		}

		//t.Logf("cryptonightv7: want: %s, got: %x", test.expected, actual)
		if fmt.Sprintf("%s", actual) != test.expected {
			t.Fatalf("scrypt: want: %s, got: %s", test.expected, actual)
			continue
		}

	}

}

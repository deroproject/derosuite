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

//import "fmt"
import "testing"
import "bufio"
import "log"
import "os"
import "strings"
import "strconv"
import "encoding/hex"

// these tests, specifically "tests_data.txt" are available in the monero  project and used here to verify whether we implement everything

func Test_Crypto(t *testing.T) {

	file, err := os.Open("tests_data.txt")
	if err != nil {
		log.Fatalf("Test file tests_data is missing, err %s ", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// parse the line
		line := scanner.Text()
		words := strings.Fields(line)

		if len(words) < 2 {
			continue
		}
		switch words[0] {
		case "check_scalar":
			scalar := HexToKey(words[1])
			expected := "true" == words[2]
			actual := ScValid(&scalar)
			if actual != expected {
				t.Fatalf("Failed %s: Expected %v, got %v.", words[0], expected, actual)
			}
		case "check_key":
			public_key := HexToKey(words[1])
			expected := "true" == words[2]
			actual := public_key.Public_Key_Valid()
			if actual != expected {
				t.Fatalf("Failed %s: Expected %v, got %v %s", words[0], expected, actual, public_key)
			}

		case "random_scalar": // ignore them
		case "hash_to_scalar":
			data, _ := hex.DecodeString(words[1])
			expected := HexToKey(words[2])
			actual := HashToScalar(data)
			if *actual != expected {
				t.Fatalf("Failed %s: Expected %v, got %v.", words[0], expected, actual)
			}
			//t.Logf("executing %s\n", expected)
		case "generate_keys": // this test is meant to test RNG ??
			key_secret := HexToKey(words[2])
			key_public := HexToKey(words[1])

			if key_public != *(key_secret.PublicKey()) {
				t.Errorf("Failed %s key generation testing failed %s ", words[0], key_secret)
			}
		case "secret_key_to_public_key":
			key_secret := HexToKey(words[1])

			expected := "true" == words[2]
			actual := key_secret.Private_Key_Valid()

			if expected != actual {
				t.Fatalf("Failed %s: Expected %v, got %v.  %s", words[0], expected, actual, key_secret)
			}

			if actual { // test only if required
				key_public := HexToKey(words[3])
				if key_public != *(key_secret.PublicKey()) {
					t.Errorf("Failed %s key generation testing failed %s ", words[0], key_secret)
				}
			}

		case "generate_key_derivation":
			public_key := HexToKey(words[1])
			private_key := HexToKey(words[2])

			expected := "true" == words[3]
			actual := public_key.Public_Key_Valid()

			if expected != actual {
				t.Fatalf(" Failed %s: Expected %v, got %v.  %s", words[0], expected, actual, public_key)
			}

			if expected == true { // yes knowingly using the same variables
				expected := HexToKey(words[4])
				actual := KeyDerivation(&public_key, &private_key)

				if expected != actual {
					t.Fatalf("1Failed %s: Expected %v, got %v.  %s", words[0], expected, actual, public_key)
				}
			}

		case "derive_public_key":
			kd := HexToKey(words[1]) //
			outIdx, _ := strconv.ParseUint(words[2], 10, 0)
			base := HexToKey(words[3])
			var expected1, actual1 bool
			var expected2, actual2 Key
			expected1 = words[4] == "true"
			if expected1 {
				expected2 = HexToKey(words[5])
			}

			actual1 = base.Public_Key_Valid()
			if actual1 != expected1 {
				t.Fatalf("%s: Expected %v, got %v.", words[0], expected1, actual1)
			}

			if expected1 {
				actual2 = kd.KeyDerivation_To_PublicKey(outIdx, base)
				if actual2 != expected2 {
					t.Fatalf("%s: Expected %v, got %v.", words[0], expected2, actual2)
				}
			}
		case "derive_secret_key":
			kd := HexToKey(words[1]) //
			outIdx, _ := strconv.ParseUint(words[2], 10, 0)
			base := HexToKey(words[3])
			expected := HexToKey(words[4])

			actual := kd.KeyDerivation_To_PrivateKey(outIdx, base)
			if actual != expected {
				t.Fatalf("%s: Expected %v, got %v.", words[0], expected, actual)
			}

		case "hash_to_point": // this is different check than HashToPoint

			hash := HexToKey(words[1])
			expected := HexToKey(words[2])

			var actual Key
			var p1 ProjectiveGroupElement
			p1.FromBytes(&hash)
			p1.ToBytes(&actual)

			if actual != expected {
				t.Fatalf("%s: Expected %v, got %v.", words[0], expected, actual)
			}
		case "hash_to_ec":
			pub := HexToKey(words[1])
			expected := HexToKey(words[2])

			var actual Key
			ex := pub.HashToEC()
			ex.ToBytes(&actual)

			if actual != expected {
				t.Fatalf("%s: Expected %s, got %s.", words[0], expected, actual)
			}

		case "generate_key_image":
			public_key := HexToKey(words[1])
			private_key := HexToKey(words[2])
			expected := HexToKey(words[3])

			actual := GenerateKeyImage(public_key, private_key)
			if actual != expected {
				t.Fatalf("%s: Expected %s, got %s.", words[0], expected, actual)
			}

		// these are ignored because they are not required DERO project is based on ringct+
		case "generate_signature":
		case "check_signature":
		case "generate_ring_signature":
		case "check_ring_signature":

		default:

			t.Fatalf("This test is not handled %s: ", words[0])

		}
	}

}

// test whether H generation is alright
func TestH(t *testing.T) {
	G := ScalarmultBase(*(d2h(1)))
	//	t.Logf("G %s \nH %s", G, H)
	actual := G.HashToPointSimple()
	if actual != H {
		t.Fatalf("H generation failed Actual %s expected %s", actual, H)
	}
}

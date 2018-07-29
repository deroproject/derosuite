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
import "encoding/hex"

const (
	ChecksumLength = 4 // for addresses
	HashLength     = 32
)

type Hash [HashLength]byte
type Checksum [ChecksumLength]byte

func (h Hash) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("%x", h[:])), nil
}

func (h *Hash) UnmarshalText(data []byte) (err error) {
	byteSlice, _ := hex.DecodeString(string(data))
	if len(byteSlice) != 32 {
		return fmt.Errorf("Incorrect hash size")
	}
	copy(h[:], byteSlice)
	return
}

// stringifier
func (h Hash) String() string {
	return fmt.Sprintf("%x", h[:])
}

// convert a hash of hex form to binary form, returns a zero hash if any error
// TODO this should be in crypto
func HashHexToHash(hash_hex string) (hash Hash) {
	hash_raw, err := hex.DecodeString(hash_hex)

	if err != nil {
		//panic(fmt.Sprintf("Cannot hex decode checkpint hash \"%s\"", hash_hex))
		return hash
	}

	if len(hash_raw) != 32 {
		//panic(fmt.Sprintf(" hash not 32 byte size Cannot hex decode checkpint hash \"%s\"", hash_hex))
		return hash
	}

	copy(hash[:], hash_raw)
	return
}

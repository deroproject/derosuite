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

package address

import "fmt"
import "bytes"
import "encoding/binary"

import "github.com/deroproject/derosuite/crypto"

// see https://cryptonote.org/cns/cns007.txt to understand address more

type Address struct {
	Network  uint64
	SpendKey crypto.Key
	ViewKey  crypto.Key

	//TODO add support for integrated address
}

const ChecksumLength = 4

type Checksum [ChecksumLength]byte

func GetChecksum(data ...[]byte) (result Checksum) {
	keccak256 := crypto.Keccak256(data...)
	copy(result[:], keccak256[:4])
	return
}

func (a *Address) Base58() (result string) {

	prefix := make([]byte, 9, 9)

	n := binary.PutUvarint(prefix, a.Network)
	prefix = prefix[:n]

	checksum := GetChecksum(prefix, a.SpendKey[:], a.ViewKey[:])
	result = EncodeDeroBase58(prefix, a.SpendKey[:], a.ViewKey[:], checksum[:])
	return
}

// stringifier
func (a Address) String() string {
	return a.Base58()
}

func NewAddress(address string) (result *Address, err error) {
	raw := DecodeDeroBase58(address)

	// donot compare length to support much more user base and be compatible with cryptonote
	if len(raw) < 69 { // 1 byte prefix + 32 byte key + 32 byte key + 4 byte checksum
		err = fmt.Errorf("Address is the wrong length")
		return
	}

	checksum := GetChecksum(raw[:len(raw)-4])
	if bytes.Compare(checksum[:], raw[len(raw)-4:]) != 0 {
		err = fmt.Errorf("Checksum does not validate")
		return
	}

	// parse network first
	address_prefix, done := binary.Uvarint(raw)
	if done <= 0 {
		err = fmt.Errorf("Network could not be parsed in address\n")
		return
	}

	raw = raw[done:]

	result = &Address{
		Network: address_prefix,
		//SpendKey: raw[0:32],
		//ViewKey:  raw[32:64],
	}
	copy(result.SpendKey[:], raw[0:32])
	copy(result.ViewKey[:], raw[32:64])

	return
}

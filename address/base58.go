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

import "strings"
import "math/big"

// all characters in the base58
const BASE58 = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

var base58Lookup = map[string]int{
	"1": 0, "2": 1, "3": 2, "4": 3, "5": 4, "6": 5, "7": 6, "8": 7,
	"9": 8, "A": 9, "B": 10, "C": 11, "D": 12, "E": 13, "F": 14, "G": 15,
	"H": 16, "J": 17, "K": 18, "L": 19, "M": 20, "N": 21, "P": 22, "Q": 23,
	"R": 24, "S": 25, "T": 26, "U": 27, "V": 28, "W": 29, "X": 30, "Y": 31,
	"Z": 32, "a": 33, "b": 34, "c": 35, "d": 36, "e": 37, "f": 38, "g": 39,
	"h": 40, "i": 41, "j": 42, "k": 43, "m": 44, "n": 45, "o": 46, "p": 47,
	"q": 48, "r": 49, "s": 50, "t": 51, "u": 52, "v": 53, "w": 54, "x": 55,
	"y": 56, "z": 57,
}
var bigBase = big.NewInt(58)

// chunk are max 8 bytes long refer https://cryptonote.org/cns/cns007.txt for more documentation
var bytes_to_base58_length_mapping = []int{
	0,  // 0 bytes of input, 0 byte of base58 output
	2,  // 1 byte of input, 2 bytes of base58 output
	3,  // 2 byte of input, 3 bytes of base58 output
	5,  // 3 byte of input, 5 bytes of base58 output
	6,  // 4 byte of input, 6 bytes of base58 output
	7,  // 5 byte of input, 7 bytes of base58 output
	9,  // 6 byte of input, 9 bytes of base58 output
	10, // 7 byte of input, 10 bytes of base58 output
	11, // 8 byte of input, 11 bytes of base58 output
}

// encode 8 byte chunk with necessary padding
func encodeChunk(raw []byte) (result string) {
	remainder := new(big.Int)
	remainder.SetBytes(raw)
	bigZero := new(big.Int)
	for remainder.Cmp(bigZero) > 0 {
		current := new(big.Int)
		remainder.DivMod(remainder, bigBase, current)
		result = string(BASE58[current.Int64()]) + result
	}

	for i := range bytes_to_base58_length_mapping {
		if i == len(raw) {
			if len(result) < bytes_to_base58_length_mapping[i] {
				result = strings.Repeat("1", (bytes_to_base58_length_mapping[i]-len(result))) + result
			}
			return result
		}
	}

	return // we never reach here, if inputs are well-formed <= 8 bytes
}

// decode max 11 char base58 to 8 byte chunk as necessary
// proper error handling is not being done
func decodeChunk(encoded string) (result []byte) {
	bigResult := big.NewInt(0)
	currentMultiplier := big.NewInt(1)
	tmp := new(big.Int)
	for i := len(encoded) - 1; i >= 0; i-- {
		// make sure decoded character is a base58 char , otherwise return
		if strings.IndexAny(BASE58, string(encoded[i])) < 0 {
			return
		}
		tmp.SetInt64(int64(base58Lookup[string(encoded[i])]))
		tmp.Mul(currentMultiplier, tmp)
		bigResult.Add(bigResult, tmp)
		currentMultiplier.Mul(currentMultiplier, bigBase)
	}

	for i := range bytes_to_base58_length_mapping {
		if bytes_to_base58_length_mapping[i] == len(encoded) {
			result = append([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0}, bigResult.Bytes()...)
			return result[len(result)-i:] // return necessary bytes, initial zero appended  as per mapping
		}
	}

	return // we never reach here, if inputs are well-formed <= 11 chars
}

// split into 8 byte chunks, process and merge back result
func EncodeDeroBase58(data ...[]byte) (result string) {
	var combined []byte
	for _, item := range data {
		combined = append(combined, item...)
	}

	fullblocks := len(combined) / 8
	for i := 0; i < fullblocks; i++ { // process any chunks in 8 byte form
		result += encodeChunk(combined[i*8 : (i+1)*8])
	}
	if len(combined)%8 > 0 { // process last partial block
		result += encodeChunk(combined[fullblocks*8:])
	}
	return
}

// split into 11 char chunks, process and merge back result
func DecodeDeroBase58(data string) (result []byte) {
	fullblocks := len(data) / 11
	for i := 0; i < fullblocks; i++ { // process partial block
		result = append(result, decodeChunk(data[i*11:(i+1)*11])...)
	}
	if len(data)%11 > 0 { // process last partial block
		result = append(result, decodeChunk(data[fullblocks*11:])...)
	}
	return
}

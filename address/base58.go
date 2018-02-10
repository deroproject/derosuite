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

func encodeChunk(raw []byte, padding int) (result string) {
	remainder := new(big.Int)
	remainder.SetBytes(raw)
	bigZero := new(big.Int)
	for remainder.Cmp(bigZero) > 0 {
		current := new(big.Int)
		remainder.DivMod(remainder, bigBase, current)
		result = string(BASE58[current.Int64()]) + result
	}
	if len(result) < padding {
		result = strings.Repeat("1", (padding-len(result))) + result
	}
	return
}

func decodeChunk(encoded string) (result []byte) {
	bigResult := big.NewInt(0)
	currentMultiplier := big.NewInt(1)
	tmp := new(big.Int)
	for i := len(encoded) - 1; i >= 0; i-- {
		tmp.SetInt64(int64(base58Lookup[string(encoded[i])]))
		tmp.Mul(currentMultiplier, tmp)
		bigResult.Add(bigResult, tmp)
		currentMultiplier.Mul(currentMultiplier, bigBase)
	}
	result = bigResult.Bytes()
	return
}

func EncodeDeroBase58(data ...[]byte) (result string) {
	var combined []byte
	for _, item := range data {
		combined = append(combined, item...)
	}
	length := len(combined)
	rounds := length / 8
	for i := 0; i < rounds; i++ {
		result += encodeChunk(combined[i*8:(i+1)*8], 11)
	}
	if length%8 > 0 {
		result += encodeChunk(combined[rounds*8:], 7)
	}
	return
}

func DecodeDeroBase58(data string) (result []byte) {
	length := len(data)
	rounds := length / 11
	for i := 0; i < rounds; i++ {
		result = append(result, decodeChunk(data[i*11:(i+1)*11])...)
	}
	if length%11 > 0 {
		result = append(result, decodeChunk(data[rounds*11:])...)
	}
	return
}

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

// TODO keccak is available in 3 of our packages we need to merge them when the alpha version is out
// Package keccak implements the Keccak (SHA-3) hash algorithm.
// http://keccak.noekeon.org / FIPS 202 draft.
package cryptonight

// import "fmt"
import "hash"

const (
	domainNone  = 1
	domainSHA3  = 0x06
	domainSHAKE = 0x1f
)

const rounds = 24

var roundConstants = []uint64{
	0x0000000000000001, 0x0000000000008082,
	0x800000000000808A, 0x8000000080008000,
	0x000000000000808B, 0x0000000080000001,
	0x8000000080008081, 0x8000000000008009,
	0x000000000000008A, 0x0000000000000088,
	0x0000000080008009, 0x000000008000000A,
	0x000000008000808B, 0x800000000000008B,
	0x8000000000008089, 0x8000000000008003,
	0x8000000000008002, 0x8000000000000080,
	0x000000000000800A, 0x800000008000000A,
	0x8000000080008081, 0x8000000000008080,
	0x0000000080000001, 0x8000000080008008,
}

type keccak struct {
	S         [25]uint64
	size      int
	blockSize int
	buf       []byte
	domain    byte
}

func newKeccak(capacity, output int, domain byte) hash.Hash {
	var h keccak
	h.size = output / 8
	h.blockSize = (200 - capacity/8)
	h.domain = domain
	return &h
}

func New224() hash.Hash {
	return newKeccak(224*2, 224, domainNone)
}

func New256() hash.Hash {
	return newKeccak(256*2, 256, domainNone)
}

func New384() hash.Hash {
	return newKeccak(384*2, 384, domainNone)
}

func New512() hash.Hash {
	return newKeccak(512*2, 512, domainNone)
}

func (k *keccak) Write(b []byte) (int, error) {
	n := len(b)

	if len(k.buf) > 0 {
		x := k.blockSize - len(k.buf)
		if x > len(b) {
			x = len(b)
		}
		k.buf = append(k.buf, b[:x]...)
		b = b[x:]

		if len(k.buf) < k.blockSize {
			return n, nil
		}

		k.absorb(k.buf)
		k.buf = nil
	}

	for len(b) >= k.blockSize {
		k.absorb(b[:k.blockSize])
		b = b[k.blockSize:]
	}

	k.buf = b

	return n, nil
}

func (k0 *keccak) Sum(b []byte) []byte {
	k := *k0
	k.final()
	return k.squeeze(b)
}

func (k *keccak) Reset() {
	for i := range k.S {
		k.S[i] = 0
	}
	k.buf = nil
}

func (k *keccak) Size() int {
	return k.size
}

func (k *keccak) BlockSize() int {
	return k.blockSize
}

func (k *keccak) absorb(block []byte) {
	if len(block) != k.blockSize {
		panic("absorb() called with invalid block size")
	}

	for i := 0; i < k.blockSize/8; i++ {
		k.S[i] ^= uint64le(block[i*8:])
	}
	keccakf(&k.S)
}

func (k *keccak) pad(block []byte) []byte {

	padded := make([]byte, k.blockSize)

	copy(padded, k.buf)
	padded[len(k.buf)] = k.domain
	padded[len(padded)-1] |= 0x80

	return padded
}

func (k *keccak) final() {
	last := k.pad(k.buf)
	k.absorb(last)
}

func (k *keccak) squeeze(b []byte) []byte {
	buf := make([]byte, 8*len(k.S))
	n := k.size
	for {
		for i := range k.S {
			putUint64le(buf[i*8:], k.S[i])
		}
		if n <= k.blockSize {
			b = append(b, buf[:n]...)
			break
		}
		b = append(b, buf[:k.blockSize]...)
		n -= k.blockSize
		keccakf(&k.S)
	}
	return b
}

func keccakf(S *[25]uint64) {
	var bc0, bc1, bc2, bc3, bc4 uint64
	var S0, S1, S2, S3, S4 uint64
	var S5, S6, S7, S8, S9 uint64
	var S10, S11, S12, S13, S14 uint64
	var S15, S16, S17, S18, S19 uint64
	var S20, S21, S22, S23, S24 uint64
	var tmp uint64

	//S[0]=0x2073692073696854;
	//S[1]=0x1747365742061
	// S[16]=0x8000000000000000;

	/* for i :=0; i< 25;i++{
	    fmt.Printf("%2d %X\n", i, S[i])
	   }
	*/

	S0, S1, S2, S3, S4 = S[0], S[1], S[2], S[3], S[4]
	S5, S6, S7, S8, S9 = S[5], S[6], S[7], S[8], S[9]
	S10, S11, S12, S13, S14 = S[10], S[11], S[12], S[13], S[14]
	S15, S16, S17, S18, S19 = S[15], S[16], S[17], S[18], S[19]
	S20, S21, S22, S23, S24 = S[20], S[21], S[22], S[23], S[24]

	for r := 0; r < rounds; r++ {
		// theta
		bc0 = S0 ^ S5 ^ S10 ^ S15 ^ S20
		bc1 = S1 ^ S6 ^ S11 ^ S16 ^ S21
		bc2 = S2 ^ S7 ^ S12 ^ S17 ^ S22
		bc3 = S3 ^ S8 ^ S13 ^ S18 ^ S23
		bc4 = S4 ^ S9 ^ S14 ^ S19 ^ S24
		tmp = bc4 ^ (bc1<<1 | bc1>>(64-1))
		S0 ^= tmp
		S5 ^= tmp
		S10 ^= tmp
		S15 ^= tmp
		S20 ^= tmp
		tmp = bc0 ^ (bc2<<1 | bc2>>(64-1))
		S1 ^= tmp
		S6 ^= tmp
		S11 ^= tmp
		S16 ^= tmp
		S21 ^= tmp
		tmp = bc1 ^ (bc3<<1 | bc3>>(64-1))
		S2 ^= tmp
		S7 ^= tmp
		S12 ^= tmp
		S17 ^= tmp
		S22 ^= tmp
		tmp = bc2 ^ (bc4<<1 | bc4>>(64-1))
		S3 ^= tmp
		S8 ^= tmp
		S13 ^= tmp
		S18 ^= tmp
		S23 ^= tmp
		tmp = bc3 ^ (bc0<<1 | bc0>>(64-1))
		S4 ^= tmp
		S9 ^= tmp
		S14 ^= tmp
		S19 ^= tmp
		S24 ^= tmp

		// rho phi
		tmp = S1
		tmp, S10 = S10, tmp<<1|tmp>>(64-1)
		tmp, S7 = S7, tmp<<3|tmp>>(64-3)
		tmp, S11 = S11, tmp<<6|tmp>>(64-6)
		tmp, S17 = S17, tmp<<10|tmp>>(64-10)
		tmp, S18 = S18, tmp<<15|tmp>>(64-15)
		tmp, S3 = S3, tmp<<21|tmp>>(64-21)
		tmp, S5 = S5, tmp<<28|tmp>>(64-28)
		tmp, S16 = S16, tmp<<36|tmp>>(64-36)
		tmp, S8 = S8, tmp<<45|tmp>>(64-45)
		tmp, S21 = S21, tmp<<55|tmp>>(64-55)
		tmp, S24 = S24, tmp<<2|tmp>>(64-2)
		tmp, S4 = S4, tmp<<14|tmp>>(64-14)
		tmp, S15 = S15, tmp<<27|tmp>>(64-27)
		tmp, S23 = S23, tmp<<41|tmp>>(64-41)
		tmp, S19 = S19, tmp<<56|tmp>>(64-56)
		tmp, S13 = S13, tmp<<8|tmp>>(64-8)
		tmp, S12 = S12, tmp<<25|tmp>>(64-25)
		tmp, S2 = S2, tmp<<43|tmp>>(64-43)
		tmp, S20 = S20, tmp<<62|tmp>>(64-62)
		tmp, S14 = S14, tmp<<18|tmp>>(64-18)
		tmp, S22 = S22, tmp<<39|tmp>>(64-39)
		tmp, S9 = S9, tmp<<61|tmp>>(64-61)
		tmp, S6 = S6, tmp<<20|tmp>>(64-20)
		S1 = tmp<<44 | tmp>>(64-44)

		// chi
		bc0 = S0
		bc1 = S1
		bc2 = S2
		bc3 = S3
		bc4 = S4
		S0 ^= (^bc1) & bc2
		S1 ^= (^bc2) & bc3
		S2 ^= (^bc3) & bc4
		S3 ^= (^bc4) & bc0
		S4 ^= (^bc0) & bc1
		bc0 = S5
		bc1 = S6
		bc2 = S7
		bc3 = S8
		bc4 = S9
		S5 ^= (^bc1) & bc2
		S6 ^= (^bc2) & bc3
		S7 ^= (^bc3) & bc4
		S8 ^= (^bc4) & bc0
		S9 ^= (^bc0) & bc1
		bc0 = S10
		bc1 = S11
		bc2 = S12
		bc3 = S13
		bc4 = S14
		S10 ^= (^bc1) & bc2
		S11 ^= (^bc2) & bc3
		S12 ^= (^bc3) & bc4
		S13 ^= (^bc4) & bc0
		S14 ^= (^bc0) & bc1
		bc0 = S15
		bc1 = S16
		bc2 = S17
		bc3 = S18
		bc4 = S19
		S15 ^= (^bc1) & bc2
		S16 ^= (^bc2) & bc3
		S17 ^= (^bc3) & bc4
		S18 ^= (^bc4) & bc0
		S19 ^= (^bc0) & bc1
		bc0 = S20
		bc1 = S21
		bc2 = S22
		bc3 = S23
		bc4 = S24
		S20 ^= (^bc1) & bc2
		S21 ^= (^bc2) & bc3
		S22 ^= (^bc3) & bc4
		S23 ^= (^bc4) & bc0
		S24 ^= (^bc0) & bc1

		// iota
		S0 ^= roundConstants[r]
	}

	S[0], S[1], S[2], S[3], S[4] = S0, S1, S2, S3, S4
	S[5], S[6], S[7], S[8], S[9] = S5, S6, S7, S8, S9
	S[10], S[11], S[12], S[13], S[14] = S10, S11, S12, S13, S14
	S[15], S[16], S[17], S[18], S[19] = S15, S16, S17, S18, S19
	S[20], S[21], S[22], S[23], S[24] = S20, S21, S22, S23, S24

}

func uint64le(v []byte) uint64 {
	return uint64(v[0]) |
		uint64(v[1])<<8 |
		uint64(v[2])<<16 |
		uint64(v[3])<<24 |
		uint64(v[4])<<32 |
		uint64(v[5])<<40 |
		uint64(v[6])<<48 |
		uint64(v[7])<<56

}

func putUint64le(v []byte, x uint64) {
	v[0] = byte(x)
	v[1] = byte(x >> 8)
	v[2] = byte(x >> 16)
	v[3] = byte(x >> 24)
	v[4] = byte(x >> 32)
	v[5] = byte(x >> 40)
	v[6] = byte(x >> 48)
	v[7] = byte(x >> 56)
}

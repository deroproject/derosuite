// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package simd

import (
	"fmt"

	"gitlab.com/nitya-sattva/go-x11/hash"
)

// HashSize holds the size of a hash in bytes.
const HashSize = int(64)

// BlockSize holds the size of a block in bytes.
const BlockSize = uintptr(128)

////////////////

type digest struct {
	ptr uintptr
	ch  uint32
	cl  uint32

	h [32]uint32

	b [BlockSize]byte
}

// New returns a new digest to compute a SIMD512 hash.
func New() hash.Digest {
	ref := &digest{}
	ref.Reset()
	return ref
}

////////////////

// Reset resets the digest to its initial state.
func (ref *digest) Reset() {
	ref.ptr = 0
	ref.cl, ref.ch = 0, 0
	copy(ref.h[:], kInit[:])
}

// Sum appends the current hash to dst and returns the result
// as a slice. It does not change the underlying hash state.
func (ref *digest) Sum(dst []byte) []byte {
	dgt := *ref
	hsh := [64]byte{}
	dgt.Close(hsh[:], 0, 0)
	return append(dst, hsh[:]...)
}

// Write more data to the running hash, never returns an error.
func (ref *digest) Write(src []byte) (int, error) {
	sln := uintptr(len(src))
	fln := len(src)

	for sln > 0 {
		cln := BlockSize - ref.ptr

		if cln > sln {
			cln = sln
		}
		sln -= cln

		copy(ref.b[ref.ptr:], src[:cln])
		src = src[cln:]

		ref.ptr += cln
		if ref.ptr == BlockSize {
			ref.compress(0)
			ref.ptr = 0

			ref.cl += 1
			if ref.cl == 0 {
				ref.ch++
			}
		}
	}

	return fln, nil
}

// Close the digest by writing the last bits and storing the hash
// in dst. This prepares the digest for reuse by calling reset. A call
// to Close with a dst that is smaller then HashSize will return an error.
func (ref *digest) Close(dst []byte, bits uint8, bcnt uint8) error {
	if ln := len(dst); HashSize > ln {
		return fmt.Errorf("Simd Close: dst min length: %d, got %d", HashSize, ln)
	}

	if ref.ptr > 0 || bcnt > 0 {
		memset(ref.b[ref.ptr:], 0)
		ref.b[ref.ptr] = uint8(bits & (0xFF << (8 - bcnt)))
		ref.compress(0)
	}

	memset(ref.b[:], 0)
	{
		low := uint32(ref.cl << 10)
		low += uint32(ref.ptr<<3) + uint32(bcnt)
		high := uint32(ref.ch<<10) + (ref.cl >> 22)
		encUInt32le(ref.b[:], low)
		encUInt32le(ref.b[4:], high)
	}
	ref.compress(1)

	for u := int(0); u < 16; u++ {
		encUInt32le(dst[(u<<2):], ref.h[u])
	}

	ref.Reset()
	return nil
}

// Size returns the number of bytes required to store the hash.
func (*digest) Size() int {
	return HashSize
}

// BlockSize returns the block size of the hash.
func (*digest) BlockSize() int {
	return int(BlockSize)
}

////////////////

func memset(dst []byte, src byte) {
	for i := range dst {
		dst[i] = src
	}
}

func decUInt32le(src []byte) uint32 {
	return (uint32(src[0]) |
		uint32(src[1])<<8 |
		uint32(src[2])<<16 |
		uint32(src[3])<<24)
}

func encUInt32le(dst []uint8, src uint32) {
	dst[0] = uint8(src)
	dst[1] = uint8(src >> 8)
	dst[2] = uint8(src >> 16)
	dst[3] = uint8(src >> 24)
}

func (ref *digest) compress(last int) {
	var q [256]int32
	var w [64]uint32
	var st [32]uint32

	mixoutRound(ref.b[:], q[:], 1<<2)
	mixoutRound(ref.b[2:], q[64:], 1<<2)

	{
		var t int32
		var u, v uintptr

		m := q[0]
		n := q[64]
		q[0] = m + n
		q[64] = m - n

		m = q[u+1]
		n = q[u+1+64]
		t = (n * kAlphaTab[v+1*2])
		t = ((t) & 0xFFFF) + ((t) >> 16)
		q[u+1] = m + t
		q[u+1+64] = m - t
		m = q[u+2]
		n = q[u+2+64]
		t = (n * kAlphaTab[v+2*2])
		t = ((t) & 0xFFFF) + ((t) >> 16)
		q[u+2] = m + t
		q[u+2+64] = m - t
		m = q[u+3]
		n = q[u+3+64]
		t = (n * kAlphaTab[v+3*2])
		t = ((t) & 0xFFFF) + ((t) >> 16)
		q[u+3] = m + t
		q[u+3+64] = m - t

		u = 4
		v = 4 * 2
		for u < 64 {
			m = q[u]
			n = q[u+(64)]
			t = (n * kAlphaTab[v+0*2])
			t = ((t) & 0xFFFF) + ((t) >> 16)
			q[u] = m + t
			q[u+(64)] = m - t

			m = q[u+1]
			n = q[u+1+64]
			t = (n * kAlphaTab[v+1*2])
			t = ((t) & 0xFFFF) + ((t) >> 16)
			q[u+1] = m + t
			q[u+1+64] = m - t
			m = q[u+2]
			n = q[u+2+64]
			t = (n * kAlphaTab[v+2*2])
			t = ((t) & 0xFFFF) + ((t) >> 16)
			q[u+2] = m + t
			q[u+2+64] = m - t
			m = q[u+3]
			n = q[u+3+64]
			t = (n * kAlphaTab[v+3*2])
			t = ((t) & 0xFFFF) + ((t) >> 16)
			q[u+3] = m + t
			q[u+3+64] = m - t

			u += 4
			v += 4 * 2
		}
	}

	mixoutRound(ref.b[1:], q[128:], 1<<2)
	mixoutRound(ref.b[3:], q[192:], 1<<2)

	{
		var t int32
		var u, v uintptr

		m := q[128]
		n := q[128+64]
		q[128] = m + n
		q[128+64] = m - n

		m = q[128+u+1]
		n = q[128+u+1+64]
		t = (n * kAlphaTab[v+1*2])
		t = ((t) & 0xFFFF) + ((t) >> 16)
		q[128+u+1] = m + t
		q[128+u+1+64] = m - t
		m = q[128+u+2]
		n = q[128+u+2+64]
		t = (n * kAlphaTab[v+2*2])
		t = ((t) & 0xFFFF) + ((t) >> 16)
		q[128+u+2] = m + t
		q[128+u+2+64] = m - t
		m = q[128+u+3]
		n = q[128+u+3+64]
		t = (n * kAlphaTab[v+3*2])
		t = ((t) & 0xFFFF) + ((t) >> 16)
		q[128+u+3] = m + t
		q[128+u+3+64] = m - t

		u = 4
		v = 4 * 2
		for u < 64 {
			m = q[128+u]
			n = q[128+u+64]
			t = (n * kAlphaTab[v+0*2])
			t = ((t) & 0xFFFF) + ((t) >> 16)
			q[128+u] = m + t
			q[128+u+64] = m - t

			m = q[128+u+1]
			n = q[128+u+1+64]
			t = (n * kAlphaTab[v+1*2])
			t = ((t) & 0xFFFF) + ((t) >> 16)
			q[128+u+1] = m + t
			q[128+u+1+64] = m - t
			m = q[128+u+2]
			n = q[128+u+2+64]
			t = (n * kAlphaTab[v+2*2])
			t = ((t) & 0xFFFF) + ((t) >> 16)
			q[128+u+2] = m + t
			q[128+u+2+64] = m - t
			m = q[128+u+3]
			n = q[128+u+3+64]
			t = (n * kAlphaTab[v+3*2])
			t = ((t) & 0xFFFF) + ((t) >> 16)
			q[128+u+3] = m + t
			q[128+u+3+64] = m - t

			u += 4
			v += 4 * 2
		}
	}

	{
		var t int32
		var u, v uintptr

		m := q[0]
		n := q[128]
		q[0] = m + n
		q[128] = m - n

		m = q[u+1]
		n = q[u+1+128]
		t = (n * kAlphaTab[v+1])
		t = ((t) & 0xFFFF) + ((t) >> 16)
		q[u+1] = m + t
		q[u+1+128] = m - t
		m = q[u+2]
		n = q[u+2+128]
		t = (n * kAlphaTab[v+2])
		t = ((t) & 0xFFFF) + ((t) >> 16)
		q[u+2] = m + t
		q[u+2+128] = m - t
		m = q[u+3]
		n = q[u+3+128]
		t = (n * kAlphaTab[v+3])
		t = ((t) & 0xFFFF) + ((t) >> 16)
		q[u+3] = m + t
		q[u+3+128] = m - t

		u = 4
		v = 4
		for u < 128 {
			m = q[u]
			n = q[u+128]
			t = (n * kAlphaTab[v+0])
			t = ((t) & 0xFFFF) + ((t) >> 16)
			q[u] = m + t
			q[u+128] = m - t

			m = q[u+1]
			n = q[u+1+128]
			t = (n * kAlphaTab[v+1])
			t = ((t) & 0xFFFF) + ((t) >> 16)
			q[u+1] = m + t
			q[u+1+128] = m - t
			m = q[u+2]
			n = q[u+2+128]
			t = (n * kAlphaTab[v+2])
			t = ((t) & 0xFFFF) + ((t) >> 16)
			q[u+2] = m + t
			q[u+2+128] = m - t
			m = q[u+3]
			n = q[u+3+128]
			t = (n * kAlphaTab[v+3])
			t = ((t) & 0xFFFF) + ((t) >> 16)
			q[u+3] = m + t
			q[u+3+128] = m - t

			u += 4
			v += 4
		}
	}

	if last == 1 {
		var tq int32
		for i := uintptr(0); i < 256; i++ {
			tq = q[i] + kYOffB[i]
			tq = (((tq) & 0xFFFF) + ((tq) >> 16))
			tq = (((tq) & 0xFF) - ((tq) >> 8))
			tq = (((tq) & 0xFF) - ((tq) >> 8))
			if tq <= 128 {
				q[i] = tq
			} else {
				q[i] = tq - 257
			}
		}
	} else {
		var tq int32
		for i := uintptr(0); i < 256; i++ {
			tq = q[i] + kYOffA[i]
			tq = (((tq) & 0xFFFF) + ((tq) >> 16))
			tq = (((tq) & 0xFF) - ((tq) >> 8))
			tq = (((tq) & 0xFF) - ((tq) >> 8))
			if tq <= 128 {
				q[i] = tq
			} else {
				q[i] = tq - 257
			}
		}
	}

	{
		b := ref.b[:]
		s := ref.h[:]
		for i := uintptr(0); i < 32; i += 8 {
			st[i+0] = s[i+0] ^ decUInt32le(b[4*(i+0):])
			st[i+1] = s[i+1] ^ decUInt32le(b[4*(i+1):])
			st[i+2] = s[i+2] ^ decUInt32le(b[4*(i+2):])
			st[i+3] = s[i+3] ^ decUInt32le(b[4*(i+3):])
			st[i+4] = s[i+4] ^ decUInt32le(b[4*(i+4):])
			st[i+5] = s[i+5] ^ decUInt32le(b[4*(i+5):])
			st[i+6] = s[i+6] ^ decUInt32le(b[4*(i+6):])
			st[i+7] = s[i+7] ^ decUInt32le(b[4*(i+7):])
		}
	}

	for u := uintptr(0); u < 64; u += 8 {
		v := uintptr(wbp[(u >> 3)])

		w[u+0] = ((uint32(q[v+2*0]*185) & uint32(0xFFFF)) +
			(uint32(q[v+2*0+1]*185) << 16))
		w[u+1] = ((uint32(q[v+2*1]*185) & uint32(0xFFFF)) +
			(uint32(q[v+2*1+1]*185) << 16))
		w[u+2] = ((uint32(q[v+2*2]*185) & uint32(0xFFFF)) +
			(uint32(q[v+2*2+1]*185) << 16))
		w[u+3] = ((uint32(q[v+2*3]*185) & uint32(0xFFFF)) +
			(uint32(q[v+2*3+1]*185) << 16))
		w[u+4] = ((uint32(q[v+2*4]*185) & uint32(0xFFFF)) +
			(uint32(q[v+2*4+1]*185) << 16))
		w[u+5] = ((uint32(q[v+2*5]*185) & uint32(0xFFFF)) +
			(uint32(q[v+2*5+1]*185) << 16))
		w[u+6] = ((uint32(q[v+2*6]*185) & uint32(0xFFFF)) +
			(uint32(q[v+2*6+1]*185) << 16))
		w[u+7] = ((uint32(q[v+2*7]*185) & uint32(0xFFFF)) +
			(uint32(q[v+2*7+1]*185) << 16))
	}

	mixinRound(st[:], w[:], 0, 3, 23, 17, 27)

	for u := uintptr(0); u < 64; u += 8 {
		v := uintptr(wbp[(u>>3)+8])

		w[u+0] = (uint32(q[v+2*0]*185) & uint32(0xFFFF)) +
			(uint32(q[v+2*0+1]*185) << 16)
		w[u+1] = (uint32(q[v+2*1]*185) & uint32(0xFFFF)) +
			(uint32(q[v+2*1+1]*185) << 16)
		w[u+2] = (uint32(q[v+2*2]*185) & uint32(0xFFFF)) +
			(uint32(q[v+2*2+1]*185) << 16)
		w[u+3] = (uint32(q[v+2*3]*185) & uint32(0xFFFF)) +
			(uint32(q[v+2*3+1]*185) << 16)
		w[u+4] = (uint32(q[v+2*4]*185) & uint32(0xFFFF)) +
			(uint32(q[v+2*4+1]*185) << 16)
		w[u+5] = (uint32(q[v+2*5]*185) & uint32(0xFFFF)) +
			(uint32(q[v+2*5+1]*185) << 16)
		w[u+6] = (uint32(q[v+2*6]*185) & uint32(0xFFFF)) +
			(uint32(q[v+2*6+1]*185) << 16)
		w[u+7] = (uint32(q[v+2*7]*185) & uint32(0xFFFF)) +
			(uint32(q[v+2*7+1]*185) << 16)
	}
	mixinRound(st[:], w[:], 1, 28, 19, 22, 7)

	for u := uintptr(0); u < 64; u += 8 {
		v := uintptr(wbp[(u>>3)+16])

		w[u+0] = ((uint32(q[v+2*0-256]*(233)) & uint32(0xFFFF)) +
			(uint32((q[v+2*0-128])*(233)) << 16))
		w[u+1] = ((uint32(q[v+2*1-256]*(233)) & uint32(0xFFFF)) +
			(uint32((q[v+2*1-128])*(233)) << 16))
		w[u+2] = ((uint32(q[v+2*2-256]*(233)) & uint32(0xFFFF)) +
			(uint32((q[v+2*2-128])*(233)) << 16))
		w[u+3] = ((uint32(q[v+2*3-256]*(233)) & uint32(0xFFFF)) +
			(uint32((q[v+2*3-128])*(233)) << 16))
		w[u+4] = ((uint32(q[v+2*4-256]*(233)) & uint32(0xFFFF)) +
			(uint32((q[v+2*4-128])*(233)) << 16))
		w[u+5] = ((uint32(q[v+2*5-256]*(233)) & uint32(0xFFFF)) +
			(uint32((q[v+2*5-128])*(233)) << 16))
		w[u+6] = ((uint32(q[v+2*6-256]*(233)) & uint32(0xFFFF)) +
			(uint32((q[v+2*6-128])*(233)) << 16))
		w[u+7] = ((uint32(q[v+2*7-256]*(233)) & uint32(0xFFFF)) +
			(uint32((q[v+2*7-128])*(233)) << 16))
	}
	mixinRound(st[:], w[:], 2, 29, 9, 15, 5)

	for u := uintptr(0); u < 64; u += 8 {
		v := uintptr(wbp[(u>>3)+24])

		w[u+0] = ((uint32(q[v+2*0-383]*233) & uint32(0xFFFF)) +
			(uint32(q[v+2*0-255]*233) << 16))
		w[u+1] = ((uint32(q[v+2*1-383]*233) & uint32(0xFFFF)) +
			(uint32(q[v+2*1-255]*233) << 16))
		w[u+2] = ((uint32(q[v+2*2-383]*233) & uint32(0xFFFF)) +
			(uint32(q[v+2*2-255]*233) << 16))
		w[u+3] = ((uint32(q[v+2*3-383]*233) & uint32(0xFFFF)) +
			(uint32(q[v+2*3-255]*233) << 16))
		w[u+4] = ((uint32(q[v+2*4-383]*233) & uint32(0xFFFF)) +
			(uint32(q[v+2*4-255]*233) << 16))
		w[u+5] = ((uint32(q[v+2*5-383]*233) & uint32(0xFFFF)) +
			(uint32(q[v+2*5-255]*233) << 16))
		w[u+6] = ((uint32(q[v+2*6-383]*233) & uint32(0xFFFF)) +
			(uint32(q[v+2*6-255]*233) << 16))
		w[u+7] = ((uint32(q[v+2*7-383]*233) & uint32(0xFFFF)) +
			(uint32(q[v+2*7-255]*233) << 16))
	}
	mixinRound(st[:], w[:], 3, 4, 13, 10, 25)

	{
		var tp uint32
		var tA [8]uint32

		sta := ref.h[:]

		tA[0] = ((st[0] << 4) | (st[0] >> (32 - 4)))
		tA[1] = ((st[1] << 4) | (st[1] >> (32 - 4)))
		tA[2] = ((st[2] << 4) | (st[2] >> (32 - 4)))
		tA[3] = ((st[3] << 4) | (st[3] >> (32 - 4)))
		tA[4] = ((st[4] << 4) | (st[4] >> (32 - 4)))
		tA[5] = ((st[5] << 4) | (st[5] >> (32 - 4)))
		tA[6] = ((st[6] << 4) | (st[6] >> (32 - 4)))
		tA[7] = ((st[7] << 4) | (st[7] >> (32 - 4)))

		tp = uint32(st[kIdxD[0]] + sta[0] +
			(((st[kIdxB[0]] ^ st[kIdxC[0]]) & st[kIdxA[0]]) ^ st[kIdxC[0]]))
		st[kIdxA[0]] = ((tp << 13) | (tp >> (32 - 13))) + tA[kPrem[4][0]]
		st[kIdxD[0]] = st[kIdxC[0]]
		st[kIdxC[0]] = st[kIdxB[0]]
		st[kIdxB[0]] = tA[0]

		tp = uint32(st[kIdxD[1]] + sta[1] +
			(((st[kIdxB[1]] ^ st[kIdxC[1]]) & st[kIdxA[1]]) ^ st[kIdxC[1]]))
		st[kIdxA[1]] = ((tp << 13) | (tp >> (32 - 13))) + tA[kPrem[4][1]]
		st[kIdxD[1]] = st[kIdxC[1]]
		st[kIdxC[1]] = st[kIdxB[1]]
		st[kIdxB[1]] = tA[1]

		tp = uint32(st[kIdxD[2]] + sta[2] +
			(((st[kIdxB[2]] ^ st[kIdxC[2]]) & st[kIdxA[2]]) ^ st[kIdxC[2]]))
		st[kIdxA[2]] = ((tp << 13) | (tp >> (32 - 13))) + tA[kPrem[4][2]]
		st[kIdxD[2]] = st[kIdxC[2]]
		st[kIdxC[2]] = st[kIdxB[2]]
		st[kIdxB[2]] = tA[2]

		tp = uint32(st[kIdxD[3]] + sta[3] +
			(((st[kIdxB[3]] ^ st[kIdxC[3]]) & st[kIdxA[3]]) ^ st[kIdxC[3]]))
		st[kIdxA[3]] = ((tp << 13) | (tp >> (32 - 13))) + tA[kPrem[4][3]]
		st[kIdxD[3]] = st[kIdxC[3]]
		st[kIdxC[3]] = st[kIdxB[3]]
		st[kIdxB[3]] = tA[3]

		tp = uint32(st[kIdxD[4]] + sta[4] +
			(((st[kIdxB[4]] ^ st[kIdxC[4]]) & st[kIdxA[4]]) ^ st[kIdxC[4]]))
		st[kIdxA[4]] = ((tp << 13) | (tp >> (32 - 13))) + tA[kPrem[4][4]]
		st[kIdxD[4]] = st[kIdxC[4]]
		st[kIdxC[4]] = st[kIdxB[4]]
		st[kIdxB[4]] = tA[4]

		tp = uint32(st[kIdxD[5]] + sta[5] +
			(((st[kIdxB[5]] ^ st[kIdxC[5]]) & st[kIdxA[5]]) ^ st[kIdxC[5]]))
		st[kIdxA[5]] = ((tp << 13) | (tp >> (32 - 13))) + tA[kPrem[4][5]]
		st[kIdxD[5]] = st[kIdxC[5]]
		st[kIdxC[5]] = st[kIdxB[5]]
		st[kIdxB[5]] = tA[5]

		tp = uint32(st[kIdxD[6]] + sta[6] +
			(((st[kIdxB[6]] ^ st[kIdxC[6]]) & st[kIdxA[6]]) ^ st[kIdxC[6]]))
		st[kIdxA[6]] = ((tp << 13) | (tp >> (32 - 13))) + tA[kPrem[4][6]]
		st[kIdxD[6]] = st[kIdxC[6]]
		st[kIdxC[6]] = st[kIdxB[6]]
		st[kIdxB[6]] = tA[6]

		tp = uint32(st[kIdxD[7]] + sta[7] +
			(((st[kIdxB[7]] ^ st[kIdxC[7]]) & st[kIdxA[7]]) ^ st[kIdxC[7]]))
		st[kIdxA[7]] = ((tp << 13) | (tp >> (32 - 13))) + tA[kPrem[4][7]]
		st[kIdxD[7]] = st[kIdxC[7]]
		st[kIdxC[7]] = st[kIdxB[7]]
		st[kIdxB[7]] = tA[7]

		tA[0] = ((st[0] << 13) | (st[0] >> (32 - 13)))
		tA[1] = ((st[1] << 13) | (st[1] >> (32 - 13)))
		tA[2] = ((st[2] << 13) | (st[2] >> (32 - 13)))
		tA[3] = ((st[3] << 13) | (st[3] >> (32 - 13)))
		tA[4] = ((st[4] << 13) | (st[4] >> (32 - 13)))
		tA[5] = ((st[5] << 13) | (st[5] >> (32 - 13)))
		tA[6] = ((st[6] << 13) | (st[6] >> (32 - 13)))
		tA[7] = ((st[7] << 13) | (st[7] >> (32 - 13)))

		tp = uint32(st[kIdxD[0]] + sta[8] +
			(((st[kIdxB[0]] ^ st[kIdxC[0]]) & st[kIdxA[0]]) ^ st[kIdxC[0]]))
		st[kIdxA[0]] = ((tp << 10) | (tp >> (32 - 10))) + tA[kPrem[5][0]]
		st[kIdxD[0]] = st[kIdxC[0]]
		st[kIdxC[0]] = st[kIdxB[0]]
		st[kIdxB[0]] = tA[0]

		tp = uint32(st[kIdxD[1]] + sta[9] +
			(((st[kIdxB[1]] ^ st[kIdxC[1]]) & st[kIdxA[1]]) ^ st[kIdxC[1]]))
		st[kIdxA[1]] = ((tp << 10) | (tp >> (32 - 10))) + tA[kPrem[5][1]]
		st[kIdxD[1]] = st[kIdxC[1]]
		st[kIdxC[1]] = st[kIdxB[1]]
		st[kIdxB[1]] = tA[1]

		tp = uint32(st[kIdxD[2]] + sta[10] +
			(((st[kIdxB[2]] ^ st[kIdxC[2]]) & st[kIdxA[2]]) ^ st[kIdxC[2]]))
		st[kIdxA[2]] = ((tp << 10) | (tp >> (32 - 10))) + tA[kPrem[5][2]]
		st[kIdxD[2]] = st[kIdxC[2]]
		st[kIdxC[2]] = st[kIdxB[2]]
		st[kIdxB[2]] = tA[2]

		tp = uint32(st[kIdxD[3]] + sta[11] +
			(((st[kIdxB[3]] ^ st[kIdxC[3]]) & st[kIdxA[3]]) ^ st[kIdxC[3]]))
		st[kIdxA[3]] = ((tp << 10) | (tp >> (32 - 10))) + tA[kPrem[5][3]]
		st[kIdxD[3]] = st[kIdxC[3]]
		st[kIdxC[3]] = st[kIdxB[3]]
		st[kIdxB[3]] = tA[3]

		tp = uint32(st[kIdxD[4]] + sta[12] +
			(((st[kIdxB[4]] ^ st[kIdxC[4]]) & st[kIdxA[4]]) ^ st[kIdxC[4]]))
		st[kIdxA[4]] = ((tp << 10) | (tp >> (32 - 10))) + tA[kPrem[5][4]]
		st[kIdxD[4]] = st[kIdxC[4]]
		st[kIdxC[4]] = st[kIdxB[4]]
		st[kIdxB[4]] = tA[4]

		tp = uint32(st[kIdxD[5]] + sta[13] +
			(((st[kIdxB[5]] ^ st[kIdxC[5]]) & st[kIdxA[5]]) ^ st[kIdxC[5]]))
		st[kIdxA[5]] = ((tp << 10) | (tp >> (32 - 10))) + tA[kPrem[5][5]]
		st[kIdxD[5]] = st[kIdxC[5]]
		st[kIdxC[5]] = st[kIdxB[5]]
		st[kIdxB[5]] = tA[5]

		tp = uint32(st[kIdxD[6]] + sta[14] +
			(((st[kIdxB[6]] ^ st[kIdxC[6]]) & st[kIdxA[6]]) ^ st[kIdxC[6]]))
		st[kIdxA[6]] = ((tp << 10) | (tp >> (32 - 10))) + tA[kPrem[5][6]]
		st[kIdxD[6]] = st[kIdxC[6]]
		st[kIdxC[6]] = st[kIdxB[6]]
		st[kIdxB[6]] = tA[6]

		tp = uint32(st[kIdxD[7]] + sta[15] +
			(((st[kIdxB[7]] ^ st[kIdxC[7]]) & st[kIdxA[7]]) ^ st[kIdxC[7]]))
		st[kIdxA[7]] = ((tp << 10) | (tp >> (32 - 10))) + tA[kPrem[5][7]]
		st[kIdxD[7]] = st[kIdxC[7]]
		st[kIdxC[7]] = st[kIdxB[7]]
		st[kIdxB[7]] = tA[7]

		tA[0] = ((st[0] << 10) | (st[0] >> (32 - 10)))
		tA[1] = ((st[1] << 10) | (st[1] >> (32 - 10)))
		tA[2] = ((st[2] << 10) | (st[2] >> (32 - 10)))
		tA[3] = ((st[3] << 10) | (st[3] >> (32 - 10)))
		tA[4] = ((st[4] << 10) | (st[4] >> (32 - 10)))
		tA[5] = ((st[5] << 10) | (st[5] >> (32 - 10)))
		tA[6] = ((st[6] << 10) | (st[6] >> (32 - 10)))
		tA[7] = ((st[7] << 10) | (st[7] >> (32 - 10)))

		tp = uint32(st[kIdxD[0]] + sta[16] +
			(((st[kIdxB[0]] ^ st[kIdxC[0]]) & st[kIdxA[0]]) ^ st[kIdxC[0]]))
		st[kIdxA[0]] = ((tp << 25) | (tp >> (32 - 25))) + tA[kPrem[6][0]]
		st[kIdxD[0]] = st[kIdxC[0]]
		st[kIdxC[0]] = st[kIdxB[0]]
		st[kIdxB[0]] = tA[0]

		tp = uint32(st[kIdxD[1]] + sta[17] +
			(((st[kIdxB[1]] ^ st[kIdxC[1]]) & st[kIdxA[1]]) ^ st[kIdxC[1]]))
		st[kIdxA[1]] = ((tp << 25) | (tp >> (32 - 25))) + tA[kPrem[6][1]]
		st[kIdxD[1]] = st[kIdxC[1]]
		st[kIdxC[1]] = st[kIdxB[1]]
		st[kIdxB[1]] = tA[1]

		tp = uint32(st[kIdxD[2]] + sta[18] +
			(((st[kIdxB[2]] ^ st[kIdxC[2]]) & st[kIdxA[2]]) ^ st[kIdxC[2]]))
		st[kIdxA[2]] = ((tp << 25) | (tp >> (32 - 25))) + tA[kPrem[6][2]]
		st[kIdxD[2]] = st[kIdxC[2]]
		st[kIdxC[2]] = st[kIdxB[2]]
		st[kIdxB[2]] = tA[2]

		tp = uint32(st[kIdxD[3]] + sta[19] +
			(((st[kIdxB[3]] ^ st[kIdxC[3]]) & st[kIdxA[3]]) ^ st[kIdxC[3]]))
		st[kIdxA[3]] = ((tp << 25) | (tp >> (32 - 25))) + tA[kPrem[6][3]]
		st[kIdxD[3]] = st[kIdxC[3]]
		st[kIdxC[3]] = st[kIdxB[3]]
		st[kIdxB[3]] = tA[3]

		tp = uint32(st[kIdxD[4]] + sta[20] +
			(((st[kIdxB[4]] ^ st[kIdxC[4]]) & st[kIdxA[4]]) ^ st[kIdxC[4]]))
		st[kIdxA[4]] = ((tp << 25) | (tp >> (32 - 25))) + tA[kPrem[6][4]]
		st[kIdxD[4]] = st[kIdxC[4]]
		st[kIdxC[4]] = st[kIdxB[4]]
		st[kIdxB[4]] = tA[4]

		tp = uint32(st[kIdxD[5]] + sta[21] +
			(((st[kIdxB[5]] ^ st[kIdxC[5]]) & st[kIdxA[5]]) ^ st[kIdxC[5]]))
		st[kIdxA[5]] = ((tp << 25) | (tp >> (32 - 25))) + tA[kPrem[6][5]]
		st[kIdxD[5]] = st[kIdxC[5]]
		st[kIdxC[5]] = st[kIdxB[5]]
		st[kIdxB[5]] = tA[5]

		tp = uint32(st[kIdxD[6]] + sta[22] +
			(((st[kIdxB[6]] ^ st[kIdxC[6]]) & st[kIdxA[6]]) ^ st[kIdxC[6]]))
		st[kIdxA[6]] = ((tp << 25) | (tp >> (32 - 25))) + tA[kPrem[6][6]]
		st[kIdxD[6]] = st[kIdxC[6]]
		st[kIdxC[6]] = st[kIdxB[6]]
		st[kIdxB[6]] = tA[6]

		tp = uint32(st[kIdxD[7]] + sta[23] +
			(((st[kIdxB[7]] ^ st[kIdxC[7]]) & st[kIdxA[7]]) ^ st[kIdxC[7]]))
		st[kIdxA[7]] = ((tp << 25) | (tp >> (32 - 25))) + tA[kPrem[6][7]]
		st[kIdxD[7]] = st[kIdxC[7]]
		st[kIdxC[7]] = st[kIdxB[7]]
		st[kIdxB[7]] = tA[7]

		tA[0] = ((st[0] << 25) | (st[0] >> (32 - 25)))
		tA[1] = ((st[1] << 25) | (st[1] >> (32 - 25)))
		tA[2] = ((st[2] << 25) | (st[2] >> (32 - 25)))
		tA[3] = ((st[3] << 25) | (st[3] >> (32 - 25)))
		tA[4] = ((st[4] << 25) | (st[4] >> (32 - 25)))
		tA[5] = ((st[5] << 25) | (st[5] >> (32 - 25)))
		tA[6] = ((st[6] << 25) | (st[6] >> (32 - 25)))
		tA[7] = ((st[7] << 25) | (st[7] >> (32 - 25)))

		tp = uint32(st[kIdxD[0]] + sta[24] +
			(((st[kIdxB[0]] ^ st[kIdxC[0]]) & st[kIdxA[0]]) ^ st[kIdxC[0]]))
		st[kIdxA[0]] = ((tp << 4) | (tp >> (32 - 4))) + tA[kPrem[0][0]]
		st[kIdxD[0]] = st[kIdxC[0]]
		st[kIdxC[0]] = st[kIdxB[0]]
		st[kIdxB[0]] = tA[0]

		tp = uint32(st[kIdxD[1]] + sta[25] +
			(((st[kIdxB[1]] ^ st[kIdxC[1]]) & st[kIdxA[1]]) ^ st[kIdxC[1]]))
		st[kIdxA[1]] = ((tp << 4) | (tp >> (32 - 4))) + tA[kPrem[0][1]]
		st[kIdxD[1]] = st[kIdxC[1]]
		st[kIdxC[1]] = st[kIdxB[1]]
		st[kIdxB[1]] = tA[1]

		tp = uint32(st[kIdxD[2]] + sta[26] +
			(((st[kIdxB[2]] ^ st[kIdxC[2]]) & st[kIdxA[2]]) ^ st[kIdxC[2]]))
		st[kIdxA[2]] = ((tp << 4) | (tp >> (32 - 4))) + tA[kPrem[0][2]]
		st[kIdxD[2]] = st[kIdxC[2]]
		st[kIdxC[2]] = st[kIdxB[2]]
		st[kIdxB[2]] = tA[2]

		tp = uint32(st[kIdxD[3]] + sta[27] +
			(((st[kIdxB[3]] ^ st[kIdxC[3]]) & st[kIdxA[3]]) ^ st[kIdxC[3]]))
		st[kIdxA[3]] = ((tp << 4) | (tp >> (32 - 4))) + tA[kPrem[0][3]]
		st[kIdxD[3]] = st[kIdxC[3]]
		st[kIdxC[3]] = st[kIdxB[3]]
		st[kIdxB[3]] = tA[3]

		tp = uint32(st[kIdxD[4]] + sta[28] +
			(((st[kIdxB[4]] ^ st[kIdxC[4]]) & st[kIdxA[4]]) ^ st[kIdxC[4]]))
		st[kIdxA[4]] = ((tp << 4) | (tp >> (32 - 4))) + tA[kPrem[0][4]]
		st[kIdxD[4]] = st[kIdxC[4]]
		st[kIdxC[4]] = st[kIdxB[4]]
		st[kIdxB[4]] = tA[4]

		tp = uint32(st[kIdxD[5]] + sta[29] +
			(((st[kIdxB[5]] ^ st[kIdxC[5]]) & st[kIdxA[5]]) ^ st[kIdxC[5]]))
		st[kIdxA[5]] = ((tp << 4) | (tp >> (32 - 4))) + tA[kPrem[0][5]]
		st[kIdxD[5]] = st[kIdxC[5]]
		st[kIdxC[5]] = st[kIdxB[5]]
		st[kIdxB[5]] = tA[5]

		tp = uint32(st[kIdxD[6]] + sta[30] +
			(((st[kIdxB[6]] ^ st[kIdxC[6]]) & st[kIdxA[6]]) ^ st[kIdxC[6]]))
		st[kIdxA[6]] = ((tp << 4) | (tp >> (32 - 4))) + tA[kPrem[0][6]]
		st[kIdxD[6]] = st[kIdxC[6]]
		st[kIdxC[6]] = st[kIdxB[6]]
		st[kIdxB[6]] = tA[6]

		tp = uint32(st[kIdxD[7]] + sta[31] +
			(((st[kIdxB[7]] ^ st[kIdxC[7]]) & st[kIdxA[7]]) ^ st[kIdxC[7]]))
		st[kIdxA[7]] = ((tp << 4) | (tp >> (32 - 4))) + tA[kPrem[0][7]]
		st[kIdxD[7]] = st[kIdxC[7]]
		st[kIdxC[7]] = st[kIdxB[7]]
		st[kIdxB[7]] = tA[7]
	}

	copy(ref.h[:], st[:])
}

func mixoutRound(x []uint8, q []int32, xt uintptr) {
	var tx int32
	var d1_0, d1_1, d1_2, d1_3, d1_4, d1_5, d1_6, d1_7 int32
	var d2_0, d2_1, d2_2, d2_3, d2_4, d2_5, d2_6, d2_7 int32

	xd := xt << 1

	{
		var sa, sb uintptr
		var x0, x1, x2, x3 int32
		var a0, a1, a2, a3 int32
		var b0, b1, b2, b3 int32

		sb = xd << 2

		x0 = int32(x[0])
		x1 = int32(x[sb])
		x2 = int32(x[2*sb])
		x3 = int32(x[3*sb])

		a0 = x0 + x2
		a1 = x0 + (x2 << 4)
		a2 = x0 - x2
		a3 = x0 - (x2 << 4)

		b0 = x1 + x3
		tx = ((x1 << 2) + (x3 << 6))
		b1 = ((tx & 0xFF) - (tx >> 8))
		b2 = (x1 << 4) - (x3 << 4)
		tx = (x1 << 6) + (x3 << 2)
		b3 = ((tx & 0xFF) - (tx >> 8))

		d1_0 = a0 + b0
		d1_1 = a1 + b1
		d1_2 = a2 + b2
		d1_3 = a3 + b3
		d1_4 = a0 - b0
		d1_5 = a1 - b1
		d1_6 = a2 - b2
		d1_7 = a3 - b3

		sa = xd << 1
		sb = xd << 2

		x0 = int32(x[sa])
		x1 = int32(x[sa+sb])
		x2 = int32(x[sa+2*sb])
		x3 = int32(x[sa+3*sb])

		a0 = x0 + x2
		a1 = x0 + (x2 << 4)
		a2 = x0 - x2
		a3 = x0 - (x2 << 4)

		b0 = x1 + x3
		tx = ((x1 << 2) + (x3 << 6))
		b1 = ((tx & 0xFF) - (tx >> 8))
		b2 = (x1 << 4) - (x3 << 4)
		tx = (x1 << 6) + (x3 << 2)
		b3 = ((tx & 0xFF) - (tx >> 8))

		d2_0 = a0 + b0
		d2_1 = a1 + b1
		d2_2 = a2 + b2
		d2_3 = a3 + b3
		d2_4 = a0 - b0
		d2_5 = a1 - b1
		d2_6 = a2 - b2
		d2_7 = a3 - b3
	}

	q[0] = d1_0 + d2_0
	q[1] = d1_1 + (d2_1 << 1)
	q[2] = d1_2 + (d2_2 << 2)
	q[3] = d1_3 + (d2_3 << 3)
	q[4] = d1_4 + (d2_4 << 4)
	q[5] = d1_5 + (d2_5 << 5)
	q[6] = d1_6 + (d2_6 << 6)
	q[7] = d1_7 + (d2_7 << 7)
	q[8] = d1_0 - d2_0
	q[9] = d1_1 - (d2_1 << 1)
	q[10] = d1_2 - (d2_2 << 2)
	q[11] = d1_3 - (d2_3 << 3)
	q[12] = d1_4 - (d2_4 << 4)
	q[13] = d1_5 - (d2_5 << 5)
	q[14] = d1_6 - (d2_6 << 6)
	q[15] = d1_7 - (d2_7 << 7)

	{
		var sa, sb uintptr
		var x0, x1, x2, x3 int32
		var a0, a1, a2, a3 int32
		var b0, b1, b2, b3 int32

		sb = xd << 2

		x0 = int32(x[(xd)])
		x1 = int32(x[(xd)+sb])
		x2 = int32(x[(xd)+2*sb])
		x3 = int32(x[(xd)+3*sb])

		a0 = x0 + x2
		a1 = x0 + (x2 << 4)
		a2 = x0 - x2
		a3 = x0 - (x2 << 4)

		b0 = x1 + x3
		tx = ((x1 << 2) + (x3 << 6))
		b1 = ((tx & 0xFF) - (tx >> 8))
		b2 = (x1 << 4) - (x3 << 4)
		tx = (x1 << 6) + (x3 << 2)
		b3 = ((tx & 0xFF) - (tx >> 8))

		d1_0 = a0 + b0
		d1_1 = a1 + b1
		d1_2 = a2 + b2
		d1_3 = a3 + b3
		d1_4 = a0 - b0
		d1_5 = a1 - b1
		d1_6 = a2 - b2
		d1_7 = a3 - b3

		sa = xd + (xd << 1)
		sb = xd << 2

		x0 = int32(x[sa])
		x1 = int32(x[sa+sb])
		x2 = int32(x[sa+2*sb])
		x3 = int32(x[sa+3*sb])

		a0 = x0 + x2
		a1 = x0 + (x2 << 4)
		a2 = x0 - x2
		a3 = x0 - (x2 << 4)

		b0 = x1 + x3
		tx = ((x1 << 2) + (x3 << 6))
		b1 = ((tx & 0xFF) - (tx >> 8))
		b2 = (x1 << 4) - (x3 << 4)
		tx = (x1 << 6) + (x3 << 2)
		b3 = ((tx & 0xFF) - (tx >> 8))

		d2_0 = a0 + b0
		d2_1 = a1 + b1
		d2_2 = a2 + b2
		d2_3 = a3 + b3
		d2_4 = a0 - b0
		d2_5 = a1 - b1
		d2_6 = a2 - b2
		d2_7 = a3 - b3
	}

	q[16+0] = d1_0 + d2_0
	q[16+1] = d1_1 + (d2_1 << 1)
	q[16+2] = d1_2 + (d2_2 << 2)
	q[16+3] = d1_3 + (d2_3 << 3)
	q[16+4] = d1_4 + (d2_4 << 4)
	q[16+5] = d1_5 + (d2_5 << 5)
	q[16+6] = d1_6 + (d2_6 << 6)
	q[16+7] = d1_7 + (d2_7 << 7)
	q[16+8] = d1_0 - d2_0
	q[16+9] = d1_1 - (d2_1 << 1)
	q[16+10] = d1_2 - (d2_2 << 2)
	q[16+11] = d1_3 - (d2_3 << 3)
	q[16+12] = d1_4 - (d2_4 << 4)
	q[16+13] = d1_5 - (d2_5 << 5)
	q[16+14] = d1_6 - (d2_6 << 6)
	q[16+15] = d1_7 - (d2_7 << 7)

	{
		var u, v uintptr

		m := q[0]
		n := q[16]
		q[0] = m + n
		q[16] = m - n

		m = q[u+1]
		n = q[u+1+16]
		tx = (n * kAlphaTab[v+1*8])
		tx = ((tx & 0xFFFF) + (tx >> 16))
		q[u+1] = m + tx
		q[u+1+16] = m - tx
		m = q[u+2]
		n = q[u+2+16]
		tx = (n * kAlphaTab[v+2*8])
		tx = ((tx & 0xFFFF) + (tx >> 16))
		q[u+2] = m + tx
		q[u+2+16] = m - tx
		m = q[u+3]
		n = q[u+3+16]
		tx = (n * kAlphaTab[v+3*8])
		tx = ((tx & 0xFFFF) + (tx >> 16))
		q[u+3] = m + tx
		q[u+3+16] = m - tx

		for u < 16 {
			u += 4
			v += 4 * 8

			m = q[u+0]
			n = q[u+0+16]
			tx = (n * kAlphaTab[v+0*8])
			tx = ((tx & 0xFFFF) + (tx >> 16))
			q[u+0] = m + tx
			q[u+0+16] = m - tx

			m = q[u+1]
			n = q[u+1+16]
			tx = (n * kAlphaTab[v+1*8])
			tx = ((tx & 0xFFFF) + (tx >> 16))
			q[u+1] = m + tx
			q[u+1+16] = m - tx
			m = q[u+2]
			n = q[u+2+16]
			tx = (n * kAlphaTab[v+2*8])
			tx = ((tx & 0xFFFF) + (tx >> 16))
			q[u+2] = m + tx
			q[u+2+16] = m - tx
			m = q[u+3]
			n = q[u+3+16]
			tx = (n * kAlphaTab[v+3*8])
			tx = ((tx & 0xFFFF) + (tx >> 16))
			q[u+3] = m + tx
			q[u+3+16] = m - tx
		}
	}

	{
		var sa, sb uintptr
		var x0, x1, x2, x3 int32
		var a0, a1, a2, a3 int32
		var b0, b1, b2, b3 int32

		sb = uintptr(xd << 2)

		x0 = int32(x[xt])
		x1 = int32(x[xt+sb])
		x2 = int32(x[xt+2*sb])
		x3 = int32(x[xt+3*sb])

		a0 = x0 + x2
		a1 = x0 + (x2 << 4)
		a2 = x0 - x2
		a3 = x0 - (x2 << 4)

		b0 = x1 + x3
		tx = ((x1 << 2) + (x3 << 6))
		b1 = ((tx & 0xFF) - (tx >> 8))
		b2 = (x1 << 4) - (x3 << 4)
		tx = (x1 << 6) + (x3 << 2)
		b3 = ((tx & 0xFF) - (tx >> 8))

		d1_0 = a0 + b0
		d1_1 = a1 + b1
		d1_2 = a2 + b2
		d1_3 = a3 + b3
		d1_4 = a0 - b0
		d1_5 = a1 - b1
		d1_6 = a2 - b2
		d1_7 = a3 - b3

		sa = xt + (xd << 1)
		sb = xd << 2

		x0 = int32(x[sa])
		x1 = int32(x[sa+sb])
		x2 = int32(x[sa+2*sb])
		x3 = int32(x[sa+3*sb])

		a0 = x0 + x2
		a1 = x0 + (x2 << 4)
		a2 = x0 - x2
		a3 = x0 - (x2 << 4)

		b0 = x1 + x3
		tx = ((x1 << 2) + (x3 << 6))
		b1 = ((tx & 0xFF) - (tx >> 8))
		b2 = (x1 << 4) - (x3 << 4)
		tx = (x1 << 6) + (x3 << 2)
		b3 = ((tx & 0xFF) - (tx >> 8))

		d2_0 = a0 + b0
		d2_1 = a1 + b1
		d2_2 = a2 + b2
		d2_3 = a3 + b3
		d2_4 = a0 - b0
		d2_5 = a1 - b1
		d2_6 = a2 - b2
		d2_7 = a3 - b3
	}

	q[32+0] = d1_0 + d2_0
	q[32+1] = d1_1 + (d2_1 << 1)
	q[32+2] = d1_2 + (d2_2 << 2)
	q[32+3] = d1_3 + (d2_3 << 3)
	q[32+4] = d1_4 + (d2_4 << 4)
	q[32+5] = d1_5 + (d2_5 << 5)
	q[32+6] = d1_6 + (d2_6 << 6)
	q[32+7] = d1_7 + (d2_7 << 7)
	q[32+8] = d1_0 - d2_0
	q[32+9] = d1_1 - (d2_1 << 1)
	q[32+10] = d1_2 - (d2_2 << 2)
	q[32+11] = d1_3 - (d2_3 << 3)
	q[32+12] = d1_4 - (d2_4 << 4)
	q[32+13] = d1_5 - (d2_5 << 5)
	q[32+14] = d1_6 - (d2_6 << 6)
	q[32+15] = d1_7 - (d2_7 << 7)

	{
		var sa, sb uintptr
		var x0, x1, x2, x3 int32
		var a0, a1, a2, a3 int32
		var b0, b1, b2, b3 int32

		sa = (xt) + (xd)
		sb = xd << 2

		x0 = int32(x[sa])
		x1 = int32(x[sa+sb])
		x2 = int32(x[sa+2*sb])
		x3 = int32(x[sa+3*sb])

		a0 = x0 + x2
		a1 = x0 + (x2 << 4)
		a2 = x0 - x2
		a3 = x0 - (x2 << 4)

		b0 = x1 + x3
		tx = ((x1 << 2) + (x3 << 6))
		b1 = ((tx & 0xFF) - (tx >> 8))
		b2 = (x1 << 4) - (x3 << 4)
		tx = (x1 << 6) + (x3 << 2)
		b3 = ((tx & 0xFF) - (tx >> 8))

		d1_0 = a0 + b0
		d1_1 = a1 + b1
		d1_2 = a2 + b2
		d1_3 = a3 + b3
		d1_4 = a0 - b0
		d1_5 = a1 - b1
		d1_6 = a2 - b2
		d1_7 = a3 - b3

		sa = (xt + xd) + (xd << 1)
		sb = xd << 2

		x0 = int32(x[sa])
		x1 = int32(x[sa+sb])
		x2 = int32(x[sa+2*sb])
		x3 = int32(x[sa+3*sb])

		a0 = x0 + x2
		a1 = x0 + (x2 << 4)
		a2 = x0 - x2
		a3 = x0 - (x2 << 4)

		b0 = x1 + x3
		tx = ((x1 << 2) + (x3 << 6))
		b1 = ((tx & 0xFF) - (tx >> 8))
		b2 = (x1 << 4) - (x3 << 4)
		tx = (x1 << 6) + (x3 << 2)
		b3 = ((tx & 0xFF) - (tx >> 8))

		d2_0 = a0 + b0
		d2_1 = a1 + b1
		d2_2 = a2 + b2
		d2_3 = a3 + b3
		d2_4 = a0 - b0
		d2_5 = a1 - b1
		d2_6 = a2 - b2
		d2_7 = a3 - b3
	}

	q[48+0] = d1_0 + d2_0
	q[48+1] = d1_1 + (d2_1 << 1)
	q[48+2] = d1_2 + (d2_2 << 2)
	q[48+3] = d1_3 + (d2_3 << 3)
	q[48+4] = d1_4 + (d2_4 << 4)
	q[48+5] = d1_5 + (d2_5 << 5)
	q[48+6] = d1_6 + (d2_6 << 6)
	q[48+7] = d1_7 + (d2_7 << 7)
	q[48+8] = d1_0 - d2_0
	q[48+9] = d1_1 - (d2_1 << 1)
	q[48+10] = d1_2 - (d2_2 << 2)
	q[48+11] = d1_3 - (d2_3 << 3)
	q[48+12] = d1_4 - (d2_4 << 4)
	q[48+13] = d1_5 - (d2_5 << 5)
	q[48+14] = d1_6 - (d2_6 << 6)
	q[48+15] = d1_7 - (d2_7 << 7)

	{
		var u, v uintptr

		m := q[(32)]
		n := q[(32)+(16)]
		q[(32)] = m + n
		q[(32)+(16)] = m - n

		m = q[(32)+u+1]
		n = q[(32)+u+1+(16)]
		tx = (n * kAlphaTab[v+1*(8)])
		tx = ((tx & 0xFFFF) + (tx >> 16))
		q[(32)+u+1] = m + tx
		q[(32)+u+1+(16)] = m - tx
		m = q[(32)+u+2]
		n = q[(32)+u+2+(16)]
		tx = (n * kAlphaTab[v+2*(8)])
		tx = ((tx & 0xFFFF) + (tx >> 16))
		q[(32)+u+2] = m + tx
		q[(32)+u+2+(16)] = m - tx
		m = q[(32)+u+3]
		n = q[(32)+u+3+(16)]
		tx = (n * kAlphaTab[v+3*(8)])
		tx = ((tx & 0xFFFF) + (tx >> 16))
		q[(32)+u+3] = m + tx
		q[(32)+u+3+(16)] = m - tx

		u = 4
		v = 4 * (8)
		for u < 16 {
			m = q[(32)+u]
			n = q[(32)+u+(16)]
			tx = (n * kAlphaTab[v+0*(8)])
			tx = ((tx & 0xFFFF) + (tx >> 16))
			q[(32)+u+0] = m + tx
			q[(32)+u+0+(16)] = m - tx

			m = q[(32)+u+1]
			n = q[(32)+u+1+(16)]
			tx = (n * kAlphaTab[v+1*(8)])
			tx = ((tx & 0xFFFF) + (tx >> 16))
			q[(32)+u+1] = m + tx
			q[(32)+u+1+(16)] = m - tx
			m = q[(32)+u+2]
			n = q[(32)+u+2+(16)]
			tx = (n * kAlphaTab[v+2*(8)])
			tx = ((tx & 0xFFFF) + (tx >> 16))
			q[(32)+u+2] = m + tx
			q[(32)+u+2+(16)] = m - tx
			m = q[(32)+u+3]
			n = q[(32)+u+3+(16)]
			tx = (n * kAlphaTab[v+3*(8)])
			tx = ((tx & 0xFFFF) + (tx >> 16))
			q[(32)+u+3] = m + tx
			q[(32)+u+3+(16)] = m - tx

			u += 4
			v += 4 * (8)
		}
	}

	{
		var u, v uintptr

		m := q[0]
		n := q[32]
		q[0] = m + n
		q[32] = m - n

		m = q[u+1]
		n = q[u+1+32]
		tx = (n * kAlphaTab[v+1*4])
		tx = ((tx & 0xFFFF) + (tx >> 16))
		q[u+1] = m + tx
		q[u+1+32] = m - tx
		m = q[u+2]
		n = q[u+2+32]
		tx = (n * kAlphaTab[v+2*4])
		tx = ((tx & 0xFFFF) + (tx >> 16))
		q[u+2] = m + tx
		q[u+2+32] = m - tx
		m = q[u+3]
		n = q[u+3+32]
		tx = (n * kAlphaTab[v+3*4])
		tx = ((tx & 0xFFFF) + (tx >> 16))
		q[u+3] = m + tx
		q[u+3+32] = m - tx

		u = 4
		v = 4 * 4
		for u < 32 {
			m = q[u]
			n = q[u+32]
			tx = (n * kAlphaTab[v+0*4])
			tx = ((tx & 0xFFFF) + (tx >> 16))
			q[u] = m + tx
			q[u+(32)] = m - tx

			m = q[u+1]
			n = q[u+1+32]
			tx = (n * kAlphaTab[v+1*4])
			tx = ((tx & 0xFFFF) + (tx >> 16))
			q[u+1] = m + tx
			q[u+1+32] = m - tx
			m = q[u+2]
			n = q[u+2+32]
			tx = (n * kAlphaTab[v+2*4])
			tx = ((tx & 0xFFFF) + (tx >> 16))
			q[u+2] = m + tx
			q[u+2+32] = m - tx
			m = q[u+3]
			n = q[u+3+32]
			tx = (n * kAlphaTab[v+3*4])
			tx = ((tx & 0xFFFF) + (tx >> 16))
			q[u+3] = m + tx
			q[u+3+32] = m - tx

			u += 4
			v += 4 * 4
		}
	}
}

func mixinRound(h, w []uint32, isp, p0, p1, p2, p3 uint32) {
	var tA [8]uint32
	var tp uint32

	tA[0] = ((h[0] << p0) | (h[0] >> (32 - p0)))
	tA[1] = ((h[1] << p0) | (h[1] >> (32 - p0)))
	tA[2] = ((h[2] << p0) | (h[2] >> (32 - p0)))
	tA[3] = ((h[3] << p0) | (h[3] >> (32 - p0)))
	tA[4] = ((h[4] << p0) | (h[4] >> (32 - p0)))
	tA[5] = ((h[5] << p0) | (h[5] >> (32 - p0)))
	tA[6] = ((h[6] << p0) | (h[6] >> (32 - p0)))
	tA[7] = ((h[7] << p0) | (h[7] >> (32 - p0)))

	tp = uint32(h[kIdxD[0]] + w[0] +
		(((h[kIdxB[0]] ^ h[kIdxC[0]]) & h[kIdxA[0]]) ^ h[kIdxC[0]]))
	h[kIdxA[0]] = ((tp << p1) | (tp >> (32 - p1))) + tA[kPrems[isp]]
	h[kIdxD[0]] = h[kIdxC[0]]
	h[kIdxC[0]] = h[kIdxB[0]]
	h[kIdxB[0]] = tA[0]

	tp = uint32(h[kIdxD[1]] + w[1] +
		(((h[kIdxB[1]] ^ h[kIdxC[1]]) & h[kIdxA[1]]) ^ h[kIdxC[1]]))
	h[kIdxA[1]] = ((tp << p1) | (tp >> (32 - p1))) + tA[kPrems[isp]^1]
	h[kIdxD[1]] = h[kIdxC[1]]
	h[kIdxC[1]] = h[kIdxB[1]]
	h[kIdxB[1]] = tA[1]

	tp = uint32(h[kIdxD[2]] + w[2] +
		(((h[kIdxB[2]] ^ h[kIdxC[2]]) & h[kIdxA[2]]) ^ h[kIdxC[2]]))
	h[kIdxA[2]] = ((tp << p1) | (tp >> (32 - p1))) + tA[kPrems[isp]^2]
	h[kIdxD[2]] = h[kIdxC[2]]
	h[kIdxC[2]] = h[kIdxB[2]]
	h[kIdxB[2]] = tA[2]

	tp = uint32(h[kIdxD[3]] + w[3] +
		(((h[kIdxB[3]] ^ h[kIdxC[3]]) & h[kIdxA[3]]) ^ h[kIdxC[3]]))
	h[kIdxA[3]] = ((tp << p1) | (tp >> (32 - p1))) + tA[kPrems[isp]^3]
	h[kIdxD[3]] = h[kIdxC[3]]
	h[kIdxC[3]] = h[kIdxB[3]]
	h[kIdxB[3]] = tA[3]

	tp = uint32(h[kIdxD[4]] + w[4] +
		(((h[kIdxB[4]] ^ h[kIdxC[4]]) & h[kIdxA[4]]) ^ h[kIdxC[4]]))
	h[kIdxA[4]] = ((tp << p1) | (tp >> (32 - p1))) + tA[kPrems[isp]^4]
	h[kIdxD[4]] = h[kIdxC[4]]
	h[kIdxC[4]] = h[kIdxB[4]]
	h[kIdxB[4]] = tA[4]

	tp = uint32(h[kIdxD[5]] + w[5] +
		(((h[kIdxB[5]] ^ h[kIdxC[5]]) & h[kIdxA[5]]) ^ h[kIdxC[5]]))
	h[kIdxA[5]] = ((tp << p1) | (tp >> (32 - p1))) + tA[kPrems[isp]^5]
	h[kIdxD[5]] = h[kIdxC[5]]
	h[kIdxC[5]] = h[kIdxB[5]]
	h[kIdxB[5]] = tA[5]

	tp = uint32(h[kIdxD[6]] + w[6] +
		(((h[kIdxB[6]] ^ h[kIdxC[6]]) & h[kIdxA[6]]) ^ h[kIdxC[6]]))
	h[kIdxA[6]] = ((tp << p1) | (tp >> (32 - p1))) + tA[kPrems[isp]^6]
	h[kIdxD[6]] = h[kIdxC[6]]
	h[kIdxC[6]] = h[kIdxB[6]]
	h[kIdxB[6]] = tA[6]

	tp = uint32(h[kIdxD[7]] + w[7] +
		(((h[kIdxB[7]] ^ h[kIdxC[7]]) & h[kIdxA[7]]) ^ h[kIdxC[7]]))
	h[kIdxA[7]] = ((tp << p1) | (tp >> (32 - p1))) + tA[kPrems[isp]^7]
	h[kIdxD[7]] = h[kIdxC[7]]
	h[kIdxC[7]] = h[kIdxB[7]]
	h[kIdxB[7]] = tA[7]

	tA[0] = ((h[0] << p1) | (h[0] >> (32 - p1)))
	tA[1] = ((h[1] << p1) | (h[1] >> (32 - p1)))
	tA[2] = ((h[2] << p1) | (h[2] >> (32 - p1)))
	tA[3] = ((h[3] << p1) | (h[3] >> (32 - p1)))
	tA[4] = ((h[4] << p1) | (h[4] >> (32 - p1)))
	tA[5] = ((h[5] << p1) | (h[5] >> (32 - p1)))
	tA[6] = ((h[6] << p1) | (h[6] >> (32 - p1)))
	tA[7] = ((h[7] << p1) | (h[7] >> (32 - p1)))

	tp = uint32(h[kIdxD[0]] + w[8] +
		(((h[kIdxB[0]] ^ h[kIdxC[0]]) & h[kIdxA[0]]) ^ h[kIdxC[0]]))
	h[kIdxA[0]] = ((tp << p2) | (tp >> (32 - p2))) + tA[kPrems[isp+1]]
	h[kIdxD[0]] = h[kIdxC[0]]
	h[kIdxC[0]] = h[kIdxB[0]]
	h[kIdxB[0]] = tA[0]

	tp = uint32(h[kIdxD[1]] + w[9] +
		(((h[kIdxB[1]] ^ h[kIdxC[1]]) & h[kIdxA[1]]) ^ h[kIdxC[1]]))
	h[kIdxA[1]] = ((tp << p2) | (tp >> (32 - p2))) + tA[kPrems[isp+1]^1]
	h[kIdxD[1]] = h[kIdxC[1]]
	h[kIdxC[1]] = h[kIdxB[1]]
	h[kIdxB[1]] = tA[1]

	tp = uint32(h[kIdxD[2]] + w[10] +
		(((h[kIdxB[2]] ^ h[kIdxC[2]]) & h[kIdxA[2]]) ^ h[kIdxC[2]]))
	h[kIdxA[2]] = ((tp << p2) | (tp >> (32 - p2))) + tA[kPrems[isp+1]^2]
	h[kIdxD[2]] = h[kIdxC[2]]
	h[kIdxC[2]] = h[kIdxB[2]]
	h[kIdxB[2]] = tA[2]

	tp = uint32(h[kIdxD[3]] + w[11] +
		(((h[kIdxB[3]] ^ h[kIdxC[3]]) & h[kIdxA[3]]) ^ h[kIdxC[3]]))
	h[kIdxA[3]] = ((tp << p2) | (tp >> (32 - p2))) + tA[kPrems[isp+1]^3]
	h[kIdxD[3]] = h[kIdxC[3]]
	h[kIdxC[3]] = h[kIdxB[3]]
	h[kIdxB[3]] = tA[3]

	tp = uint32(h[kIdxD[4]] + w[12] +
		(((h[kIdxB[4]] ^ h[kIdxC[4]]) & h[kIdxA[4]]) ^ h[kIdxC[4]]))
	h[kIdxA[4]] = ((tp << p2) | (tp >> (32 - p2))) + tA[kPrems[isp+1]^4]
	h[kIdxD[4]] = h[kIdxC[4]]
	h[kIdxC[4]] = h[kIdxB[4]]
	h[kIdxB[4]] = tA[4]

	tp = uint32(h[kIdxD[5]] + w[13] +
		(((h[kIdxB[5]] ^ h[kIdxC[5]]) & h[kIdxA[5]]) ^ h[kIdxC[5]]))
	h[kIdxA[5]] = ((tp << p2) | (tp >> (32 - p2))) + tA[kPrems[isp+1]^5]
	h[kIdxD[5]] = h[kIdxC[5]]
	h[kIdxC[5]] = h[kIdxB[5]]
	h[kIdxB[5]] = tA[5]

	tp = uint32(h[kIdxD[6]] + w[14] +
		(((h[kIdxB[6]] ^ h[kIdxC[6]]) & h[kIdxA[6]]) ^ h[kIdxC[6]]))
	h[kIdxA[6]] = ((tp << p2) | (tp >> (32 - p2))) + tA[kPrems[isp+1]^6]
	h[kIdxD[6]] = h[kIdxC[6]]
	h[kIdxC[6]] = h[kIdxB[6]]
	h[kIdxB[6]] = tA[6]

	tp = uint32(h[kIdxD[7]] + w[15] +
		(((h[kIdxB[7]] ^ h[kIdxC[7]]) & h[kIdxA[7]]) ^ h[kIdxC[7]]))
	h[kIdxA[7]] = ((tp << p2) | (tp >> (32 - p2))) + tA[kPrems[isp+1]^7]
	h[kIdxD[7]] = h[kIdxC[7]]
	h[kIdxC[7]] = h[kIdxB[7]]
	h[kIdxB[7]] = tA[7]

	tA[0] = ((h[0] << p2) | (h[0] >> (32 - p2)))
	tA[1] = ((h[1] << p2) | (h[1] >> (32 - p2)))
	tA[2] = ((h[2] << p2) | (h[2] >> (32 - p2)))
	tA[3] = ((h[3] << p2) | (h[3] >> (32 - p2)))
	tA[4] = ((h[4] << p2) | (h[4] >> (32 - p2)))
	tA[5] = ((h[5] << p2) | (h[5] >> (32 - p2)))
	tA[6] = ((h[6] << p2) | (h[6] >> (32 - p2)))
	tA[7] = ((h[7] << p2) | (h[7] >> (32 - p2)))

	tp = uint32(h[kIdxD[0]] + w[16] +
		(((h[kIdxB[0]] ^ h[kIdxC[0]]) & h[kIdxA[0]]) ^ h[kIdxC[0]]))
	h[kIdxA[0]] = ((tp << p3) | (tp >> (32 - p3))) + tA[kPrems[isp+2]]
	h[kIdxD[0]] = h[kIdxC[0]]
	h[kIdxC[0]] = h[kIdxB[0]]
	h[kIdxB[0]] = tA[0]

	tp = uint32(h[kIdxD[1]] + w[17] +
		(((h[kIdxB[1]] ^ h[kIdxC[1]]) & h[kIdxA[1]]) ^ h[kIdxC[1]]))
	h[kIdxA[1]] = ((tp << p3) | (tp >> (32 - p3))) + tA[kPrems[isp+2]^1]
	h[kIdxD[1]] = h[kIdxC[1]]
	h[kIdxC[1]] = h[kIdxB[1]]
	h[kIdxB[1]] = tA[1]

	tp = uint32(h[kIdxD[2]] + w[18] +
		(((h[kIdxB[2]] ^ h[kIdxC[2]]) & h[kIdxA[2]]) ^ h[kIdxC[2]]))
	h[kIdxA[2]] = ((tp << p3) | (tp >> (32 - p3))) + tA[kPrems[isp+2]^2]
	h[kIdxD[2]] = h[kIdxC[2]]
	h[kIdxC[2]] = h[kIdxB[2]]
	h[kIdxB[2]] = tA[2]

	tp = uint32(h[kIdxD[3]] + w[19] +
		(((h[kIdxB[3]] ^ h[kIdxC[3]]) & h[kIdxA[3]]) ^ h[kIdxC[3]]))
	h[kIdxA[3]] = ((tp << p3) | (tp >> (32 - p3))) + tA[kPrems[isp+2]^3]
	h[kIdxD[3]] = h[kIdxC[3]]
	h[kIdxC[3]] = h[kIdxB[3]]
	h[kIdxB[3]] = tA[3]

	tp = uint32(h[kIdxD[4]] + w[20] +
		(((h[kIdxB[4]] ^ h[kIdxC[4]]) & h[kIdxA[4]]) ^ h[kIdxC[4]]))
	h[kIdxA[4]] = ((tp << p3) | (tp >> (32 - p3))) + tA[kPrems[isp+2]^4]
	h[kIdxD[4]] = h[kIdxC[4]]
	h[kIdxC[4]] = h[kIdxB[4]]
	h[kIdxB[4]] = tA[4]

	tp = uint32(h[kIdxD[5]] + w[21] +
		(((h[kIdxB[5]] ^ h[kIdxC[5]]) & h[kIdxA[5]]) ^ h[kIdxC[5]]))
	h[kIdxA[5]] = ((tp << p3) | (tp >> (32 - p3))) + tA[kPrems[isp+2]^5]
	h[kIdxD[5]] = h[kIdxC[5]]
	h[kIdxC[5]] = h[kIdxB[5]]
	h[kIdxB[5]] = tA[5]

	tp = uint32(h[kIdxD[6]] + w[22] +
		(((h[kIdxB[6]] ^ h[kIdxC[6]]) & h[kIdxA[6]]) ^ h[kIdxC[6]]))
	h[kIdxA[6]] = ((tp << p3) | (tp >> (32 - p3))) + tA[kPrems[isp+2]^6]
	h[kIdxD[6]] = h[kIdxC[6]]
	h[kIdxC[6]] = h[kIdxB[6]]
	h[kIdxB[6]] = tA[6]

	tp = uint32(h[kIdxD[7]] + w[23] +
		(((h[kIdxB[7]] ^ h[kIdxC[7]]) & h[kIdxA[7]]) ^ h[kIdxC[7]]))
	h[kIdxA[7]] = ((tp << p3) | (tp >> (32 - p3))) + tA[kPrems[isp+2]^7]
	h[kIdxD[7]] = h[kIdxC[7]]
	h[kIdxC[7]] = h[kIdxB[7]]
	h[kIdxB[7]] = tA[7]

	tA[0] = ((h[0] << p3) | (h[0] >> (32 - p3)))
	tA[1] = ((h[1] << p3) | (h[1] >> (32 - p3)))
	tA[2] = ((h[2] << p3) | (h[2] >> (32 - p3)))
	tA[3] = ((h[3] << p3) | (h[3] >> (32 - p3)))
	tA[4] = ((h[4] << p3) | (h[4] >> (32 - p3)))
	tA[5] = ((h[5] << p3) | (h[5] >> (32 - p3)))
	tA[6] = ((h[6] << p3) | (h[6] >> (32 - p3)))
	tA[7] = ((h[7] << p3) | (h[7] >> (32 - p3)))

	tp = uint32(h[kIdxD[0]] + w[24] +
		(((h[kIdxB[0]] ^ h[kIdxC[0]]) & h[kIdxA[0]]) ^ h[kIdxC[0]]))
	h[kIdxA[0]] = ((tp << p0) | (tp >> (32 - p0))) + tA[kPrems[isp+3]]
	h[kIdxD[0]] = h[kIdxC[0]]
	h[kIdxC[0]] = h[kIdxB[0]]
	h[kIdxB[0]] = tA[0]

	tp = uint32(h[kIdxD[1]] + w[25] +
		(((h[kIdxB[1]] ^ h[kIdxC[1]]) & h[kIdxA[1]]) ^ h[kIdxC[1]]))
	h[kIdxA[1]] = ((tp << p0) | (tp >> (32 - p0))) + tA[kPrems[isp+3]^1]
	h[kIdxD[1]] = h[kIdxC[1]]
	h[kIdxC[1]] = h[kIdxB[1]]
	h[kIdxB[1]] = tA[1]

	tp = uint32(h[kIdxD[2]] + w[26] +
		(((h[kIdxB[2]] ^ h[kIdxC[2]]) & h[kIdxA[2]]) ^ h[kIdxC[2]]))
	h[kIdxA[2]] = ((tp << p0) | (tp >> (32 - p0))) + tA[kPrems[isp+3]^2]
	h[kIdxD[2]] = h[kIdxC[2]]
	h[kIdxC[2]] = h[kIdxB[2]]
	h[kIdxB[2]] = tA[2]

	tp = uint32(h[kIdxD[3]] + w[27] +
		(((h[kIdxB[3]] ^ h[kIdxC[3]]) & h[kIdxA[3]]) ^ h[kIdxC[3]]))
	h[kIdxA[3]] = ((tp << p0) | (tp >> (32 - p0))) + tA[kPrems[isp+3]^3]
	h[kIdxD[3]] = h[kIdxC[3]]
	h[kIdxC[3]] = h[kIdxB[3]]
	h[kIdxB[3]] = tA[3]

	tp = uint32(h[kIdxD[4]] + w[28] +
		(((h[kIdxB[4]] ^ h[kIdxC[4]]) & h[kIdxA[4]]) ^ h[kIdxC[4]]))
	h[kIdxA[4]] = ((tp << p0) | (tp >> (32 - p0))) + tA[kPrems[isp+3]^4]
	h[kIdxD[4]] = h[kIdxC[4]]
	h[kIdxC[4]] = h[kIdxB[4]]
	h[kIdxB[4]] = tA[4]

	tp = uint32(h[kIdxD[5]] + w[29] +
		(((h[kIdxB[5]] ^ h[kIdxC[5]]) & h[kIdxA[5]]) ^ h[kIdxC[5]]))
	h[kIdxA[5]] = ((tp << p0) | (tp >> (32 - p0))) + tA[kPrems[isp+3]^5]
	h[kIdxD[5]] = h[kIdxC[5]]
	h[kIdxC[5]] = h[kIdxB[5]]
	h[kIdxB[5]] = tA[5]

	tp = uint32(h[kIdxD[6]] + w[30] +
		(((h[kIdxB[6]] ^ h[kIdxC[6]]) & h[kIdxA[6]]) ^ h[kIdxC[6]]))
	h[kIdxA[6]] = ((tp << p0) | (tp >> (32 - p0))) + tA[kPrems[isp+3]^6]
	h[kIdxD[6]] = h[kIdxC[6]]
	h[kIdxC[6]] = h[kIdxB[6]]
	h[kIdxB[6]] = tA[6]

	tp = uint32(h[kIdxD[7]] + w[31] +
		(((h[kIdxB[7]] ^ h[kIdxC[7]]) & h[kIdxA[7]]) ^ h[kIdxC[7]]))
	h[kIdxA[7]] = ((tp << p0) | (tp >> (32 - p0))) + tA[kPrems[isp+3]^7]
	h[kIdxD[7]] = h[kIdxC[7]]
	h[kIdxC[7]] = h[kIdxB[7]]
	h[kIdxB[7]] = tA[7]

	tA[0] = ((h[0] << p0) | (h[0] >> (32 - p0)))
	tA[1] = ((h[1] << p0) | (h[1] >> (32 - p0)))
	tA[2] = ((h[2] << p0) | (h[2] >> (32 - p0)))
	tA[3] = ((h[3] << p0) | (h[3] >> (32 - p0)))
	tA[4] = ((h[4] << p0) | (h[4] >> (32 - p0)))
	tA[5] = ((h[5] << p0) | (h[5] >> (32 - p0)))
	tA[6] = ((h[6] << p0) | (h[6] >> (32 - p0)))
	tA[7] = ((h[7] << p0) | (h[7] >> (32 - p0)))

	tp = uint32(h[kIdxD[0]] + w[32] +
		((h[kIdxA[0]] & h[kIdxB[0]]) | ((h[kIdxA[0]] | h[kIdxB[0]]) & h[kIdxC[0]])))
	h[kIdxA[0]] = ((tp << p1) | (tp >> (32 - p1))) + tA[kPrems[isp+4]]
	h[kIdxD[0]] = h[kIdxC[0]]
	h[kIdxC[0]] = h[kIdxB[0]]
	h[kIdxB[0]] = tA[0]

	tp = uint32(h[kIdxD[1]] + w[33] +
		((h[kIdxA[1]] & h[kIdxB[1]]) | ((h[kIdxA[1]] | h[kIdxB[1]]) & h[kIdxC[1]])))
	h[kIdxA[1]] = ((tp << p1) | (tp >> (32 - p1))) + tA[kPrems[isp+4]^1]
	h[kIdxD[1]] = h[kIdxC[1]]
	h[kIdxC[1]] = h[kIdxB[1]]
	h[kIdxB[1]] = tA[1]

	tp = uint32(h[kIdxD[2]] + w[34] +
		((h[kIdxA[2]] & h[kIdxB[2]]) | ((h[kIdxA[2]] | h[kIdxB[2]]) & h[kIdxC[2]])))
	h[kIdxA[2]] = ((tp << p1) | (tp >> (32 - p1))) + tA[kPrems[isp+4]^2]
	h[kIdxD[2]] = h[kIdxC[2]]
	h[kIdxC[2]] = h[kIdxB[2]]
	h[kIdxB[2]] = tA[2]

	tp = uint32(h[kIdxD[3]] + w[35] +
		((h[kIdxA[3]] & h[kIdxB[3]]) | ((h[kIdxA[3]] | h[kIdxB[3]]) & h[kIdxC[3]])))
	h[kIdxA[3]] = ((tp << p1) | (tp >> (32 - p1))) + tA[kPrems[isp+4]^3]
	h[kIdxD[3]] = h[kIdxC[3]]
	h[kIdxC[3]] = h[kIdxB[3]]
	h[kIdxB[3]] = tA[3]

	tp = uint32(h[kIdxD[4]] + w[36] +
		((h[kIdxA[4]] & h[kIdxB[4]]) | ((h[kIdxA[4]] | h[kIdxB[4]]) & h[kIdxC[4]])))
	h[kIdxA[4]] = ((tp << p1) | (tp >> (32 - p1))) + tA[kPrems[isp+4]^4]
	h[kIdxD[4]] = h[kIdxC[4]]
	h[kIdxC[4]] = h[kIdxB[4]]
	h[kIdxB[4]] = tA[4]

	tp = uint32(h[kIdxD[5]] + w[37] +
		((h[kIdxA[5]] & h[kIdxB[5]]) | ((h[kIdxA[5]] | h[kIdxB[5]]) & h[kIdxC[5]])))
	h[kIdxA[5]] = ((tp << p1) | (tp >> (32 - p1))) + tA[kPrems[isp+4]^5]
	h[kIdxD[5]] = h[kIdxC[5]]
	h[kIdxC[5]] = h[kIdxB[5]]
	h[kIdxB[5]] = tA[5]

	tp = uint32(h[kIdxD[6]] + w[38] +
		((h[kIdxA[6]] & h[kIdxB[6]]) | ((h[kIdxA[6]] | h[kIdxB[6]]) & h[kIdxC[6]])))
	h[kIdxA[6]] = ((tp << p1) | (tp >> (32 - p1))) + tA[kPrems[isp+4]^6]
	h[kIdxD[6]] = h[kIdxC[6]]
	h[kIdxC[6]] = h[kIdxB[6]]
	h[kIdxB[6]] = tA[6]

	tp = uint32(h[kIdxD[7]] + w[39] +
		((h[kIdxA[7]] & h[kIdxB[7]]) | ((h[kIdxA[7]] | h[kIdxB[7]]) & h[kIdxC[7]])))
	h[kIdxA[7]] = ((tp << p1) | (tp >> (32 - p1))) + tA[kPrems[isp+4]^7]
	h[kIdxD[7]] = h[kIdxC[7]]
	h[kIdxC[7]] = h[kIdxB[7]]
	h[kIdxB[7]] = tA[7]

	tA[0] = ((h[0] << p1) | (h[0] >> (32 - p1)))
	tA[1] = ((h[1] << p1) | (h[1] >> (32 - p1)))
	tA[2] = ((h[2] << p1) | (h[2] >> (32 - p1)))
	tA[3] = ((h[3] << p1) | (h[3] >> (32 - p1)))
	tA[4] = ((h[4] << p1) | (h[4] >> (32 - p1)))
	tA[5] = ((h[5] << p1) | (h[5] >> (32 - p1)))
	tA[6] = ((h[6] << p1) | (h[6] >> (32 - p1)))
	tA[7] = ((h[7] << p1) | (h[7] >> (32 - p1)))

	tp = uint32(h[kIdxD[0]] + w[40] +
		((h[kIdxA[0]] & h[kIdxB[0]]) | ((h[kIdxA[0]] | h[kIdxB[0]]) & h[kIdxC[0]])))
	h[kIdxA[0]] = ((tp << p2) | (tp >> (32 - p2))) + tA[kPrems[isp+5]]
	h[kIdxD[0]] = h[kIdxC[0]]
	h[kIdxC[0]] = h[kIdxB[0]]
	h[kIdxB[0]] = tA[0]

	tp = uint32(h[kIdxD[1]] + w[41] +
		((h[kIdxA[1]] & h[kIdxB[1]]) | ((h[kIdxA[1]] | h[kIdxB[1]]) & h[kIdxC[1]])))
	h[kIdxA[1]] = ((tp << p2) | (tp >> (32 - p2))) + tA[kPrems[isp+5]^1]
	h[kIdxD[1]] = h[kIdxC[1]]
	h[kIdxC[1]] = h[kIdxB[1]]
	h[kIdxB[1]] = tA[1]

	tp = uint32(h[kIdxD[2]] + w[42] +
		((h[kIdxA[2]] & h[kIdxB[2]]) | ((h[kIdxA[2]] | h[kIdxB[2]]) & h[kIdxC[2]])))
	h[kIdxA[2]] = ((tp << p2) | (tp >> (32 - p2))) + tA[kPrems[isp+5]^2]
	h[kIdxD[2]] = h[kIdxC[2]]
	h[kIdxC[2]] = h[kIdxB[2]]
	h[kIdxB[2]] = tA[2]

	tp = uint32(h[kIdxD[3]] + w[43] +
		((h[kIdxA[3]] & h[kIdxB[3]]) | ((h[kIdxA[3]] | h[kIdxB[3]]) & h[kIdxC[3]])))
	h[kIdxA[3]] = ((tp << p2) | (tp >> (32 - p2))) + tA[kPrems[isp+5]^3]
	h[kIdxD[3]] = h[kIdxC[3]]
	h[kIdxC[3]] = h[kIdxB[3]]
	h[kIdxB[3]] = tA[3]

	tp = uint32(h[kIdxD[4]] + w[44] +
		((h[kIdxA[4]] & h[kIdxB[4]]) | ((h[kIdxA[4]] | h[kIdxB[4]]) & h[kIdxC[4]])))
	h[kIdxA[4]] = ((tp << p2) | (tp >> (32 - p2))) + tA[kPrems[isp+5]^4]
	h[kIdxD[4]] = h[kIdxC[4]]
	h[kIdxC[4]] = h[kIdxB[4]]
	h[kIdxB[4]] = tA[4]

	tp = uint32(h[kIdxD[5]] + w[45] +
		((h[kIdxA[5]] & h[kIdxB[5]]) | ((h[kIdxA[5]] | h[kIdxB[5]]) & h[kIdxC[5]])))
	h[kIdxA[5]] = ((tp << p2) | (tp >> (32 - p2))) + tA[kPrems[isp+5]^5]
	h[kIdxD[5]] = h[kIdxC[5]]
	h[kIdxC[5]] = h[kIdxB[5]]
	h[kIdxB[5]] = tA[5]

	tp = uint32(h[kIdxD[6]] + w[46] +
		((h[kIdxA[6]] & h[kIdxB[6]]) | ((h[kIdxA[6]] | h[kIdxB[6]]) & h[kIdxC[6]])))
	h[kIdxA[6]] = ((tp << p2) | (tp >> (32 - p2))) + tA[kPrems[isp+5]^6]
	h[kIdxD[6]] = h[kIdxC[6]]
	h[kIdxC[6]] = h[kIdxB[6]]
	h[kIdxB[6]] = tA[6]

	tp = uint32(h[kIdxD[7]] + w[47] +
		((h[kIdxA[7]] & h[kIdxB[7]]) | ((h[kIdxA[7]] | h[kIdxB[7]]) & h[kIdxC[7]])))
	h[kIdxA[7]] = ((tp << p2) | (tp >> (32 - p2))) + tA[kPrems[isp+5]^7]
	h[kIdxD[7]] = h[kIdxC[7]]
	h[kIdxC[7]] = h[kIdxB[7]]
	h[kIdxB[7]] = tA[7]

	tA[0] = ((h[0] << p2) | (h[0] >> (32 - p2)))
	tA[1] = ((h[1] << p2) | (h[1] >> (32 - p2)))
	tA[2] = ((h[2] << p2) | (h[2] >> (32 - p2)))
	tA[3] = ((h[3] << p2) | (h[3] >> (32 - p2)))
	tA[4] = ((h[4] << p2) | (h[4] >> (32 - p2)))
	tA[5] = ((h[5] << p2) | (h[5] >> (32 - p2)))
	tA[6] = ((h[6] << p2) | (h[6] >> (32 - p2)))
	tA[7] = ((h[7] << p2) | (h[7] >> (32 - p2)))

	tp = uint32(h[kIdxD[0]] + w[48] +
		((h[kIdxA[0]] & h[kIdxB[0]]) | ((h[kIdxA[0]] | h[kIdxB[0]]) & h[kIdxC[0]])))
	h[kIdxA[0]] = ((tp << p3) | (tp >> (32 - p3))) + tA[kPrems[isp+6]]
	h[kIdxD[0]] = h[kIdxC[0]]
	h[kIdxC[0]] = h[kIdxB[0]]
	h[kIdxB[0]] = tA[0]

	tp = uint32(h[kIdxD[1]] + w[49] +
		((h[kIdxA[1]] & h[kIdxB[1]]) | ((h[kIdxA[1]] | h[kIdxB[1]]) & h[kIdxC[1]])))
	h[kIdxA[1]] = ((tp << p3) | (tp >> (32 - p3))) + tA[kPrems[isp+6]^1]
	h[kIdxD[1]] = h[kIdxC[1]]
	h[kIdxC[1]] = h[kIdxB[1]]
	h[kIdxB[1]] = tA[1]

	tp = uint32(h[kIdxD[2]] + w[50] +
		((h[kIdxA[2]] & h[kIdxB[2]]) | ((h[kIdxA[2]] | h[kIdxB[2]]) & h[kIdxC[2]])))
	h[kIdxA[2]] = ((tp << p3) | (tp >> (32 - p3))) + tA[kPrems[isp+6]^2]
	h[kIdxD[2]] = h[kIdxC[2]]
	h[kIdxC[2]] = h[kIdxB[2]]
	h[kIdxB[2]] = tA[2]

	tp = uint32(h[kIdxD[3]] + w[51] +
		((h[kIdxA[3]] & h[kIdxB[3]]) | ((h[kIdxA[3]] | h[kIdxB[3]]) & h[kIdxC[3]])))
	h[kIdxA[3]] = ((tp << p3) | (tp >> (32 - p3))) + tA[kPrems[isp+6]^3]
	h[kIdxD[3]] = h[kIdxC[3]]
	h[kIdxC[3]] = h[kIdxB[3]]
	h[kIdxB[3]] = tA[3]

	tp = uint32(h[kIdxD[4]] + w[52] +
		((h[kIdxA[4]] & h[kIdxB[4]]) | ((h[kIdxA[4]] | h[kIdxB[4]]) & h[kIdxC[4]])))
	h[kIdxA[4]] = ((tp << p3) | (tp >> (32 - p3))) + tA[kPrems[isp+6]^4]
	h[kIdxD[4]] = h[kIdxC[4]]
	h[kIdxC[4]] = h[kIdxB[4]]
	h[kIdxB[4]] = tA[4]

	tp = uint32(h[kIdxD[5]] + w[53] +
		((h[kIdxA[5]] & h[kIdxB[5]]) | ((h[kIdxA[5]] | h[kIdxB[5]]) & h[kIdxC[5]])))
	h[kIdxA[5]] = ((tp << p3) | (tp >> (32 - p3))) + tA[kPrems[isp+6]^5]
	h[kIdxD[5]] = h[kIdxC[5]]
	h[kIdxC[5]] = h[kIdxB[5]]
	h[kIdxB[5]] = tA[5]

	tp = uint32(h[kIdxD[6]] + w[54] +
		((h[kIdxA[6]] & h[kIdxB[6]]) | ((h[kIdxA[6]] | h[kIdxB[6]]) & h[kIdxC[6]])))
	h[kIdxA[6]] = ((tp << p3) | (tp >> (32 - p3))) + tA[kPrems[isp+6]^6]
	h[kIdxD[6]] = h[kIdxC[6]]
	h[kIdxC[6]] = h[kIdxB[6]]
	h[kIdxB[6]] = tA[6]

	tp = uint32(h[kIdxD[7]] + w[55] +
		((h[kIdxA[7]] & h[kIdxB[7]]) | ((h[kIdxA[7]] | h[kIdxB[7]]) & h[kIdxC[7]])))
	h[kIdxA[7]] = ((tp << p3) | (tp >> (32 - p3))) + tA[kPrems[isp+6]^7]
	h[kIdxD[7]] = h[kIdxC[7]]
	h[kIdxC[7]] = h[kIdxB[7]]
	h[kIdxB[7]] = tA[7]

	tA[0] = ((h[0] << p3) | (h[0] >> (32 - p3)))
	tA[1] = ((h[1] << p3) | (h[1] >> (32 - p3)))
	tA[2] = ((h[2] << p3) | (h[2] >> (32 - p3)))
	tA[3] = ((h[3] << p3) | (h[3] >> (32 - p3)))
	tA[4] = ((h[4] << p3) | (h[4] >> (32 - p3)))
	tA[5] = ((h[5] << p3) | (h[5] >> (32 - p3)))
	tA[6] = ((h[6] << p3) | (h[6] >> (32 - p3)))
	tA[7] = ((h[7] << p3) | (h[7] >> (32 - p3)))

	tp = uint32(h[kIdxD[0]] + w[56] +
		((h[kIdxA[0]] & h[kIdxB[0]]) | ((h[kIdxA[0]] | h[kIdxB[0]]) & h[kIdxC[0]])))
	h[kIdxA[0]] = ((tp << p0) | (tp >> (32 - p0))) + tA[kPrems[isp+7]]
	h[kIdxD[0]] = h[kIdxC[0]]
	h[kIdxC[0]] = h[kIdxB[0]]
	h[kIdxB[0]] = tA[0]

	tp = uint32(h[kIdxD[1]] + w[57] +
		((h[kIdxA[1]] & h[kIdxB[1]]) | ((h[kIdxA[1]] | h[kIdxB[1]]) & h[kIdxC[1]])))
	h[kIdxA[1]] = ((tp << p0) | (tp >> (32 - p0))) + tA[kPrems[isp+7]^1]
	h[kIdxD[1]] = h[kIdxC[1]]
	h[kIdxC[1]] = h[kIdxB[1]]
	h[kIdxB[1]] = tA[1]

	tp = uint32(h[kIdxD[2]] + w[58] +
		((h[kIdxA[2]] & h[kIdxB[2]]) | ((h[kIdxA[2]] | h[kIdxB[2]]) & h[kIdxC[2]])))
	h[kIdxA[2]] = ((tp << p0) | (tp >> (32 - p0))) + tA[kPrems[isp+7]^2]
	h[kIdxD[2]] = h[kIdxC[2]]
	h[kIdxC[2]] = h[kIdxB[2]]
	h[kIdxB[2]] = tA[2]

	tp = uint32(h[kIdxD[3]] + w[59] +
		((h[kIdxA[3]] & h[kIdxB[3]]) | ((h[kIdxA[3]] | h[kIdxB[3]]) & h[kIdxC[3]])))
	h[kIdxA[3]] = ((tp << p0) | (tp >> (32 - p0))) + tA[kPrems[isp+7]^3]
	h[kIdxD[3]] = h[kIdxC[3]]
	h[kIdxC[3]] = h[kIdxB[3]]
	h[kIdxB[3]] = tA[3]

	tp = uint32(h[kIdxD[4]] + w[60] +
		((h[kIdxA[4]] & h[kIdxB[4]]) | ((h[kIdxA[4]] | h[kIdxB[4]]) & h[kIdxC[4]])))
	h[kIdxA[4]] = ((tp << p0) | (tp >> (32 - p0))) + tA[kPrems[isp+7]^4]
	h[kIdxD[4]] = h[kIdxC[4]]
	h[kIdxC[4]] = h[kIdxB[4]]
	h[kIdxB[4]] = tA[4]

	tp = uint32(h[kIdxD[5]] + w[61] +
		((h[kIdxA[5]] & h[kIdxB[5]]) | ((h[kIdxA[5]] | h[kIdxB[5]]) & h[kIdxC[5]])))
	h[kIdxA[5]] = ((tp << p0) | (tp >> (32 - p0))) + tA[kPrems[isp+7]^5]
	h[kIdxD[5]] = h[kIdxC[5]]
	h[kIdxC[5]] = h[kIdxB[5]]
	h[kIdxB[5]] = tA[5]

	tp = uint32(h[kIdxD[6]] + w[62] +
		((h[kIdxA[6]] & h[kIdxB[6]]) | ((h[kIdxA[6]] | h[kIdxB[6]]) & h[kIdxC[6]])))
	h[kIdxA[6]] = ((tp << p0) | (tp >> (32 - p0))) + tA[kPrems[isp+7]^6]
	h[kIdxD[6]] = h[kIdxC[6]]
	h[kIdxC[6]] = h[kIdxB[6]]
	h[kIdxB[6]] = tA[6]

	tp = uint32(h[kIdxD[7]] + w[63] +
		((h[kIdxA[7]] & h[kIdxB[7]]) | ((h[kIdxA[7]] | h[kIdxB[7]]) & h[kIdxC[7]])))
	h[kIdxA[7]] = ((tp << p0) | (tp >> (32 - p0))) + tA[kPrems[isp+7]^7]
	h[kIdxD[7]] = h[kIdxC[7]]
	h[kIdxC[7]] = h[kIdxB[7]]
	h[kIdxB[7]] = tA[7]
}

////////////////

var kInit = []uint32{
	uint32(0x0BA16B95), uint32(0x72F999AD),
	uint32(0x9FECC2AE), uint32(0xBA3264FC),
	uint32(0x5E894929), uint32(0x8E9F30E5),
	uint32(0x2F1DAA37), uint32(0xF0F2C558),
	uint32(0xAC506643), uint32(0xA90635A5),
	uint32(0xE25B878B), uint32(0xAAB7878F),
	uint32(0x88817F7A), uint32(0x0A02892B),
	uint32(0x559A7550), uint32(0x598F657E),
	uint32(0x7EEF60A1), uint32(0x6B70E3E8),
	uint32(0x9C1714D1), uint32(0xB958E2A8),
	uint32(0xAB02675E), uint32(0xED1C014F),
	uint32(0xCD8D65BB), uint32(0xFDB7A257),
	uint32(0x09254899), uint32(0xD699C7BC),
	uint32(0x9019B6DC), uint32(0x2B9022E4),
	uint32(0x8FA14956), uint32(0x21BF9BD3),
	uint32(0xB94D0943), uint32(0x6FFDDC22),
}

var kIdxA = [8]uint8{
	0, 1, 2, 3, 4, 5, 6, 7,
}
var kIdxB = [8]uint8{
	8, 9, 10, 11, 12, 13, 14, 15,
}
var kIdxC = [8]uint8{
	16, 17, 18, 19, 20, 21, 22, 23,
}
var kIdxD = [8]uint8{
	24, 25, 26, 27, 28, 29, 30, 31,
}

var kPrems = []uint8{
	1, 6, 2, 3, 5, 7, 4, 1, 6, 2, 3,
}

var kPrem = [7][8]uint8{
	{1, 0, 3, 2, 5, 4, 7, 6},
	{6, 7, 4, 5, 2, 3, 0, 1},
	{2, 3, 0, 1, 6, 7, 4, 5},
	{3, 2, 1, 0, 7, 6, 5, 4},
	{5, 4, 7, 6, 1, 0, 3, 2},
	{7, 6, 5, 4, 3, 2, 1, 0},
	{4, 5, 6, 7, 0, 1, 2, 3},
}

var wbp = [32]uintptr{
	4 << 4, 6 << 4, 0 << 4, 2 << 4,
	7 << 4, 5 << 4, 3 << 4, 1 << 4,
	15 << 4, 11 << 4, 12 << 4, 8 << 4,
	9 << 4, 13 << 4, 10 << 4, 14 << 4,
	17 << 4, 18 << 4, 23 << 4, 20 << 4,
	22 << 4, 21 << 4, 16 << 4, 19 << 4,
	30 << 4, 24 << 4, 25 << 4, 31 << 4,
	27 << 4, 29 << 4, 28 << 4, 26 << 4,
}

var kAlphaTab = []int32{
	1, 41, 139, 45, 46, 87, 226, 14, 60, 147, 116, 130,
	190, 80, 196, 69, 2, 82, 21, 90, 92, 174, 195, 28,
	120, 37, 232, 3, 123, 160, 135, 138, 4, 164, 42, 180,
	184, 91, 133, 56, 240, 74, 207, 6, 246, 63, 13, 19,
	8, 71, 84, 103, 111, 182, 9, 112, 223, 148, 157, 12,
	235, 126, 26, 38, 16, 142, 168, 206, 222, 107, 18, 224,
	189, 39, 57, 24, 213, 252, 52, 76, 32, 27, 79, 155,
	187, 214, 36, 191, 121, 78, 114, 48, 169, 247, 104, 152,
	64, 54, 158, 53, 117, 171, 72, 125, 242, 156, 228, 96,
	81, 237, 208, 47, 128, 108, 59, 106, 234, 85, 144, 250,
	227, 55, 199, 192, 162, 217, 159, 94, 256, 216, 118, 212,
	211, 170, 31, 243, 197, 110, 141, 127, 67, 177, 61, 188,
	255, 175, 236, 167, 165, 83, 62, 229, 137, 220, 25, 254,
	134, 97, 122, 119, 253, 93, 215, 77, 73, 166, 124, 201,
	17, 183, 50, 251, 11, 194, 244, 238, 249, 186, 173, 154,
	146, 75, 248, 145, 34, 109, 100, 245, 22, 131, 231, 219,
	241, 115, 89, 51, 35, 150, 239, 33, 68, 218, 200, 233,
	44, 5, 205, 181, 225, 230, 178, 102, 70, 43, 221, 66,
	136, 179, 143, 209, 88, 10, 153, 105, 193, 203, 99, 204,
	140, 86, 185, 132, 15, 101, 29, 161, 176, 20, 49, 210,
	129, 149, 198, 151, 23, 172, 113, 7, 30, 202, 58, 65,
	95, 40, 98, 163,
}

var kYOffA = []int32{
	1, 163, 98, 40, 95, 65, 58, 202, 30, 7, 113, 172,
	23, 151, 198, 149, 129, 210, 49, 20, 176, 161, 29, 101,
	15, 132, 185, 86, 140, 204, 99, 203, 193, 105, 153, 10,
	88, 209, 143, 179, 136, 66, 221, 43, 70, 102, 178, 230,
	225, 181, 205, 5, 44, 233, 200, 218, 68, 33, 239, 150,
	35, 51, 89, 115, 241, 219, 231, 131, 22, 245, 100, 109,
	34, 145, 248, 75, 146, 154, 173, 186, 249, 238, 244, 194,
	11, 251, 50, 183, 17, 201, 124, 166, 73, 77, 215, 93,
	253, 119, 122, 97, 134, 254, 25, 220, 137, 229, 62, 83,
	165, 167, 236, 175, 255, 188, 61, 177, 67, 127, 141, 110,
	197, 243, 31, 170, 211, 212, 118, 216, 256, 94, 159, 217,
	162, 192, 199, 55, 227, 250, 144, 85, 234, 106, 59, 108,
	128, 47, 208, 237, 81, 96, 228, 156, 242, 125, 72, 171,
	117, 53, 158, 54, 64, 152, 104, 247, 169, 48, 114, 78,
	121, 191, 36, 214, 187, 155, 79, 27, 32, 76, 52, 252,
	213, 24, 57, 39, 189, 224, 18, 107, 222, 206, 168, 142,
	16, 38, 26, 126, 235, 12, 157, 148, 223, 112, 9, 182,
	111, 103, 84, 71, 8, 19, 13, 63, 246, 6, 207, 74,
	240, 56, 133, 91, 184, 180, 42, 164, 4, 138, 135, 160,
	123, 3, 232, 37, 120, 28, 195, 174, 92, 90, 21, 82,
	2, 69, 196, 80, 190, 130, 116, 147, 60, 14, 226, 87,
	46, 45, 139, 41,
}

var kYOffB = []int32{
	2, 203, 156, 47, 118, 214, 107, 106, 45, 93, 212, 20,
	111, 73, 162, 251, 97, 215, 249, 53, 211, 19, 3, 89,
	49, 207, 101, 67, 151, 130, 223, 23, 189, 202, 178, 239,
	253, 127, 204, 49, 76, 236, 82, 137, 232, 157, 65, 79,
	96, 161, 176, 130, 161, 30, 47, 9, 189, 247, 61, 226,
	248, 90, 107, 64, 0, 88, 131, 243, 133, 59, 113, 115,
	17, 236, 33, 213, 12, 191, 111, 19, 251, 61, 103, 208,
	57, 35, 148, 248, 47, 116, 65, 119, 249, 178, 143, 40,
	189, 129, 8, 163, 204, 227, 230, 196, 205, 122, 151, 45,
	187, 19, 227, 72, 247, 125, 111, 121, 140, 220, 6, 107,
	77, 69, 10, 101, 21, 65, 149, 171, 255, 54, 101, 210,
	139, 43, 150, 151, 212, 164, 45, 237, 146, 184, 95, 6,
	160, 42, 8, 204, 46, 238, 254, 168, 208, 50, 156, 190,
	106, 127, 34, 234, 68, 55, 79, 18, 4, 130, 53, 208,
	181, 21, 175, 120, 25, 100, 192, 178, 161, 96, 81, 127,
	96, 227, 210, 248, 68, 10, 196, 31, 9, 167, 150, 193,
	0, 169, 126, 14, 124, 198, 144, 142, 240, 21, 224, 44,
	245, 66, 146, 238, 6, 196, 154, 49, 200, 222, 109, 9,
	210, 141, 192, 138, 8, 79, 114, 217, 68, 128, 249, 94,
	53, 30, 27, 61, 52, 135, 106, 212, 70, 238, 30, 185,
	10, 132, 146, 136, 117, 37, 251, 150, 180, 188, 247, 156,
	236, 192, 108, 86,
}

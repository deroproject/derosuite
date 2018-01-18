// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package shavite

import (
	"fmt"

	"gitlab.com/nitya-sattva/go-x11/aesr"
	"gitlab.com/nitya-sattva/go-x11/hash"
)

// HashSize holds the size of a hash in bytes.
const HashSize = int(64)

// BlockSize holds the size of a block in bytes.
const BlockSize = uintptr(128)

////////////////

type digest struct {
	ptr uintptr

	h [16]uint32
	c [4]uint32

	b [BlockSize]byte
}

// New returns a new digest to compute a SHAVITE512 hash.
func New() hash.Digest {
	ref := &digest{}
	ref.Reset()
	return ref
}

////////////////

// Reset resets the digest to its initial state.
func (ref *digest) Reset() {
	ref.ptr = 0
	copy(ref.h[:], kInit[:])
	ref.c[0], ref.c[1] = 0, 0
	ref.c[2], ref.c[3] = 0, 0
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
	ptr := ref.ptr

	for sln > 0 {
		cln := BlockSize - ptr

		if cln > sln {
			cln = sln
		}
		sln -= cln

		copy(ref.b[ptr:], src[:cln])
		src = src[cln:]
		ptr += cln

		if ptr == BlockSize {
			ref.c[0] += 1024
			if ref.c[0] == 0 {
				ref.c[1] += 1
				if ref.c[1] == 0 {
					ref.c[2] += 1
					if ref.c[2] == 0 {
						ref.c[3] += 1
					}
				}
			}

			ref.compress()
			ptr = 0
		}
	}

	ref.ptr = ptr
	return fln, nil
}

// Close the digest by writing the last bits and storing the hash
// in dst. This prepares the digest for reuse by calling reset. A call
// to Close with a dst that is smaller then HashSize will return an error.
func (ref *digest) Close(dst []byte, bits uint8, bcnt uint8) error {
	if ln := len(dst); HashSize > ln {
		return fmt.Errorf("Shavite Close: dst min length: %d, got %d", HashSize, ln)
	}

	var cnt [4]uint32

	ptr := ref.ptr
	buf := ref.b[:]

	ref.c[0] += uint32(ptr<<3) + uint32(bcnt)
	cnt[0] = ref.c[0]
	cnt[1] = ref.c[1]
	cnt[2] = ref.c[2]
	cnt[3] = ref.c[3]

	z := uint8(0x80) >> bcnt
	z = uint8((bits & -z) | z)

	if (ptr == 0) && (bcnt == 0) {
		buf[0] = 0x80
		memset(buf[1:110], 0)
		ref.c[0], ref.c[1] = 0, 0
		ref.c[2], ref.c[3] = 0, 0
	} else if ptr < 110 {
		buf[ptr] = z
		ptr += 1
		memset(buf[ptr:(ptr+(110-ptr))], 0)
	} else {
		buf[ptr] = z
		ptr += 1
		memset(buf[ptr:(ptr+(128-ptr))], 0)

		ref.compress()
		memset(buf[:], 0)
		ref.c[0], ref.c[1] = 0, 0
		ref.c[2], ref.c[3] = 0, 0
	}

	encUInt32le(buf[110:], cnt[0])
	encUInt32le(buf[114:], cnt[1])
	encUInt32le(buf[118:], cnt[2])
	encUInt32le(buf[122:], cnt[3])

	buf[126] = 0
	buf[127] = 2

	ref.compress()

	for u := uintptr(0); u < 16; u++ {
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

func (ref *digest) compress() {
	var p0, p1, p2, p3, p4, p5, p6, p7 uint32
	var p8, p9, pA, pB, pC, pD, pE, pF uint32
	var t0, t1, t2, t3 uint32
	var rk [448]uint32

	for i := uintptr(0); i < 32; i++ {
		rk[i] = decUInt32le(ref.b[(i * 4):])
		//rk[i] = decUInt32be(ref.b[(i * 4):])
	}

	idx := uintptr(32)
	for idx < 448 {
		for s := uintptr(0); s < 4; s++ {
			t0 = rk[idx-31]
			t1 = rk[idx-30]
			t2 = rk[idx-29]
			t3 = rk[idx-32]
			t0, t1, t2, t3 = aesr.Round32sle(t0, t1, t2, t3)
			rk[idx+0] = t0 ^ rk[idx-4]
			rk[idx+1] = t1 ^ rk[idx-3]
			rk[idx+2] = t2 ^ rk[idx-2]
			rk[idx+3] = t3 ^ rk[idx-1]
			if idx == 32 {
				rk[32] ^= ref.c[0]
				rk[33] ^= ref.c[1]
				rk[34] ^= ref.c[2]
				rk[35] ^= uint32(^ref.c[3])
			} else if idx == 440 {
				rk[440] ^= ref.c[1]
				rk[441] ^= ref.c[0]
				rk[442] ^= ref.c[3]
				rk[443] ^= uint32(^ref.c[2])
			}
			idx += 4

			t0 = rk[idx-31]
			t1 = rk[idx-30]
			t2 = rk[idx-29]
			t3 = rk[idx-32]
			t0, t1, t2, t3 = aesr.Round32sle(t0, t1, t2, t3)
			rk[idx+0] = t0 ^ rk[idx-4]
			rk[idx+1] = t1 ^ rk[idx-3]
			rk[idx+2] = t2 ^ rk[idx-2]
			rk[idx+3] = t3 ^ rk[idx-1]
			if idx == 164 {
				rk[164] ^= ref.c[3]
				rk[165] ^= ref.c[2]
				rk[166] ^= ref.c[1]
				rk[167] ^= uint32(^ref.c[0])
			} else if idx == 316 {
				rk[316] ^= ref.c[2]
				rk[317] ^= ref.c[3]
				rk[318] ^= ref.c[0]
				rk[319] ^= uint32(^ref.c[1])
			}
			idx += 4
		}

		if idx != 448 {
			for s := uintptr(0); s < 8; s++ {
				rk[idx+0] = rk[idx-32] ^ rk[idx-7]
				rk[idx+1] = rk[idx-31] ^ rk[idx-6]
				rk[idx+2] = rk[idx-30] ^ rk[idx-5]
				rk[idx+3] = rk[idx-29] ^ rk[idx-4]
				idx += 4
			}
		}
	}

	p0 = ref.h[0x0]
	p1 = ref.h[0x1]
	p2 = ref.h[0x2]
	p3 = ref.h[0x3]
	p4 = ref.h[0x4]
	p5 = ref.h[0x5]
	p6 = ref.h[0x6]
	p7 = ref.h[0x7]
	p8 = ref.h[0x8]
	p9 = ref.h[0x9]
	pA = ref.h[0xA]
	pB = ref.h[0xB]
	pC = ref.h[0xC]
	pD = ref.h[0xD]
	pE = ref.h[0xE]
	pF = ref.h[0xF]

	idx = 0
	for r := uint32(0); r < 14; r++ {
		t0 = p4 ^ rk[idx+0]
		t1 = p5 ^ rk[idx+1]
		t2 = p6 ^ rk[idx+2]
		t3 = p7 ^ rk[idx+3]
		t0, t1, t2, t3 = aesr.Round32sle(t0, t1, t2, t3)
		t0 ^= rk[idx+4]
		t1 ^= rk[idx+5]
		t2 ^= rk[idx+6]
		t3 ^= rk[idx+7]
		t0, t1, t2, t3 = aesr.Round32sle(t0, t1, t2, t3)
		t0 ^= rk[idx+8]
		t1 ^= rk[idx+9]
		t2 ^= rk[idx+10]
		t3 ^= rk[idx+11]
		t0, t1, t2, t3 = aesr.Round32sle(t0, t1, t2, t3)
		t0 ^= rk[idx+12]
		t1 ^= rk[idx+13]
		t2 ^= rk[idx+14]
		t3 ^= rk[idx+15]
		t0, t1, t2, t3 = aesr.Round32sle(t0, t1, t2, t3)
		p0 ^= t0
		p1 ^= t1
		p2 ^= t2
		p3 ^= t3

		idx += 16

		t0 = pC ^ rk[idx+0]
		t1 = pD ^ rk[idx+1]
		t2 = pE ^ rk[idx+2]
		t3 = pF ^ rk[idx+3]
		t0, t1, t2, t3 = aesr.Round32sle(t0, t1, t2, t3)
		t0 ^= rk[idx+4]
		t1 ^= rk[idx+5]
		t2 ^= rk[idx+6]
		t3 ^= rk[idx+7]
		t0, t1, t2, t3 = aesr.Round32sle(t0, t1, t2, t3)
		t0 ^= rk[idx+8]
		t1 ^= rk[idx+9]
		t2 ^= rk[idx+10]
		t3 ^= rk[idx+11]
		t0, t1, t2, t3 = aesr.Round32sle(t0, t1, t2, t3)
		t0 ^= rk[idx+12]
		t1 ^= rk[idx+13]
		t2 ^= rk[idx+14]
		t3 ^= rk[idx+15]
		t0, t1, t2, t3 = aesr.Round32sle(t0, t1, t2, t3)
		p8 ^= t0
		p9 ^= t1
		pA ^= t2
		pB ^= t3

		idx += 16

		t0 = pC
		pC = p8
		p8 = p4
		p4 = p0
		p0 = t0

		t0 = pD
		pD = p9
		p9 = p5
		p5 = p1
		p1 = t0

		t0 = pE
		pE = pA
		pA = p6
		p6 = p2
		p2 = t0

		t0 = pF
		pF = pB
		pB = p7
		p7 = p3
		p3 = t0
	}

	ref.h[0x0] ^= p0
	ref.h[0x1] ^= p1
	ref.h[0x2] ^= p2
	ref.h[0x3] ^= p3
	ref.h[0x4] ^= p4
	ref.h[0x5] ^= p5
	ref.h[0x6] ^= p6
	ref.h[0x7] ^= p7
	ref.h[0x8] ^= p8
	ref.h[0x9] ^= p9
	ref.h[0xA] ^= pA
	ref.h[0xB] ^= pB
	ref.h[0xC] ^= pC
	ref.h[0xD] ^= pD
	ref.h[0xE] ^= pE
	ref.h[0xF] ^= pF
}

////////////////

var kInit = [16]uint32{
	uint32(0x72FCCDD8), uint32(0x79CA4727),
	uint32(0x128A077B), uint32(0x40D55AEC),
	uint32(0xD1901A06), uint32(0x430AE307),
	uint32(0xB29F5CD1), uint32(0xDF07FBFC),
	uint32(0x8E45D73D), uint32(0x681AB538),
	uint32(0xBDE86578), uint32(0xDD577E47),
	uint32(0xE275EADE), uint32(0x502D9FCD),
	uint32(0xB9357178), uint32(0x022A4B9A),
}

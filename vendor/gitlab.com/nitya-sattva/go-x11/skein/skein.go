// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package skein

import (
	"fmt"

	"gitlab.com/nitya-sattva/go-x11/hash"
)

// HashSize holds the size of a hash in bytes.
const HashSize = int(64)

// BlockSize holds the size of a block in bytes.
const BlockSize = uintptr(64)

////////////////

type digest struct {
	ptr uintptr
	cnt uint64

	h [8]uint64

	b [BlockSize]byte
}

// New returns a new digest to compute a BLAKE512 hash.
func New() hash.Digest {
	ref := &digest{}
	ref.Reset()
	return ref
}

////////////////

// Reset resets the digest to its initial state.
func (ref *digest) Reset() {
	ref.ptr, ref.cnt = 0, 0
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
	ptr := ref.ptr

	if sln <= (BlockSize - ptr) {
		copy(ref.b[ptr:], src)
		ref.ptr += sln
		return int(sln), nil
	}

	cnt := ref.cnt

	b := ref.b[:]
	h := [27]uint64{}
	h[0] = ref.h[0]
	h[1] = ref.h[1]
	h[2] = ref.h[2]
	h[3] = ref.h[3]
	h[4] = ref.h[4]
	h[5] = ref.h[5]
	h[6] = ref.h[6]
	h[7] = ref.h[7]

	var first uint8
	if cnt == 0 {
		first = uint8(1 << 7)
	}

	if ptr == BlockSize {
		cnt += 1
		compress(b, h[:], uint64(96+first), cnt, 0)
		first = 0
		ptr = 0
	}

	cln := BlockSize - ptr

	if cln > sln {
		cln = sln
	}
	sln -= cln

	copy(b[ptr:], src[:cln])
	src = src[cln:]
	ptr += cln

	for sln > 0 {
		if ptr == BlockSize {
			cnt += 1
			compress(b, h[:], uint64(96+first), cnt, 0)
			first = 0
			ptr = 0
		}

		cln := BlockSize - ptr

		if cln > sln {
			cln = sln
		}
		sln -= cln

		copy(b[ptr:], src[:cln])
		src = src[cln:]
		ptr += cln

	}

	ref.h[0] = h[0]
	ref.h[1] = h[1]
	ref.h[2] = h[2]
	ref.h[3] = h[3]
	ref.h[4] = h[4]
	ref.h[5] = h[5]
	ref.h[6] = h[6]
	ref.h[7] = h[7]

	ref.ptr = ptr
	ref.cnt = cnt
	return fln, nil
}

// Close the digest by writing the last bits and storing the hash
// in dst. This prepares the digest for reuse by calling reset. A call
// to Close with a dst that is smaller then HashSize will return an error.
func (ref *digest) Close(dst []byte, bits uint8, bcnt uint8) error {
	if ln := len(dst); HashSize > ln {
		return fmt.Errorf("Skein Close: dst min length: %d, got %d", HashSize, ln)
	}

	if bcnt != 0 {
		off := uint8(0x80) >> bcnt
		x := [1]uint8{uint8((bits & -off) | off)}
		ref.Write(x[:])
	}

	ptr := ref.ptr
	cnt := ref.cnt

	b := ref.b[:]
	h := [27]uint64{}
	h[0] = ref.h[0]
	h[1] = ref.h[1]
	h[2] = ref.h[2]
	h[3] = ref.h[3]
	h[4] = ref.h[4]
	h[5] = ref.h[5]
	h[6] = ref.h[6]
	h[7] = ref.h[7]

	memset(b[ptr:ptr+(BlockSize-ptr)], 0)

	var etv uint64
	if cnt == 0 {
		etv = 352 + uint64(1<<7)
	} else {
		etv = 352
	}

	if bcnt != 0 {
		etv += 1
	}

	for i := uintptr(0); i < 2; i++ {
		compress(b, h[:], etv, cnt, ptr)
		if i == 0 {
			memset(b[:], 0)
			etv = 510
			ptr = 8
			cnt = 0
		}
	}

	for u := uintptr(0); u < 64; u += 8 {
		encUInt64le(dst[u:], h[u>>3])
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

func decUInt64le(src []byte) uint64 {
	return (uint64(src[0]) |
		uint64(src[1])<<8 |
		uint64(src[2])<<16 |
		uint64(src[3])<<24 |
		uint64(src[4])<<32 |
		uint64(src[5])<<40 |
		uint64(src[6])<<48 |
		uint64(src[7])<<56)
}

func encUInt64le(dst []byte, src uint64) {
	dst[0] = uint8(src)
	dst[1] = uint8(src >> 8)
	dst[2] = uint8(src >> 16)
	dst[3] = uint8(src >> 24)
	dst[4] = uint8(src >> 32)
	dst[5] = uint8(src >> 40)
	dst[6] = uint8(src >> 48)
	dst[7] = uint8(src >> 56)
}

func compress(buf []uint8, h []uint64, etv, cnt uint64, ext uintptr) {
	var t0, t1, t2 uint64

	m0 := decUInt64le(buf[0:])
	m1 := decUInt64le(buf[8:])
	m2 := decUInt64le(buf[16:])
	m3 := decUInt64le(buf[24:])
	m4 := decUInt64le(buf[32:])
	m5 := decUInt64le(buf[40:])
	m6 := decUInt64le(buf[48:])
	m7 := decUInt64le(buf[56:])

	p0 := m0
	p1 := m1
	p2 := m2
	p3 := m3
	p4 := m4
	p5 := m5
	p6 := m6
	p7 := m7

	t0 = uint64(cnt<<6) + uint64(ext)
	t1 = (cnt >> 58) + (uint64(etv) << 55)
	t2 = t0 ^ t1

	h[8] = (((h[0] ^ h[1]) ^ (h[2] ^ h[3])) ^
		((h[4] ^ h[5]) ^ (h[6] ^ h[7])) ^
		uint64(0x1BD11BDAA9FC1A22))

	for u := uintptr(0); u <= 15; u += 3 {
		h[u+9] = h[u+0]
		h[u+10] = h[u+1]
		h[u+11] = h[u+2]
	}

	var tmp uint64
	for u := uintptr(0); u < 9; u++ {
		s := uint64(u << 1)

		p0 = uint64(p0 + h[s+0])
		p1 = uint64(p1 + h[s+1])
		p2 = uint64(p2 + h[s+2])
		p3 = uint64(p3 + h[s+3])
		p4 = uint64(p4 + h[s+4])
		p5 = uint64(p5 + h[s+5] + t0)
		p6 = uint64(p6 + h[s+6] + t1)
		p7 = uint64(p7 + h[s+7] + s)

		p0 = (p0 + p1)
		p1 = ((p1 << 46) | (p1 >> (64 - 46))) ^ p0
		p2 = (p2 + p3)
		p3 = ((p3 << 36) | (p3 >> (64 - 36))) ^ p2
		p4 = (p4 + p5)
		p5 = ((p5 << 19) | (p5 >> (64 - 19))) ^ p4
		p6 = (p6 + p7)
		p7 = ((p7 << 37) | (p7 >> (64 - 37))) ^ p6

		p2 = (p2 + p1)
		p1 = ((p1 << 33) | (p1 >> (64 - 33))) ^ p2
		p4 = (p4 + p7)
		p7 = ((p7 << 27) | (p7 >> (64 - 27))) ^ p4
		p6 = (p6 + p5)
		p5 = ((p5 << 14) | (p5 >> (64 - 14))) ^ p6
		p0 = (p0 + p3)
		p3 = ((p3 << 42) | (p3 >> (64 - 42))) ^ p0

		p4 = (p4 + p1)
		p1 = ((p1 << 17) | (p1 >> (64 - 17))) ^ p4
		p6 = (p6 + p3)
		p3 = ((p3 << 49) | (p3 >> (64 - 49))) ^ p6
		p0 = (p0 + p5)
		p5 = ((p5 << 36) | (p5 >> (64 - 36))) ^ p0
		p2 = (p2 + p7)
		p7 = ((p7 << 39) | (p7 >> (64 - 39))) ^ p2

		p6 = (p6 + p1)
		p1 = ((p1 << 44) | (p1 >> (64 - 44))) ^ p6
		p0 = (p0 + p7)
		p7 = ((p7 << 9) | (p7 >> (64 - 9))) ^ p0
		p2 = (p2 + p5)
		p5 = ((p5 << 54) | (p5 >> (64 - 54))) ^ p2
		p4 = (p4 + p3)
		p3 = ((p3 << 56) | (p3 >> (64 - 56))) ^ p4

		p0 = (p0 + h[s+1])
		p1 = (p1 + h[s+2])
		p2 = (p2 + h[s+3])
		p3 = (p3 + h[s+4])
		p4 = (p4 + h[s+5])
		p5 = (p5 + h[s+6] + t1)
		p6 = (p6 + h[s+7] + t2)
		p7 = (p7 + h[s+8] + (s + 1))

		p0 = (p0 + p1)
		p1 = ((p1 << 39) | (p1 >> (64 - 39))) ^ p0
		p2 = (p2 + p3)
		p3 = ((p3 << 30) | (p3 >> (64 - 30))) ^ p2
		p4 = (p4 + p5)
		p5 = ((p5 << 34) | (p5 >> (64 - 34))) ^ p4
		p6 = (p6 + p7)
		p7 = ((p7 << 24) | (p7 >> (64 - 24))) ^ p6

		p2 = (p2 + p1)
		p1 = ((p1 << 13) | (p1 >> (64 - 13))) ^ p2
		p4 = (p4 + p7)
		p7 = ((p7 << 50) | (p7 >> (64 - 50))) ^ p4
		p6 = (p6 + p5)
		p5 = ((p5 << 10) | (p5 >> (64 - 10))) ^ p6
		p0 = (p0 + p3)
		p3 = ((p3 << 17) | (p3 >> (64 - 17))) ^ p0

		p4 = (p4 + p1)
		p1 = ((p1 << 25) | (p1 >> (64 - 25))) ^ p4
		p6 = (p6 + p3)
		p3 = ((p3 << 29) | (p3 >> (64 - 29))) ^ p6
		p0 = (p0 + p5)
		p5 = ((p5 << 39) | (p5 >> (64 - 39))) ^ p0
		p2 = (p2 + p7)
		p7 = ((p7 << 43) | (p7 >> (64 - 43))) ^ p2

		p6 = (p6 + p1)
		p1 = ((p1 << 8) | (p1 >> (64 - 8))) ^ p6
		p0 = (p0 + p7)
		p7 = ((p7 << 35) | (p7 >> (64 - 35))) ^ p0
		p2 = (p2 + p5)
		p5 = ((p5 << 56) | (p5 >> (64 - 56))) ^ p2
		p4 = (p4 + p3)
		p3 = ((p3 << 22) | (p3 >> (64 - 22))) ^ p4

		tmp = t2
		t2 = t1
		t1 = t0
		t0 = tmp
	}

	p0 += h[18+0]
	p1 += h[18+1]
	p2 += h[18+2]
	p3 += h[18+3]
	p4 += h[18+4]
	p5 += h[18+5] + t0
	p6 += h[18+6] + t1
	p7 += h[18+7] + 18

	h[0] = m0 ^ p0
	h[1] = m1 ^ p1
	h[2] = m2 ^ p2
	h[3] = m3 ^ p3
	h[4] = m4 ^ p4
	h[5] = m5 ^ p5
	h[6] = m6 ^ p6
	h[7] = m7 ^ p7
}

////////////////

var kInit = [8]uint64{
	uint64(0x4903ADFF749C51CE), uint64(0x0D95DE399746DF03),
	uint64(0x8FD1934127C79BCE), uint64(0x9A255629FF352CB1),
	uint64(0x5DB62599DF6CA7B0), uint64(0xEABE394CA9D5C3F4),
	uint64(0x991112C71A75B523), uint64(0xAE18A40B660FCC33),
}

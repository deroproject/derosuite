// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package bmw

import (
	"fmt"

	"gitlab.com/nitya-sattva/go-x11/hash"
)

// HashSize holds the size of a hash in bytes.
const HashSize = uintptr(64)

// BlockSize holds the size of a block in bytes.
const BlockSize = uintptr(128)

////////////////

type digest struct {
	ptr uintptr
	cnt uint64

	h [16]uint64

	b [BlockSize]byte
}

// New returns a new digest compute a BMW512 hash.
func New() hash.Digest {
	ref := &digest{}
	ref.Reset()
	return ref
}

////////////////

// Reset resets the digest to its initial state.
func (ref *digest) Reset() {
	ref.ptr = 0
	ref.cnt = 0
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

	ht := [16]uint64{}
	h1 := ref.h[:]
	h2 := ht[:]

	ref.cnt += uint64(sln << 3)

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
			compress(ref.b[:], h1, h2)
			h1, h2 = h2[:], h1[:]
			ptr = 0
		}
	}

	copy(ref.h[:], h1[:])

	ref.ptr = ptr
	return fln, nil
}

// Close the digest by writing the last bits and storing the hash
// in dst. This prepares the digest for reuse by calling reset. A call
// to Close with a dst that is smaller then HashSize will return an error.
func (ref *digest) Close(dst []byte, bits uint8, bcnt uint8) error {
	if ln := len(dst); HashSize > uintptr(ln) {
		return fmt.Errorf("Bmw Close: dst min length: %d, got %d", HashSize, ln)
	}

	buf := ref.b[:]
	ptr := ref.ptr + 1

	{
		off := uint8(0x80) >> bcnt
		buf[ref.ptr] = uint8((bits & -off) | off)
	}

	var h1, h2 [16]uint64
	if ptr > (BlockSize - 8) {
		memset(buf[ptr:], 0)
		compress(buf, ref.h[:], h1[:])
		ref.h = h1
		ptr = 0
	}

	memset(buf[ptr:(BlockSize-8)], 0)
	encUInt64le(buf[BlockSize-8:], ref.cnt+uint64(bcnt))

	compress(buf, ref.h[:], h2[:])
	for u := uint8(0); u < 16; u++ {
		encUInt64le(buf[(u*8):], h2[u])
	}

	compress(buf, kFinal[:], h1[:])
	for u := uint8(0); u < 8; u++ {
		encUInt64le(dst[(u*8):], h1[8+u])
	}

	ref.Reset()
	return nil
}

// Size returns the number of bytes required to store the hash.
func (*digest) Size() int {
	return int(HashSize)
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

func compress(src []uint8, hv, dh []uint64) {
	var xl, xh uint64

	mv := [16]uint64{}
	qt := [32]uint64{}
	mv[0x0] = decUInt64le(src[0:])
	mv[0x1] = decUInt64le(src[8:])
	mv[0x2] = decUInt64le(src[16:])
	mv[0x3] = decUInt64le(src[24:])
	mv[0x4] = decUInt64le(src[32:])
	mv[0x5] = decUInt64le(src[40:])
	mv[0x6] = decUInt64le(src[48:])
	mv[0x7] = decUInt64le(src[56:])
	mv[0x8] = decUInt64le(src[64:])
	mv[0x9] = decUInt64le(src[72:])
	mv[0xA] = decUInt64le(src[80:])
	mv[0xB] = decUInt64le(src[88:])
	mv[0xC] = decUInt64le(src[96:])
	mv[0xD] = decUInt64le(src[104:])
	mv[0xE] = decUInt64le(src[112:])
	mv[0xF] = decUInt64le(src[120:])

	{
		xv := [16]uint64{}
		xv[0x0] = mv[0x0] ^ hv[0x0]
		xv[0x1] = mv[0x1] ^ hv[0x1]
		xv[0x2] = mv[0x2] ^ hv[0x2]
		xv[0x3] = mv[0x3] ^ hv[0x3]
		xv[0x4] = mv[0x4] ^ hv[0x4]
		xv[0x5] = mv[0x5] ^ hv[0x5]
		xv[0x6] = mv[0x6] ^ hv[0x6]
		xv[0x7] = mv[0x7] ^ hv[0x7]
		xv[0x8] = mv[0x8] ^ hv[0x8]
		xv[0x9] = mv[0x9] ^ hv[0x9]
		xv[0xA] = mv[0xA] ^ hv[0xA]
		xv[0xB] = mv[0xB] ^ hv[0xB]
		xv[0xC] = mv[0xC] ^ hv[0xC]
		xv[0xD] = mv[0xD] ^ hv[0xD]
		xv[0xE] = mv[0xE] ^ hv[0xE]
		xv[0xF] = mv[0xF] ^ hv[0xF]

		qt[0] = hv[0x1] + shiftBits0(
			xv[0x5]-xv[0x7]+xv[0xA]+xv[0xD]+xv[0xE],
		)
		qt[1] = hv[0x2] + shiftBits1(
			xv[0x6]-xv[0x8]+xv[0xB]+xv[0xE]-xv[0xF],
		)
		qt[2] = hv[0x3] + shiftBits2(
			xv[0x0]+xv[0x7]+xv[0x9]-xv[0xC]+xv[0xF],
		)
		qt[3] = hv[0x4] + shiftBits3(
			xv[0x0]-xv[0x1]+xv[0x8]-xv[0xA]+xv[0xD],
		)
		qt[4] = hv[0x5] + shiftBits4(
			xv[0x1]+xv[0x2]+xv[0x9]-xv[0xB]-xv[0xE],
		)
		qt[5] = hv[0x6] + shiftBits0(
			xv[0x3]-xv[0x2]+xv[0xA]-xv[0xC]+xv[0xF],
		)
		qt[6] = hv[0x7] + shiftBits1(
			xv[0x4]-xv[0x0]-xv[0x3]-xv[0xB]+xv[0xD],
		)
		qt[7] = hv[0x8] + shiftBits2(
			xv[0x1]-xv[0x4]-xv[0x5]-xv[0xC]-xv[0xE],
		)
		qt[8] = hv[0x9] + shiftBits3(
			xv[0x2]-xv[0x5]-xv[0x6]+xv[0xD]-xv[0xF],
		)
		qt[9] = hv[0xA] + shiftBits4(
			xv[0x0]-xv[0x3]+xv[0x6]-xv[0x7]+xv[0xE],
		)
		qt[10] = hv[0xB] + shiftBits0(
			xv[0x8]-xv[0x1]-xv[0x4]-xv[0x7]+xv[0xF],
		)
		qt[11] = hv[0xC] + shiftBits1(
			xv[0x8]-xv[0x0]-xv[0x2]-xv[0x5]+xv[0x9],
		)
		qt[12] = hv[0xD] + shiftBits2(
			xv[0x1]+xv[0x3]-xv[0x6]-xv[0x9]+xv[0xA],
		)
		qt[13] = hv[0xE] + shiftBits3(
			xv[0x2]+xv[0x4]+xv[0x7]+xv[0xA]+xv[0xB],
		)
		qt[14] = hv[0xF] + shiftBits4(
			xv[0x3]-xv[0x5]+xv[0x8]-xv[0xB]-xv[0xC],
		)
		qt[15] = hv[0x0] + shiftBits0(
			xv[0xC]-xv[0x4]-xv[0x6]-xv[0x9]+xv[0xD],
		)
	}

	qt[16] = expandOne(16, qt[:], mv[:], hv[:])
	qt[17] = expandOne(17, qt[:], mv[:], hv[:])
	qt[18] = expandTwo(18, qt[:], mv[:], hv[:])
	qt[19] = expandTwo(19, qt[:], mv[:], hv[:])
	qt[20] = expandTwo(20, qt[:], mv[:], hv[:])
	qt[21] = expandTwo(21, qt[:], mv[:], hv[:])
	qt[22] = expandTwo(22, qt[:], mv[:], hv[:])
	qt[23] = expandTwo(23, qt[:], mv[:], hv[:])
	qt[24] = expandTwo(24, qt[:], mv[:], hv[:])
	qt[25] = expandTwo(25, qt[:], mv[:], hv[:])
	qt[26] = expandTwo(26, qt[:], mv[:], hv[:])
	qt[27] = expandTwo(27, qt[:], mv[:], hv[:])
	qt[28] = expandTwo(28, qt[:], mv[:], hv[:])
	qt[29] = expandTwo(29, qt[:], mv[:], hv[:])
	qt[30] = expandTwo(30, qt[:], mv[:], hv[:])
	qt[31] = expandTwo(31, qt[:], mv[:], hv[:])

	xl = qt[16] ^ qt[17] ^ qt[18] ^ qt[19] ^ qt[20] ^ qt[21] ^ qt[22] ^ qt[23]
	xh = xl ^ qt[24] ^ qt[25] ^ qt[26] ^ qt[27] ^ qt[28] ^ qt[29] ^ qt[30] ^ qt[31]

	dh[0x0] = ((xh << 5) ^ (qt[16] >> 5) ^ mv[0x0]) + (xl ^ qt[24] ^ qt[0])
	dh[0x1] = ((xh >> 7) ^ (qt[17] << 8) ^ mv[0x1]) + (xl ^ qt[25] ^ qt[1])
	dh[0x2] = ((xh >> 5) ^ (qt[18] << 5) ^ mv[0x2]) + (xl ^ qt[26] ^ qt[2])
	dh[0x3] = ((xh >> 1) ^ (qt[19] << 5) ^ mv[0x3]) + (xl ^ qt[27] ^ qt[3])
	dh[0x4] = ((xh >> 3) ^ (qt[20] << 0) ^ mv[0x4]) + (xl ^ qt[28] ^ qt[4])
	dh[0x5] = ((xh << 6) ^ (qt[21] >> 6) ^ mv[0x5]) + (xl ^ qt[29] ^ qt[5])
	dh[0x6] = ((xh >> 4) ^ (qt[22] << 6) ^ mv[0x6]) + (xl ^ qt[30] ^ qt[6])
	dh[0x7] = ((xh >> 11) ^ (qt[23] << 2) ^ mv[0x7]) + (xl ^ qt[31] ^ qt[7])

	dh[0x8] = ((dh[0x4] << 9) | (dh[0x4] >> (64 - 9)))
	dh[0x8] += (xh ^ qt[24] ^ mv[0x8]) + ((xl << 8) ^ qt[23] ^ qt[8])
	dh[0x9] = ((dh[0x5] << 10) | (dh[0x5] >> (64 - 10)))
	dh[0x9] += (xh ^ qt[25] ^ mv[0x9]) + ((xl >> 6) ^ qt[16] ^ qt[9])
	dh[0xA] = ((dh[0x6] << 11) | (dh[0x6] >> (64 - 11)))
	dh[0xA] += (xh ^ qt[26] ^ mv[0xA]) + ((xl << 6) ^ qt[17] ^ qt[10])
	dh[0xB] = ((dh[0x7] << 12) | (dh[0x7] >> (64 - 12)))
	dh[0xB] += (xh ^ qt[27] ^ mv[0xB]) + ((xl << 4) ^ qt[18] ^ qt[11])
	dh[0xC] = ((dh[0x0] << 13) | (dh[0x0] >> (64 - 13)))
	dh[0xC] += (xh ^ qt[28] ^ mv[0xC]) + ((xl >> 3) ^ qt[19] ^ qt[12])
	dh[0xD] = ((dh[0x1] << 14) | (dh[0x1] >> (64 - 14)))
	dh[0xD] += (xh ^ qt[29] ^ mv[0xD]) + ((xl >> 4) ^ qt[20] ^ qt[13])
	dh[0xE] = ((dh[0x2] << 15) | (dh[0x2] >> (64 - 15)))
	dh[0xE] += (xh ^ qt[30] ^ mv[0xE]) + ((xl >> 7) ^ qt[21] ^ qt[14])
	dh[0xF] = ((dh[0x3] << 16) | (dh[0x3] >> (64 - 16)))
	dh[0xF] += (xh ^ qt[31] ^ mv[0xF]) + ((xl >> 2) ^ qt[22] ^ qt[15])
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

func shiftBits0(x uint64) uint64 {
	return ((x >> 1) ^ (x << 3) ^
		((x << 4) | (x >> (64 - 4))) ^
		((x << 37) | (x >> (64 - 37))))
}

func shiftBits1(x uint64) uint64 {
	return ((x >> 1) ^ (x << 2) ^
		((x << 13) | (x >> (64 - 13))) ^
		((x << 43) | (x >> (64 - 43))))
}

func shiftBits2(x uint64) uint64 {
	return ((x >> 2) ^ (x << 1) ^
		((x << 19) | (x >> (64 - 19))) ^
		((x << 53) | (x >> (64 - 53))))
}

func shiftBits3(x uint64) uint64 {
	return ((x >> 2) ^ (x << 2) ^
		((x << 28) | (x >> (64 - 28))) ^
		((x << 59) | (x >> (64 - 59))))
}

func shiftBits4(x uint64) uint64 {
	return ((x >> 1) ^ x)
}

func shiftBits5(x uint64) uint64 {
	return ((x >> 2) ^ x)
}

func rolBits(idx, off uint8, mv []uint64) uint64 {
	x := mv[(idx+off)&15]
	n := uint64((idx+off)&15) + 1
	return (x << n) | (x >> (64 - n))
}

func addEltBits(idx uint8, mv, hv []uint64) uint64 {
	kbt := uint64(idx+16) * uint64(0x0555555555555555)
	return ((rolBits(idx, 0, mv) + rolBits(idx, 3, mv) -
		rolBits(idx, 10, mv) + kbt) ^ (hv[(idx+7)&15]))
}

func expandOne(idx uint8, qt, mv, hv []uint64) uint64 {
	return (shiftBits1(qt[idx-0x10]) + shiftBits2(qt[idx-0x0F]) +
		shiftBits3(qt[idx-0x0E]) + shiftBits0(qt[idx-0x0D]) +
		shiftBits1(qt[idx-0x0C]) + shiftBits2(qt[idx-0x0B]) +
		shiftBits3(qt[idx-0x0A]) + shiftBits0(qt[idx-0x09]) +
		shiftBits1(qt[idx-0x08]) + shiftBits2(qt[idx-0x07]) +
		shiftBits3(qt[idx-0x06]) + shiftBits0(qt[idx-0x05]) +
		shiftBits1(qt[idx-0x04]) + shiftBits2(qt[idx-0x03]) +
		shiftBits3(qt[idx-0x02]) + shiftBits0(qt[idx-0x01]) +
		addEltBits(uint8(idx-16), mv, hv))
}

func expandTwo(idx uint8, qt, mv, hv []uint64) uint64 {
	return (qt[idx-0x10] + ((qt[idx-0x0F] << 5) | (qt[idx-0x0F] >> (64 - 5))) +
		qt[idx-0x0E] + ((qt[idx-0x0D] << 11) | (qt[idx-0x0D] >> (64 - 11))) +
		qt[idx-0x0C] + ((qt[idx-0x0B] << 27) | (qt[idx-0x0B] >> (64 - 27))) +
		qt[idx-0x0A] + ((qt[idx-0x09] << 32) | (qt[idx-0x09] >> (64 - 32))) +
		qt[idx-0x08] + ((qt[idx-0x07] << 37) | (qt[idx-0x07] >> (64 - 37))) +
		qt[idx-0x06] + ((qt[idx-0x05] << 43) | (qt[idx-0x05] >> (64 - 43))) +
		qt[idx-0x04] + ((qt[idx-0x03] << 53) | (qt[idx-0x03] >> (64 - 53))) +
		shiftBits4(qt[idx-0x02]) + shiftBits5(qt[idx-0x01]) +
		addEltBits(uint8(idx-16), mv, hv))
}

////////////////

var kInit = [16]uint64{
	uint64(0x8081828384858687), uint64(0x88898A8B8C8D8E8F),
	uint64(0x9091929394959697), uint64(0x98999A9B9C9D9E9F),
	uint64(0xA0A1A2A3A4A5A6A7), uint64(0xA8A9AAABACADAEAF),
	uint64(0xB0B1B2B3B4B5B6B7), uint64(0xB8B9BABBBCBDBEBF),
	uint64(0xC0C1C2C3C4C5C6C7), uint64(0xC8C9CACBCCCDCECF),
	uint64(0xD0D1D2D3D4D5D6D7), uint64(0xD8D9DADBDCDDDEDF),
	uint64(0xE0E1E2E3E4E5E6E7), uint64(0xE8E9EAEBECEDEEEF),
	uint64(0xF0F1F2F3F4F5F6F7), uint64(0xF8F9FAFBFCFDFEFF),
}

var kFinal = [16]uint64{
	uint64(0xaaaaaaaaaaaaaaa0), uint64(0xaaaaaaaaaaaaaaa1),
	uint64(0xaaaaaaaaaaaaaaa2), uint64(0xaaaaaaaaaaaaaaa3),
	uint64(0xaaaaaaaaaaaaaaa4), uint64(0xaaaaaaaaaaaaaaa5),
	uint64(0xaaaaaaaaaaaaaaa6), uint64(0xaaaaaaaaaaaaaaa7),
	uint64(0xaaaaaaaaaaaaaaa8), uint64(0xaaaaaaaaaaaaaaa9),
	uint64(0xaaaaaaaaaaaaaaaa), uint64(0xaaaaaaaaaaaaaaab),
	uint64(0xaaaaaaaaaaaaaaac), uint64(0xaaaaaaaaaaaaaaad),
	uint64(0xaaaaaaaaaaaaaaae), uint64(0xaaaaaaaaaaaaaaaf),
}

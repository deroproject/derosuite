// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blake

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

	h [8]uint64
	t [2]uint64

	b [BlockSize]byte
}

// New returns a new digest compute a BLAKE512 hash.
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
	ref.t[0], ref.t[1] = 0, 0
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

	if sln < (BlockSize - ptr) {
		copy(ref.b[ptr:], src)
		ref.ptr += sln
		return int(sln), nil
	}

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
			h := ref.h[:]
			t := ref.t[:]

			t[0] = t[0] + 1024
			if t[0] < 1024 {
				t[1] += 1
			}
			ptr = 0

			var vec [16]uint64
			vec[0x0] = h[0]
			vec[0x1] = h[1]
			vec[0x2] = h[2]
			vec[0x3] = h[3]
			vec[0x4] = h[4]
			vec[0x5] = h[5]
			vec[0x6] = h[6]
			vec[0x7] = h[7]
			vec[0x8] = kSpec[0]
			vec[0x9] = kSpec[1]
			vec[0xA] = kSpec[2]
			vec[0xB] = kSpec[3]
			vec[0xC] = t[0] ^ kSpec[4]
			vec[0xD] = t[0] ^ kSpec[5]
			vec[0xE] = t[1] ^ kSpec[6]
			vec[0xF] = t[1] ^ kSpec[7]

			var mat [16]uint64
			mat[0x0] = decUInt64be(ref.b[0:])
			mat[0x1] = decUInt64be(ref.b[8:])
			mat[0x2] = decUInt64be(ref.b[16:])
			mat[0x3] = decUInt64be(ref.b[24:])
			mat[0x4] = decUInt64be(ref.b[32:])
			mat[0x5] = decUInt64be(ref.b[40:])
			mat[0x6] = decUInt64be(ref.b[48:])
			mat[0x7] = decUInt64be(ref.b[56:])
			mat[0x8] = decUInt64be(ref.b[64:])
			mat[0x9] = decUInt64be(ref.b[72:])
			mat[0xA] = decUInt64be(ref.b[80:])
			mat[0xB] = decUInt64be(ref.b[88:])
			mat[0xC] = decUInt64be(ref.b[96:])
			mat[0xD] = decUInt64be(ref.b[104:])
			mat[0xE] = decUInt64be(ref.b[112:])
			mat[0xF] = decUInt64be(ref.b[120:])

			var sig []uint8
			for r := uint8(0); r < 16; r++ {
				sig = kSigma[r][:]

				vec[0x0] = (vec[0x0] + vec[0x4] + (mat[sig[0x0]] ^ kSpec[sig[0x1]]))
				vec[0xC] = (((vec[0xC] ^ vec[0x0]) << (32)) | ((vec[0xC] ^ vec[0x0]) >> 32))
				vec[0x8] = (vec[0x8] + vec[0xC])
				vec[0x4] = (((vec[0x4] ^ vec[0x8]) << (39)) | ((vec[0x4] ^ vec[0x8]) >> 25))
				vec[0x0] = (vec[0x0] + vec[0x4] + (mat[sig[0x1]] ^ kSpec[sig[0x0]]))
				vec[0xC] = (((vec[0xC] ^ vec[0x0]) << (48)) | ((vec[0xC] ^ vec[0x0]) >> 16))
				vec[0x8] = (vec[0x8] + vec[0xC])
				vec[0x4] = (((vec[0x4] ^ vec[0x8]) << (53)) | ((vec[0x4] ^ vec[0x8]) >> 11))

				vec[0x1] = (vec[0x1] + vec[0x5] + (mat[sig[0x2]] ^ kSpec[sig[0x3]]))
				vec[0xD] = (((vec[0xD] ^ vec[0x1]) << (32)) | ((vec[0xD] ^ vec[0x1]) >> 32))
				vec[0x9] = (vec[0x9] + vec[0xD])
				vec[0x5] = (((vec[0x5] ^ vec[0x9]) << (39)) | ((vec[0x5] ^ vec[0x9]) >> 25))
				vec[0x1] = (vec[0x1] + vec[0x5] + (mat[sig[0x3]] ^ kSpec[sig[0x2]]))
				vec[0xD] = (((vec[0xD] ^ vec[0x1]) << (48)) | ((vec[0xD] ^ vec[0x1]) >> 16))
				vec[0x9] = (vec[0x9] + vec[0xD])
				vec[0x5] = (((vec[0x5] ^ vec[0x9]) << (53)) | ((vec[0x5] ^ vec[0x9]) >> 11))

				vec[0x2] = (vec[0x2] + vec[0x6] + (mat[sig[0x4]] ^ kSpec[sig[0x5]]))
				vec[0xE] = (((vec[0xE] ^ vec[0x2]) << (32)) | ((vec[0xE] ^ vec[0x2]) >> 32))
				vec[0xA] = (vec[0xA] + vec[0xE])
				vec[0x6] = (((vec[0x6] ^ vec[0xA]) << (39)) | ((vec[0x6] ^ vec[0xA]) >> 25))
				vec[0x2] = (vec[0x2] + vec[0x6] + (mat[sig[0x5]] ^ kSpec[sig[0x4]]))
				vec[0xE] = (((vec[0xE] ^ vec[0x2]) << (48)) | ((vec[0xE] ^ vec[0x2]) >> 16))
				vec[0xA] = (vec[0xA] + vec[0xE])
				vec[0x6] = (((vec[0x6] ^ vec[0xA]) << (53)) | ((vec[0x6] ^ vec[0xA]) >> 11))

				vec[0x3] = (vec[0x3] + vec[0x7] + (mat[sig[0x6]] ^ kSpec[sig[0x7]]))
				vec[0xF] = (((vec[0xF] ^ vec[0x3]) << (32)) | ((vec[0xF] ^ vec[0x3]) >> 32))
				vec[0xB] = (vec[0xB] + vec[0xF])
				vec[0x7] = (((vec[0x7] ^ vec[0xB]) << (39)) | ((vec[0x7] ^ vec[0xB]) >> 25))
				vec[0x3] = (vec[0x3] + vec[0x7] + (mat[sig[0x7]] ^ kSpec[sig[0x6]]))
				vec[0xF] = (((vec[0xF] ^ vec[0x3]) << (48)) | ((vec[0xF] ^ vec[0x3]) >> 16))
				vec[0xB] = (vec[0xB] + vec[0xF])
				vec[0x7] = (((vec[0x7] ^ vec[0xB]) << (53)) | ((vec[0x7] ^ vec[0xB]) >> 11))

				vec[0x0] = (vec[0x0] + vec[0x5] + (mat[sig[0x8]] ^ kSpec[sig[0x9]]))
				vec[0xF] = (((vec[0xF] ^ vec[0x0]) << (32)) | ((vec[0xF] ^ vec[0x0]) >> 32))
				vec[0xA] = (vec[0xA] + vec[0xF])
				vec[0x5] = (((vec[0x5] ^ vec[0xA]) << (39)) | ((vec[0x5] ^ vec[0xA]) >> 25))
				vec[0x0] = (vec[0x0] + vec[0x5] + (mat[sig[0x9]] ^ kSpec[sig[0x8]]))
				vec[0xF] = (((vec[0xF] ^ vec[0x0]) << (48)) | ((vec[0xF] ^ vec[0x0]) >> 16))
				vec[0xA] = (vec[0xA] + vec[0xF])
				vec[0x5] = (((vec[0x5] ^ vec[0xA]) << (53)) | ((vec[0x5] ^ vec[0xA]) >> 11))

				vec[0x1] = (vec[0x1] + vec[0x6] + (mat[sig[0xA]] ^ kSpec[sig[0xB]]))
				vec[0xC] = (((vec[0xC] ^ vec[0x1]) << (32)) | ((vec[0xC] ^ vec[0x1]) >> 32))
				vec[0xB] = (vec[0xB] + vec[0xC])
				vec[0x6] = (((vec[0x6] ^ vec[0xB]) << (39)) | ((vec[0x6] ^ vec[0xB]) >> 25))
				vec[0x1] = (vec[0x1] + vec[0x6] + (mat[sig[0xB]] ^ kSpec[sig[0xA]]))
				vec[0xC] = (((vec[0xC] ^ vec[0x1]) << (48)) | ((vec[0xC] ^ vec[0x1]) >> 16))
				vec[0xB] = (vec[0xB] + vec[0xC])
				vec[0x6] = (((vec[0x6] ^ vec[0xB]) << (53)) | ((vec[0x6] ^ vec[0xB]) >> 11))

				vec[0x2] = (vec[0x2] + vec[0x7] + (mat[sig[0xC]] ^ kSpec[sig[0xD]]))
				vec[0xD] = (((vec[0xD] ^ vec[0x2]) << (32)) | ((vec[0xD] ^ vec[0x2]) >> 32))
				vec[0x8] = (vec[0x8] + vec[0xD])
				vec[0x7] = (((vec[0x7] ^ vec[0x8]) << (39)) | ((vec[0x7] ^ vec[0x8]) >> 25))
				vec[0x2] = (vec[0x2] + vec[0x7] + (mat[sig[0xD]] ^ kSpec[sig[0xC]]))
				vec[0xD] = (((vec[0xD] ^ vec[0x2]) << (48)) | ((vec[0xD] ^ vec[0x2]) >> 16))
				vec[0x8] = (vec[0x8] + vec[0xD])
				vec[0x7] = (((vec[0x7] ^ vec[0x8]) << (53)) | ((vec[0x7] ^ vec[0x8]) >> 11))

				vec[0x3] = (vec[0x3] + vec[0x4] + (mat[sig[0xE]] ^ kSpec[sig[0xF]]))
				vec[0xE] = (((vec[0xE] ^ vec[0x3]) << (32)) | ((vec[0xE] ^ vec[0x3]) >> 32))
				vec[0x9] = (vec[0x9] + vec[0xE])
				vec[0x4] = (((vec[0x4] ^ vec[0x9]) << (39)) | ((vec[0x4] ^ vec[0x9]) >> 25))
				vec[0x3] = (vec[0x3] + vec[0x4] + (mat[sig[0xF]] ^ kSpec[sig[0xE]]))
				vec[0xE] = (((vec[0xE] ^ vec[0x3]) << (48)) | ((vec[0xE] ^ vec[0x3]) >> 16))
				vec[0x9] = (vec[0x9] + vec[0xE])
				vec[0x4] = (((vec[0x4] ^ vec[0x9]) << (53)) | ((vec[0x4] ^ vec[0x9]) >> 11))
			}

			h[0] ^= vec[0x0] ^ vec[0x8]
			h[1] ^= vec[0x1] ^ vec[0x9]
			h[2] ^= vec[0x2] ^ vec[0xA]
			h[3] ^= vec[0x3] ^ vec[0xB]
			h[4] ^= vec[0x4] ^ vec[0xC]
			h[5] ^= vec[0x5] ^ vec[0xD]
			h[6] ^= vec[0x6] ^ vec[0xE]
			h[7] ^= vec[0x7] ^ vec[0xF]
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
		return fmt.Errorf("Blake Close: dst min length: %d, got %d", HashSize, ln)
	}

	ptr := ref.ptr
	bln := uint64((ref.ptr << 3) + uintptr(bcnt))

	tpl := ref.t[0] + bln
	tph := ref.t[1]

	var buf [BlockSize]uint8

	{
		off := uint8(0x80) >> bcnt
		buf[ptr] = uint8((bits & -off) | off)
	}

	if ptr == 0 && bcnt == 0 {
		ref.t[0] = uint64(0xFFFFFFFFFFFFFC00)
		ref.t[1] = uint64(0xFFFFFFFFFFFFFFFF)
	} else if ref.t[0] == 0 {
		ref.t[0] = uint64(0xFFFFFFFFFFFFFC00) + bln
		ref.t[1] = uint64(ref.t[1] - 1)
	} else {
		ref.t[0] -= 1024 - bln
	}

	if bln <= 894 {
		buf[111] |= 1

		encUInt64be(buf[112:], tph)
		encUInt64be(buf[120:], tpl)
		ref.Write(buf[ptr:])
	} else {
		ref.Write(buf[ptr:])

		ref.t[0] = uint64(0xFFFFFFFFFFFFFC00)
		ref.t[1] = uint64(0xFFFFFFFFFFFFFFFF)

		memset(buf[:112], 0)
		buf[111] = 1

		encUInt64be(buf[112:], tph)
		encUInt64be(buf[120:], tpl)
		ref.Write(buf[:])
	}

	for k := uintptr(0); k < 8; k++ {
		encUInt64be(dst[(k<<3):], ref.h[k])
	}

	ref.Reset()
	return nil
}

// Size returns the number of bytes Sum will return.
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

func decUInt64be(src []byte) uint64 {
	return (uint64(src[0])<<56 |
		uint64(src[1])<<48 |
		uint64(src[2])<<40 |
		uint64(src[3])<<32 |
		uint64(src[4])<<24 |
		uint64(src[5])<<16 |
		uint64(src[6])<<8 |
		uint64(src[7]))
}

func encUInt64be(dst []byte, src uint64) {
	dst[0] = uint8(src >> 56)
	dst[1] = uint8(src >> 48)
	dst[2] = uint8(src >> 40)
	dst[3] = uint8(src >> 32)
	dst[4] = uint8(src >> 24)
	dst[5] = uint8(src >> 16)
	dst[6] = uint8(src >> 8)
	dst[7] = uint8(src)
}

////////////////

var kInit = [8]uint64{
	uint64(0x6A09E667F3BCC908), uint64(0xBB67AE8584CAA73B),
	uint64(0x3C6EF372FE94F82B), uint64(0xA54FF53A5F1D36F1),
	uint64(0x510E527FADE682D1), uint64(0x9B05688C2B3E6C1F),
	uint64(0x1F83D9ABFB41BD6B), uint64(0x5BE0CD19137E2179),
}

var kSpec = [16]uint64{
	uint64(0x243F6A8885A308D3), uint64(0x13198A2E03707344),
	uint64(0xA4093822299F31D0), uint64(0x082EFA98EC4E6C89),
	uint64(0x452821E638D01377), uint64(0xBE5466CF34E90C6C),
	uint64(0xC0AC29B7C97C50DD), uint64(0x3F84D5B5B5470917),
	uint64(0x9216D5D98979FB1B), uint64(0xD1310BA698DFB5AC),
	uint64(0x2FFD72DBD01ADFB7), uint64(0xB8E1AFED6A267E96),
	uint64(0xBA7C9045F12C7F99), uint64(0x24A19947B3916CF7),
	uint64(0x0801F2E2858EFC16), uint64(0x636920D871574E69),
}

var kSigma = [16][16]uint8{
	{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
	{14, 10, 4, 8, 9, 15, 13, 6, 1, 12, 0, 2, 11, 7, 5, 3},
	{11, 8, 12, 0, 5, 2, 15, 13, 10, 14, 3, 6, 7, 1, 9, 4},
	{7, 9, 3, 1, 13, 12, 11, 14, 2, 6, 5, 10, 4, 0, 15, 8},
	{9, 0, 5, 7, 2, 4, 10, 15, 14, 1, 11, 12, 6, 8, 3, 13},
	{2, 12, 6, 10, 0, 11, 8, 3, 4, 13, 7, 5, 15, 14, 1, 9},
	{12, 5, 1, 15, 14, 13, 4, 10, 0, 7, 6, 3, 9, 2, 8, 11},
	{13, 11, 7, 14, 12, 1, 3, 9, 5, 0, 15, 4, 8, 6, 2, 10},
	{6, 15, 14, 9, 11, 3, 0, 8, 12, 2, 13, 7, 1, 4, 10, 5},
	{10, 2, 8, 4, 7, 6, 1, 5, 15, 11, 9, 14, 3, 12, 13, 0},
	{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
	{14, 10, 4, 8, 9, 15, 13, 6, 1, 12, 0, 2, 11, 7, 5, 3},
	{11, 8, 12, 0, 5, 2, 15, 13, 10, 14, 3, 6, 7, 1, 9, 4},
	{7, 9, 3, 1, 13, 12, 11, 14, 2, 6, 5, 10, 4, 0, 15, 8},
	{9, 0, 5, 7, 2, 4, 10, 15, 14, 1, 11, 12, 6, 8, 3, 13},
	{2, 12, 6, 10, 0, 11, 8, 3, 4, 13, 7, 5, 15, 14, 1, 9},
}

// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package keccak

import (
	"fmt"

	"gitlab.com/nitya-sattva/go-x11/hash"
)

// HashSize holds the size of a hash in bytes.
const HashSize = int(64)

// BlockSize holds the size of a block in bytes.
const BlockSize = uintptr(72)

////////////////

type digest struct {
	ptr uintptr
	cnt uintptr

	h [25]uint64

	b [144]byte
}

// New returns a new digest compute a KECCAK512 hash.
func New() hash.Digest {
	ref := &digest{}
	ref.Reset()
	return ref
}

////////////////

// Reset resets the digest to its initial state.
func (ref *digest) Reset() {
	ref.ptr = 0
	ref.cnt = 200 - (512 >> 2)

	h := ref.h[:]
	h[0] = uint64(0x0)
	h[1] = uint64(0xFFFFFFFFFFFFFFFF)
	h[2] = uint64(0xFFFFFFFFFFFFFFFF)
	h[3] = uint64(0x0)
	h[4] = uint64(0x0)
	h[5] = uint64(0x0)
	h[6] = uint64(0x0)
	h[7] = uint64(0x0)
	h[8] = uint64(0xFFFFFFFFFFFFFFFF)
	h[9] = uint64(0x0)
	h[10] = uint64(0x0)
	h[11] = uint64(0x0)
	h[12] = uint64(0xFFFFFFFFFFFFFFFF)
	h[13] = uint64(0x0)
	h[14] = uint64(0x0)
	h[15] = uint64(0x0)
	h[16] = uint64(0x0)
	h[17] = uint64(0xFFFFFFFFFFFFFFFF)
	h[18] = uint64(0x0)
	h[19] = uint64(0x0)
	h[20] = uint64(0xFFFFFFFFFFFFFFFF)
	h[21] = uint64(0x0)
	h[22] = uint64(0x0)
	h[23] = uint64(0x0)
	h[24] = uint64(0x0)
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

	buf := ref.b[:]
	sta := ref.h[:]

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
			sta[0] ^= decUInt64le(buf[0:])
			sta[1] ^= decUInt64le(buf[8:])
			sta[2] ^= decUInt64le(buf[16:])
			sta[3] ^= decUInt64le(buf[24:])
			sta[4] ^= decUInt64le(buf[32:])
			sta[5] ^= decUInt64le(buf[40:])
			sta[6] ^= decUInt64le(buf[48:])
			sta[7] ^= decUInt64le(buf[56:])
			sta[8] ^= decUInt64le(buf[64:])

			for j := uintptr(0); j < 24; j++ {
				var t0, t1, t2, t3, t4, tp uint64

				{
					var tt0, tt1, tt2, tt3 uint64

					tt0 = sta[1] ^ sta[6]
					tt1 = sta[11] ^ sta[16]
					tt0 = tt0 ^ sta[21]
					tt0 = tt0 ^ tt1
					tt0 = (tt0 << 1) | (tt0 >> (64 - 1))
					tt2 = sta[4] ^ sta[9]
					tt3 = sta[14] ^ sta[19]
					tt0 = tt0 ^ sta[24]
					tt2 = tt2 ^ tt3
					t0 = tt0 ^ tt2

					tt0 = sta[2] ^ sta[7]
					tt1 = sta[12] ^ sta[17]
					tt0 = tt0 ^ sta[22]
					tt0 = tt0 ^ tt1
					tt0 = (tt0 << 1) | (tt0 >> (64 - 1))
					tt2 = sta[0] ^ sta[5]
					tt3 = sta[10] ^ sta[15]
					tt0 = tt0 ^ sta[20]
					tt2 = tt2 ^ tt3
					t1 = tt0 ^ tt2

					tt0 = sta[3] ^ sta[8]
					tt1 = sta[13] ^ sta[18]
					tt0 = tt0 ^ sta[23]
					tt0 = tt0 ^ tt1
					tt0 = (tt0 << 1) | (tt0 >> (64 - 1))
					tt2 = sta[1] ^ sta[6]
					tt3 = sta[11] ^ sta[16]
					tt0 = tt0 ^ sta[21]
					tt2 = tt2 ^ tt3
					t2 = tt0 ^ tt2

					tt0 = sta[4] ^ sta[9]
					tt1 = sta[14] ^ sta[19]
					tt0 = tt0 ^ sta[24]
					tt0 = tt0 ^ tt1
					tt0 = (tt0 << 1) | (tt0 >> (64 - 1))
					tt2 = sta[2] ^ sta[7]
					tt3 = sta[12] ^ sta[17]
					tt0 = tt0 ^ sta[22]
					tt2 = tt2 ^ tt3
					t3 = tt0 ^ tt2

					tt0 = sta[0] ^ sta[5]
					tt1 = sta[10] ^ sta[15]
					tt0 = tt0 ^ sta[20]
					tt0 = tt0 ^ tt1
					tt0 = (tt0 << 1) | (tt0 >> (64 - 1))
					tt2 = sta[3] ^ sta[8]
					tt3 = sta[13] ^ sta[18]
					tt0 = tt0 ^ sta[23]
					tt2 = tt2 ^ tt3
					t4 = tt0 ^ tt2
				}

				sta[0] = sta[0] ^ t0
				sta[1] = sta[1] ^ t1
				sta[2] = sta[2] ^ t2
				sta[3] = sta[3] ^ t3
				sta[4] = sta[4] ^ t4
				sta[5] = sta[5] ^ t0
				sta[6] = sta[6] ^ t1
				sta[7] = sta[7] ^ t2
				sta[8] = sta[8] ^ t3
				sta[9] = sta[9] ^ t4
				sta[10] = sta[10] ^ t0
				sta[11] = sta[11] ^ t1
				sta[12] = sta[12] ^ t2
				sta[13] = sta[13] ^ t3
				sta[14] = sta[14] ^ t4
				sta[15] = sta[15] ^ t0
				sta[16] = sta[16] ^ t1
				sta[17] = sta[17] ^ t2
				sta[18] = sta[18] ^ t3
				sta[23] = sta[23] ^ t3
				sta[19] = sta[19] ^ t4
				sta[20] = sta[20] ^ t0
				sta[22] = sta[22] ^ t2
				sta[21] = sta[21] ^ t1
				sta[24] = sta[24] ^ t4

				sta[1] = (sta[1] << 1) | (sta[1] >> (64 - 1))
				sta[2] = (sta[2] << 62) | (sta[2] >> (64 - 62))
				sta[3] = (sta[3] << 28) | (sta[3] >> (64 - 28))
				sta[4] = (sta[4] << 27) | (sta[4] >> (64 - 27))
				sta[5] = (sta[5] << 36) | (sta[5] >> (64 - 36))
				sta[6] = (sta[6] << 44) | (sta[6] >> (64 - 44))
				sta[7] = (sta[7] << 6) | (sta[7] >> (64 - 6))
				sta[8] = (sta[8] << 55) | (sta[8] >> (64 - 55))
				sta[9] = (sta[9] << 20) | (sta[9] >> (64 - 20))
				sta[10] = (sta[10] << 3) | (sta[10] >> (64 - 3))
				sta[11] = (sta[11] << 10) | (sta[11] >> (64 - 10))
				sta[12] = (sta[12] << 43) | (sta[12] >> (64 - 43))
				sta[13] = (sta[13] << 25) | (sta[13] >> (64 - 25))
				sta[14] = (sta[14] << 39) | (sta[14] >> (64 - 39))
				sta[15] = (sta[15] << 41) | (sta[15] >> (64 - 41))
				sta[16] = (sta[16] << 45) | (sta[16] >> (64 - 45))
				sta[17] = (sta[17] << 15) | (sta[17] >> (64 - 15))
				sta[18] = (sta[18] << 21) | (sta[18] >> (64 - 21))
				sta[19] = (sta[19] << 8) | (sta[19] >> (64 - 8))
				sta[20] = (sta[20] << 18) | (sta[20] >> (64 - 18))
				sta[21] = (sta[21] << 2) | (sta[21] >> (64 - 2))
				sta[22] = (sta[22] << 61) | (sta[22] >> (64 - 61))
				sta[23] = (sta[23] << 56) | (sta[23] >> (64 - 56))
				sta[24] = (sta[24] << 14) | (sta[24] >> (64 - 14))

				tp = ^sta[12]
				t0 = sta[6] | sta[12]
				t0 = sta[0] ^ t0
				t1 = tp | sta[18]
				t1 = sta[6] ^ t1
				t2 = sta[18] & sta[24]
				t2 = sta[12] ^ t2
				t3 = sta[24] | sta[0]
				t3 = sta[18] ^ t3
				t4 = sta[0] & sta[6]
				t4 = sta[24] ^ t4

				sta[0] = t0
				sta[6] = t1
				sta[12] = t2
				sta[18] = t3
				sta[24] = t4

				tp = ^sta[22]
				t0 = sta[9] | sta[10]
				t0 = sta[3] ^ t0
				t1 = sta[10] & sta[16]
				t1 = sta[9] ^ t1
				t2 = sta[16] | tp
				t2 = sta[10] ^ t2
				t3 = sta[22] | sta[3]
				t3 = sta[16] ^ t3
				t4 = sta[3] & sta[9]
				t4 = sta[22] ^ t4

				sta[3] = t0
				sta[9] = t1
				sta[10] = t2
				sta[16] = t3
				sta[22] = t4

				tp = ^sta[19]
				t0 = sta[7] | sta[13]
				t0 = sta[1] ^ t0
				t1 = sta[13] & sta[19]
				t1 = sta[7] ^ t1
				t2 = tp & sta[20]
				t2 = sta[13] ^ t2
				t3 = sta[20] | sta[1]
				t3 = tp ^ t3
				t4 = sta[1] & sta[7]
				t4 = sta[20] ^ t4

				sta[1] = t0
				sta[7] = t1
				sta[13] = t2
				sta[19] = t3
				sta[20] = t4

				tp = ^sta[17]
				t0 = sta[5] & sta[11]
				t0 = sta[4] ^ t0
				t1 = sta[11] | sta[17]
				t1 = sta[5] ^ t1
				t2 = tp | sta[23]
				t2 = sta[11] ^ t2
				t3 = sta[23] & sta[4]
				t3 = tp ^ t3
				t4 = sta[4] | sta[5]
				t4 = sta[23] ^ t4

				sta[4] = t0
				sta[5] = t1
				sta[11] = t2
				sta[17] = t3
				sta[23] = t4

				tp = ^sta[8]
				t0 = tp & sta[14]
				t0 = sta[2] ^ t0
				t1 = sta[14] | sta[15]
				t1 = tp ^ t1
				t2 = sta[15] & sta[21]
				t2 = sta[14] ^ t2
				t3 = sta[21] | sta[2]
				t3 = sta[15] ^ t3
				t4 = sta[2] & sta[8]
				t4 = sta[21] ^ t4

				sta[2] = t0
				sta[8] = t1
				sta[14] = t2
				sta[15] = t3
				sta[21] = t4

				sta[0] = sta[0] ^ kSpec[j+0]

				t0 = sta[5]
				sta[5] = sta[3]
				sta[3] = sta[18]
				sta[18] = sta[17]
				sta[17] = sta[11]
				sta[11] = sta[7]
				sta[7] = sta[10]
				sta[10] = sta[1]
				sta[1] = sta[6]
				sta[6] = sta[9]
				sta[9] = sta[22]
				sta[22] = sta[14]
				sta[14] = sta[20]
				sta[20] = sta[2]
				sta[2] = sta[12]
				sta[12] = sta[13]
				sta[13] = sta[19]
				sta[19] = sta[23]
				sta[23] = sta[15]
				sta[15] = sta[4]
				sta[4] = sta[24]
				sta[24] = sta[21]
				sta[21] = sta[8]
				sta[8] = sta[16]
				sta[16] = t0
			}

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
		return fmt.Errorf("Keccak Close: dst min length: %d, got %d", HashSize, ln)
	}

	var tln uintptr
	var tmp [73]uint8

	off := uint8((uint16(0x100) | uint16(bits&0xFF)) >> (8 - bcnt))

	if ref.ptr == (72 - 1) {
		if bcnt == 7 {
			tmp[0] = off
			tmp[72] = 0x80
			tln = 1 + 72
		} else {
			tmp[0] = uint8(off | 0x80)
			tln = 1
		}
	} else {
		tln = 72 - ref.ptr
		tmp[0] = off
		tmp[tln-1] = 0x80
	}
	ref.Write(tmp[:tln])

	ref.h[1] = ^ref.h[1]
	ref.h[2] = ^ref.h[2]
	ref.h[8] = ^ref.h[8]
	ref.h[12] = ^ref.h[12]
	ref.h[17] = ^ref.h[17]
	ref.h[20] = ^ref.h[20]

	for u := uintptr(0); u < 64; u += 8 {
		encUInt64le(dst[u:], ref.h[(u>>3)])
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

////////////////

var kSpec = []uint64{
	uint64(0x0000000000000001), uint64(0x0000000000008082),
	uint64(0x800000000000808A), uint64(0x8000000080008000),
	uint64(0x000000000000808B), uint64(0x0000000080000001),
	uint64(0x8000000080008081), uint64(0x8000000000008009),
	uint64(0x000000000000008A), uint64(0x0000000000000088),
	uint64(0x0000000080008009), uint64(0x000000008000000A),
	uint64(0x000000008000808B), uint64(0x800000000000008B),
	uint64(0x8000000000008089), uint64(0x8000000000008003),
	uint64(0x8000000000008002), uint64(0x8000000000000080),
	uint64(0x000000000000800A), uint64(0x800000008000000A),
	uint64(0x8000000080008081), uint64(0x8000000000008080),
	uint64(0x0000000080000001), uint64(0x8000000080008008),
}

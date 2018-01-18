// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package echo

import (
	"fmt"

	"gitlab.com/nitya-sattva/go-x11/aesr"
	"gitlab.com/nitya-sattva/go-x11/hash"
)

// HashSize holds the size of a hash in bytes.
const HashSize = uintptr(64)

// BlockSize holds the size of a block in bytes.
const BlockSize = uintptr(128)

////////////////

type digest struct {
	ptr uintptr

	h [8][2]uint64
	c [4]uint32

	b [BlockSize]byte
}

// New returns a new digest compute a ECHO512 hash.
func New() hash.Digest {
	ref := &digest{}
	ref.Reset()
	return ref
}

////////////////

// Reset resets the digest to its initial state.
func (ref *digest) Reset() {
	ref.ptr = 0

	ref.h[0][0] = uint64(8 * HashSize)
	ref.h[0][1] = 0
	ref.h[1][0] = uint64(8 * HashSize)
	ref.h[1][1] = 0
	ref.h[2][0] = uint64(8 * HashSize)
	ref.h[2][1] = 0
	ref.h[3][0] = uint64(8 * HashSize)
	ref.h[3][1] = 0
	ref.h[4][0] = uint64(8 * HashSize)
	ref.h[4][1] = 0
	ref.h[5][0] = uint64(8 * HashSize)
	ref.h[5][1] = 0
	ref.h[6][0] = uint64(8 * HashSize)
	ref.h[6][1] = 0
	ref.h[7][0] = uint64(8 * HashSize)
	ref.h[7][1] = 0

	memset32(ref.c[:], 0)
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
			ref.c[0] += 1024
			if ref.c[0] < 1024 {
				ref.c[1] += 1
				if ref.c[1] == 0 {
					ref.c[2] += 1
					if ref.c[2] == 0 {
						ref.c[3] += 1
					}
				}
			}

			compress(ref)
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
	if ln := len(dst); HashSize > uintptr(ln) {
		return fmt.Errorf("Echo Close: dst min length: %d, got %d", HashSize, ln)
	}

	ptr := ref.ptr
	buf := ref.b[:]

	eln := uint32(ptr<<3) + uint32(bcnt)

	ref.c[0] += eln
	if ref.c[0] < eln {
		ref.c[1] += 1
		if ref.c[1] == 0 {
			ref.c[2] += 1
			if ref.c[2] == 0 {
				ref.c[3] += 1
			}
		}
	}

	var tmp [64]uint8
	encUInt32le(tmp[:], ref.c[0])
	encUInt32le(tmp[4:], ref.c[1])
	encUInt32le(tmp[8:], ref.c[2])
	encUInt32le(tmp[12:], ref.c[3])

	if eln == 0 {
		memset32(ref.c[:], 0)
	}

	{
		off := uint8(0x80) >> bcnt
		buf[ptr] = uint8((bits & -off) | off)
	}

	ptr += 1
	memset8(buf[ptr:], 0)

	if ptr > (BlockSize - 18) {
		compress(ref)
		memset8(buf[:], 0)
		memset32(ref.c[:], 0)
	}

	encUInt16le(buf[(BlockSize-18):], uint16(16<<5))
	copy(buf[(BlockSize-16):], tmp[:])
	compress(ref)

	h := ref.h[:]
	for x := uintptr(0); x < 4; x++ {
		for y := uintptr(0); y < 2; y++ {
			encUInt64le(dst[(((x*2)+y)*8):], h[x][y])
		}
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

func memset8(dst []byte, src byte) {
	for i := range dst {
		dst[i] = src
	}
}

func memset32(dst []uint32, src uint32) {
	for i := range dst {
		dst[i] = src
	}
}

func compress(ref *digest) {
	var w [16][2]uint64

	k0 := ref.c[0]
	k1 := ref.c[1]
	k2 := ref.c[2]
	k3 := ref.c[3]

	for i := uintptr(0); i < 8; i++ {
		w[i][0] = ref.h[i][0]
		w[i+8][0] = decUInt64le(ref.b[(16 * i):])
		w[i][1] = ref.h[i][1]
		w[i+8][1] = decUInt64le(ref.b[((16 * i) + 8):])
	}

	var a0, a1, a2 uint64
	var b0, b1, b2 uint64

	var w0, w1, w2, w3 uint64

	var x0, x1, x2, x3 uint32
	var y0, y1, y2, y3 uint32

	for u := uintptr(0); u < 10; u++ {
		for n := uintptr(0); n < 16; n++ {
			x0 = uint32(w[n][0])
			x1 = uint32(w[n][0] >> 32)
			x2 = uint32(w[n][1])
			x3 = uint32(w[n][1] >> 32)

			y0, y1, y2, y3 = aesr.Round32ble(
				x0, x1, x2, x3,
				k0, k1, k2, k3,
			)
			x0, x1, x2, x3 = aesr.Round32ble(
				y0, y1, y2, y3,
				0, 0, 0, 0,
			)

			w[n][0] = uint64(x0) | (uint64(x1) << 32)
			w[n][1] = uint64(x2) | (uint64(x3) << 32)

			k0 += 1
			if k0 == 0 {
				k1 += 1
				if k1 == 0 {
					k2 += 1
					if k2 == 0 {
						k3 += 1
					}
				}
			}
		}

		a0 = w[1][0]
		w[1][0] = w[5][0]
		w[5][0] = w[9][0]
		w[9][0] = w[13][0]
		w[13][0] = a0

		a0 = w[1][1]
		w[1][1] = w[5][1]
		w[5][1] = w[9][1]
		w[9][1] = w[13][1]
		w[13][1] = a0

		a0 = w[2][0]
		w[2][0] = w[10][0]
		w[10][0] = a0

		a0 = w[6][0]
		w[6][0] = w[14][0]
		w[14][0] = a0

		a0 = w[2][1]
		w[2][1] = w[10][1]
		w[10][1] = a0

		a0 = w[6][1]
		w[6][1] = w[14][1]
		w[14][1] = a0

		a0 = w[15][0]
		w[15][0] = w[11][0]
		w[11][0] = w[7][0]
		w[7][0] = w[3][0]
		w[3][0] = a0

		a0 = w[15][1]
		w[15][1] = w[11][1]
		w[11][1] = w[7][1]
		w[7][1] = w[3][1]
		w[3][1] = a0

		for n := uintptr(0); n < 2; n++ {
			w0 = w[0][n]
			w1 = w[1][n]
			w2 = w[2][n]
			w3 = w[3][n]
			a0 = w0 ^ w1
			a1 = w1 ^ w2
			a2 = w2 ^ w3
			b0 = (((a0&uint64(0x8080808080808080))>>7)*27 ^
				((a0 & uint64(0x7F7F7F7F7F7F7F7F)) << 1))
			b1 = (((a1&uint64(0x8080808080808080))>>7)*27 ^
				((a1 & uint64(0x7F7F7F7F7F7F7F7F)) << 1))
			b2 = (((a2&uint64(0x8080808080808080))>>7)*27 ^
				((a2 & uint64(0x7F7F7F7F7F7F7F7F)) << 1))
			w[0][n] = b0 ^ a1 ^ w3
			w[1][n] = b1 ^ w0 ^ a2
			w[2][n] = b2 ^ a0 ^ w3
			w[3][n] = b0 ^ b1 ^ b2 ^ a0 ^ w2

			w0 = w[4][n]
			w1 = w[5][n]
			w2 = w[6][n]
			w3 = w[7][n]
			a0 = w0 ^ w1
			a1 = w1 ^ w2
			a2 = w2 ^ w3
			b0 = (((a0&uint64(0x8080808080808080))>>7)*27 ^
				((a0 & uint64(0x7F7F7F7F7F7F7F7F)) << 1))
			b1 = (((a1&uint64(0x8080808080808080))>>7)*27 ^
				((a1 & uint64(0x7F7F7F7F7F7F7F7F)) << 1))
			b2 = (((a2&uint64(0x8080808080808080))>>7)*27 ^
				((a2 & uint64(0x7F7F7F7F7F7F7F7F)) << 1))
			w[4][n] = b0 ^ a1 ^ w3
			w[5][n] = b1 ^ w0 ^ a2
			w[6][n] = b2 ^ a0 ^ w3
			w[7][n] = b0 ^ b1 ^ b2 ^ a0 ^ w2

			w0 = w[8][n]
			w1 = w[9][n]
			w2 = w[10][n]
			w3 = w[11][n]
			a0 = w0 ^ w1
			a1 = w1 ^ w2
			a2 = w2 ^ w3
			b0 = (((a0&uint64(0x8080808080808080))>>7)*27 ^
				((a0 & uint64(0x7F7F7F7F7F7F7F7F)) << 1))
			b1 = (((a1&uint64(0x8080808080808080))>>7)*27 ^
				((a1 & uint64(0x7F7F7F7F7F7F7F7F)) << 1))
			b2 = (((a2&uint64(0x8080808080808080))>>7)*27 ^
				((a2 & uint64(0x7F7F7F7F7F7F7F7F)) << 1))
			w[8][n] = b0 ^ a1 ^ w3
			w[9][n] = b1 ^ w0 ^ a2
			w[10][n] = b2 ^ a0 ^ w3
			w[11][n] = b0 ^ b1 ^ b2 ^ a0 ^ w2

			w0 = w[12][n]
			w1 = w[13][n]
			w2 = w[14][n]
			w3 = w[15][n]
			a0 = w0 ^ w1
			a1 = w1 ^ w2
			a2 = w2 ^ w3
			b0 = (((a0&uint64(0x8080808080808080))>>7)*27 ^
				((a0 & uint64(0x7F7F7F7F7F7F7F7F)) << 1))
			b1 = (((a1&uint64(0x8080808080808080))>>7)*27 ^
				((a1 & uint64(0x7F7F7F7F7F7F7F7F)) << 1))
			b2 = (((a2&uint64(0x8080808080808080))>>7)*27 ^
				((a2 & uint64(0x7F7F7F7F7F7F7F7F)) << 1))
			w[12][n] = b0 ^ a1 ^ w3
			w[13][n] = b1 ^ w0 ^ a2
			w[14][n] = b2 ^ a0 ^ w3
			w[15][n] = b0 ^ b1 ^ b2 ^ a0 ^ w2
		}
	}

	h := ref.h[:]
	for x := uintptr(0); x < 8; x++ {
		for y := uintptr(0); y < 2; y++ {
			h[x][y] ^= decUInt64le(ref.b[(((x*2)+y)*8):]) ^ w[x][y] ^ w[x+8][y]
		}
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

func encUInt16le(dst []byte, src uint16) {
	dst[0] = uint8(src)
	dst[1] = uint8(src >> 8)
}

func encUInt32le(dst []uint8, src uint32) {
	dst[0] = uint8(src)
	dst[1] = uint8(src >> 8)
	dst[2] = uint8(src >> 16)
	dst[3] = uint8(src >> 24)
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

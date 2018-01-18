// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package cubed

import (
	"fmt"

	"gitlab.com/nitya-sattva/go-x11/hash"
)

// HashSize holds the size of a hash in bytes.
const HashSize = int(64)

// BlockSize holds the size of a block in bytes.
const BlockSize = uintptr(32)

////////////////

type digest struct {
	ptr uintptr

	h [32]uint32

	b [BlockSize]byte
}

// New returns a new digest compute a CUBEHASH512 hash.
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
		copy(ref.b[ptr:], src[:sln])
		ref.ptr += sln
		return int(sln), nil
	}

	st := ref.h[:]
	buf := ref.b[:]
	for sln > 0 {
		cln := BlockSize - ptr

		if cln > sln {
			cln = sln
		}
		sln -= cln

		copy(buf[ptr:], src[:cln])
		src = src[cln:]
		ptr += cln

		if ptr == BlockSize {
			st[0x00] ^= decUInt32le(buf[0:])
			st[0x01] ^= decUInt32le(buf[4:])
			st[0x02] ^= decUInt32le(buf[8:])
			st[0x03] ^= decUInt32le(buf[12:])
			st[0x04] ^= decUInt32le(buf[16:])
			st[0x05] ^= decUInt32le(buf[20:])
			st[0x06] ^= decUInt32le(buf[24:])
			st[0x07] ^= decUInt32le(buf[28:])

			runRounds(st)
			runRounds(st)

			runRounds(st)
			runRounds(st)

			runRounds(st)
			runRounds(st)

			runRounds(st)
			runRounds(st)

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
		return fmt.Errorf("Cubed Close: dst min length: %d, got %d", HashSize, ln)
	}
	st := ref.h[:]

	{
		buf := ref.b[:]
		ptr := ref.ptr + 1

		z := uint8(0x80) >> bcnt
		buf[ref.ptr] = uint8((bits & -z) | z)

		memset(buf[ptr:], 0)
		st[0x00] ^= decUInt32le(buf[0:])
		st[0x01] ^= decUInt32le(buf[4:])
		st[0x02] ^= decUInt32le(buf[8:])
		st[0x03] ^= decUInt32le(buf[12:])
		st[0x04] ^= decUInt32le(buf[16:])
		st[0x05] ^= decUInt32le(buf[20:])
		st[0x06] ^= decUInt32le(buf[24:])
		st[0x07] ^= decUInt32le(buf[28:])
	}

	for i := uint8(0); i < 11; i++ {
		runRounds(st)
		runRounds(st)

		runRounds(st)
		runRounds(st)

		runRounds(st)
		runRounds(st)

		runRounds(st)
		runRounds(st)

		if i == 0 {
			st[0x1F] ^= uint32(1)
		}
	}

	for i := uint8(0); i < 16; i++ {
		encUInt32le(dst[(i<<2):], ref.h[i])
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

func decUInt32le(src []byte) uint32 {
	return (uint32(src[0]) |
		uint32(src[1])<<8 |
		uint32(src[2])<<16 |
		uint32(src[3])<<24)
}

func encUInt32le(dst []byte, src uint32) {
	dst[0] = uint8(src)
	dst[1] = uint8(src >> 8)
	dst[2] = uint8(src >> 16)
	dst[3] = uint8(src >> 24)
}

func runRounds(st []uint32) {
	st[0x10] = (st[0x00] + st[0x10])
	st[0x00] = (st[0x00] << 7) | (st[0x00] >> (32 - 7))
	st[0x11] = (st[0x01] + st[0x11])
	st[0x01] = (st[0x01] << 7) | (st[0x01] >> (32 - 7))
	st[0x12] = (st[0x02] + st[0x12])
	st[0x02] = (st[0x02] << 7) | (st[0x02] >> (32 - 7))
	st[0x13] = (st[0x03] + st[0x13])
	st[0x03] = (st[0x03] << 7) | (st[0x03] >> (32 - 7))
	st[0x14] = (st[0x04] + st[0x14])
	st[0x04] = (st[0x04] << 7) | (st[0x04] >> (32 - 7))
	st[0x15] = (st[0x05] + st[0x15])
	st[0x05] = (st[0x05] << 7) | (st[0x05] >> (32 - 7))
	st[0x16] = (st[0x06] + st[0x16])
	st[0x06] = (st[0x06] << 7) | (st[0x06] >> (32 - 7))
	st[0x17] = (st[0x07] + st[0x17])
	st[0x07] = (st[0x07] << 7) | (st[0x07] >> (32 - 7))
	st[0x18] = (st[0x08] + st[0x18])
	st[0x08] = (st[0x08] << 7) | (st[0x08] >> (32 - 7))
	st[0x19] = (st[0x09] + st[0x19])
	st[0x09] = (st[0x09] << 7) | (st[0x09] >> (32 - 7))
	st[0x1A] = (st[0x0A] + st[0x1A])
	st[0x0A] = (st[0x0A] << 7) | (st[0x0A] >> (32 - 7))
	st[0x1B] = (st[0x0B] + st[0x1B])
	st[0x0B] = (st[0x0B] << 7) | (st[0x0B] >> (32 - 7))
	st[0x1C] = (st[0x0C] + st[0x1C])
	st[0x0C] = (st[0x0C] << 7) | (st[0x0C] >> (32 - 7))
	st[0x1D] = (st[0x0D] + st[0x1D])
	st[0x0D] = (st[0x0D] << 7) | (st[0x0D] >> (32 - 7))
	st[0x1E] = (st[0x0E] + st[0x1E])
	st[0x0E] = (st[0x0E] << 7) | (st[0x0E] >> (32 - 7))
	st[0x1F] = (st[0x0F] + st[0x1F])
	st[0x0F] = (st[0x0F] << 7) | (st[0x0F] >> (32 - 7))
	st[0x08] ^= st[0x10]
	st[0x09] ^= st[0x11]
	st[0x0A] ^= st[0x12]
	st[0x0B] ^= st[0x13]
	st[0x0C] ^= st[0x14]
	st[0x0D] ^= st[0x15]
	st[0x0E] ^= st[0x16]
	st[0x0F] ^= st[0x17]
	st[0x00] ^= st[0x18]
	st[0x01] ^= st[0x19]
	st[0x02] ^= st[0x1A]
	st[0x03] ^= st[0x1B]
	st[0x04] ^= st[0x1C]
	st[0x05] ^= st[0x1D]
	st[0x06] ^= st[0x1E]
	st[0x07] ^= st[0x1F]
	st[0x12] = (st[0x08] + st[0x12])
	st[0x08] = (st[0x08] << 11) | (st[0x08] >> (32 - 11))
	st[0x13] = (st[0x09] + st[0x13])
	st[0x09] = (st[0x09] << 11) | (st[0x09] >> (32 - 11))
	st[0x10] = (st[0x0A] + st[0x10])
	st[0x0A] = (st[0x0A] << 11) | (st[0x0A] >> (32 - 11))
	st[0x11] = (st[0x0B] + st[0x11])
	st[0x0B] = (st[0x0B] << 11) | (st[0x0B] >> (32 - 11))
	st[0x16] = (st[0x0C] + st[0x16])
	st[0x0C] = (st[0x0C] << 11) | (st[0x0C] >> (32 - 11))
	st[0x17] = (st[0x0D] + st[0x17])
	st[0x0D] = (st[0x0D] << 11) | (st[0x0D] >> (32 - 11))
	st[0x14] = (st[0x0E] + st[0x14])
	st[0x0E] = (st[0x0E] << 11) | (st[0x0E] >> (32 - 11))
	st[0x15] = (st[0x0F] + st[0x15])
	st[0x0F] = (st[0x0F] << 11) | (st[0x0F] >> (32 - 11))
	st[0x1A] = (st[0x00] + st[0x1A])
	st[0x00] = (st[0x00] << 11) | (st[0x00] >> (32 - 11))
	st[0x1B] = (st[0x01] + st[0x1B])
	st[0x01] = (st[0x01] << 11) | (st[0x01] >> (32 - 11))
	st[0x18] = (st[0x02] + st[0x18])
	st[0x02] = (st[0x02] << 11) | (st[0x02] >> (32 - 11))
	st[0x19] = (st[0x03] + st[0x19])
	st[0x03] = (st[0x03] << 11) | (st[0x03] >> (32 - 11))
	st[0x1E] = (st[0x04] + st[0x1E])
	st[0x04] = (st[0x04] << 11) | (st[0x04] >> (32 - 11))
	st[0x1F] = (st[0x05] + st[0x1F])
	st[0x05] = (st[0x05] << 11) | (st[0x05] >> (32 - 11))
	st[0x1C] = (st[0x06] + st[0x1C])
	st[0x06] = (st[0x06] << 11) | (st[0x06] >> (32 - 11))
	st[0x1D] = (st[0x07] + st[0x1D])
	st[0x07] = (st[0x07] << 11) | (st[0x07] >> (32 - 11))
	st[0x0C] ^= st[0x12]
	st[0x0D] ^= st[0x13]
	st[0x0E] ^= st[0x10]
	st[0x0F] ^= st[0x11]
	st[0x08] ^= st[0x16]
	st[0x09] ^= st[0x17]
	st[0x0A] ^= st[0x14]
	st[0x0B] ^= st[0x15]
	st[0x04] ^= st[0x1A]
	st[0x05] ^= st[0x1B]
	st[0x06] ^= st[0x18]
	st[0x07] ^= st[0x19]
	st[0x00] ^= st[0x1E]
	st[0x01] ^= st[0x1F]
	st[0x02] ^= st[0x1C]
	st[0x03] ^= st[0x1D]

	st[0x13] = (st[0x0C] + st[0x13])
	st[0x0C] = (st[0x0C] << 7) | (st[0x0C] >> (32 - 7))
	st[0x12] = (st[0x0D] + st[0x12])
	st[0x0D] = (st[0x0D] << 7) | (st[0x0D] >> (32 - 7))
	st[0x11] = (st[0x0E] + st[0x11])
	st[0x0E] = (st[0x0E] << 7) | (st[0x0E] >> (32 - 7))
	st[0x10] = (st[0x0F] + st[0x10])
	st[0x0F] = (st[0x0F] << 7) | (st[0x0F] >> (32 - 7))
	st[0x17] = (st[0x08] + st[0x17])
	st[0x08] = (st[0x08] << 7) | (st[0x08] >> (32 - 7))
	st[0x16] = (st[0x09] + st[0x16])
	st[0x09] = (st[0x09] << 7) | (st[0x09] >> (32 - 7))
	st[0x15] = (st[0x0A] + st[0x15])
	st[0x0A] = (st[0x0A] << 7) | (st[0x0A] >> (32 - 7))
	st[0x14] = (st[0x0B] + st[0x14])
	st[0x0B] = (st[0x0B] << 7) | (st[0x0B] >> (32 - 7))
	st[0x1B] = (st[0x04] + st[0x1B])
	st[0x04] = (st[0x04] << 7) | (st[0x04] >> (32 - 7))
	st[0x1A] = (st[0x05] + st[0x1A])
	st[0x05] = (st[0x05] << 7) | (st[0x05] >> (32 - 7))
	st[0x19] = (st[0x06] + st[0x19])
	st[0x06] = (st[0x06] << 7) | (st[0x06] >> (32 - 7))
	st[0x18] = (st[0x07] + st[0x18])
	st[0x07] = (st[0x07] << 7) | (st[0x07] >> (32 - 7))
	st[0x1F] = (st[0x00] + st[0x1F])
	st[0x00] = (st[0x00] << 7) | (st[0x00] >> (32 - 7))
	st[0x1E] = (st[0x01] + st[0x1E])
	st[0x01] = (st[0x01] << 7) | (st[0x01] >> (32 - 7))
	st[0x1D] = (st[0x02] + st[0x1D])
	st[0x02] = (st[0x02] << 7) | (st[0x02] >> (32 - 7))
	st[0x1C] = (st[0x03] + st[0x1C])
	st[0x03] = (st[0x03] << 7) | (st[0x03] >> (32 - 7))
	st[0x04] ^= st[0x13]
	st[0x05] ^= st[0x12]
	st[0x06] ^= st[0x11]
	st[0x07] ^= st[0x10]
	st[0x00] ^= st[0x17]
	st[0x01] ^= st[0x16]
	st[0x02] ^= st[0x15]
	st[0x03] ^= st[0x14]
	st[0x0C] ^= st[0x1B]
	st[0x0D] ^= st[0x1A]
	st[0x0E] ^= st[0x19]
	st[0x0F] ^= st[0x18]
	st[0x08] ^= st[0x1F]
	st[0x09] ^= st[0x1E]
	st[0x0A] ^= st[0x1D]
	st[0x0B] ^= st[0x1C]
	st[0x11] = (st[0x04] + st[0x11])
	st[0x04] = (st[0x04] << 11) | (st[0x04] >> (32 - 11))
	st[0x10] = (st[0x05] + st[0x10])
	st[0x05] = (st[0x05] << 11) | (st[0x05] >> (32 - 11))
	st[0x13] = (st[0x06] + st[0x13])
	st[0x06] = (st[0x06] << 11) | (st[0x06] >> (32 - 11))
	st[0x12] = (st[0x07] + st[0x12])
	st[0x07] = (st[0x07] << 11) | (st[0x07] >> (32 - 11))
	st[0x15] = (st[0x00] + st[0x15])
	st[0x00] = (st[0x00] << 11) | (st[0x00] >> (32 - 11))
	st[0x14] = (st[0x01] + st[0x14])
	st[0x01] = (st[0x01] << 11) | (st[0x01] >> (32 - 11))
	st[0x17] = (st[0x02] + st[0x17])
	st[0x02] = (st[0x02] << 11) | (st[0x02] >> (32 - 11))
	st[0x16] = (st[0x03] + st[0x16])
	st[0x03] = (st[0x03] << 11) | (st[0x03] >> (32 - 11))
	st[0x19] = (st[0x0C] + st[0x19])
	st[0x0C] = (st[0x0C] << 11) | (st[0x0C] >> (32 - 11))
	st[0x18] = (st[0x0D] + st[0x18])
	st[0x0D] = (st[0x0D] << 11) | (st[0x0D] >> (32 - 11))
	st[0x1B] = (st[0x0E] + st[0x1B])
	st[0x0E] = (st[0x0E] << 11) | (st[0x0E] >> (32 - 11))
	st[0x1A] = (st[0x0F] + st[0x1A])
	st[0x0F] = (st[0x0F] << 11) | (st[0x0F] >> (32 - 11))
	st[0x1D] = (st[0x08] + st[0x1D])
	st[0x08] = (st[0x08] << 11) | (st[0x08] >> (32 - 11))
	st[0x1C] = (st[0x09] + st[0x1C])
	st[0x09] = (st[0x09] << 11) | (st[0x09] >> (32 - 11))
	st[0x1F] = (st[0x0A] + st[0x1F])
	st[0x0A] = (st[0x0A] << 11) | (st[0x0A] >> (32 - 11))
	st[0x1E] = (st[0x0B] + st[0x1E])
	st[0x0B] = (st[0x0B] << 11) | (st[0x0B] >> (32 - 11))
	st[0x00] ^= st[0x11]
	st[0x01] ^= st[0x10]
	st[0x02] ^= st[0x13]
	st[0x03] ^= st[0x12]
	st[0x04] ^= st[0x15]
	st[0x05] ^= st[0x14]
	st[0x06] ^= st[0x17]
	st[0x07] ^= st[0x16]
	st[0x08] ^= st[0x19]
	st[0x09] ^= st[0x18]
	st[0x0A] ^= st[0x1B]
	st[0x0B] ^= st[0x1A]
	st[0x0C] ^= st[0x1D]
	st[0x0D] ^= st[0x1C]
	st[0x0E] ^= st[0x1F]
	st[0x0F] ^= st[0x1E]
}

////////////////

var kInit = [32]uint32{
	uint32(0x2AEA2A61), uint32(0x50F494D4), uint32(0x2D538B8B),
	uint32(0x4167D83E), uint32(0x3FEE2313), uint32(0xC701CF8C),
	uint32(0xCC39968E), uint32(0x50AC5695), uint32(0x4D42C787),
	uint32(0xA647A8B3), uint32(0x97CF0BEF), uint32(0x825B4537),
	uint32(0xEEF864D2), uint32(0xF22090C4), uint32(0xD0E5CD33),
	uint32(0xA23911AE), uint32(0xFCD398D9), uint32(0x148FE485),
	uint32(0x1B017BEF), uint32(0xB6444532), uint32(0x6A536159),
	uint32(0x2FF5781C), uint32(0x91FA7934), uint32(0x0DBADEA9),
	uint32(0xD65C8A2B), uint32(0xA5A70E75), uint32(0xB1C62456),
	uint32(0xBC796576), uint32(0x1921C8F7), uint32(0xE7989AF1),
	uint32(0x7795D246), uint32(0xD43E3B44),
}

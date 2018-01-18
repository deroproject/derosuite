// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package jhash

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
	cnt uintptr

	h [16]uint64

	b [BlockSize]byte
}

// New returns a new digest compute a JH512 hash.
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
	buf := ref.b[:]
	ptr := ref.ptr

	if sln < (BlockSize - ptr) {
		copy(buf[ptr:], src)
		ref.ptr += sln
		return int(sln), nil
	}

	var hi, lo [8]uint64
	hi[0] = ref.h[0x0]
	lo[0] = ref.h[0x1]
	hi[1] = ref.h[0x2]
	lo[1] = ref.h[0x3]
	hi[2] = ref.h[0x4]
	lo[2] = ref.h[0x5]
	hi[3] = ref.h[0x6]
	lo[3] = ref.h[0x7]
	hi[4] = ref.h[0x8]
	lo[4] = ref.h[0x9]
	hi[5] = ref.h[0xA]
	lo[5] = ref.h[0xB]
	hi[6] = ref.h[0xC]
	lo[6] = ref.h[0xD]
	hi[7] = ref.h[0xE]
	lo[7] = ref.h[0xF]

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
			m0h := decUInt64le(buf[0:])
			m0l := decUInt64le(buf[8:])
			m1h := decUInt64le(buf[16:])
			m1l := decUInt64le(buf[24:])
			m2h := decUInt64le(buf[32:])
			m2l := decUInt64le(buf[40:])
			m3h := decUInt64le(buf[48:])
			m3l := decUInt64le(buf[56:])

			hi[0] ^= m0h
			lo[0] ^= m0l
			hi[1] ^= m1h
			lo[1] ^= m1l
			hi[2] ^= m2h
			lo[2] ^= m2l
			hi[3] ^= m3h
			lo[3] ^= m3l

			for r := uint64(0); r < 42; r += 7 {
				slMutateExtend(r+0, 0, hi[:], lo[:])
				slMutateExtend(r+1, 1, hi[:], lo[:])
				slMutateExtend(r+2, 2, hi[:], lo[:])
				slMutateExtend(r+3, 3, hi[:], lo[:])
				slMutateExtend(r+4, 4, hi[:], lo[:])
				slMutateExtend(r+5, 5, hi[:], lo[:])
				slMutateBasic(r+6, hi[:], lo[:])
			}

			hi[4] ^= m0h
			lo[4] ^= m0l
			hi[5] ^= m1h
			lo[5] ^= m1l
			hi[6] ^= m2h
			lo[6] ^= m2l
			hi[7] ^= m3h
			lo[7] ^= m3l

			ref.cnt++
			ptr = 0
		}
	}

	ref.h[0x0] = hi[0]
	ref.h[0x1] = lo[0]
	ref.h[0x2] = hi[1]
	ref.h[0x3] = lo[1]
	ref.h[0x4] = hi[2]
	ref.h[0x5] = lo[2]
	ref.h[0x6] = hi[3]
	ref.h[0x7] = lo[3]
	ref.h[0x8] = hi[4]
	ref.h[0x9] = lo[4]
	ref.h[0xA] = hi[5]
	ref.h[0xB] = lo[5]
	ref.h[0xC] = hi[6]
	ref.h[0xD] = lo[6]
	ref.h[0xE] = hi[7]
	ref.h[0xF] = lo[7]

	ref.ptr = ptr
	return fln, nil
}

// Close the digest by writing the last bits and storing the hash
// in dst. This prepares the digest for reuse by calling reset. A call
// to Close with a dst that is smaller then HashSize will return an error.
func (ref *digest) Close(dst []byte, bits uint8, bcnt uint8) error {
	if ln := len(dst); HashSize > ln {
		return fmt.Errorf("JHash Close: dst min length: %d, got %d", HashSize, ln)
	}

	var ocnt uintptr
	var buf [128]uint8

	{
		off := uint8(0x80) >> bcnt
		buf[0] = uint8((bits & -off) | off)
	}

	if ref.ptr == 0 && bcnt == 0 {
		ocnt = 47
	} else {
		ocnt = 111 - ref.ptr
	}

	l0 := uint64(bcnt)
	l0 += uint64(ref.cnt << 9)
	l0 += uint64(ref.ptr << 3)
	l1 := uint64(ref.cnt >> 55)

	encUInt64be(buf[ocnt+1:], l1)
	encUInt64be(buf[ocnt+9:], l0)

	ref.Write(buf[:ocnt+17])

	for u := uintptr(0); u < 8; u++ {
		encUInt64le(dst[(u<<3):], ref.h[u+8])
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

func slMutateBasic(r uint64, hi, lo []uint64) {
	var tmp uint64

	tmp = kSpec[(r<<2)+0]
	hi[6] = ^hi[6]
	hi[0] ^= tmp & ^hi[4]
	tmp = tmp ^ (hi[0] & hi[2])
	hi[0] ^= hi[4] & hi[6]
	hi[6] ^= ^hi[2] & hi[4]
	hi[2] ^= hi[0] & hi[4]
	hi[4] ^= hi[0] & ^hi[6]
	hi[0] ^= hi[2] | hi[6]
	hi[6] ^= hi[2] & hi[4]
	hi[2] ^= tmp & hi[0]
	hi[4] ^= tmp

	tmp = kSpec[(r<<2)+1]
	lo[6] = ^lo[6]
	lo[0] ^= tmp & ^lo[4]
	tmp = tmp ^ (lo[0] & lo[2])
	lo[0] ^= lo[4] & lo[6]
	lo[6] ^= ^lo[2] & lo[4]
	lo[2] ^= lo[0] & lo[4]
	lo[4] ^= lo[0] & ^lo[6]
	lo[0] ^= lo[2] | lo[6]
	lo[6] ^= lo[2] & lo[4]
	lo[2] ^= tmp & lo[0]
	lo[4] ^= tmp

	tmp = kSpec[(r<<2)+2]
	hi[7] = ^hi[7]
	hi[1] ^= tmp & ^hi[5]
	tmp = tmp ^ (hi[1] & hi[3])
	hi[1] ^= hi[5] & hi[7]
	hi[7] ^= ^hi[3] & hi[5]
	hi[3] ^= hi[1] & hi[5]
	hi[5] ^= hi[1] & ^hi[7]
	hi[1] ^= hi[3] | hi[7]
	hi[7] ^= hi[3] & hi[5]
	hi[3] ^= tmp & hi[1]
	hi[5] ^= tmp

	tmp = kSpec[(r<<2)+3]
	lo[7] = ^lo[7]
	lo[1] ^= tmp & ^lo[5]
	tmp = tmp ^ (lo[1] & lo[3])
	lo[1] ^= lo[5] & lo[7]
	lo[7] ^= ^lo[3] & lo[5]
	lo[3] ^= lo[1] & lo[5]
	lo[5] ^= lo[1] & ^lo[7]
	lo[1] ^= lo[3] | lo[7]
	lo[7] ^= lo[3] & lo[5]
	lo[3] ^= tmp & lo[1]
	lo[5] ^= tmp

	hi[1] ^= hi[2]
	hi[3] ^= hi[4]
	hi[5] ^= hi[6] ^ hi[0]
	hi[7] ^= hi[0]
	hi[0] ^= hi[3]
	hi[2] ^= hi[5]
	hi[4] ^= hi[7] ^ hi[1]
	hi[6] ^= hi[1]

	lo[1] ^= lo[2]
	lo[3] ^= lo[4]
	lo[5] ^= lo[6] ^ lo[0]
	lo[7] ^= lo[0]
	lo[0] ^= lo[3]
	lo[2] ^= lo[5]
	lo[4] ^= lo[7] ^ lo[1]
	lo[6] ^= lo[1]

	tmp = hi[1]
	hi[1] = lo[1]
	lo[1] = tmp

	tmp = hi[3]
	hi[3] = lo[3]
	lo[3] = tmp

	tmp = hi[5]
	hi[5] = lo[5]
	lo[5] = tmp

	tmp = hi[7]
	hi[7] = lo[7]
	lo[7] = tmp
}

func slMutateExtend(r, ro uint64, hi, lo []uint64) {
	var tmp uint64

	tmp = kSpec[(r<<2)+0]
	hi[6] = ^hi[6]
	hi[0] ^= tmp & ^hi[4]
	tmp = tmp ^ (hi[0] & hi[2])
	hi[0] ^= hi[4] & hi[6]
	hi[6] ^= ^hi[2] & hi[4]
	hi[2] ^= hi[0] & hi[4]
	hi[4] ^= hi[0] & ^hi[6]
	hi[0] ^= hi[2] | hi[6]
	hi[6] ^= hi[2] & hi[4]
	hi[2] ^= tmp & hi[0]
	hi[4] ^= tmp

	tmp = kSpec[(r<<2)+1]
	lo[6] = ^lo[6]
	lo[0] ^= tmp & ^lo[4]
	tmp = tmp ^ (lo[0] & lo[2])
	lo[0] ^= lo[4] & lo[6]
	lo[6] ^= ^lo[2] & lo[4]
	lo[2] ^= lo[0] & lo[4]
	lo[4] ^= lo[0] & ^lo[6]
	lo[0] ^= lo[2] | lo[6]
	lo[6] ^= lo[2] & lo[4]
	lo[2] ^= tmp & lo[0]
	lo[4] ^= tmp

	tmp = kSpec[(r<<2)+2]
	hi[7] = ^hi[7]
	hi[1] ^= tmp & ^hi[5]
	tmp = tmp ^ (hi[1] & hi[3])
	hi[1] ^= hi[5] & hi[7]
	hi[7] ^= ^hi[3] & hi[5]
	hi[3] ^= hi[1] & hi[5]
	hi[5] ^= hi[1] & ^hi[7]
	hi[1] ^= hi[3] | hi[7]
	hi[7] ^= hi[3] & hi[5]
	hi[3] ^= tmp & hi[1]
	hi[5] ^= tmp

	tmp = kSpec[(r<<2)+3]
	lo[7] = ^lo[7]
	lo[1] ^= tmp & ^lo[5]
	tmp = tmp ^ (lo[1] & lo[3])
	lo[1] ^= lo[5] & lo[7]
	lo[7] ^= ^lo[3] & lo[5]
	lo[3] ^= lo[1] & lo[5]
	lo[5] ^= lo[1] & ^lo[7]
	lo[1] ^= lo[3] | lo[7]
	lo[7] ^= lo[3] & lo[5]
	lo[3] ^= tmp & lo[1]
	lo[5] ^= tmp

	hi[1] ^= hi[2]
	hi[3] ^= hi[4]
	hi[5] ^= hi[6] ^ hi[0]
	hi[7] ^= hi[0]
	hi[0] ^= hi[3]
	hi[2] ^= hi[5]
	hi[4] ^= hi[7] ^ hi[1]
	hi[6] ^= hi[1]

	lo[1] ^= lo[2]
	lo[3] ^= lo[4]
	lo[5] ^= lo[6] ^ lo[0]
	lo[7] ^= lo[0]
	lo[0] ^= lo[3]
	lo[2] ^= lo[5]
	lo[4] ^= lo[7] ^ lo[1]
	lo[6] ^= lo[1]

	tmp = (hi[1] & (kWrapValue[ro])) << (kWrapOffset[ro])
	hi[1] = ((hi[1] >> (kWrapOffset[ro])) & (kWrapValue[ro])) | tmp
	tmp = (lo[1] & (kWrapValue[ro])) << (kWrapOffset[ro])
	lo[1] = ((lo[1] >> (kWrapOffset[ro])) & (kWrapValue[ro])) | tmp

	tmp = (hi[3] & (kWrapValue[ro])) << (kWrapOffset[ro])
	hi[3] = ((hi[3] >> (kWrapOffset[ro])) & (kWrapValue[ro])) | tmp
	tmp = (lo[3] & (kWrapValue[ro])) << (kWrapOffset[ro])
	lo[3] = ((lo[3] >> (kWrapOffset[ro])) & (kWrapValue[ro])) | tmp

	tmp = (hi[5] & (kWrapValue[ro])) << (kWrapOffset[ro])
	hi[5] = ((hi[5] >> (kWrapOffset[ro])) & (kWrapValue[ro])) | tmp
	tmp = (lo[5] & (kWrapValue[ro])) << (kWrapOffset[ro])
	lo[5] = ((lo[5] >> (kWrapOffset[ro])) & (kWrapValue[ro])) | tmp

	tmp = (hi[7] & (kWrapValue[ro])) << (kWrapOffset[ro])
	hi[7] = ((hi[7] >> (kWrapOffset[ro])) & (kWrapValue[ro])) | tmp
	tmp = (lo[7] & (kWrapValue[ro])) << (kWrapOffset[ro])
	lo[7] = ((lo[7] >> (kWrapOffset[ro])) & (kWrapValue[ro])) | tmp
}

////////////////

var kInit = []uint64{
	uint64(0x17aa003e964bd16f), uint64(0x43d5157a052e6a63),
	uint64(0x0bef970c8d5e228a), uint64(0x61c3b3f2591234e9),
	uint64(0x1e806f53c1a01d89), uint64(0x806d2bea6b05a92a),
	uint64(0xa6ba7520dbcc8e58), uint64(0xf73bf8ba763a0fa9),
	uint64(0x694ae34105e66901), uint64(0x5ae66f2e8e8ab546),
	uint64(0x243c84c1d0a74710), uint64(0x99c15a2db1716e3b),
	uint64(0x56f8b19decf657cf), uint64(0x56b116577c8806a7),
	uint64(0xfb1785e6dffcc2e3), uint64(0x4bdd8ccc78465a54),
}

var kSpec = []uint64{
	uint64(0x67f815dfa2ded572), uint64(0x571523b70a15847b),
	uint64(0xf6875a4d90d6ab81), uint64(0x402bd1c3c54f9f4e),
	uint64(0x9cfa455ce03a98ea), uint64(0x9a99b26699d2c503),
	uint64(0x8a53bbf2b4960266), uint64(0x31a2db881a1456b5),
	uint64(0xdb0e199a5c5aa303), uint64(0x1044c1870ab23f40),
	uint64(0x1d959e848019051c), uint64(0xdccde75eadeb336f),
	uint64(0x416bbf029213ba10), uint64(0xd027bbf7156578dc),
	uint64(0x5078aa3739812c0a), uint64(0xd3910041d2bf1a3f),
	uint64(0x907eccf60d5a2d42), uint64(0xce97c0929c9f62dd),
	uint64(0xac442bc70ba75c18), uint64(0x23fcc663d665dfd1),
	uint64(0x1ab8e09e036c6e97), uint64(0xa8ec6c447e450521),
	uint64(0xfa618e5dbb03f1ee), uint64(0x97818394b29796fd),
	uint64(0x2f3003db37858e4a), uint64(0x956a9ffb2d8d672a),
	uint64(0x6c69b8f88173fe8a), uint64(0x14427fc04672c78a),
	uint64(0xc45ec7bd8f15f4c5), uint64(0x80bb118fa76f4475),
	uint64(0xbc88e4aeb775de52), uint64(0xf4a3a6981e00b882),
	uint64(0x1563a3a9338ff48e), uint64(0x89f9b7d524565faa),
	uint64(0xfde05a7c20edf1b6), uint64(0x362c42065ae9ca36),
	uint64(0x3d98fe4e433529ce), uint64(0xa74b9a7374f93a53),
	uint64(0x86814e6f591ff5d0), uint64(0x9f5ad8af81ad9d0e),
	uint64(0x6a6234ee670605a7), uint64(0x2717b96ebe280b8b),
	uint64(0x3f1080c626077447), uint64(0x7b487ec66f7ea0e0),
	uint64(0xc0a4f84aa50a550d), uint64(0x9ef18e979fe7e391),
	uint64(0xd48d605081727686), uint64(0x62b0e5f3415a9e7e),
	uint64(0x7a205440ec1f9ffc), uint64(0x84c9f4ce001ae4e3),
	uint64(0xd895fa9df594d74f), uint64(0xa554c324117e2e55),
	uint64(0x286efebd2872df5b), uint64(0xb2c4a50fe27ff578),
	uint64(0x2ed349eeef7c8905), uint64(0x7f5928eb85937e44),
	uint64(0x4a3124b337695f70), uint64(0x65e4d61df128865e),
	uint64(0xe720b95104771bc7), uint64(0x8a87d423e843fe74),
	uint64(0xf2947692a3e8297d), uint64(0xc1d9309b097acbdd),
	uint64(0xe01bdc5bfb301b1d), uint64(0xbf829cf24f4924da),
	uint64(0xffbf70b431bae7a4), uint64(0x48bcf8de0544320d),
	uint64(0x39d3bb5332fcae3b), uint64(0xa08b29e0c1c39f45),
	uint64(0x0f09aef7fd05c9e5), uint64(0x34f1904212347094),
	uint64(0x95ed44e301b771a2), uint64(0x4a982f4f368e3be9),
	uint64(0x15f66ca0631d4088), uint64(0xffaf52874b44c147),
	uint64(0x30c60ae2f14abb7e), uint64(0xe68c6eccc5b67046),
	uint64(0x00ca4fbd56a4d5a4), uint64(0xae183ec84b849dda),
	uint64(0xadd1643045ce5773), uint64(0x67255c1468cea6e8),
	uint64(0x16e10ecbf28cdaa3), uint64(0x9a99949a5806e933),
	uint64(0x7b846fc220b2601f), uint64(0x1885d1a07facced1),
	uint64(0xd319dd8da15b5932), uint64(0x46b4a5aac01c9a50),
	uint64(0xba6b04e467633d9f), uint64(0x7eee560bab19caf6),
	uint64(0x742128a9ea79b11f), uint64(0xee51363b35f7bde9),
	uint64(0x76d350755aac571d), uint64(0x01707da3fec2463a),
	uint64(0x42d8a498afc135f7), uint64(0x79676b9e20eced78),
	uint64(0xa8db3aea15638341), uint64(0x832c83324d3bc3fa),
	uint64(0xf347271c1f3b40a7), uint64(0x9a762db734f04059),
	uint64(0xfd4f21d26c4e3ee7), uint64(0xef5957dc398dfdb8),
	uint64(0xdaeb492b490c9b8d), uint64(0x0d70f36849d7a25b),
	uint64(0x84558d7ad0ae3b7d), uint64(0x658ef8e4f0e9a5f5),
	uint64(0x533b1036f4a2b8a0), uint64(0x5aec3e759e07a80c),
	uint64(0x4f88e85692946891), uint64(0x4cbcbaf8555cb05b),
	uint64(0x7b9487f3993bbbe3), uint64(0x5d1c6b72d6f4da75),
	uint64(0x6db334dc28acae64), uint64(0x71db28b850a5346c),
	uint64(0x2a518d10f2e261f8), uint64(0xfc75dd593364dbe3),
	uint64(0xa23fce43f1bcac1c), uint64(0xb043e8023cd1bb67),
	uint64(0x75a12988ca5b0a33), uint64(0x5c5316b44d19347f),
	uint64(0x1e4d790ec3943b92), uint64(0x3fafeeb6d7757479),
	uint64(0x21391abef7d4a8ea), uint64(0x5127234c097ef45c),
	uint64(0xd23c32ba5324a326), uint64(0xadd5a66d4a17a344),
	uint64(0x08c9f2afa63e1db5), uint64(0x563c6b91983d5983),
	uint64(0x4d608672a17cf84c), uint64(0xf6c76e08cc3ee246),
	uint64(0x5e76bcb1b333982f), uint64(0x2ae6c4efa566d62b),
	uint64(0x36d4c1bee8b6f406), uint64(0x6321efbc1582ee74),
	uint64(0x69c953f40d4ec1fd), uint64(0x26585806c45a7da7),
	uint64(0x16fae0061614c17e), uint64(0x3f9d63283daf907e),
	uint64(0x0cd29b00e3f2c9d2), uint64(0x300cd4b730ceaa5f),
	uint64(0x9832e0f216512a74), uint64(0x9af8cee3d830eb0d),
	uint64(0x9279f1b57b9ec54b), uint64(0xd36886046ee651ff),
	uint64(0x316796e6574d239b), uint64(0x05750a17f3a6e6cc),
	uint64(0xce6c3213d98176b1), uint64(0x62a205f88452173c),
	uint64(0x47154778b3cb2bf4), uint64(0x486a9323825446ff),
	uint64(0x65655e4e0758df38), uint64(0x8e5086fc897cfcf2),
	uint64(0x86ca0bd0442e7031), uint64(0x4e477830a20940f0),
	uint64(0x8338f7d139eea065), uint64(0xbd3a2ce437e95ef7),
	uint64(0x6ff8130126b29721), uint64(0xe7de9fefd1ed44a3),
	uint64(0xd992257615dfa08b), uint64(0xbe42dc12f6f7853c),
	uint64(0x7eb027ab7ceca7d8), uint64(0xdea83eaada7d8d53),
	uint64(0xd86902bd93ce25aa), uint64(0xf908731afd43f65a),
	uint64(0xa5194a17daef5fc0), uint64(0x6a21fd4c33664d97),
	uint64(0x701541db3198b435), uint64(0x9b54cdedbb0f1eea),
	uint64(0x72409751a163d09a), uint64(0xe26f4791bf9d75f6),
}

var kWrapValue = []uint64{
	uint64(0x5555555555555555),
	uint64(0x3333333333333333),
	uint64(0x0F0F0F0F0F0F0F0F),
	uint64(0x00FF00FF00FF00FF),
	uint64(0x0000FFFF0000FFFF),
	uint64(0x00000000FFFFFFFF),
}

var kWrapOffset = []uint64{
	1, 2, 4, 8, 16, 32,
}

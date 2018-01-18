// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package groest

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

// New returns a new digest compute a GROESTL512 hash.
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

	for u := uintptr(0); u < 15; u++ {
		ref.h[u] = 0
	}
	ref.h[15] = ((uint64(512&0xFF) << 56) | (uint64(512&0xFF00) << 40))
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

	h := ref.h[:]
	b := ref.b[:]

	for sln > 0 {
		cln := BlockSize - ptr

		if cln > sln {
			cln = sln
		}
		sln -= cln

		copy(b[ptr:], src[:cln])
		src = src[cln:]
		ptr += cln

		if ptr == BlockSize {
			var g, m [16]uint64

			for u := uint64(0); u < 16; u++ {
				m[u] = decUInt64le(b[(u << 3):])
				g[u] = m[u] ^ h[u]
			}

			gs := g[:]
			for r := uint64(0); r < 14; r += 2 {
				gRounds(r+0, gs)
				gRounds(r+1, gs)
			}

			ms := m[:]
			for r := uint64(0); r < 14; r += 2 {
				mRounds(r+0, ms)
				mRounds(r+1, ms)
			}

			for u := uint64(0); u < 16; u++ {
				h[u] ^= g[u] ^ m[u]
			}

			ref.cnt++
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
		return fmt.Errorf("Groest Close: dst min length: %d, got %d", HashSize, ln)
	}

	ptr := ref.ptr
	cnt := uint64(0)

	pln := uintptr(0)
	pad := [136]uint8{}

	{
		off := uint8(0x80) >> bcnt
		pad[0] = uint8((bits & -off) | off)
	}

	if ptr < 120 {
		pln = 128 - ptr
		cnt = ref.cnt + 1
	} else {
		pln = 256 - ptr
		cnt = ref.cnt + 2
	}

	encUInt64be(pad[(pln-8):], cnt)
	ref.Write(pad[:pln])

	h := ref.h[:]
	g := [16]uint64{}

	gs := g[:]
	copy(gs, h[:])

	for r := uint64(0); r < 14; r += 2 {
		gRounds(r+0, gs)
		gRounds(r+1, gs)
	}

	for u := uintptr(0); u < 16; u++ {
		h[u] ^= g[u]
	}

	for u := uintptr(0); u < 8; u++ {
		encUInt64le(pad[(u<<3):], h[u+8])
	}

	copy(dst[:], pad[:])

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

func gRounds(r uint64, g []uint64) {
	g[0x0] ^= (r + uint64(0x00))
	g[0x1] ^= (r + uint64(0x10))
	g[0x2] ^= (r + uint64(0x20))
	g[0x3] ^= (r + uint64(0x30))
	g[0x4] ^= (r + uint64(0x40))
	g[0x5] ^= (r + uint64(0x50))
	g[0x6] ^= (r + uint64(0x60))
	g[0x7] ^= (r + uint64(0x70))
	g[0x8] ^= (r + uint64(0x80))
	g[0x9] ^= (r + uint64(0x90))
	g[0xA] ^= (r + uint64(0xA0))
	g[0xB] ^= (r + uint64(0xB0))
	g[0xC] ^= (r + uint64(0xC0))
	g[0xD] ^= (r + uint64(0xD0))
	g[0xE] ^= (r + uint64(0xE0))
	g[0xF] ^= (r + uint64(0xF0))

	var t [16]uint64
	var tp, ix uint64

	for u := uint64(0); u < 16; u += 4 {
		t[u] = kTab0[((g[(u+0)&0xF]) & 0xFF)]
		tp = kTab0[(((g[(u+1)&0xF]) >> 8) & 0xFF)]
		t[u] ^= (tp << 8) | (tp >> (64 - 8))
		tp = kTab0[(((g[(u+2)&0xF]) >> 16) & 0xFF)]
		t[u] ^= (tp << 16) | (tp >> (64 - 16))
		tp = kTab0[(((g[(u+3)&0xF]) >> 24) & 0xFF)]
		t[u] ^= (tp << 24) | (tp >> (64 - 24))
		t[u] ^= kTab4[(((g[(u+4)&0xF]) >> 32) & 0xFF)]
		tp = kTab4[(((g[(u+5)&0xF]) >> 40) & 0xFF)]
		t[u] ^= (tp << 8) | (tp >> (64 - 8))
		tp = kTab4[(((g[(u+6)&0xF]) >> 48) & 0xFF)]
		t[u] ^= (tp << 16) | (tp >> (64 - 16))
		tp = kTab4[(((g[(u+11)&0xF]) >> 56) & 0xFF)]
		t[u] ^= (tp << 24) | (tp >> (64 - 24))

		ix = u + 1
		t[ix] = kTab0[((g[(u+1)&0xF]) & 0xFF)]
		tp = kTab0[(((g[(u+2)&0xF]) >> 8) & 0xFF)]
		t[ix] ^= (tp << 8) | (tp >> (64 - 8))
		tp = kTab0[(((g[(u+3)&0xF]) >> 16) & 0xFF)]
		t[ix] ^= (tp << 16) | (tp >> (64 - 16))
		tp = kTab0[(((g[(u+4)&0xF]) >> 24) & 0xFF)]
		t[ix] ^= (tp << 24) | (tp >> (64 - 24))
		t[ix] ^= kTab4[(((g[(u+5)&0xF]) >> 32) & 0xFF)]
		tp = kTab4[(((g[(u+6)&0xF]) >> 40) & 0xFF)]
		t[ix] ^= (tp << 8) | (tp >> (64 - 8))
		tp = kTab4[(((g[(u+7)&0xF]) >> 48) & 0xFF)]
		t[ix] ^= (tp << 16) | (tp >> (64 - 16))
		tp = kTab4[(((g[(u+12)&0xF]) >> 56) & 0xFF)]
		t[ix] ^= (tp << 24) | (tp >> (64 - 24))

		ix = u + 2
		t[ix] = kTab0[((g[(u+2)&0xF]) & 0xFF)]
		tp = kTab0[(((g[(u+3)&0xF]) >> 8) & 0xFF)]
		t[ix] ^= (tp << 8) | (tp >> (64 - 8))
		tp = kTab0[(((g[(u+4)&0xF]) >> 16) & 0xFF)]
		t[ix] ^= (tp << 16) | (tp >> (64 - 16))
		tp = kTab0[(((g[(u+5)&0xF]) >> 24) & 0xFF)]
		t[ix] ^= (tp << 24) | (tp >> (64 - 24))
		t[ix] ^= kTab4[(((g[(u+6)&0xF]) >> 32) & 0xFF)]
		tp = kTab4[(((g[(u+7)&0xF]) >> 40) & 0xFF)]
		t[ix] ^= (tp << 8) | (tp >> (64 - 8))
		tp = kTab4[(((g[(u+8)&0xF]) >> 48) & 0xFF)]
		t[ix] ^= (tp << 16) | (tp >> (64 - 16))
		tp = kTab4[(((g[(u+13)&0xF]) >> 56) & 0xFF)]
		t[ix] ^= (tp << 24) | (tp >> (64 - 24))

		ix = u + 3
		t[ix] = kTab0[((g[(u+3)&0xF]) & 0xFF)]
		tp = kTab0[(((g[(u+4)&0xF]) >> 8) & 0xFF)]
		t[ix] ^= (tp << 8) | (tp >> (64 - 8))
		tp = kTab0[(((g[(u+5)&0xF]) >> 16) & 0xFF)]
		t[ix] ^= (tp << 16) | (tp >> (64 - 16))
		tp = kTab0[(((g[(u+6)&0xF]) >> 24) & 0xFF)]
		t[ix] ^= (tp << 24) | (tp >> (64 - 24))
		t[ix] ^= kTab4[(((g[(u+7)&0xF]) >> 32) & 0xFF)]
		tp = kTab4[(((g[(u+8)&0xF]) >> 40) & 0xFF)]
		t[ix] ^= (tp << 8) | (tp >> (64 - 8))
		tp = kTab4[(((g[(u+9)&0xF]) >> 48) & 0xFF)]
		t[ix] ^= (tp << 16) | (tp >> (64 - 16))
		tp = kTab4[(((g[(u+14)&0xF]) >> 56) & 0xFF)]
		t[ix] ^= (tp << 24) | (tp >> (64 - 24))
	}

	copy(g, t[:])
}

func mRounds(r uint64, m []uint64) {
	m[0x0] ^= ((r << 56) ^ (^(uint64(0x00) << 56)))
	m[0x1] ^= ((r << 56) ^ (^(uint64(0x10) << 56)))
	m[0x2] ^= ((r << 56) ^ (^(uint64(0x20) << 56)))
	m[0x3] ^= ((r << 56) ^ (^(uint64(0x30) << 56)))
	m[0x4] ^= ((r << 56) ^ (^(uint64(0x40) << 56)))
	m[0x5] ^= ((r << 56) ^ (^(uint64(0x50) << 56)))
	m[0x6] ^= ((r << 56) ^ (^(uint64(0x60) << 56)))
	m[0x7] ^= ((r << 56) ^ (^(uint64(0x70) << 56)))
	m[0x8] ^= ((r << 56) ^ (^(uint64(0x80) << 56)))
	m[0x9] ^= ((r << 56) ^ (^(uint64(0x90) << 56)))
	m[0xA] ^= ((r << 56) ^ (^(uint64(0xA0) << 56)))
	m[0xB] ^= ((r << 56) ^ (^(uint64(0xB0) << 56)))
	m[0xC] ^= ((r << 56) ^ (^(uint64(0xC0) << 56)))
	m[0xD] ^= ((r << 56) ^ (^(uint64(0xD0) << 56)))
	m[0xE] ^= ((r << 56) ^ (^(uint64(0xE0) << 56)))
	m[0xF] ^= ((r << 56) ^ (^(uint64(0xF0) << 56)))

	var t [16]uint64
	var tp, ix uint64

	for u := uint64(0); u < 16; u += 4 {
		t[u] = kTab0[((m[(u+1)&0xF]) & 0xFF)]
		tp = kTab0[(((m[(u+3)&0xF]) >> 8) & 0xFF)]
		t[u] ^= (tp << 8) | (tp >> (64 - 8))
		tp = kTab0[(((m[(u+5)&0xF]) >> 16) & 0xFF)]
		t[u] ^= (tp << 16) | (tp >> (64 - 16))
		tp = kTab0[(((m[(u+11)&0xF]) >> 24) & 0xFF)]
		t[u] ^= (tp << 24) | (tp >> (64 - 24))
		t[u] ^= kTab4[(((m[(u+0)&0xF]) >> 32) & 0xFF)]
		tp = kTab4[(((m[(u+2)&0xF]) >> 40) & 0xFF)]
		t[u] ^= (tp << 8) | (tp >> (64 - 8))
		tp = kTab4[(((m[(u+4)&0xF]) >> 48) & 0xFF)]
		t[u] ^= (tp << 16) | (tp >> (64 - 16))
		tp = kTab4[(((m[(u+6)&0xF]) >> 56) & 0xFF)]
		t[u] ^= (tp << 24) | (tp >> (64 - 24))

		ix = u + 1
		t[ix] = kTab0[((m[(u+2)&0xF]) & 0xFF)]
		tp = kTab0[(((m[(u+4)&0xF]) >> 8) & 0xFF)]
		t[ix] ^= (tp << 8) | (tp >> (64 - 8))
		tp = kTab0[(((m[(u+6)&0xF]) >> 16) & 0xFF)]
		t[ix] ^= (tp << 16) | (tp >> (64 - 16))
		tp = kTab0[(((m[(u+12)&0xF]) >> 24) & 0xFF)]
		t[ix] ^= (tp << 24) | (tp >> (64 - 24))
		t[ix] ^= kTab4[(((m[(u+1)&0xF]) >> 32) & 0xFF)]
		tp = kTab4[(((m[(u+3)&0xF]) >> 40) & 0xFF)]
		t[ix] ^= (tp << 8) | (tp >> (64 - 8))
		tp = kTab4[(((m[(u+5)&0xF]) >> 48) & 0xFF)]
		t[ix] ^= (tp << 16) | (tp >> (64 - 16))
		tp = kTab4[(((m[(u+7)&0xF]) >> 56) & 0xFF)]
		t[ix] ^= (tp << 24) | (tp >> (64 - 24))

		ix = u + 2
		t[ix] = kTab0[((m[(u+3)&0xF]) & 0xFF)]
		tp = kTab0[(((m[(u+5)&0xF]) >> 8) & 0xFF)]
		t[ix] ^= (tp << 8) | (tp >> (64 - 8))
		tp = kTab0[(((m[(u+7)&0xF]) >> 16) & 0xFF)]
		t[ix] ^= (tp << 16) | (tp >> (64 - 16))
		tp = kTab0[(((m[(u+13)&0xF]) >> 24) & 0xFF)]
		t[ix] ^= (tp << 24) | (tp >> (64 - 24))
		t[ix] ^= kTab4[(((m[(u+2)&0xF]) >> 32) & 0xFF)]
		tp = kTab4[(((m[(u+4)&0xF]) >> 40) & 0xFF)]
		t[ix] ^= (tp << 8) | (tp >> (64 - 8))
		tp = kTab4[(((m[(u+6)&0xF]) >> 48) & 0xFF)]
		t[ix] ^= (tp << 16) | (tp >> (64 - 16))
		tp = kTab4[(((m[(u+8)&0xF]) >> 56) & 0xFF)]
		t[ix] ^= (tp << 24) | (tp >> (64 - 24))

		ix = u + 3
		t[ix] = kTab0[((m[(u+4)&0xF]) & 0xFF)]
		tp = kTab0[(((m[(u+6)&0xF]) >> 8) & 0xFF)]
		t[ix] ^= (tp << 8) | (tp >> (64 - 8))
		tp = kTab0[(((m[(u+8)&0xF]) >> 16) & 0xFF)]
		t[ix] ^= (tp << 16) | (tp >> (64 - 16))
		tp = kTab0[(((m[(u+14)&0xF]) >> 24) & 0xFF)]
		t[ix] ^= (tp << 24) | (tp >> (64 - 24))
		t[ix] ^= kTab4[(((m[(u+3)&0xF]) >> 32) & 0xFF)]
		tp = kTab4[(((m[(u+5)&0xF]) >> 40) & 0xFF)]
		t[ix] ^= (tp << 8) | (tp >> (64 - 8))
		tp = kTab4[(((m[(u+7)&0xF]) >> 48) & 0xFF)]
		t[ix] ^= (tp << 16) | (tp >> (64 - 16))
		tp = kTab4[(((m[(u+9)&0xF]) >> 56) & 0xFF)]
		t[ix] ^= (tp << 24) | (tp >> (64 - 24))
	}

	copy(m, t[:])
}

////////////////

var kTab0 = []uint64{
	uint64(0xc6a597f4a5f432c6), uint64(0xf884eb9784976ff8),
	uint64(0xee99c7b099b05eee), uint64(0xf68df78c8d8c7af6),
	uint64(0xff0de5170d17e8ff), uint64(0xd6bdb7dcbddc0ad6),
	uint64(0xdeb1a7c8b1c816de), uint64(0x915439fc54fc6d91),
	uint64(0x6050c0f050f09060), uint64(0x0203040503050702),
	uint64(0xcea987e0a9e02ece), uint64(0x567dac877d87d156),
	uint64(0xe719d52b192bcce7), uint64(0xb56271a662a613b5),
	uint64(0x4de69a31e6317c4d), uint64(0xec9ac3b59ab559ec),
	uint64(0x8f4505cf45cf408f), uint64(0x1f9d3ebc9dbca31f),
	uint64(0x894009c040c04989), uint64(0xfa87ef92879268fa),
	uint64(0xef15c53f153fd0ef), uint64(0xb2eb7f26eb2694b2),
	uint64(0x8ec90740c940ce8e), uint64(0xfb0bed1d0b1de6fb),
	uint64(0x41ec822fec2f6e41), uint64(0xb3677da967a91ab3),
	uint64(0x5ffdbe1cfd1c435f), uint64(0x45ea8a25ea256045),
	uint64(0x23bf46dabfdaf923), uint64(0x53f7a602f7025153),
	uint64(0xe496d3a196a145e4), uint64(0x9b5b2ded5bed769b),
	uint64(0x75c2ea5dc25d2875), uint64(0xe11cd9241c24c5e1),
	uint64(0x3dae7ae9aee9d43d), uint64(0x4c6a98be6abef24c),
	uint64(0x6c5ad8ee5aee826c), uint64(0x7e41fcc341c3bd7e),
	uint64(0xf502f1060206f3f5), uint64(0x834f1dd14fd15283),
	uint64(0x685cd0e45ce48c68), uint64(0x51f4a207f4075651),
	uint64(0xd134b95c345c8dd1), uint64(0xf908e9180818e1f9),
	uint64(0xe293dfae93ae4ce2), uint64(0xab734d9573953eab),
	uint64(0x6253c4f553f59762), uint64(0x2a3f54413f416b2a),
	uint64(0x080c10140c141c08), uint64(0x955231f652f66395),
	uint64(0x46658caf65afe946), uint64(0x9d5e21e25ee27f9d),
	uint64(0x3028607828784830), uint64(0x37a16ef8a1f8cf37),
	uint64(0x0a0f14110f111b0a), uint64(0x2fb55ec4b5c4eb2f),
	uint64(0x0e091c1b091b150e), uint64(0x2436485a365a7e24),
	uint64(0x1b9b36b69bb6ad1b), uint64(0xdf3da5473d4798df),
	uint64(0xcd26816a266aa7cd), uint64(0x4e699cbb69bbf54e),
	uint64(0x7fcdfe4ccd4c337f), uint64(0xea9fcfba9fba50ea),
	uint64(0x121b242d1b2d3f12), uint64(0x1d9e3ab99eb9a41d),
	uint64(0x5874b09c749cc458), uint64(0x342e68722e724634),
	uint64(0x362d6c772d774136), uint64(0xdcb2a3cdb2cd11dc),
	uint64(0xb4ee7329ee299db4), uint64(0x5bfbb616fb164d5b),
	uint64(0xa4f65301f601a5a4), uint64(0x764decd74dd7a176),
	uint64(0xb76175a361a314b7), uint64(0x7dcefa49ce49347d),
	uint64(0x527ba48d7b8ddf52), uint64(0xdd3ea1423e429fdd),
	uint64(0x5e71bc937193cd5e), uint64(0x139726a297a2b113),
	uint64(0xa6f55704f504a2a6), uint64(0xb96869b868b801b9),
	uint64(0x0000000000000000), uint64(0xc12c99742c74b5c1),
	uint64(0x406080a060a0e040), uint64(0xe31fdd211f21c2e3),
	uint64(0x79c8f243c8433a79), uint64(0xb6ed772ced2c9ab6),
	uint64(0xd4beb3d9bed90dd4), uint64(0x8d4601ca46ca478d),
	uint64(0x67d9ce70d9701767), uint64(0x724be4dd4bddaf72),
	uint64(0x94de3379de79ed94), uint64(0x98d42b67d467ff98),
	uint64(0xb0e87b23e82393b0), uint64(0x854a11de4ade5b85),
	uint64(0xbb6b6dbd6bbd06bb), uint64(0xc52a917e2a7ebbc5),
	uint64(0x4fe59e34e5347b4f), uint64(0xed16c13a163ad7ed),
	uint64(0x86c51754c554d286), uint64(0x9ad72f62d762f89a),
	uint64(0x6655ccff55ff9966), uint64(0x119422a794a7b611),
	uint64(0x8acf0f4acf4ac08a), uint64(0xe910c9301030d9e9),
	uint64(0x0406080a060a0e04), uint64(0xfe81e798819866fe),
	uint64(0xa0f05b0bf00baba0), uint64(0x7844f0cc44ccb478),
	uint64(0x25ba4ad5bad5f025), uint64(0x4be3963ee33e754b),
	uint64(0xa2f35f0ef30eaca2), uint64(0x5dfeba19fe19445d),
	uint64(0x80c01b5bc05bdb80), uint64(0x058a0a858a858005),
	uint64(0x3fad7eecadecd33f), uint64(0x21bc42dfbcdffe21),
	uint64(0x7048e0d848d8a870), uint64(0xf104f90c040cfdf1),
	uint64(0x63dfc67adf7a1963), uint64(0x77c1ee58c1582f77),
	uint64(0xaf75459f759f30af), uint64(0x426384a563a5e742),
	uint64(0x2030405030507020), uint64(0xe51ad12e1a2ecbe5),
	uint64(0xfd0ee1120e12effd), uint64(0xbf6d65b76db708bf),
	uint64(0x814c19d44cd45581), uint64(0x1814303c143c2418),
	uint64(0x26354c5f355f7926), uint64(0xc32f9d712f71b2c3),
	uint64(0xbee16738e13886be), uint64(0x35a26afda2fdc835),
	uint64(0x88cc0b4fcc4fc788), uint64(0x2e395c4b394b652e),
	uint64(0x93573df957f96a93), uint64(0x55f2aa0df20d5855),
	uint64(0xfc82e39d829d61fc), uint64(0x7a47f4c947c9b37a),
	uint64(0xc8ac8befacef27c8), uint64(0xbae76f32e73288ba),
	uint64(0x322b647d2b7d4f32), uint64(0xe695d7a495a442e6),
	uint64(0xc0a09bfba0fb3bc0), uint64(0x199832b398b3aa19),
	uint64(0x9ed12768d168f69e), uint64(0xa37f5d817f8122a3),
	uint64(0x446688aa66aaee44), uint64(0x547ea8827e82d654),
	uint64(0x3bab76e6abe6dd3b), uint64(0x0b83169e839e950b),
	uint64(0x8cca0345ca45c98c), uint64(0xc729957b297bbcc7),
	uint64(0x6bd3d66ed36e056b), uint64(0x283c50443c446c28),
	uint64(0xa779558b798b2ca7), uint64(0xbce2633de23d81bc),
	uint64(0x161d2c271d273116), uint64(0xad76419a769a37ad),
	uint64(0xdb3bad4d3b4d96db), uint64(0x6456c8fa56fa9e64),
	uint64(0x744ee8d24ed2a674), uint64(0x141e28221e223614),
	uint64(0x92db3f76db76e492), uint64(0x0c0a181e0a1e120c),
	uint64(0x486c90b46cb4fc48), uint64(0xb8e46b37e4378fb8),
	uint64(0x9f5d25e75de7789f), uint64(0xbd6e61b26eb20fbd),
	uint64(0x43ef862aef2a6943), uint64(0xc4a693f1a6f135c4),
	uint64(0x39a872e3a8e3da39), uint64(0x31a462f7a4f7c631),
	uint64(0xd337bd5937598ad3), uint64(0xf28bff868b8674f2),
	uint64(0xd532b156325683d5), uint64(0x8b430dc543c54e8b),
	uint64(0x6e59dceb59eb856e), uint64(0xdab7afc2b7c218da),
	uint64(0x018c028f8c8f8e01), uint64(0xb16479ac64ac1db1),
	uint64(0x9cd2236dd26df19c), uint64(0x49e0923be03b7249),
	uint64(0xd8b4abc7b4c71fd8), uint64(0xacfa4315fa15b9ac),
	uint64(0xf307fd090709faf3), uint64(0xcf25856f256fa0cf),
	uint64(0xcaaf8feaafea20ca), uint64(0xf48ef3898e897df4),
	uint64(0x47e98e20e9206747), uint64(0x1018202818283810),
	uint64(0x6fd5de64d5640b6f), uint64(0xf088fb83888373f0),
	uint64(0x4a6f94b16fb1fb4a), uint64(0x5c72b8967296ca5c),
	uint64(0x3824706c246c5438), uint64(0x57f1ae08f1085f57),
	uint64(0x73c7e652c7522173), uint64(0x975135f351f36497),
	uint64(0xcb238d652365aecb), uint64(0xa17c59847c8425a1),
	uint64(0xe89ccbbf9cbf57e8), uint64(0x3e217c6321635d3e),
	uint64(0x96dd377cdd7cea96), uint64(0x61dcc27fdc7f1e61),
	uint64(0x0d861a9186919c0d), uint64(0x0f851e9485949b0f),
	uint64(0xe090dbab90ab4be0), uint64(0x7c42f8c642c6ba7c),
	uint64(0x71c4e257c4572671), uint64(0xccaa83e5aae529cc),
	uint64(0x90d83b73d873e390), uint64(0x06050c0f050f0906),
	uint64(0xf701f5030103f4f7), uint64(0x1c12383612362a1c),
	uint64(0xc2a39ffea3fe3cc2), uint64(0x6a5fd4e15fe18b6a),
	uint64(0xaef94710f910beae), uint64(0x69d0d26bd06b0269),
	uint64(0x17912ea891a8bf17), uint64(0x995829e858e87199),
	uint64(0x3a2774692769533a), uint64(0x27b94ed0b9d0f727),
	uint64(0xd938a948384891d9), uint64(0xeb13cd351335deeb),
	uint64(0x2bb356ceb3cee52b), uint64(0x2233445533557722),
	uint64(0xd2bbbfd6bbd604d2), uint64(0xa9704990709039a9),
	uint64(0x07890e8089808707), uint64(0x33a766f2a7f2c133),
	uint64(0x2db65ac1b6c1ec2d), uint64(0x3c22786622665a3c),
	uint64(0x15922aad92adb815), uint64(0xc92089602060a9c9),
	uint64(0x874915db49db5c87), uint64(0xaaff4f1aff1ab0aa),
	uint64(0x5078a0887888d850), uint64(0xa57a518e7a8e2ba5),
	uint64(0x038f068a8f8a8903), uint64(0x59f8b213f8134a59),
	uint64(0x0980129b809b9209), uint64(0x1a1734391739231a),
	uint64(0x65daca75da751065), uint64(0xd731b553315384d7),
	uint64(0x84c61351c651d584), uint64(0xd0b8bbd3b8d303d0),
	uint64(0x82c31f5ec35edc82), uint64(0x29b052cbb0cbe229),
	uint64(0x5a77b4997799c35a), uint64(0x1e113c3311332d1e),
	uint64(0x7bcbf646cb463d7b), uint64(0xa8fc4b1ffc1fb7a8),
	uint64(0x6dd6da61d6610c6d), uint64(0x2c3a584e3a4e622c),
}

var kTab4 = []uint64{
	uint64(0xa5f432c6c6a597f4), uint64(0x84976ff8f884eb97),
	uint64(0x99b05eeeee99c7b0), uint64(0x8d8c7af6f68df78c),
	uint64(0x0d17e8ffff0de517), uint64(0xbddc0ad6d6bdb7dc),
	uint64(0xb1c816dedeb1a7c8), uint64(0x54fc6d91915439fc),
	uint64(0x50f090606050c0f0), uint64(0x0305070202030405),
	uint64(0xa9e02ececea987e0), uint64(0x7d87d156567dac87),
	uint64(0x192bcce7e719d52b), uint64(0x62a613b5b56271a6),
	uint64(0xe6317c4d4de69a31), uint64(0x9ab559ecec9ac3b5),
	uint64(0x45cf408f8f4505cf), uint64(0x9dbca31f1f9d3ebc),
	uint64(0x40c04989894009c0), uint64(0x879268fafa87ef92),
	uint64(0x153fd0efef15c53f), uint64(0xeb2694b2b2eb7f26),
	uint64(0xc940ce8e8ec90740), uint64(0x0b1de6fbfb0bed1d),
	uint64(0xec2f6e4141ec822f), uint64(0x67a91ab3b3677da9),
	uint64(0xfd1c435f5ffdbe1c), uint64(0xea25604545ea8a25),
	uint64(0xbfdaf92323bf46da), uint64(0xf702515353f7a602),
	uint64(0x96a145e4e496d3a1), uint64(0x5bed769b9b5b2ded),
	uint64(0xc25d287575c2ea5d), uint64(0x1c24c5e1e11cd924),
	uint64(0xaee9d43d3dae7ae9), uint64(0x6abef24c4c6a98be),
	uint64(0x5aee826c6c5ad8ee), uint64(0x41c3bd7e7e41fcc3),
	uint64(0x0206f3f5f502f106), uint64(0x4fd15283834f1dd1),
	uint64(0x5ce48c68685cd0e4), uint64(0xf407565151f4a207),
	uint64(0x345c8dd1d134b95c), uint64(0x0818e1f9f908e918),
	uint64(0x93ae4ce2e293dfae), uint64(0x73953eabab734d95),
	uint64(0x53f597626253c4f5), uint64(0x3f416b2a2a3f5441),
	uint64(0x0c141c08080c1014), uint64(0x52f66395955231f6),
	uint64(0x65afe94646658caf), uint64(0x5ee27f9d9d5e21e2),
	uint64(0x2878483030286078), uint64(0xa1f8cf3737a16ef8),
	uint64(0x0f111b0a0a0f1411), uint64(0xb5c4eb2f2fb55ec4),
	uint64(0x091b150e0e091c1b), uint64(0x365a7e242436485a),
	uint64(0x9bb6ad1b1b9b36b6), uint64(0x3d4798dfdf3da547),
	uint64(0x266aa7cdcd26816a), uint64(0x69bbf54e4e699cbb),
	uint64(0xcd4c337f7fcdfe4c), uint64(0x9fba50eaea9fcfba),
	uint64(0x1b2d3f12121b242d), uint64(0x9eb9a41d1d9e3ab9),
	uint64(0x749cc4585874b09c), uint64(0x2e724634342e6872),
	uint64(0x2d774136362d6c77), uint64(0xb2cd11dcdcb2a3cd),
	uint64(0xee299db4b4ee7329), uint64(0xfb164d5b5bfbb616),
	uint64(0xf601a5a4a4f65301), uint64(0x4dd7a176764decd7),
	uint64(0x61a314b7b76175a3), uint64(0xce49347d7dcefa49),
	uint64(0x7b8ddf52527ba48d), uint64(0x3e429fdddd3ea142),
	uint64(0x7193cd5e5e71bc93), uint64(0x97a2b113139726a2),
	uint64(0xf504a2a6a6f55704), uint64(0x68b801b9b96869b8),
	uint64(0x0000000000000000), uint64(0x2c74b5c1c12c9974),
	uint64(0x60a0e040406080a0), uint64(0x1f21c2e3e31fdd21),
	uint64(0xc8433a7979c8f243), uint64(0xed2c9ab6b6ed772c),
	uint64(0xbed90dd4d4beb3d9), uint64(0x46ca478d8d4601ca),
	uint64(0xd970176767d9ce70), uint64(0x4bddaf72724be4dd),
	uint64(0xde79ed9494de3379), uint64(0xd467ff9898d42b67),
	uint64(0xe82393b0b0e87b23), uint64(0x4ade5b85854a11de),
	uint64(0x6bbd06bbbb6b6dbd), uint64(0x2a7ebbc5c52a917e),
	uint64(0xe5347b4f4fe59e34), uint64(0x163ad7eded16c13a),
	uint64(0xc554d28686c51754), uint64(0xd762f89a9ad72f62),
	uint64(0x55ff99666655ccff), uint64(0x94a7b611119422a7),
	uint64(0xcf4ac08a8acf0f4a), uint64(0x1030d9e9e910c930),
	uint64(0x060a0e040406080a), uint64(0x819866fefe81e798),
	uint64(0xf00baba0a0f05b0b), uint64(0x44ccb4787844f0cc),
	uint64(0xbad5f02525ba4ad5), uint64(0xe33e754b4be3963e),
	uint64(0xf30eaca2a2f35f0e), uint64(0xfe19445d5dfeba19),
	uint64(0xc05bdb8080c01b5b), uint64(0x8a858005058a0a85),
	uint64(0xadecd33f3fad7eec), uint64(0xbcdffe2121bc42df),
	uint64(0x48d8a8707048e0d8), uint64(0x040cfdf1f104f90c),
	uint64(0xdf7a196363dfc67a), uint64(0xc1582f7777c1ee58),
	uint64(0x759f30afaf75459f), uint64(0x63a5e742426384a5),
	uint64(0x3050702020304050), uint64(0x1a2ecbe5e51ad12e),
	uint64(0x0e12effdfd0ee112), uint64(0x6db708bfbf6d65b7),
	uint64(0x4cd45581814c19d4), uint64(0x143c24181814303c),
	uint64(0x355f792626354c5f), uint64(0x2f71b2c3c32f9d71),
	uint64(0xe13886bebee16738), uint64(0xa2fdc83535a26afd),
	uint64(0xcc4fc78888cc0b4f), uint64(0x394b652e2e395c4b),
	uint64(0x57f96a9393573df9), uint64(0xf20d585555f2aa0d),
	uint64(0x829d61fcfc82e39d), uint64(0x47c9b37a7a47f4c9),
	uint64(0xacef27c8c8ac8bef), uint64(0xe73288babae76f32),
	uint64(0x2b7d4f32322b647d), uint64(0x95a442e6e695d7a4),
	uint64(0xa0fb3bc0c0a09bfb), uint64(0x98b3aa19199832b3),
	uint64(0xd168f69e9ed12768), uint64(0x7f8122a3a37f5d81),
	uint64(0x66aaee44446688aa), uint64(0x7e82d654547ea882),
	uint64(0xabe6dd3b3bab76e6), uint64(0x839e950b0b83169e),
	uint64(0xca45c98c8cca0345), uint64(0x297bbcc7c729957b),
	uint64(0xd36e056b6bd3d66e), uint64(0x3c446c28283c5044),
	uint64(0x798b2ca7a779558b), uint64(0xe23d81bcbce2633d),
	uint64(0x1d273116161d2c27), uint64(0x769a37adad76419a),
	uint64(0x3b4d96dbdb3bad4d), uint64(0x56fa9e646456c8fa),
	uint64(0x4ed2a674744ee8d2), uint64(0x1e223614141e2822),
	uint64(0xdb76e49292db3f76), uint64(0x0a1e120c0c0a181e),
	uint64(0x6cb4fc48486c90b4), uint64(0xe4378fb8b8e46b37),
	uint64(0x5de7789f9f5d25e7), uint64(0x6eb20fbdbd6e61b2),
	uint64(0xef2a694343ef862a), uint64(0xa6f135c4c4a693f1),
	uint64(0xa8e3da3939a872e3), uint64(0xa4f7c63131a462f7),
	uint64(0x37598ad3d337bd59), uint64(0x8b8674f2f28bff86),
	uint64(0x325683d5d532b156), uint64(0x43c54e8b8b430dc5),
	uint64(0x59eb856e6e59dceb), uint64(0xb7c218dadab7afc2),
	uint64(0x8c8f8e01018c028f), uint64(0x64ac1db1b16479ac),
	uint64(0xd26df19c9cd2236d), uint64(0xe03b724949e0923b),
	uint64(0xb4c71fd8d8b4abc7), uint64(0xfa15b9acacfa4315),
	uint64(0x0709faf3f307fd09), uint64(0x256fa0cfcf25856f),
	uint64(0xafea20cacaaf8fea), uint64(0x8e897df4f48ef389),
	uint64(0xe920674747e98e20), uint64(0x1828381010182028),
	uint64(0xd5640b6f6fd5de64), uint64(0x888373f0f088fb83),
	uint64(0x6fb1fb4a4a6f94b1), uint64(0x7296ca5c5c72b896),
	uint64(0x246c54383824706c), uint64(0xf1085f5757f1ae08),
	uint64(0xc752217373c7e652), uint64(0x51f36497975135f3),
	uint64(0x2365aecbcb238d65), uint64(0x7c8425a1a17c5984),
	uint64(0x9cbf57e8e89ccbbf), uint64(0x21635d3e3e217c63),
	uint64(0xdd7cea9696dd377c), uint64(0xdc7f1e6161dcc27f),
	uint64(0x86919c0d0d861a91), uint64(0x85949b0f0f851e94),
	uint64(0x90ab4be0e090dbab), uint64(0x42c6ba7c7c42f8c6),
	uint64(0xc457267171c4e257), uint64(0xaae529ccccaa83e5),
	uint64(0xd873e39090d83b73), uint64(0x050f090606050c0f),
	uint64(0x0103f4f7f701f503), uint64(0x12362a1c1c123836),
	uint64(0xa3fe3cc2c2a39ffe), uint64(0x5fe18b6a6a5fd4e1),
	uint64(0xf910beaeaef94710), uint64(0xd06b026969d0d26b),
	uint64(0x91a8bf1717912ea8), uint64(0x58e87199995829e8),
	uint64(0x2769533a3a277469), uint64(0xb9d0f72727b94ed0),
	uint64(0x384891d9d938a948), uint64(0x1335deebeb13cd35),
	uint64(0xb3cee52b2bb356ce), uint64(0x3355772222334455),
	uint64(0xbbd604d2d2bbbfd6), uint64(0x709039a9a9704990),
	uint64(0x8980870707890e80), uint64(0xa7f2c13333a766f2),
	uint64(0xb6c1ec2d2db65ac1), uint64(0x22665a3c3c227866),
	uint64(0x92adb81515922aad), uint64(0x2060a9c9c9208960),
	uint64(0x49db5c87874915db), uint64(0xff1ab0aaaaff4f1a),
	uint64(0x7888d8505078a088), uint64(0x7a8e2ba5a57a518e),
	uint64(0x8f8a8903038f068a), uint64(0xf8134a5959f8b213),
	uint64(0x809b92090980129b), uint64(0x1739231a1a173439),
	uint64(0xda75106565daca75), uint64(0x315384d7d731b553),
	uint64(0xc651d58484c61351), uint64(0xb8d303d0d0b8bbd3),
	uint64(0xc35edc8282c31f5e), uint64(0xb0cbe22929b052cb),
	uint64(0x7799c35a5a77b499), uint64(0x11332d1e1e113c33),
	uint64(0xcb463d7b7bcbf646), uint64(0xfc1fb7a8a8fc4b1f),
	uint64(0xd6610c6d6dd6da61), uint64(0x3a4e622c2c3a584e),
}

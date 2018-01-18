// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package luffa

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

	h [5][8]uint32

	b [32]byte
}

// New returns a new digest compute a LUFFA512 hash.
func New() hash.Digest {
	ref := &digest{}
	ref.Reset()
	return ref
}

////////////////

// Reset resets the digest to its initial state.
func (ref *digest) Reset() {
	ref.ptr = 0
	for x := range kInit {
		for y := range kInit[x] {
			ref.h[x][y] = kInit[x][y]
		}
	}
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
		copy(ref.b[ptr:], src)
		ref.ptr += sln
		return int(sln), nil
	}

	var V00, V01, V02, V03, V04, V05, V06, V07 uint32
	var V10, V11, V12, V13, V14, V15, V16, V17 uint32
	var V20, V21, V22, V23, V24, V25, V26, V27 uint32
	var V30, V31, V32, V33, V34, V35, V36, V37 uint32
	var V40, V41, V42, V43, V44, V45, V46, V47 uint32

	V00 = ref.h[0][0]
	V01 = ref.h[0][1]
	V02 = ref.h[0][2]
	V03 = ref.h[0][3]
	V04 = ref.h[0][4]
	V05 = ref.h[0][5]
	V06 = ref.h[0][6]
	V07 = ref.h[0][7]
	V10 = ref.h[1][0]
	V11 = ref.h[1][1]
	V12 = ref.h[1][2]
	V13 = ref.h[1][3]
	V14 = ref.h[1][4]
	V15 = ref.h[1][5]
	V16 = ref.h[1][6]
	V17 = ref.h[1][7]
	V20 = ref.h[2][0]
	V21 = ref.h[2][1]
	V22 = ref.h[2][2]
	V23 = ref.h[2][3]
	V24 = ref.h[2][4]
	V25 = ref.h[2][5]
	V26 = ref.h[2][6]
	V27 = ref.h[2][7]
	V30 = ref.h[3][0]
	V31 = ref.h[3][1]
	V32 = ref.h[3][2]
	V33 = ref.h[3][3]
	V34 = ref.h[3][4]
	V35 = ref.h[3][5]
	V36 = ref.h[3][6]
	V37 = ref.h[3][7]
	V40 = ref.h[4][0]
	V41 = ref.h[4][1]
	V42 = ref.h[4][2]
	V43 = ref.h[4][3]
	V44 = ref.h[4][4]
	V45 = ref.h[4][5]
	V46 = ref.h[4][6]
	V47 = ref.h[4][7]

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
			{
				var ts uint32
				var M0, M1, M2, M3, M4, M5, M6, M7 uint32
				var a0, a1, a2, a3, a4, a5, a6, a7 uint32
				var b0, b1, b2, b3, b4, b5, b6, b7 uint32

				M0 = decUInt32be(buf[0:])
				M1 = decUInt32be(buf[4:])
				M2 = decUInt32be(buf[8:])
				M3 = decUInt32be(buf[12:])
				M4 = decUInt32be(buf[16:])
				M5 = decUInt32be(buf[20:])
				M6 = decUInt32be(buf[24:])
				M7 = decUInt32be(buf[28:])

				a0 = V00 ^ V10
				a1 = V01 ^ V11
				a2 = V02 ^ V12
				a3 = V03 ^ V13
				a4 = V04 ^ V14
				a5 = V05 ^ V15
				a6 = V06 ^ V16
				a7 = V07 ^ V17

				b0 = V20 ^ V30
				b1 = V21 ^ V31
				b2 = V22 ^ V32
				b3 = V23 ^ V33
				b4 = V24 ^ V34
				b5 = V25 ^ V35
				b6 = V26 ^ V36
				b7 = V27 ^ V37

				a0 ^= b0
				a1 ^= b1
				a2 ^= b2
				a3 ^= b3
				a4 ^= b4
				a5 ^= b5
				a6 ^= b6
				a7 ^= b7

				a0 ^= V40
				a1 ^= V41
				a2 ^= V42
				a3 ^= V43
				a4 ^= V44
				a5 ^= V45
				a6 ^= V46
				a7 ^= V47

				ts = a7
				a7 = a6
				a6 = a5
				a5 = a4
				a4 = a3 ^ ts
				a3 = a2 ^ ts
				a2 = a1
				a1 = a0 ^ ts
				a0 = ts

				V00 ^= a0
				V01 ^= a1
				V02 ^= a2
				V03 ^= a3
				V04 ^= a4
				V05 ^= a5
				V06 ^= a6
				V07 ^= a7

				V10 ^= a0
				V11 ^= a1
				V12 ^= a2
				V13 ^= a3
				V14 ^= a4
				V15 ^= a5
				V16 ^= a6
				V17 ^= a7

				V20 ^= a0
				V21 ^= a1
				V22 ^= a2
				V23 ^= a3
				V24 ^= a4
				V25 ^= a5
				V26 ^= a6
				V27 ^= a7

				V30 ^= a0
				V31 ^= a1
				V32 ^= a2
				V33 ^= a3
				V34 ^= a4
				V35 ^= a5
				V36 ^= a6
				V37 ^= a7

				V40 ^= a0
				V41 ^= a1
				V42 ^= a2
				V43 ^= a3
				V44 ^= a4
				V45 ^= a5
				V46 ^= a6
				V47 ^= a7

				ts = V07
				b7 = V06
				b6 = V05
				b5 = V04
				b4 = V03 ^ ts
				b3 = V02 ^ ts
				b2 = V01
				b1 = V00 ^ ts
				b0 = ts

				b0 ^= V10
				b1 ^= V11
				b2 ^= V12
				b3 ^= V13
				b4 ^= V14
				b5 ^= V15
				b6 ^= V16
				b7 ^= V17

				ts = V17
				V17 = V16
				V16 = V15
				V15 = V14
				V14 = V13 ^ ts
				V13 = V12 ^ ts
				V12 = V11
				V11 = V10 ^ ts
				V10 = ts

				V10 ^= V20
				V11 ^= V21
				V12 ^= V22
				V13 ^= V23
				V14 ^= V24
				V15 ^= V25
				V16 ^= V26
				V17 ^= V27

				ts = V27
				V27 = V26
				V26 = V25
				V25 = V24
				V24 = V23 ^ ts
				V23 = V22 ^ ts
				V22 = V21
				V21 = V20 ^ ts
				V20 = ts

				V20 ^= V30
				V21 ^= V31
				V22 ^= V32
				V23 ^= V33
				V24 ^= V34
				V25 ^= V35
				V26 ^= V36
				V27 ^= V37

				ts = V37
				V37 = V36
				V36 = V35
				V35 = V34
				V34 = V33 ^ ts
				V33 = V32 ^ ts
				V32 = V31
				V31 = V30 ^ ts
				V30 = ts

				V30 ^= V40
				V31 ^= V41
				V32 ^= V42
				V33 ^= V43
				V34 ^= V44
				V35 ^= V45
				V36 ^= V46
				V37 ^= V47

				ts = V47
				V47 = V46
				V46 = V45
				V45 = V44
				V44 = V43 ^ ts
				V43 = V42 ^ ts
				V42 = V41
				V41 = V40 ^ ts
				V40 = ts

				V40 ^= V00
				V41 ^= V01
				V42 ^= V02
				V43 ^= V03
				V44 ^= V04
				V45 ^= V05
				V46 ^= V06
				V47 ^= V07

				ts = b7
				V07 = b6
				V06 = b5
				V05 = b4
				V04 = b3 ^ ts
				V03 = b2 ^ ts
				V02 = b1
				V01 = b0 ^ ts
				V00 = ts

				V00 ^= V40
				V01 ^= V41
				V02 ^= V42
				V03 ^= V43
				V04 ^= V44
				V05 ^= V45
				V06 ^= V46
				V07 ^= V47

				ts = V47
				V47 = V46
				V46 = V45
				V45 = V44
				V44 = V43 ^ ts
				V43 = V42 ^ ts
				V42 = V41
				V41 = V40 ^ ts
				V40 = ts

				V40 ^= V30
				V41 ^= V31
				V42 ^= V32
				V43 ^= V33
				V44 ^= V34
				V45 ^= V35
				V46 ^= V36
				V47 ^= V37

				ts = V37
				V37 = V36
				V36 = V35
				V35 = V34
				V34 = V33 ^ ts
				V33 = V32 ^ ts
				V32 = V31
				V31 = V30 ^ ts
				V30 = ts

				V30 ^= V20
				V31 ^= V21
				V32 ^= V22
				V33 ^= V23
				V34 ^= V24
				V35 ^= V25
				V36 ^= V26
				V37 ^= V27

				ts = V27
				V27 = V26
				V26 = V25
				V25 = V24
				V24 = V23 ^ ts
				V23 = V22 ^ ts
				V22 = V21
				V21 = V20 ^ ts
				V20 = ts

				V20 ^= V10
				V21 ^= V11
				V22 ^= V12
				V23 ^= V13
				V24 ^= V14
				V25 ^= V15
				V26 ^= V16
				V27 ^= V17

				ts = V17
				V17 = V16
				V16 = V15
				V15 = V14
				V14 = V13 ^ ts
				V13 = V12 ^ ts
				V12 = V11
				V11 = V10 ^ ts
				V10 = ts

				V10 ^= b0
				V11 ^= b1
				V12 ^= b2
				V13 ^= b3
				V14 ^= b4
				V15 ^= b5
				V16 ^= b6
				V17 ^= b7

				V00 ^= M0
				V01 ^= M1
				V02 ^= M2
				V03 ^= M3
				V04 ^= M4
				V05 ^= M5
				V06 ^= M6
				V07 ^= M7

				ts = M7
				M7 = M6
				M6 = M5
				M5 = M4
				M4 = M3 ^ ts
				M3 = M2 ^ ts
				M2 = M1
				M1 = M0 ^ ts
				M0 = ts

				V10 ^= M0
				V11 ^= M1
				V12 ^= M2
				V13 ^= M3
				V14 ^= M4
				V15 ^= M5
				V16 ^= M6
				V17 ^= M7

				ts = M7
				M7 = M6
				M6 = M5
				M5 = M4
				M4 = M3 ^ ts
				M3 = M2 ^ ts
				M2 = M1
				M1 = M0 ^ ts
				M0 = ts

				V20 ^= M0
				V21 ^= M1
				V22 ^= M2
				V23 ^= M3
				V24 ^= M4
				V25 ^= M5
				V26 ^= M6
				V27 ^= M7

				ts = M7
				M7 = M6
				M6 = M5
				M5 = M4
				M4 = M3 ^ ts
				M3 = M2 ^ ts
				M2 = M1
				M1 = M0 ^ ts
				M0 = ts

				V30 ^= M0
				V31 ^= M1
				V32 ^= M2
				V33 ^= M3
				V34 ^= M4
				V35 ^= M5
				V36 ^= M6
				V37 ^= M7

				ts = M7
				M7 = M6
				M6 = M5
				M5 = M4
				M4 = M3 ^ ts
				M3 = M2 ^ ts
				M2 = M1
				M1 = M0 ^ ts
				M0 = ts

				V40 ^= M0
				V41 ^= M1
				V42 ^= M2
				V43 ^= M3
				V44 ^= M4
				V45 ^= M5
				V46 ^= M6
				V47 ^= M7
			}

			{
				var ul, uh, vl, vh, tws uint32
				var W0, W1, W2, W3, W4, W5, W6, W7, tw uint64

				V14 = ((V14 << 1) | (V14 >> (32 - 1)))
				V15 = ((V15 << 1) | (V15 >> (32 - 1)))
				V16 = ((V16 << 1) | (V16 >> (32 - 1)))
				V17 = ((V17 << 1) | (V17 >> (32 - 1)))
				V24 = ((V24 << 2) | (V24 >> (32 - 2)))
				V25 = ((V25 << 2) | (V25 >> (32 - 2)))
				V26 = ((V26 << 2) | (V26 >> (32 - 2)))
				V27 = ((V27 << 2) | (V27 >> (32 - 2)))
				V34 = ((V34 << 3) | (V34 >> (32 - 3)))
				V35 = ((V35 << 3) | (V35 >> (32 - 3)))
				V36 = ((V36 << 3) | (V36 >> (32 - 3)))
				V37 = ((V37 << 3) | (V37 >> (32 - 3)))
				V44 = ((V44 << 4) | (V44 >> (32 - 4)))
				V45 = ((V45 << 4) | (V45 >> (32 - 4)))
				V46 = ((V46 << 4) | (V46 >> (32 - 4)))
				V47 = ((V47 << 4) | (V47 >> (32 - 4)))

				W0 = uint64(V00) | (uint64(V10) << 32)
				W1 = uint64(V01) | (uint64(V11) << 32)
				W2 = uint64(V02) | (uint64(V12) << 32)
				W3 = uint64(V03) | (uint64(V13) << 32)
				W4 = uint64(V04) | (uint64(V14) << 32)
				W5 = uint64(V05) | (uint64(V15) << 32)
				W6 = uint64(V06) | (uint64(V16) << 32)
				W7 = uint64(V07) | (uint64(V17) << 32)

				for r := uintptr(0); r < 8; r++ {
					tw = W0
					W0 |= W1
					W2 ^= W3
					W1 = ^W1
					W0 ^= W3
					W3 &= tw
					W1 ^= W3
					W3 ^= W2
					W2 &= W0
					W0 = ^W0
					W2 ^= W1
					W1 |= W3
					tw ^= W1
					W3 ^= W2
					W2 &= W1
					W1 ^= W0
					W0 = tw

					tw = W5
					W5 |= W6
					W7 ^= W4
					W6 = ^W6
					W5 ^= W4
					W4 &= tw
					W6 ^= W4
					W4 ^= W7
					W7 &= W5
					W5 = ^W5
					W7 ^= W6
					W6 |= W4
					tw ^= W6
					W4 ^= W7
					W7 &= W6
					W6 ^= W5
					W5 = tw

					W4 ^= W0
					ul = uint32(W0)
					uh = uint32((W0 >> 32))
					vl = uint32(W4)
					vh = uint32((W4 >> 32))
					ul = ((ul << 2) | (ul >> (32 - 2))) ^ vl
					vl = ((vl << 14) | (vl >> (32 - 14))) ^ ul
					ul = ((ul << 10) | (ul >> (32 - 10))) ^ vl
					vl = ((vl << 1) | (vl >> (32 - 1)))
					uh = ((uh << 2) | (uh >> (32 - 2))) ^ vh
					vh = ((vh << 14) | (vh >> (32 - 14))) ^ uh
					uh = ((uh << 10) | (uh >> (32 - 10))) ^ vh
					vh = ((vh << 1) | (vh >> (32 - 1)))
					W0 = uint64(ul) | (uint64(uh) << 32)
					W4 = uint64(vl) | (uint64(vh) << 32)

					W5 ^= W1
					ul = uint32(W1)
					uh = uint32((W1 >> 32))
					vl = uint32(W5)
					vh = uint32((W5 >> 32))
					ul = ((ul << 2) | (ul >> (32 - 2))) ^ vl
					vl = ((vl << 14) | (vl >> (32 - 14))) ^ ul
					ul = ((ul << 10) | (ul >> (32 - 10))) ^ vl
					vl = ((vl << 1) | (vl >> (32 - 1)))
					uh = ((uh << 2) | (uh >> (32 - 2))) ^ vh
					vh = ((vh << 14) | (vh >> (32 - 14))) ^ uh
					uh = ((uh << 10) | (uh >> (32 - 10))) ^ vh
					vh = ((vh << 1) | (vh >> (32 - 1)))
					W1 = uint64(ul) | (uint64(uh) << 32)
					W5 = uint64(vl) | (uint64(vh) << 32)

					W6 ^= W2
					ul = uint32(W2)
					uh = uint32((W2 >> 32))
					vl = uint32(W6)
					vh = uint32((W6 >> 32))
					ul = ((ul << 2) | (ul >> (32 - 2))) ^ vl
					vl = ((vl << 14) | (vl >> (32 - 14))) ^ ul
					ul = ((ul << 10) | (ul >> (32 - 10))) ^ vl
					vl = ((vl << 1) | (vl >> (32 - 1)))
					uh = ((uh << 2) | (uh >> (32 - 2))) ^ vh
					vh = ((vh << 14) | (vh >> (32 - 14))) ^ uh
					uh = ((uh << 10) | (uh >> (32 - 10))) ^ vh
					vh = ((vh << 1) | (vh >> (32 - 1)))
					W2 = uint64(ul) | (uint64(uh) << 32)
					W6 = uint64(vl) | (uint64(vh) << 32)

					W7 ^= W3
					ul = uint32(W3)
					uh = uint32((W3 >> 32))
					vl = uint32(W7)
					vh = uint32((W7 >> 32))
					ul = ((ul << 2) | (ul >> (32 - 2))) ^ vl
					vl = ((vl << 14) | (vl >> (32 - 14))) ^ ul
					ul = ((ul << 10) | (ul >> (32 - 10))) ^ vl
					vl = ((vl << 1) | (vl >> (32 - 1)))
					uh = ((uh << 2) | (uh >> (32 - 2))) ^ vh
					vh = ((vh << 14) | (vh >> (32 - 14))) ^ uh
					uh = ((uh << 10) | (uh >> (32 - 10))) ^ vh
					vh = ((vh << 1) | (vh >> (32 - 1)))
					W3 = uint64(ul) | (uint64(uh) << 32)
					W7 = uint64(vl) | (uint64(vh) << 32)

					W0 ^= kRCW010[r]
					W4 ^= kRCW014[r]
				}

				V00 = uint32(W0)
				V10 = uint32((W0 >> 32))
				V01 = uint32(W1)
				V11 = uint32((W1 >> 32))
				V02 = uint32(W2)
				V12 = uint32((W2 >> 32))
				V03 = uint32(W3)
				V13 = uint32((W3 >> 32))
				V04 = uint32(W4)
				V14 = uint32((W4 >> 32))
				V05 = uint32(W5)
				V15 = uint32((W5 >> 32))
				V06 = uint32(W6)
				V16 = uint32((W6 >> 32))
				V07 = uint32(W7)
				V17 = uint32((W7 >> 32))

				W0 = uint64(V20) | (uint64(V30) << 32)
				W1 = uint64(V21) | (uint64(V31) << 32)
				W2 = uint64(V22) | (uint64(V32) << 32)
				W3 = uint64(V23) | (uint64(V33) << 32)
				W4 = uint64(V24) | (uint64(V34) << 32)
				W5 = uint64(V25) | (uint64(V35) << 32)
				W6 = uint64(V26) | (uint64(V36) << 32)
				W7 = uint64(V27) | (uint64(V37) << 32)

				for r := uintptr(0); r < 8; r++ {
					tw = W0
					W0 |= W1
					W2 ^= W3
					W1 = ^W1
					W0 ^= W3
					W3 &= tw
					W1 ^= W3
					W3 ^= W2
					W2 &= W0
					W0 = ^W0
					W2 ^= W1
					W1 |= W3
					tw ^= W1
					W3 ^= W2
					W2 &= W1
					W1 ^= W0
					W0 = tw

					tw = W5
					W5 |= W6
					W7 ^= W4
					W6 = ^W6
					W5 ^= W4
					W4 &= tw
					W6 ^= W4
					W4 ^= W7
					W7 &= W5
					W5 = ^W5
					W7 ^= W6
					W6 |= W4
					tw ^= W6
					W4 ^= W7
					W7 &= W6
					W6 ^= W5
					W5 = tw

					W4 ^= W0
					ul = uint32(W0)
					uh = uint32((W0 >> 32))
					vl = uint32(W4)
					vh = uint32((W4 >> 32))
					ul = ((ul << 2) | (ul >> (32 - 2))) ^ vl
					vl = ((vl << 14) | (vl >> (32 - 14))) ^ ul
					ul = ((ul << 10) | (ul >> (32 - 10))) ^ vl
					vl = ((vl << 1) | (vl >> (32 - 1)))
					uh = ((uh << 2) | (uh >> (32 - 2))) ^ vh
					vh = ((vh << 14) | (vh >> (32 - 14))) ^ uh
					uh = ((uh << 10) | (uh >> (32 - 10))) ^ vh
					vh = ((vh << 1) | (vh >> (32 - 1)))
					W0 = uint64(ul) | (uint64(uh) << 32)
					W4 = uint64(vl) | (uint64(vh) << 32)

					W5 ^= W1
					ul = uint32(W1)
					uh = uint32((W1 >> 32))
					vl = uint32(W5)
					vh = uint32((W5 >> 32))
					ul = ((ul << 2) | (ul >> (32 - 2))) ^ vl
					vl = ((vl << 14) | (vl >> (32 - 14))) ^ ul
					ul = ((ul << 10) | (ul >> (32 - 10))) ^ vl
					vl = ((vl << 1) | (vl >> (32 - 1)))
					uh = ((uh << 2) | (uh >> (32 - 2))) ^ vh
					vh = ((vh << 14) | (vh >> (32 - 14))) ^ uh
					uh = ((uh << 10) | (uh >> (32 - 10))) ^ vh
					vh = ((vh << 1) | (vh >> (32 - 1)))
					W1 = uint64(ul) | (uint64(uh) << 32)
					W5 = uint64(vl) | (uint64(vh) << 32)

					W6 ^= W2
					ul = uint32(W2)
					uh = uint32((W2 >> 32))
					vl = uint32(W6)
					vh = uint32((W6 >> 32))
					ul = ((ul << 2) | (ul >> (32 - 2))) ^ vl
					vl = ((vl << 14) | (vl >> (32 - 14))) ^ ul
					ul = ((ul << 10) | (ul >> (32 - 10))) ^ vl
					vl = ((vl << 1) | (vl >> (32 - 1)))
					uh = ((uh << 2) | (uh >> (32 - 2))) ^ vh
					vh = ((vh << 14) | (vh >> (32 - 14))) ^ uh
					uh = ((uh << 10) | (uh >> (32 - 10))) ^ vh
					vh = ((vh << 1) | (vh >> (32 - 1)))
					W2 = uint64(ul) | (uint64(uh) << 32)
					W6 = uint64(vl) | (uint64(vh) << 32)

					W7 ^= W3
					ul = uint32(W3)
					uh = uint32((W3 >> 32))
					vl = uint32(W7)
					vh = uint32((W7 >> 32))
					ul = ((ul << 2) | (ul >> (32 - 2))) ^ vl
					vl = ((vl << 14) | (vl >> (32 - 14))) ^ ul
					ul = ((ul << 10) | (ul >> (32 - 10))) ^ vl
					vl = ((vl << 1) | (vl >> (32 - 1)))
					uh = ((uh << 2) | (uh >> (32 - 2))) ^ vh
					vh = ((vh << 14) | (vh >> (32 - 14))) ^ uh
					uh = ((uh << 10) | (uh >> (32 - 10))) ^ vh
					vh = ((vh << 1) | (vh >> (32 - 1)))
					W3 = uint64(ul) | (uint64(uh) << 32)
					W7 = uint64(vl) | (uint64(vh) << 32)

					W0 ^= kRCW230[r]
					W4 ^= kRCW234[r]
				}

				V20 = uint32(W0)
				V30 = uint32((W0 >> 32))
				V21 = uint32(W1)
				V31 = uint32((W1 >> 32))
				V22 = uint32(W2)
				V32 = uint32((W2 >> 32))
				V23 = uint32(W3)
				V33 = uint32((W3 >> 32))
				V24 = uint32(W4)
				V34 = uint32((W4 >> 32))
				V25 = uint32(W5)
				V35 = uint32((W5 >> 32))
				V26 = uint32(W6)
				V36 = uint32((W6 >> 32))
				V27 = uint32(W7)
				V37 = uint32((W7 >> 32))

				for r := uintptr(0); r < 8; r++ {
					tws = V40
					V40 |= V41
					V42 ^= V43
					V41 = ^V41
					V40 ^= V43
					V43 &= tws
					V41 ^= V43
					V43 ^= V42
					V42 &= V40
					V40 = ^V40
					V42 ^= V41
					V41 |= V43
					tws ^= V41
					V43 ^= V42
					V42 &= V41
					V41 ^= V40
					V40 = tws

					tws = V45
					V45 |= V46
					V47 ^= V44
					V46 = ^V46
					V45 ^= V44
					V44 &= tws
					V46 ^= V44
					V44 ^= V47
					V47 &= V45
					V45 = ^V45
					V47 ^= V46
					V46 |= V44
					tws ^= V46
					V44 ^= V47
					V47 &= V46
					V46 ^= V45
					V45 = tws

					V44 ^= V40
					V40 = ((V40 << 2) | (V40 >> (32 - 2))) ^ V44
					V44 = ((V44 << 14) | (V44 >> (32 - 14))) ^ V40
					V40 = ((V40 << 10) | (V40 >> (32 - 10))) ^ V44
					V44 = ((V44 << 1) | (V44 >> (32 - 1)))

					V45 ^= V41
					V41 = ((V41 << 2) | (V41 >> (32 - 2))) ^ V45
					V45 = ((V45 << 14) | (V45 >> (32 - 14))) ^ V41
					V41 = ((V41 << 10) | (V41 >> (32 - 10))) ^ V45
					V45 = ((V45 << 1) | (V45 >> (32 - 1)))

					V46 ^= V42
					V42 = ((V42 << 2) | (V42 >> (32 - 2))) ^ V46
					V46 = ((V46 << 14) | (V46 >> (32 - 14))) ^ V42
					V42 = ((V42 << 10) | (V42 >> (32 - 10))) ^ V46
					V46 = ((V46 << 1) | (V46 >> (32 - 1)))

					V47 ^= V43
					V43 = ((V43 << 2) | (V43 >> (32 - 2))) ^ V47
					V47 = ((V47 << 14) | (V47 >> (32 - 14))) ^ V43
					V43 = ((V43 << 10) | (V43 >> (32 - 10))) ^ V47
					V47 = ((V47 << 1) | (V47 >> (32 - 1)))

					V40 ^= kRC40[r]
					V44 ^= kRC44[r]
				}
			}

			ptr = 0
		}
	}

	ref.h[0][0] = V00
	ref.h[0][1] = V01
	ref.h[0][2] = V02
	ref.h[0][3] = V03
	ref.h[0][4] = V04
	ref.h[0][5] = V05
	ref.h[0][6] = V06
	ref.h[0][7] = V07
	ref.h[1][0] = V10
	ref.h[1][1] = V11
	ref.h[1][2] = V12
	ref.h[1][3] = V13
	ref.h[1][4] = V14
	ref.h[1][5] = V15
	ref.h[1][6] = V16
	ref.h[1][7] = V17
	ref.h[2][0] = V20
	ref.h[2][1] = V21
	ref.h[2][2] = V22
	ref.h[2][3] = V23
	ref.h[2][4] = V24
	ref.h[2][5] = V25
	ref.h[2][6] = V26
	ref.h[2][7] = V27
	ref.h[3][0] = V30
	ref.h[3][1] = V31
	ref.h[3][2] = V32
	ref.h[3][3] = V33
	ref.h[3][4] = V34
	ref.h[3][5] = V35
	ref.h[3][6] = V36
	ref.h[3][7] = V37
	ref.h[4][0] = V40
	ref.h[4][1] = V41
	ref.h[4][2] = V42
	ref.h[4][3] = V43
	ref.h[4][4] = V44
	ref.h[4][5] = V45
	ref.h[4][6] = V46
	ref.h[4][7] = V47

	ref.ptr = ptr
	return fln, nil
}

// Close the digest by writing the last bits and storing the hash
// in dst. This prepares the digest for reuse by calling reset. A call
// to Close with a dst that is smaller then HashSize will return an error.
func (ref *digest) Close(dst []byte, bits uint8, bcnt uint8) error {
	if ln := len(dst); HashSize > ln {
		return fmt.Errorf("Luffa Close: dst min length: %d, got %d", HashSize, ln)
	}

	buf := ref.b[:]
	ptr := ref.ptr + 1

	{
		off := uint8(0x80) >> bcnt
		buf[ref.ptr] = uint8((bits & -off) | off)
	}

	memset(buf[ptr:], 0)

	var V00, V01, V02, V03, V04, V05, V06, V07 uint32
	var V10, V11, V12, V13, V14, V15, V16, V17 uint32
	var V20, V21, V22, V23, V24, V25, V26, V27 uint32
	var V30, V31, V32, V33, V34, V35, V36, V37 uint32
	var V40, V41, V42, V43, V44, V45, V46, V47 uint32

	V00 = ref.h[0][0]
	V01 = ref.h[0][1]
	V02 = ref.h[0][2]
	V03 = ref.h[0][3]
	V04 = ref.h[0][4]
	V05 = ref.h[0][5]
	V06 = ref.h[0][6]
	V07 = ref.h[0][7]
	V10 = ref.h[1][0]
	V11 = ref.h[1][1]
	V12 = ref.h[1][2]
	V13 = ref.h[1][3]
	V14 = ref.h[1][4]
	V15 = ref.h[1][5]
	V16 = ref.h[1][6]
	V17 = ref.h[1][7]
	V20 = ref.h[2][0]
	V21 = ref.h[2][1]
	V22 = ref.h[2][2]
	V23 = ref.h[2][3]
	V24 = ref.h[2][4]
	V25 = ref.h[2][5]
	V26 = ref.h[2][6]
	V27 = ref.h[2][7]
	V30 = ref.h[3][0]
	V31 = ref.h[3][1]
	V32 = ref.h[3][2]
	V33 = ref.h[3][3]
	V34 = ref.h[3][4]
	V35 = ref.h[3][5]
	V36 = ref.h[3][6]
	V37 = ref.h[3][7]
	V40 = ref.h[4][0]
	V41 = ref.h[4][1]
	V42 = ref.h[4][2]
	V43 = ref.h[4][3]
	V44 = ref.h[4][4]
	V45 = ref.h[4][5]
	V46 = ref.h[4][6]
	V47 = ref.h[4][7]

	for i := uintptr(0); i < 3; i++ {
		{
			var ts uint32
			var M0, M1, M2, M3, M4, M5, M6, M7 uint32
			var a0, a1, a2, a3, a4, a5, a6, a7 uint32
			var b0, b1, b2, b3, b4, b5, b6, b7 uint32

			M0 = decUInt32be(buf[0:])
			M1 = decUInt32be(buf[4:])
			M2 = decUInt32be(buf[8:])
			M3 = decUInt32be(buf[12:])
			M4 = decUInt32be(buf[16:])
			M5 = decUInt32be(buf[20:])
			M6 = decUInt32be(buf[24:])
			M7 = decUInt32be(buf[28:])

			a0 = V00 ^ V10
			a1 = V01 ^ V11
			a2 = V02 ^ V12
			a3 = V03 ^ V13
			a4 = V04 ^ V14
			a5 = V05 ^ V15
			a6 = V06 ^ V16
			a7 = V07 ^ V17

			b0 = V20 ^ V30
			b1 = V21 ^ V31
			b2 = V22 ^ V32
			b3 = V23 ^ V33
			b4 = V24 ^ V34
			b5 = V25 ^ V35
			b6 = V26 ^ V36
			b7 = V27 ^ V37

			a0 ^= b0
			a1 ^= b1
			a2 ^= b2
			a3 ^= b3
			a4 ^= b4
			a5 ^= b5
			a6 ^= b6
			a7 ^= b7

			a0 ^= V40
			a1 ^= V41
			a2 ^= V42
			a3 ^= V43
			a4 ^= V44
			a5 ^= V45
			a6 ^= V46
			a7 ^= V47

			ts = a7
			a7 = a6
			a6 = a5
			a5 = a4
			a4 = a3 ^ ts
			a3 = a2 ^ ts
			a2 = a1
			a1 = a0 ^ ts
			a0 = ts

			V00 ^= a0
			V01 ^= a1
			V02 ^= a2
			V03 ^= a3
			V04 ^= a4
			V05 ^= a5
			V06 ^= a6
			V07 ^= a7

			V10 ^= a0
			V11 ^= a1
			V12 ^= a2
			V13 ^= a3
			V14 ^= a4
			V15 ^= a5
			V16 ^= a6
			V17 ^= a7

			V20 ^= a0
			V21 ^= a1
			V22 ^= a2
			V23 ^= a3
			V24 ^= a4
			V25 ^= a5
			V26 ^= a6
			V27 ^= a7

			V30 ^= a0
			V31 ^= a1
			V32 ^= a2
			V33 ^= a3
			V34 ^= a4
			V35 ^= a5
			V36 ^= a6
			V37 ^= a7

			V40 ^= a0
			V41 ^= a1
			V42 ^= a2
			V43 ^= a3
			V44 ^= a4
			V45 ^= a5
			V46 ^= a6
			V47 ^= a7

			ts = V07
			b7 = V06
			b6 = V05
			b5 = V04
			b4 = V03 ^ ts
			b3 = V02 ^ ts
			b2 = V01
			b1 = V00 ^ ts
			b0 = ts

			b0 ^= V10
			b1 ^= V11
			b2 ^= V12
			b3 ^= V13
			b4 ^= V14
			b5 ^= V15
			b6 ^= V16
			b7 ^= V17

			ts = V17
			V17 = V16
			V16 = V15
			V15 = V14
			V14 = V13 ^ ts
			V13 = V12 ^ ts
			V12 = V11
			V11 = V10 ^ ts
			V10 = ts

			V10 ^= V20
			V11 ^= V21
			V12 ^= V22
			V13 ^= V23
			V14 ^= V24
			V15 ^= V25
			V16 ^= V26
			V17 ^= V27

			ts = V27
			V27 = V26
			V26 = V25
			V25 = V24
			V24 = V23 ^ ts
			V23 = V22 ^ ts
			V22 = V21
			V21 = V20 ^ ts
			V20 = ts

			V20 ^= V30
			V21 ^= V31
			V22 ^= V32
			V23 ^= V33
			V24 ^= V34
			V25 ^= V35
			V26 ^= V36
			V27 ^= V37

			ts = V37
			V37 = V36
			V36 = V35
			V35 = V34
			V34 = V33 ^ ts
			V33 = V32 ^ ts
			V32 = V31
			V31 = V30 ^ ts
			V30 = ts

			V30 ^= V40
			V31 ^= V41
			V32 ^= V42
			V33 ^= V43
			V34 ^= V44
			V35 ^= V45
			V36 ^= V46
			V37 ^= V47

			ts = V47
			V47 = V46
			V46 = V45
			V45 = V44
			V44 = V43 ^ ts
			V43 = V42 ^ ts
			V42 = V41
			V41 = V40 ^ ts
			V40 = ts

			V40 ^= V00
			V41 ^= V01
			V42 ^= V02
			V43 ^= V03
			V44 ^= V04
			V45 ^= V05
			V46 ^= V06
			V47 ^= V07

			ts = b7
			V07 = b6
			V06 = b5
			V05 = b4
			V04 = b3 ^ ts
			V03 = b2 ^ ts
			V02 = b1
			V01 = b0 ^ ts
			V00 = ts

			V00 ^= V40
			V01 ^= V41
			V02 ^= V42
			V03 ^= V43
			V04 ^= V44
			V05 ^= V45
			V06 ^= V46
			V07 ^= V47

			ts = V47
			V47 = V46
			V46 = V45
			V45 = V44
			V44 = V43 ^ ts
			V43 = V42 ^ ts
			V42 = V41
			V41 = V40 ^ ts
			V40 = ts

			V40 ^= V30
			V41 ^= V31
			V42 ^= V32
			V43 ^= V33
			V44 ^= V34
			V45 ^= V35
			V46 ^= V36
			V47 ^= V37

			ts = V37
			V37 = V36
			V36 = V35
			V35 = V34
			V34 = V33 ^ ts
			V33 = V32 ^ ts
			V32 = V31
			V31 = V30 ^ ts
			V30 = ts

			V30 ^= V20
			V31 ^= V21
			V32 ^= V22
			V33 ^= V23
			V34 ^= V24
			V35 ^= V25
			V36 ^= V26
			V37 ^= V27

			ts = V27
			V27 = V26
			V26 = V25
			V25 = V24
			V24 = V23 ^ ts
			V23 = V22 ^ ts
			V22 = V21
			V21 = V20 ^ ts
			V20 = ts

			V20 ^= V10
			V21 ^= V11
			V22 ^= V12
			V23 ^= V13
			V24 ^= V14
			V25 ^= V15
			V26 ^= V16
			V27 ^= V17

			ts = V17
			V17 = V16
			V16 = V15
			V15 = V14
			V14 = V13 ^ ts
			V13 = V12 ^ ts
			V12 = V11
			V11 = V10 ^ ts
			V10 = ts

			V10 ^= b0
			V11 ^= b1
			V12 ^= b2
			V13 ^= b3
			V14 ^= b4
			V15 ^= b5
			V16 ^= b6
			V17 ^= b7

			V00 ^= M0
			V01 ^= M1
			V02 ^= M2
			V03 ^= M3
			V04 ^= M4
			V05 ^= M5
			V06 ^= M6
			V07 ^= M7

			ts = M7
			M7 = M6
			M6 = M5
			M5 = M4
			M4 = M3 ^ ts
			M3 = M2 ^ ts
			M2 = M1
			M1 = M0 ^ ts
			M0 = ts

			V10 ^= M0
			V11 ^= M1
			V12 ^= M2
			V13 ^= M3
			V14 ^= M4
			V15 ^= M5
			V16 ^= M6
			V17 ^= M7

			ts = M7
			M7 = M6
			M6 = M5
			M5 = M4
			M4 = M3 ^ ts
			M3 = M2 ^ ts
			M2 = M1
			M1 = M0 ^ ts
			M0 = ts

			V20 ^= M0
			V21 ^= M1
			V22 ^= M2
			V23 ^= M3
			V24 ^= M4
			V25 ^= M5
			V26 ^= M6
			V27 ^= M7

			ts = M7
			M7 = M6
			M6 = M5
			M5 = M4
			M4 = M3 ^ ts
			M3 = M2 ^ ts
			M2 = M1
			M1 = M0 ^ ts
			M0 = ts

			V30 ^= M0
			V31 ^= M1
			V32 ^= M2
			V33 ^= M3
			V34 ^= M4
			V35 ^= M5
			V36 ^= M6
			V37 ^= M7

			ts = M7
			M7 = M6
			M6 = M5
			M5 = M4
			M4 = M3 ^ ts
			M3 = M2 ^ ts
			M2 = M1
			M1 = M0 ^ ts
			M0 = ts

			V40 ^= M0
			V41 ^= M1
			V42 ^= M2
			V43 ^= M3
			V44 ^= M4
			V45 ^= M5
			V46 ^= M6
			V47 ^= M7
		}

		{
			var ul, uh, vl, vh, tws uint32
			var W0, W1, W2, W3, W4, W5, W6, W7, tw uint64

			V14 = ((V14 << 1) | (V14 >> (32 - 1)))
			V15 = ((V15 << 1) | (V15 >> (32 - 1)))
			V16 = ((V16 << 1) | (V16 >> (32 - 1)))
			V17 = ((V17 << 1) | (V17 >> (32 - 1)))
			V24 = ((V24 << 2) | (V24 >> (32 - 2)))
			V25 = ((V25 << 2) | (V25 >> (32 - 2)))
			V26 = ((V26 << 2) | (V26 >> (32 - 2)))
			V27 = ((V27 << 2) | (V27 >> (32 - 2)))
			V34 = ((V34 << 3) | (V34 >> (32 - 3)))
			V35 = ((V35 << 3) | (V35 >> (32 - 3)))
			V36 = ((V36 << 3) | (V36 >> (32 - 3)))
			V37 = ((V37 << 3) | (V37 >> (32 - 3)))
			V44 = ((V44 << 4) | (V44 >> (32 - 4)))
			V45 = ((V45 << 4) | (V45 >> (32 - 4)))
			V46 = ((V46 << 4) | (V46 >> (32 - 4)))
			V47 = ((V47 << 4) | (V47 >> (32 - 4)))

			W0 = uint64(V00) | (uint64(V10) << 32)
			W1 = uint64(V01) | (uint64(V11) << 32)
			W2 = uint64(V02) | (uint64(V12) << 32)
			W3 = uint64(V03) | (uint64(V13) << 32)
			W4 = uint64(V04) | (uint64(V14) << 32)
			W5 = uint64(V05) | (uint64(V15) << 32)
			W6 = uint64(V06) | (uint64(V16) << 32)
			W7 = uint64(V07) | (uint64(V17) << 32)

			for r := uintptr(0); r < 8; r++ {
				tw = W0
				W0 |= W1
				W2 ^= W3
				W1 = ^W1
				W0 ^= W3
				W3 &= tw
				W1 ^= W3
				W3 ^= W2
				W2 &= W0
				W0 = ^W0
				W2 ^= W1
				W1 |= W3
				tw ^= W1
				W3 ^= W2
				W2 &= W1
				W1 ^= W0
				W0 = tw

				tw = W5
				W5 |= W6
				W7 ^= W4
				W6 = ^W6
				W5 ^= W4
				W4 &= tw
				W6 ^= W4
				W4 ^= W7
				W7 &= W5
				W5 = ^W5
				W7 ^= W6
				W6 |= W4
				tw ^= W6
				W4 ^= W7
				W7 &= W6
				W6 ^= W5
				W5 = tw

				W4 ^= W0
				ul = uint32(W0)
				uh = uint32((W0 >> 32))
				vl = uint32(W4)
				vh = uint32((W4 >> 32))
				ul = ((ul << 2) | (ul >> (32 - 2))) ^ vl
				vl = ((vl << 14) | (vl >> (32 - 14))) ^ ul
				ul = ((ul << 10) | (ul >> (32 - 10))) ^ vl
				vl = ((vl << 1) | (vl >> (32 - 1)))
				uh = ((uh << 2) | (uh >> (32 - 2))) ^ vh
				vh = ((vh << 14) | (vh >> (32 - 14))) ^ uh
				uh = ((uh << 10) | (uh >> (32 - 10))) ^ vh
				vh = ((vh << 1) | (vh >> (32 - 1)))
				W0 = uint64(ul) | (uint64(uh) << 32)
				W4 = uint64(vl) | (uint64(vh) << 32)

				W5 ^= W1
				ul = uint32(W1)
				uh = uint32((W1 >> 32))
				vl = uint32(W5)
				vh = uint32((W5 >> 32))
				ul = ((ul << 2) | (ul >> (32 - 2))) ^ vl
				vl = ((vl << 14) | (vl >> (32 - 14))) ^ ul
				ul = ((ul << 10) | (ul >> (32 - 10))) ^ vl
				vl = ((vl << 1) | (vl >> (32 - 1)))
				uh = ((uh << 2) | (uh >> (32 - 2))) ^ vh
				vh = ((vh << 14) | (vh >> (32 - 14))) ^ uh
				uh = ((uh << 10) | (uh >> (32 - 10))) ^ vh
				vh = ((vh << 1) | (vh >> (32 - 1)))
				W1 = uint64(ul) | (uint64(uh) << 32)
				W5 = uint64(vl) | (uint64(vh) << 32)

				W6 ^= W2
				ul = uint32(W2)
				uh = uint32((W2 >> 32))
				vl = uint32(W6)
				vh = uint32((W6 >> 32))
				ul = ((ul << 2) | (ul >> (32 - 2))) ^ vl
				vl = ((vl << 14) | (vl >> (32 - 14))) ^ ul
				ul = ((ul << 10) | (ul >> (32 - 10))) ^ vl
				vl = ((vl << 1) | (vl >> (32 - 1)))
				uh = ((uh << 2) | (uh >> (32 - 2))) ^ vh
				vh = ((vh << 14) | (vh >> (32 - 14))) ^ uh
				uh = ((uh << 10) | (uh >> (32 - 10))) ^ vh
				vh = ((vh << 1) | (vh >> (32 - 1)))
				W2 = uint64(ul) | (uint64(uh) << 32)
				W6 = uint64(vl) | (uint64(vh) << 32)

				W7 ^= W3
				ul = uint32(W3)
				uh = uint32((W3 >> 32))
				vl = uint32(W7)
				vh = uint32((W7 >> 32))
				ul = ((ul << 2) | (ul >> (32 - 2))) ^ vl
				vl = ((vl << 14) | (vl >> (32 - 14))) ^ ul
				ul = ((ul << 10) | (ul >> (32 - 10))) ^ vl
				vl = ((vl << 1) | (vl >> (32 - 1)))
				uh = ((uh << 2) | (uh >> (32 - 2))) ^ vh
				vh = ((vh << 14) | (vh >> (32 - 14))) ^ uh
				uh = ((uh << 10) | (uh >> (32 - 10))) ^ vh
				vh = ((vh << 1) | (vh >> (32 - 1)))
				W3 = uint64(ul) | (uint64(uh) << 32)
				W7 = uint64(vl) | (uint64(vh) << 32)

				W0 ^= kRCW010[r]
				W4 ^= kRCW014[r]
			}

			V00 = uint32(W0)
			V10 = uint32((W0 >> 32))
			V01 = uint32(W1)
			V11 = uint32((W1 >> 32))
			V02 = uint32(W2)
			V12 = uint32((W2 >> 32))
			V03 = uint32(W3)
			V13 = uint32((W3 >> 32))
			V04 = uint32(W4)
			V14 = uint32((W4 >> 32))
			V05 = uint32(W5)
			V15 = uint32((W5 >> 32))
			V06 = uint32(W6)
			V16 = uint32((W6 >> 32))
			V07 = uint32(W7)
			V17 = uint32((W7 >> 32))

			W0 = uint64(V20) | (uint64(V30) << 32)
			W1 = uint64(V21) | (uint64(V31) << 32)
			W2 = uint64(V22) | (uint64(V32) << 32)
			W3 = uint64(V23) | (uint64(V33) << 32)
			W4 = uint64(V24) | (uint64(V34) << 32)
			W5 = uint64(V25) | (uint64(V35) << 32)
			W6 = uint64(V26) | (uint64(V36) << 32)
			W7 = uint64(V27) | (uint64(V37) << 32)

			for r := uintptr(0); r < 8; r++ {
				tw = W0
				W0 |= W1
				W2 ^= W3
				W1 = ^W1
				W0 ^= W3
				W3 &= tw
				W1 ^= W3
				W3 ^= W2
				W2 &= W0
				W0 = ^W0
				W2 ^= W1
				W1 |= W3
				tw ^= W1
				W3 ^= W2
				W2 &= W1
				W1 ^= W0
				W0 = tw

				tw = W5
				W5 |= W6
				W7 ^= W4
				W6 = ^W6
				W5 ^= W4
				W4 &= tw
				W6 ^= W4
				W4 ^= W7
				W7 &= W5
				W5 = ^W5
				W7 ^= W6
				W6 |= W4
				tw ^= W6
				W4 ^= W7
				W7 &= W6
				W6 ^= W5
				W5 = tw

				W4 ^= W0
				ul = uint32(W0)
				uh = uint32((W0 >> 32))
				vl = uint32(W4)
				vh = uint32((W4 >> 32))
				ul = ((ul << 2) | (ul >> (32 - 2))) ^ vl
				vl = ((vl << 14) | (vl >> (32 - 14))) ^ ul
				ul = ((ul << 10) | (ul >> (32 - 10))) ^ vl
				vl = ((vl << 1) | (vl >> (32 - 1)))
				uh = ((uh << 2) | (uh >> (32 - 2))) ^ vh
				vh = ((vh << 14) | (vh >> (32 - 14))) ^ uh
				uh = ((uh << 10) | (uh >> (32 - 10))) ^ vh
				vh = ((vh << 1) | (vh >> (32 - 1)))
				W0 = uint64(ul) | (uint64(uh) << 32)
				W4 = uint64(vl) | (uint64(vh) << 32)

				W5 ^= W1
				ul = uint32(W1)
				uh = uint32((W1 >> 32))
				vl = uint32(W5)
				vh = uint32((W5 >> 32))
				ul = ((ul << 2) | (ul >> (32 - 2))) ^ vl
				vl = ((vl << 14) | (vl >> (32 - 14))) ^ ul
				ul = ((ul << 10) | (ul >> (32 - 10))) ^ vl
				vl = ((vl << 1) | (vl >> (32 - 1)))
				uh = ((uh << 2) | (uh >> (32 - 2))) ^ vh
				vh = ((vh << 14) | (vh >> (32 - 14))) ^ uh
				uh = ((uh << 10) | (uh >> (32 - 10))) ^ vh
				vh = ((vh << 1) | (vh >> (32 - 1)))
				W1 = uint64(ul) | (uint64(uh) << 32)
				W5 = uint64(vl) | (uint64(vh) << 32)

				W6 ^= W2
				ul = uint32(W2)
				uh = uint32((W2 >> 32))
				vl = uint32(W6)
				vh = uint32((W6 >> 32))
				ul = ((ul << 2) | (ul >> (32 - 2))) ^ vl
				vl = ((vl << 14) | (vl >> (32 - 14))) ^ ul
				ul = ((ul << 10) | (ul >> (32 - 10))) ^ vl
				vl = ((vl << 1) | (vl >> (32 - 1)))
				uh = ((uh << 2) | (uh >> (32 - 2))) ^ vh
				vh = ((vh << 14) | (vh >> (32 - 14))) ^ uh
				uh = ((uh << 10) | (uh >> (32 - 10))) ^ vh
				vh = ((vh << 1) | (vh >> (32 - 1)))
				W2 = uint64(ul) | (uint64(uh) << 32)
				W6 = uint64(vl) | (uint64(vh) << 32)

				W7 ^= W3
				ul = uint32(W3)
				uh = uint32((W3 >> 32))
				vl = uint32(W7)
				vh = uint32((W7 >> 32))
				ul = ((ul << 2) | (ul >> (32 - 2))) ^ vl
				vl = ((vl << 14) | (vl >> (32 - 14))) ^ ul
				ul = ((ul << 10) | (ul >> (32 - 10))) ^ vl
				vl = ((vl << 1) | (vl >> (32 - 1)))
				uh = ((uh << 2) | (uh >> (32 - 2))) ^ vh
				vh = ((vh << 14) | (vh >> (32 - 14))) ^ uh
				uh = ((uh << 10) | (uh >> (32 - 10))) ^ vh
				vh = ((vh << 1) | (vh >> (32 - 1)))
				W3 = uint64(ul) | (uint64(uh) << 32)
				W7 = uint64(vl) | (uint64(vh) << 32)

				W0 ^= kRCW230[r]
				W4 ^= kRCW234[r]
			}

			V20 = uint32(W0)
			V30 = uint32((W0 >> 32))
			V21 = uint32(W1)
			V31 = uint32((W1 >> 32))
			V22 = uint32(W2)
			V32 = uint32((W2 >> 32))
			V23 = uint32(W3)
			V33 = uint32((W3 >> 32))
			V24 = uint32(W4)
			V34 = uint32((W4 >> 32))
			V25 = uint32(W5)
			V35 = uint32((W5 >> 32))
			V26 = uint32(W6)
			V36 = uint32((W6 >> 32))
			V27 = uint32(W7)
			V37 = uint32((W7 >> 32))

			for r := uintptr(0); r < 8; r++ {
				tws = V40
				V40 |= V41
				V42 ^= V43
				V41 = ^V41
				V40 ^= V43
				V43 &= tws
				V41 ^= V43
				V43 ^= V42
				V42 &= V40
				V40 = ^V40
				V42 ^= V41
				V41 |= V43
				tws ^= V41
				V43 ^= V42
				V42 &= V41
				V41 ^= V40
				V40 = tws

				tws = V45
				V45 |= V46
				V47 ^= V44
				V46 = ^V46
				V45 ^= V44
				V44 &= tws
				V46 ^= V44
				V44 ^= V47
				V47 &= V45
				V45 = ^V45
				V47 ^= V46
				V46 |= V44
				tws ^= V46
				V44 ^= V47
				V47 &= V46
				V46 ^= V45
				V45 = tws

				V44 ^= V40
				V40 = ((V40 << 2) | (V40 >> (32 - 2))) ^ V44
				V44 = ((V44 << 14) | (V44 >> (32 - 14))) ^ V40
				V40 = ((V40 << 10) | (V40 >> (32 - 10))) ^ V44
				V44 = ((V44 << 1) | (V44 >> (32 - 1)))

				V45 ^= V41
				V41 = ((V41 << 2) | (V41 >> (32 - 2))) ^ V45
				V45 = ((V45 << 14) | (V45 >> (32 - 14))) ^ V41
				V41 = ((V41 << 10) | (V41 >> (32 - 10))) ^ V45
				V45 = ((V45 << 1) | (V45 >> (32 - 1)))

				V46 ^= V42
				V42 = ((V42 << 2) | (V42 >> (32 - 2))) ^ V46
				V46 = ((V46 << 14) | (V46 >> (32 - 14))) ^ V42
				V42 = ((V42 << 10) | (V42 >> (32 - 10))) ^ V46
				V46 = ((V46 << 1) | (V46 >> (32 - 1)))

				V47 ^= V43
				V43 = ((V43 << 2) | (V43 >> (32 - 2))) ^ V47
				V47 = ((V47 << 14) | (V47 >> (32 - 14))) ^ V43
				V43 = ((V43 << 10) | (V43 >> (32 - 10))) ^ V47
				V47 = ((V47 << 1) | (V47 >> (32 - 1)))

				V40 ^= kRC40[r]
				V44 ^= kRC44[r]
			}
		}

		switch i {
		case 0:
			memset(buf[:], 0)
			break
		case 1:
			encUInt32be(dst[0:], V00^V10^V20^V30^V40)
			encUInt32be(dst[4:], V01^V11^V21^V31^V41)
			encUInt32be(dst[8:], V02^V12^V22^V32^V42)
			encUInt32be(dst[12:], V03^V13^V23^V33^V43)
			encUInt32be(dst[16:], V04^V14^V24^V34^V44)
			encUInt32be(dst[20:], V05^V15^V25^V35^V45)
			encUInt32be(dst[24:], V06^V16^V26^V36^V46)
			encUInt32be(dst[28:], V07^V17^V27^V37^V47)
			break
		case 2:
			encUInt32be(dst[32:], V00^V10^V20^V30^V40)
			encUInt32be(dst[36:], V01^V11^V21^V31^V41)
			encUInt32be(dst[40:], V02^V12^V22^V32^V42)
			encUInt32be(dst[44:], V03^V13^V23^V33^V43)
			encUInt32be(dst[48:], V04^V14^V24^V34^V44)
			encUInt32be(dst[52:], V05^V15^V25^V35^V45)
			encUInt32be(dst[56:], V06^V16^V26^V36^V46)
			encUInt32be(dst[60:], V07^V17^V27^V37^V47)
			break
		}
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

func decUInt32be(src []byte) uint32 {
	return (uint32(src[0])<<24 |
		uint32(src[1])<<16 |
		uint32(src[2])<<8 |
		uint32(src[3]))
}

func encUInt32be(dst []byte, src uint32) {
	dst[0] = uint8(src >> 24)
	dst[1] = uint8(src >> 16)
	dst[2] = uint8(src >> 8)
	dst[3] = uint8(src)
}

////////////////

var kInit = [5][8]uint32{
	{
		uint32(0x6d251e69), uint32(0x44b051e0),
		uint32(0x4eaa6fb4), uint32(0xdbf78465),
		uint32(0x6e292011), uint32(0x90152df4),
		uint32(0xee058139), uint32(0xdef610bb),
	},
	{
		uint32(0xc3b44b95), uint32(0xd9d2f256),
		uint32(0x70eee9a0), uint32(0xde099fa3),
		uint32(0x5d9b0557), uint32(0x8fc944b3),
		uint32(0xcf1ccf0e), uint32(0x746cd581),
	},
	{
		uint32(0xf7efc89d), uint32(0x5dba5781),
		uint32(0x04016ce5), uint32(0xad659c05),
		uint32(0x0306194f), uint32(0x666d1836),
		uint32(0x24aa230a), uint32(0x8b264ae7),
	},
	{
		uint32(0x858075d5), uint32(0x36d79cce),
		uint32(0xe571f7d7), uint32(0x204b1f67),
		uint32(0x35870c6a), uint32(0x57e9e923),
		uint32(0x14bcb808), uint32(0x7cde72ce),
	},
	{
		uint32(0x6c68e9be), uint32(0x5ec41e22),
		uint32(0xc825b7c7), uint32(0xaffb4363),
		uint32(0xf5df3999), uint32(0x0fc688f1),
		uint32(0xb07224cc), uint32(0x03e86cea),
	},
}

var kRC40 = [8]uint32{
	uint32(0xf0d2e9e3), uint32(0xac11d7fa),
	uint32(0x1bcb66f2), uint32(0x6f2d9bc9),
	uint32(0x78602649), uint32(0x8edae952),
	uint32(0x3b6ba548), uint32(0xedae9520),
}
var kRC44 = [8]uint32{
	uint32(0x5090d577), uint32(0x2d1925ab),
	uint32(0xb46496ac), uint32(0xd1925ab0),
	uint32(0x29131ab6), uint32(0x0fc053c3),
	uint32(0x3f014f0c), uint32(0xfc053c31),
}

var kRCW010 = [8]uint64{
	uint64(0xb6de10ed303994a6), uint64(0x70f47aaec0e65299),
	uint64(0x0707a3d46cc33a12), uint64(0x1c1e8f51dc56983e),
	uint64(0x707a3d451e00108f), uint64(0xaeb285627800423d),
	uint64(0xbaca15898f5b7882), uint64(0x40a46f3e96e1db12),
}
var kRCW014 = [8]uint64{
	uint64(0x01685f3de0337818), uint64(0x05a17cf4441ba90d),
	uint64(0xbd09caca7f34d442), uint64(0xf4272b289389217f),
	uint64(0x144ae5cce5a8bce6), uint64(0xfaa7ae2b5274baf4),
	uint64(0x2e48f1c126889ba7), uint64(0xb923c7049a226e9d),
}
var kRCW230 = [8]uint64{
	uint64(0xb213afa5fc20d9d2), uint64(0xc84ebe9534552e25),
	uint64(0x4e608a227ad8818f), uint64(0x56d858fe8438764a),
	uint64(0x343b138fbb6de032), uint64(0xd0ec4e3dedb780c8),
	uint64(0x2ceb4882d9847356), uint64(0xb3ad2208a2c78434),
}
var kRCW234 = [8]uint64{
	uint64(0xe028c9bfe25e72c1), uint64(0x44756f91e623bb72),
	uint64(0x7e8fce325c58a4a4), uint64(0x956548be1e38e2e7),
	uint64(0xfe191be278e38b9d), uint64(0x3cb226e527586719),
	uint64(0x5944a28e36eda57f), uint64(0xa1c4c355703aace7),
}

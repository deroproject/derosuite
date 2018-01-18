// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package x11

import (
	"gitlab.com/nitya-sattva/go-x11/blake"
	"gitlab.com/nitya-sattva/go-x11/bmw"
	"gitlab.com/nitya-sattva/go-x11/cubed"
	"gitlab.com/nitya-sattva/go-x11/echo"
	"gitlab.com/nitya-sattva/go-x11/groest"
	"gitlab.com/nitya-sattva/go-x11/hash"
	"gitlab.com/nitya-sattva/go-x11/jhash"
	"gitlab.com/nitya-sattva/go-x11/keccak"
	"gitlab.com/nitya-sattva/go-x11/luffa"
	"gitlab.com/nitya-sattva/go-x11/shavite"
	"gitlab.com/nitya-sattva/go-x11/simd"
	"gitlab.com/nitya-sattva/go-x11/skein"
)

////////////////

// Hash contains the state objects
// required to perform the x11.Hash.
type Hash struct {
	tha [64]byte
	thb [64]byte

	blake   hash.Digest
	bmw     hash.Digest
	cubed   hash.Digest
	echo    hash.Digest
	groest  hash.Digest
	jhash   hash.Digest
	keccak  hash.Digest
	luffa   hash.Digest
	shavite hash.Digest
	simd    hash.Digest
	skein   hash.Digest
}

// New returns a new object to compute a x11 hash.
func New() *Hash {
	ref := &Hash{}
	ref.blake = blake.New()
	ref.bmw = bmw.New()
	ref.cubed = cubed.New()
	ref.echo = echo.New()
	ref.groest = groest.New()
	ref.jhash = jhash.New()
	ref.keccak = keccak.New()
	ref.luffa = luffa.New()
	ref.shavite = shavite.New()
	ref.simd = simd.New()
	ref.skein = skein.New()
	return ref
}

// Hash computes the hash from the src bytes and stores the result in dst.
func (ref *Hash) Hash(src []byte, dst []byte) {
	ta := ref.tha[:]
	tb := ref.thb[:]

	ref.blake.Write(src)
	ref.blake.Close(tb, 0, 0)

	ref.bmw.Write(tb)
	ref.bmw.Close(ta, 0, 0)

	ref.groest.Write(ta)
	ref.groest.Close(tb, 0, 0)

	ref.skein.Write(tb)
	ref.skein.Close(ta, 0, 0)

	ref.jhash.Write(ta)
	ref.jhash.Close(tb, 0, 0)

	ref.keccak.Write(tb)
	ref.keccak.Close(ta, 0, 0)

	ref.luffa.Write(ta)
	ref.luffa.Close(tb, 0, 0)

	ref.cubed.Write(tb)
	ref.cubed.Close(ta, 0, 0)

	ref.shavite.Write(ta)
	ref.shavite.Close(tb, 0, 0)

	ref.simd.Write(tb)
	ref.simd.Close(ta, 0, 0)

	ref.echo.Write(ta)
	ref.echo.Close(tb, 0, 0)

	copy(dst, tb)
}

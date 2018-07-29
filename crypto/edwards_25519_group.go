// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crypto

// Group elements are members of the elliptic curve -x^2 + y^2 = 1 + d * x^2 *
// y^2 where d = -121665/121666.
//
// Several representations are used:
//   ProjectiveGroupElement: (X:Y:Z) satisfying x=X/Z, y=Y/Z
//   ExtendedGroupElement: (X:Y:Z:T) satisfying x=X/Z, y=Y/Z, XY=ZT
//   CompletedGroupElement: ((X:Z),(Y:T)) satisfying x=X/Z, y=Y/T
//   PreComputedGroupElement: (y+x,y-x,2dxy)

type ProjectiveGroupElement struct {
	X, Y, Z FieldElement
}

type ExtendedGroupElement struct {
	X, Y, Z, T FieldElement
}

type CompletedGroupElement struct {
	X, Y, Z, T FieldElement
}

type PreComputedGroupElement struct {
	yPlusX, yMinusX, xy2d FieldElement
}

type CachedGroupElement struct {
	yPlusX, yMinusX, Z, T2d FieldElement
}

func (c *CachedGroupElement) Zero() {
	FeOne(&c.yPlusX)  //c.yPlusX.One()
	FeOne(&c.yMinusX) //c.yMinusX.One()
	FeOne(&c.Z)       //c.Z.One()
	FeZero(&c.T2d)    //c.T2d.Zero()
}

func (p *ProjectiveGroupElement) Zero() {
	FeZero(&p.X)
	FeOne(&p.Y)
	FeOne(&p.Z)
}

func (p *ProjectiveGroupElement) Double(r *CompletedGroupElement) {
	var t0 FieldElement

	FeSquare(&r.X, &p.X)
	FeSquare(&r.Z, &p.Y)
	FeSquare2(&r.T, &p.Z)
	FeAdd(&r.Y, &p.X, &p.Y)
	FeSquare(&t0, &r.Y)
	FeAdd(&r.Y, &r.Z, &r.X)
	FeSub(&r.Z, &r.Z, &r.X)
	FeSub(&r.X, &t0, &r.Y)
	FeSub(&r.T, &r.T, &r.Z)
}

func (p *ProjectiveGroupElement) ToBytes(s *Key) {
	var recip, x, y FieldElement

	FeInvert(&recip, &p.Z)
	FeMul(&x, &p.X, &recip)
	FeMul(&y, &p.Y, &recip)
	FeToBytes(s, &y)
	s[31] ^= FeIsNegative(&x) << 7
}

func (p *ExtendedGroupElement) Zero() {
	FeZero(&p.X)
	FeOne(&p.Y)
	FeOne(&p.Z)
	FeZero(&p.T)
}

func (p *ExtendedGroupElement) Double(r *CompletedGroupElement) {
	var q ProjectiveGroupElement
	p.ToProjective(&q)
	q.Double(r)
}

func (p *ExtendedGroupElement) ToCached(r *CachedGroupElement) {
	FeAdd(&r.yPlusX, &p.Y, &p.X)
	FeSub(&r.yMinusX, &p.Y, &p.X)
	FeCopy(&r.Z, &p.Z)
	FeMul(&r.T2d, &p.T, &d2)
}

// build precomputed element for
func (p *ExtendedGroupElement) ToPreComputed(r *PreComputedGroupElement) {
	FeAdd(&r.yPlusX, &p.Y, &p.X)
	FeSub(&r.yMinusX, &p.Y, &p.X)
	FeMul(&r.xy2d, &p.X, &p.Y)
	FeMul(&r.xy2d, &r.xy2d, &d2)
}

func (p *ExtendedGroupElement) ToProjective(r *ProjectiveGroupElement) {
	FeCopy(&r.X, &p.X)
	FeCopy(&r.Y, &p.Y)
	FeCopy(&r.Z, &p.Z)
}

func (p *ExtendedGroupElement) ToBytes(s *Key) {
	var recip, x, y FieldElement

	FeInvert(&recip, &p.Z)
	FeMul(&x, &p.X, &recip)
	FeMul(&y, &p.Y, &recip)
	FeToBytes(s, &y)
	s[31] ^= FeIsNegative(&x) << 7
}

func (p *ExtendedGroupElement) FromBytes(s *Key) bool {
	var u, v, v3, vxx, check FieldElement

	// expanded FeFromBytes (with canonical check)
	h0 := load4(s[:])
	h1 := load3(s[4:]) << 6
	h2 := load3(s[7:]) << 5
	h3 := load3(s[10:]) << 3
	h4 := load3(s[13:]) << 2
	h5 := load4(s[16:])
	h6 := load3(s[20:]) << 7
	h7 := load3(s[23:]) << 5
	h8 := load3(s[26:]) << 4
	h9 := (load3(s[29:]) & 8388607) << 2

	// Validate the number to be canonical
	if h9 == 33554428 && h8 == 268435440 && h7 == 536870880 && h6 == 2147483520 &&
		h5 == 4294967295 && h4 == 67108860 && h3 == 134217720 && h2 == 536870880 &&
		h1 == 1073741760 && h0 >= 4294967277 {
		return false
	}

	FeFromBytes(&p.Y, s)
	FeOne(&p.Z)
	FeSquare(&u, &p.Y)
	FeMul(&v, &u, &d)
	FeSub(&u, &u, &p.Z) // y = y^2-1
	FeAdd(&v, &v, &p.Z) // v = dy^2+1

	FeSquare(&v3, &v)
	FeMul(&v3, &v3, &v) // v3 = v^3
	FeSquare(&p.X, &v3)
	FeMul(&p.X, &p.X, &v)
	FeMul(&p.X, &p.X, &u) // x = uv^7

	fePow22523(&p.X, &p.X) // x = (uv^7)^((q-5)/8)
	FeMul(&p.X, &p.X, &v3)
	FeMul(&p.X, &p.X, &u) // x = uv^3(uv^7)^((q-5)/8)

	var tmpX, tmp2 Key

	FeSquare(&vxx, &p.X)
	FeMul(&vxx, &vxx, &v)
	FeSub(&check, &vxx, &u) // vx^2-u
	if FeIsNonZero(&check) == 1 {
		FeAdd(&check, &vxx, &u) // vx^2+u
		if FeIsNonZero(&check) == 1 {
			return false
		}
		FeMul(&p.X, &p.X, &SqrtM1)

		FeToBytes(&tmpX, &p.X)
		for i, v := range tmpX {
			tmp2[31-i] = v
		}
	}

	if FeIsNegative(&p.X) != (s[31] >> 7) {
		// If x = 0, the sign must be positive
		if FeIsNonZero(&p.X) == 0 {
			return false
		}
		FeNeg(&p.X, &p.X)
	}

	FeMul(&p.T, &p.X, &p.Y)
	return true
}

func FeDivPowM1(out, u, v *FieldElement) {
	var v3, uv7, t0 FieldElement

	FeSquare(&v3, v)
	FeMul(&v3, &v3, v) /* v3 = v^3 */
	FeSquare(&uv7, &v3)
	FeMul(&uv7, &uv7, v)
	FeMul(&uv7, &uv7, u) /* uv7 = uv^7 */

	fePow22523(&t0, &uv7)
	/* t0 = (uv^7)^((q-5)/8) */
	FeMul(&t0, &t0, &v3)
	FeMul(out, &t0, u) /* u^(m+1)v^(-(m+1)) */
}
func (p *ProjectiveGroupElement) FromBytes(s *Key) {

	// the original code processes it in a different way
	// so we do it in 2 steps
	// first parse using the original code, convert into 32 bit form
	// then pack back to a 32 bytes
	// now we use 64 bit code to parse
	var tmps [32]byte
	{
		h0 := load4(s[:])
		h1 := load3(s[4:]) << 6
		h2 := load3(s[7:]) << 5
		h3 := load3(s[10:]) << 3
		h4 := load3(s[13:]) << 2
		h5 := load4(s[16:])
		h6 := load3(s[20:]) << 7
		h7 := load3(s[23:]) << 5
		h8 := load3(s[26:]) << 4
		h9 := load3(s[29:]) << 2
		var carry [10]int64
		carry[9] = (h9 + int64(1<<24)) >> 25
		h0 += carry[9] * 19
		h9 -= carry[9] << 25
		carry[1] = (h1 + int64(1<<24)) >> 25
		h2 += carry[1]
		h1 -= carry[1] << 25
		carry[3] = (h3 + int64(1<<24)) >> 25
		h4 += carry[3]
		h3 -= carry[3] << 25
		carry[5] = (h5 + int64(1<<24)) >> 25
		h6 += carry[5]
		h5 -= carry[5] << 25
		carry[7] = (h7 + int64(1<<24)) >> 25
		h8 += carry[7]
		h7 -= carry[7] << 25

		carry[0] = (h0 + int64(1<<25)) >> 26
		h1 += carry[0]
		h0 -= carry[0] << 26
		carry[2] = (h2 + int64(1<<25)) >> 26
		h3 += carry[2]
		h2 -= carry[2] << 26
		carry[4] = (h4 + int64(1<<25)) >> 26
		h5 += carry[4]
		h4 -= carry[4] << 26
		carry[6] = (h6 + int64(1<<25)) >> 26
		h7 += carry[6]
		h6 -= carry[6] << 26
		carry[8] = (h8 + int64(1<<25)) >> 26
		h9 += carry[8]
		h8 -= carry[8] << 26

		var u FieldElement32
		u[0] = int32(h0)
		u[1] = int32(h1)
		u[2] = int32(h2)
		u[3] = int32(h3)
		u[4] = int32(h4)
		u[5] = int32(h5)
		u[6] = int32(h6)
		u[7] = int32(h7)
		u[8] = int32(h8)
		u[9] = int32(h9)

		FeToBytes32(&tmps, &u)

	}
	var u, v, w, x, y, z FieldElement

	// the preprocessed key is stored in tmps
	var tmp_key Key
	copy(tmp_key[:], tmps[:])
	FeFromBytes(&u, &tmp_key)

	FeSquare2(&v, &u) /* 2 * u^2 */
	FeOne(&w)
	FeAdd(&w, &v, &w) /* w = 2 * u^2 + 1 */
	FeSquare(&x, &w)  /* w^2 */

	FeMul(&y, &FeMa2, &v)    /* -2 * A^2 * u^2 */
	FeAdd(&x, &x, &y)        /* x = w^2 - 2 * A^2 * u^2 */
	FeDivPowM1(&p.X, &w, &x) /* (w / x)^(m + 1) */
	FeSquare(&y, &p.X)
	FeMul(&x, &y, &x)
	FeSub(&y, &w, &x)
	FeCopy(&z, &FeMa)
	isNegative := false
	var sign byte
	if FeIsNonZero(&y) != 0 {
		FeAdd(&y, &w, &x)
		if FeIsNonZero(&y) != 0 {
			isNegative = true
		} else {
			FeMul(&p.X, &p.X, &FeFffb1)
		}
	} else {
		FeMul(&p.X, &p.X, &FeFffb2)
	}
	if isNegative {
		FeMul(&x, &x, &FeSqrtM1)
		FeSub(&y, &w, &x)
		if FeIsNonZero(&y) != 0 {
			FeAdd(&y, &w, &x)
			FeMul(&p.X, &p.X, &FeFffb3)
		} else {
			FeMul(&p.X, &p.X, &FeFffb4)
		}
		/* p.X = sqrt(A * (A + 2) * w / x) */
		/* z = -A */
		sign = 1
	} else {
		FeMul(&p.X, &p.X, &u) /* u * sqrt(2 * A * (A + 2) * w / x) */
		FeMul(&z, &z, &v)     /* -2 * A * u^2 */
		sign = 0
	}
	if FeIsNegative(&p.X) != sign {
		FeNeg(&p.X, &p.X)
	}
	FeAdd(&p.Z, &z, &w)
	FeSub(&p.Y, &z, &w)
	FeMul(&p.X, &p.X, &p.Z)

}

func (p *CompletedGroupElement) ToProjective(r *ProjectiveGroupElement) {
	FeMul(&r.X, &p.X, &p.T)
	FeMul(&r.Y, &p.Y, &p.Z)
	FeMul(&r.Z, &p.Z, &p.T)
}

func (p *CompletedGroupElement) ToExtended(r *ExtendedGroupElement) {
	FeMul(&r.X, &p.X, &p.T)
	FeMul(&r.Y, &p.Y, &p.Z)
	FeMul(&r.Z, &p.Z, &p.T)
	FeMul(&r.T, &p.X, &p.Y)
}

func (p *PreComputedGroupElement) Zero() {
	FeOne(&p.yPlusX)
	FeOne(&p.yMinusX)
	FeZero(&p.xy2d)
}

// exported version for ringct
func GeAdd(r *CompletedGroupElement, p *ExtendedGroupElement, q *CachedGroupElement) {
	geAdd(r, p, q)
}
func geAdd(r *CompletedGroupElement, p *ExtendedGroupElement, q *CachedGroupElement) {
	var t0 FieldElement

	FeAdd(&r.X, &p.Y, &p.X)
	FeSub(&r.Y, &p.Y, &p.X)
	FeMul(&r.Z, &r.X, &q.yPlusX)
	FeMul(&r.Y, &r.Y, &q.yMinusX)
	FeMul(&r.T, &q.T2d, &p.T)
	FeMul(&r.X, &p.Z, &q.Z)
	FeAdd(&t0, &r.X, &r.X)
	FeSub(&r.X, &r.Z, &r.Y)
	FeAdd(&r.Y, &r.Z, &r.Y)
	FeAdd(&r.Z, &t0, &r.T)
	FeSub(&r.T, &t0, &r.T)
}
func geMixedAdd(r *CompletedGroupElement, p *ExtendedGroupElement, q *PreComputedGroupElement) {
	var t0 FieldElement

	FeAdd(&r.X, &p.Y, &p.X)
	FeSub(&r.Y, &p.Y, &p.X)
	FeMul(&r.Z, &r.X, &q.yPlusX)
	FeMul(&r.Y, &r.Y, &q.yMinusX)
	FeMul(&r.T, &q.xy2d, &p.T)
	FeAdd(&t0, &p.Z, &p.Z)
	FeSub(&r.X, &r.Z, &r.Y)
	FeAdd(&r.Y, &r.Z, &r.Y)
	FeAdd(&r.Z, &t0, &r.T)
	FeSub(&r.T, &t0, &r.T)
}

func geSub(r *CompletedGroupElement, p *ExtendedGroupElement, q *CachedGroupElement) {
	var t0 FieldElement

	FeAdd(&r.X, &p.Y, &p.X)
	FeSub(&r.Y, &p.Y, &p.X)
	FeMul(&r.Z, &r.X, &q.yMinusX)
	FeMul(&r.Y, &r.Y, &q.yPlusX)
	FeMul(&r.T, &q.T2d, &p.T)
	FeMul(&r.X, &p.Z, &q.Z)
	FeAdd(&t0, &r.X, &r.X)
	FeSub(&r.X, &r.Z, &r.Y)
	FeAdd(&r.Y, &r.Z, &r.Y)
	FeSub(&r.Z, &t0, &r.T)
	FeAdd(&r.T, &t0, &r.T)
}

func geMixedSub(r *CompletedGroupElement, p *ExtendedGroupElement, q *PreComputedGroupElement) {
	var t0 FieldElement

	FeAdd(&r.X, &p.Y, &p.X)
	FeSub(&r.Y, &p.Y, &p.X)
	FeMul(&r.Z, &r.X, &q.yMinusX)
	FeMul(&r.Y, &r.Y, &q.yPlusX)
	FeMul(&r.T, &q.xy2d, &p.T)
	FeAdd(&t0, &p.Z, &p.Z)
	FeSub(&r.X, &r.Z, &r.Y)
	FeAdd(&r.Y, &r.Z, &r.Y)
	FeSub(&r.Z, &t0, &r.T)
	FeAdd(&r.T, &t0, &r.T)
}

func slide(r *[256]int8, a *Key) {
	for i := range r {
		r[i] = int8(1 & (a[i>>3] >> uint(i&7)))
	}

	for i := range r {
		if r[i] != 0 {
			for b := 1; b <= 6 && i+b < 256; b++ {
				if r[i+b] != 0 {
					if r[i]+(r[i+b]<<uint(b)) <= 15 {
						r[i] += r[i+b] << uint(b)
						r[i+b] = 0
					} else if r[i]-(r[i+b]<<uint(b)) >= -15 {
						r[i] -= r[i+b] << uint(b)
						for k := i + b; k < 256; k++ {
							if r[k] == 0 {
								r[k] = 1
								break
							}
							r[k] = 0
						}
					} else {
						break
					}
				}
			}
		}
	}
}

// sets r = a*A + b*G
// this is an optimised version, unoptimised version is 4 lines below
func GeDoubleScalarMultVartime(r *ProjectiveGroupElement, a *Key, A *ExtendedGroupElement, b *Key) {
	var Ai [8]CachedGroupElement // A,3A,5A,7A,9A,11A,13A,15A
	GePrecompute(&Ai, A)
	GeDoubleScalarMultPrecompVartime2(r, a, &Ai, b, &GBASE_Cached)
}

// GeDoubleScalarMultVartime sets r = a*A + b*B
// where a = a[0]+256*a[1]+...+256^31 a[31].
// and b = b[0]+256*b[1]+...+256^31 b[31].
// B is the Ed25519 base point (x,4/5) with x positive.
/*
func GeDoubleScalarMultVartime(r *ProjectiveGroupElement, a *Key, A *ExtendedGroupElement, b *Key) {
	var aSlide, bSlide [256]int8
	var Ai [8]CachedGroupElement // A,3A,5A,7A,9A,11A,13A,15A
	var t CompletedGroupElement
	var u, A2 ExtendedGroupElement
	var i int

	slide(&aSlide, a)
	slide(&bSlide, b)

	A.ToCached(&Ai[0])
	A.Double(&t)
	t.ToExtended(&A2)

	for i := 0; i < 7; i++ {
		geAdd(&t, &A2, &Ai[i])
		t.ToExtended(&u)
		u.ToCached(&Ai[i+1])
	}

	r.Zero()

	for i = 255; i >= 0; i-- {
		if aSlide[i] != 0 || bSlide[i] != 0 {
			break
		}
	}

	for ; i >= 0; i-- {
		r.Double(&t)

		if aSlide[i] > 0 {
			t.ToExtended(&u)
			geAdd(&t, &u, &Ai[aSlide[i]/2])
		} else if aSlide[i] < 0 {
			t.ToExtended(&u)
			geSub(&t, &u, &Ai[(-aSlide[i])/2])
		}

		if bSlide[i] > 0 {
			t.ToExtended(&u)
			geMixedAdd(&t, &u, &bi[bSlide[i]/2])
		} else if bSlide[i] < 0 {
			t.ToExtended(&u)
			geMixedSub(&t, &u, &bi[(-bSlide[i])/2])
		}

		t.ToProjective(r)
	}
}
*/

// equal returns 1 if b == c and 0 otherwise, assuming that b and c are
// non-negative.
func equal(b, c int32) int32 {
	x := uint32(b ^ c)
	x--
	return int32(x >> 31)
}

// negative returns 1 if b < 0 and 0 otherwise.
func negative(b int32) int32 {
	return (b >> 31) & 1
}

func PreComputedGroupElementCMove(t, u *PreComputedGroupElement, b int32) {
	FeCMove(&t.yPlusX, &u.yPlusX, b)
	FeCMove(&t.yMinusX, &u.yMinusX, b)
	FeCMove(&t.xy2d, &u.xy2d, b)
}

func selectPoint(t *PreComputedGroupElement, pos int32, b int32) {
	var minusT PreComputedGroupElement
	bNegative := negative(b)
	bAbs := b - (((-bNegative) & b) << 1)

	t.Zero()
	for i := int32(0); i < 8; i++ {
		PreComputedGroupElementCMove(t, &base[pos][i], equal(bAbs, i+1))
	}
	FeCopy(&minusT.yPlusX, &t.yMinusX)
	FeCopy(&minusT.yMinusX, &t.yPlusX)
	FeNeg(&minusT.xy2d, &t.xy2d)
	PreComputedGroupElementCMove(t, &minusT, bNegative)
}

// GeScalarMultBase computes h = a*B, where
//   a = a[0]+256*a[1]+...+256^31 a[31]
//   B is the Ed25519 base point (x,4/5) with x positive.
//
// Preconditions:
//   a[31] <= 127
func GeScalarMultBase(h *ExtendedGroupElement, a *Key) {
	var e [64]int8

	for i, v := range a {
		e[2*i] = int8(v & 15)
		e[2*i+1] = int8((v >> 4) & 15)
	}

	// each e[i] is between 0 and 15 and e[63] is between 0 and 7.

	carry := int8(0)
	for i := 0; i < 63; i++ {
		e[i] += carry
		carry = (e[i] + 8) >> 4
		e[i] -= carry << 4
	}
	e[63] += carry
	// each e[i] is between -8 and 8.

	h.Zero()
	var t PreComputedGroupElement
	var r CompletedGroupElement
	for i := int32(1); i < 64; i += 2 {
		selectPoint(&t, i/2, int32(e[i]))
		geMixedAdd(&r, h, &t)
		r.ToExtended(h)
	}

	var s ProjectiveGroupElement

	h.Double(&r)
	r.ToProjective(&s)
	s.Double(&r)
	r.ToProjective(&s)
	s.Double(&r)
	r.ToProjective(&s)
	s.Double(&r)
	r.ToExtended(h)

	for i := int32(0); i < 64; i += 2 {
		selectPoint(&t, i/2, int32(e[i]))
		geMixedAdd(&r, h, &t)
		r.ToExtended(h)
	}
}

// r = 8 * t
func GeMul8(r *CompletedGroupElement, t *ProjectiveGroupElement) {
	var u ProjectiveGroupElement
	t.Double(r)
	r.ToProjective(&u)
	u.Double(r)
	r.ToProjective(&u)
	u.Double(r)
}

// caches s into an array of CachedGroupElements for scalar multiplication later
func GePrecompute(r *[8]CachedGroupElement, s *ExtendedGroupElement) {
	var t CompletedGroupElement
	var s2, u ExtendedGroupElement
	s.ToCached(&r[0])
	s.Double(&t)
	t.ToExtended(&s2)
	for i := 0; i < 7; i++ {
		geAdd(&t, &s2, &r[i])
		t.ToExtended(&u)
		u.ToCached(&r[i+1])
	}
}

// sets r = a*A + b*B
// where Bi is the [8]CachedGroupElement consisting of
// B,3B,5B,7B,9B,11B,13B,15B
func GeDoubleScalarMultPrecompVartime2(r *ProjectiveGroupElement, a *Key, Ai *[8]CachedGroupElement, b *Key, Bi *[8]CachedGroupElement) {
	var aSlide, bSlide [256]int8
	//var Ai [8]CachedGroupElement // A,3A,5A,7A,9A,11A,13A,15A
	var t CompletedGroupElement
	var u ExtendedGroupElement
	var i int
	slide(&aSlide, a)
	slide(&bSlide, b)
	//GePrecompute(&Ai, A)
	r.Zero()
	for i = 255; i >= 0; i-- {
		if aSlide[i] != 0 || bSlide[i] != 0 {
			break
		}
	}
	for ; i >= 0; i-- {
		r.Double(&t)
		if aSlide[i] > 0 {
			t.ToExtended(&u)
			geAdd(&t, &u, &Ai[aSlide[i]/2])
		} else if aSlide[i] < 0 {
			t.ToExtended(&u)
			geSub(&t, &u, &Ai[(-aSlide[i])/2])
		}
		if bSlide[i] > 0 {
			t.ToExtended(&u)
			geAdd(&t, &u, &Bi[bSlide[i]/2])
		} else if bSlide[i] < 0 {
			t.ToExtended(&u)
			geSub(&t, &u, &Bi[(-bSlide[i])/2])
		}
		t.ToProjective(r)
	}
	return
}

// sets r = a*A + b*B
// where Bi is the [8]CachedGroupElement consisting of
// B,3B,5B,7B,9B,11B,13B,15B
func GeDoubleScalarMultPrecompVartime(r *ProjectiveGroupElement, a *Key, A *ExtendedGroupElement, b *Key, Bi *[8]CachedGroupElement) {
	var Ai [8]CachedGroupElement // A,3A,5A,7A,9A,11A,13A,15A
	GePrecompute(&Ai, A)
	GeDoubleScalarMultPrecompVartime2(r, a, &Ai, b, Bi)
}

func CachedGroupElementCMove(t, u *CachedGroupElement, b int32) {
	if b == 0 {
		return
	}
	FeCMove(&t.yPlusX, &u.yPlusX, b)
	FeCMove(&t.yMinusX, &u.yMinusX, b)
	FeCMove(&t.Z, &u.Z, b)
	FeCMove(&t.T2d, &u.T2d, b)
}

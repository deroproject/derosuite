package crypto

// FieldElement32 represents an element of the field GF(2^255 - 19). An element
// t, entries t[0]...t[9], represents the integer t[0]+2^26 t[1]+2^51 t[2]+2^77
// t[3]+2^102 t[4]+...+2^230 t[9].  Bounds on each t[i] vary depending on
// context.
type FieldElement32 [10]int32

// FieldElement64 represents an element of the field GF(2^255-19). An element t
// represents the integer t[0] + t[1]*2^51 + t[2]*2^102 + t[3]*2^153 +
// t[4]*2^204.
type FieldElement64 [5]uint64

func FeFromBytes32(dst *FieldElement32, src *[32]byte) {
	h0 := load4(src[:])
	h1 := load3(src[4:]) << 6
	h2 := load3(src[7:]) << 5
	h3 := load3(src[10:]) << 3
	h4 := load3(src[13:]) << 2
	h5 := load4(src[16:])
	h6 := load3(src[20:]) << 7
	h7 := load3(src[23:]) << 5
	h8 := load3(src[26:]) << 4
	h9 := (load3(src[29:]) & 8388607) << 2

	feCombine(dst, h0, h1, h2, h3, h4, h5, h6, h7, h8, h9)
}

func FeToBytes32(s *[32]byte, h *FieldElement32) {
	var carry [10]int32

	q := (19*h[9] + (1 << 24)) >> 25
	q = (h[0] + q) >> 26
	q = (h[1] + q) >> 25
	q = (h[2] + q) >> 26
	q = (h[3] + q) >> 25
	q = (h[4] + q) >> 26
	q = (h[5] + q) >> 25
	q = (h[6] + q) >> 26
	q = (h[7] + q) >> 25
	q = (h[8] + q) >> 26
	q = (h[9] + q) >> 25

	// Goal: Output h-(2^255-19)q, which is between 0 and 2^255-20.
	h[0] += 19 * q
	// Goal: Output h-2^255 q, which is between 0 and 2^255-20.

	carry[0] = h[0] >> 26
	h[1] += carry[0]
	h[0] -= carry[0] << 26
	carry[1] = h[1] >> 25
	h[2] += carry[1]
	h[1] -= carry[1] << 25
	carry[2] = h[2] >> 26
	h[3] += carry[2]
	h[2] -= carry[2] << 26
	carry[3] = h[3] >> 25
	h[4] += carry[3]
	h[3] -= carry[3] << 25
	carry[4] = h[4] >> 26
	h[5] += carry[4]
	h[4] -= carry[4] << 26
	carry[5] = h[5] >> 25
	h[6] += carry[5]
	h[5] -= carry[5] << 25
	carry[6] = h[6] >> 26
	h[7] += carry[6]
	h[6] -= carry[6] << 26
	carry[7] = h[7] >> 25
	h[8] += carry[7]
	h[7] -= carry[7] << 25
	carry[8] = h[8] >> 26
	h[9] += carry[8]
	h[8] -= carry[8] << 26
	carry[9] = h[9] >> 25
	h[9] -= carry[9] << 25
	// h10 = carry9

	// Goal: Output h[0]+...+2^255 h10-2^255 q, which is between 0 and 2^255-20.
	// Have h[0]+...+2^230 h[9] between 0 and 2^255-1;
	// evidently 2^255 h10-2^255 q = 0.
	// Goal: Output h[0]+...+2^230 h[9].

	s[0] = byte(h[0] >> 0)
	s[1] = byte(h[0] >> 8)
	s[2] = byte(h[0] >> 16)
	s[3] = byte((h[0] >> 24) | (h[1] << 2))
	s[4] = byte(h[1] >> 6)
	s[5] = byte(h[1] >> 14)
	s[6] = byte((h[1] >> 22) | (h[2] << 3))
	s[7] = byte(h[2] >> 5)
	s[8] = byte(h[2] >> 13)
	s[9] = byte((h[2] >> 21) | (h[3] << 5))
	s[10] = byte(h[3] >> 3)
	s[11] = byte(h[3] >> 11)
	s[12] = byte((h[3] >> 19) | (h[4] << 6))
	s[13] = byte(h[4] >> 2)
	s[14] = byte(h[4] >> 10)
	s[15] = byte(h[4] >> 18)
	s[16] = byte(h[5] >> 0)
	s[17] = byte(h[5] >> 8)
	s[18] = byte(h[5] >> 16)
	s[19] = byte((h[5] >> 24) | (h[6] << 1))
	s[20] = byte(h[6] >> 7)
	s[21] = byte(h[6] >> 15)
	s[22] = byte((h[6] >> 23) | (h[7] << 3))
	s[23] = byte(h[7] >> 5)
	s[24] = byte(h[7] >> 13)
	s[25] = byte((h[7] >> 21) | (h[8] << 4))
	s[26] = byte(h[8] >> 4)
	s[27] = byte(h[8] >> 12)
	s[28] = byte((h[8] >> 20) | (h[9] << 6))
	s[29] = byte(h[9] >> 2)
	s[30] = byte(h[9] >> 10)
	s[31] = byte(h[9] >> 18)
}

func feCombine(h *FieldElement32, h0, h1, h2, h3, h4, h5, h6, h7, h8, h9 int64) {
	var c0, c1, c2, c3, c4, c5, c6, c7, c8, c9 int64

	/*
	  |h0| <= (1.1*1.1*2^52*(1+19+19+19+19)+1.1*1.1*2^50*(38+38+38+38+38))
	    i.e. |h0| <= 1.2*2^59; narrower ranges for h2, h4, h6, h8
	  |h1| <= (1.1*1.1*2^51*(1+1+19+19+19+19+19+19+19+19))
	    i.e. |h1| <= 1.5*2^58; narrower ranges for h3, h5, h7, h9
	*/

	c0 = (h0 + (1 << 25)) >> 26
	h1 += c0
	h0 -= c0 << 26
	c4 = (h4 + (1 << 25)) >> 26
	h5 += c4
	h4 -= c4 << 26
	/* |h0| <= 2^25 */
	/* |h4| <= 2^25 */
	/* |h1| <= 1.51*2^58 */
	/* |h5| <= 1.51*2^58 */

	c1 = (h1 + (1 << 24)) >> 25
	h2 += c1
	h1 -= c1 << 25
	c5 = (h5 + (1 << 24)) >> 25
	h6 += c5
	h5 -= c5 << 25
	/* |h1| <= 2^24; from now on fits into int32 */
	/* |h5| <= 2^24; from now on fits into int32 */
	/* |h2| <= 1.21*2^59 */
	/* |h6| <= 1.21*2^59 */

	c2 = (h2 + (1 << 25)) >> 26
	h3 += c2
	h2 -= c2 << 26
	c6 = (h6 + (1 << 25)) >> 26
	h7 += c6
	h6 -= c6 << 26
	/* |h2| <= 2^25; from now on fits into int32 unchanged */
	/* |h6| <= 2^25; from now on fits into int32 unchanged */
	/* |h3| <= 1.51*2^58 */
	/* |h7| <= 1.51*2^58 */

	c3 = (h3 + (1 << 24)) >> 25
	h4 += c3
	h3 -= c3 << 25
	c7 = (h7 + (1 << 24)) >> 25
	h8 += c7
	h7 -= c7 << 25
	/* |h3| <= 2^24; from now on fits into int32 unchanged */
	/* |h7| <= 2^24; from now on fits into int32 unchanged */
	/* |h4| <= 1.52*2^33 */
	/* |h8| <= 1.52*2^33 */

	c4 = (h4 + (1 << 25)) >> 26
	h5 += c4
	h4 -= c4 << 26
	c8 = (h8 + (1 << 25)) >> 26
	h9 += c8
	h8 -= c8 << 26
	/* |h4| <= 2^25; from now on fits into int32 unchanged */
	/* |h8| <= 2^25; from now on fits into int32 unchanged */
	/* |h5| <= 1.01*2^24 */
	/* |h9| <= 1.51*2^58 */

	c9 = (h9 + (1 << 24)) >> 25
	h0 += c9 * 19
	h9 -= c9 << 25
	/* |h9| <= 2^24; from now on fits into int32 unchanged */
	/* |h0| <= 1.8*2^37 */

	c0 = (h0 + (1 << 25)) >> 26
	h1 += c0
	h0 -= c0 << 26
	/* |h0| <= 2^25; from now on fits into int32 unchanged */
	/* |h1| <= 1.01*2^24 */

	h[0] = int32(h0)
	h[1] = int32(h1)
	h[2] = int32(h2)
	h[3] = int32(h3)
	h[4] = int32(h4)
	h[5] = int32(h5)
	h[6] = int32(h6)
	h[7] = int32(h7)
	h[8] = int32(h8)
	h[9] = int32(h9)
}

func FeFromBytes64(v *FieldElement64, x *[32]byte) {
	v[0] = uint64(x[0])
	v[0] |= uint64(x[1]) << 8
	v[0] |= uint64(x[2]) << 16
	v[0] |= uint64(x[3]) << 24
	v[0] |= uint64(x[4]) << 32
	v[0] |= uint64(x[5]) << 40
	v[0] |= uint64(x[6]&7) << 48

	v[1] = uint64(x[6]) >> 3
	v[1] |= uint64(x[7]) << 5
	v[1] |= uint64(x[8]) << 13
	v[1] |= uint64(x[9]) << 21
	v[1] |= uint64(x[10]) << 29
	v[1] |= uint64(x[11]) << 37
	v[1] |= uint64(x[12]&63) << 45

	v[2] = uint64(x[12]) >> 6
	v[2] |= uint64(x[13]) << 2
	v[2] |= uint64(x[14]) << 10
	v[2] |= uint64(x[15]) << 18
	v[2] |= uint64(x[16]) << 26
	v[2] |= uint64(x[17]) << 34
	v[2] |= uint64(x[18]) << 42
	v[2] |= uint64(x[19]&1) << 50

	v[3] = uint64(x[19]) >> 1
	v[3] |= uint64(x[20]) << 7
	v[3] |= uint64(x[21]) << 15
	v[3] |= uint64(x[22]) << 23
	v[3] |= uint64(x[23]) << 31
	v[3] |= uint64(x[24]) << 39
	v[3] |= uint64(x[25]&15) << 47

	v[4] = uint64(x[25]) >> 4
	v[4] |= uint64(x[26]) << 4
	v[4] |= uint64(x[27]) << 12
	v[4] |= uint64(x[28]) << 20
	v[4] |= uint64(x[29]) << 28
	v[4] |= uint64(x[30]) << 36
	v[4] |= uint64(x[31]&127) << 44
}

func FeToBytes64(r *[32]byte, v *FieldElement64) {
	var t FieldElement64
	feReduce64(&t, v)

	r[0] = byte(t[0] & 0xff)
	r[1] = byte((t[0] >> 8) & 0xff)
	r[2] = byte((t[0] >> 16) & 0xff)
	r[3] = byte((t[0] >> 24) & 0xff)
	r[4] = byte((t[0] >> 32) & 0xff)
	r[5] = byte((t[0] >> 40) & 0xff)
	r[6] = byte((t[0] >> 48))

	r[6] ^= byte((t[1] << 3) & 0xf8)
	r[7] = byte((t[1] >> 5) & 0xff)
	r[8] = byte((t[1] >> 13) & 0xff)
	r[9] = byte((t[1] >> 21) & 0xff)
	r[10] = byte((t[1] >> 29) & 0xff)
	r[11] = byte((t[1] >> 37) & 0xff)
	r[12] = byte((t[1] >> 45))

	r[12] ^= byte((t[2] << 6) & 0xc0)
	r[13] = byte((t[2] >> 2) & 0xff)
	r[14] = byte((t[2] >> 10) & 0xff)
	r[15] = byte((t[2] >> 18) & 0xff)
	r[16] = byte((t[2] >> 26) & 0xff)
	r[17] = byte((t[2] >> 34) & 0xff)
	r[18] = byte((t[2] >> 42) & 0xff)
	r[19] = byte((t[2] >> 50))

	r[19] ^= byte((t[3] << 1) & 0xfe)
	r[20] = byte((t[3] >> 7) & 0xff)
	r[21] = byte((t[3] >> 15) & 0xff)
	r[22] = byte((t[3] >> 23) & 0xff)
	r[23] = byte((t[3] >> 31) & 0xff)
	r[24] = byte((t[3] >> 39) & 0xff)
	r[25] = byte((t[3] >> 47))

	r[25] ^= byte((t[4] << 4) & 0xf0)
	r[26] = byte((t[4] >> 4) & 0xff)
	r[27] = byte((t[4] >> 12) & 0xff)
	r[28] = byte((t[4] >> 20) & 0xff)
	r[29] = byte((t[4] >> 28) & 0xff)
	r[30] = byte((t[4] >> 36) & 0xff)
	r[31] = byte((t[4] >> 44))
}

const maskLow51Bits = (1 << 51) - 1

func feReduce64(t, v *FieldElement64) {
	// Copy v
	*t = *v

	// Let v = v[0] + v[1]*2^51 + v[2]*2^102 + v[3]*2^153 + v[4]*2^204
	// Reduce each limb below 2^51, propagating carries.
	t[1] += t[0] >> 51
	t[0] = t[0] & maskLow51Bits
	t[2] += t[1] >> 51
	t[1] = t[1] & maskLow51Bits
	t[3] += t[2] >> 51
	t[2] = t[2] & maskLow51Bits
	t[4] += t[3] >> 51
	t[3] = t[3] & maskLow51Bits
	t[0] += (t[4] >> 51) * 19
	t[4] = t[4] & maskLow51Bits

	// We now have a field element t < 2^255, but need t <= 2^255-19

	// Get the carry bit
	c := (t[0] + 19) >> 51
	c = (t[1] + c) >> 51
	c = (t[2] + c) >> 51
	c = (t[3] + c) >> 51
	c = (t[4] + c) >> 51

	t[0] += 19 * c

	t[1] += t[0] >> 51
	t[0] = t[0] & maskLow51Bits
	t[2] += t[1] >> 51
	t[1] = t[1] & maskLow51Bits
	t[3] += t[2] >> 51
	t[2] = t[2] & maskLow51Bits
	t[4] += t[3] >> 51
	t[3] = t[3] & maskLow51Bits
	// no additional carry
	t[4] = t[4] & maskLow51Bits
}

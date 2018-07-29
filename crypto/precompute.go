// Copyright 2017-2018 DERO Project. All rights reserved.
// Use of this source code in any form is governed by RESEARCH license.
// license can be found in the LICENSE file.
// GPG: 0F39 E425 8C65 3947 702A  8234 08B2 0360 A03A 9DE8
//
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY
// EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL
// THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
// PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT,
// STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF
// THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

// this file includes the code related to precomputation generation
// and scalar multiplication
// each table is roughly 40 KB
// 32 such tables form a super table, 32 * 40KB

package crypto

import "fmt"

// table types
type PRECOMPUTE_TABLE [256]CachedGroupElement
type SUPER_PRECOMPUTE_TABLE [32]PRECOMPUTE_TABLE

type FAST_TABLE [256]PreComputedGroupElement

var x FAST_TABLE

//import "fmt"
//import "time"
//import "bytes"

// precompuate table of the form A,2A,4A,8A,16A.....2^255A
// se gemul8
// pre compute table size in bytes 256 * 4 field elemets * 5 limbs * 8 bytes
func GenPrecompute(table *PRECOMPUTE_TABLE, A Key) {
	var lbi, tmp ExtendedGroupElement
	lbi.FromBytes(&A)
	lbi.ToCached(&table[0]) // store first entry
	lbi.ToPreComputed(&x[0])

	var t ProjectiveGroupElement
	var r CompletedGroupElement
	lbi.ToProjective(&t)

	var output Key

	//fmt.Printf("%d %s\n",0, A)

	for i := 1; i < 256; i++ {
		t.Double(&r)
		r.ToProjective(&t)
		r.ToExtended(&tmp)
		tmp.ToCached(&table[i])
		tmp.ToBytes(&output)

		tmp.ToPreComputed(&x[i])

		//fmt.Printf("%d %s\n",i, output)

	}

	lbi.ToProjective(&t)
	GeMul8(&r, &t)
	r.ToExtended(&tmp)

	tmp.ToBytes(&output)

	//fmt.Printf("%d %s\n",8, output)

}

const BITS_PER_BYTE = (8)

// it finds the highest bit that is high
// remember the scalar is stored in little Endian
func GetBit(x *Key, pos uint) int {
	if (((x)[(pos)/BITS_PER_BYTE]) & (0x1 << ((pos) % BITS_PER_BYTE))) != 0 {
		return 1
	}
	return 0
}

// fastest  Fixed based multiplicatipnn in 7.5 microseconds
func ScalarMultPrecompute(output *ExtendedGroupElement, scalar *Key, table *PRECOMPUTE_TABLE) {
	output.Zero() // make it identity

	// extract all bits of scalar
	for i := uint(0); i < 256; i++ {
		if GetBit(scalar, i) == 1 { // add precomputed table points
			var c CompletedGroupElement
			geAdd(&c, output, &table[i])
			c.ToExtended(output)

		}
	}
}

// this generates a very large 32*256 cached elements table roughly 1.28 MB
func GenSuperPrecompute(stable *SUPER_PRECOMPUTE_TABLE, ptable *PRECOMPUTE_TABLE) {

	var scalar Key

	for i := 0; i < 32; i++ {
		Sc_0(&scalar)
		for j := 0; j < 256; j++ {
			scalar[i] = byte(j)

			var result_extended ExtendedGroupElement
			ScalarMultPrecompute(&result_extended, &scalar, ptable)

			result_extended.ToCached(&stable[i][j])
		}
	}

}

func ScalarMultSuperPrecompute(output *ExtendedGroupElement, scalar *Key, stable *SUPER_PRECOMPUTE_TABLE) {
	output.Zero() // make it identity

	// extract byte wise scalar and do point add
	for i := uint(0); i < 32; i++ {
		var c CompletedGroupElement
		geAdd(&c, output, &stable[i][scalar[i]])
		c.ToExtended(output)

	}
}

// generate tables of the form, first element is identity
// 0,A,2A,3A........255A
func MulPrecompute(r *PRECOMPUTE_TABLE, s *ExtendedGroupElement) {
	var t CompletedGroupElement
	var s2, u ExtendedGroupElement

	var id ExtendedGroupElement
	id.Zero()
	id.ToCached(&r[0])
	s.ToCached(&r[1])
	s.Double(&t)
	t.ToExtended(&s2)
	for i := 0; i < 254; i++ {
		geAdd(&t, &s2, &r[i])
		t.ToExtended(&u)
		u.ToCached(&r[i+2])
	}
}

// this table can be used for  double scalar, double base multiplication
// only if the bases are fixed and scalars change
func GenDoublePrecompute(table *PRECOMPUTE_TABLE, A Key, B Key) {

	var ATABLE, BTABLE PRECOMPUTE_TABLE

	var a_ex, b_ex ExtendedGroupElement
	a_ex.FromBytes(&A)
	b_ex.FromBytes(&B)

	MulPrecompute(&ATABLE, &a_ex)
	MulPrecompute(&BTABLE, &b_ex)

	var c CompletedGroupElement

	var extended_identity ExtendedGroupElement
	extended_identity.Zero()

	var output Key

	//fmt.Printf("%d %s\n",0, A)

	for i := 0; i < 256; i++ {

		// split into 2 parts
		apart := i & 0xf
		bpart := (i >> 4) & 0xf

		var apoint, bpoint, result ExtendedGroupElement

		// extract elements from table
		extended_identity.Zero() // extract a point
		geAdd(&c, &extended_identity, &ATABLE[apart])
		c.ToExtended(&apoint)

		extended_identity.Zero() // extract b point
		geAdd(&c, &extended_identity, &BTABLE[bpart])
		c.ToExtended(&bpoint)

		// calculate  result = A+B
		var bcached CachedGroupElement
		var c CompletedGroupElement
		bpoint.ToCached(&bcached)
		geAdd(&c, &apoint, &bcached)
		c.ToExtended(&result)

		// store result to cache
		result.ToCached(&table[i])

		apoint.ToBytes(&output)
		//fmt.Printf("%2x %s\n",i, output)

	}

	//fmt.Printf("%d %s\n",256, output)
}

var z = fmt.Sprintf("uy")

// r = 16 * t
func GeMul16(r *CompletedGroupElement, t *ProjectiveGroupElement) {
	var u ProjectiveGroupElement
	t.Double(r)
	r.ToProjective(&u)
	u.Double(r)
	r.ToProjective(&u)
	u.Double(r)
	r.ToProjective(&u)
	u.Double(r)
}

func multprecompscalar(output *ExtendedGroupElement, s *Key, table *PRECOMPUTE_TABLE) {

	var c CompletedGroupElement
	var p ProjectiveGroupElement

	var output_bytes Key

	output.Zero()
	p.Zero()

	for i := uint(255); i >= 0; i-- {
		p.Double(&c)
		c.ToProjective(&p)
		c.ToExtended(output)

		if GetBit(s, i) == 1 {
			geAdd(&c, output, &table[1])
			c.ToExtended(output)
		} else {
			geAdd(&c, output, &table[0])
			c.ToExtended(output)
		}

		output.ToProjective(&p) // for doubling

		output.ToBytes(&output_bytes)
		//fmt.Printf("%d output %s\n", i,output_bytes)

		if i == 0 {
			break
		}
	}
	/*
	   	p.Double(&c)
	   	c.ToProjective(&p)
	   	p.Double(&c)
	   	c.ToProjective(&p)
	   	p.Double(&c)
	   	c.ToProjective(&p)
	   	p.Double(&c)
	   	c.ToProjective(&p)
	   	p.Double(&c)
	   	c.ToProjective(&p)
	   	p.Double(&c)
	   	c.ToProjective(&p)
	   	p.Double(&c)
	   	c.ToProjective(&p)
	   	p.Double(&c)
	   	c.ToProjective(&p)


	   	c.ToExtended(output)


	   	{ // process high nibble first
	   		//point := ((s1[i]>>4) & 0xf) //| (((s2[i]>>4) & 0xf)<<4)

	   		//fmt.Printf("%d  hpoint %d\n",i, point )
	   		geAdd(&c, output, &ATABLE[s[i]])
	   		c.ToExtended(output)
	   	}

	   /*
	   	{ // process low nibble now
	   		point := ((s1[i]) & 0xf) //| (((s2[i]) & 0xf)<<4)
	   		fmt.Printf("%d  lpoint %d\n",i, point )
	   		geAdd(&c, output, &ATABLE[point])
	   		c.ToExtended(output)
	   	}
	*/
	/*	output.ToBytes(&output_bytes)

		fmt.Printf("%d output %s\n", i,output_bytes)
		output.ToProjective(&p) // for doubling
	*/
	//}

}

// does output = s1*A + s2*B using precomputed tables for A and B
// providing scalars in wrong order will give wrong results
// bases are as they were provide in the order
// tables must have been prepared using GenDoublePrecompute
func DoubleScalarDoubleBaseMulPrecomputed(output *ExtendedGroupElement, s1, s2 *Key, table *PRECOMPUTE_TABLE) {

	var c CompletedGroupElement
	var p ProjectiveGroupElement

	var output_bytes Key
	_ = output_bytes

	output.Zero()
	p.Zero()

	for i := 31; i >= 0; i-- {

		// we are processing 4 bits at a time
		p.Double(&c)
		c.ToProjective(&p)
		p.Double(&c)
		c.ToProjective(&p)
		p.Double(&c)
		c.ToProjective(&p)
		p.Double(&c)
		//c.ToProjective(&p)
		c.ToExtended(output)

		{ // process high nibble first
			point := ((s1[i] >> 4) & 0xf) | (((s2[i] >> 4) & 0xf) << 4)

			geAdd(&c, output, &table[point])
			//c.ToExtended(output)
			c.ToProjective(&p)
		}

		// again 4 bits at a time
		p.Double(&c)
		c.ToProjective(&p)
		p.Double(&c)
		c.ToProjective(&p)
		p.Double(&c)
		c.ToProjective(&p)
		p.Double(&c)
		//c.ToProjective(&p)
		c.ToExtended(output)

		{ // process low nibble now
			point := ((s1[i]) & 0xf) | (((s2[i]) & 0xf) << 4)
			//fmt.Printf("%d  lpoint %d\n",i, point )
			geAdd(&c, output, &table[point])
			c.ToExtended(output)
		}
		//output.ToBytes(&output_bytes)

		//fmt.Printf("%d output %s\n", i,output_bytes)

		output.ToProjective(&p) // for doubling

	}

}

// does output = s1*A + s2*B using precomputed tables for A and B
// providing scalars in wrong order will give wrong results
// bases are as they were provide in the order
// tables must have been prepared using GenDoublePrecompute
// this multiplies 64 double scalars and is used for bulletproofs
func DoubleScalarDoubleBaseMulPrecomputed64(output *ExtendedGroupElement, s1, s2 []Key, table []PRECOMPUTE_TABLE) {

	if len(s1) != 64 || len(s2) != 64 || len(table) != 64 {
		panic("DoubleScalarDoubleBaseMulPrecomputed64 requires 64 members")
	}
	var c CompletedGroupElement
	var p ProjectiveGroupElement

	var output_bytes Key
	_ = output_bytes

	output.Zero()
	p.Zero()

	for i := 31; i >= 0; i-- {

		// we are processing 4 bits at a time
		p.Double(&c)
		c.ToProjective(&p)
		p.Double(&c)
		c.ToProjective(&p)
		p.Double(&c)
		c.ToProjective(&p)
		p.Double(&c)
		//c.ToProjective(&p)
		c.ToExtended(output)

		for j := 0; j < 64; j++ { // process high nibble first
			point := ((s1[j][i] >> 4) & 0xf) | (((s2[j][i] >> 4) & 0xf) << 4)
			if point != 0 { // skip if point is zero
				geAdd(&c, output, &table[j][point])
				c.ToExtended(output)
			}

		}
		c.ToProjective(&p)

		// again 4 bits at a time
		p.Double(&c)
		c.ToProjective(&p)
		p.Double(&c)
		c.ToProjective(&p)
		p.Double(&c)
		c.ToProjective(&p)
		p.Double(&c)
		//c.ToProjective(&p)
		c.ToExtended(output)

		for j := 0; j < 64; j++ { // process low nibble now
			point := ((s1[j][i]) & 0xf) | (((s2[j][i]) & 0xf) << 4)
			if point != 0 { // skip if point is zero
				//fmt.Printf("%d  lpoint %d\n",i, point )
				geAdd(&c, output, &table[j][point])
				c.ToExtended(output)
			}
		}
		//output.ToBytes(&output_bytes)

		//fmt.Printf("%d output %s\n", i,output_bytes)

		output.ToProjective(&p) // for doubling

	}

}

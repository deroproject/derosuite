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

package ringct

//import "fmt"

const ATOMS = 64 // 64 bit in the amount field

type bits64 [ATOMS]bool

// implementation of d2b from rctTypes.cpp
// lays out the number from lowest bit at pos 0 and highest at bit 63
func d2b_uint64_to_bits(amount uint64) bits64 {
	var bits bits64
	for i := 0; amount != 0; i++ {
		if (amount & 1) == 1 {
			bits[i] = true
		}
		amount = amount >> 1
	}
	return bits
}

//ProveRange and VerifyRange
//ProveRange gives C, and mask such that \sumCi = C
//   c.f. http://eprint.iacr.org/2015/1098 section 5.1
//   and Ci is a commitment to either 0 or 2^i, i=0,...,63
//   thus this proves that "amount" is in [0, 2^64]
//   mask is a such that C = aG + bH, and b = amount
//VerifyRange verifies that \sum Ci = C and that each Ci is a commitment to 0 or 2^i
// this function proves a range using Pedersen  commitment and borromean signatures
// implemented in cryptonote rctSigs.cpp
func ProveRange(C *Key, mask *Key, amount uint64) *RangeSig {
	Sc_0(mask)
	copy(C[:], (*identity())[:]) // set C to identity

	var ai Key64
	var Cih Key64
	var sig RangeSig

	bits := d2b_uint64_to_bits(amount)
	//fmt.Printf("bits %+v\n", bits)

	for i := 0; i < ATOMS; i++ {
		ai[i] = *(RandomScalar()) // grab a random key
		// Sc_0(&ai[i]); // make random key zero  for tesing puprpose // BUG if line is uncommented
		ScReduce32(&ai[i]) // reduce it
		// fmt.Printf("ai[%2d] %x\n",i, ai[i])

		sig.ci[i] = ScalarmultBase(ai[i])
		// fmt.Printf("ci[%2d] %x\n",i, sig.ci[i])
		if bits[i] {
			AddKeys(&sig.ci[i], &sig.ci[i], &H2[i])
		}

		SubKeys(&Cih[i], &sig.ci[i], &H2[i])
		ScAdd(mask, mask, &ai[i])
		AddKeys(C, C, &sig.ci[i])
	}

	//fmt.Print("C   %x\n", *C)

	// TODO caculate Borromean signature here
	sig.asig = GenerateBorromean(ai, sig.ci, Cih, bits)

	return &sig
}

func VerifyRange(c *Key, as RangeSig) bool {
	var CiH Key64
	tmp := identity()
	for i := 0; i < 64; i++ {
		SubKeys(&CiH[i], &as.ci[i], &H2[i])
		AddKeys(tmp, tmp, &as.ci[i])
	}

	//	fmt.Printf("C   %x\n", *c)
	//        fmt.Printf("tmp %x\n", *tmp)
	if *c != *tmp {
		return false
	}
	//return true
	return VerifyBorromean(&as.asig, &as.ci, &CiH)
}

//Borromean (c.f. gmax/andytoshi's paper)
func GenerateBorromean(x Key64, P1 Key64, P2 Key64, indices bits64) BoroSig {
	var bb BoroSig
	var alpha Key64
	var L [2]Key64
	var c Key

	var data_bytes []byte

	for ii := 0; ii < ATOMS; ii++ {
		var naught, prime int
		if indices[ii] {
			naught = 1
		} else {
			naught = 0
		}
		prime = (naught + 1) % 2 // basically it is the inverse of naught

		alpha[ii] = skGen() // generate a new random scalar
		L[naught][ii] = ScalarmultBase(alpha[ii])

		if naught == 0 {
			bb.s1[ii] = skGen()
			c = *(HashToScalar(L[naught][ii][:]))
			AddKeys2(&L[prime][ii], &bb.s1[ii], &c, &P2[ii])
		}
		// original cryptonote does NOT clear out some unset bytes, verify whether it may be a problem for them
		data_bytes = append(data_bytes, L[1][ii][:]...)
	}
	// take the hash of the L1 keys all 64 of them
	// we have been collecting them above
	bb.ee = *(HashToScalar(data_bytes))

	//     fmt.Printf("bb.ee   %x\n", bb.ee)

	var LL, cc Key
	for jj := 0; jj < ATOMS; jj++ {
		if indices[jj] == false {
			ScMulSub(&bb.s0[jj], &x[jj], &bb.ee, &alpha[jj])
		} else {
			bb.s0[jj] = skGen()
			AddKeys2(&LL, &bb.s0[jj], &bb.ee, &P1[jj])
			cc = *(HashToScalar(LL[:]))
			ScMulSub(&bb.s1[jj], &x[jj], &cc, &alpha[jj])
		}
	}

	return bb
}

// Verify the Borromean sig
func VerifyBorromean(b *BoroSig, p1, p2 *Key64) bool {
	var data []byte
	tmp, tmp2 := new(Key), new(Key)
	for i := 0; i < 64; i++ {
		AddKeys2(tmp, &b.s0[i], &b.ee, &p1[i])
		tmp3 := HashToScalar(tmp[:])
		AddKeys2(tmp2, &b.s1[i], tmp3, &p2[i])
		data = append(data, tmp2[:]...)
	}
	computed := HashToScalar(data)

	//        fmt.Printf("comp    %x\n", computed)
	return *computed == b.ee
}

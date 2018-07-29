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

import "fmt"
import "math/big"
import "encoding/binary"

import "github.com/deroproject/derosuite/crypto"

// see the  references such as original paper and multiple implementations
// https://eprint.iacr.org/2017/1066.pdf
// https://blog.chain.com/faster-bulletproofs-with-ristretto-avx2-29450b4490cd
const maxN = 64

var Hi [maxN]crypto.Key
var Gi [maxN]crypto.Key
var Hi_Precomputed [maxN][8]crypto.CachedGroupElement
var Gi_Precomputed [maxN][8]crypto.CachedGroupElement
var TWO crypto.Key = crypto.HexToKey("0200000000000000000000000000000000000000000000000000000000000000")
var oneN = vector_powers(crypto.Identity, maxN)
var twoN = vector_powers(TWO, maxN)
var ip12 = inner_product(oneN, twoN)

func ProveRangeBulletproof(C *crypto.Key, mask *crypto.Key, amount uint64) BulletProof {
	tmpmask := crypto.SkGen()
	copy(mask[:], tmpmask[:])
	proof := BULLETPROOF_Prove_Amount(amount, mask)
	if len(proof.V) != 1 {
		panic(fmt.Sprintf("V has not exactly one element"))
	}
	copy(C[:], proof.V[0][:]) //C = proof.V[0];
	return *proof
}

func get_exponent(base crypto.Key, idx uint64) crypto.Key {

	salt := "bulletproof"
	var idx_buf [9]byte

	idx_buf_size := binary.PutUvarint(idx_buf[:], idx)

	hash_buf := append(base[:], []byte(salt)...)
	hash_buf = append(hash_buf, idx_buf[:idx_buf_size]...)

	output_hash_good := crypto.Key(crypto.Keccak256(hash_buf[:]))

	return crypto.Key(output_hash_good.HashToPoint())

}

// initialize some hard coded constants
func init() {
	for i := uint64(0); i < maxN; i++ {
		Hi[i] = get_exponent(crypto.H, i*2)

		He := new(crypto.ExtendedGroupElement)
		He.FromBytes(&Hi[i])
		crypto.GePrecompute(&Hi_Precomputed[i], He)

		Gi[i] = get_exponent(crypto.H, i*2+1)
		Ge := new(crypto.ExtendedGroupElement)
		Ge.FromBytes(&Gi[i])
		crypto.GePrecompute(&Gi_Precomputed[i], Ge)

		//fmt.Printf("%2d CONSTANTS %s %s\n",i, Hi[i],Gi[i] )

		//panic("DEAD")
	}

	//fmt.Printf("ip12 inner_product %s \n",ip12 )

}

// Given two scalar arrays, construct a vector commitment
func vector_exponent(a []crypto.Key, b []crypto.Key) (result crypto.Key) {

	if len(a) != len(b) {
		panic("Incompatible sizes of a and b")
	}

	if len(a) < maxN {
		panic("Incompatible sizes of a and maxN")
	}

	result = crypto.Identity
	for i := range a {
		var term crypto.Key
		crypto.AddKeys3_3(&term, &a[i], &Gi_Precomputed[i], &b[i], &Hi_Precomputed[i])
		crypto.AddKeys(&result, &result, &term)
	}
	return
}

// Compute a custom vector-scalar commitment
func vector_exponent_custom(A []crypto.Key, B []crypto.Key, a []crypto.Key, b []crypto.Key) (result crypto.Key) {

	if !(len(A) == len(B)) {
		panic("Incompatible sizes of A and B")
	}

	if !(len(a) == len(b)) {
		panic("Incompatible sizes of a and b")
	}
	if !(len(a) == len(A)) {
		panic("Incompatible sizes of a and A")
	}

	if !(len(a) <= maxN) {
		panic("Incompatible sizes of a and maxN")
	}

	result = crypto.Identity
	for i := range a {
		var term crypto.Key

		var B_Precomputed [8]crypto.CachedGroupElement
		Be := new(crypto.ExtendedGroupElement)
		Be.FromBytes(&B[i])
		crypto.GePrecompute(&B_Precomputed, Be)

		crypto.AddKeys3(&term, &a[i], &A[i], &b[i], &B_Precomputed)
		crypto.AddKeys(&result, &result, &term)
	}
	return result

}

// Given a scalar, construct a vector of powers
// NOTE: the below function has bug where the function will panic  if n  == 0 or n == 1 
// However, the code has hardcoded number n = 64, so this is not exploitable as such in current form
func vector_powers(x crypto.Key, n int64) (res []crypto.Key) {
        
        if n < 2 {
            panic("vector powers only support 64 bit inputs/outputs")
        }
        
	res = make([]crypto.Key, n, n)
        res[0] = crypto.Identity  // first 2 are setup manually
	res[1] = x
	
	for i := int64(2); i < n; i++ {
		crypto.ScMul(&res[i], &res[i-1], &x)

		//	fmt.Printf("vector power %2d %s %s %s\n", i, res[i], res[i-1],x)
	}

	return
}

// Given two scalar arrays, construct the inner product
func inner_product(a []crypto.Key, b []crypto.Key) (result crypto.Key) {
	if len(a) != len(b) {
		panic("Incompatible sizes of a and b")
	}
	result = crypto.Zero
	for i := range a {
		crypto.ScMulAdd(&result, &a[i], &b[i], &result)
	}
	return
}

// Given two scalar arrays, construct the Hadamard product
func hadamard(a []crypto.Key, b []crypto.Key) (result []crypto.Key) {
	if len(a) != len(b) {
		panic("Incompatible sizes of a and b")
	}
	result = make([]crypto.Key, len(a), len(a))
	for i := range a {
		crypto.ScMul(&result[i], &a[i], &b[i])
	}
	return
}

// Given two curvepoint arrays, construct the Hadamard product
func hadamard2(a []crypto.Key, b []crypto.Key) (result []crypto.Key) {
	if len(a) != len(b) {
		panic("Incompatible sizes of a and b")
	}
	result = make([]crypto.Key, len(a), len(a))
	for i := range a {
		crypto.AddKeys(&result[i], &a[i], &b[i])
	}
	return
}

// Add two vectors
func vector_add(a []crypto.Key, b []crypto.Key) (result []crypto.Key) {
	if len(a) != len(b) {
		panic("Incompatible sizes of a and b")
	}
	result = make([]crypto.Key, len(a), len(a))
	for i := range a {
		crypto.ScAdd(&result[i], &a[i], &b[i])
	}
	return
}

// substract two vectors
func vector_subtract(a []crypto.Key, b []crypto.Key) (result []crypto.Key) {
	if len(a) != len(b) {
		panic("Incompatible sizes of a and b")
	}
	result = make([]crypto.Key, len(a), len(a))
	for i := range a {
		crypto.ScSub(&result[i], &a[i], &b[i])
	}
	return
}

// Multiply a vector and a scalar
func vector_scalar(a []crypto.Key, x crypto.Key) (result []crypto.Key) {
	result = make([]crypto.Key, len(a), len(a))
	for i := range a {
		crypto.ScMul(&result[i], &a[i], &x)
	}
	return
}

// Exponentiate a vector and a scalar
func vector_scalar2(a []crypto.Key, x crypto.Key) (result []crypto.Key) {
	result = make([]crypto.Key, len(a), len(a))
	for i := range a {
		result[i] = *crypto.ScalarMultKey(&a[i], &x)
	}
	return
}

func Reverse(x crypto.Key) (result crypto.Key) {

	result = x
	// A key is in little-endian, but the big package wants the bytes in
	// big-endian, so reverse them.
	blen := len(x) // its hardcoded 32 bytes, so why do len but lets do it
	for i := 0; i < blen/2; i++ {
		result[i], result[blen-1-i] = result[blen-1-i], result[i]
	}

	return
}

// Compute the inverse of a scalar, the stupid way
func invert_scalar(x crypto.Key) (inverse_result crypto.Key) {

	reversex := Reverse(x)
	bigX := new(big.Int).SetBytes(reversex[:])

	reverseL := Reverse(crypto.CurveOrder()) // as speed improvements it can be made constant
	bigL := new(big.Int).SetBytes(reverseL[:])

	var inverse big.Int
	inverse.ModInverse(bigX, bigL)

	inverse_bytes := inverse.Bytes()

	if len(inverse_bytes) > 32 {
		panic("Inverse cannot be more than 32 bytes in this domain")
	}

	for i, j := 0, len(inverse_bytes)-1; i < j; i, j = i+1, j-1 {
		inverse_bytes[i], inverse_bytes[j] = inverse_bytes[j], inverse_bytes[i]
	}
	copy(inverse_result[:], inverse_bytes[:]) // copy the bytes  as they should be

	return
}

//Compute the slice of a vector, copy and and return it
func slice_vector(a []crypto.Key, start, stop int) (result []crypto.Key) {
	if !(start < int(len(a))) {
		panic("Invalid start index")
	}

	if !(stop <= int(len(a))) {
		panic("Invalid stop index")
	}

	if !(start < stop) {
		panic("Invalid start/stop index")
	}

	result = make([]crypto.Key, stop-start, stop-start)
	for i := start; i < stop; i++ {
		result[i-start] = a[i]
	}
	return
}

// mash of 2 keys with existing mash
func hash_cache_mash2(hashcache *crypto.Key, mash0, mash1 crypto.Key) crypto.Key {

	data_bytes := append(hashcache[:], mash0[:]...)
	data_bytes = append(data_bytes, mash1[:]...)

	mash_result := *(crypto.HashToScalar(data_bytes))
	copy(hashcache[:], mash_result[:]) // update hash cache in place

	return mash_result // and return the mash
}

// mash of 3 keys with existing mash
func hash_cache_mash3(hashcache *crypto.Key, mash0, mash1, mash2 crypto.Key) crypto.Key {

	data_bytes := append(hashcache[:], mash0[:]...)
	data_bytes = append(data_bytes, mash1[:]...)
	data_bytes = append(data_bytes, mash2[:]...)

	mash_result := *(crypto.HashToScalar(data_bytes))
	copy(hashcache[:], mash_result[:]) // update hash cache in place

	return mash_result // and return the mash
}

// mash of 4 keys with existing mash
func hash_cache_mash4(hashcache *crypto.Key, mash0, mash1, mash2, mash3 crypto.Key) crypto.Key {

	data_bytes := append(hashcache[:], mash0[:]...)
	data_bytes = append(data_bytes, mash1[:]...)
	data_bytes = append(data_bytes, mash2[:]...)
	data_bytes = append(data_bytes, mash3[:]...)

	mash_result := *(crypto.HashToScalar(data_bytes))
	copy(hashcache[:], mash_result[:]) // update hash cache in place

	return mash_result // and return the mash
}

// this function is exactly similiar to AddKeys
// add two points together and return the result
func AddKeys_return(k1, k2 *crypto.Key) (result crypto.Key) {
	crypto.AddKeys(&result, k1, k2)
	return
}

//Given a value v (0..2^N-1) and a mask gamma, construct a range proof
func BULLETPROOF_Prove(sv *crypto.Key, gamma *crypto.Key) *BulletProof {
	const logN = int(6) // log2(64)
	const N = int(64)   // 1 << logN

	var V crypto.Key
	var aL, aR [N]crypto.Key
	var A, S crypto.Key

	// prove V
	crypto.AddKeys2(&V, gamma, sv, &crypto.H)

	// prove aL,aR
	// the entire amount in uint64 is extracted into bits and
	// different action taken if bit is zero or bit is one
	for i := N - 1; i >= 0; i-- {
		if (sv[i/8] & (1 << (uint64(i) % 8))) >= 1 {
			aL[i] = crypto.Identity
		} else {
			aL[i] = crypto.Zero
		}
		crypto.ScSub(&aR[i], &aL[i], &crypto.Identity)
	}

	hashcache := *(crypto.HashToScalar(V[:]))

	// prove STEP 1

	// PAPER LINES 38-39
	alpha := crypto.SkGen()
	ve := vector_exponent(aL[:], aR[:])

	alpha_base_tmp := crypto.ScalarmultBase(alpha)
	crypto.AddKeys(&A, &ve, &alpha_base_tmp)

	// PAPER LINES 40-42
	var sL, sR [N]crypto.Key
	for i := range sL {
		sL[i] = crypto.SkGen()
		sR[i] = crypto.SkGen()
	}
	//rct::keyV sL = rct::skvGen(N), sR = rct::skvGen(N);
	rho := crypto.SkGen()
	ve = vector_exponent(sL[:], sR[:])
	rho_base_tmp := crypto.ScalarmultBase(rho)
	crypto.AddKeys(&S, &ve, &rho_base_tmp)

	// PAPER LINES 43-45
	y := hash_cache_mash2(&hashcache, A, S)  //   rct::key y = hash_cache_mash(hash_cache, A, S);
	hashcache = *(crypto.HashToScalar(y[:])) // rct::key z = hash_cache = rct::hash_to_scalar(y);
	z := hashcache

	// Polynomial construction before PAPER LINE 46
	t0 := crypto.Zero // rct::key t0 = rct::zero();
	t1 := crypto.Zero // rct::key t1 = rct::zero();
	t2 := crypto.Zero // rct::key t2 = rct::zero();

	yN := vector_powers(y, int64(N)) // const auto yN = vector_powers(y, N);

	ip1y := inner_product(oneN, yN)      //rct::key ip1y = inner_product(oneN, yN);
	crypto.ScMulAdd(&t0, &z, &ip1y, &t0) // sc_muladd(t0.bytes, z.bytes, ip1y.bytes, t0.bytes);

	var zsq crypto.Key                  //rct::key zsq;
	crypto.ScMul(&zsq, &z, &z)          // sc_mul(zsq.bytes, z.bytes, z.bytes);
	crypto.ScMulAdd(&t0, &zsq, sv, &t0) // sc_muladd(t0.bytes, zsq.bytes, sv.bytes, t0.bytes);

	k := crypto.Zero                     // rct::key k = rct::zero();
	crypto.ScMulSub(&k, &zsq, &ip1y, &k) //sc_mulsub(k.bytes, zsq.bytes, ip1y.bytes, k.bytes);

	var zcu crypto.Key                   //  rct::key zcu;
	crypto.ScMul(&zcu, &zsq, &z)         //sc_mul(zcu.bytes, zsq.bytes, z.bytes);
	crypto.ScMulSub(&k, &zcu, &ip12, &k) //sc_mulsub(k.bytes, zcu.bytes, ip12.bytes, k.bytes);
	crypto.ScAdd(&t0, &t0, &k)           //sc_add(t0.bytes, t0.bytes, k.bytes);

	if DEBUGGING_MODE { // verify intermediate variables for correctness
		test_t0 := crypto.Zero                                  //rct::key test_t0 = rct::zero();
		iph := inner_product(aL[:], hadamard(aR[:], yN))        // rct::key iph = inner_product(aL, hadamard(aR, yN));
		crypto.ScAdd(&test_t0, &test_t0, &iph)                  //sc_add(test_t0.bytes, test_t0.bytes, iph.bytes);
		ips := inner_product(vector_subtract(aL[:], aR[:]), yN) //rct::key ips = inner_product(vector_subtract(aL, aR), yN);
		crypto.ScMulAdd(&test_t0, &z, &ips, &test_t0)           // sc_muladd(test_t0.bytes, z.bytes, ips.bytes, test_t0.bytes);
		ipt := inner_product(twoN, aL[:])                       // rct::key ipt = inner_product(twoN, aL);
		crypto.ScMulAdd(&test_t0, &zsq, &ipt, &test_t0)         // sc_muladd(test_t0.bytes, zsq.bytes, ipt.bytes, test_t0.bytes);
		crypto.ScAdd(&test_t0, &test_t0, &k)                    // sc_add(test_t0.bytes, test_t0.bytes, k.bytes);

		//CHECK_AND_ASSERT_THROW_MES(t0 == test_t0, "t0 check failed");
		if t0 != test_t0 {
			panic("t0 check failed")
		}

		//fmt.Printf("t0      %s\ntest_t0 %s\n",t0,test_t0)

	}

	// STEP 1 complete above

	// STEP 2 starts

	HyNsR := hadamard(yN, sR[:])            // const auto HyNsR = hadamard(yN, sR);
	vpIz := vector_scalar(oneN, z)          //  const auto vpIz = vector_scalar(oneN, z);
	vp2zsq := vector_scalar(twoN, zsq)      //  const auto vp2zsq = vector_scalar(twoN, zsq);
	aL_vpIz := vector_subtract(aL[:], vpIz) //  const auto aL_vpIz = vector_subtract(aL, vpIz);
	aR_vpIz := vector_add(aR[:], vpIz)      //const auto aR_vpIz = vector_add(aR, vpIz);

	ip1 := inner_product(aL_vpIz, HyNsR) // rct::key ip1 = inner_product(aL_vpIz, HyNsR);
	crypto.ScAdd(&t1, &t1, &ip1)         //   sc_add(t1.bytes, t1.bytes, ip1.bytes);

	ip2 := inner_product(sL[:], vector_add(hadamard(yN, aR_vpIz), vp2zsq)) // rct::key ip2 = inner_product(sL, vector_add(hadamard(yN, aR_vpIz), vp2zsq));
	crypto.ScAdd(&t1, &t1, &ip2)                                           // sc_add(t1.bytes, t1.bytes, ip2.bytes);

	ip3 := inner_product(sL[:], HyNsR) // rct::key ip3 = inner_product(sL, HyNsR);
	crypto.ScAdd(&t2, &t2, &ip3)       //sc_add(t2.bytes, t2.bytes, ip3.bytes);

	// PAPER LINES 47-48
	tau1 := crypto.SkGen() //   rct::key tau1 = rct::skGen(), tau2 = rct::skGen();
	tau2 := crypto.SkGen()

	// rct::key T1 = rct::addKeys(rct::scalarmultKey(rct::H, t1), rct::scalarmultBase(tau1));
	tau1_base := crypto.ScalarmultBase(tau1)
	T1 := AddKeys_return(crypto.ScalarMultKey(&crypto.H, &t1), &tau1_base)

	//rct::key T2 = rct::addKeys(rct::scalarmultKey(rct::H, t2), rct::scalarmultBase(tau2));
	tau2_base := crypto.ScalarmultBase(tau2)
	T2 := AddKeys_return(crypto.ScalarMultKey(&crypto.H, &t2), &tau2_base)

	// PAPER LINES 49-51
	x := hash_cache_mash3(&hashcache, z, T1, T2) //rct::key x = hash_cache_mash(hash_cache, z, T1, T2);

	// PAPER LINES 52-53
	taux := crypto.Zero                        // rct::key taux = rct::zero();
	crypto.ScMul(&taux, &tau1, &x)             //sc_mul(taux.bytes, tau1.bytes, x.bytes);
	var xsq crypto.Key                         //rct::key xsq;
	crypto.ScMul(&xsq, &x, &x)                 //sc_mul(xsq.bytes, x.bytes, x.bytes);
	crypto.ScMulAdd(&taux, &tau2, &xsq, &taux) // sc_muladd(taux.bytes, tau2.bytes, xsq.bytes, taux.bytes);
	crypto.ScMulAdd(&taux, gamma, &zsq, &taux) //sc_muladd(taux.bytes, gamma.bytes, zsq.bytes, taux.bytes);

	var mu crypto.Key                      //rct::key mu;
	crypto.ScMulAdd(&mu, &x, &rho, &alpha) //sc_muladd(mu.bytes, x.bytes, rho.bytes, alpha.bytes);

	// PAPER LINES 54-57
	l := vector_add(aL_vpIz, vector_scalar(sL[:], x))                                   //rct::keyV l = vector_add(aL_vpIz, vector_scalar(sL, x));
	r := vector_add(hadamard(yN, vector_add(aR_vpIz, vector_scalar(sR[:], x))), vp2zsq) // rct::keyV r = vector_add(hadamard(yN, vector_add(aR_vpIz, vector_scalar(sR, x))), vp2zsq);

	// STEP 2 complete

	// STEP 3 starts
	t := inner_product(l, r) //rct::key t = inner_product(l, r);

	//DEBUG: Test if the l and r vectors match the polynomial forms
	if DEBUGGING_MODE {
		var test_t crypto.Key

		crypto.ScMulAdd(&test_t, &t1, &x, &t0)       // sc_muladd(test_t.bytes, t1.bytes, x.bytes, t0.bytes);
		crypto.ScMulAdd(&test_t, &t2, &xsq, &test_t) //sc_muladd(test_t.bytes, t2.bytes, xsq.bytes, test_t.bytes);

		if test_t != t {
			//panic("test_t check failed")
		}

		//fmt.Printf("t      %s\ntest_t %s\n",t,test_t)
	}

	// PAPER LINES 32-33
	x_ip := hash_cache_mash4(&hashcache, x, taux, mu, t) //rct::key x_ip = hash_cache_mash(hash_cache, x, taux, mu, t);

	// These are used in the inner product rounds
	// declared in step 4 //size_t nprime = N;
	var Gprime, Hprime, aprime, bprime []crypto.Key
	Gprime = make([]crypto.Key, N, N) //rct::keyV Gprime(N);
	Hprime = make([]crypto.Key, N, N) //rct::keyV Hprime(N);
	aprime = make([]crypto.Key, N, N) // rct::keyV aprime(N);
	bprime = make([]crypto.Key, N, N) //rct::keyV bprime(N);

	yinv := invert_scalar(y)   //const rct::key yinv = invert(y);
	yinvpow := crypto.Identity //          rct::key yinvpow = rct::identity();

	for i := 0; i < N; i++ { ///for (size_t i = 0; i < N; ++i)
		Gprime[i] = Gi[i]                                     //                       Gprime[i] = Gi[i];
		Hprime[i] = *(crypto.ScalarMultKey(&Hi[i], &yinvpow)) //Hprime[i] = scalarmultKey(Hi[i], yinvpow);
		crypto.ScMul(&yinvpow, &yinvpow, &yinv)               //sc_mul(yinvpow.bytes, yinvpow.bytes, yinv.bytes);

		aprime[i] = l[i] // aprime[i] = l[i];
		bprime[i] = r[i] // bprime[i] = r[i];
	}

	// STEP 3 complete

	// STEP 4 starts
	round := 0
	nprime := N
	//var L,R,w [logN]crypto.Key  // w is the challenge x in the inner product protocol
	L := make([]crypto.Key, logN, logN)
	R := make([]crypto.Key, logN, logN)
	w := make([]crypto.Key, logN, logN)
	var tmp crypto.Key

	// PAPER LINE 13
	for nprime > 1 { // while (nprime > 1)
		// PAPER LINE 15
		nprime /= 2 // nprime /= 2;

		// PAPER LINES 16-17
		cL := inner_product(slice_vector(aprime[:], 0, nprime), slice_vector(bprime[:], nprime, len(bprime))) // rct::key cL = inner_product(slice(aprime, 0, nprime), slice(bprime, nprime, bprime.size()));
		cR := inner_product(slice_vector(aprime[:], nprime, len(aprime)), slice_vector(bprime[:], 0, nprime)) // rct::key cR = inner_product(slice(aprime, nprime, aprime.size()), slice(bprime, 0, nprime));

		// PAPER LINES 18-19
		//L[round] = vector_exponent_custom(slice(Gprime, nprime, Gprime.size()), slice(Hprime, 0, nprime), slice(aprime, 0, nprime), slice(bprime, nprime, bprime.size()));

		L[round] = vector_exponent_custom(slice_vector(Gprime[:], nprime, len(Gprime)), slice_vector(Hprime[:], 0, nprime), slice_vector(aprime[:], 0, nprime), slice_vector(bprime[:], nprime, len(bprime)))
		crypto.ScMul(&tmp, &cL, &x_ip)                                              //    sc_mul(tmp.bytes, cL.bytes, x_ip.bytes);
		crypto.AddKeys(&L[round], &L[round], crypto.ScalarMultKey(&crypto.H, &tmp)) //rct::addKeys(L[round], L[round], rct::scalarmultKey(rct::H, tmp));
		//R[round] = vector_exponent_custom(slice(Gprime, 0, nprime), slice(Hprime, nprime, Hprime.size()), slice(aprime, nprime, aprime.size()), slice(bprime, 0, nprime));
		R[round] = vector_exponent_custom(slice_vector(Gprime[:], 0, nprime), slice_vector(Hprime[:], nprime, len(Hprime)), slice_vector(aprime[:], nprime, len(aprime)), slice_vector(bprime[:], 0, nprime))
		crypto.ScMul(&tmp, &cR, &x_ip)                                              // sc_mul(tmp.bytes, cR.bytes, x_ip.bytes);
		crypto.AddKeys(&R[round], &R[round], crypto.ScalarMultKey(&crypto.H, &tmp)) // rct::addKeys(R[round], R[round], rct::scalarmultKey(rct::H, tmp));

		// PAPER LINES 21-22
		w[round] = hash_cache_mash2(&hashcache, L[round], R[round]) //   w[round] = hash_cache_mash(hash_cache, L[round], R[round]);

		// PAPER LINES 24-25
		winv := invert_scalar(w[round]) //const rct::key winv = invert(w[round]);
		//Gprime = hadamard2(vector_scalar2(slice(Gprime, 0, nprime), winv), vector_scalar2(slice(Gprime, nprime, Gprime.size()), w[round]));
		Gprime = hadamard2(vector_scalar2(slice_vector(Gprime[:], 0, nprime), winv), vector_scalar2(slice_vector(Gprime[:], nprime, len(Gprime)), w[round]))

		//Hprime = hadamard2(vector_scalar2(slice(Hprime, 0, nprime), w[round]), vector_scalar2(slice(Hprime, nprime, Hprime.size()), winv));
		Hprime = hadamard2(vector_scalar2(slice_vector(Hprime[:], 0, nprime), w[round]), vector_scalar2(slice_vector(Hprime[:], nprime, len(Hprime)), winv))

		// PAPER LINES 28-29
		//aprime = vector_add(vector_scalar(slice(aprime, 0, nprime), w[round]), vector_scalar(slice(aprime, nprime, aprime.size()), winv));
		aprime = vector_add(vector_scalar(slice_vector(aprime[:], 0, nprime), w[round]), vector_scalar(slice_vector(aprime[:], nprime, len(aprime)), winv))

		//bprime = vector_add(vector_scalar(slice(bprime, 0, nprime), winv), vector_scalar(slice(bprime, nprime, bprime.size()), w[round]));
		bprime = vector_add(vector_scalar(slice_vector(bprime[:], 0, nprime), winv), vector_scalar(slice_vector(bprime[:], nprime, len(bprime)), w[round]))

		round++

	}

	return &BulletProof{
		V:    []crypto.Key{V},
		A:    A,
		S:    S,
		T1:   T1,
		T2:   T2,
		taux: taux,
		mu:   mu,
		L:    L,
		R:    R,
		a:    aprime[0],
		b:    bprime[0],
		t:    t,
	}
}

// prove an amount
func BULLETPROOF_Prove_Amount(v uint64, gamma *crypto.Key) *BulletProof {
	sv := crypto.Zero

	sv[0] = byte(v & 255)
	sv[1] = byte((v >> 8) & 255)
	sv[2] = byte((v >> 16) & 255)
	sv[3] = byte((v >> 24) & 255)
	sv[4] = byte((v >> 32) & 255)
	sv[5] = byte((v >> 40) & 255)
	sv[6] = byte((v >> 48) & 255)
	sv[7] = byte((v >> 56) & 255)

	return BULLETPROOF_Prove(&sv, gamma)
}


func (proof *BulletProof)BULLETPROOF_BasicChecks() (result bool){
   
    // check whether any of the values in the proof are not 0 or 1 
	if proof.V[0] == crypto.Zero  ||
           proof.A ==  crypto.Zero || 
           proof.S ==  crypto.Zero || 
            proof.T1 ==  crypto.Zero || 
             proof.T2 ==  crypto.Zero || 
              proof.taux ==  crypto.Zero || 
               proof.mu ==  crypto.Zero || 
                  proof.a ==  crypto.Zero || 
                   proof.b ==  crypto.Zero ||
                    proof.t ==  crypto.Zero{
                       return false
                   }
        for  i := range proof.L {
            if  proof.L[i] ==  crypto.Zero ||  proof.R[i] ==  crypto.Zero{
                return false
            }
        }
                   
	if proof.V[0] == crypto.Identity  ||
           proof.A ==  crypto.Identity || 
           proof.S ==  crypto.Identity || 
            proof.T1 ==  crypto.Identity || 
             proof.T2 ==  crypto.Identity || 
              proof.taux ==  crypto.Identity || 
               proof.mu ==  crypto.Identity || 
                  proof.a ==  crypto.Identity || 
                   proof.b ==  crypto.Identity ||
                    proof.t ==  crypto.Identity{
                       return false
                   }

        for  i := range proof.L {
            if  proof.L[i] ==  crypto.Identity ||  proof.R[i] ==  crypto.Identity{
                return false
            }
        }                   
	
	
	// time to verify that cofactors cannnot be exploited 
	curve_order := crypto.CurveOrder()
	if *crypto.ScalarMultKey(&proof.V[0], &curve_order) != crypto.Identity {
            return false
        }
        
	if *crypto.ScalarMultKey(&proof.A, &curve_order) != crypto.Identity {
            return false
        }	
	if *crypto.ScalarMultKey(&proof.S, &curve_order) != crypto.Identity {
            return false
        }
	if *crypto.ScalarMultKey(&proof.T1, &curve_order) != crypto.Identity {
            return false
        }
	if *crypto.ScalarMultKey(&proof.T2, &curve_order) != crypto.Identity {
            return false
        }
        /* 
	if *crypto.ScalarMultKey(&proof.taux, &curve_order) != crypto.Identity {
            return false
        }
	if *crypto.ScalarMultKey(&proof.mu, &curve_order) != crypto.Identity {
            return false
        }
        */
         for  i := range proof.L {
            if *crypto.ScalarMultKey(&proof.L[i], &curve_order) != crypto.Identity {
                return false
            }
            if *crypto.ScalarMultKey(&proof.R[i], &curve_order) != crypto.Identity {
            return false
        }
        } 
        

/*	
        if *crypto.ScalarMultKey(&proof.a, &curve_order) != crypto.Identity {
            return false
        }
	if *crypto.ScalarMultKey(&proof.b, &curve_order) != crypto.Identity {
            return false
        }

	if *crypto.ScalarMultKey(&proof.t, &curve_order) != crypto.Identity {
            return false
        }
*/

   return true 
}

func (proof *BulletProof) BULLETPROOF_Verify() (result bool) {

	defer func() { // safety so if anything wrong happens, verification fails
		if r := recover(); r != nil {
			result = false
		}
	}()

	if !(len(proof.V) == 1) {
		//V does not have exactly one element
		return false
	}

	if len(proof.L) != len(proof.R) {
		//Mismatched L and R sizes
		return false
	}
	if len(proof.L) == 0 {
		// Empty Proof
		return false
	}

	if len(proof.L) != 6 {
		//Proof is not for 64 bits
		return false
	}
	
	
	// these checks try to filter out rogue inputs
	if proof.BULLETPROOF_BasicChecks() == false{
            return false
        }
    
        
		
	logN := len(proof.L)
	N := int(1 << uint(logN))

	// reconstruct the challenges
	hashcache := *(crypto.HashToScalar(proof.V[0][:]))  //rct::key hash_cache = rct::hash_to_scalar(proof.V[0]);
	y := hash_cache_mash2(&hashcache, proof.A, proof.S) //  rct::key y = hash_cache_mash(hash_cache, proof.A, proof.S);

	hashcache = *(crypto.HashToScalar(y[:])) // rct::key z = hash_cache = rct::hash_to_scalar(y);
	z := hashcache
	x := hash_cache_mash3(&hashcache, z, proof.T1, proof.T2) //rct::key x = hash_cache_mash(hash_cache, z, proof.T1, proof.T2);

	x_ip := hash_cache_mash4(&hashcache, x, proof.taux, proof.mu, proof.t) //rct::key x_ip = hash_cache_mash(hash_cache, x, proof.taux, proof.mu, proof.t);

	// PAPER LINE 61
	//rct::key L61Left = rct::addKeys(rct::scalarmultBase(proof.taux), rct::scalarmultKey(rct::H, proof.t));
	taux_base := crypto.ScalarmultBase(proof.taux)
	L61Left := AddKeys_return(&taux_base, crypto.ScalarMultKey(&crypto.H, &proof.t))

	k := crypto.Zero                 //rct::key k = rct::zero();
	yN := vector_powers(y, int64(N)) //const auto yN = vector_powers(y, N);
	ip1y := inner_product(oneN, yN)  //rct::key ip1y = inner_product(oneN, yN);
	zsq := crypto.Zero               //rct::key zsq;
	crypto.ScMul(&zsq, &z, &z)       //sc_mul(zsq.bytes, z.bytes, z.bytes);

	var tmp, tmp2 crypto.Key             //rct::key tmp, tmp2;
	crypto.ScMulSub(&k, &zsq, &ip1y, &k) //  sc_mulsub(k.bytes, zsq.bytes, ip1y.bytes, k.bytes);
	var zcu crypto.Key                   //rct::key zcu;
	crypto.ScMul(&zcu, &zsq, &z)         //sc_mul(zcu.bytes, zsq.bytes, z.bytes);
	crypto.ScMulSub(&k, &zcu, &ip12, &k) //sc_mulsub(k.bytes, zcu.bytes, ip12.bytes, k.bytes);

	crypto.ScMulAdd(&tmp, &z, &ip1y, &k)                 // sc_muladd(tmp.bytes, z.bytes, ip1y.bytes, k.bytes);
	L61Right := *(crypto.ScalarMultKey(&crypto.H, &tmp)) //rct::key L61Right = rct::scalarmultKey(rct::H, tmp);

	tmp = *(crypto.ScalarMultKey(&proof.V[0], &zsq)) //tmp = rct::scalarmultKey(proof.V[0], zsq);
	crypto.AddKeys(&L61Right, &L61Right, &tmp)       //rct::addKeys(L61Right, L61Right, tmp);

	tmp = *(crypto.ScalarMultKey(&proof.T1, &x)) // tmp = rct::scalarmultKey(proof.T1, x);
	crypto.AddKeys(&L61Right, &L61Right, &tmp)   //ct::addKeys(L61Right, L61Right, tmp);

	var xsq crypto.Key                             //rct::key xsq;
	crypto.ScMul(&xsq, &x, &x)                     // sc_mul(xsq.bytes, x.bytes, x.bytes);
	tmp = *(crypto.ScalarMultKey(&proof.T2, &xsq)) //tmp = rct::scalarmultKey(proof.T2, xsq);
	crypto.AddKeys(&L61Right, &L61Right, &tmp)     //rct::addKeys(L61Right, L61Right, tmp);

	if !(L61Right == L61Left) {
		//MERROR("Verification failure at step 1");
		// fmt.Printf("erification failure at step 1")
		return false
	}

	//fmt.Printf("Verification passed at step 1")

	// PAPER LINE 62
	P := AddKeys_return(&proof.A, crypto.ScalarMultKey(&proof.S, &x)) //rct::key P = rct::addKeys(proof.A, rct::scalarmultKey(proof.S, x));

	// Compute the number of rounds for the inner product
	rounds := len(proof.L)

	// PAPER LINES 21-22
	// The inner product challenges are computed per round
	w := make([]crypto.Key, rounds, rounds) //  rct::keyV w(rounds);
	for i := 0; i < rounds; i++ {           ///for (size_t i = 0; i < rounds; ++i)
		w[i] = hash_cache_mash2(&hashcache, proof.L[i], proof.R[i]) //w[i] = hash_cache_mash(hash_cache, proof.L[i], proof.R[i]);
	}

	// Basically PAPER LINES 24-25
	// Compute the curvepoints from G[i] and H[i]
	inner_prod := crypto.Identity // rct::key inner_prod = rct::identity();
	yinvpow := crypto.Identity    // rct::key yinvpow = rct::identity();
	ypow := crypto.Identity       // rct::key ypow = rct::identity();

	yinv := invert_scalar(y)                   //const rct::key yinv = invert(y);
	winv := make([]crypto.Key, rounds, rounds) //rct::keyV winv(rounds);
	for i := 0; i < rounds; i++ {              //for (size_t i = 0; i < rounds; ++i)
		winv[i] = invert_scalar(w[i]) //	winv[i] = invert(w[i]);
	}

	for i := 0; i < N; i++ { //for (size_t i = 0; i < N; ++i)

		// Convert the index to binary IN REVERSE and construct the scalar exponent
		g_scalar := proof.a                         //rct::key g_scalar = proof.a;
		h_scalar := crypto.Zero                     // rct::key h_scalar;
		crypto.ScMul(&h_scalar, &proof.b, &yinvpow) //sc_mul(h_scalar.bytes, proof.b.bytes, yinvpow.bytes);

		// is this okay ???
		for j := rounds; j > 0; { // for (size_t j = rounds; j-- > 0; )
			j--
			// FIXME below len can be ommitted and represents rounds
			J := len(w) - j - 1 //size_t J = w.size() - j - 1;

			if i&((1)<<uint(j)) == 0 { /////if ((i & (((size_t)1)<<j)) == 0)
				crypto.ScMul(&g_scalar, &g_scalar, &winv[J]) //sc_mul(g_scalar.bytes, g_scalar.bytes, winv[J].bytes);
				crypto.ScMul(&h_scalar, &h_scalar, &w[J])    // sc_mul(h_scalar.bytes, h_scalar.bytes, w[J].bytes);
			} else {
				crypto.ScMul(&g_scalar, &g_scalar, &w[J])    //sc_mul(g_scalar.bytes, g_scalar.bytes, w[J].bytes);
				crypto.ScMul(&h_scalar, &h_scalar, &winv[J]) //sc_mul(h_scalar.bytes, h_scalar.bytes, winv[J].bytes);
			}
		}

		// Adjust the scalars using the exponents from PAPER LINE 62
		crypto.ScAdd(&g_scalar, &g_scalar, &z)                // sc_add(g_scalar.bytes, g_scalar.bytes, z.bytes);
		crypto.ScMul(&tmp, &zsq, &twoN[i])                    //sc_mul(tmp.bytes, zsq.bytes, twoN[i].bytes);
		crypto.ScMulAdd(&tmp, &z, &ypow, &tmp)                //sc_muladd(tmp.bytes, z.bytes, ypow.bytes, tmp.bytes);
		crypto.ScMulSub(&h_scalar, &tmp, &yinvpow, &h_scalar) //  sc_mulsub(h_scalar.bytes, tmp.bytes, yinvpow.bytes, h_scalar.bytes);

		// Now compute the basepoint's scalar multiplication
		// Each of these could be written as a multiexp operation instead
		// cross-check this line again
		// TODO can be a major  performance improvement
		// TODO maybe this can be used https://boringssl.googlesource.com/boringssl/+/2357/crypto/ec/wnaf.c
		// https://github.com/bitcoin-core/secp256k1/blob/master/src/ecmult_impl.h
		crypto.AddKeys3_3(&tmp, &g_scalar, &Gi_Precomputed[i], &h_scalar, &Hi_Precomputed[i]) //rct::addKeys3(tmp, g_scalar, Gprecomp[i], h_scalar, Hprecomp[i]);
		crypto.AddKeys(&inner_prod, &inner_prod, &tmp)                                        //rct::addKeys(inner_prod, inner_prod, tmp);

		if i != N-1 {
			crypto.ScMul(&yinvpow, &yinvpow, &yinv) //sc_mul(yinvpow.bytes, yinvpow.bytes, yinv.bytes);
			crypto.ScMul(&ypow, &ypow, &y)          //sc_mul(ypow.bytes, ypow.bytes, y.bytes);
		}
	}

	//fmt.Printf("inner prod original %s\n",inner_prod)

	// PAPER LINE 26
	var pprime crypto.Key                       //rct::key pprime;
	crypto.ScSub(&tmp, &crypto.Zero, &proof.mu) //sc_sub(tmp.bytes, rct::zero().bytes, proof.mu.bytes);

	tmp_base := crypto.ScalarmultBase(tmp)
	crypto.AddKeys(&pprime, &P, &tmp_base) //rct::addKeys(pprime, P, rct::scalarmultBase(tmp));

	for i := 0; i < rounds; i++ { //for (size_t i = 0; i < rounds; ++i)

		crypto.ScMul(&tmp, &w[i], &w[i])        //sc_mul(tmp.bytes, w[i].bytes, w[i].bytes);
		crypto.ScMul(&tmp2, &winv[i], &winv[i]) //sc_mul(tmp2.bytes, winv[i].bytes, winv[i].bytes);
		//#if 1
		// ge_dsmp cacheL, cacheR;
		// rct::precomp(cacheL, proof.L[i]);
		//rct::precomp(cacheR, proof.R[i]);

		ProofLi := new(crypto.ExtendedGroupElement)
		ProofLi.FromBytes(&proof.L[i])

		ProofRi := new(crypto.ExtendedGroupElement)
		ProofRi.FromBytes(&proof.R[i])

		var ProofLi_precomputed [8]crypto.CachedGroupElement // A,3A,5A,7A,9A,11A,13A,15A
		crypto.GePrecompute(&ProofLi_precomputed, ProofLi)

		var ProofRi_precomputed [8]crypto.CachedGroupElement // A,3A,5A,7A,9A,11A,13A,15A
		crypto.GePrecompute(&ProofRi_precomputed, ProofRi)

		// optimise these at the end only if possible
		crypto.AddKeys3_3(&tmp, &tmp, &ProofLi_precomputed, &tmp2, &ProofRi_precomputed) //rct::addKeys3(tmp, tmp, cacheL, tmp2, cacheR);
		crypto.AddKeys(&pprime, &pprime, &tmp)                                           //rct::addKeys(pprime, pprime, tmp);

		//#endif
	}

	crypto.ScMul(&tmp, &proof.t, &x_ip)                                     // sc_mul(tmp.bytes, proof.t.bytes, x_ip.bytes);
	crypto.AddKeys(&pprime, &pprime, crypto.ScalarMultKey(&crypto.H, &tmp)) //rct::addKeys(pprime, pprime, rct::scalarmultKey(rct::H, tmp));

	crypto.ScMul(&tmp, &proof.a, &proof.b)         //  sc_mul(tmp.bytes, proof.a.bytes, proof.b.bytes);
	crypto.ScMul(&tmp, &tmp, &x_ip)                // sc_mul(tmp.bytes, tmp.bytes, x_ip.bytes);
	tmp = *(crypto.ScalarMultKey(&crypto.H, &tmp)) //tmp = rct::scalarmultKey(rct::H, tmp);
	crypto.AddKeys(&tmp, &tmp, &inner_prod)        //rct::addKeys(tmp, tmp, inner_prod);

	if !(pprime == tmp) {
		// MERROR("Verification failure at step 2");
		// fmt.Printf("Verification failure at step 2");
		return false
	}

	//fmt.Printf("\n prime      %s\n tmp %s  bulletproof verified successfully\n", pprime, tmp)

	return true
}

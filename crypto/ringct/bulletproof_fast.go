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

//import "math/big"
//import "encoding/binary"
import "sync"

import "github.com/deroproject/derosuite/crypto"

var HiS, GiS [maxN]crypto.SUPER_PRECOMPUTE_TABLE
var once sync.Once

// this should be called after Hi, Gi are setup
func precompute_tables() {
	fmt.Sprintf("junk")
	for i := 0; i < maxN; i++ {
		var tmp crypto.PRECOMPUTE_TABLE

		crypto.GenPrecompute(&tmp, Hi[i])        // generate pre comp table
		crypto.GenSuperPrecompute(&HiS[i], &tmp) // generate His precomp table

		crypto.GenPrecompute(&tmp, Gi[i])        // generate pre comp table
		crypto.GenSuperPrecompute(&GiS[i], &tmp) // generate His precomp table

	}
}

// see the  references such as original paper and multiple implementations
// https://eprint.iacr.org/2017/1066.pdf
// https://blog.chain.com/faster-bulletproofs-with-ristretto-avx2-29450b4490cd

func (proof *BulletProof) BULLETPROOF_Verify_fast() (result bool) {

	defer func() { // safety so if anything wrong happens, verification fails
		if r := recover(); r != nil {
			result = false
		}
	}()

	once.Do(precompute_tables) // generate pre compute tables

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

	var intermediate_inner_prod crypto.ExtendedGroupElement
	intermediate_inner_prod.Zero() // make identity

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
		//crypto.AddKeys3_3(&tmp, &g_scalar, &Gi_Precomputed[i], &h_scalar, &Hi_Precomputed[i]) //rct::addKeys3(tmp, g_scalar, Gprecomp[i], h_scalar, Hprecomp[i]);
		//crypto.AddKeys(&inner_prod, &inner_prod, &tmp)                                        //rct::addKeys(inner_prod, inner_prod, tmp);

		var first, second, scratch crypto.ExtendedGroupElement
		var cached crypto.CachedGroupElement
		var c crypto.CompletedGroupElement
		crypto.ScalarMultSuperPrecompute(&first, &g_scalar, &GiS[i])

		crypto.ScalarMultSuperPrecompute(&second, &h_scalar, &HiS[i])

		second.ToCached(&cached)
		crypto.GeAdd(&c, &first, &cached) // add both points together

		c.ToExtended(&scratch)
		scratch.ToCached(&cached)
		crypto.GeAdd(&c, &intermediate_inner_prod, &cached) //add with intermediate resule
		c.ToExtended(&intermediate_inner_prod)

		if i != N-1 {
			crypto.ScMul(&yinvpow, &yinvpow, &yinv) //sc_mul(yinvpow.bytes, yinvpow.bytes, yinv.bytes);
			crypto.ScMul(&ypow, &ypow, &y)          //sc_mul(ypow.bytes, ypow.bytes, y.bytes);
		}
	}

	intermediate_inner_prod.ToBytes(&inner_prod)
	//inner_prod = crypto.Multiscalarmult_compatibility(scalars,points)
	//inner_prod = crypto.Multiscalarmult(scalars,points)

	// fmt.Printf("inner prod fast %s\n",inner_prod)

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

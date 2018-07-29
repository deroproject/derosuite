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

//import "sync"
//import "sync/atomic"

import "github.com/deroproject/derosuite/crypto"

/* this files handles the generation and verification in ringct simple */

// NOTE the transaction must have been expanded earlier and must have a key image, mixring etc
// this is implementation of verRctMG from rctSigs.cpp file
func (r *RctSig) VerifyRCTSimple_Core() (result bool) {

	result = false
	if !(r.sigType == RCTTypeSimple || r.sigType == RCTTypeSimpleBulletproof) {
		if DEBUGGING_MODE {
			fmt.Printf("Signature NOT RingCT Simple or bulletproof  type, verification failed\n")
		}
		result = false
		return
	}

	pre_mlsag_hash := crypto.Key(Get_pre_mlsag_hash(r))

	// loop through all the inputs
	for inputi := 0; inputi < len(r.pseudoOuts); inputi++ {

		rows := 1
		cols := len(r.MixRing[inputi])

		if cols <= 2 {
			result = false
		}

		M := make([][]crypto.Key, cols) // lets create the double dimensional array
		for i := 0; i < cols; i++ {
			M[i] = make([]crypto.Key, rows+1, rows+1)
		}

		//create the matrix to mg sig
		for i := 0; i < cols; i++ {
			M[i][0] = r.MixRing[inputi][i].Destination
			crypto.SubKeys(&M[i][1], &r.MixRing[inputi][i].Mask, &r.pseudoOuts[inputi])
		}

		// do the mlsag verification
		result = MLSAG_Ver(pre_mlsag_hash, M, &r.MlsagSigs[inputi], rows, r)

		if result == false { // verification of 1 one vin failed mark, entire TX as failed
			if DEBUGGING_MODE {
				fmt.Printf("RCT Simple  signature verification failed for input %d\n", inputi)
			}

			return
		}

	}

	// we are here means everything went smoothly
	if DEBUGGING_MODE {
		fmt.Printf(" RCT Simple Signature successfully verified\n")
	}
	// result is already true so

	return
}

// structure using which information is fed from wallet
// these are only used while proving ringct simple
type Input_info struct {
	Amount       uint64      // amount in this input
	Key_image    crypto.Hash // keyimage in this input
	Index        int         // index position within ring members
	Index_Global uint64
	Ring_Members []uint64 // ring members  already sorted absolute
	Pubs         []CtKey  // public keys from ring members ( secret key from our input)
	Sk           CtKey    // secret key for the input
}

type Output_info struct {
	Amount           uint64     // only first output is locked
	Public_View_Key  crypto.Key // taken from address
	Public_Spend_Key crypto.Key // taken from address
	Scalar_Key       crypto.Key // used to encrypt amounts
	// Destination crypto.Key
	//Addr  address.Address
}

// this will prove  ringct signature
// message is the tx prefix hash
// inputs contain data of each and every input, together with the ring members and other data
// outputs contains the output amount and key to encode the amount
// fees is the fees to provide
// this function is equivalent to genRctSimple in rctSigs.cpp
func (r *RctSig) Gen_RingCT_Simple(Message crypto.Hash, inputs []Input_info, outputs []Output_info, fees uint64) {

	r.sigType = RCTTypeSimple
	r.Message = crypto.Key(Message)
	r.txFee = fees

	var sumouts, sumpouts crypto.Key

	for i := range outputs {
		var public_mask, secret_mask crypto.Key
		r.rangeSigs = append(r.rangeSigs, *(ProveRange(&public_mask, &secret_mask, outputs[i].Amount)))

		//fmt.Printf("SK %s\n", secret_mask )
		//fmt.Printf("PK %s\n", public_mask )
		crypto.ScAdd(&sumouts, &sumouts, &secret_mask)
		// copy public mask to outpk
		r.OutPk = append(r.OutPk, CtKey{Mask: public_mask})

		// create tuple and encrypt it, then add it the signature
		tuple := ECdhTuple{Mask: secret_mask, Amount: *d2h(outputs[i].Amount)}
		ecdhEncode(&tuple, outputs[i].Scalar_Key)

		r.ECdhInfo = append(r.ECdhInfo, tuple) // add encrypted tuple to signature
	}

	a := make([]crypto.Key, len(inputs), len(inputs))
	r.pseudoOuts = make([]crypto.Key, len(inputs), len(inputs))
	// generate pseudoOuts so as any one can verify that sum(inputs) = sum(outputs)
	for i := 0; i < len(inputs)-1; i++ { // we need to adjust fees in the last one
		a[i] = crypto.SkGen() // generate random key

		// Sc_0(&a[i]);  // temporary for debugging puprpose, make it zero

		crypto.ScReduce32(&a[i]) // reduce it for crypto purpose
		crypto.ScAdd(&sumpouts, &a[i], &sumpouts)
		genC(&r.pseudoOuts[i], &a[i], inputs[i].Amount)
	}

	//  fmt.Printf("input len %d\n", len(inputs))
	crypto.ScSub(&a[len(inputs)-1], &sumouts, &sumpouts)
	genC(&r.pseudoOuts[len(inputs)-1], &a[len(inputs)-1], inputs[len(inputs)-1].Amount)

	// fmt.Printf("RCT range signature verification status %v\n", r.VerifyRctSimple())

	message := crypto.Key(Get_pre_mlsag_hash(r))
	for i := range inputs {

		r.MlsagSigs = append(r.MlsagSigs, proveRctMGSimple(message, inputs[i].Pubs, inputs[i].Sk, a[i], r.pseudoOuts[i], inputs[i].Index))

		r.MixRing = append(r.MixRing, inputs[i].Pubs) // setup mixring for temoprary validation
		r.MlsagSigs[i].II = make([]crypto.Key, 1, 1)
		r.MlsagSigs[i].II[0] = crypto.Key(inputs[i].Key_image)

		// fmt.Printf("rv.sigs size %d  \n", len(r.MlsagSigs[i].ss))

	}

	// temporary setup mixrings so tx can be verified h
	//  fmt.Printf("ringct verified %+v\n ",r.VerifyRCTSimple_Core())

}

// this will prove  ringct signature using bullet proof ranges
// message is the tx prefix hash
// inputs contain data of each and every input, together with the ring members and other data
// outputs contains the output amount and key to encode the amount
// fees is the fees to provide
// this function is equivalent to genRctSimple in rctSigs.cpp
func (r *RctSig) Gen_RingCT_Simple_BulletProof(Message crypto.Hash, inputs []Input_info, outputs []Output_info, fees uint64) {

	r.sigType = RCTTypeSimpleBulletproof
	r.Message = crypto.Key(Message)
	r.txFee = fees

	var sumouts, sumpouts crypto.Key

	for i := range outputs {
		var public_mask, secret_mask crypto.Key
		var public_maskc, secret_maskc crypto.Key

		r.BulletSigs = append(r.BulletSigs, ProveRangeBulletproof(&public_maskc, &secret_maskc, outputs[i].Amount))

		public_mask = crypto.Key(public_maskc)
		secret_mask = crypto.Key(secret_maskc)
		//fmt.Printf("SK %s\n", secret_mask )
		//fmt.Printf("PK %s\n", public_mask )
		crypto.ScAdd(&sumouts, &sumouts, &secret_mask)
		// copy public mask to outpk
		r.OutPk = append(r.OutPk, CtKey{Mask: public_mask})

		// create tuple and encrypt it, then add it the signature
		tuple := ECdhTuple{Mask: secret_mask, Amount: *d2h(outputs[i].Amount)}
		ecdhEncode(&tuple, outputs[i].Scalar_Key)

		r.ECdhInfo = append(r.ECdhInfo, tuple) // add encrypted tuple to signature
	}

	a := make([]crypto.Key, len(inputs), len(inputs))
	r.pseudoOuts = make([]crypto.Key, len(inputs), len(inputs))
	// generate pseudoOuts so as any one can verify that sum(inputs) = sum(outputs)
	for i := 0; i < len(inputs)-1; i++ { // we need to adjust fees in the last one
		a[i] = crypto.SkGen() // generate random key

		// Sc_0(&a[i]);  // temporary for debugging puprpose, make it zero

		crypto.ScReduce32(&a[i]) // reduce it for crypto purpose
		crypto.ScAdd(&sumpouts, &a[i], &sumpouts)
		genC(&r.pseudoOuts[i], &a[i], inputs[i].Amount)
	}

	//  fmt.Printf("input len %d\n", len(inputs))
	crypto.ScSub(&a[len(inputs)-1], &sumouts, &sumpouts)
	genC(&r.pseudoOuts[len(inputs)-1], &a[len(inputs)-1], inputs[len(inputs)-1].Amount)

	// fmt.Printf("RCT range signature verification status %v\n", r.VerifyRctSimple())

	message := crypto.Key(Get_pre_mlsag_hash(r))
	for i := range inputs {

		r.MlsagSigs = append(r.MlsagSigs, proveRctMGSimple(message, inputs[i].Pubs, inputs[i].Sk, a[i], r.pseudoOuts[i], inputs[i].Index))

		r.MixRing = append(r.MixRing, inputs[i].Pubs) // setup mixring for temoprary validation
		r.MlsagSigs[i].II = make([]crypto.Key, 1, 1)
		r.MlsagSigs[i].II[0] = crypto.Key(inputs[i].Key_image)

		// fmt.Printf("rv.sigs size %d  \n", len(r.MlsagSigs[i].ss))

	}

	// temporary setup mixrings so tx can be verified h
	//  fmt.Printf("ringct verified %+v\n ",r.VerifyRCTSimple_Core())

}

//Ring-ct MG sigs Simple
//   Simple version for when we assume only
//       post rct inputs
//       here pubs is a vector of (P, C) length mixin
//   inSk is x, a_in corresponding to signing index from the inputs
//       a_out, Cout is for the output commitment
//       index is the signing index..
func proveRctMGSimple(message crypto.Key, pubs []CtKey, inSk CtKey, a crypto.Key, Cout crypto.Key, index int) (msig MlsagSig) {
	rows := 1
	cols := len(pubs)

	if len(pubs) < 1 {
		panic("Pubs are empty")
	}

	//tmp := make([]Key, rows+1, rows+1)
	sk := make([]crypto.Key, rows+1, rows+1)

	// next 5 lines are quite common
	M := make([][]crypto.Key, cols)
	for i := 0; i < (cols); i++ {
		M[i] = make([]crypto.Key, rows+1, rows+1)
		for j := 0; j < (rows + 1); j++ { // yes there is an extra column
			M[i][j] = Identity // fill it with identity
		}
	}

	for i := 0; i < (cols); i++ {
		M[i][0] = pubs[i].Destination
		crypto.SubKeys(&M[i][1], &pubs[i].Mask, &Cout)
		sk[0] = inSk.Destination // these 2 lines can be moved out of loop, but original version implemented here
		crypto.ScSub(&sk[1], &inSk.Mask, &a)

	}

	// call mlsag gen
	return MLSAG_Gen(message, M, sk, index, rows)
}

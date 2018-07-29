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

import "io"
import "fmt"
import "bytes"
import "sync"
import "sync/atomic"

import "github.com/deroproject/derosuite/crypto"

// enable debuggin mode within ringct
// if true debugging mode enabled
const DEBUGGING_MODE = false

// XMR testnet bulletproofs are serialized  differently than how we do it
// XMR pseudoouts are marked as prunable and serialized at the end
// while we consider them as base
// below switch enables compatibility for developments purposes and should be turned
// false after development is finished
// some automated tests will automatically enable it temporarily
var XMR_COMPATIBILITY bool = false

func setcompatibility(value bool) {
	fmt.Printf("compatibility %+v", value)
}

// TODO this package need serious love of atleast few weeks
// but atleast the parser and serdes works
// we neeed to expand everthing so as chances of a bug slippping in becomes very low
// NOTE:DO NOT waste time implmenting pre-RCT code

const (
	RCTTypeNull = iota
	RCTTypeFull //  we do generate these but they are accepted
	RCTTypeSimple
	RCTTypeFullBulletproof // we DO NOT parse/support/generate these
	RCTTypeSimpleBulletproof
)

// Pedersen Commitment is generated from this struct
// C = aG + bH where a = mask and b = amount
// senderPk is the one-time public key for ECDH exchange
type ECdhTuple struct {
	Mask   crypto.Key `msgpack:"M"`
	Amount crypto.Key `msgpack:"A"`
	//	senderPk Key
}

// Range proof commitments
type Key64 [64]crypto.Key // for Borromean

// Range Signature
// Essentially data for a Borromean Signature
type RangeSig struct {
	asig BoroSig
	ci   Key64
}

// Borromean Signature
type BoroSig struct {
	s0 Key64
	s1 Key64
	ee crypto.Key
}

// size of  single bullet proof range
// we are currently only implementing non-aggregate version only
// as aggregate have following benefits and disadvntages
// 1) they are logarithmic in size but verification is linear, thus aggregate version may make it very easy to DOD
// 2) they can only be used for 2^n outputs  not for randon n
// 3) are very optimised and speedy to verify
// serialised size ((2*6 + 4 + 5)*32 + 3) * n_outputs;
type BulletProof struct {
	V []crypto.Key // 1 * 32           // extra 1 byte for length

	// 4
	A  crypto.Key // 1 * 32
	S  crypto.Key // 1 * 32
	T1 crypto.Key // 1 * 32
	T2 crypto.Key // 1 * 32

	// final 2/5
	taux crypto.Key // 1 * 32
	mu   crypto.Key // 1 * 32

	// 2*6
	L []crypto.Key // 6 * 32  // space requirements while serializing, extra 1 byte for length
	R []crypto.Key // 6 * 32 // space requirements while serializing, extra 1 byte for length

	// final 3/5
	a crypto.Key // 1 * 32
	b crypto.Key // 1 * 32
	t crypto.Key // 1 * 32
}

// MLSAG (Multilayered Linkable Spontaneous Anonymous Group) Signature
type MlsagSig struct {
	ss [][]crypto.Key
	cc crypto.Key   // this stores the starting point
	II []crypto.Key // this stores the keyimage, but is taken from the tx/blockchain,it is NOT serialized
}

// Confidential Transaction Keys, mask is Pedersen Commitment
// most of the time, it holds public keys, except (transaction making ) where it holds private keys
type CtKey struct {
	Destination crypto.Key `msgpack:"D"` // this is the destination and needs to expanded from blockchain
	Mask        crypto.Key `msgpack:"M"` // this is the public key amount/commitment homomorphic mask
}

// Ring Confidential Signature parts that we have to keep
type RctSigBase struct {
	sigType    uint8
	Message    crypto.Key // transaction prefix hash
	MixRing    [][]CtKey  // this is not serialized
	pseudoOuts []crypto.Key
	ECdhInfo   []ECdhTuple
	OutPk      []CtKey // only mask amount is serialized
	txFee      uint64

	Txid crypto.Hash // this field is extra and only used for logging purposes to track which txid was at fault
}

// Ring Confidential Signature parts that we can just prune later
type RctSigPrunable struct {
	rangeSigs  []RangeSig    //borrowmean range proof
	BulletSigs []BulletProof // bulletproofs range proofs
	MlsagSigs  []MlsagSig    // there can be as many mlsagsigs as many vins
}

// Ring Confidential Signature struct that can verify everything
type RctSig struct {
	RctSigBase
	RctSigPrunable
}

func (k *Key64) Serialize() (result []byte) {
	for _, key := range k {
		result = append(result, key[:]...)
	}
	return
}

func (b *BoroSig) Serialize() (result []byte) {
	result = append(b.s0.Serialize(), b.s1.Serialize()...)
	result = append(result, b.ee[:]...)
	return
}

func (r *RangeSig) Serialize() (result []byte) {
	result = append(r.asig.Serialize(), r.ci.Serialize()...)
	return
}

func (m *MlsagSig) Serialize() (result []byte) {
	for i := 0; i < len(m.ss); i++ {
		for j := 0; j < len(m.ss[i]); j++ {
			result = append(result, m.ss[i][j][:]...)
		}
	}
	result = append(result, m.cc[:]...)
	return
}

func (r *RctSigBase) SerializeBase() (result []byte) {
	result = []byte{r.sigType}
	// Null type returns right away
	if r.sigType == RCTTypeNull {
		return
	}
	result = append(result, Uint64ToBytes(r.txFee)...)
	if r.sigType == RCTTypeSimple {
		for _, input := range r.pseudoOuts {
			result = append(result, input[:]...)
		}
	}

	if XMR_COMPATIBILITY == false { // we consider pseudo outs are always part of base
		if r.sigType == RCTTypeSimpleBulletproof {
			for _, input := range r.pseudoOuts {
				result = append(result, input[:]...)
			}
		}
	}

	for _, ecdh := range r.ECdhInfo {
		result = append(result, ecdh.Mask[:]...)
		result = append(result, ecdh.Amount[:]...)
	}
	for _, ctKey := range r.OutPk {
		result = append(result, ctKey.Mask[:]...)
	}
	return
}

func (r *RctSigBase) BaseHash() (result crypto.Hash) {
	result = crypto.Keccak256(r.SerializeBase())
	return
}

func (r *RctSig) SerializePrunable() (result []byte) {
	if r.sigType == RCTTypeNull {
		return
	}
	for _, rangeSig := range r.rangeSigs {
		result = append(result, rangeSig.Serialize()...)
	}
	for _, bp := range r.BulletSigs {
		result = append(result, bp.Serialize()...)
	}
	for _, mlsagSig := range r.MlsagSigs {
		result = append(result, mlsagSig.Serialize()...)
	}

	if XMR_COMPATIBILITY == true {
		// XMR pseudoouts are serialized differently  and considered prunable and at the end
		if r.sigType == RCTTypeSimpleBulletproof {
			for _, input := range r.pseudoOuts {
				result = append(result, input[:]...)
			}
		}
	}
	return
}

func (r *RctSig) Get_Sig_Type() byte {
	return r.sigType
}

func (r *RctSig) Get_TX_Fee() (result uint64) {
	if r.sigType == RCTTypeNull {
		panic("RCTTypeNull cannot have TX fee")
	}
	return r.txFee
}

func (r *RctSig) PrunableHash() (result crypto.Hash) {
	if r.sigType == RCTTypeNull {
		return
	}
	result = crypto.Keccak256(r.SerializePrunable())
	return
}

// this is the function which should be used by external world
// if any exceptions occur while handling, we simply return false
// transaction must be expanded before verification
// coinbase transactions are always success, since they are tied to PoW of block
func (r *RctSig) Verify() (result bool) {

	result = false
	defer func() { // safety so if anything wrong happens, verification fails
		if r := recover(); r != nil {
			//connection.logger.Fatalf("Recovered while Verify transaction", r)
			fmt.Printf("Recovered while Verify transaction")
			result = false
		}
	}()

	switch r.sigType {
	case RCTTypeNull:
		return false /// this is only possible for miner tx
	case RCTTypeFull:
		return r.VerifyRctFull()
	case RCTTypeSimple:
		return r.VerifyRctSimple()
	case RCTTypeFullBulletproof:
		return false // these TX are NOT supported
	case RCTTypeSimpleBulletproof:
		return r.VerifyRctSimpleBulletProof()

	default:
		return false
	}

	// can never reach here
	//return false
}

// Verify a RCTTypeSimple RingCT Signature
func (r *RctSig) VerifyRctSimple() bool {
	sumOutPks := identity()
	for _, ctKey := range r.OutPk {
		crypto.AddKeys(sumOutPks, sumOutPks, &ctKey.Mask)
	}
	//txFeeKey := ScalarMultH(d2h(r.txFee))
	txFeeKey := Commitment_From_Amount(r.txFee)
	crypto.AddKeys(sumOutPks, sumOutPks, &txFeeKey)
	sumPseudoOuts := identity()
	for _, pseudoOut := range r.pseudoOuts {
		crypto.AddKeys(sumPseudoOuts, sumPseudoOuts, &pseudoOut)
	}
	if *sumPseudoOuts != *sumOutPks {
		return false
	}

	/*
		// verify range single threaded
		for i, ctKey := range r.OutPk {
			if !VerifyRange(&ctKey.Mask, r.rangeSigs[i]) {
				return false
			}
		}
	*/

	// verify borrowmean range  in parallelised form
	fail_count := uint64(0)
	wg := sync.WaitGroup{}
	wg.Add(len(r.OutPk))

	for i, ctKey := range r.OutPk {
		go func(index int, ckey CtKey) {
			if !VerifyRange(&ckey.Mask, r.rangeSigs[index]) {
				atomic.AddUint64(&fail_count, 1) // increase fail count by 1
			}
			wg.Done()
		}(i, ctKey)
	}
	wg.Wait()
	if fail_count > 0 { // check the result
		return false
	}

	if len(r.pseudoOuts) != len(r.MlsagSigs) { //if the signatures are partial reject
		return false
	}

	return r.VerifyRCTSimple_Core()
}

// Verify a RCTTypeSimple RingCT Signature
func (r *RctSig) VerifyRctSimpleBulletProof() bool {
	sumOutPks := identity()
	for _, ctKey := range r.OutPk {
		crypto.AddKeys(sumOutPks, sumOutPks, &ctKey.Mask)
	}
	//txFeeKey := ScalarMultH(d2h(r.txFee))
	txFeeKey := Commitment_From_Amount(r.txFee)
	crypto.AddKeys(sumOutPks, sumOutPks, &txFeeKey)
	sumPseudoOuts := identity()
	for _, pseudoOut := range r.pseudoOuts {
		crypto.AddKeys(sumPseudoOuts, sumPseudoOuts, &pseudoOut)
	}
	if *sumPseudoOuts != *sumOutPks {
		return false
	}

	if len(r.pseudoOuts) != len(r.MlsagSigs) { //if the signatures are partial reject
		return false
	}

	/* verify bulletproof in single threaded */
	for i, _ := range r.OutPk {
		r.BulletSigs[i].V = []crypto.Key{crypto.Key(r.OutPk[i].Mask)}
		//if !r.BulletSigs[i].BULLETPROOF_Verify() {
		if !r.BulletSigs[i].BULLETPROOF_Verify_ultrafast() {
			return false
		}
	}

	/*
		// verify range  in parallelised form
		fail_count := int32(0)
		wg := sync.WaitGroup{}
		wg.Add(len(r.OutPk))

		for i, _ := range r.OutPk {
			r.BulletSigs[i].V = []crypto.Key{crypto.Key(r.OutPk[i].Mask)}
		}
		for i, _ := range r.OutPk {
			go func(index int){
				if !r.BulletSigs[index].BULLETPROOF_Verify() {
				//if !r.BulletSigs[index].BULLETPROOF_Verify_ultrafast() {
					atomic.AddInt32(&fail_count, 1) // increase fail count by 1
				}
				wg.Done()
			}(i)
		}
		wg.Wait()
		if fail_count > 0 { // check the result
				return false
		}
	*/

	return r.VerifyRCTSimple_Core()
}
func (r *RctSig) VerifyRctFull() bool {
	for i, ctKey := range r.OutPk {
		if !VerifyRange(&ctKey.Mask, r.rangeSigs[i]) {
			return false
		}
	}

	if 1 != len(r.MlsagSigs) { //if the signatures are partial reject
		return false
	}

	return r.VerifyRCTFull_Core()
}

func ParseCtKey(buf io.Reader) (result CtKey, err error) {
	if result.Mask, err = crypto.ParseKey(buf); err != nil {
		return
	}
	return
}

func ParseKey64(buf io.Reader) (result Key64, err error) {
	for i := 0; i < 64; i++ {
		if result[i], err = crypto.ParseKey(buf); err != nil {
			return
		}
	}
	return
}

// parse Borromean signature
func ParseBoroSig(buf io.Reader) (result BoroSig, err error) {
	if result.s0, err = ParseKey64(buf); err != nil {
		return
	}
	if result.s1, err = ParseKey64(buf); err != nil {
		return
	}
	if result.ee, err = crypto.ParseKey(buf); err != nil {
		return
	}
	return
}

// range data consists of Single Borromean sig and 64 keys for 64 bits
func ParseRangeSig(buf io.Reader) (result RangeSig, err error) {
	if result.asig, err = ParseBoroSig(buf); err != nil {
		return
	}
	if result.ci, err = ParseKey64(buf); err != nil {
		return
	}
	return
}

// parser for ringct signature
// we need to be extra cautious as almost anything cam come as input
func ParseRingCtSignature(buf io.Reader, nInputs, nOutputs, nMixin int) (result *RctSig, err error) {
	r := new(RctSig)
	sigType := make([]byte, 1)
	_, err = buf.Read(sigType)
	if err != nil {
		return
	}
	r.sigType = uint8(sigType[0])
	if r.sigType == RCTTypeNull {
		result = r
		return
	}

	/* This triggers go vet saying suspect OR
	         if (r.sigType != RCTTypeFull) || (r.sigType != RCTTypeSimple) {
			err = fmt.Errorf("Bad signature Type %d", r.sigType)
	                return
		}*/

	switch r.sigType {
	case RCTTypeFull:
	case RCTTypeSimple:
	case RCTTypeSimpleBulletproof:

	case RCTTypeFullBulletproof:
		err = fmt.Errorf("Bad signature Type %d", r.sigType)
		return
	default:
		err = fmt.Errorf("Bad signature Type %d", r.sigType)
		return

	}

	r.txFee, err = ReadVarInt(buf)
	if err != nil {
		return
	}
	var nMg, nSS int

	// pseudoouts for bulletproofs are serialised at the end
	if r.sigType == RCTTypeSimple || r.sigType == RCTTypeSimpleBulletproof {
		nMg = nInputs
		nSS = 2
		r.pseudoOuts = make([]crypto.Key, nInputs)

		if r.sigType == RCTTypeSimple {
			for i := 0; i < nInputs; i++ {
				if r.pseudoOuts[i], err = crypto.ParseKey(buf); err != nil {
					return
				}
			}
		}

		if XMR_COMPATIBILITY == false { // parse our serialized
			if r.sigType == RCTTypeSimpleBulletproof {
				for i := 0; i < nInputs; i++ {
					if r.pseudoOuts[i], err = crypto.ParseKey(buf); err != nil {
						return
					}
				}
			}
		}
	} else {
		nMg = 1
		nSS = nInputs + 1
	}
	r.ECdhInfo = make([]ECdhTuple, nOutputs)
	for i := 0; i < nOutputs; i++ {
		if r.ECdhInfo[i].Mask, err = crypto.ParseKey(buf); err != nil {
			return
		}
		if r.ECdhInfo[i].Amount, err = crypto.ParseKey(buf); err != nil {
			return
		}
	}
	r.OutPk = make([]CtKey, nOutputs)
	for i := 0; i < nOutputs; i++ {
		if r.OutPk[i], err = ParseCtKey(buf); err != nil {
			return
		}
	}

	switch r.sigType {
	case RCTTypeFull, RCTTypeSimple:
		r.rangeSigs = make([]RangeSig, nOutputs)
		for i := 0; i < nOutputs; i++ {
			if r.rangeSigs[i], err = ParseRangeSig(buf); err != nil {
				return
			}
		}

	case RCTTypeFullBulletproof, RCTTypeSimpleBulletproof:
		r.BulletSigs = make([]BulletProof, nOutputs)
		for i := 0; i < nOutputs; i++ {
			if r.BulletSigs[i], err = ParseBulletProof(buf); err != nil {
				return
			}

		}
	}

	r.MlsagSigs = make([]MlsagSig, nMg)
	for i := 0; i < nMg; i++ {
		r.MlsagSigs[i].ss = make([][]crypto.Key, nMixin+1)
		for j := 0; j < nMixin+1; j++ {
			r.MlsagSigs[i].ss[j] = make([]crypto.Key, nSS)
			for k := 0; k < nSS; k++ {
				if r.MlsagSigs[i].ss[j][k], err = crypto.ParseKey(buf); err != nil {
					return
				}
			}
		}
		if r.MlsagSigs[i].cc, err = crypto.ParseKey(buf); err != nil {
			return
		}
	}

	if XMR_COMPATIBILITY == true { // parse XMR bulletproofs
		// parse pseudoouts for bulletproofs
		if r.sigType == RCTTypeSimpleBulletproof {
			for i := 0; i < nInputs; i++ {
				if r.pseudoOuts[i], err = crypto.ParseKey(buf); err != nil {
					return
				}
			}
		}
	}

	//fmt.Printf("mlsag sigs %+v \n",r.MlsagSigs)
	result = r
	return
}

func ParseBulletProof(buf io.Reader) (b BulletProof, err error) {
	if b.A, err = crypto.ParseKey(buf); err != nil {
		return
	}
	if b.S, err = crypto.ParseKey(buf); err != nil {
		return
	}
	if b.T1, err = crypto.ParseKey(buf); err != nil {
		return
	}
	if b.T2, err = crypto.ParseKey(buf); err != nil {
		return
	}
	if b.taux, err = crypto.ParseKey(buf); err != nil {
		return
	}
	if b.mu, err = crypto.ParseKey(buf); err != nil {
		return
	}

	lcount, err := ReadVarInt(buf)
	if err != nil {
		return
	}

	if lcount > 1024 {
		err = fmt.Errorf("Detected dos bulletproof  bad L %d", lcount)
		return
	}
	b.L = make([]crypto.Key, lcount, lcount)

	for i := range b.L {
		if b.L[i], err = crypto.ParseKey(buf); err != nil {
			return
		}
	}

	rcount, err := ReadVarInt(buf)
	if err != nil {
		return
	}
	if rcount > 1024 {
		err = fmt.Errorf("Detected dos bulletproof  bad L %d", rcount)
		return
	}
	b.R = make([]crypto.Key, rcount, rcount)

	for i := range b.R {
		if b.R[i], err = crypto.ParseKey(buf); err != nil {
			return
		}
	}

	if b.a, err = crypto.ParseKey(buf); err != nil {
		return
	}
	if b.b, err = crypto.ParseKey(buf); err != nil {
		return
	}
	if b.t, err = crypto.ParseKey(buf); err != nil {
		return
	}

	return
}

// serialize the bullet proof
func (b *BulletProof) Serialize() (result []byte) {
	var buf bytes.Buffer
	buf.Write(b.A[:])
	buf.Write(b.S[:])
	buf.Write(b.T1[:])
	buf.Write(b.T2[:])
	buf.Write(b.taux[:])
	buf.Write(b.mu[:])
	buf.Write(Uint64ToBytes(uint64(len(b.L))))
	for i := range b.L {
		buf.Write(b.L[i][:])
	}
	buf.Write(Uint64ToBytes(uint64(len(b.R))))
	for i := range b.R {
		buf.Write(b.R[i][:])
	}
	buf.Write(b.a[:])
	buf.Write(b.b[:])
	buf.Write(b.t[:])
	return buf.Bytes()
}

/*
   //Elliptic Curve Diffie Helman: encodes and decodes the amount b and mask a
   // where C= aG + bH
   void ecdhEncode(ecdhTuple & unmasked, const key & sharedSec) {
       key sharedSec1 = hash_to_scalar(sharedSec);
       key sharedSec2 = hash_to_scalar(sharedSec1);
       //encode
       sc_add(unmasked.mask.bytes, unmasked.mask.bytes, sharedSec1.bytes);
       sc_add(unmasked.amount.bytes, unmasked.amount.bytes, sharedSec2.bytes);
   }
   void ecdhDecode(ecdhTuple & masked, const key & sharedSec) {
       key sharedSec1 = hash_to_scalar(sharedSec);
       key sharedSec2 = hash_to_scalar(sharedSec1);
       //decode
       sc_sub(masked.mask.bytes, masked.mask.bytes, sharedSec1.bytes);
       sc_sub(masked.amount.bytes, masked.amount.bytes, sharedSec2.bytes);
   }
*/
func ecdhEncode(tuple *ECdhTuple, shared_secret crypto.Key) {
	shared_secret1 := crypto.HashToScalar(shared_secret[:])
	shared_secret2 := crypto.HashToScalar(shared_secret1[:])

	// encode
	crypto.ScAdd(&tuple.Mask, &tuple.Mask, shared_secret1)
	crypto.ScAdd(&tuple.Amount, &tuple.Amount, shared_secret2)
}

func ecdhDecode(tuple *ECdhTuple, shared_secret crypto.Key) {
	shared_secret1 := crypto.HashToScalar(shared_secret[:])
	shared_secret2 := crypto.HashToScalar(shared_secret1[:])

	// decode
	crypto.ScSub(&tuple.Mask, &tuple.Mask, shared_secret1)
	crypto.ScSub(&tuple.Amount, &tuple.Amount, shared_secret2)
}

// decode and verify a previously encrypted tuple
// the keys come in from the wallet
// tuple is the encoded data
// skey is the secret scalar key
// outpk is public key used to verify whether the decode was sucessfull
func Decode_Amount(tuple ECdhTuple, skey crypto.Key, outpk crypto.Key) (amount uint64, mask crypto.Key, result bool) {
	var Ctmp crypto.Key

	ecdhDecode(&tuple, skey) // decode the amounts

	// saniity check similiar to  original one
	// addKeys2(Ctmp, mask, amount, H);
	crypto.AddKeys2(&Ctmp, &tuple.Mask, &tuple.Amount, &H)

	if Ctmp != outpk {
		fmt.Printf("warning, amount decoded incorrectly, will be unable to spend")
		result = false
		return
	}
	amount = h2d(tuple.Amount)
	mask = tuple.Mask

	result = true
	return
}

/* from rctOps.cpp
//generates C =aG + bH from b, a is given..
    void genC(key & C, const key & a, xmr_amount amount) {
        key bH = scalarmultH(d2h(amount));
        addKeys1(C, a, bH);
    }
*/
// Commit X amount to random  // see Commitment_From_Amount and ZeroCommitment_From_Amount in key.go
func genC(C *crypto.Key, a *crypto.Key, amount uint64) {
	bH := crypto.ScalarMultH(d2h(amount))
	aG := crypto.ScalarmultBase(*a)
	crypto.AddKeys(C, &aG, bH)
}

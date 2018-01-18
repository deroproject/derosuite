package ringct

import "io"
import "fmt"

import "github.com/deroproject/derosuite/crypto"


// TODO this package need serious love of atleast few weeks
// but atleast the parser and serdes works
// we neeed to expand everthing so as chances of a bug slippping in becomes very low
// NOTE:DO NOT waste time implmenting pre-RCT code

const (
	RCTTypeNull = iota
	RCTTypeFull
	RCTTypeSimple
)

// Pedersen Commitment is generated from this struct
// C = aG + bH where a = mask and b = amount
// senderPk is the one-time public key for ECDH exchange
type ecdhTuple struct {
	mask     Key
	amount   Key
	senderPk Key
}

// Range proof commitments
type Key64 [64]Key


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
	ee Key
}

// MLSAG (Multilayered Linkable Spontaneous Anonymous Group) Signature
type MlsagSig struct {
	ss [][]Key 
	cc Key    // this stores the starting point
	II []Key // this stores the keyimage, but is taken from the tx,it is NOT serialized
}


// Confidential Transaction Keys, mask is Pedersen Commitment
// most of the time, it holds public keys, except where it holds private keys
type CtKey struct {
	Destination Key  // this is the destination and needs to expanded from blockchain
	Mask        Key  // this is the public key mask
}

// Ring Confidential Signature parts that we have to keep
type RctSigBase struct {
	sigType    uint8
	Message    Key  // transaction prefix hash
	MixRing    [][]CtKey // this is not serialized
	pseudoOuts []Key
	ecdhInfo   []ecdhTuple
	outPk      []CtKey // only mask amount is serialized 
	txFee      uint64
	
	Txid       crypto.Hash // this field is extra and only used for logging purposes to track which txid was at fault
}

// Ring Confidential Signature parts that we can just prune later
type RctSigPrunable struct {
	rangeSigs []RangeSig
	MlsagSigs []MlsagSig
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
	for _, ecdh := range r.ecdhInfo {
		result = append(result, ecdh.mask[:]...)
		result = append(result, ecdh.amount[:]...)
	}
	for _, ctKey := range r.outPk {
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
	for _, mlsagSig := range r.MlsagSigs {
		result = append(result, mlsagSig.Serialize()...)
	}
	return
}


func (r *RctSig) Get_Sig_Type() (byte) {
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
func  (r *RctSig) Verify() (result bool) {
    
    result = false
    defer func() {  // safety so if anything wrong happens, verification fails
        if r := recover(); r != nil {
            //connection.logger.Fatalf("Recovered while Verify transaction", r)
            fmt.Printf("Recovered while Verify transaction")
            result = false
        }}()
    
    switch r.sigType {
        case  RCTTypeNull:  return true /// this is only possible for miner tx
        case RCTTypeFull :  return r.VerifyRctFull()
        case RCTTypeSimple: return r.VerifyRctSimple()
            
        default :
            return false
    }
    
    
    return false
}


// Verify a RCTTypeSimple RingCT Signature
func (r *RctSig) VerifyRctSimple() bool {
	sumOutPks := identity()
	for _, ctKey := range r.outPk {
		AddKeys(sumOutPks, sumOutPks, &ctKey.Mask)
	}
	txFeeKey := ScalarMultH(d2h(r.txFee))
	AddKeys(sumOutPks, sumOutPks, txFeeKey)
	sumPseudoOuts := identity()
	for _, pseudoOut := range r.pseudoOuts {
		AddKeys(sumPseudoOuts, sumPseudoOuts, &pseudoOut)
	}
	if *sumPseudoOuts != *sumOutPks {
		return false
	}
	for i, ctKey := range r.outPk {
		if !VerifyRange(&ctKey.Mask, r.rangeSigs[i]) {
			return false
		}
	}
	
	// BUG BUG we are not verifying mlsag here, Do it once the core finishes
	
	return true
}

func (r *RctSig) VerifyRctFull() bool {
	for i, ctKey := range r.outPk {
		if !VerifyRange(&ctKey.Mask, r.rangeSigs[i]) {
			return false
		}
	}
	
	// BUG BUG we are not verifying mlsag here, Do it once the core is finished
	return true
}

func ParseCtKey(buf io.Reader) (result CtKey, err error) {
	if result.Mask, err = ParseKey(buf); err != nil {
		return
	}
	return
}

func ParseKey64(buf io.Reader) (result Key64, err error) {
	for i := 0; i < 64; i++ {
		if result[i], err = ParseKey(buf); err != nil {
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
	if result.ee, err = ParseKey(buf); err != nil {
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
	if r.sigType != RCTTypeFull || r.sigType != RCTTypeSimple {
		err = fmt.Errorf("Bad signature Type %d", r.sigType)
	}
	r.txFee, err = ReadVarInt(buf)
	if err != nil {
		return
	}
	var nMg, nSS int
	if r.sigType == RCTTypeSimple {
		nMg = nInputs
		nSS = 2
		r.pseudoOuts = make([]Key, nInputs)
		for i := 0; i < nInputs; i++ {
			if r.pseudoOuts[i], err = ParseKey(buf); err != nil {
				return
			}
		}
	} else {
		nMg = 1
		nSS = nInputs + 1
	}
	r.ecdhInfo = make([]ecdhTuple, nOutputs)
	for i := 0; i < nOutputs; i++ {
		if r.ecdhInfo[i].mask, err = ParseKey(buf); err != nil {
			return
		}
		if r.ecdhInfo[i].amount, err = ParseKey(buf); err != nil {
			return
		}
	}
	r.outPk = make([]CtKey, nOutputs)
	for i := 0; i < nOutputs; i++ {
		if r.outPk[i], err = ParseCtKey(buf); err != nil {
			return
		}
	}
	r.rangeSigs = make([]RangeSig, nOutputs)
	for i := 0; i < nOutputs; i++ {
		if r.rangeSigs[i], err = ParseRangeSig(buf); err != nil {
			return
		}
	}
	r.MlsagSigs = make([]MlsagSig, nMg)
	for i := 0; i < nMg; i++ {
		r.MlsagSigs[i].ss = make([][]Key, nMixin+1)
		for j := 0; j < nMixin+1; j++ {
			r.MlsagSigs[i].ss[j] = make([]Key, nSS)
			for k := 0; k < nSS; k++ {
				if r.MlsagSigs[i].ss[j][k], err = ParseKey(buf); err != nil {
					return
				}
			}
		}
		if r.MlsagSigs[i].cc, err = ParseKey(buf); err != nil {
			return
		}
	}
	result = r
	return
}

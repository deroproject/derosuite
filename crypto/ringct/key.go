package ringct


import "io"
import "crypto/rand"

import "github.com/deroproject/derosuite/crypto"



const KeyLength = 32


// Key can be a Scalar or a Point
type Key [KeyLength]byte

func (p *Key) FromBytes(b [KeyLength]byte) {
	*p = b
}

func (p *Key) ToBytes() (result [KeyLength]byte) {
	result = [KeyLength]byte(*p)
	return
}

func (p *Key) PubKey() (pubKey *Key) {
	point := new(ExtendedGroupElement)
	GeScalarMultBase(point, p)
	pubKey = new(Key)
	point.ToBytes(pubKey)
	return
}

// Creates a point on the Edwards Curve by hashing the key
func (p *Key) HashToEC() (result *ExtendedGroupElement) {
	result = new(ExtendedGroupElement)
	var p1 ProjectiveGroupElement
	var p2 CompletedGroupElement
	h := Key(crypto.Keccak256(p[:]))
	p1.FromBytes(&h)
	GeMul8(&p2, &p1)
	p2.ToExtended(result)
	return
}

func RandomScalar() (result *Key) {
	result = new(Key)
	var reduceFrom [KeyLength * 2]byte
	tmp := make([]byte, KeyLength*2)
	rand.Read(tmp)
	copy(reduceFrom[:], tmp)
	ScReduce(result, &reduceFrom)
	return
}

func NewKeyPair() (privKey *Key, pubKey *Key) {
	privKey = RandomScalar()
	pubKey = privKey.PubKey()
	return
}

func ParseKey(buf io.Reader) (result Key, err error) {
	key := make([]byte, KeyLength)
	if _, err = buf.Read(key); err != nil {
		return
	}
	copy(result[:], key)
	return
}

/*
   //does a * G where a is a scalar and G is the curve basepoint
    key scalarmultBase(const key & a) {
        ge_p3 point;
        key aG;
        sc_reduce32copy(aG.bytes, a.bytes); //do this beforehand
        ge_scalarmult_base(&point, aG.bytes);
        ge_p3_tobytes(aG.bytes, &point);
        return aG;
    }
  */  
//does a * G where a is a scalar and G is the curve basepoint

func ScalarmultBase(a Key) (aG Key){
    reduce32copy :=  a
    ScReduce32(&reduce32copy)
    point := new(ExtendedGroupElement)
    GeScalarMultBase(point, &a)
    point.ToBytes(&aG)
    return aG
}    

// generates a key which can be used as private key or mask
// this function is similiar to  RandomScalar except for reduce32, TODO can we merge both
func skGen() Key {
    skey := RandomScalar()
    ScReduce32(skey)
    return *skey
}


func (k *Key) ToExtended() (result *ExtendedGroupElement) {
	result = new(ExtendedGroupElement)
	result.FromBytes(k)
	return
}

func identity() (result *Key) {
	result = new(Key)
	result[0] = 1
	return
}

// convert a uint64 to a scalar
func d2h(val uint64) (result *Key) {
	result = new(Key)
	for i := 0; val > 0; i++ {
		result[i] = byte(val & 0xFF)
		val /= 256
	}
	return
}

func HashToScalar(data ...[]byte) (result *Key) {
	result = new(Key)
	*result = Key(crypto.Keccak256(data...))
	ScReduce32(result)
	return
}

// multiply a scalar by H (second curve point of Pedersen Commitment)
func ScalarMultH(scalar *Key) (result *Key) {
	h := new(ExtendedGroupElement)
	h.FromBytes(&H)
	resultPoint := new(ProjectiveGroupElement)
	GeScalarMult(resultPoint, scalar, h)
	result = new(Key)
	resultPoint.ToBytes(result)
	return
}

// add two points together
func AddKeys(sum, k1, k2 *Key) {
	a := k1.ToExtended()
	b := new(CachedGroupElement)
	k2.ToExtended().ToCached(b)
	c := new(CompletedGroupElement)
	geAdd(c, a, b)
	tmp := new(ExtendedGroupElement)
	c.ToExtended(tmp)
	tmp.ToBytes(sum)
	return
}

// compute a*G + b*B
func AddKeys2(result, a, b, B *Key) {
	BPoint := B.ToExtended()
	RPoint := new(ProjectiveGroupElement)
	GeDoubleScalarMultVartime(RPoint, b, BPoint, a)
	RPoint.ToBytes(result)
	return
}

// subtract two points A - B
func SubKeys(diff, k1, k2 *Key) {
	a := k1.ToExtended()
	b := new(CachedGroupElement)
	k2.ToExtended().ToCached(b)
	c := new(CompletedGroupElement)
	geSub(c, a, b)
	tmp := new(ExtendedGroupElement)
	c.ToExtended(tmp)
	tmp.ToBytes(diff)
	return
}

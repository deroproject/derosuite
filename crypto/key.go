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

package crypto

import "io"
import "fmt"
import "bytes"
import "crypto/rand"
import "encoding/hex"
import "encoding/binary"

const KeyLength = 32

// Key can be a Scalar or a Point
type Key [KeyLength]byte

func (k Key) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("%x", k[:])), nil
}

func (k *Key) UnmarshalText(data []byte) (err error) {
	byteSlice, _ := hex.DecodeString(string(data))
	if len(byteSlice) != 32 {
		return fmt.Errorf("Incorrect key size")
	}
	copy(k[:], byteSlice)
	return
}

func (k Key) String() string {
	return fmt.Sprintf("%x", k[:])
}

func (p *Key) FromBytes(b [KeyLength]byte) {
	*p = b
}

func (p *Key) ToBytes() (result [KeyLength]byte) {
	result = [KeyLength]byte(*p)
	return
}

// convert a hex string to a key
func HexToKey(h string) (result Key) {
	byteSlice, _ := hex.DecodeString(h)
	if len(byteSlice) != 32 {
		panic("Incorrect key size")
	}
	copy(result[:], byteSlice)
	return
}

func HexToHash(h string) (result Hash) {
	byteSlice, _ := hex.DecodeString(h)
	if len(byteSlice) != 32 {
		panic("Incorrect key size")
	}
	copy(result[:], byteSlice)
	return
}

// generates a public from the secret key
func (p *Key) PublicKey() (pubKey *Key) {
	point := new(ExtendedGroupElement)
	GeScalarMultBase(point, p)
	pubKey = new(Key)
	point.ToBytes(pubKey)
	return
}

// tests whether the key is valid ( represents a point on the curve )
// this is equivalent to bool crypto_ops::check_key(const public_key &key)
func (k *Key) Public_Key_Valid() bool {
	var point ExtendedGroupElement
	return point.FromBytes(k)
}

func (k *Key) Private_Key_Valid() bool {
	return Sc_check(k)
}

// Creates a point on the Edwards Curve by hashing the key
func (p *Key) HashToEC() (result *ExtendedGroupElement) {
	result = new(ExtendedGroupElement)
	var p1 ProjectiveGroupElement
	var p2 CompletedGroupElement
	h := Key(Keccak256(p[:]))
	p1.FromBytes(&h)

	// fmt.Printf("p1 %+v\n", p1)
	GeMul8(&p2, &p1)
	p2.ToExtended(result)
	return
}

func (p *Key) HashToPoint() (result Key) {
	extended := p.HashToEC()
	extended.ToBytes(&result)
	return
}

// compatible with hashToPointSimple
// NOTE: this is incompatible with HashToPoint ( though it should have been)
// there are no side-effects or degradtion of crypto, due to this
// however, the mistakes have to kept as they were in original code base
// this function is only used to generate H from G
func (p *Key) HashToPointSimple() (result Key) {
	h := Key(Keccak256(p[:]))
	extended := new(ExtendedGroupElement)
	extended.FromBytes(&h)

	// convert extended to projective
	var p1 ProjectiveGroupElement

	extended.ToProjective(&p1)
	var p2 CompletedGroupElement

	GeMul8(&p2, &p1)
	p2.ToExtended(extended)
	extended.ToBytes(&result)
	return
}

// this uses random number generator from the OS
func RandomScalar() (result *Key) {
	result = new(Key)
	var reduceFrom [KeyLength * 2]byte
	tmp := make([]byte, KeyLength*2)
	rand.Read(tmp)
	copy(reduceFrom[:], tmp)
	ScReduce(result, &reduceFrom)
	return
}

// generate a new private-public key pair
func NewKeyPair() (privKey *Key, pubKey *Key) {
	privKey = RandomScalar()
	pubKey = privKey.PublicKey()
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

func ScalarmultBase(a Key) (aG Key) {
	reduce32copy := a
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
func SkGen() Key {
	return skGen()
}

func (k *Key) ToExtended() (result *ExtendedGroupElement) {
	result = new(ExtendedGroupElement)
	result.FromBytes(k)
	return
}

// bothe the function resturn identity of the ed25519 curve
func identity() (result *Key) {
	result = new(Key)
	result[0] = 1
	return
}

func CurveIdentity() (result Key) {
	result = Identity
	return result
}

func CurveOrder() (result Key) {
	result = L
	return result
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
	*result = Key(Keccak256(data...))
	ScReduce32(result)
	return
}

// does a * P where a is a scalar and P is an arbitrary point
func ScalarMultKey(Point *Key, scalar *Key) (result *Key) {
	var P ExtendedGroupElement
	P.FromBytes(Point)
	var resultPoint ProjectiveGroupElement
	GeScalarMult(&resultPoint, scalar, &P)
	result = new(Key)
	resultPoint.ToBytes(result)
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
	var b CachedGroupElement
	k2.ToExtended().ToCached(&b)
	var c CompletedGroupElement
	geAdd(&c, a, &b)
	var tmp ExtendedGroupElement
	c.ToExtended(&tmp)
	tmp.ToBytes(sum)
	return
}

// compute a*G + b*B
func AddKeys2(result, a, b, B *Key) {
	BPoint := B.ToExtended()
	var RPoint ProjectiveGroupElement
	GeDoubleScalarMultVartime(&RPoint, b, BPoint, a)
	RPoint.ToBytes(result)
	return
}

//addKeys3
//aAbB = a*A + b*B where a, b are scalars, A, B are curve points
//B must be input after applying "precomp"
func AddKeys3(result *Key, a *Key, A *Key, b *Key, B_Precomputed *[8]CachedGroupElement) {
	var A_Point ExtendedGroupElement
	A_Point.FromBytes(A)

	var result_projective ProjectiveGroupElement
	GeDoubleScalarMultPrecompVartime(&result_projective, a, &A_Point, b, B_Precomputed)
	result_projective.ToBytes(result)

}

//addKeys3_3  this is similiar to addkeys3 except it allows for use of precomputed A,B
//aAbB = a*A + b*B where a, b are scalars, A, B are curve points
//A,B must be input after applying "precomp"
func AddKeys3_3(result *Key, a *Key, A_Precomputed *[8]CachedGroupElement, b *Key, B_Precomputed *[8]CachedGroupElement) {
	var result_projective ProjectiveGroupElement
	GeDoubleScalarMultPrecompVartime2(&result_projective, a, A_Precomputed, b, B_Precomputed)
	result_projective.ToBytes(result)

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

// zero fill the key
func Sc_0(k *Key) {
	for i := 0; i < 32; i++ {
		k[i] = 0
	}
}

// RandomPubKey takes a random scalar, interprets it as a point on the curve
//  remember the low order bug and do more auditing of the entire thing
func RandomPubKey() (result *Key) {
	result = new(Key)
	p3 := new(ExtendedGroupElement)
	var p1 ProjectiveGroupElement
	var p2 CompletedGroupElement
	h := RandomScalar()
	p1.FromBytes(h)
	GeMul8(&p2, &p1)
	p2.ToExtended(p3)
	p3.ToBytes(result)
	return
}

// this is the main key derivation function and is the crux
// when deriving keys in the case user  A wants to send DERO to another user B ( this is outgoing case)
// public key is B's view key
// private keys is TX private key
// if user B wants to derive key, he needs to   ( this is incoming case )
// public key is TX public key
// private is B's private keys
// HOPE the above is  clean and clear

func KeyDerivation(pub *Key, priv *Key) (KeyDerivation Key) {
	var point ExtendedGroupElement
	var point2 ProjectiveGroupElement
	var point3 CompletedGroupElement

	if !priv.Private_Key_Valid() {
		panic("Invalid private key.")
	}
	tmp := *pub
	if !point.FromBytes(&tmp) {
		panic("Invalid public key.")
	}

	tmp = *priv
	GeScalarMult(&point2, &tmp, &point)
	GeMul8(&point3, &point2)
	point3.ToProjective(&point2)

	point2.ToBytes(&tmp)
	return tmp
}

// the origincal c implementation needs to be checked for varint overflow
// we also need to check the compatibility of golang varint with cryptonote implemented varint
// outputIndex is the position of output within that specific transaction
func (k *Key) KeyDerivationToScalar(outputIndex uint64) (scalar *Key) {
	tmp := make([]byte, 12, 12)

	length := binary.PutUvarint(tmp, outputIndex)
	tmp = tmp[:length]

	var buf bytes.Buffer
	buf.Write(k[:])
	buf.Write(tmp)
	scalar = HashToScalar(buf.Bytes())
	return
}

// generate ephermal keys  from a key derivation
// base key is the B's public spend key or A's private spend key
// outputIndex is the position of output within that specific transaction
func (kd *Key) KeyDerivation_To_PublicKey(outputIndex uint64, baseKey Key) Key {

	var point1, point2 ExtendedGroupElement
	var point3 CachedGroupElement
	var point4 CompletedGroupElement
	var point5 ProjectiveGroupElement

	tmp := baseKey
	if !point1.FromBytes(&tmp) {
		panic("Invalid public key.")
	}
	scalar := kd.KeyDerivationToScalar(outputIndex)
	GeScalarMultBase(&point2, scalar)
	point2.ToCached(&point3)
	geAdd(&point4, &point1, &point3)
	point4.ToProjective(&point5)
	point5.ToBytes(&tmp)
	return tmp
}

// generate ephermal keys  from a key derivation
// base key is the A's private spend key
// outputIndex is the position of output within that specific transaction
func (kd *Key) KeyDerivation_To_PrivateKey(outputIndex uint64, baseKey Key) Key {
	if !baseKey.Private_Key_Valid() {
		panic("Invalid private key.")
	}
	scalar := kd.KeyDerivationToScalar(outputIndex)

	tmp := baseKey
	ScAdd(&tmp, &tmp, scalar)
	return tmp
}

// NewKeyImage creates a new KeyImage from the given public and private keys.
// The keys are usually the ephemeral keys derived using KeyDerivation.
func GenerateKeyImage(pub Key, private Key) Key {
	var proj ProjectiveGroupElement

	ext := pub.HashToEC()
	GeScalarMult(&proj, &private, ext)

	var ki Key
	proj.ToBytes(&ki)
	return ki
}

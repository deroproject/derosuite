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
import "crypto/rand"

import "github.com/deroproject/derosuite/crypto"

const KeyLength = 32

// Key can be a Scalar or a Point
type Key [KeyLength]byte

func (k Key) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("%x", k[:])), nil
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

func (p *Key) HashToPoint() (result Key) {
	extended := p.HashToEC()
	extended.ToBytes(&result)
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

//32 byte key to uint long long
// if the key holds a value > 2^64
// then the value in the first 8 bytes is returned
func h2d(input Key) (value uint64) {
	for j := 7; j >= 0; j-- {
		value = (value*256 + uint64(input[j]))
	}
	return value
}

func HashToScalar(data ...[]byte) (result *Key) {
	result = new(Key)
	*result = Key(crypto.Keccak256(data...))
	ScReduce32(result)
	return
}

// does a * P where a is a scalar and P is an arbitrary point
func ScalarMultKey(Point *Key, scalar *Key) (result *Key) {
	P := new(ExtendedGroupElement)
	P.FromBytes(Point)
	resultPoint := new(ProjectiveGroupElement)
	GeScalarMult(resultPoint, scalar, P)
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

//addKeys3
//aAbB = a*A + b*B where a, b are scalars, A, B are curve points
//B must be input after applying "precomp"
func AddKeys3(result *Key, a *Key, A *Key, b *Key, B_Precomputed *[8]CachedGroupElement) {
	A_Point := new(ExtendedGroupElement)
	A_Point.FromBytes(A)

	result_projective := new(ProjectiveGroupElement)
	GeDoubleScalarMultPrecompVartime(result_projective, a, A_Point, b, B_Precomputed)
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

// this gives you a commitment from an amount
// this is used to convert tx fee or miner tx amount to commitment
func Commitment_From_Amount(amount uint64) Key {
	return *(ScalarMultH(d2h(amount)))
}

// this is used to convert miner tx commitment to  mask
// equivalent to rctOps.cpp zeroCommit
func ZeroCommitment_From_Amount(amount uint64) Key {
	mask := *(identity())
	mask = ScalarmultBase(mask)
	am := d2h(amount)
	bH := ScalarMultH(am)
	AddKeys(&mask, &mask, bH)
	return mask
}

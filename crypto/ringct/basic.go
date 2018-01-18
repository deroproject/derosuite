package ringct

import "encoding/hex"
import "github.com/deroproject/derosuite/crypto"

// convert a hex string to a key
// a copy of these functions exists in the crypto package also
func HexToKey(h string) (result Key) {
	byteSlice, _ := hex.DecodeString(h)
        if len(byteSlice) != 32 {
            panic("Incorrect key size")
        }
	copy(result[:], byteSlice)
	return
}

func HexToHash(h string) (result crypto.Hash) {
	byteSlice, _ := hex.DecodeString(h)
        if len(byteSlice) != 32 {
            panic("Incorrect key size")
        }
	copy(result[:], byteSlice)
	return
}

// zero fill the key
func Sc_0(k *Key) {
    for i:=0; i < 32;i++{
        k[i]=0
    }
}

// RandomPubKey takes a random scalar, interprets it as a point on the curve
// and then multiplies by 8 to make it a point in the Group
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

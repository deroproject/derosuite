package crypto

import "fmt"

const KeyLength = 32


// Key can be a Scalar or a Point
type Key [KeyLength]byte

func (k Key) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("%x", k[:])), nil
}

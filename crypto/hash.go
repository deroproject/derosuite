package crypto

import "fmt"
import "encoding/hex"

const (
	ChecksumLength = 4 // for addresses
	HashLength     = 32
)

type Hash [HashLength]byte
type Checksum [ChecksumLength]byte

func (h Hash) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("%x", h[:])), nil
}

// convert a hash of hex form to binary form, returns a zero hash if any error
// TODO this should be in crypto
func HashHexToHash(hash_hex string) (hash Hash) {
	hash_raw, err := hex.DecodeString(hash_hex)

	if err != nil {
		//panic(fmt.Sprintf("Cannot hex decode checkpint hash \"%s\"", hash_hex))
		return hash
	}

	if len(hash_raw) != 32 {
		//panic(fmt.Sprintf(" hash not 32 byte size Cannot hex decode checkpint hash \"%s\"", hash_hex))
		return hash
	}

	copy(hash[:], hash_raw)
	return
}

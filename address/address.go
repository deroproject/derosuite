package address

import "fmt"
import "bytes"
import "encoding/binary"

import "github.com/deroproject/derosuite/crypto"


type Address struct {
	Network     uint64
	SpendingKey []byte
	ViewingKey  []byte

	//TODO add support for integrated address
}

const ChecksumLength = 4

type Checksum [ChecksumLength]byte

func GetChecksum(data ...[]byte) (result Checksum) {
	keccak256 := crypto.Keccak256(data...)
	copy(result[:], keccak256[:4])
	return
}

func (a *Address) Base58() (result string) {

	prefix := make([]byte, 9, 9)

	n := binary.PutUvarint(prefix, a.Network)
	prefix = prefix[:n]

	checksum := GetChecksum(prefix, a.SpendingKey, a.ViewingKey)
	result = EncodeDeroBase58(prefix, a.SpendingKey, a.ViewingKey, checksum[:])
	return
}

func NewAddress(address string) (result *Address, err error) {
	raw := DecodeDeroBase58(address)

	// donot compare length to support much more user base and be compatible with cryptonote
	if len(raw) < 69 { // 1 byte prefix + 32 byte key + 32 byte key + 4 byte checksum
		err = fmt.Errorf("Address is the wrong length")
		return
	}

	checksum := GetChecksum(raw[:len(raw)-4])
	if bytes.Compare(checksum[:], raw[len(raw)-4:]) != 0 {
		err = fmt.Errorf("Checksum does not validate")
		return
	}

	// parse network first
	address_prefix, done := binary.Uvarint(raw)
	if done <= 0 {
		err = fmt.Errorf("Network could not be parsed in address\n")
		return
	}

	raw = raw[done:]

	result = &Address{
		Network:     address_prefix,
		SpendingKey: raw[0:32],
		ViewingKey:  raw[32:64],
	}

	return
}

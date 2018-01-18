package blockchain

// version 1 transaction support
const SIGNATURE_V1_LENGTH = 64

type Signature_v1 struct {
	R [32]byte
	C [32]byte
}

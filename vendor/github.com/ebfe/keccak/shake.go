package keccak

import (
	"hash"
)

// NewSHAKE128 returns a new hash.Hash computing SHAKE128 with a n*8 bit output as specified in the FIPS 202 draft.
func NewSHAKE128(n int) hash.Hash {
	return newKeccak(128*2, n*8, domainSHAKE)
}

// NewSHAKE256 returns a new hash.Hash computing SHAKE256 with a n*8 bit output as specified in the FIPS 202 draft.
func NewSHAKE256(n int) hash.Hash {
	return newKeccak(256*2, n*8, domainSHAKE)
}

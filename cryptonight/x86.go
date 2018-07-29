// +build amd64

package cryptonight

import "github.com/intel-go/cpuid"

func init() {
	if cpuid.HasFeature(cpuid.AES) {
		HardwareAES = true
	}
}

// defined in assembly
func encryptAESRound(xk *uint32, dst, src *uint32)

// defined in assembly
func encrypt10AESRound(xk *uint32, dst, src *uint32)

// defined in assembly
func mul6464128(input *uint64)

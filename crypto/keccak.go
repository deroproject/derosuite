package crypto


import "github.com/ebfe/keccak"


// quick keccak wrapper
func Keccak256(data ...[]byte) (result Hash) {
	h := keccak.New256()
	for _, b := range data {
		h.Write(b)
	}
	r := h.Sum(nil)
	copy(result[:], r)
	return
}

func Keccak512(data ...[]byte) (result Hash) {
	h := keccak.New512()
	for _, b := range data {
		h.Write(b)
	}
	r := h.Sum(nil)
	copy(result[:], r)
	return
}

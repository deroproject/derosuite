package walletapi

import "crypto/sha256"

// all keys are derived as follows ( entirely deterministic )
// ( key is derived from master secret and user supplied key)
// it's only 1 way
// all keys and bucket names are stored using this , except the metadata bucket
// all values are stored using salsa20 aead
// since all keys are dependent on the master password, almost 99.99% analsis are rendered useless
func (w *Wallet) Key2Key(Key []byte) []byte {
	h := sha256.New()
	h.Write(w.master_password)
	h.Write(Key)
	hash := h.Sum(nil)
	return hash[:]
}

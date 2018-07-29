package walletapi

import "testing"
import "crypto/sha256"

// functional test whether  the wrappers are okay
func Test_AEAD_Cipher(t *testing.T) {

	var key = sha256.Sum256([]byte("test"))
	var data = []byte("data")

	encrypted, err := EncryptWithKey(key[:], data)

	if err != nil {
		t.Fatalf("AEAD cipher failed err %s", err)
	}

	//t.Logf("Encrypted data %x %s", encrypted, string(encrypted))

	decrypted, err := DecryptWithKey(key[:], encrypted)
	if err != nil {
		t.Fatalf("AEAD cipher decryption failed, err %s", err)
	}

	if string(decrypted) != "data" {
		t.Fatalf("AEAD cipher encryption/decryption failed")
	}
}

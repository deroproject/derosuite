package walletapi

import "fmt"
import "crypto/rand"

import "golang.org/x/crypto/chacha20poly1305"

// all data in encrypted within the storage using this, PERIOD
// all data has a new nonce, appended to the the data , last 12 bytes
func EncryptWithKey(Key []byte, Data []byte) (result []byte, err error) {
	nonce := make([]byte, chacha20poly1305.NonceSize, chacha20poly1305.NonceSize)
	cipher, err := chacha20poly1305.New(Key)
	if err != nil {
		return
	}

	_, err = rand.Read(nonce)
	if err != nil {
		return
	}
	Data = cipher.Seal(Data[:0], nonce, Data, nil) // is this okay

	result = append(Data, nonce...) // append nonce
	return
}

// extract 12 byte nonce from the data and deseal the data
func DecryptWithKey(Key []byte, Data []byte) (result []byte, err error) {

	// make sure data is atleast 28 byte, 16 bytes of AEAD cipher and 12 bytes of nonce
	if len(Data) < 28 {
		err = fmt.Errorf("Invalid data")
		return
	}

	data_without_nonce := Data[0 : len(Data)-chacha20poly1305.NonceSize]

	nonce := Data[len(Data)-chacha20poly1305.NonceSize:]

	cipher, err := chacha20poly1305.New(Key)
	if err != nil {
		return
	}

	return cipher.Open(result[:0], nonce, data_without_nonce, nil) // result buffer should be different

}

// use master keys, everytime required
func (w *Wallet) Encrypt(Data []byte) (result []byte, err error) {
	return EncryptWithKey(w.master_password, Data)
}

// use master keys, everytime required
func (w *Wallet) Decrypt(Data []byte) (result []byte, err error) {
	return DecryptWithKey(w.master_password, Data)
}

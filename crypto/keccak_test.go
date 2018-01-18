package crypto

import "testing"
import "encoding/hex"



// convert a hex string to a key
func HexToKey(h string) (result Key) {
	byteSlice, _ := hex.DecodeString(h)
        if len(byteSlice) != 32 {
            panic("Incorrect key size")
        }
	copy(result[:], byteSlice)
	return
}

func HexToHash(h string) (result Hash) {
	byteSlice, _ := hex.DecodeString(h)
        if len(byteSlice) != 32 {
            panic("Incorrect key size")
        }
	copy(result[:], byteSlice)
	return
}

func TestKeccak256(t *testing.T) {
	tests := []struct {
		name       string
		messageHex string
		wantHex    string
	}{
		{
			name:       "from monero 1",
			messageHex: "c8fedd380dbae40ffb52",
			wantHex:    "8e41962058b7422e7404253121489a3e63d186ed115086919a75105661483ba9",
		},
		{
			name:       "from monero 2",
			messageHex: "5020c4d530b6ec6cb4d9",
			wantHex:    "8a597f11961935e32e0adeab2ce48b3df2d907c9b26619dad22f42ff65ab7593",
		},
		{
			name:       "hello",
			messageHex: "68656c6c6f",
			wantHex:    "1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8",
		},
		{
			name:       "from monero cryptotest.pl",
			messageHex: "0f3fe9c20b24a11bf4d6d1acd335c6a80543f1f0380590d7323caf1390c78e88",
			wantHex:    "73b7a236f2a97c4e1805f7a319f1283e3276598567757186c526caf9a49e0a92",
		},

	}
	for _, test := range tests {
		message, _ := hex.DecodeString(test.messageHex)
		got := Keccak256(message)
		want := HexToHash(test.wantHex)
		if want != got {
			t.Errorf("want %x, got %x", want, got)
		}
	}
}

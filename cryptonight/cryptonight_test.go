package cryptonight

import "testing"
import "encoding/hex"

func Test_Cryptonight_Hash(t *testing.T) {

	// there are 4 sub algorithms, basically blake , groestl , jh , skein
	// other things are common, these tect vectors have been manually pulled from c codes

	blake_hash := cryptonight([]byte("This is a testi" + "\x01"))

	if hex.EncodeToString(blake_hash) != "7958d1afe0c46670642c0341f92e89bf6de6a2573ef89742237162e66ea4a121" {
		t.Error("Cryptonight blake_hash testing Failed\n")
		return

	}

	// this is from cryptonote whitepaper
	groestl_hash := cryptonight([]byte("This is a test" + "\x01"))

	if hex.EncodeToString(groestl_hash) != "a084f01d1437a09c6985401b60d43554ae105802c5f5d8a9b3253649c0be6605" {
		t.Error("Cryptonight testing Failed\n")
		return

	}

	jh_hash := cryptonight([]byte("This is a test2" + "\x01"))

	if hex.EncodeToString(jh_hash) != "6f93b51852d1a47277c62e720bf0e10bf90e92123847be246f67e3fd2639f4b4" {
		t.Error("Cryptonight testing Failed\n")
		return

	}

	skein_hash := cryptonight([]byte("This is a testw" + "\x01"))

	if hex.EncodeToString(skein_hash) != "3174ef437b24fd30e81d307d9b7d02ba21eb6f627cafc9d8134ea63adc4985b0" {
		t.Error("Cryptonight testing Failed\n")
		return

	}

}

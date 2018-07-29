package walletapi

import "os"
import "bytes"
import "path/filepath"
import "testing"

//import "fmt"

import "github.com/deroproject/derosuite/crypto"

// quick testing of wallet creation
func Test_Wallet_DB(t *testing.T) {

	temp_db := filepath.Join(os.TempDir(), "dero_temporary_test_wallet.db")

	os.Remove(temp_db)
	w, err := Create_Encrypted_Wallet(temp_db, "QWER", *crypto.RandomScalar())
	if err != nil {
		t.Fatalf("Cannot create encrypted wallet, err %s", err)
	}
	w.Close_Encrypted_Wallet()

	w, err = Open_Encrypted_Wallet(temp_db, "QWER")
	if err != nil {
		t.Fatalf("Cannot open encrypted wallet, err %s", err)
	}

	os.Remove(temp_db)

	//  test deterministc keys
	key := []byte("test")

	if !bytes.Equal(w.Key2Key(key), w.Key2Key(key)) {
		t.Fatalf("Key2Key failed")
	}

	w.Close_Encrypted_Wallet()

}

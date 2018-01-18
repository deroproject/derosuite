package address

import "fmt"
import "bytes"
import "testing"
import "encoding/hex"

import "github.com/deroproject/derosuite/config"



func TestAddressError(t *testing.T) {
	_, err := NewAddress("")
	/*want := "Address is the wrong length"
	if err != want {
		t.Errorf("want: %s, got: %s", want, err)
	}
	*/
	_, err = NewAddress("46w3n5EGhBeYmKvQRsd8UK9GhvcbYWQDobJape3NLMMFEjFZnJ3CnRmeKspubQGiP8iMTwFEX2QiBsjUkjKT4SSPd3fK1")
	want := fmt.Errorf("Checksum does not validate")
	if err.Error() != want.Error() {
		t.Errorf("want: %s, got: %s", want, err)
	}
}

func TestAddress(t *testing.T) {
    
        const Monero_MainNetwork = 18
        const Monero_TestNetwork = 53

	tests := []struct {
		name           string
		Network        uint64
		SpendingKeyHex string
		ViewingKeyHex  string
		Address        string
	}{
		{
			name:           "generic",
			Network:        Monero_MainNetwork,
			SpendingKeyHex: "8c1a9d5ff5aaf1c3cdeb2a1be62f07a34ae6b15fe47a254c8bc240f348271679",
			ViewingKeyHex:  "0a29b163e392eb9416a52907fd7d3b84530f8d02ff70b1f63e72fdcb54cf7fe1",
			Address:        "46w3n5EGhBeZkYmKvQRsd8UK9GhvcbYWQDobJape3NLMMFEjFZnJ3CnRmeKspubQGiP8iMTwFEX2QiBsjUkjKT4SSPd3fKp",
		},
		{
			name:           "generic 2",
			Network:        Monero_MainNetwork,
			SpendingKeyHex: "5007b84275af9a173c2080683afce90b2157ab640c18ddd5ce3e060a18a9ce99",
			ViewingKeyHex:  "27024b45150037b677418fcf11ba9675494ffdf994f329b9f7a8f8402b7934a0",
			Address:        "44f1Y84r9Lu4tQdLWRxV122rygfhUeVBrcmBaqcYCwUHScmf1ht8DFLXX9YN4T7nPPLcpqYLUdrFiY77nQYeH9RuK9gg4p6",
		},
		{
			name:           "require 1 padding in middle",
			Network:        Monero_MainNetwork,
			SpendingKeyHex: "6add197bd82866e8bfbf1dc2fdf49873ec5f679059652da549cd806f2b166756",
			ViewingKeyHex:  "f5cf2897088fda0f7ac1c42491ed7d558a46ee41d0c81d038fd53ff4360afda0",
			Address:        "45fzHekTd5FfvxWBPYX2TqLPbtWjaofxYUeWCi6BRQXYFYd85sY2qw73bAuKhqY7deFJr6pN3STY81bZ9x2Zf4nGKASksqe",
		},
		{
			name:           "require 1 padding in last chunk",
			Network:        Monero_MainNetwork,
			SpendingKeyHex: "50defe92d88b19aaf6bf66f061dd4380b79866a4122b25a03bceb571767dbe7b",
			ViewingKeyHex:  "f8f6f28283921bf5a17f0bcf4306233fc25ce9b6276154ad0de22aebc5c67702",
			Address:        "44grjkXtDHJVbZgtU1UKnrNXidcHfZ3HWToU5WjR3KgHMjgwrYLjXC6i5vm3HCp4vnBfYaNEyNiuZVwqtHD2SenS1JBRyco",
		},
		{
			name:           "testnet",
			Network:        Monero_TestNetwork,
			SpendingKeyHex: "8de9cce254e60cd940abf6c77ef344c3a21fad74320e45734fbfcd5870e5c875",
			ViewingKeyHex:  "27024b45150037b677418fcf11ba9675494ffdf994f329b9f7a8f8402b7934a0",
			Address:        "9xYZvCDf6aFdLd7Qawg5XHZitWLKoeFvcLHfe5GxsGCFLbXSWeQNKciXX9YN4T7nPPLcpqYLUdrFiY77nQYeH9RuK9bogZJ",
		},

		{
			name:           "DERO testnet",
			Network:        config.Testnet.Public_Address_Prefix,
			SpendingKeyHex: "ffb4baf32792d38d36c5f1792201d1cff142a10bad6aa088090156a35858739d",
			ViewingKeyHex:  "0ea428a9608fc9dc06acceea608ac97cc9119647b943941a381306548ee43455",
			Address:        "dETosYceeTxRZQBk5hQzN51JepzZn5H24JqR96q7mY7ZFo6JhJKPNSKR3vs9ES1ibyQDQgeRheDP6CJbb7AKJY2H9eacz2RtPy",
		},
	}
	var base58 string
	var spendingKey, viewingKey []byte
	for _, test := range tests {
		spendingKey, _ = hex.DecodeString(test.SpendingKeyHex)
		viewingKey, _ = hex.DecodeString(test.ViewingKeyHex)

		_ = spendingKey
		_ = viewingKey

		address, err := NewAddress(test.Address)

		if err != nil {
			t.Errorf("%s: Failed while parsing address %s", test.name, err)
			continue
		}
		_ = address

		if address.Network != test.Network {
			t.Errorf("%s: want: %d, got: %d", test.name, test.Network, address.Network)
			continue
		}

		if bytes.Compare(address.SpendingKey, spendingKey) != 0 {
			t.Errorf("%s: want: %x, got: %x", test.name, spendingKey, address.SpendingKey)
			continue
		}
		if bytes.Compare(address.ViewingKey, viewingKey) != 0 {
			t.Errorf("%s: want: %x, got: %x", test.name, viewingKey, address.ViewingKey)
			continue
		}

		base58 = address.Base58()
		if base58 != test.Address {
			t.Errorf("%s: want: %s, got: %s", test.name, test.Address, base58)
			continue
		}

	}
}

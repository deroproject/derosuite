// Copyright 2017-2018 DERO Project. All rights reserved.
// Use of this source code in any form is governed by RESEARCH license.
// license can be found in the LICENSE file.
// GPG: 0F39 E425 8C65 3947 702A  8234 08B2 0360 A03A 9DE8
//
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY
// EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL
// THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
// PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT,
// STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF
// THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package address

import "fmt"
import "bytes"
import "testing"
import "encoding/hex"

import "github.com/deroproject/derosuite/config"

func TestAddressError(t *testing.T) {
	_, err := NewAddress("")
	want := fmt.Errorf("Address is not complete")
	if err.Error() != want.Error() {
		t.Fatalf("want: %s, got: %s", want, err)
	}

	_, err = NewAddress("dERoNzsi5WW1ABhQ1UGLwoLqBU6sbzvyuS4cCi4PGzW7QRM5TH4MUf3QvZUBNJCYSDPw6K495eroGe24cf75uDdD2QwWy9pchN")
	want = fmt.Errorf("Checksum failed")
	if err.Error() != want.Error() {
		t.Fatalf("want: %s, got: %s", want, err)
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
		{
			name:           "DERO mainnet requires padding in second block",
			Network:        config.Mainnet.Public_Address_Prefix,
			SpendingKeyHex: "10a80329a700f25c9892a696de768f5bdc73cafe6095d647e5707c04f48c0481",
			ViewingKeyHex:  "b0fa8ca43a8f07681274ddd8fa891aea4222aa8027dd516bc144317a042547c4",
			Address:        "dERoNzsi5WW1ABhQ1UGLwoLqBU6sbzvyuS4cCi4PGzW7QRM5TH4MUf3QvZUBNJCYSDPw6K495eroGe24cf75uDdD2QwWy9pchM",
		},
	}
	var base58 string
	var spendingKey, viewingKey []byte
	for _, test := range tests {
		spendingKey, _ = hex.DecodeString(test.SpendingKeyHex)
		viewingKey, _ = hex.DecodeString(test.ViewingKeyHex)

		address, err := NewAddress(test.Address)
		if err != nil {
			t.Fatalf("%s: Failed while parsing address %s", test.name, err)
			continue
		}

		if address.Network != test.Network {
			t.Fatalf("%s: want: %d, got: %d", test.name, test.Network, address.Network)
			continue
		}

		if bytes.Compare(address.SpendKey[:], spendingKey) != 0 {
			t.Fatalf("%s: want: %x, got: %s", test.name, spendingKey, address.SpendKey)
			continue
		}
		if bytes.Compare(address.ViewKey[:], viewingKey) != 0 {
			t.Fatalf("%s: want: %x, got: %s", test.name, viewingKey, address.ViewKey)
			continue
		}

		base58 = address.Base58()
		if base58 != test.Address {
			t.Fatalf("%s: want: %s, got: %s", test.name, test.Address, base58)
			continue
		}

	}
}

// more explaination here https://monero.stackexchange.com/questions/1910/how-do-payment-ids-work
// test case created from here https://xmr.llcoins.net/addresstests.html
func TestIntegratedAddress(t *testing.T) {

	const Monero_MainNetwork = 18
	const Monero_MainNetwork_Integrated = 19
	const Monero_TestNetwork = 53

	tests := []struct {
		name           string
		Network        uint64
		NetworkI       uint64
		SpendingKeyHex string
		ViewingKeyHex  string
		PaymentID      string
		Address        string
		AddressI       string
	}{
		{
			name:           "generic",
			Network:        Monero_MainNetwork,
			NetworkI:       Monero_MainNetwork_Integrated,
			SpendingKeyHex: "80d3eca27896f549abc41dd941d08a4c82cff165a7f8bc4c3c0841cffd11c095",
			ViewingKeyHex:  "7849297236cd7c0d6c69a3c8c179c038d3c1c434735741bb3c8995c3c9d6f2ac",
			PaymentID:      "90470a40196034b5",
			Address:        "46WGHoGHRT2DKhdr4BxzhXDoFe5NBjNm1Dka5144aXZHS13cAoUQWRq3FE2gcT3LJjAWJ6fGWq8t8YKRqwwit8vmLT6tcxK",

			AddressI: "4GCwJc5n2iYDKhdr4BxzhXDoFe5NBjNm1Dka5144aXZHS13cAoUQWRq3FE2gcT3LJjAWJ6fGWq8t8YKRqwwit8vmVs5oxyLeWQsMWmcgkC",
		},

		{
			name:           "generic",
			Network:        config.Mainnet.Public_Address_Prefix,
			NetworkI:       config.Mainnet.Public_Address_Prefix_Integrated,
			SpendingKeyHex: "bd7393b76af23611e6e0eb1e4974bcb5688fceea6ad8a1b08435a4e68fcb7b8c",
			ViewingKeyHex:  "c828aa405d78c3a0b0a7263d2cb82811d4c6ee3374ada5cc753d8196a271b3d2",
			PaymentID:      "0cbd6e050cf3b73c",
			Address:        "dERoiVavtPjhWkdEPp17RJLXVoHkr2ucMdEbgGgpskhLb33732LBifWMCZhPga3EcjXoYqfM9jRv3W3bnWUSpdmK5Jur1PhN6P",

			AddressI: "dERijfr9y7XhWkdEPp17RJLXVoHkr2ucMdEbgGgpskhLb33732LBifWMCZhPga3EcjXoYqfM9jRv3W3bnWUSpdmKL24FBjG6ctTAEg1jrhDHh",
		},
	}

	var base58 string
	var spendingKey, viewingKey []byte
	for _, test := range tests {
		spendingKey, _ = hex.DecodeString(test.SpendingKeyHex)
		viewingKey, _ = hex.DecodeString(test.ViewingKeyHex)

		address, err := NewAddress(test.Address)
		if err != nil {
			t.Fatalf("%s: Failed while parsing address %s", test.name, err)
			continue
		}

		if address.Network != test.Network {
			t.Errorf("%s: want: %d, got: %d", test.name, test.Network, address.Network)
			continue
		}

		if bytes.Compare(address.SpendKey[:], spendingKey) != 0 {
			t.Fatalf("%s: want: %x, got: %s", test.name, spendingKey, address.SpendKey)
			continue
		}
		if bytes.Compare(address.ViewKey[:], viewingKey) != 0 {
			t.Fatalf("%s: want: %x, got: %s", test.name, viewingKey, address.ViewKey)
			continue
		}

		base58 = address.Base58()
		if base58 != test.Address {
			t.Fatalf("%s: want: %s, got: %s", test.name, test.Address, base58)
			continue
		}

		address, err = NewAddress(test.AddressI)
		if err != nil {
			t.Fatalf("%s: Failed while parsing address %s", test.name, err)
			continue
		}

		base58 = address.Base58()
		if base58 != test.AddressI {
			t.Fatalf("%s: want: %s, got: %s", test.name, test.AddressI, base58)
			continue
		}

		if fmt.Sprintf("%x", address.PaymentID) != test.PaymentID {
			t.Fatalf("%s: PaymentID want: %s, got: %s", test.name, test.PaymentID, address.PaymentID)
		}

	}

}

func Test_Bruteforce_IntegratedAddress(t *testing.T) {
    var AddressI string = "dERijfr9y7XhWkdEPp17RJLXVoHkr2ucMdEbgGgpskhLb33732LBifWMCZhPga3EcjXoYqfM9jRv3W3bnWUSpdmKL24FBjG6ctTAEg1jrhDHh"
    
    var PaymentID string = "0cbd6e050cf3b73c"
    
    
    for i := 0; i < 100000;i++  {
    address, err := NewAddress(AddressI)
		if err != nil {
			t.Fatalf("%s: Failed while parsing address %s", AddressI, err)
			continue
		}
		if fmt.Sprintf("%x",address.PaymentID) != PaymentID{
                    t.Fatalf("Payment ID failed at loop %d", i)
                }
    }
}

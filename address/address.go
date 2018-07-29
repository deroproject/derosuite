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
import "encoding/binary"

import "github.com/deroproject/derosuite/config"
import "github.com/deroproject/derosuite/crypto"

// see https://cryptonote.org/cns/cns007.txt to understand address more

type Address struct {
	Network   uint64
	SpendKey  crypto.Key // these are public keys only
	ViewKey   crypto.Key // these are public keys only
	PaymentID []byte     //integrated payment id is 8 bytes
	// 8 byte version is encrypted on the blockchain
	// 32 byte version is dumped on the chain openly
}

const ChecksumLength = 4

type Checksum [ChecksumLength]byte

func GetChecksum(data ...[]byte) (result Checksum) {
	keccak256 := crypto.Keccak256(data...)
	copy(result[:], keccak256[:4])
	return
}

func (a *Address) Base58() (result string) {
	prefix := make([]byte, 9, 9)

	n := binary.PutUvarint(prefix, a.Network)
	prefix = prefix[:n]

	// convert address to string ( include payment ID if prefix says so )
	switch a.Network {

	case 19:
		fallthrough // for testing purpose monero integrated address
	case config.Mainnet.Public_Address_Prefix_Integrated:
		fallthrough
	case config.Testnet.Public_Address_Prefix_Integrated:
		checksum := GetChecksum(prefix, a.SpendKey[:], a.ViewKey[:], a.PaymentID)
		result = EncodeDeroBase58(prefix, a.SpendKey[:], a.ViewKey[:], a.PaymentID, checksum[:])

		// normal addresses without prefix
	case config.Mainnet.Public_Address_Prefix:
		fallthrough
	case config.Testnet.Public_Address_Prefix:
		fallthrough
	default:
		checksum := GetChecksum(prefix, a.SpendKey[:], a.ViewKey[:])
		result = EncodeDeroBase58(prefix, a.SpendKey[:], a.ViewKey[:], checksum[:])
	}

	return
}

// stringifier
func (a Address) String() string {
	return a.Base58()
}

// tells whether address is mainnet address
func (a *Address) IsMainnet() bool {
	if a.Network == config.Mainnet.Public_Address_Prefix ||
		a.Network == config.Mainnet.Public_Address_Prefix_Integrated {
		return true
	}
	return false
}

// tells whether address is mainnet address
func (a *Address) IsIntegratedAddress() bool {
	if a.Network == config.Testnet.Public_Address_Prefix_Integrated ||
		a.Network == config.Mainnet.Public_Address_Prefix_Integrated {
		return true
	}
	return false
}

// tells whether address belongs to DERO Network
func (a *Address) IsDERONetwork() bool {
	if a.Network == config.Mainnet.Public_Address_Prefix ||
		a.Network == config.Mainnet.Public_Address_Prefix_Integrated ||
		a.Network == config.Testnet.Public_Address_Prefix ||
		a.Network == config.Testnet.Public_Address_Prefix_Integrated {
		return true
	}
	return false
}

func NewAddress(address string) (result *Address, err error) {
	raw := DecodeDeroBase58(address)

	// donot compare length to support much more user base and be compatible with cryptonote
	if len(raw) < 69 { // 1 byte prefix + 32 byte key + 32 byte key + 4 byte checksum
		err = fmt.Errorf("Address is not complete")
		return
	}
	checksum := GetChecksum(raw[:len(raw)-4])
	if bytes.Compare(checksum[:], raw[len(raw)-4:]) != 0 {
		err = fmt.Errorf("Checksum failed")
		return
	}
	raw = raw[0 : len(raw)-4] // remove the checksum

	// parse network first
	address_prefix, done := binary.Uvarint(raw)
	if done <= 0 {
		err = fmt.Errorf("Network could not be parsed in address\n")
		return
	}

	raw = raw[done:]

	result = &Address{
		Network: address_prefix,
		//SpendKey: raw[0:32],
		//ViewKey:  raw[32:64],
	}
	copy(result.SpendKey[:], raw[0:32])
	copy(result.ViewKey[:], raw[32:64])

	//
	switch address_prefix { // if network id is integrated address
	case 19:
		fallthrough //Monero_MainNetwork_Integrated: for testing purposes only for compatible reasons
	case config.Mainnet.Public_Address_Prefix_Integrated:
		fallthrough // DERO mainnet integrated address
	case config.Testnet.Public_Address_Prefix_Integrated: // DERO testnet integrated address
		if len(raw[64:]) == 8 { // 8 byte encrypted payment id + 4 bytes
			result.PaymentID = raw[64:]
		} else if len(raw[64:]) == 32 { // 32 byte unencrypted payment ID
			result.PaymentID = raw[64:]
		} else {
			err = fmt.Errorf("Invalid payment ID in address\n")
			return
		}
	}

	return
}

// create a new address from keys
func NewAddressFromKeys(spendkey, viewkey crypto.Key) (result *Address) {
	result = &Address{
		Network:  config.Mainnet.Public_Address_Prefix,
		SpendKey: spendkey,
		ViewKey:  viewkey,
	}
	return
}

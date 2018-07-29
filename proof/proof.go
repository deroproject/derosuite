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

package proof

import "fmt"
import "encoding/hex"

import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/address"
import "github.com/deroproject/derosuite/transaction"
import "github.com/deroproject/derosuite/crypto/ringct"
//import "github.com/deroproject/derosuite/walletapi" // to decode encrypted payment ID
 

// used to encrypt payment id
const ENCRYPTED_PAYMENT_ID_TAIL = 0x8d

// this function is used to encrypt/decrypt payment id
// as the operation is symmetric XOR, is the same in both direction
func EncryptDecryptPaymentID(derivation crypto.Key, tx_public crypto.Key, input []byte) (output []byte) {
	// input must be exactly 8 bytes long
	if len(input) != 8 {
		panic("Encrypted payment ID must be exactly 8 bytes long")
	}

	var tmp_buf [33]byte
	copy(tmp_buf[:], derivation[:]) // copy derivation key to buffer
	tmp_buf[32] = ENCRYPTED_PAYMENT_ID_TAIL

	// take hash
	hash := crypto.Keccak256(tmp_buf[:]) // take hash of entire 33 bytes, 32 bytes derivation key, 1 byte tail

	output = make([]byte, 8, 8)
	for i := range input {
		output[i] = input[i] ^ hash[i] // xor the bytes with the hash
	}

	return
}

// this function will prove detect and decode output amount for the tx
func Prove(input_key string, input_addr string, input_tx string) (indexes []uint64, amounts []uint64, payids [][]byte, err error) {
	var tx_secret_key crypto.Key
	var tx transaction.Transaction

	if len(input_key) != 64 {
		err = fmt.Errorf("Invalid input key size")
		return
	}

	tx_secret_key_raw, err := hex.DecodeString(input_key)
	if err != nil {
		return
	}
	copy(tx_secret_key[:], tx_secret_key_raw[:32])

	addr, err := address.NewAddress(input_addr)
	if err != nil {
		return
	}

	tx_hex, err := hex.DecodeString(input_tx)
	if err != nil {
		return
	}

	err = tx.DeserializeHeader(tx_hex)
	if err != nil {
		return
	}

	// okay all inputs have been parsed

	switch tx.RctSignature.Get_Sig_Type() {
	case 0: // miner tx, for miner tx we can only prove that the output belongs to address, TODO
		//fmt.Printf("TX is coinbase and does NOT have encrypted OUTPUTS\n")
		err = fmt.Errorf("TX is coinbase and does NOT have encrypted OUTPUTS")
		return

	}
	
	var PayID8  []byte
	
	if tx.Parse_Extra() {
            // will decrypt payment ID if encrypted 
		if _, ok := tx.PaymentID_map[transaction.TX_EXTRA_NONCE_ENCRYPTED_PAYMENT_ID]; ok {
			PayID8 = tx.PaymentID_map[transaction.TX_EXTRA_NONCE_ENCRYPTED_PAYMENT_ID].([]byte)
		}

	}

	// ringct full/simple tx come here
	derivation := crypto.KeyDerivation(&addr.ViewKey, &tx_secret_key) // keyderivation using output address
	found := false

	// Vout can be only specific type rest all make th fail case
	for i := 0; i < len(tx.Vout); i++ {
		index_within_tx := i

		ehphermal_public_key := derivation.KeyDerivation_To_PublicKey(uint64(index_within_tx), addr.SpendKey)

		if ehphermal_public_key == tx.Vout[i].Target.(transaction.Txout_to_key).Key {
                        found = true
			//fmt.Printf("Output at index %d belongs to  %s\n",i,addr.String())
			indexes = append(indexes, uint64(index_within_tx))

			// we must decode output amounts also
			scalar_key := *(derivation.KeyDerivationToScalar(uint64(index_within_tx)))

			mask := tx.RctSignature.OutPk[i].Mask

			ECDHTuple := tx.RctSignature.ECdhInfo[i]

			amount, _, result := ringct.Decode_Amount(ECDHTuple, scalar_key, mask)
			if result {
				// fmt.Printf("Amount is ~ %0.8f\n", float64(amount)/(1000000000000.0))
				amounts = append(amounts, amount)
                                
                            
                            if len(PayID8) == 8 {
                                 decrypted_pay_id := EncryptDecryptPaymentID(derivation,scalar_key,PayID8)
                                 payids = append(payids,decrypted_pay_id)                                
                            }
			
			}else{
                            err = fmt.Errorf("TX belongs to user but amount could NOT be decoded")
                            return
                        }
                        
		}

	}

	_ = found
	/*if found {
	      fmt.Printf("Found outputs\n")
	  }else{
	      fmt.Printf("Outputs do not belong to this address\n")
	  }*/
        
        if !found{
            err = fmt.Errorf("Wrong TX Key or wrong address or Outputs do not belong to this address")
            
        }

	return

}

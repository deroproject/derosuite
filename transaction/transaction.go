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

package transaction

import "fmt"
import "bytes"
import "encoding/binary"

import "github.com/romana/rlog"

import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/crypto/ringct"

const TXIN_GEN = byte(0xff)
const TXIN_TO_SCRIPT = byte(0)
const TXIN_TO_SCRIPTHASH = byte(1)
const TXIN_TO_KEY = byte(2)

const TXOUT_TO_SCRIPT = byte(0)
const TXOUT_TO_SCRIPTHASH = byte(1)
const TXOUT_TO_KEY = byte(2)

var TX_IN_NAME = map[byte]string{
	TXIN_GEN:           "Coinbase",
	TXIN_TO_SCRIPT:     "To Script",
	TXIN_TO_SCRIPTHASH: "To Script hash",
	TXIN_TO_KEY:        "To key",
}

const TRANSACTION = byte(0xcc)
const BLOCK = byte(0xbb)

/*
VARIANT_TAG(binary_archive, cryptonote::txin_to_script, 0x0);
VARIANT_TAG(binary_archive, cryptonote::txin_to_scripthash, 0x1);
VARIANT_TAG(binary_archive, cryptonote::txin_to_key, 0x2);
VARIANT_TAG(binary_archive, cryptonote::txout_to_script, 0x0);
VARIANT_TAG(binary_archive, cryptonote::txout_to_scripthash, 0x1);
VARIANT_TAG(binary_archive, cryptonote::txout_to_key, 0x2);
VARIANT_TAG(binary_archive, cryptonote::transaction, 0xcc);
VARIANT_TAG(binary_archive, cryptonote::block, 0xbb);
*/
/* outputs */

type Txout_to_script struct {
	// std::vector<crypto::public_key> keys;
	//  std::vector<uint8_t> script;

	Keys   [][32]byte
	Script []byte

	/*  BEGIN_SERIALIZE_OBJECT()
	      FIELD(keys)
	      FIELD(script)
	    END_SERIALIZE()

	*/
}

type Txout_to_scripthash struct {
	//crypto::hash hash;
	Hash [32]byte
}

type Txout_to_key struct {
	Key crypto.Key
	//	Mask [32]byte `json:"-"`
	/*txout_to_key() { }
	  txout_to_key(const crypto::public_key &_key) : key(_key) { }
	  crypto::public_key key;*/

}

// there can be only 4 types if inputs

// used by miner
type Txin_gen struct {
	Height uint64 // stored as varint
}

type Txin_to_script struct {
	Prev    [32]byte
	Prevout uint64
	Sigset  []byte

	/* BEGIN_SERIALIZE_OBJECT()
	     FIELD(prev)
	     VARINT_FIELD(prevout)
	     FIELD(sigset)
	   END_SERIALIZE()
	*/
}

type Txin_to_scripthash struct {
	Prev    [32]byte
	Prevout uint64
	Script  Txout_to_script
	Sigset  []byte

	/* BEGIN_SERIALIZE_OBJECT()
	     FIELD(prev)
	     VARINT_FIELD(prevout)
	     FIELD(script)
	     FIELD(sigset)
	   END_SERIALIZE()
	*/
}

type Txin_to_key struct {
	Amount      uint64
	Key_offsets []uint64 // this is encoded as a varint for length and then all offsets are stored as varint
	//crypto::key_image k_image;      // double spending protection
	K_image crypto.Hash `json:"k_image"` // key image

	/* BEGIN_SERIALIZE_OBJECT()
	     VARINT_FIELD(amount)
	     FIELD(key_offsets)
	     FIELD(k_image)
	   END_SERIALIZE()
	*/
}

type Txin_v interface{} // it can only be txin_gen, txin_to_script, txin_to_scripthash, txin_to_key

type Tx_out struct {
	Amount uint64
	Target interface{} // txout_target_v ;, it can only be  txout_to_script, txout_to_scripthash, txout_to_key

	/* BEGIN_SERIALIZE_OBJECT()
	     VARINT_FIELD(amount)
	     FIELD(target)
	   END_SERIALIZE()
	*/

}

// the core transaction
type Transaction_Prefix struct {
	Version       uint64 `json:"version"`
	Unlock_Time   uint64 `json:"unlock_time"` // used to lock first output
	Vin           []Txin_v
	Vout          []Tx_out
	Extra         []byte
	Extra_map     map[EXTRA_TAG]interface{} `json:"-"` // all information parsed from extra is placed here
	PaymentID_map map[EXTRA_TAG]interface{} `json:"-"` // payments id parsed or set are placed her
	ExtraType     byte                      `json:"-"` // NOT used, candidate for deletion
}

type Transaction struct {
	Transaction_Prefix
	// same as Transaction_Prefix
	// Signature  not sure of what form
	Signature []Signature_v1 `json:"-"` // old format, the array size is always equal to vin length,
	//Signature_RCT RCT_Signature  // version 2

	RctSignature *ringct.RctSig
	Expanded     bool `json:"-"`
}

func (tx *Transaction) GetHash() (result crypto.Hash) {
	switch tx.Version {

	/*case 1:
	result = crypto.Hash(crypto.Keccak256(tx.SerializeHeader()))
	*/

	case 2:
		// version 2 requires first computing 3 separate hashes
		// prefix, rctBase and rctPrunable
		// and then hashing the hashes together to get the final hash
		prefixHash := tx.GetPrefixHash()
		rctBaseHash := tx.RctSignature.BaseHash()
		rctPrunableHash := tx.RctSignature.PrunableHash()
		result = crypto.Hash(crypto.Keccak256(prefixHash[:], rctBaseHash[:], rctPrunableHash[:]))
	default:
		panic("Transaction version unknown")

	}

	return
}

func (tx *Transaction) GetPrefixHash() (result crypto.Hash) {
	result = crypto.Keccak256(tx.SerializeHeader())
	return result
}

// returns whether the tx is coinbase
func (tx *Transaction) IsCoinbase() (result bool) {
	switch tx.Vin[0].(type) {
	case Txin_gen:
		return true
	default:
		return false
	}
}

func (tx *Transaction) DeserializeHeader(buf []byte) (err error) {

	Key_offset_count := uint64(0) // used to calculate expected signatures in v1

	Mixin := -1

	tx.Clear() // clear existing

	//Mixin_count := 0 // for signature purpose

	done := 0
	tx.Version, done = binary.Uvarint(buf)
	if done <= 0 {
		return fmt.Errorf("Invalid Version in Transaction\n")
	}

	if tx.Version != 2 {
		return fmt.Errorf("Transaction version not equal to 2 \n")
	}
	rlog.Tracef(10, "transaction version %d\n", tx.Version)

	buf = buf[done:]
	tx.Unlock_Time, done = binary.Uvarint(buf)
	if done <= 0 {
		return fmt.Errorf("Invalid Unlock_Time in Transaction\n")
	}
	buf = buf[done:]

	// parse vin length
	vin_length, done := binary.Uvarint(buf)
	if done <= 0 {
		return fmt.Errorf("Invalid Vin length in Transaction\n")
	}
	buf = buf[done:]

	if vin_length == 0 {
		return fmt.Errorf("Vin input cannot be zero in Transaction\n")

	}

	rlog.Tracef(10, "vin length %d\n", vin_length)

	for i := uint64(0); i < vin_length; i++ {

		vin_type := buf[0]

		buf = buf[1:] // consume 1 more byte

		rlog.Tracef(10, "Processing i %d  vin_type %s  hex %x\n", i, TX_IN_NAME[vin_type], buf[:40])

		switch vin_type {
		case TXIN_GEN:
			rlog.Tracef(10, "Coinbase transaction\n")

			var current_vin Txin_gen
			current_vin.Height, done = binary.Uvarint(buf)
			if done <= 0 {
				return fmt.Errorf("Invalid Height for Txin_gen vin in Transaction\n")
			}
			buf = buf[done:]
			tx.Vin = append(tx.Vin, current_vin)

		/*case TXIN_TO_SCRIPT:
			panic("TXIN_TO_SCRIPT not implemented")
		case TXIN_TO_SCRIPTHASH:
			panic("TXIN_TO_SCRIPTHASH not implemented")
		*/
		case TXIN_TO_KEY:
			var current_vin Txin_to_key

			// parse Amount
			current_vin.Amount, done = binary.Uvarint(buf)
			if done <= 0 {
				return fmt.Errorf("Invalid Amount for Txin_to_key vin in Transaction\n")
			}
			buf = buf[done:]

			if current_vin.Amount != 0 {
				return fmt.Errorf("V2 Transactions have all input amount as zero\n")
			}

			//fmt.Printf("Remaining data %x\n", buf[:20]);

			mixin_count, done := binary.Uvarint(buf)
			if done <= 0 {
				return fmt.Errorf("Invalid offset_count for Txin_to_key vin in Transaction\n")
			}
			buf = buf[done:]

			// safety check mixin cannot be larger than say x
			// do we need a safety check
			if mixin_count > 4000 {
				return fmt.Errorf("Mixin cannot be larger than 400\n")
			}

			if Mixin < 0 {
				Mixin = int(mixin_count)
			}

			if Mixin != int(mixin_count) { // all vins must have same mixin
				return fmt.Errorf("Different mixin in Transaction\n")

			}

			//Mixin_input_count += Mixin

			for j := uint64(0); j < mixin_count; j++ {
				offset, done := binary.Uvarint(buf)
				if done <= 0 {
					return fmt.Errorf("Invalid key offset for Txin_to_key vin in Transaction\n")
				}
				buf = buf[done:]
				current_vin.Key_offsets = append(current_vin.Key_offsets, offset)
			}

			Key_offset_count += mixin_count

			copy(current_vin.K_image[:], buf[:32]) // copy key image

			buf = buf[32:] // consume key image bytes

			tx.Vin = append(tx.Vin, current_vin)

			// panic("TXIN_TO_KEY not implemented")

		default:
			return fmt.Errorf("Invalid VIN type in Transaction")

		}

	}

	//fmt.Printf("TX before vout %+v\n", tx)

	//fmt.Printf("buf before vout length %x\n", buf)
	vout_length, done := binary.Uvarint(buf)
	if done <= 0 {
		return fmt.Errorf("Invalid Vout length in Transaction\n")
	}
	buf = buf[done:]

	if vout_length == 0 {
		return fmt.Errorf("Vout cannot be zero in Transaction\n")
	}

	for i := uint64(0); i < vout_length; i++ {

		// amount is decoded earlier

		amount, done := binary.Uvarint(buf)
		if done <= 0 {
			return fmt.Errorf("Invalid Amount in Transaction\n")
		}
		buf = buf[done:]

		/*if amount != 0 {
		    return fmt.Errorf("V2 Transactions have all output amount as zero\n")
		}*/

		// decode vout type

		vout_type := buf[0]
		buf = buf[1:] // consume 1 more byte

		rlog.Tracef(10, "Vout Amount length %d  vout type %d   \n", amount, vout_type)

		/*if tx.Version == 2 && amount != 0 { // version 2 must have amount 0
			return fmt.Errorf("Amount must be zero in Transaction\n")
		}*/

		switch vout_type {
		/*case TXOUT_TO_SCRIPT:
			//fmt.Printf("out to script\n")
			panic("TXOUT_TO_SCRIPT not implemented")
		case TXOUT_TO_SCRIPTHASH:
			//fmt.Printf("out to scripthash\n")
			var current_vout Txout_to_scripthash
			copy(current_vout.Hash[:], buf[0:32])
			tx.Vout = append(tx.Vout, Tx_out{Amount: amount, Target: current_vout})

			buf = buf[32:]

			//panic("TXOUT_TO_SCRIPTHASH not implemented")
		*/
		case TXOUT_TO_KEY:
			//fmt.Printf("out to key\n")

			var current_vout Txout_to_key

			copy(current_vout.Key[:], buf[0:32])
			buf = buf[32:]

			//Mixin_input_count++

			tx.Vout = append(tx.Vout, Tx_out{Amount: amount, Target: current_vout})

		default:
			return fmt.Errorf("Invalid VOUT type in Transaction\n")

		}

	}

	// fmt.Printf("Extra %x\n", buf)
	// decode extra
	extra_length, done := binary.Uvarint(buf)
	if done <= 0 {
		return fmt.Errorf("Invalid Extra length in Transaction\n")
	}
	buf = buf[done:]

	// BUG extra needs to be processed in a loop till we load all extra fields

	//tx.ExtraType = buf[0]
	//    buf = buf[1:] // consume 1 more byte

	//    extra_length--

	rlog.Tracef(8, "extra len %d  have %d \n", extra_length, len(buf))
	tx.Extra = buf[:extra_length]

	// whatever is leftover is signature
	buf = buf[extra_length:] // consume more bytes

	switch tx.Version {
	/*case 1: // old style signatures, load value
	for i := uint64(0); i < Key_offset_count; i++ {
		var s Signature_v1
		copy(s.R[:], buf[:32])
		copy(s.C[:], buf[32:64])
		tx.Signature = append(tx.Signature, s)
		buf = buf[SIGNATURE_V1_LENGTH:]
	}
	*/

	case 2:
		bufreader := bytes.NewReader(buf)

		Mixin -= 1 // one is ours, rest are mixin

		tx.RctSignature, err = ringct.ParseRingCtSignature(bufreader, len(tx.Vin), len(tx.Vout), Mixin)
		if err != nil {
			return err
		}

		// we can expand and set some bulletproofs for later on verification,
		/*
					if tx.RctSignature.Get_Sig_Type() == ringct.RCTTypeSimpleBulletproof || tx.RctSignature.Get_Sig_Type() == ringct.RCTTypeFullBulletproof {

			                    if len(tx.RctSignature.ECdhInfo) != len(tx.Vout) ||len(tx.Vout) != len(tx.RctSignature.BulletSigs) {
			                        return fmt.Errorf("Invalid Bulletproof signature in Transaction\n")
			                    }
			                    for  i := range tx.RctSignature.ECdhInfo {
			                        tx.RctSignature.BulletSigs[i].V = []crypto.Key{ crypto.Key(tx.RctSignature.OutPk[i].Mask) }

			                        fmt.Printf("verifying BP now  \t ")

			                        if tx.RctSignature.BulletSigs[i].BULLETPROOF_Verify() {
			                         fmt.Printf("Bulletproof verification done successfully")
			                        }else{
			                            fmt.Printf("Bulletproof verification failed %+v \n", tx.RctSignature.BulletSigs[i])
			                        }
			                    }


			                }*/

	default:
		return fmt.Errorf("Version 1 is NOT supported\n")
	}

	/* we must deserialize signature some where else




	   //fmt.Printf("extra bytes %x\n",buf)

	   //fmt.Printf("signature len %d  should be %d\n",len(buf),len(tx.Vin)*SIGNATURE_V1_LENGTH)
	   fmt.Printf("signature len %d  should be %d\n",len(buf),Key_offset_count*SIGNATURE_V1_LENGTH)



	   switch tx.Version {
	       case 1 : // old style signatures, load value
	                for i := uint64(0); i < Key_offset_count;i++{
	                    var s Signature_v1
	                    copy(s.R[:],buf[:32])
	                    copy(s.C[:],buf[32:64])
	                    tx.Signature = append(tx.Signature, s)
	                    buf = buf[SIGNATURE_V1_LENGTH:]
	                }
	       case 2:
	                tx.Signature_RCT.Type, done = binary.Uvarint(buf)
	               if done <= 0 {
	                       return fmt.Errorf("Invalid RCT signature in Transaction\n")
	               }
	                   buf = buf[done:]

	                   switch tx.Signature_RCT.Type {

	                       case 0 : // no signature break

	                       case 1 :



	                   tx.Signature_RCT.TxnFee, done = binary.Uvarint(buf)
	               if done <= 0 {
	                       return fmt.Errorf("Invalid txn fee in Transaction\n")
	               }
	                   buf = buf[done:]

	                   fmt.Printf("RCT signature type %d  Fee %d\n",tx.Signature_RCT.Type, tx.Signature_RCT.TxnFee)


	                   // how many masked inputs depends on number of masked outouts
	                   for i := (0); i < len(tx.Vout);i++{
	                       // read masked input
	                       var info ECDHinfo
	                       copy(info.Mask[:], buf[0:32])
	                       copy(info.Amount[:], buf[32:64])
	                       tx.Signature_RCT.Amounts = append(tx.Signature_RCT.Amounts, info)
	                      buf = buf[64:]
	                   }

	                   // now parse the public keys
	                   for i := (0); i < len(tx.Vout);i++{
	                       // read masked input
	                       var tmp [32]byte
	                       copy(tmp[:], buf[0:32])

	                       tx.Signature_RCT.OutPK = append(tx.Signature_RCT.OutPK, tmp)
	                      buf = buf[32:]
	                   }

	                       case 2 : // panic("ringct type 2 currently not handled")


	                       default:
	                           panic("unknown signature style")


	                   }

	       default:
	           panic("unknown transaction version \n")



	   }


	*/

	rlog.Tracef(8, "TX deserialized %+v\n", tx)

	/*
		data.Local_time = binary.LittleEndian.Uint64( buf[24:], )

		 data.Local_Port = binary.LittleEndian.Uint32( buf[41:])

		_ = data.Network_UUID.UnmarshalBinary(buf[58:58+16])


		data.Peer_ID = binary.LittleEndian.Uint64( buf[83:] )
	*/
	return nil //fmt.Errorf("Done Transaction\n")

}

// calculated prefi has signature
func (tx *Transaction) Clear() {
	// clean the transaction everything
	tx.Version = 0
	tx.Unlock_Time = 0
	tx.Vin = tx.Vin[:0]
	tx.Vout = tx.Vout[:0]
	tx.Extra = tx.Extra[:0]

}

func (tx *Transaction) SerializeHeader() []byte {

	var serialised_header bytes.Buffer

	buf := make([]byte, binary.MaxVarintLen64)

	n := binary.PutUvarint(buf, tx.Version)
	serialised_header.Write(buf[:n])

	n = binary.PutUvarint(buf, tx.Unlock_Time)
	serialised_header.Write(buf[:n])

	if len(tx.Vin) < 1 {
		panic("No vins")
	}

	n = binary.PutUvarint(buf, uint64(len(tx.Vin)))
	serialised_header.Write(buf[:n])

	for _, current_vin := range tx.Vin {
		switch current_vin.(type) {
		case Txin_gen:
			serialised_header.WriteByte(TXIN_GEN)
			n = binary.PutUvarint(buf, current_vin.(Txin_gen).Height)
			serialised_header.Write(buf[:n])

		case Txin_to_key:
			serialised_header.WriteByte(TXIN_TO_KEY)
			n = binary.PutUvarint(buf, current_vin.(Txin_to_key).Amount)
			serialised_header.Write(buf[:n])

			// number of Ring member
			n = binary.PutUvarint(buf, uint64(len(current_vin.(Txin_to_key).Key_offsets)))
			serialised_header.Write(buf[:n])

			// write ring members
			for _, offset := range current_vin.(Txin_to_key).Key_offsets {
				n = binary.PutUvarint(buf, offset)
				serialised_header.Write(buf[:n])

			}

			// dump key image, interface needs a concrete type feor accessing array
			cvin := current_vin.(Txin_to_key)
			serialised_header.Write(cvin.K_image[:])

		}
	}

	// time to serialize vouts

	if len(tx.Vout) < 1 {
		panic("No vout")
	}

	n = binary.PutUvarint(buf, uint64(len(tx.Vout)))
	serialised_header.Write(buf[:n])

	for _, current_vout := range tx.Vout {

		// dump amount
		n := binary.PutUvarint(buf, current_vout.Amount)
		serialised_header.Write(buf[:n])

		switch current_vout.Target.(type) {
		case Txout_to_key:

			serialised_header.WriteByte(TXOUT_TO_KEY)

			target := current_vout.Target.(Txout_to_key)
			serialised_header.Write(target.Key[:])

			//serialised_header.Write(current_vout.Target.(Txout_to_key).Key[:])

		default:
			panic("This type of Txout not suppported")

		}
	}

	// dump any extras
	n = binary.PutUvarint(buf, uint64(len(tx.Extra)))
	serialised_header.Write(buf[:n])

	//rlog.Tracef("Extra length %d while serializing\n ", len(tx.Extra))

	serialised_header.Write(tx.Extra[:])

	return serialised_header.Bytes()

}

// serialize entire transaction include signature
func (tx *Transaction) Serialize() []byte {

	header_bytes := tx.SerializeHeader()
	base_bytes := tx.RctSignature.SerializeBase()
	prunable := tx.RctSignature.SerializePrunable()

	buf := append(header_bytes, base_bytes...)
	buf = append(buf, prunable...)

	return buf

}

/*

func (tx *Transaction) IsCoinbase() (result bool){

  // check whether the type is Txin.get

   if len(tx.Vin) != 0 { // coinbase transactions have no vin
    return
   }

   if tx.Vout[0].(Target) != 0 { // coinbase transactions have no vin
    return
   }


}*/

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

package block

import "fmt"

//import "sort"
import "bytes"
import "runtime/debug"
import "encoding/hex"
import "encoding/binary"

import "github.com/ebfe/keccak"
import "github.com/romana/rlog"

import "github.com/deroproject/derosuite/crypto"

//import "github.com/deroproject/derosuite/config"
import "github.com/deroproject/derosuite/cryptonight"
import "github.com/deroproject/derosuite/transaction"

// these are defined  in file
//https://github.com/monero-project/monero/src/cryptonote_basic/cryptonote_basic.h
type Block_Header struct {
	Major_Version uint64                  `json:"major_version"`
	Minor_Version uint64                  `json:"minor_version"`
	Timestamp     uint64                  `json:"timestamp"`
	Nonce         uint32                  `json:"nonce"` // TODO make nonce 32 byte array for infinite work capacity
	ExtraNonce    [32]byte                `json:"-"`
	Miner_TX      transaction.Transaction `json:"miner_tx"`
}

type Block struct {
	Block_Header
	Proof     [32]byte      `json:"-"` // Reserved for future
	Tips      []crypto.Hash `json:"tips"`
	Tx_hashes []crypto.Hash `json:"tx_hashes"`
}

// we process incoming blocks in this format
type Complete_Block struct {
	Bl  *Block
	Txs []*transaction.Transaction
}

// see spec here https://cryptonote.org/cns/cns003.txt
// this function gets the block identifier hash
// this has been simplified and varint length has been removed
func (bl *Block) GetHash() (hash crypto.Hash) {
	long_header := bl.GetBlockWork()

	// keccak hash of this above blob, gives the block id
	return crypto.Keccak256(long_header)
}

// converts a block, into a getwork style work, ready for either submitting the block
// or doing Pow Calculations
func (bl *Block) GetBlockWork() []byte {

	var buf []byte // bitcoin/litecoin getworks are 80 bytes
	var scratch [8]byte

	buf = append(buf, []byte{byte(bl.Major_Version), byte(bl.Minor_Version), 0, 0, 0, 0, 0}...) // 0 first 7 bytes are version in little endia format

	binary.LittleEndian.PutUint32(buf[2:6], uint32(bl.Timestamp))
	header_hash := crypto.Keccak256(bl.getserializedheaderforwork()) // 0 + 7

	buf = append(buf, header_hash[:]...) // 0 + 7 + 32  = 39

	binary.LittleEndian.PutUint32(scratch[0:4], bl.Nonce) // check whether it needs to be big endian
	buf = append(buf, scratch[:4]...)                     // 0 + 7 + 32  + 4 = 43

	// next place the ExtraNonce
	buf = append(buf, bl.ExtraNonce[:]...) // total 7 + 32 + 4 + 32

	buf = append(buf, 0) // total 7 + 32 + 4 + 32 + 1 = 76

	if len(buf) != 76 {
		panic(fmt.Sprintf("Getwork not equal to  76 bytes  actual %d", len(buf)))

	}
	return buf
}

// copy the nonce and the extra nonce from the getwork to the block
func (bl *Block) CopyNonceFromBlockWork(work []byte) (err error) {
	if len(work) < 74 {
		return fmt.Errorf("work buffer is Invalid")
	}

	bl.Timestamp = uint64(binary.LittleEndian.Uint32(work[2:]))
	bl.Nonce = binary.LittleEndian.Uint32(work[7+32:])
	copy(bl.ExtraNonce[:], work[7+32+4:75])
	return
}

// copy the nonce and the extra nonce from the getwork to the block
func (bl *Block) SetExtraNonce(extranonce []byte) (err error) {

	if len(extranonce) == 0 {
		return fmt.Errorf("extra nonce is invalid")
	}
	max := len(extranonce)
	if max > 32 {
		max = 32
	}
	copy(bl.ExtraNonce[:], extranonce[0:max])
	return
}

// clear extra nonce
func (bl *Block) ClearExtraNonce() {
	for i := range bl.ExtraNonce {
		bl.ExtraNonce[i] = 0
	}
}

// clear nonce
func (bl *Block) ClearNonce() {
	bl.Nonce = 0
}

// Get PoW hash , this is very slow function
func (bl *Block) GetPoWHash() (hash crypto.Hash) {
	long_header := bl.GetBlockWork()
	rlog.Tracef(9, "longheader %x\n", long_header)
	tmphash := cryptonight.SlowHash(long_header)
	//   tmphash := crypto.Scrypt_1024_1_1_256(long_header)
	copy(hash[:], tmphash[:32])

	return
}

// serialize block header for calculating PoW
func (bl *Block) getserializedheaderforwork() []byte {
	var serialised bytes.Buffer

	buf := make([]byte, binary.MaxVarintLen64)

	n := binary.PutUvarint(buf, uint64(bl.Major_Version))
	serialised.Write(buf[:n])

	n = binary.PutUvarint(buf, uint64(bl.Minor_Version))
	serialised.Write(buf[:n])

	// it is placed in pow
	//n = binary.PutUvarint(buf, bl.Timestamp)
	//serialised.Write(buf[:n])

	// write miner tx
	serialised.Write(bl.Miner_TX.Serialize())

	// write tips,, merkle tree should be replaced with something faster
	tips_treehash := bl.GetTipsHash()
	n = binary.PutUvarint(buf, uint64(len(tips_treehash)))
	serialised.Write(buf[:n])
	serialised.Write(tips_treehash[:]) // actual tips hash

	tx_treehash := bl.GetTXSHash()                        // hash of all transactions
	n = binary.PutUvarint(buf, uint64(len(bl.Tx_hashes))) // count of all transactions
	serialised.Write(buf[:n])
	serialised.Write(tx_treehash[:]) // actual transactions hash

	return serialised.Bytes()
}

// serialize block header
func (bl *Block) SerializeHeader() []byte {

	var serialised bytes.Buffer

	buf := make([]byte, binary.MaxVarintLen64)

	n := binary.PutUvarint(buf, uint64(bl.Major_Version))
	serialised.Write(buf[:n])

	n = binary.PutUvarint(buf, uint64(bl.Minor_Version))
	serialised.Write(buf[:n])

	n = binary.PutUvarint(buf, bl.Timestamp)
	serialised.Write(buf[:n])

	binary.LittleEndian.PutUint32(buf[0:8], bl.Nonce) // check whether it needs to be big endian
	serialised.Write(buf[:4])

	serialised.Write(bl.ExtraNonce[:])

	// write miner address
	serialised.Write(bl.Miner_TX.Serialize())

	return serialised.Bytes()

}

// serialize entire block ( block_header + miner_tx + tx_list )
func (bl *Block) Serialize() []byte {
	var serialized bytes.Buffer
	buf := make([]byte, binary.MaxVarintLen64)

	header := bl.SerializeHeader()
	serialized.Write(header)

	serialized.Write(bl.Proof[:]) // write proof  NOT implemented

	// miner tx should always be coinbase
	//minex_tx := bl.Miner_tx.Serialize()
	//serialized.Write(minex_tx)

	n := binary.PutUvarint(buf, uint64(len(bl.Tips)))
	serialized.Write(buf[:n])

	for _, hash := range bl.Tips {
		serialized.Write(hash[:])
	}

	//fmt.Printf("serializing tx hashes %d\n", len(bl.Tx_hashes))

	n = binary.PutUvarint(buf, uint64(len(bl.Tx_hashes)))
	serialized.Write(buf[:n])

	for _, hash := range bl.Tx_hashes {
		serialized.Write(hash[:])
	}

	return serialized.Bytes()
}

// get block transactions tree hash
func (bl *Block) GetTipsHash() (result crypto.Hash) {

	/*if len(bl.Tips) == 0 { // case for genesis block
	  panic("Block does NOT refer any tips")
	 }*/

	// add all the remaining hashes
	h := keccak.New256()
	for i := range bl.Tips {
		h.Write(bl.Tips[i][:])
	}
	r := h.Sum(nil)
	copy(result[:], r)
	return
}

// get block transactions
// we have discarded the merkle tree and have shifted to a plain version
func (bl *Block) GetTXSHash() (result crypto.Hash) {
	h := keccak.New256()
	for i := range bl.Tx_hashes {
		h.Write(bl.Tx_hashes[i][:])
	}
	r := h.Sum(nil)
	copy(result[:], r)

	return
}

// only parses header
func (bl *Block) DeserializeHeader(buf []byte) (err error) {
	done := 0
	bl.Major_Version, done = binary.Uvarint(buf)
	if done <= 0 {
		return fmt.Errorf("Invalid Major Version in Block\n")
	}
	buf = buf[done:]

	bl.Minor_Version, done = binary.Uvarint(buf)
	if done <= 0 {
		return fmt.Errorf("Invalid Minor Version in Block\n")
	}
	buf = buf[done:]

	bl.Timestamp, done = binary.Uvarint(buf)
	if done <= 0 {
		return fmt.Errorf("Invalid Timestamp in Block\n")
	}
	buf = buf[done:]

	//copy(bl.Prev_Hash[:], buf[:32]) // hash is always 32 byte
	//buf = buf[32:]

	bl.Nonce = binary.LittleEndian.Uint32(buf)

	buf = buf[4:]

	copy(bl.ExtraNonce[:], buf[0:32])

	buf = buf[32:]

	// parse miner tx
	err = bl.Miner_TX.DeserializeHeader(buf)
	if err != nil {
		return err
	}

	return
}

//parse entire block completely
func (bl *Block) Deserialize(buf []byte) (err error) {
	done := 0

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic while deserialising block, block hex_dump below to make a testcase/debug\n")
			fmt.Printf("%s\n", hex.EncodeToString(buf))

			fmt.Printf("Recovered while parsing block, Stack trace below block_hash ")
			fmt.Printf("Stack trace  \n%s", debug.Stack())
			err = fmt.Errorf("Invalid Block")
			return
		}
	}()

	err = bl.DeserializeHeader(buf)
	if err != nil {
		return fmt.Errorf("Block Header could not be parsed\n")
	}

	buf = buf[len(bl.SerializeHeader()):] // skup number of bytes processed

	// read 32 byte proof
	copy(bl.Proof[:], buf[0:32])
	buf = buf[32:]

	// header finished here

	// read and parse transaction
	/*err = bl.Miner_tx.DeserializeHeader(buf)

	if err != nil {
		return fmt.Errorf("Cannot parse miner TX  %x", buf)
	}

	// if tx was parse, make sure it's coin base
	if len(bl.Miner_tx.Vin) != 1 || bl.Miner_tx.Vin[0].(transaction.Txin_gen).Height > config.MAX_CHAIN_HEIGHT {
		// serialize transaction again to get the tx size, so as parsing could continue
		return fmt.Errorf("Invalid Miner TX")
	}

	miner_tx_serialized_size := bl.Miner_tx.Serialize()
	buf = buf[len(miner_tx_serialized_size):]
	*/

	tips_count, done := binary.Uvarint(buf)
	if done <= 0 || done > 1 {
		return fmt.Errorf("Invalid Tips count in Block\n")
	}
	buf = buf[done:]

	// remember first tx is merkle root

	for i := uint64(0); i < tips_count; i++ {
		//fmt.Printf("Parsing transaction hash %d  tx_count %d\n", i, tx_count)
		var h crypto.Hash
		copy(h[:], buf[:32])
		buf = buf[32:]

		bl.Tips = append(bl.Tips, h)

	}

	//fmt.Printf("miner tx %x\n", miner_tx_serialized_size)
	// read number of transactions
	tx_count, done := binary.Uvarint(buf)
	if done <= 0 {
		return fmt.Errorf("Invalid Tx count in Block\n")
	}
	buf = buf[done:]

	// remember first tx is merkle root

	for i := uint64(0); i < tx_count; i++ {
		//fmt.Printf("Parsing transaction hash %d  tx_count %d\n", i, tx_count)
		var h crypto.Hash
		copy(h[:], buf[:32])
		buf = buf[32:]

		bl.Tx_hashes = append(bl.Tx_hashes, h)

	}

	//fmt.Printf("%d member in tx hashes \n",len(bl.Tx_hashes))

	return

}

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
import "bytes"
import "encoding/hex"
import "encoding/binary"

import "github.com/romana/rlog"

import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/config"
import "github.com/deroproject/derosuite/cryptonight"
import "github.com/deroproject/derosuite/transaction"

// these are defined  in file
//https://github.com/monero-project/monero/src/cryptonote_basic/cryptonote_basic.h
type Block_Header struct {
	Major_Version uint32      `json:"major_version"`
	Minor_Version uint32      `json:"minor_version"`
	Timestamp     uint64      `json:"timestamp"`
	Prev_Hash     crypto.Hash `json:"prev_id"`
	Nonce         uint32      `json:"nonce"`
}

type Block struct {
	Block_Header
	Miner_tx    transaction.Transaction `json:"miner_tx"`
	Merkle_Root crypto.Hash             `json:"-"`
	Tx_hashes   []crypto.Hash           `json:"tx_hashes"`

	treehash crypto.Hash
}

// we process incoming blocks in this format
type Complete_Block struct {
	Bl  *Block
	Txs []*transaction.Transaction
}

// see spec here https://cryptonote.org/cns/cns003.txt
// this function gets the block identifier hash
func (bl *Block) GetHash() (hash crypto.Hash) {
	buf := make([]byte, binary.MaxVarintLen64)
	long_header := bl.GetBlockWork()
	length := uint64(len(long_header))
	n := binary.PutUvarint(buf, length) //
	buf = buf[:n]
	block_id_blob := append(buf, long_header...)

	// keccak hash of this above blob, gives the block id
	hash2 := crypto.Keccak256(block_id_blob)
	return crypto.Hash(hash2)
}

// converts a block, into a getwork style work, ready for either submitting the block
// or doing Pow Calculations
func (bl *Block) GetBlockWork() []byte {
	buf := make([]byte, binary.MaxVarintLen64)
	header := bl.SerializeHeader()
	tx_treehash := bl.GetTreeHash() // treehash of all transactions

	// length of all transactions
	n := binary.PutUvarint(buf, uint64(len(bl.Tx_hashes)+1)) // +1 for miner TX
	buf = buf[:n]

	long_header := append(header, tx_treehash[:]...)
	long_header = append(long_header, buf...)

	return long_header

}

// Get PoW hash , this is very slow function
func (bl *Block) GetPoWHash() (hash crypto.Hash) {
	long_header := bl.GetBlockWork()
	rlog.Tracef(9, "longheader %x\n", long_header)
	tmphash := cryptonight.SlowHash(long_header)
	copy(hash[:], tmphash[:32])

	return
}

// Reward is, total amount in the miner tx - fees
func (bl *Block) GetReward() uint64 {
	total_amount := bl.Miner_tx.Vout[0].Amount
	total_fees := uint64(0)
	// load all the TX and get the fees, since we are in a post rct world
	// extract the fees from the rct sig
	return total_amount - total_fees
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

	serialised.Write(bl.Prev_Hash[:32]) // write previous ID

	binary.LittleEndian.PutUint32(buf[0:8], bl.Nonce) // check whether it needs to be big endian
	serialised.Write(buf[:4])

	return serialised.Bytes()

}

// serialize entire block ( block_header + miner_tx + tx_list )
func (bl *Block) Serialize() []byte {
	var serialized bytes.Buffer
	buf := make([]byte, binary.MaxVarintLen64)

	header := bl.SerializeHeader()
	serialized.Write(header)

	// miner tx should always be coinbase
	minex_tx := bl.Miner_tx.Serialize()
	serialized.Write(minex_tx)

	//fmt.Printf("serializing tx hashes %d\n", len(bl.Tx_hashes))

	n := binary.PutUvarint(buf, uint64(len(bl.Tx_hashes)))
	serialized.Write(buf[:n])

	for _, hash := range bl.Tx_hashes {
		serialized.Write(hash[:])
	}

	return serialized.Bytes()
}

// get block transactions tree hash
func (bl *Block) GetTreeHash() (hash crypto.Hash) {
	var hash_list []crypto.Hash

	hash_list = append(hash_list, bl.Miner_tx.GetHash())
	// add all the remaining hashes
	for i := range bl.Tx_hashes {
		hash_list = append(hash_list, bl.Tx_hashes[i])
	}

	return TreeHash(hash_list)

}

// input is the list of transactions hashes
func TreeHash(hashes []crypto.Hash) (hash crypto.Hash) {

	switch len(hashes) {
	case 0:
		panic("Treehash cannot have 0 transactions, atleast miner tx will be present")
	case 1:
		copy(hash[:], hashes[0][:32])
	case 2:
		var buf []byte
		for i := 0; i < len(hashes); i++ {
			buf = append(buf, hashes[i][:32]...)
		}
		tmp_hash := crypto.Keccak256(buf)
		copy(hash[:], tmp_hash[:32])

	default:

		count := uint64(len(hashes))
		cnt := tree_hash_cnt(count)

		//fmt.Printf("cnt %d count %d\n",cnt, count)
		ints := make([]byte, 32*cnt, 32*cnt)

		hashes_buf := make([]byte, 32*count, 32*count)

		for i := uint64(0); i < count; i++ {
			copy(hashes_buf[i*32:], hashes[i][:32]) // copy hashes 1 by 1
		}

		for i := uint64(0); i < ((2 * cnt) - count); i++ {
			copy(ints[i*32:], hashes[i][:32]) // copy hashes 1 by 1
		}

		i := ((2 * cnt) - count)
		j := ((2 * cnt) - count)
		for ; j < cnt; i, j = i+2, j+1 {
			hash := crypto.Keccak256(hashes_buf[i*32 : (i*32)+64]) // find hash of 64 bytes
			copy(ints[j*32:], hash[:32])
		}
		if i != count {
			panic("please fix tree hash")
		}

		for cnt > 2 {
			cnt = cnt >> 1
			i = 0
			j = 0
			for ; j < cnt; i, j = i+2, j+1 {
				hash := crypto.Keccak256(ints[i*32 : (i*32)+64]) // find hash of 64 bytes
				copy(ints[j*32:], hash[:32])
			}
		}

		hash = crypto.Hash(crypto.Keccak256(ints[0:64])) // find hash of 64 bytes
	}
	return
}

// see crypto/tree-hash.c
// this function has a naughty history
func tree_hash_cnt(count uint64) uint64 {
	pow := uint64(2)
	for pow < count {
		pow = pow << 1
	}
	return pow >> 1
}

func (bl *Block) Deserialize(buf []byte) (err error) {
	done := 0
	var tmp uint64

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Panic while deserialising block, block hex_dump below to make a testcase/debug\n")
			fmt.Printf("%s", hex.EncodeToString(buf))
			err = fmt.Errorf("Invalid Block")
			return
		}
	}()

	tmp, done = binary.Uvarint(buf)
	if done <= 0 {
		return fmt.Errorf("Invalid Version in Block\n")
	}
	buf = buf[done:]

	bl.Major_Version = uint32(tmp)

	if uint64(bl.Major_Version) != tmp {
		return fmt.Errorf("Invalid Block major version")
	}

	tmp, done = binary.Uvarint(buf)
	if done <= 0 {
		return fmt.Errorf("Invalid minor Version in Block\n")
	}
	buf = buf[done:]

	bl.Minor_Version = uint32(tmp)

	if uint64(bl.Minor_Version) != tmp {
		return fmt.Errorf("Invalid Block minor version")
	}

	bl.Timestamp, done = binary.Uvarint(buf)
	if done <= 0 {
		return fmt.Errorf("Invalid Timestamp in Block\n")
	}
	buf = buf[done:]

	copy(bl.Prev_Hash[:], buf[:32]) // hash is always 32 byte
	buf = buf[32:]

	bl.Nonce = binary.LittleEndian.Uint32(buf)
	buf = buf[4:] // nonce is always 4 bytes

	// read and parse transaction
	err = bl.Miner_tx.DeserializeHeader(buf)

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

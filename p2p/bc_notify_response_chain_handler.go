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

package p2p

import "bytes"
import "encoding/binary"

import "github.com/romana/rlog"

import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/crypto"

/* the data structure which needs to be serialised is    defined in cryptonote_protocol_defs.h
* struct request
   {
     uint64_t start_height;
     uint64_t total_height;
     uint64_t cumulative_difficulty;
     std::list<crypto::hash> m_block_ids;

     BEGIN_KV_SERIALIZE_MAP()
       KV_SERIALIZE(cumulative_difficulty)
       KV_SERIALIZE_CONTAINER_POD_AS_BLOB(m_block_ids)
       KV_SERIALIZE(start_height)
       KV_SERIALIZE(total_height)
     END_KV_SERIALIZE_MAP()
   };
*/

// This request only comes  when we have sent the BC_NOTIFY_REQUEST_CHAIN command
// handle BC_NOTIFY_RESPONSE_CHAIN_ENTRY
func Handle_BC_Notify_Response_Chain_Entry(connection *Connection,
	i_command_header *Levin_Header, buf []byte) {

	// deserialize data header
	var i_data_header Levin_Data_Header // incoming data header

	err := i_data_header.DeSerialize(buf)

	if err != nil {
		//fmt.Printf("We should destroy connection here, data header cnot deserialized")
		rlog.Tracef(4, "Data header deserialisation failed. Disconnect peer \n")
		connection.Exit = true
		return
	}

	pos := bytes.Index(i_data_header.Data, []byte("cumulative_difficulty")) // at this point to
	if pos == -1 {
		rlog.Tracef(4, "Cumulative difficulty deserialisation failed. Disconnect peer \n")
		connection.Exit = true
		return
	}

	cumulative_difficulty := binary.LittleEndian.Uint64(i_data_header.Data[pos+22:])

	rlog.Tracef(4, "Cumalative difficulty %d  %x\n", cumulative_difficulty, cumulative_difficulty)

	pos = bytes.Index(i_data_header.Data, []byte("start_height")) // at this point to
	if pos == -1 {
		panic("start_height not found, its mandatory\n")

	}

	start_height := binary.LittleEndian.Uint64(i_data_header.Data[pos+13:])

	rlog.Tracef(4, "start_height %d  %x\n", start_height, start_height)

	pos = bytes.Index(i_data_header.Data, []byte("m_block_ids")) // at this point to
	if pos == -1 {
		panic("m_block_ids not found, its mandatory\n")
	}

	// decode  data length ( though we know it from buffer size, but still verify it )
	buf = i_data_header.Data[pos+11+1:]
	data_length, done := Decode_Boost_Varint(buf)
	//fmt.Printf("data length %d ,   hex  %x\n", data_length, buf[:8])
	buf = buf[done:]

	if data_length == 0 {
		rlog.Tracef(4, "Peer says it does not have even genesis block, so disconnect")
		connection.Exit = true
		return
	}
	if (data_length % 32) != 0 { // sanity check
		rlog.Tracef(2, "We should destroy connection here, packet mismatch")
		connection.Exit = true
		return
	}

	rlog.Tracef(4, "Number of Blocks id in chain BC_NOTIFY_RESPONSE_CHAIN_ENTRY %d \n", data_length/32)

	var block_list []crypto.Hash

	for i := uint64(0); i < data_length/32; i++ {
		var bhash crypto.Hash
		copy(bhash[:], buf[i*32:(i+1)*32])
		// only request block that we donot have

		if chain.Get_Height() < 20 {
			block_list = append(block_list, bhash)
			rlog.Tracef(5, "%2d hash  %x\n", i, bhash[:])
		} else {
			if !chain.Block_Exists(bhash) {
				block_list = append(block_list, bhash)
				rlog.Tracef(5, "%2d hash  %x\n", i, bhash[:])
			}
		}

	}

	// server will kill us, if we queue more than 1000 blocks
	if len(block_list) > 900 {
		block_list = block_list[:899]
	}

	// make sure the genesis block is same
	// if peer provided us a genesis block make sure, its ours
	if start_height == 0 && block_list[0] != globals.Config.Genesis_Block_Hash {
		rlog.Tracef(4, "Peer's genesis block is different from our, so disconnect")
		connection.Exit = true
		return
	}

	// we must queue the hashes so as to fetch them

	//block_list = block_list[:0]
	var hash crypto.Hash

	/*
	   //big_block ,_ := hex.DecodeString("9ba23efe505f9674dc24c150edbdbe57abc3ec6636aa4c1659e811b389c0b30b") // zero tx
	   //big_block ,_ := hex.DecodeString("14371eeddca0f3ce9b992b3e2a0e482920497d87dd20002456f3a844b04a3318") // single tx

	   big_block ,_ := hex.DecodeString("a31a17bb26b2ec37479ee3a02f53dd94860611457811d0c23ce892f4e87e1697") // 4 tx

	   copy(hash[:],big_block[:32])
	   block_list = append(block_list, hash)


	   big_tx, _ := hex.DecodeString("af6f12d56f32f58623a834c4f12c5443976346a2f877ef78798c39496bd00559") //  single tx





	   var tx_list []ringct.Hash
	   copy(hash[:],big_tx[:32])
	   tx_list = append(tx_list, hash)
	*/
	_ = hash
	var tx_list []crypto.Hash

	/*if len(block_list) > 5 {
	    block_list= block_list[:5]
	}*/

	Send_BC_Notify_Request_GetObjects(connection, block_list, tx_list[:0])

}

// header from boost packet
/* header bytes
0000   01 11 01 01 01 01 02 01 01 10 15 63 75 6d 75 6c  ...........cumul
0010   61 74 69 76 65 5f 64 69 66 66 69 63 75 6c 74 79  ative_difficulty
0020   05 ab 61 00 00 00 00 00 00 0b 6d 5f 62 6c 6f 63  ..a.......m_bloc
0030   6b 5f 69 64 73 0a 82 72 02 00 63 34 12 de 21 ea  k_ids..r..c4..!.


// suffix_bytes
44d0   4c 16 59 e8 11 b3 89 c0 b3 0b 0c 73 74 61 72 74  L.Y........start
44e0   5f 68 65 69 67 68 74 05 00 00 00 00 00 00 00 00  _height.........
44f0   0c 74 6f 74 61 6c 5f 68 65 69 67 68 74 05 e5 04  .total_height...
4500   00 00 00 00 00 00                                ......
*/
// send the PEER the blocks he needs to download our version of chain
// this is send every 5 seconds
func Send_BC_Notify_Response_Chain_Entry(connection *Connection, block_list []crypto.Hash, start_height, current_height, diff uint64) {

	connection.Lock()

	header_bytes := []byte{0x01, 0x11, 0x01, 0x01, 0x01, 0x01, 0x02, 0x01, 0x01, 0x10, 0x15, 0x63, 0x75, 0x6d, 0x75, 0x6c,
		0x61, 0x74, 0x69, 0x76, 0x65, 0x5f, 0x64, 0x69, 0x66, 0x66, 0x69, 0x63, 0x75, 0x6c, 0x74, 0x79,
		0x05, 0xab, 0x61, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0b, 0x6d, 0x5f, 0x62, 0x6c, 0x6f, 0x63,
		0x6b, 0x5f, 0x69, 0x64, 0x73, 0x0a}

	suffix_bytes := []byte{0x0c, 0x73, 0x74, 0x61, 0x72, 0x74,
		0x5f, 0x68, 0x65, 0x69, 0x67, 0x68, 0x74, 0x05, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x0c, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x5f, 0x68, 0x65, 0x69, 0x67, 0x68, 0x74, 0x05, 0xe5, 0x04,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	binary.LittleEndian.PutUint64(header_bytes[33:], diff)
	binary.LittleEndian.PutUint64(suffix_bytes[14:], start_height)
	binary.LittleEndian.PutUint64(suffix_bytes[36:], current_height)

	// now append boost variant length, and then append all the hashes

	buf := make([]byte, 8, 8)
	done := Encode_Boost_Varint(buf, uint64(len(block_list)*32)) // encode length of buffer
	buf = buf[:done]

	var o_command_header Levin_Header

	data_bytes := append(header_bytes, buf...)

	// convert and append all hashes to bytes
	for _, hash := range block_list {
		data_bytes = append(data_bytes, hash[:32]...)
	}

	data_bytes = append(data_bytes, suffix_bytes...)

	o_command_header.CB = uint64(len(data_bytes))

	o_command_header.Command = BC_NOTIFY_RESPONSE_CHAIN_ENTRY
	o_command_header.ReturnData = false
	o_command_header.Flags = LEVIN_PACKET_REQUEST

	o_command_header_bytes, _ := o_command_header.Serialize()

	connection.Conn.Write(o_command_header_bytes)
	connection.Conn.Write(data_bytes)

	//fmt.Printf("len of command header %d\n", len(o_command_header_bytes))
	//fmt.Printf("len of data header %d\n", len(data_bytes))

	connection.Unlock()
}

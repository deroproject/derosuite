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

import "github.com/romana/rlog"

import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/crypto"

// when 2 peers communiate either both are in sync or async
// if async, reply to the request below, with your state and other will supply you  list of block ids
// handle BC_NOTIFY_REQUEST_CHAIN
func Handle_BC_Notify_Chain(connection *Connection,
	i_command_header *Levin_Header, buf []byte) {

	// deserialize data header
	var i_data_header Levin_Data_Header // incoming data header

	err := i_data_header.DeSerialize(buf)

	if err != nil {
		connection.logger.Debugf("We should destroy connection here, data header cnot deserialized")
		connection.Exit = true
		return
	}

	buf = i_data_header.Data[11:] // 11 bytes boost header, ignore it

	// decode remain data length ( though we know it from buffer size, but still verify it )

	data_length, done := Decode_Boost_Varint(buf)
	buf = buf[done:]

	if data_length == 0 {
		rlog.Tracef(4, "Peer says it does not have even genesis block, so disconnect")
		connection.Exit = true
		return
	}
	if (data_length % 32) != 0 { // sanity check
		rlog.Tracef(4, "We should destroy connection here, packet mismatch")
		connection.Exit = true
		return
	}

	rlog.Tracef(4, "Number of hashes %d \n", data_length/32)

	var block_list []crypto.Hash

	for i := uint64(0); i < data_length/32; i++ {
		var bhash crypto.Hash
		copy(bhash[:], buf[i*32:(i+1)*32])

		block_list = append(block_list, bhash)

		rlog.Tracef(5, "%2d hash  %x\n", i, bhash[:])

	}

	// the data is like this, first 10 blocks, then block are in 2^n power and the last block is genesis

	// make sure the genesis block is same

	if block_list[len(block_list)-1] != globals.Config.Genesis_Block_Hash {
		connection.logger.Debugf("Peer's genesis block is different from our, so disconnect")
		connection.Exit = true
		return
	}

	// we must give user our version of the chain
	start_height := uint64(0)

	for i := 0; i < len(block_list); i++ { // find the common point in our chain
		if chain.Block_Exists(block_list[i]) {
			start_height = chain.Load_Height_for_BL_ID(block_list[i])
			rlog.Tracef(4, "Found common point in chain at hash %x\n", block_list[i])
			break
		}
	}

	// send atleast 16001 block or till the top
	stop_height := chain.Get_Height()

	if (stop_height - start_height) > 1001 { // send MAX 512 KB block hashes
		stop_height = start_height + 1002
	}

	block_list = block_list[:0]

	for i := start_height; i < stop_height; i++ {
		hash, _ := chain.Load_BL_ID_at_Height(i)
		block_list = append(block_list, hash)
	}

	rlog.Tracef(2, "Prepared list of %d block header to send \n", len(block_list))

	Send_BC_Notify_Response_Chain_Entry(connection, block_list, start_height, chain.Get_Height(), 1)

}

// header from boost packet
//0060   00 00 00 01 11 01 01 01 01 02 01 [] 01 04 09 62 6c  ..............bl
//0070   6f 63 6b 5f 69 64 73 0a [] 81 0a 9b a2 3e fe 50 5f  ock_ids.....>.P_

// send the Peer our chain status, so he can give us the latest chain or updated block_ids
// this is only sent when we are in different from peers
func Send_BC_Notify_Chain_Command(connection *Connection) {

	connection.Lock()

	header_bytes := []byte{ /*0x01, 0x04,*/ 0x09, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x5f, 0x69, 0x64, 0x73, 0x0a}

	_ = header_bytes
	// now append boost variant length, and then append all the hashes

	var block_list []crypto.Hash

	// add your blocks here

	our_height := chain.Get_Height()
	if our_height < 20 { // if height is less than 20, fetch from genesis block

	} else { // send blocks in reverse
		for i := uint64(1); our_height < 11; i++ {
			hash, err := chain.Load_BL_ID_at_Height(our_height - i)
			_ = err
			block_list = append(block_list, hash)
		}

		our_height = our_height - 11

		// now seend all our block id in log, 2nd, 4th, 8th,  etc
		for ; our_height > 0; our_height = our_height >> 1 {
			hash, err := chain.Load_BL_ID_at_Height(our_height)
			_ = err
			block_list = append(block_list, hash)

		}
	}

	// final block is always genesis block
	block_list = append(block_list, globals.Config.Genesis_Block_Hash)

	buf := make([]byte, 8, 8)
	done := Encode_Boost_Varint(buf, uint64(len(block_list)*32)) // encode length of buffer
	buf = buf[:done]

	var o_command_header Levin_Header
	var o_data_header Levin_Data_Header

	o_data_header.Data = append(header_bytes, buf...)

	// convert and append all hashes to bytes
	for _, hash := range block_list {
		o_data_header.Data = append(o_data_header.Data, hash[:32]...)
	}

	o_data_bytes, _ := o_data_header.Serialize()

	o_data_bytes[9] = 0x4

	o_command_header.CB = uint64(len(o_data_bytes))

	o_command_header.Command = BC_NOTIFY_REQUEST_CHAIN
	o_command_header.ReturnData = false
	o_command_header.Flags = LEVIN_PACKET_REQUEST

	o_command_header_bytes, _ := o_command_header.Serialize()

	connection.Conn.Write(o_command_header_bytes)
	connection.Conn.Write(o_data_bytes)

	//fmt.Printf("len of command header %d\n", len(o_command_header_bytes))
	//fmt.Printf("len of data header %d\n", len(o_data_bytes))

	connection.Unlock()
}

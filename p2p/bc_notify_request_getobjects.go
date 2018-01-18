package p2p


import "fmt"
import "bytes"
import "encoding/binary"

import "github.com/romana/rlog"

import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/blockchain"




// The peer triggers this it wants some blocks or txs
func Handle_BC_Notify_Request_GetObjects(connection *Connection,
	i_command_header *Levin_Header, buf []byte) {

	// deserialize data header
	var i_data_header Levin_Data_Header // incoming data header

	err := i_data_header.DeSerialize(buf)

	if err != nil {
		connection.logger.Debugf("We should destroy connection here, data header cnot deserialized")
		connection.Exit = true
		return
	}

	pos := bytes.Index(buf, []byte("\x06blocks\x0a")) // at this point to
	if pos < 0 {
		rlog.Tracef(4, "NOTIFY_REQUEST_GET_OBJECTS doesnot contains blocks. Disconnect peer \n")
	} else { // we have some block ids, extract and serve them

		tmp_slice := buf[pos+8:]
		rlog.Tracef(4, " varint %x", tmp_slice[:8])

		data_length, done := Decode_Boost_Varint(tmp_slice)
		tmp_slice = tmp_slice[done:]

		if data_length == 0 {
			rlog.Tracef(4, "Peer says it does not have even genesis block, so disconnect")
			connection.Exit = true
		        return
		}
		rlog.Tracef(4, "Data size %d", data_length)

		if (data_length % 32) != 0 { // sanity check
			rlog.Tracef(4, "We should destroy connection here, packet mismatch")
			connection.Exit = true
		        return
		}

		rlog.Tracef(4, "Number of hashes %d tmp_slice %x \n", data_length/32, tmp_slice[:32])

		var block_list []crypto.Hash

		for i := uint64(0); i < data_length/32; i++ {
			var bhash crypto.Hash
			copy(bhash[:], tmp_slice[i*32:(i+1)*32])

			block_list = append(block_list, bhash)

			rlog.Tracef(4,"%2d hash  %x\n", i, bhash[:])

		}

		// send each block independently

		/*if len(block_list) == 1 {
			Send_Single_Block_to_Peer(connection, block_list[0])
		} else*/ { // we need to send al blocks in 1 go
			Send_Blocks_to_Peer(connection, block_list)

		}

	}

	// we must give user  data

}

func boost_serialisation_block(hash crypto.Hash) []byte {

	block_header := []byte{0x04, 0x05, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x0a}

	txs_header := []byte{0x03, 0x74, 0x78, 0x73, 0x8a}
	bl, err := chain.Load_BL_FROM_ID(hash)

	_ = err

	if len(bl.Tx_hashes) >= 1 {
		block_header[0] = 0x8
	}

	block_serialized := bl.Serialize()

	// add a varint len
	buf := make([]byte, 8, 8)
	done := Encode_Boost_Varint(buf, uint64(len(block_serialized))) // encode length of buffer
	buf = buf[:done]
	block_header = append(block_header, buf...)
	block_header = append(block_header, block_serialized...)

	if len(bl.Tx_hashes) >= 1 {
		block_header = append(block_header, txs_header...)

		// add txs length
		buf := make([]byte, 8, 8)
		done := Encode_Boost_Varint(buf, uint64(len(bl.Tx_hashes))) // encode length of buffer
		buf = buf[:done]
		block_header = append(block_header, buf...)

		for i := range bl.Tx_hashes {
			tx, err := chain.Load_TX_FROM_ID(bl.Tx_hashes[i])

			if err != nil {
				rlog.Tracef(1,"ERR Cannot load tx from DB\n")
				return block_header
			}

			tx_serialized := tx.Serialize()
			buf := make([]byte, 8, 8)
			done := Encode_Boost_Varint(buf, uint64(len(tx_serialized))) // encode length of buffer
			buf = buf[:done]
			block_header = append(block_header, buf...)
			block_header = append(block_header, tx_serialized...)

		}
	}

	return block_header

}

func Send_Blocks_to_Peer(connection *Connection, block_list []crypto.Hash) {

	blocks_header := []byte{0x01, 0x11, 0x01, 0x01, 0x01, 0x01, 0x02, 0x01, 0x01, 0x08, 0x06, 0x62, 0x6c, 0x6f, 0x63, 0x6b,
		0x73, 0x8c} // this is followed by a varint count of blocks

	trailer := []byte{0x19, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x5f, 0x62,
		0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x5f, 0x68, 0x65, 0x69, 0x67, 0x68, 0x74, 0x05}

	buf := make([]byte, 8, 8)
	binary.LittleEndian.PutUint64(buf, chain.Get_Height())
	trailer = append(trailer, buf...)

	done := Encode_Boost_Varint(buf, uint64(len(block_list))) // encode length of buffer
	buf = buf[:done]

	result := append(blocks_header, buf...)

	for i := range block_list {
		block := boost_serialisation_block(block_list[i])
		result = append(result, block...)
	}

	result = append(result, trailer...)

	var o_command_header Levin_Header
	o_command_header.CB = uint64(len(result))

	o_command_header.Command = BC_NOTIFY_RESPONSE_GET_OBJECTS
	o_command_header.ReturnData = false
	o_command_header.Flags = LEVIN_PACKET_REQUEST

	o_command_header_bytes, _ := o_command_header.Serialize()

	connection.Conn.Write(o_command_header_bytes)
	connection.Conn.Write(result)

}

func Send_Single_Block_to_Peer(connection *Connection, hash crypto.Hash) {
	bl, err := chain.Load_BL_FROM_ID(hash)
	if err == nil {
		if len(bl.Tx_hashes) == 0 {
			Send_Block_with_ZERO_TX(connection, bl)
		} else {
			Send_Block_with_TX(connection, bl)
		}
	}
}

/*
header
00009F94  01 11 01 01 01 01 02 01  01 08 06 62 6c 6f 63 6b   ........ ...block
00009FA4  73 8c 04 08 05 62 6c 6f  63 6b 0a                  s....blo ck......

trailer

000173C4                    19 63  75 72 72 65 6e 74 5f 62   nm.'...c urrent_b
000173D4  6c 6f 63 6b 63 68 61 69  6e 5f 68 65 69 67 68 74   lockchai n_height
000173E4  05 ec 04 00 00 00 00 00  00
*/

// if a block is with out TX, send it in this format
func Send_Block_with_ZERO_TX(connection *Connection, bl *blockchain.Block) {
	fmt.Printf("sending block with zero tx")

	header := []byte{0x01, 0x11, 0x01, 0x01, 0x01, 0x01, 0x02, 0x01, 0x01, 0x08, 0x06, 0x62, 0x6c, 0x6f, 0x63, 0x6b,
		0x73, 0x8c, 0x04, 0x04, 0x05, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x0a}
	trailer := []byte{0x19, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74, 0x5f, 0x62,
		0x6c, 0x6f, 0x63, 0x6b, 0x63, 0x68, 0x61, 0x69, 0x6e, 0x5f, 0x68, 0x65, 0x69, 0x67, 0x68, 0x74, 0x05}

	buf := make([]byte, 8, 8)
	binary.LittleEndian.PutUint64(buf, chain.Get_Height())
	trailer = append(trailer, buf...)

	block_serialized := bl.Serialize()
	done := Encode_Boost_Varint(buf, uint64(len(block_serialized))) // encode length of buffer
	buf = buf[:done]

	header = append(header, buf...)
	header = append(header, block_serialized...)
	header = append(header, trailer...)

	var o_command_header Levin_Header
	o_command_header.CB = uint64(len(header))

	o_command_header.Command = BC_NOTIFY_RESPONSE_GET_OBJECTS
	o_command_header.ReturnData = false
	o_command_header.Flags = LEVIN_PACKET_REQUEST

	o_command_header_bytes, _ := o_command_header.Serialize()

	connection.Conn.Write(o_command_header_bytes)
	connection.Conn.Write(header)

}

// if a block is with TX, send it in this format
func Send_Block_with_TX(connection *Connection, bl *blockchain.Block) {

	panic("Sending block with TX not implmented\n")
}

// header from boost packet
//0060   00 00 00 01 11 01 01 01 01 02 01 [] 01 04 09 62 6c  ..............bl
//0070   6f 63 6b 5f 69 64 73 0a [] 81 0a 9b a2 3e fe 50 5f  ock_ids.....>.P_

// send the peer the blocks hash and transactions hash, that we need
// this function, splits the request and serves
// we split the request into 2 requests
func Send_BC_Notify_Request_GetObjects(connection *Connection, block_list []crypto.Hash, tx_list []crypto.Hash) {

	connection.Lock()

	txs_header_bytes := []byte{0x03, 't', 'x', 's', 0x0a}
	blocks_header_bytes := []byte{0x06, 'b', 'l', 'o', 'c', 'k', 's', 0x0a}

	// now append boost variant length, and then append all the hashes

	// add your blocks here

	// adding genesis block
	//block_list = append(block_list,globals.Config.Genesis_Block_Hash)

	/*for i := 0; i < 100;i++{
	  block_list = append(block_list,globals.Config.Genesis_Block_Hash)
	  }*/

	//  split block request in independent request  , so the response comes independent

	if len(block_list) > 0 {
		rlog.Tracef(4, "Sending block request for  %d blocks \n", len(block_list))

		for i := range block_list {
			buf := make([]byte, 8, 8)
			done := Encode_Boost_Varint(buf, uint64(32)) // encode length of buffer
			buf = buf[:done]

			var o_command_header Levin_Header
			var o_data_header Levin_Data_Header

			o_data_header.Data = append(blocks_header_bytes, buf...)

			// convert and append all hashes to bytes
			o_data_header.Data = append(o_data_header.Data, block_list[i][:32]...)

			o_data_bytes, _ := o_data_header.Serialize()

			o_data_bytes[9] = 0x4

			o_command_header.CB = uint64(len(o_data_bytes))

			o_command_header.Command = BC_NOTIFY_REQUEST_GET_OBJECTS
			o_command_header.ReturnData = false
			o_command_header.Flags = LEVIN_PACKET_REQUEST

			o_command_header_bytes, _ := o_command_header.Serialize()

			connection.Conn.Write(o_command_header_bytes)
			connection.Conn.Write(o_data_bytes)
		}
	}

	if len(tx_list) > 0 {

		rlog.Tracef(4, "Sending tx request for  %d tx \n", len(tx_list))

		buf := make([]byte, 8, 8)
		done := Encode_Boost_Varint(buf, uint64(len(tx_list)*32)) // encode length of buffer
		buf = buf[:done]

		var o_command_header Levin_Header
		var o_data_header Levin_Data_Header

		o_data_header.Data = append(txs_header_bytes, buf...)

		// convert and append all hashes to bytes
		for _, hash := range tx_list {
			o_data_header.Data = append(o_data_header.Data, hash[:32]...)
		}

		o_data_bytes, _ := o_data_header.Serialize()

		o_data_bytes[9] = 0x4

		o_command_header.CB = uint64(len(o_data_bytes))

		o_command_header.Command = BC_NOTIFY_REQUEST_GET_OBJECTS
		o_command_header.ReturnData = false
		o_command_header.Flags = LEVIN_PACKET_REQUEST

		o_command_header_bytes, _ := o_command_header.Serialize()

		connection.Conn.Write(o_command_header_bytes)
		connection.Conn.Write(o_data_bytes)
	}

	//fmt.Printf("len of command header %d\n", len(o_command_header_bytes))
	//fmt.Printf("len of data header %d\n", len(o_data_bytes))

	connection.Unlock()
}

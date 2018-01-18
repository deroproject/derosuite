package p2p


import "bytes"

import "github.com/romana/rlog"
import log "github.com/sirupsen/logrus"

import "github.com/deroproject/derosuite/blockchain"

// if the incoming blob contains block with included transactions
//00009F94  01 11 01 01 01 01 02 01  01 08 06 62 6c 6f 63 6b   ........ ...block
//00009FA4  73 8c 04 08 05 62 6c 6f  63 6b 0a fd 03 06 06 cd   s....blo ck......

// if the incoming blob contains block without any tx
//00009EB4  01 11 01 01 01 01 02 01  01 08 06 62 6c 6f 63 6b   ........ ...block
//00009EC4  73 8c 08 04 05 62 6c 6f  63 6b 0a e5 01 01 00 00   s....blo ck......

// if the incoming blob only contains a TX

// FIXME this code can also be shared by NOTIFY_NEW_BLOCK, NOTIFY_NEW_TRANSACTIONS

// we trigger this if we want to request any TX or block from the peer
func Handle_BC_Notify_Response_GetObjects(connection *Connection,
	i_command_header *Levin_Header, buf []byte) {

	var bl blockchain.Block
	complete_block := false

	// deserialize data header
	var i_data_header Levin_Data_Header // incoming data header

	err := i_data_header.DeSerialize(buf)

	if err != nil {
		log.Debugf("We should destroy connection here, data header cnot deserialized")
		return
	}

	// check whether the response contains block
	pos := bytes.Index(i_data_header.Data, []byte("blocks")) // at this point to

	buf = i_data_header.Data

	if pos > 0 { // the data contains atleast 1 block
		pos += 6
		buf = i_data_header.Data[pos:]
		pos := bytes.Index(buf, []byte("block"))
		// find inner position of block
		pos = pos + 6 // jump to varint length position and decode

		buf = buf[pos:]
		block_length, done := Decode_Boost_Varint(buf)
		rlog.Tracef(9, "Block length %d %x\n", block_length, buf[:8])
		buf = buf[done:]

		block_buf := buf[:block_length]

		err = bl.Deserialize(block_buf)
		if err != nil {
			log.Debugf("Block could not be deserialized successfully err %s\n", err)
			log.Debugf("We should destroy connection here, block not deserialized")
			return
		}

		hash := bl.GetHash()
		rlog.Tracef(9, "Block deserialized successfully  %x\n", hash[:32])
		rlog.Tracef(9, "Tx hash length %d\n", len(bl.Tx_hashes))
		for i := range bl.Tx_hashes {
			rlog.Tracef(9, "%d tx %x\n", i, bl.Tx_hashes[i][:32])
		}
		// point buffer to check whether any more tx exist
		buf = buf[block_length:]
		complete_block = true
	}

	pos = bytes.Index(buf, []byte("\x03txs\x8a")) // at this point to

	if pos > -1 {
		rlog.Tracef(9, "txt pos %d\n", pos)

		buf = buf[pos+5:]
		// decode remain data length ( though we know it from buffer size, but still verify it )

		tx_count, done := Decode_Boost_Varint(buf)
		buf = buf[done:]

		for i := uint64(0); i < tx_count; i++ {

			var tx blockchain.Transaction

			tx_len, done := Decode_Boost_Varint(buf)
			buf = buf[done:]
			rlog.Tracef(9, "tx count %d  i %d  tx_len %d\n", tx_count, i, tx_len)

			tx_bytes := buf[:tx_len]

			// deserialize and verrify transaction

			err = tx.DeserializeHeader(tx_bytes)

			if err != nil {
				log.Debugf("Transaction could not be deserialized\n")

			} else {
				hash := tx.GetHash()
				rlog.Tracef(9, "Transaction deserialised successfully  hash %x\n", hash[:32])

				// add tx to block chain, we must verify that the tx has been mined

				chain.Add_TX(&tx)
			}

			buf = buf[tx_len:] // setup for next tx

		}
	}

	// at this point, if it's a block we should try to add it to block chain
	if complete_block {
		// add block to chain
		log.Debugf("Found a block we should add it to our chain\n")
		chain.Chain_Add(&bl)
	}

	// add all transaction to chains

}

package p2p


import "bytes"

import "github.com/romana/rlog"

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
func Handle_BC_Notify_New_Transactions(connection *Connection,
	i_command_header *Levin_Header, buf []byte) {

	// deserialize data header
	var i_data_header Levin_Data_Header // incoming data header

	err := i_data_header.DeSerialize(buf)

	if err != nil {
		connection.logger.Debugf("We should destroy connection here, data header cnot deserialized")
		connection.Exit = true
		return
	}

	connection.logger.Debugf("Incoming NOTIFY_NEW_TRANSACTIONS")

	// check whether the response contains block
	pos := bytes.Index(i_data_header.Data, []byte("blocks")) // at this point to

	buf = i_data_header.Data

	pos = bytes.Index(buf, []byte("\x03txs\x8a")) // at this point to

	if pos > -1 {
		rlog.Tracef(3, "txt pos %d", pos)

		buf = buf[pos+5:]
		// decode remain data length ( though we know it from buffer size, but still verify it )

		tx_count, done := Decode_Boost_Varint(buf)
		buf = buf[done:]

		for i := uint64(0); i < tx_count; i++ {

			var tx blockchain.Transaction

			tx_len, done := Decode_Boost_Varint(buf)
			buf = buf[done:]
			rlog.Tracef(3, "tx count %d  i %d  tx_len %d\n", tx_count, i, tx_len)

			tx_bytes := buf[:tx_len]

			// deserialize and verrify transaction

			err = tx.DeserializeHeader(tx_bytes)

			if err != nil {
				connection.logger.Warnf("Transaction could not be deserialized\n") // we should disconnect peer

			} else {
				hash := tx.GetHash()
				rlog.Tracef(2, "Transaction deserialised successfully  hash %x\n", hash[:32])
				// add tx to mem pool, we must verify that the tx  is valid at this point in time

				// we should check , whether we shoul add the tx to pool

				//chain.Add_TX(&tx)
				// we should add TX to pool
			}

			buf = buf[tx_len:] // setup for next tx

		}
	}

}

package p2p

//import "fmt"
import "bytes"
import "time"
import "encoding/binary"

import "github.com/romana/rlog"
import log "github.com/sirupsen/logrus"

// outgoing response needs to be as follows
/*
0020   01 11 01 01 01 01 02 01 01 08 0a 6c 6f 63 61     ............loca
0030   6c 5f 74 69 6d 65 05 c9 ea 45 5a 00 00 00 00 0c  l_time...EZ.....
0040   70 61 79 6c 6f 61 64 5f 64 61 74 61 0c 10 15 63  payload_data...c
0050   75 6d 75 6c 61 74 69 76 65 5f 64 69 66 66 69 63  umulative_diffic
0060   75 6c 74 79 05 37 62 00 00 00 00 00 00 0e 63 75  ulty.7b.......cu
0070   72 72 65 6e 74 5f 68 65 69 67 68 74 05 ec 04 00  rrent_height....
0080   00 00 00 00 00 06 74 6f 70 5f 69 64 0a 80 85 d9  ......top_id....
0090   d2 f6 cd ee 1b 87 dd d1 ac 3d 15 db 4d 72 63 ca  .........=..Mrc.
00a0   1c 43 37 db 53 78 7f 03 3b 74 f6 fc 45 0e 0b 74  .C7.Sx..;t..E..t
00b0   6f 70 5f 76 65 72 73 69 6f 6e 08 06              op_version..
*/

// handle P2P_COMMAND_TIMED_SYNC_T
func Handle_P2P_Timed_Sync(connection *Connection,
	i_command_header *Levin_Header, buf []byte) {

	// deserialize data header
	var i_data_header Levin_Data_Header // incoming data header

	err := i_data_header.DeSerialize(buf)

	if err != nil {
		log.Debugf("Invalid P2P_COMMAND_TIMED_SYNC_T, disconnecting peer")
		connection.Exit = true
		return
	}

	// parse incoming core data

	var peer_core_data CORE_DATA

	pos := bytes.Index(i_data_header.Data, []byte("payload_data")) // at this point to node data and should be parsed as such

	if pos < 0 {
		log.Debugf("Invalid P2P_COMMAND_TIMED_SYNC_T, disconnecting peer")
		connection.Exit = true
		return
	}
	err = peer_core_data.DeSerialize(i_data_header.Data[pos-1:])

	if err != nil {
		log.Debugf("Invalid P2P_COMMAND_TIMED_SYNC_T, disconnecting peer")
		connection.Exit = true
		return

	}

	rlog.Trace(5, "Incoming core data %+v \n", peer_core_data)

	// TODO if cumulative difficulty at this top mismatches ours start resync
	// if height is more than ours, start resync
	var our_core_data CORE_DATA
	// fill the structure with our chain data
	our_core_data.Top_ID = chain.Get_Top_ID()
	our_core_data.Cumulative_Difficulty = chain.Load_Block_Cumulative_Difficulty(our_core_data.Top_ID) // get cumulative difficulty for top block
	our_core_data.Current_Height = chain.Load_Height_for_BL_ID(our_core_data.Top_ID)
	our_core_data.Top_Version = 6

	serialised_bytes, _ := our_core_data.Serialize()

	header_bytes := []byte{0x01, 0x11, 0x01, 0x01, 0x01, 0x01, 0x02, 0x01, 0x01, 0x08, 0x0a, 0x6c, 0x6f, 0x63, 0x61,
		0x6c, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x05,
		/* time bytes start here */ 0xc9, 0xea, 0x45, 0x5a, 0x00, 0x00, 0x00, 0x00}

	binary.LittleEndian.PutUint64(header_bytes[22:], uint64(time.Now().Unix()))

	//fmt.Printf("header %x  serialised_bytes %x\n", header_bytes, serialised_bytes)

	response_bytes := append(serialised_bytes, header_bytes...)

	// create a new response header

	var o_command_header Levin_Header
	//var o_data_header  Levin_Data_Header

	o_command_header.CB = uint64(len(response_bytes))

	o_command_header.Command = P2P_COMMAND_TIMED_SYNC
	o_command_header.ReturnData = false
	o_command_header.Flags = LEVIN_PACKET_RESPONSE

	o_command_header_bytes, _ := o_command_header.Serialize()

	connection.Conn.Write(o_command_header_bytes)
	connection.Conn.Write(response_bytes[:])

	connection.Last_Height = peer_core_data.Current_Height
	connection.Top_Version = uint64(peer_core_data.Top_Version)
	connection.Top_ID = peer_core_data.Top_ID
	connection.Cumulative_Difficulty = peer_core_data.Cumulative_Difficulty
	connection.State = ACTIVE

	// lets check whether we need to resync with this peer
	if chain.IsLagging(peer_core_data.Cumulative_Difficulty, peer_core_data.Current_Height, peer_core_data.Top_ID) {
		log.Debugf("We need to resync with the peer")
		// set mode to syncronising
		Send_BC_Notify_Chain_Command(connection)
	}

}

/* we will never send this request, so we donot need to parse response
func Send_P2P_Timed_Sync(connection *Connection){

connection.Lock()

var o_command_header Levin_Header
var o_data_header  Levin_Data_Header

o_data_bytes,_ := o_data_header.Serialize()

o_command_header.CB = uint64(len(o_data_bytes))

o_command_header.Command = P2P_COMMAND_REQUEST_SUPPORT_FLAGS
o_command_header.ReturnData = true
o_command_header.Flags = LEVIN_PACKET_REQUEST

o_command_header_bytes,_ := o_command_header.Serialize()

connection.Conn.Write(o_command_header_bytes)
connection.Conn.Write(o_data_bytes)

connection.Command_queue.PushBack(uint32(P2P_COMMAND_REQUEST_SUPPORT_FLAGS))

connection.Unlock()
}
*/

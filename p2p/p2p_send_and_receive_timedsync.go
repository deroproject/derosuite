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

// this file sends a timed sync packet and parses response

/*  below is a timed sync request
000001DE  01 21 01 01 01 01 01 01  87 00 00 00 00 00 00 00   .!...... ........
000001EE  01 ea 03 00 00 00 00 00  00 01 00 00 00 01 00 00   ........ ........
000001FE  00                                                 .
000001FF  01 11 01 01 01 01 02 01  01 04 0c 70 61 79 6c 6f   ........ ...paylo
0000020F  61 64 5f 64 61 74 61 0c  10 15 63 75 6d 75 6c 61   ad_data. ..cumula
0000021F  74 69 76 65 5f 64 69 66  66 69 63 75 6c 74 79 05   tive_dif ficulty.
0000022F  8e 4a f1 09 da 00 00 00  0e 63 75 72 72 65 6e 74   .J...... .current
0000023F  5f 68 65 69 67 68 74 05  53 4b 00 00 00 00 00 00   _height. SK......
0000024F  06 74 6f 70 5f 69 64 0a  80 1d ca fd 23 08 50 54   .top_id. ....#.PT
0000025F  16 e0 41 c8 33 e6 db 91  ca 33 6b b1 fb af 9d a1   ..A.3... .3k.....
0000026F  99 f2 77 85 5d a1 e8 2a  fa 0b 74 6f 70 5f 76 65   ..w.]..* ..top_ve
0000027F  72 73 69 6f 6e 08 06                               rsion..


00000000  01 21 01 01 01 01 01 01  87 00 00 00 00 00 00 00   .!...... ........
00000010  01 ea 03 00 00 00 00 00  00 01 00 00 00 01 00 00   ........ ........
00000020  00                                                 .
00000021  01 11 01 01 01 01 02 01  01 04 0c 70 61 79 6c 6f   ........ ...paylo
00000031  61 64 5f 64 61 74 61 0c  10 15 63 75 6d 75 6c 61   ad_data. ..cumula
00000041  74 69 76 65 5f 64 69 66  66 69 63 75 6c 74 79 05   tive_dif ficulty.
00000051  2e 9d 54 af fd 00 00 00  0e 63 75 72 72 65 6e 74   ..T..... .current
00000061  5f 68 65 69 67 68 74 05  b6 4c 00 00 00 00 00 00   _height. .L......
00000071  06 74 6f 70 5f 69 64 0a  80 64 52 25 11 88 5a ca   .top_id. .dR%..Z.
00000081  10 c8 f9 28 c5 ad a4 03  e1 4f 6d 68 23 8c c5 ea   ...(.... .Omh#...
00000091  73 e6 16 58 37 d1 96 22  07 0b 74 6f 70 5f 76 65   s..X7.." ..top_ve
000000A1  72 73 69 6f 6e 08 06

01 21 01 01 01 01 01 01 87 00 00 00 00 00 00 00
01 ea 03 00 00 00 00 00  00 01 00 00 00 01 00 00
00
01 11 01 01 01 01 02 01  01 04 0c 70 61 79 6c 6f
61 64 5f 64 61 74 61 0c  10 15 63 75 6d 75 6c 61
74 69 76 65 5f 64 69 66  66 69 63 75 6c 74 79 05
2e 9d 54 af fd 00 00 00  0e 63 75 72 72 65 6e 74
5f 68 65 69 67 68 74 b6  4c00000000000000
06 74 6f 70 64 52 25 11  88 5a ca 10 c8 f9 28 c5
ada403e14f6d68238cc5ea73e6165837
d19622075da1e82afa0b746f705f7665
72 73 69 6f 6e 08 06

01210101010101018700000000000000
01ea0300000000000001000000010000
00
011101010101020101040c7061796c6f
61645f646174610c101563756d756c61
746976655f646966666963756c747905
2e9d54affd0000000e63757272656e74
5f 68 65 69 67 68 74 05  b6 4c 00 00 00 00 00 00
06 74 6f 70 5f 69 64 0a 80  1d ca fd 23 08 50 64
522511885aca10c8f928c5ada403e14f
6d68238cc5ea73e6165837d196220765
7273696f6e0806

*/
func Send_Timed_Sync(connection *Connection) {

	request_packet := []byte{0x01, 0x21, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01,
		0x87, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x01, 0xea, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00,
		0x00,
		0x01, 0x11, 0x01, 0x01, 0x01, 0x01, 0x02, 0x01,
		0x01, 0x04, 0x0c, 0x70, 0x61, 0x79, 0x6c, 0x6f,
		0x61, 0x64, 0x5f, 0x64, 0x61, 0x74, 0x61, 0x0c,
		0x10, 0x15, 0x63, 0x75, 0x6d, 0x75, 0x6c, 0x61,
		0x74, 0x69, 0x76, 0x65, 0x5f, 0x64, 0x69, 0x66,
		0x66, 0x69, 0x63, 0x75, 0x6c, 0x74, 0x79, 0x05,
		0x8e, 0x4a, 0xf1, 0x09, 0xda, 0x00, 0x00, 0x00,
		0x0e, 0x63, 0x75, 0x72, 0x72, 0x65, 0x6e, 0x74,
		0x5f, 0x68, 0x65, 0x69, 0x67, 0x68, 0x74, 0x05,
		0x53, 0x4b, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x06, 0x74, 0x6f, 0x70, 0x5f, 0x69, 0x64, 0x0a,
		0x80, 0x1d, 0xca, 0xfd, 0x23, 0x08, 0x50, 0x54,
		0x16, 0xe0, 0x41, 0xc8, 0x33, 0xe6, 0xdb, 0x91,
		0xca, 0x33, 0x6b, 0xb1, 0xfb, 0xaf, 0x9d, 0xa1,
		0x99, 0xf2, 0x77, 0x85, 0x5d, 0xa1, 0xe8, 0x2a,
		0xfa, 0x0b, 0x74, 0x6f, 0x70, 0x5f, 0x76, 0x65,
		0x72, 0x73, 0x69, 0x6f, 0x6e, 0x08, 0x06}

	// write our data into the packet

	top_id := chain.Get_Top_ID()
	cumulative_diff := chain.Load_Block_Cumulative_Difficulty(top_id)
	height := chain.Get_Height()
	top_version := byte(6)

	// now lets write that into the buffer
	binary.LittleEndian.PutUint64(request_packet[81:], cumulative_diff)
	binary.LittleEndian.PutUint64(request_packet[105:], height)
	copy(request_packet[122:], top_id[:])
	request_packet[len(request_packet)-1] = top_version

	//connection.logger.Infof("sync request packet %x  top_id %x", request_packet, top_id)
	connection.Lock()

	defer connection.Unlock()
	connection.Conn.Write(request_packet)

	connection.Command_queue.PushBack(uint32(P2P_COMMAND_TIMED_SYNC))

}

// handle response of above command
// response comes as below
/*
   00000000  01 21 01 01 01 01 01 01  9b 00 00 00 00 00 00 00   .!...... ........
   00000010  00 ea 03 00 00 00 00 00  00 02 00 00 00 01 00 00   ........ ........
   00000020  00                                                 .
   00000021  0c 70 61 79 6c 6f 61 64  5f 64 61 74 61 0c 10 15   .payload _data...
   00000031  63 75 6d 75 6c 61 74 69  76 65 5f 64 69 66 66 69   cumulati ve_diffi
   00000041  63 75 6c 74 79 05 8e 4a  f1 09 da 00 00 00 0e 63   culty..J .......c
   00000051  75 72 72 65 6e 74 5f 68  65 69 67 68 74 05 52 4b   urrent_h eight.RK
   00000061  00 00 00 00 00 00 06 74  6f 70 5f 69 64 0a 80 1d   .......t op_id...
   00000071  ca fd 23 08 50 54 16 e0  41 c8 33 e6 db 91 ca 33   ..#.PT.. A.3....3
   00000081  6b b1 fb af 9d a1 99 f2  77 85 5d a1 e8 2a fa 0b   k....... w.]..*..
   00000091  74 6f 70 5f 76 65 72 73  69 6f 6e 08 06 01 11 01   top_vers ion.....
   000000A1  01 01 01 02 01 01 08 0a  6c 6f 63 61 6c 5f 74 69   ........ local_ti
   000000B1  6d 65 05 df 3d 5b 5a 00  00 00 00                  me..=[Z. ...
*/

func Handle_P2P_Timed_Sync_Response(connection *Connection,
	i_command_header *Levin_Header, buf []byte) {

	var i_data_header Levin_Data_Header // incoming data header, deserialize it
	var peer_core_data CORE_DATA
	err := i_data_header.DeSerialize(buf)

	if err != nil {
		connection.logger.Debugf("Invalid Levin Data header, disconnecting peer")
		connection.Exit = true
		return
	}

	// parse incoming core data
	pos := bytes.Index(i_data_header.Data, []byte("payload_data")) // at this point to node data and should be parsed as such

	if pos < 0 {
		connection.logger.Debugf("Invalid P2P_COMMAND_TIMED_SYNC_T, could not find payload_data, disconnecting peer")
		connection.Exit = true
		return
	}
	err = peer_core_data.DeSerialize(i_data_header.Data[pos-1:])

	if err != nil {
		connection.logger.Debugf("Invalid P2P_COMMAND_TIMED_SYNC_T, could not deserialize core_data, disconnecting peer")
		connection.Exit = true
		return

	}

	rlog.Trace(5, "Incoming core data %+v \n", peer_core_data)

	// lets check whether we need to resync with this peer
	if chain.IsLagging(peer_core_data.Cumulative_Difficulty, peer_core_data.Current_Height, peer_core_data.Top_ID) {
		connection.logger.Debugf("We need to resync with the peer")
		// set mode to syncronising
		Send_BC_Notify_Chain_Command(connection)
	}

}

package p2p


import "fmt"
import "net"
import "time"

import "github.com/romana/rlog"
import log "github.com/sirupsen/logrus"

import "github.com/deroproject/derosuite/globals"



// the connection starts with P2P handshake

// send the hand shake
func Send_Handshake(connection *Connection) {

	// first request support flags

	if connection.Exit {
		return
	}

	Send_SupportFlags_Command(connection)

	connection.Lock()

	// lets do handshake
	var d Node_Data
	var c CORE_DATA
	var data_header Levin_Data_Header
	var levin_header Levin_Header

	d.Network_UUID = globals.Config.Network_ID
	d.Peer_ID = (uint64)(time.Now().Unix())
	c.Current_Height = chain.Get_Height()
	c.Cumulative_Difficulty = chain.Get_Difficulty()
	c.Top_Version = 6

	// top block id from the genesis or other block
	c.Top_ID = chain.Get_Top_ID()

	ds, _ := d.Serialize()
	cs, _ := c.Serialize()

	data_header.Data = ds
	data_header.Data = append(data_header.Data, cs...)

	levin_data, _ := data_header.Serialize()
	levin_header.CB = uint64(len(levin_data))

	levin_header.Command = P2P_COMMAND_HANDSHAKE
	levin_header.ReturnData = true
	levin_header.Flags = LEVIN_PACKET_REQUEST

	header_bytes, _ := levin_header.Serialize()

	connection.Conn.Write(header_bytes)
	connection.Conn.Write(levin_data)

	connection.Command_queue.PushBack(uint32(P2P_COMMAND_HANDSHAKE))

	connection.Unlock()
}

// handle P2P_COMMAND_HANDSHAKE,
// we must send a response
// our response is a boost compatible response which is parseable by old cryptonote daemons
func Handle_P2P_Handshake_Command(connection *Connection,
	i_command_header *Levin_Header, buf []byte) {
    
    
       connection.logger.Infof("Handshake request arrived, we must parse it")

	// extract peers from our list, and insert them into the response
	// max 250 peers can be send ( since we are aiming compatibility with  old daemons)

	var reply Node_Data_Response

	//panic("Handle_P2P_Handshake needs to fixed and tested")

	reply.NodeData.Network_UUID = globals.Config.Network_ID
	reply.NodeData.Peer_ID = (uint64)(OUR_PEER_ID)
	reply.NodeData.Local_time = uint64(time.Now().Unix())

	reply.CoreData.Current_Height = chain.Get_Height()
	reply.CoreData.Cumulative_Difficulty = chain.Get_Difficulty()
	reply.CoreData.Top_Version = 6

	// top block id from the genesis or other block
	// this data is from the block chain
	// this data must be the top block that we see till now
	reply.CoreData.Top_ID = chain.Get_Top_ID()

	for i := 0; i < 250; i++ {
		reply.PeerArray = append(reply.PeerArray,
			Peer_Info{IP: net.IPv4(byte(i), byte(i), byte(i), byte(i)), Port: 0, ID: 0, LastSeen: 0})
	}

	var o_command_header Levin_Header
	var o_data_header Levin_Data_Header

	o_data_header.Data, _ = reply.Serialize()
	data, _ := o_data_header.Serialize()

	// mark as containing 4 elements
	data[9] = 0x10

	o_command_header.CB = uint64(len(data))

	o_command_header.Command = P2P_COMMAND_HANDSHAKE
	o_command_header.ReturnData = false
	o_command_header.ReturnCode = 1 // send as good response
	o_command_header.Flags = LEVIN_PACKET_RESPONSE

	o_command_header_bytes, _ := o_command_header.Serialize()

	connection.Conn.Write(o_command_header_bytes)
	connection.Conn.Write(data)

	rlog.Tracef(4, "Sending handshake response\n")
        
        
        Handle_P2P_Handshake_Command_Response(connection, i_command_header,buf) // parse incoming response

}

/* handles response of our p2p command, parses data etc*/

func Handle_P2P_Handshake_Command_Response(connection *Connection,
	i_command_header *Levin_Header, buf []byte) {

	var reply Node_Data_Response

	// deserialize data header
	var i_data_header Levin_Data_Header // incoming data header

	err := i_data_header.DeSerialize(buf)

	if err != nil {

		connection.logger.WithFields(log.Fields{
			"ip": connection.Addr.IP,
		}).Debugf("Disconnecting client, handshake could not be deserialized")

		return
	}

	if reply.DeSerialize(i_data_header.Data) != nil {

		logger.WithFields(log.Fields{
			"ip": connection.Addr.IP,
		}).Debugf("Disconnecting client, handshake could not be deserialized")

		return
	}

	if reply.NodeData.Network_UUID != globals.Config.Network_ID {

		logger.WithFields(log.Fields{
			"ip": connection.Addr.IP,
			"id": reply.NodeData.Network_UUID,
		}).Debugf("Disconnecting client, Wrong network ID")

		return

	}

	// we need to kick the peer if the height is something specific and peer id is less than ours
	// TODO right we are not doing it
	connection.Peer_ID = reply.NodeData.Peer_ID
	connection.Port = reply.NodeData.Local_Port
	connection.Last_Height = reply.CoreData.Current_Height
	connection.Top_Version = uint64(reply.CoreData.Top_Version)
	connection.Top_ID = reply.CoreData.Top_ID
	connection.Cumulative_Difficulty = reply.CoreData.Cumulative_Difficulty
	connection.State = ACTIVE

	connection.logger.WithFields(log.Fields{
		"PeerHeight":  reply.CoreData.Current_Height,
		"Top Version": reply.CoreData.Top_Version,
		"Top_ID":      fmt.Sprintf("%x", reply.CoreData.Top_ID),
	}).Debugf("Successful Handshake with Peer")

	// lets check whether we need to resync with this peer
	if chain.IsLagging(reply.CoreData.Cumulative_Difficulty, reply.CoreData.Current_Height, reply.CoreData.Top_ID) {
		logger.Debugf("We need to resync with the peer")
		// set mode to syncronising
		Send_BC_Notify_Chain_Command(connection)
	}

}

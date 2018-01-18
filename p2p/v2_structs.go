package p2p


// This file defines the structure for the protocol which is a msgp encoded ( which is standard)
// msgp  would cause an easy rewrite of p2p layer even in c, ruby or rust etc as future may demand
// the protocol is length prefixed msgp payload
// though we can http2 stream features,  they may become compilcated as the project evolve
// the prefix length is 4 bytes, little endian encoded ( so a frame can be 4GB in size)

//import "net"

// the protocol is completely asyncronous, except for first handshake, so the node remain undetectable to external network scans, the detection cost is to atleast send a handshake packet*/

// used to parse incoming packet for for command , so as a repective command command could be triggered
type Common struct {
	Command uint64
	Height  uint64
	Cumulative_Difficulty uint64
	Top_ID           [32]byte // 32 bytes of Top block
	Top_Version  uint64 // this basically represents the hard fork version
}

// at start, client sends handshake and server will respond to handshake
type Handshake struct {
        Common     // add all fields of Common
	Local_Time       uint64
	Local_Port       uint32
	ID               uint64
	Network_ID       []byte // 16 bytes
	PeerList       []Peer_Info
	Extension_List []string
}

type Sync struct {
        Common  // add all fields of common
}
    

const V2_COMMAND_HANDSHAKE=1  // commands are syncronous and must be responded within 10 secs
const V2_COMMAND_SYNC=2
const V2_COMMAND_CHAIN_REQUEST=3
const V2_COMMAND_CHAIN_RESPONSE=4 
const V2_COMMAND_OBJECTS_REQUEST=5
const V2_COMMAND_OBJECTS_RESPONSE=6

const V2_NOTIFY_NEW_OBJECTS=0x80000001  // all notifications come here, such as new block, new txs



type Chain_Request struct {
	Block_list [][32]byte
}

type Chain_Response struct {
	Block_list [][32]byte
}

type Object_Request struct {
    Block_list [][32]byte
    Tx_list [][32]byte
}

type Complete_Block struct {
    Block []byte
    Tx [][]byte
}

type Object_Response struct {
    Blocks []Complete_Block
    Tx_list [][32]byte
}

type Notify_New_Objects struct {
    Block []byte
    Tx_list [][]byte
}




// each packet has to be parsed twice
// the following

package p2p

// these are defined in cryptonote_protocol_defs.h

const BC_COMMANDS_POOL_BASE = 2000
const BC_NOTIFY_NEW_BLOCK = BC_COMMANDS_POOL_BASE + 1        // arrival of new block
const BC_NOTIFY_NEW_TRANSACTIONS = BC_COMMANDS_POOL_BASE + 2 // arrival of new transactions

const BC_NOTIFY_REQUEST_GET_OBJECTS = BC_COMMANDS_POOL_BASE + 3  // get objects
const BC_NOTIFY_RESPONSE_GET_OBJECTS = BC_COMMANDS_POOL_BASE + 4 // carries payload

// where is the 5th one ??

const BC_NOTIFY_REQUEST_CHAIN = BC_COMMANDS_POOL_BASE + 6        // provides chain state
const BC_NOTIFY_RESPONSE_CHAIN_ENTRY = BC_COMMANDS_POOL_BASE + 7 // request and reponse, for block height

// used to print names for command
var BC_COMMAND_NAME = map[uint32]string{

	// below are p2p commands
	P2P_COMMAND_HANDSHAKE:             "P2P_COMMAND_HANDSHAKE",
	P2P_COMMAND_TIMED_SYNC:            "P2P_COMMAND_TIMED_SYNC",
	P2P_COMMAND_PING:                  "P2P_COMMAND_PING",
	P2P_COMMAND_REQUEST_SUPPORT_FLAGS: "P2P_COMMAND_REQUEST_SUPPORT_FLAGS",

	// below are cryptonote based commands
}

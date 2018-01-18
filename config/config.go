package config

import "github.com/satori/go.uuid"
import "github.com/deroproject/derosuite/crypto"


// all global configuration variables are picked from here

var BLOCK_TIME = uint64(120)

// we are ignoring leap seconds from calculations

// coin emiision related settings
var COIN_MONEY_SUPPLY = uint64(18446744073709551615) // 2^64-1
var COIN_EMISSION_SPEED_FACTOR = uint64(20)
var COIN_DIFFICULTY_TARGET = uint64(120)                 // this is a feeder to emission formula
var COIN_FINAL_SUBSIDY_PER_MINUTE = uint64(300000000000) // 0.3 DERO per minute = 157680 per year roughly
var CRYPTONOTE_REWARD_BLOCKS_WINDOW = uint64(100)        // last 100 blocks are used to create

var MAX_CHAIN_HEIGHT = uint64(2147483648) // 2^31

// we use this for scheduled hardforks
var CURRENT_BLOCK_MAJOR_VERSION = 6
var CURRENT_BLOCK_MINOR_VERSION = 6
var CRYPTONOTE_BLOCK_GRANTED_FULL_REWARD_ZONE = uint64(300000) // after this block size , reward calculated differently

// consider last 30 blocks for calculating difficulty
var DIFFICULTY_BLOCKS_COUNT_V2 = 30

const PROJECT_NAME = "dero"
const POOLDATA_FILENAME = "poolstate.bin"

//const CRYPTONOTE_BLOCKCHAINDATA_FILENAME      "data.mdb" // these decisions are made by storage layer
//#define CRYPTONOTE_BLOCKCHAINDATA_LOCK_FILENAME "lock.mdb"
const P2P_NET_DATA_FILENAME = "p2pstate.bin"

// we can have number of chains running for testing reasons
type CHAIN_CONFIG struct {
	Name                             string
	Network_ID                       uuid.UUID // network ID
	Public_Address_Prefix            uint64
	Public_Address_Prefix_Integrated uint64

	P2P_Default_Port uint32
	RPC_Default_Port uint32

	Genesis_Nonce uint32

	Genesis_Block_Hash crypto.Hash

	Genesis_Tx string
}

var Mainnet = CHAIN_CONFIG{Name: "mainnet",
	Network_ID:                       uuid.FromBytesOrNil([]byte{0x59, 0xd7, 0xf7, 0xe9, 0xdd, 0x48, 0xd5, 0xfd, 0x13, 0x0a, 0xf6, 0xe0, 0x9a, 0xec, 0xb9, 0x23}),
	Public_Address_Prefix:            0xc8ed8, //for dERo
	Public_Address_Prefix_Integrated: 0xa0ed8, //for dERi
	P2P_Default_Port:                 18090,
	RPC_Default_Port:                 18091,
	Genesis_Nonce:                    10000,

	Genesis_Block_Hash: crypto.Hash([32]byte{0x36, 0x2d, 0x61, 0x48, 0xd6, 0x83, 0x08, 0x2d,
		0x94, 0x2e, 0x53, 0xdd, 0xb5, 0x0d, 0xaf, 0x54,
		0x6a, 0x10, 0x92, 0xda, 0x76, 0x98, 0x2d, 0x5b,
		0xd4, 0xf1, 0x3d, 0x0d, 0xf0, 0x74, 0xec, 0x2f}),

	Genesis_Tx: "" +
		"02" + // version
		"3c" + // unlock time
		"01" + // vin length
		"ff" + // vin #1
		"00" + // height gen input
		"01" + // vout length
		"ffffffffffff07" + // output #1 amount
		"02" + // output 1 type
		"0bf6522f9152fa26cd1fc5c022b1a9e13dab697f3acf4b4d0ca6950a867a1943" + // output #1 key
		"21" + // extra length in bytes
		"01" + // extra pubkey tag
		"1d92826d0656958865a035264725799f39f6988faa97d532f972895de849496d" + // tx pubkey
		"00", // RCT signature none
}

var Testnet = CHAIN_CONFIG{Name: "testnet",
	Network_ID:                       uuid.FromBytesOrNil([]byte{0x59, 0xd7, 0xf7, 0xe9, 0xdd, 0x48, 0xd5, 0xfd, 0x13, 0x0a, 0xf6, 0xe0, 0x9a, 0xec, 0xb9, 0x24}),
	Public_Address_Prefix:            0x6cf58, // for dETo
	Public_Address_Prefix_Integrated: 0x44f58, //for dETi
	P2P_Default_Port:                 28090,
	RPC_Default_Port:                 28091,
	Genesis_Nonce:                    10001,

	Genesis_Block_Hash: crypto.Hash([32]byte{0x63, 0x34, 0x12, 0xde, 0x21, 0xea, 0xcb, 0xf0,
		0x03, 0xe0, 0xfb, 0x9b, 0x7f, 0xcb, 0xca, 0x97,
		0x6d, 0xff, 0xd4, 0x3e, 0x3f, 0x06, 0x9e, 0x55,
		0xfa, 0xf1, 0xc5, 0xb4, 0x46, 0x2b, 0x59, 0x3a}),

	Genesis_Tx: "" +
		"02" + // version
		"3c" + // unlock time
		"01" + // vin length
		"ff" + // vin #1
		"00" + // height gen input
		"01" + // vout length
		"ffffffffffff07" + // output #1 amount
		"02" + // output 1 type
		"0bf6522f9152fa26cd1fc5c022b1a9e13dab697f3acf4b4d0ca6950a867a1943" + // output #1 key
		"21" + // extra length in bytes
		"01" + // extra pubkey tag
		"1d92826d0656958865a035264725799f39f6988faa97d532f972895de849496d" + // tx pubkey
		"00", // RCT signature none

}

// on init this variable is updated to setup global config in 1 go
//var Current_Config CHAIN_CONFIG

func init() {
	//Current_Config = Mainnet // default is mainnnet
	//Current_Config = Testnet // default is mainnnet
}

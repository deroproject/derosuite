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

var MINER_TX_AMOUNT_UNLOCK = uint64(60)  // miner tx will need 60 blocks to mature
var NORMAL_TX_AMOUNT_UNLOCK = uint64(10) // normal transfers will mature at 10th (9 blocks distance) blocks to mature

// this is used to find whether output is locked to height or time
// see input maturity to find how it works
// if locked is less than this, then it is considered locked to height else epoch
var CRYPTONOTE_MAX_BLOCK_NUMBER = uint64(500000000)

var MAX_CHAIN_HEIGHT = uint64(2147483648) // 2^31

// we use this for scheduled hardforks
var CURRENT_BLOCK_MAJOR_VERSION = 6
var CURRENT_BLOCK_MINOR_VERSION = 6

// this is also the minimum block size
var CRYPTONOTE_BLOCK_GRANTED_FULL_REWARD_ZONE = uint64(300000) // after this block size , reward calculated differently
var CRYPTONOTE_MAX_TX_SIZE = uint64(100 * 1024 * 1024)         // 100 MB, we must rationalize it

// we only accept blocks which are this much into future, 2 hours
const CRYPTONOTE_BLOCK_FUTURE_TIME_LIMIT = 60 * 60 * 20

// block is checked that the timestamp is not less than the median of this many blocks
const BLOCKCHAIN_TIMESTAMP_CHECK_WINDOW = 60

// consider last 30 blocks for calculating difficulty
const DIFFICULTY_BLOCKS_COUNT_V2 = 30

// FEE calculation constants are here
// the constants can be found in cryptonote_config.h
const DYNAMIC_FEE_PER_KB_BASE_FEE_V5 = uint64((2000000000 * 60000) / 300000)
const DYNAMIC_FEE_PER_KB_BASE_BLOCK_REWARD = uint64(10000000000000) // 10 * pow(10,12)

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

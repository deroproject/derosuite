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

// though testing hash complete successfully with 3 secs block time, however
// consider homeusers/developing countries we will be targetting  9 secs
// later hardforks can make it lower by 1 sec, say every 6 months or so, until the system reaches 3 secs
// by that time, networking,space requirements  and cryptonote tx processing requiremtn will probably outgrow homeusers
// since most mining nodes will be running in datacenter, 3 secs  blocks c
const BLOCK_TIME = uint64(12)

// we are ignoring leap seconds from calculations

// coin emiision related settings
const COIN_MONEY_SUPPLY = uint64(18446744073709551615) // 2^64-1
const COIN_EMISSION_SPEED_FACTOR = uint64(20)
const COIN_DIFFICULTY_TARGET = uint64(120)                 // this is a feeder to emission formula
const COIN_FINAL_SUBSIDY_PER_MINUTE = uint64(300000000000) // 0.3 DERO per minute = 157680 per year roughly
const CRYPTONOTE_REWARD_BLOCKS_WINDOW = uint64(100)        // last 100 blocks are used to create

const MINER_TX_AMOUNT_UNLOCK = uint64(60)  // miner tx will need 60 blocks to mature
const NORMAL_TX_AMOUNT_UNLOCK = uint64(11) // normal transfers will mature at 10th (9 blocks distance) blocks to mature

// these are used to configure mainnet hard fork
const HARDFORK_1_END = int64(1)

//const HARDFORK_1_TOTAL_SUPPLY = uint64(2000000000000000000 ) // this is used to mark total supply
// till 95532 (includind)  4739519967524007940
// 95543   4739807553788105597
// 95549   4739964392976757069
// 95550   4739990536584241377
const MAINNET_HARDFORK_1_TOTAL_SUPPLY = uint64(4739990536584241377)

const TESTNET_HARDFORK_1_TOTAL_SUPPLY = uint64(4319584000000000000)

// this is used to find whether output is locked to height or time
// see input maturity to find how it works
// if locked is less than this, then it is considered locked to height else epoch
const CRYPTONOTE_MAX_BLOCK_NUMBER = uint64(500000000)

const MAX_CHAIN_HEIGHT = uint64(2147483648) // 2^31

// this is also the minimum block size
// no longer used for emission as our block sizes are now fixed
const CRYPTONOTE_BLOCK_GRANTED_FULL_REWARD_ZONE = uint64(300000) // after this block size , reward calculated differently

// max block deviation of 2 seconds is allowed
const CRYPTONOTE_FUTURE_TIME_LIMIT = 2

// 1.25 MB block every 12 secs is equal to roughly 75 TX per second
// if we consider side blocks, TPS increase to > 100 TPS
// we can easily improve TPS by changing few parameters in this file
// the resources compute/network may not be easy for the developing countries
// we need to trade of TPS  as per community
const CRYPTONOTE_MAX_BLOCK_SIZE = uint64((1 * 1024 * 1024) + (256*1024 )) // max block size limit

const CRYPTONOTE_MAX_TX_SIZE = 300 * 1024 // max size

const MAX_VOUT = 8   // max payees, 6, 7 is change,   8th will be rejected
const MIN_MIXIN = 5  //  >= 5 ,   4 mixin will be rejected
const MAX_MIXIN = 14 // <= 13,   14th will rejected

// ATLANTIS FEE calculation constants are here
const FEE_PER_KB = uint64(1000000000) // .001 dero per kb

// mainnet botstraps at 200 MH
//const MAINNET_BOOTSTRAP_DIFFICULTY = uint64(200 *  1000* 1000 * BLOCK_TIME)
const MAINNET_BOOTSTRAP_DIFFICULTY = uint64(200 *1000*1000 * BLOCK_TIME)
const MAINNET_MINIMUM_DIFFICULTY = uint64(1000*1000 * BLOCK_TIME) // 2KH

// testnet bootstraps at 1 MH
//const  TESTNET_BOOTSTRAP_DIFFICULTY = uint64(1000*1000*BLOCK_TIME)
const TESTNET_BOOTSTRAP_DIFFICULTY = uint64(800 * BLOCK_TIME) // testnet bootstrap at 800 H/s
const TESTNET_MINIMUM_DIFFICULTY = uint64(800 * BLOCK_TIME) // 800 H


// this single parameter controls lots of various parameters
// within the consensus, it should never go below 7
// if changed responsibly, we can have one second  or lower blocks (ignoring chain bloat/size issues)
// gives immense scalability,
// increasing this means, you need to change  maturity limits also
const STABLE_LIMIT = int64(8)

// we can have number of chains running for testing reasons
type CHAIN_CONFIG struct {
	Name                             string
	Network_ID                       uuid.UUID // network ID
	Public_Address_Prefix            uint64
	Public_Address_Prefix_Integrated uint64

	P2P_Default_Port        int
	RPC_Default_Port        int
	Wallet_RPC_Default_Port int

	Genesis_Nonce uint32

	Genesis_Block_Hash crypto.Hash

	Genesis_Tx string
}

var Mainnet = CHAIN_CONFIG{Name: "mainnet",
	Network_ID:                       uuid.FromBytesOrNil([]byte{0x59, 0xd7, 0xf7, 0xe9, 0xdd, 0x48, 0xd5, 0xfd, 0x13, 0x0a, 0xf6, 0xe0, 0x9a, 0x11, 0x22, 0x33}),
	Public_Address_Prefix:            0xc8ed8, //for dERo  823000
	Public_Address_Prefix_Integrated: 0xa0ed8, //for dERi  659160
	P2P_Default_Port:                 20202,
	RPC_Default_Port:                 20206,
	Wallet_RPC_Default_Port:          20209,
	Genesis_Nonce:                    10000,

	Genesis_Block_Hash: crypto.HashHexToHash("e14e318562db8d22f8d00bd41c7938807c7ff70e4380acc6f7f2427cf49f474a"),

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

var Testnet = CHAIN_CONFIG{Name: "testnet", // testnet will always have last 3 bytes 0
	Network_ID:                       uuid.FromBytesOrNil([]byte{0x59, 0xd7, 0xf7, 0xe9, 0xdd, 0x48, 0xd5, 0xfd, 0x13, 0x0a, 0xf6, 0xe0, 0x9a, 0x04, 0x00, 0x00}),
	Public_Address_Prefix:            0x6cf58, // for dETo 446296
	Public_Address_Prefix_Integrated: 0x44f58, // for dETi 282456
	P2P_Default_Port:                 30303,
	RPC_Default_Port:                 30306,
	Wallet_RPC_Default_Port:          30309,
	Genesis_Nonce:                    10001,

	Genesis_Block_Hash: crypto.HashHexToHash("7be4a8f27bcadf556132dba38c2d3d78214beec8a959be17caf172317122927a"),

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

// the constants can be found in cryptonote_config.h
// these are still here for previous emission functions, they are not used directly for atlantis
const DYNAMIC_FEE_PER_KB_BASE_FEE_V5 = uint64((2000000000 * 60000) / 300000)
const DYNAMIC_FEE_PER_KB_BASE_BLOCK_REWARD = uint64(1000000000000) // 1 * pow(10,12)

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

// this package contains only struct definitions
// in order to avoid the dependency on block chain by any package requiring access to rpc
// and other structures
// having the structures was causing the build times of explorer/wallet to be more than couple of secs
// so separated the logic from the structures

package structures

import "github.com/deroproject/derosuite/globals"

//import "github.com/deroproject/derosuite/structures"

type (
	GetBlockHeaderByHeight_Params struct {
		Height uint64 `json:"height"`
	} // no params
	GetBlockHeaderByHeight_Result struct {
		Block_Header BlockHeader_Print `json:"block_header"`
		Status       string            `json:"status"`
	}
)

type (
	GetBlockHeaderByTopoHeight_Params struct {
		TopoHeight uint64 `json:"topoheight"`
	} // no params
	/* //reponse is same as GetBlockHeaderByHeight_Result
	GetBlockHeaderByHeight_Result struct {
		Block_Header BlockHeader_Print `json:"block_header"`
		Status       string            `json:"status"`
	}*/
)

// GetBlockHeaderByHash
type (
	GetBlockHeaderByHash_Params struct {
		Hash string `json:"hash"`
	} // no params
	GetBlockHeaderByHash_Result struct {
		Block_Header BlockHeader_Print `json:"block_header"`
		Status       string            `json:"status"`
	}
)

// get block count
type (
	GetBlockCount_Params struct {
		// NO params
	}
	GetBlockCount_Result struct {
		Count  uint64 `json:"count"`
		Status string `json:"status"`
	}
)

// getblock
type (
	GetBlock_Params struct {
		Hash   string `json:"hash,omitempty"`   // Monero Daemon breaks if both are provided
		Height uint64 `json:"height,omitempty"` // Monero Daemon breaks if both are provided
	} // no params
	GetBlock_Result struct {
		Blob         string            `json:"blob"`
		Json         string            `json:"json"`
		Block_Header BlockHeader_Print `json:"block_header"`
		Status       string            `json:"status"`
	}
)

// get block template request response
type (
	GetBlockTemplate_Params struct {
		Wallet_Address string `json:"wallet_address"`
		Reserve_size   uint64 `json:"reserve_size"`
	}
	GetBlockTemplate_Result struct {
		Blocktemplate_blob string `json:"blocktemplate_blob"`
		Blockhashing_blob  string `json:"blockhashing_blob"`
		Expected_reward    uint64 `json:"expected_reward"`
		Difficulty         uint64 `json:"difficulty"`
		Height             uint64 `json:"height"`
		Prev_Hash          string `json:"prev_hash"`
		Reserved_Offset    uint64 `json:"reserved_offset"`
		Epoch              uint64 `json:"epoch"` // used to expire pool jobs
		Status             string `json:"status"`
	}
)

type ( // array without name containing block template in hex
	SubmitBlock_Params struct {
		X []string
	}
	SubmitBlock_Result struct {
                BLID string `json:"blid"`
		Status string `json:"status"`
	}
)

type (
	GetLastBlockHeader_Params struct{} // no params
	GetLastBlockHeader_Result struct {
		Block_Header BlockHeader_Print `json:"block_header"`
		Status       string            `json:"status"`
	}
)

type (
	GetTxPool_Params struct{} // no params
	GetTxPool_Result struct {
		Tx_list []string `json:"txs,omitempty"`
		Status  string   `json:"status"`
	}
)

// get height http response as json
type (
	Daemon_GetHeight_Result struct {
		Height       uint64 `json:"height"`
		StableHeight int64  `json:"stableheight"`
		TopoHeight   int64  `json:"topoheight"`

		Status string `json:"status"`
	}
)

type (
	On_GetBlockHash_Params struct {
		X [1]uint64
	}
	On_GetBlockHash_Result struct{}
)

type (
	GetTransaction_Params struct {
		Tx_Hashes []string `json:"txs_hashes"`
		Decode    uint64   `json:"decode_as_json,omitempty"` // Monero Daemon breaks if this sent
	} // no params
	GetTransaction_Result struct {
		Txs_as_hex  []string          `json:"txs_as_hex"`
		Txs_as_json []string          `json:"txs_as_json"`
		Txs         []Tx_Related_Info `json:"txs"`
		Status      string            `json:"status"`
	}

	Tx_Related_Info struct {
		As_Hex         string                     `json:"as_hex"`
		As_Json        string                     `json:"as_json"`
		Block_Height   int64                      `json:"block_height"`
		Reward         uint64                     `json:"reward"`  // miner tx rewards are decided by the protocol during execution
		Ignored        bool                       `json:"ignored"` // tell whether this tx is okau as per client protocol or bein ignored
		In_pool        bool                       `json:"in_pool"`
		Output_Indices []uint64                   `json:"output_indices"`
		Tx_hash        string                     `json:"tx_hash"`
		ValidBlock     string                     `json:"valid_block"`   // TX is valid in this block
		InvalidBlock   []string                   `json:"invalid_block"` // TX is invalid in this block,  0 or more
		Ring           [][]globals.TX_Output_Data `json:"ring"`
	}
)

type (
	Is_Key_Image_Spent_Params struct {
		Key_images []string `json:"key_images"`
	} // no params
	Is_Key_Image_Spent_Result struct {
		Spent_Status []int  `json:"spent_status"` // 0 if okay, 1 spent in block chain, 2 spent in pool
		Status       string `json:"status"`
	}
)

type (
	SendRawTransaction_Params struct {
		Tx_as_hex string `json:"tx_as_hex"`
	}
	SendRawTransaction_Result struct {
		Status        string `json:"status"`
		DoubleSpend   bool   `json:"double_spend"`
		FeeTooLow     bool   `json:"fee_too_low"`
		InvalidInput  bool   `json:"invalid_input"`
		InvalidOutput bool   `json:"invalid_output"`
		Low_Mixin     bool   `json:"low_mixin"`
		Non_rct       bool   `json:"not_rct"`
		NotRelayed    bool   `json:"not_relayed"`
		Overspend     bool   `json:"overspend"`
		TooBig        bool   `json:"too_big"`
		Reason        string `json:"string"`
	}
)

/*
{
  "id": "0",
  "jsonrpc": "2.0",
  "result": {
    "alt_blocks_count": 5,
    "difficulty": 972165250,
    "grey_peerlist_size": 2280,
    "height": 993145,
    "incoming_connections_count": 0,
    "outgoing_connections_count": 8,
    "status": "OK",
    "target": 60,
    "target_height": 993137,
    "testnet": false,
    "top_block_hash": "",
    "tx_count": 564287,
    "tx_pool_size": 45,
    "white_peerlist_size": 529
  }
}*/
type (
	GetInfo_Params struct{} // no params
	GetInfo_Result struct {
		Alt_Blocks_Count           uint64  `json:"alt_blocks_count"`
		Difficulty                 uint64  `json:"difficulty"`
		Grey_PeerList_Size         uint64  `json:"grey_peerlist_size"`
		Height                     int64   `json:"height"`
		StableHeight               int64   `json:"stableheight"`
		TopoHeight                 int64   `json:"topoheight"`
		AverageBlockTime50         float32 `json:"averageblocktime50"`
		Incoming_connections_count uint64  `json:"incoming_connections_count"`
		Outgoing_connections_count uint64  `json:"outgoing_connections_count"`
		Target                     uint64  `json:"target"`
		Target_Height              uint64  `json:"target_height"`
		Testnet                    bool    `json:"testnet"`
		Top_block_hash             string  `json:"top_block_hash"`
		Tx_count                   uint64  `json:"tx_count"`
		Tx_pool_size               uint64  `json:"tx_pool_size"`
		Dynamic_fee_per_kb         uint64  `json:"dynamic_fee_per_kb"` // our addition
		Total_Supply               uint64  `json:"total_supply"`       // our addition
		Median_Block_Size          uint64  `json:"median_block_size"`  // our addition
		White_peerlist_size        uint64  `json:"white_peerlist_size"`
		Version                    string  `json:"version"`

		Status string `json:"status"`
	}
)

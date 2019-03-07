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

// this package contains only struct definitions required to implement wallet rpc (compatible with C daemon)
// in order to avoid the dependency on block chain by any package requiring access to rpc
// and other structures
// having the structures was causing the build times of explorer/wallet to be more than couple of secs
// so separated the logic from the structures

package structures

type (
	GetBalance_Params struct{} // no params
	GetBalance_Result struct {
		Balance          uint64 `json:"balance"`
		Unlocked_Balance uint64 `json:"unlocked_balance"`
	}
)

type (
	GetAddress_Params struct{} // no params
	GetAddress_Result struct {
		Address string `json:"address"`
	}
)

type (
	GetHeight_Params struct{} // no params
	GetHeight_Result struct {
		Height uint64 `json:"height"`
	}
)

type (
	Destination struct {
		Amount  uint64 `json:"amount"`
		Address string `json:"address"`
	}

	Transfer_Params struct {
		Destinations []Destination `json:"destinations"`
		Fee          uint64        `json:"fee"`
		Mixin        uint64        `json:"mixin"`
		Unlock_time  uint64        `json:"unlock_time"`
		Payment_ID   string        `json:"payment_id"`
		Get_tx_key   bool          `json:"get_tx_key"`
		Priority     uint64        `json:"priority"`
		Do_not_relay bool          `json:"do_not_relay"`
		Get_tx_hex   bool          `json:"get_tx_hex"`
	} // no params
	Transfer_Result struct {
		Fee     uint64 `json:"fee"`
		Tx_key  string `json:"tx_key"`
		Tx_hash string `json:"tx_hash"`
		Tx_blob string `json:"tx_blob"`
	}
)

//transfer split
type (
	TransferSplit_Params Transfer_Params
	TransferSplit_Result struct {
		Fee_list     []uint64 `json:"fee_list"`
		Amount_list  []uint64 `json:"amount_list"`
		Tx_key_list  []string `json:"tx_key_list"`
		Tx_hash_list []string `json:"tx_hash_list"`
		Tx_blob_list []string `json:"tx_blob_list"`
	}
)

// each outgoing transaction will have this detail,
// if the wallet is recreated from scratch this information will be lost
type Outgoing_Transfer_Details struct {
	TXID string `json:"txid,omitempty"`
	PaymentID  string  `json:"paymentid,omitempty"` // actual payment id
	Fees  uint64      `json:"fees,omitempty"` // fees
	Amount           []uint64 `json:"amount,omitempty"`    // amount sent
	Daddress   []string `json:"to,omitempty"`// taken from address
	TXsecretkey string `json:"tx_secret_key,omitempty"` //secret key for transaction
}


type (
	Transfer_Details struct {
		TXID        string `json:"tx_hash"`
		Payment_ID  string `json:"payment_id,omitempty"`
		Height      uint64 `json:"block_height"`
		Timestamp   uint64 `json:"timestamp,omitempty"`
		Amount      uint64 `json:"amount"`
		Fees        uint64 `json:"fee,omitempty"`
		Unlock_time uint64 `json:"unlock_time"`
		Destinations []Destination `json:"destinations"`
		Note string `json:"note,omitempty"`
		Type string `json:"type,omitempty"`
	}

	Get_Transfers_Params struct {
		In               bool   `json:"in"`
		Out              bool   `json:"out"`
		Pending          bool   `json:"pending"`
		Failed           bool   `json:"failed"`
		Pool             bool   `json:"pool"`
		Filter_by_Height bool   `json:"filter_by_height"`
		Min_Height       uint64 `json:"min_height"`
		Max_Height       uint64 `json:"max_height"`
	}
	Get_Transfers_Result struct {
		In      []Transfer_Details `json:"in,omitempty"`
		Out     []Transfer_Details `json:"out,omitempty"`
		Pending []Transfer_Details `json:"pending,omitempty"`
		Failed  []Transfer_Details `json:"failed,omitempty"`
		Pool    []Transfer_Details `json:"pool,omitempty"`
	}
)

// Get_Bulk_Payments
type (
	Get_Bulk_Payments_Params struct {
		Payment_IDs      []string `json:"payment_ids"`
		Min_block_height uint64   `json:"min_block_height"`
	} // no params
	Get_Bulk_Payments_Result struct {
		Payments []Transfer_Details `json:"payments,omitempty"`
	}
)

// query_key
type (
	Query_Key_Params struct {
		Key_type string `json:"key_type"`
	} // no params
	Query_Key_Result struct {
		Key string `json:"key"`
	}
)

// make_integrated_address_handler
type (
	Make_Integrated_Address_Params struct {
		Payment_id string `json:"payment_id"` // 16 or 64 hex encoded payment ID
	} // no params
	Make_Integrated_Address_Result struct {
		Integrated_Address string `json:"integrated_address"`
		Payment_id         string `json:"payment_id"`
	}
)

// split_integrated_address_handler
type (
	Split_Integrated_Address_Params struct {
		Integrated_Address string `json:"integrated_address"`
	} // no params
	Split_Integrated_Address_Result struct {
		Standard_Address string `json:"standard_address"`
		Payment_id       string `json:"payment_id"`
	}
)

// Get_Transfer_By_TXID
type (
	Get_Transfer_By_TXID_Params struct {
		TXID string `json:"txid"`
	}
	Get_Transfer_By_TXID_Result struct {
		Transfer Transfer_Details `json:"payments,omitempty"`
	}
)

/*
// GetBlockHeaderByHash
type (

	GetBlockHeaderByHash_Params  struct {
		Hash string `json:"hash"`
	} // no params
	GetBlockHeaderByHash_Result struct {
		Block_Header BlockHeader_Print `json:"block_header"`
		Status       string                       `json:"status"`
	}
)

// get block count
type (
	GetBlockCount_Params  struct {
		// NO params
	}
	GetBlockCount_Result struct {
		Count  uint64 `json:"count"`
		Status string `json:"status"`
	}
)

// getblock
type (

	GetBlock_Params  struct {
		Hash   string `json:"hash,omitempty"`   // Monero Daemon breaks if both are provided
		Height uint64 `json:"height,omitempty"` // Monero Daemon breaks if both are provided
	} // no params
	GetBlock_Result struct {
		Blob         string                       `json:"blob"`
		Json         string                       `json:"json"`
		Block_Header BlockHeader_Print `json:"block_header"`
		Status       string                       `json:"status"`
	}
)


// get block template request response
type (

	GetBlockTemplate_Params  struct {
		Wallet_Address string `json:"wallet_address"`
		Reserve_size   uint64 `json:"reserve_size"`
	}
	GetBlockTemplate_Result struct {
		Blocktemplate_blob string `json:"blocktemplate_blob"`
		Expected_reward   uint64   `json:"expected_reward"`
		Difficulty         uint64 `json:"difficulty"`
		Height             uint64 `json:"height"`
		Prev_Hash          string `json:"prev_hash"`
		Reserved_Offset    uint64 `json:"reserved_offset"`
		Status             string `json:"status"`
	}
)

type (// array without name containing block template in hex
	SubmitBlock_Params  struct {
		X []string
	}
	SubmitBlock_Result struct {
		Status             string `json:"status"`
	}
)


type (

	GetLastBlockHeader_Params  struct{} // no params
	GetLastBlockHeader_Result  struct {
		Block_Header BlockHeader_Print `json:"block_header"`
		Status       string                       `json:"status"`
	}
)


type (

	GetTxPool_Params  struct{} // no params
	GetTxPool_Result  struct {
		Tx_list []string `json:"txs,omitempty"`
		Status  string   `json:"status"`
	}
)

type (

	On_GetBlockHash_Params  struct {
		X [1]uint64
	}
	On_GetBlockHash_Result struct{}
)




type (

	GetTransaction_Params  struct {
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
		As_Hex         string   `json:"as_hex"`
		As_Json        string   `json:"as_json"`
		Block_Height   int64    `json:"block_height"`
		In_pool        bool     `json:"in_pool"`
		Output_Indices []uint64 `json:"output_indices"`
		Tx_hash        string   `json:"tx_hash"`
	}
)

type (

	SendRawTransaction_Params  struct {
		Tx_as_hex string `json:"tx_as_hex"`

	}
	SendRawTransaction_Result struct {
		Status      string            `json:"status"`
		DoubleSpend bool  `json:"double_spend"`
		FeeTooLow bool    `json:"fee_too_low"`
		InvalidInput bool `json:"invalid_input"`
		InvalidOutput bool `json:"invalid_output"`
		Low_Mixin bool    `json:"low_mixin"`
		Non_rct bool  `json:"not_rct"`
		NotRelayed bool `json:"not_relayed"`
		Overspend bool `json:"overspend"`
		TooBig bool   `json:"too_big"`
		Reason  string `json:"string"`

	}

)


*/

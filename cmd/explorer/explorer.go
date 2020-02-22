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

package main

// this file implements the explorer for DERO blockchain
// this needs only RPC access
// NOTE: Only use data exported from within the RPC interface, do direct use of exported variables  fom packages
// NOTE: we can use structs defined within the RPCserver package
// This is being developed to track down and confirm some bugs
// NOTE: This is NO longer entirely compliant with the xyz RPC interface ( the pool part is not compliant), currently and can be used as it for their chain,
// atleast for the last 1 year

// TODO: error handling is non-existant ( as this was built up in hrs ). Add proper error handling
//

import "time"
import "fmt"
import "net"
import "bytes"
import "strings"
import "strconv"
import "encoding/hex"
import "net/http"
import "html/template"
import "encoding/json"
import "io/ioutil"

import "github.com/docopt/docopt-go"
import log "github.com/sirupsen/logrus"
import "github.com/ybbus/jsonrpc"

import "github.com/deroproject/derosuite/block"
import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/transaction"
import "github.com/deroproject/derosuite/structures"
import "github.com/deroproject/derosuite/proof"

var command_line string = `dero_explorer
DERO Atlantis Explorer: A secure, private blockchain with smart-contracts

Usage:
  dero_explorer [--help] [--version] [--debug] [--rpc-server-address=<127.0.0.1:18091>] [--http-address=<0.0.0.0:8080>] 
  dero_explorer -h | --help
  dero_explorer --version

Options:
  -h --help     Show this screen.
  --version     Show version.
  --debug       Debug mode enabled, print log messages
  --rpc-server-address=<127.0.0.1:18091>  connect to this daemon port as client
  --http-address=<0.0.0.0:8080>    explorer listens on this port to serve user requests`

var rpcClient *jsonrpc.RPCClient
var netClient *http.Client
var endpoint string
var replacer = strings.NewReplacer("h", ":", "m", ":", "s", "")

func main() {
	var err error
	var arguments map[string]interface{}

	arguments, err = docopt.Parse(command_line, nil, true, "DERO Explorer : work in progress", false)

	if err != nil {
		log.Fatalf("Error while parsing options err: %s\n", err)
	}

	if arguments["--debug"].(bool) == true {
		log.SetLevel(log.DebugLevel)
	}

	log.Debugf("Arguments %+v", arguments)
	log.Infof("DERO Atlantis Exporer :  This is under heavy development, use it for testing/evaluations purpose only")
	log.Infof("Copyright 2017-2020 DERO Project. All rights reserved.")
	endpoint = "127.0.0.1:30306"
	if arguments["--rpc-server-address"] != nil {
		endpoint = arguments["--rpc-server-address"].(string)
	}

	log.Infof("using RPC endpoint %s", endpoint)

	listen_address := "0.0.0.0:8081"
	if arguments["--http-address"] != nil {
		listen_address = arguments["--http-address"].(string)
	}
	log.Infof("Will listen on %s", listen_address)

	// create client
	rpcClient = jsonrpc.NewRPCClient("http://" + endpoint + "/json_rpc")

	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}

	netClient = &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}

	// execute rpc to service
	response, err := rpcClient.Call("get_info")

	if err == nil {
		log.Infof("Connection to RPC server successful")
	} else {
		log.Fatalf("Connection to RPC server Failed err %s", err)
	}
	var info structures.GetInfo_Result
	err = response.GetObject(&info)

	fmt.Printf("%+v  err %s\n", info, err)

	http.HandleFunc("/search", search_handler)
	http.HandleFunc("/page/", page_handler)
	http.HandleFunc("/block/", block_handler)
	http.HandleFunc("/txpool/", txpool_handler)
	http.HandleFunc("/tx/", tx_handler)
	http.HandleFunc("/", root_handler)

	fmt.Printf("Listening for requests\n")
	err = http.ListenAndServe(listen_address, nil)
	log.Warnf("Listening to port %s err : %s", listen_address, err)

}

// all the tx info which ever needs to be printed
type txinfo struct {
	Hex          string // raw tx
	Height       string // height at which tx was mined
	Depth        int64
	Timestamp    uint64 // timestamp
	Age          string //  time diff from current time
	Block_time   string // UTC time from block header
	Epoch        uint64 // Epoch time
	In_Pool      bool   // whether tx was in pool
	Hash         string // hash for hash
	PrefixHash   string // prefix hash
	Version      int    // version of tx
	Size         string // size of tx in KB
	Sizeuint64   uint64 // size of tx in bytes
	Fee          string // fee in TX
	Feeuint64    uint64 // fee in atomic units
	In           int    // inputs counts
	Out          int    // outputs counts
	Amount       string
	CoinBase     bool     // is tx coin base
	Extra        string   // extra within tx
	Keyimages    []string // key images within tx
	OutAddress   []string // contains output secret key
	OutOffset    []uint64 // contains index offsets
	Type         string   // ringct or ruffct ( bulletproof)
	ValidBlock   string   // the tx is valid in which block
	InvalidBlock []string // the tx is invalid in which block
	Skipped      bool     // this is only valid, when a block is being listed
	Ring_size    int
	Ring         [][]globals.TX_Output_Data

	TXpublickey string
	PayID32     string // 32 byte payment ID
	PayID8      string // 8 byte encrypted payment ID


	Proof_address string // address agains which which the proving ran
	Proof_index  int64  // proof satisfied for which index
	Proof_amount string // decoded amount 
	Proof_PayID8 string // decrypted 8 byte payment id
	Proof_error  string // error if any while decoding proof

}

// any information for block which needs to be printed
type block_info struct {
	Major_Version uint64
	Minor_Version uint64
	Height        int64
	TopoHeight    int64
	Depth         int64
	Timestamp     uint64
	Hash          string
	Tips          []string
	Nonce         uint64
	Fees          string
	Reward        string
	Size          string
	Age           string //  time diff from current time
	Block_time    string // UTC time from block header
	Epoch         uint64 // Epoch time
	Outputs       string
	Mtx           txinfo
	Txs           []txinfo
	Orphan_Status bool
	SyncBlock     bool // whether the block is sync block
	Tx_Count      int
}

// load and setup block_info from rpc
// if hash is less than 64 bytes then it is considered a height parameter
func load_block_from_rpc(info *block_info, block_hash string, recursive bool) (err error) {
	var bl block.Block
	var bresult structures.GetBlock_Result

	var block_height int
	var block_bin []byte
	if len(block_hash) != 64 { // parameter is a height
		fmt.Sscanf(block_hash, "%d", &block_height)
		// user requested block height
		log.Debugf("User requested block at height %d  user input %s", block_height, block_hash)
		response, err := rpcClient.CallNamed("getblock", map[string]interface{}{"height": uint64(block_height)})

		if err != nil {
			return err
		}
		err = response.GetObject(&bresult)
		if err != nil {
			return err
		}

	} else { // parameter is the hex blob

		log.Debugf("User requested block %s", block_hash)
		response, err := rpcClient.CallNamed("getblock", map[string]interface{}{"hash": block_hash})

		if err != nil {
			log.Warnf("err %s ", err)
			return err
		}
		if response.Error != nil {
			log.Warnf("err %s ", response.Error)
			return fmt.Errorf("No Such block or other Error")
		}

		err = response.GetObject(&bresult)
		if err != nil {
			return err
		}
	}

	// fmt.Printf("block %d  %+v\n",i, bresult)
	info.TopoHeight = bresult.Block_Header.TopoHeight
	info.Height = bresult.Block_Header.Height
	info.Depth = bresult.Block_Header.Depth

	duration_second := (uint64(time.Now().UTC().Unix()) - bresult.Block_Header.Timestamp)
	info.Age = replacer.Replace((time.Duration(duration_second) * time.Second).String())
	info.Block_time = time.Unix(int64(bresult.Block_Header.Timestamp), 0).Format("2006-01-02 15:04:05")
	info.Epoch = bresult.Block_Header.Timestamp
	info.Outputs = fmt.Sprintf("%.03f", float32(bresult.Block_Header.Reward)/1000000000000.0)
	info.Size = "N/A"
	info.Hash = bresult.Block_Header.Hash
	//info.Prev_Hash = bresult.Block_Header.Prev_Hash
	info.Tips = bresult.Block_Header.Tips
	info.Orphan_Status = bresult.Block_Header.Orphan_Status
	info.SyncBlock = bresult.Block_Header.SyncBlock
	info.Nonce = bresult.Block_Header.Nonce
	info.Major_Version = bresult.Block_Header.Major_Version
	info.Minor_Version = bresult.Block_Header.Minor_Version
	info.Reward = fmt.Sprintf("%.03f", float32(bresult.Block_Header.Reward)/1000000000000.0)

	block_bin, _ = hex.DecodeString(bresult.Blob)

	//log.Infof("block %+v bresult %+v ", bl, bresult)

	bl.Deserialize(block_bin)

	if recursive {
		// fill in miner tx info

		err = load_tx_from_rpc(&info.Mtx, bl.Miner_TX.GetHash().String()) //TODO handle error

		// miner tx reward is calculated on runtime due to client protocol reasons in dero atlantis
		// feed what is calculated by the daemon
		info.Mtx.Amount = fmt.Sprintf("%.012f", float64(bresult.Block_Header.Reward)/1000000000000)

		log.Infof("loading tx from rpc %s  %s", bl.Miner_TX.GetHash().String(), err)
		info.Tx_Count = len(bl.Tx_hashes)

		fees := uint64(0)
		size := uint64(0)
		// if we have any other tx load them also
		for i := 0; i < len(bl.Tx_hashes); i++ {
			var tx txinfo
			err = load_tx_from_rpc(&tx, bl.Tx_hashes[i].String()) //TODO handle error
			if tx.ValidBlock != bresult.Block_Header.Hash {       // track skipped status
				tx.Skipped = true
			}
			info.Txs = append(info.Txs, tx)
			fees += tx.Feeuint64
			size += tx.Sizeuint64
		}

		info.Fees = fmt.Sprintf("%.03f", float32(fees)/1000000000000.0)
		info.Size = fmt.Sprintf("%.03f", float32(size)/1024)

	}

	return
}

// this will fill up the info struct from the tx
func load_tx_info_from_tx(info *txinfo, tx *transaction.Transaction) (err error) {
	info.Hash = tx.GetHash().String()
	info.PrefixHash = tx.GetPrefixHash().String()
	info.Size = fmt.Sprintf("%.03f", float32(len(tx.Serialize()))/1024)
	info.Sizeuint64 = uint64(len(tx.Serialize()))
	info.Version = int(tx.Version)
	info.Extra = fmt.Sprintf("%x", tx.Extra)
	info.In = len(tx.Vin)
	info.Out = len(tx.Vout)

	if tx.Parse_Extra() {

		// store public key if present
		if _, ok := tx.Extra_map[transaction.TX_PUBLIC_KEY]; ok {
			info.TXpublickey = tx.Extra_map[transaction.TX_PUBLIC_KEY].(crypto.Key).String()
		}

		// store payment IDs if present
		if _, ok := tx.PaymentID_map[transaction.TX_EXTRA_NONCE_ENCRYPTED_PAYMENT_ID]; ok {
			info.PayID8 = fmt.Sprintf("%x", tx.PaymentID_map[transaction.TX_EXTRA_NONCE_ENCRYPTED_PAYMENT_ID].([]byte))
		} else if _, ok := tx.PaymentID_map[transaction.TX_EXTRA_NONCE_PAYMENT_ID]; ok {
			info.PayID32 = fmt.Sprintf("%x", tx.PaymentID_map[transaction.TX_EXTRA_NONCE_PAYMENT_ID].([]byte))
		}

	}

	if !tx.IsCoinbase() {
		info.Fee = fmt.Sprintf("%.012f", float64(tx.RctSignature.Get_TX_Fee())/1000000000000)
		info.Feeuint64 = tx.RctSignature.Get_TX_Fee()
		info.Amount = "?"

		info.Ring_size = len(tx.Vin[0].(transaction.Txin_to_key).Key_offsets)
		for i := 0; i < len(tx.Vin); i++ {
			info.Keyimages = append(info.Keyimages, fmt.Sprintf("%s ring members %+v", tx.Vin[i].(transaction.Txin_to_key).K_image, tx.Vin[i].(transaction.Txin_to_key).Key_offsets))
		}
	} else {
		info.CoinBase = true
		info.In = 0

		//if info.
		//info.Amount = fmt.Sprintf("%.012f", float64(tx.Vout[0].Amount)/1000000000000)
	}

	for i := 0; i < len(tx.Vout); i++ {
		info.OutAddress = append(info.OutAddress, tx.Vout[i].Target.(transaction.Txout_to_key).Key.String())
	}

	// if outputs cannot be located, do not panic
	// this will be the case for pool transactions
	if len(info.OutAddress) != len(info.OutOffset) {
		info.OutOffset = make([]uint64, len(info.OutAddress), len(info.OutAddress))
	}

	switch tx.RctSignature.Get_Sig_Type() {
	case 0:
		info.Type = "RingCT/0"
	case 1:
		info.Type = "RingCT/1 MG"
	case 2:
		info.Type = "RingCT/2 Simple"
	case 3:
		info.Type = "RingCT/3 Full bulletproof"
	case 4:
		info.Type = "RingCT/4 Simple Bulletproof"
	}

	if !info.In_Pool { // find the age of block and other meta
		var blinfo block_info
		err := load_block_from_rpc(&blinfo, fmt.Sprintf("%s", info.Height), false) // we only need block data and not data of txs
		if err != nil {
			return err
		}

		//    fmt.Printf("Blinfo %+v height %d", blinfo, info.Height);

		info.Age = blinfo.Age
		info.Block_time = blinfo.Block_time
		info.Epoch = blinfo.Epoch
		info.Timestamp = blinfo.Epoch
		info.Depth = blinfo.Depth

	}

	return nil
}

// load and setup txinfo from rpc
func load_tx_from_rpc(info *txinfo, txhash string) (err error) {
	var tx_params structures.GetTransaction_Params
	var tx_result structures.GetTransaction_Result

	//fmt.Printf("Requesting tx data %s", txhash);
	tx_params.Tx_Hashes = append(tx_params.Tx_Hashes, txhash)

	request_bytes, err := json.Marshal(&tx_params)
	response, err := http.Post("http://"+endpoint+"/gettransactions", "application/json", bytes.NewBuffer(request_bytes))
	if err != nil {
		fmt.Printf("err while requesting tx err %s\n", err)
		return
	}
	buf, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("err while reading reponse body err %s\n", err)
		return
	}

	err = json.Unmarshal(buf, &tx_result)
	if err != nil {
		fmt.Printf("err while parsing reponse body err %s\n", err)
		return
	}

	// fmt.Printf("TX response %+v", tx_result)

	if tx_result.Status != "OK" {
		return fmt.Errorf("No Such TX RPC error status %s", tx_result.Status)
	}

	var tx transaction.Transaction

	if len(tx_result.Txs_as_hex[0]) < 50 {
		return
	}

	info.Hex =  tx_result.Txs_as_hex[0]

	tx_bin, _ := hex.DecodeString(tx_result.Txs_as_hex[0])
	tx.DeserializeHeader(tx_bin)

	// fill as much info required from headers
	if tx_result.Txs[0].In_pool {
		info.In_Pool = true
	} else {
		info.Height = fmt.Sprintf("%d", tx_result.Txs[0].Block_Height)
	}

	for x := range tx_result.Txs[0].Output_Indices {
		info.OutOffset = append(info.OutOffset, tx_result.Txs[0].Output_Indices[x])
	}

	if tx.IsCoinbase() { // fill miner tx reward from what the chain tells us
		info.Amount = fmt.Sprintf("%.012f", float64(uint64(tx_result.Txs[0].Reward))/1000000000000)
	}

	info.ValidBlock = tx_result.Txs[0].ValidBlock
	info.InvalidBlock = tx_result.Txs[0].InvalidBlock

	info.Ring = tx_result.Txs[0].Ring

	//fmt.Printf("tx_result %+v\n",tx_result.Txs)

	return load_tx_info_from_tx(info, &tx)
}

func block_handler(w http.ResponseWriter, r *http.Request) {
	param := ""
	fmt.Sscanf(r.URL.EscapedPath(), "/block/%s", &param)

	var blinfo block_info
	err := load_block_from_rpc(&blinfo, param, true)
	_ = err

	// execute template now
	data := map[string]interface{}{}

	data["title"] = "DERO Atlantis BlockChain Explorer(v1)"
	data["servertime"] = time.Now().UTC().Format("2006-01-02 15:04:05")
	data["block"] = blinfo

	t, err := template.New("foo").Parse(header_template + block_template + footer_template + notfound_page_template)

	err = t.ExecuteTemplate(w, "block", data)
	if err != nil {
		return
	}

	return

	//     fmt.Fprint(w, "This is a valid block")

}

func tx_handler(w http.ResponseWriter, r *http.Request) {
	var info txinfo
	tx_hex := ""
	fmt.Sscanf(r.URL.EscapedPath(), "/tx/%s", &tx_hex)
	txhash := crypto.HashHexToHash(tx_hex)
	log.Debugf("user requested TX %s", tx_hex)

	err := load_tx_from_rpc(&info, txhash.String()) //TODO handle error
	_ = err

	// check whether user requested proof

	tx_secret_key := r.PostFormValue("txprvkey")
	daddress :=  r.PostFormValue("deroaddress")
	raw_tx_data := r.PostFormValue("raw_tx_data")

	if raw_tx_data !=  "" { // gives ability to prove transactions not in the blockchain
		info.Hex = raw_tx_data
	}

	log.Debugf("tx key %s address %s  tx %s", tx_secret_key, daddress, tx_hex)
	
	if tx_secret_key != ""  && daddress != "" {

	// there may be more than 1 amounts, only first one is shown
	indexes, amounts, payids, err := proof.Prove(tx_secret_key,daddress,info.Hex)

	_ = indexes
	_ = amounts
	if err == nil { //&& len(amounts) > 0 && len(indexes) > 0{
		log.Debugf("Successfully proved transaction %s len(payids) %d",tx_hex, len(payids))
		info.Proof_index = int64(indexes[0])
		info.Proof_address =  daddress
		info.Proof_amount = globals.FormatMoney12(amounts[0])
		if len(payids) >=1 {
			info.Proof_PayID8 = fmt.Sprintf("%x", payids[0]) // decrypted payment ID
		}
	}else{
		log.Debugf("err while proving %s",err)
		if err != nil {
		info.Proof_error = err.Error()
	}

	}
}




	// execute template now
	data := map[string]interface{}{}

	data["title"] = "DERO Atlantis BlockChain Explorer(v1)"
	data["servertime"] = time.Now().UTC().Format("2006-01-02 15:04:05")
	data["info"] = info

	t, err := template.New("foo").Parse(header_template + tx_template + footer_template + notfound_page_template)

	err = t.ExecuteTemplate(w, "tx", data)
	if err != nil {
		return
	}

	return

}

func pool_handler(w http.ResponseWriter, r *http.Request) {

	fmt.Fprint(w, "This is a valid pool")

}

// if there is any error, we return back empty
// if pos is wrong we return back
// pos is descending order
func fill_tx_structure(pos int, size_in_blocks int) (data []block_info) {

	for i := pos; i > (pos-size_in_blocks) && i >= 0; i-- { // query blocks by topo height
		var blinfo block_info
		err := load_block_from_rpc(&blinfo, fmt.Sprintf("%d", i), true)
		if err == nil {
			data = append(data, blinfo)
		}
	}
	return
}

func show_page(w http.ResponseWriter, page int) {
	data := map[string]interface{}{}
	var info structures.GetInfo_Result

	data["title"] = "DERO Atlantis BlockChain Explorer(v1)"
	data["servertime"] = time.Now().UTC().Format("2006-01-02 15:04:05")

	t, err := template.New("foo").Parse(header_template + txpool_template + main_template + paging_template + footer_template + notfound_page_template)

	// collect all the data afresh
	// execute rpc to service
	response, err := rpcClient.Call("get_info")

	if err != nil {
		goto exit_error
	}

	err = response.GetObject(&info)
	if err != nil {
		goto exit_error
	}

	//fmt.Printf("get info %+v", info)

	data["Network_Difficulty"] = info.Difficulty
	data["hash_rate"] = fmt.Sprintf("%.03f", float32(info.Difficulty)/float32(info.Target*1000))
	data["txpool_size"] = info.Tx_pool_size
	data["testnet"] = info.Testnet
	data["averageblocktime50"] = info.AverageBlockTime50
	data["fee_per_kb"] = float64(info.Dynamic_fee_per_kb) / 1000000000000
	data["median_block_size"] = fmt.Sprintf("%.02f", float32(info.Median_Block_Size)/1024)
	data["total_supply"] = info.Total_Supply

	if page == 0  { // use requested invalid page, give current page
		page = int(info.TopoHeight)/10
	}

	data["previous_page"] = page - 1
	if page <= 1 {
		data["previous_page"] = 1
	}
	data["current_page"] = page
	if (int(info.TopoHeight) % 10)  == 0 {
		data["total_page"] = (int(info.TopoHeight) / 10)  
	}else{
		data["total_page"] = (int(info.TopoHeight) / 10)  
	}
	

	data["next_page"] = page + 1
	if (page + 1) > data["total_page"].(int) {
		data["next_page"] = page
	}

	fill_tx_pool_info(data, 25)

	if page == 1{ // page 1 has 11 blocks, it does not show genesis block
		data["block_array"] = fill_tx_structure(int(page*10), 12)
		}else{
			if int(info.TopoHeight)-int(page*10) > 10{
				data["block_array"] = fill_tx_structure(int(page*10), 10)	
			}else{
				data["block_array"] = fill_tx_structure(int(info.TopoHeight), int(info.TopoHeight)-int(page*10))	
			}
			
		}
	

	err = t.ExecuteTemplate(w, "main", data)
	if err != nil {
		goto exit_error
	}

	return

exit_error:
	fmt.Fprintf(w, "Error occurred err %s", err)

}

func txpool_handler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{}
	var info structures.GetInfo_Result

	data["title"] = "DERO Atlantis BlockChain Explorer(v1)"
	data["servertime"] = time.Now().UTC().Format("2006-01-02 15:04:05")

	t, err := template.New("foo").Parse(header_template + txpool_template + main_template + paging_template + footer_template + txpool_page_template + notfound_page_template)

	// collect all the data afresh
	// execute rpc to service
	response, err := rpcClient.Call("get_info")

	if err != nil {
		goto exit_error
	}

	err = response.GetObject(&info)
	if err != nil {
		goto exit_error
	}

	//fmt.Printf("get info %+v", info)

	data["Network_Difficulty"] = info.Difficulty
	data["hash_rate"] = fmt.Sprintf("%.03f", float32(info.Difficulty/1000000)/float32(info.Target))
	data["txpool_size"] = info.Tx_pool_size
	data["testnet"] = info.Testnet
	data["fee_per_kb"] = float64(info.Dynamic_fee_per_kb) / 1000000000000
	data["median_block_size"] = fmt.Sprintf("%.02f", float32(info.Median_Block_Size)/1024)
	data["total_supply"] = info.Total_Supply
	data["averageblocktime50"] = info.AverageBlockTime50

	fill_tx_pool_info(data, 500) // show only 500 txs

	err = t.ExecuteTemplate(w, "txpool_page", data)
	if err != nil {
		goto exit_error
	}

	return

exit_error:
	fmt.Fprintf(w, "Error occurred err %s", err)

}

// shows a page
func page_handler(w http.ResponseWriter, r *http.Request) {
	page := 0
	page_string := r.URL.EscapedPath()
	fmt.Sscanf(page_string, "/page/%d", &page)
	log.Debugf("user requested page %d", page)
	show_page(w, page)
}

// root shows page 0
func root_handler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("Showing main page")
	show_page(w, 0)
}

// search handler, finds the items using rpc bruteforce
func search_handler(w http.ResponseWriter, r *http.Request) {
	var info structures.GetInfo_Result

	log.Debugf("Showing search page")

	values, ok := r.URL.Query()["value"]

	if !ok || len(values) < 1 {
		show_page(w, 0)
		return
	}



	// Query()["key"] will return an array of items,
	// we only want the single item.
	value := strings.TrimSpace(values[0])
	good := false

	response, err := rpcClient.Call("get_info")
	if err != nil {
		goto exit_error
	}
	err = response.GetObject(&info)
	if err != nil {
		goto exit_error
	}

	if len(value) != 64 {
		if s, err := strconv.ParseInt(value, 10, 64); err == nil && s >= 0 && s <= info.TopoHeight{
					good = true
				}
	}else{ // check whether the string can be hex decoded
		t, err := hex.DecodeString(value)
		if err != nil || len(t) != 32{

			}else{
				good = true
			}
	}


	// value should be either 64 hex chars or a topoheight which should be less than current topoheight

if good{
	// check whether the page is block or tx or height
	var blinfo block_info
	var tx txinfo
	err := load_block_from_rpc(&blinfo, value, false)
	if err == nil {
		log.Debugf("Redirecting user to block page")
		http.Redirect(w, r, "/block/"+value, 302)
		return
	}

	err = load_tx_from_rpc(&tx, value) //TODO handle error
	if err == nil {
		log.Debugf("Redirecting user to tx page")
		http.Redirect(w, r, "/tx/"+value, 302)

		return
	}
}

	{ // show error page
		data := map[string]interface{}{}
	var info structures.GetInfo_Result

	data["title"] = "DERO Atlantis BlockChain Explorer(v1)"
	data["servertime"] = time.Now().UTC().Format("2006-01-02 15:04:05")

	t, err := template.New("foo").Parse(header_template + txpool_template + main_template + paging_template + footer_template + txpool_page_template + notfound_page_template)

	// collect all the data afresh
	// execute rpc to service
	

	err = response.GetObject(&info)
	if err != nil {
		goto exit_error
	}

	//fmt.Printf("get info %+v", info)

	data["Network_Difficulty"] = info.Difficulty
	data["hash_rate"] = fmt.Sprintf("%.03f", float32(info.Difficulty/1000000)/float32(info.Target))
	data["txpool_size"] = info.Tx_pool_size
	data["testnet"] = info.Testnet
	data["fee_per_kb"] = float64(info.Dynamic_fee_per_kb) / 1000000000000
	data["median_block_size"] = fmt.Sprintf("%.02f", float32(info.Median_Block_Size)/1024)
	data["total_supply"] = info.Total_Supply
	data["averageblocktime50"] = info.AverageBlockTime50

	err = t.ExecuteTemplate(w, "notfound_page", data)
	if err == nil {
		return

	}

	}
exit_error:
	show_page(w, 0)
	return

}

// fill all the tx pool info as per requested
func fill_tx_pool_info(data map[string]interface{}, max_count int) error {

	var txs []txinfo
	var txpool structures.GetTxPool_Result

	data["mempool"] = txs // initialize with empty data
	// collect all the data afresh
	// execute rpc to service
	response, err := rpcClient.Call("gettxpool")

	if err != nil {
		return fmt.Errorf("gettxpool rpc failed")
	}

	err = response.GetObject(&txpool)
	if err != nil {
		return fmt.Errorf("gettxpool rpc failed")
	}

	for i := range txpool.Tx_list {
		var info txinfo
		err := load_tx_from_rpc(&info, txpool.Tx_list[i]) //TODO handle error
		if err != nil {
			continue
		}
		txs = append(txs, info)

		if len(txs) >= max_count {
			break
		}
	}

	data["mempool"] = txs
	return nil

}

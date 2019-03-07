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

package walletapi

/* this file handles communication with the daemon
 * this includes receiving output information
 * *
 */
import "io"
import "os"
import "fmt"
import "net"
import "time"
import "sync"
import "bytes"
import "net/http"
import "bufio"
import "strings"
import "runtime"
import "compress/gzip"
import "encoding/hex"
import "encoding/json"
import "runtime/debug"

import "github.com/romana/rlog"

//import "github.com/pierrec/lz4"
import "github.com/ybbus/jsonrpc"
import "github.com/vmihailenco/msgpack"

import "github.com/deroproject/derosuite/config"
import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/structures"
import "github.com/deroproject/derosuite/transaction"

// this global variable should be within wallet structure
var Connected bool = false

// there should be no global variables, so multiple wallets can run at the same time with different assset
var rpcClient *jsonrpc.RPCClient
var netClient *http.Client
var endpoint string

var output_lock sync.Mutex

// returns whether wallet was online some time ago
func (w *Wallet) IsDaemonOnlineCached() bool {
	return Connected
}

// currently process url  with compatibility for older ip address
func buildurl(endpoint string) string {
    if strings.IndexAny(endpoint,"abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") >= 0 { // url is already complete
        return strings.TrimSuffix(endpoint,"/")
    }else{
        return "http://" + endpoint
    }
    
    
}

// this is as simple as it gets
// single threaded communication to get the daemon status and height
// this will tell whether the wallet can connection successfully to  daemon or not
func (w *Wallet) IsDaemonOnline() (err error) {

    if globals.Arguments["--remote"] == true  && globals.IsMainnet() {
        w.Daemon_Endpoint = config.REMOTE_DAEMON
    }

	// if user provided endpoint has error, use default
	if w.Daemon_Endpoint == "" {
		w.Daemon_Endpoint = "127.0.0.1:" + fmt.Sprintf("%d", config.Mainnet.RPC_Default_Port)
		if !globals.IsMainnet() {
			w.Daemon_Endpoint = "127.0.0.1:" + fmt.Sprintf("%d", config.Testnet.RPC_Default_Port)
		}
	}

	if globals.Arguments["--daemon-address"] != nil {
		w.Daemon_Endpoint = globals.Arguments["--daemon-address"].(string)
	}

	rlog.Infof("Daemon endpoint %s", w.Daemon_Endpoint)

	// TODO enable socks support here
	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second, // 5 second timeout
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}

	netClient = &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}

	// create client
	rpcClient = jsonrpc.NewRPCClient(buildurl(w.Daemon_Endpoint) + "/json_rpc")

	// execute rpc to service
	response, err := rpcClient.Call("get_info")

        // notify user of any state change
	// if daemon connection breaks or comes live again
	if err == nil {
		if !Connected {
			rlog.Infof("Connection to RPC server successful %s", buildurl(w.Daemon_Endpoint))
			Connected = true
		}
	} else {
		rlog.Errorf("Error executing getinfo_rpc err %s", err)

		if Connected {
			rlog.Warnf("Connection to RPC server Failed err %s endpoint %s ", err, buildurl(w.Daemon_Endpoint))
			fmt.Printf("Connection to RPC server Failed err %s endpoint %s ", err, buildurl(w.Daemon_Endpoint))

		}
		Connected = false

		return
	}
	var info structures.GetInfo_Result
	err = response.GetObject(&info)
	if err != nil {
		rlog.Errorf("Daemon getinfo RPC parsing error err: %s\n", err)
		Connected = false
		return
	}
	// detect whether both are in different modes
	//  daemon is in testnet and wallet in mainnet or
	// daemon
	if info.Testnet != !globals.IsMainnet() {
		err = fmt.Errorf("Mainnet/TestNet  is different between wallet/daemon.Please run daemon/wallet without --testnet")
		rlog.Criticalf("%s", err)
		return
	}

	w.Lock()
	defer w.Unlock()

	if info.Height >= 0 {
		w.Daemon_Height = uint64(info.Height)
		w.Daemon_TopoHeight = info.TopoHeight
	}
	w.dynamic_fees_per_kb = info.Dynamic_fee_per_kb // set fee rate, it can work for quite some time,

	return nil
}

// do the entire sync
// lagging behind is the NOT the major problem
// the problem is the frequent soft-forks
// once an input has been detected, we must check whether the block is not orphan
// we must do this without leaking ourselves to the daemon itself
// this means, we will have to keep track of the chain in the wallet also, ofcouse for syncing purposes only
// we have a bucket where we store  topo height to block hash links , of course in encrypted form
// during syncing we do a a binary search style matching and then sync up from that point
func (w *Wallet) DetectSyncPoint() (start_sync_at_height uint64, err error) {
	err = w.IsDaemonOnline()
	if err != nil { // if we cannot connect with daemon bail out now
		return
	}

	rlog.Tracef(2, "Detection of Sync Point started")

	w.Lock()

	minimum := int64(0)
	maximum := int64(w.account.TopoHeight)

	// if wallet height >  daemon height, we can only do sanity chec
	if maximum > w.Daemon_TopoHeight {
		w.Unlock()
		err = fmt.Errorf("Wallet can never be ahead than daemon. Make sure daemon is synced")
		return
	}
	w.Unlock()

	//rlog.Debugf("min %d max %d", minimum, maximum)

	//corruption_point := uint64(40000)
	// maximum is overridden the the current blockchain height
	//base := minimum
	for minimum <= maximum { //  with just 24 requests we can find, the mismatch point,in 16 million chain height
		// get hash at height (maximum-minimum)/2
		// compare it with what is stored in wallet
		// if match, increase minimum
		// of no match , decrease maximum
		median := ((minimum + maximum) / 2)

		// try to load first
		rlog.Tracef(4, "min %d max %d median %d\n", minimum, maximum, median)

		var local_hash []byte
		var response *jsonrpc.RPCResponse
		local_hash, err = w.load_key_value(BLOCKCHAIN_UNIVERSE, []byte(HEIGHT_TO_BLOCK_BUCKET), itob(uint64(median)))
		if err != nil {
			maximum = median - 1
			continue
		}

		/* if len(local_hash) == 32 && median >= corruption_point {
		   local_hash[0]=0;
		   local_hash[1]=1;
		   local_hash[2]=2;
		   local_hash[3]=3;
		   local_hash[4]=4;

		  }*/

		response, err = rpcClient.CallNamed("getblockheaderbytopoheight", map[string]interface{}{"topoheight": median})
		if err != nil {
			rlog.Errorf("Connection to RPC server Failed err %s", err)
			return
		}

		// parse response
		if response.Error != nil {
			rlog.Errorf("Connection to RPC server Failed err %s", response.Error)
			return
		}

		var bresult structures.GetBlockHeaderByHeight_Result

		err = response.GetObject(&bresult)
		if err != nil {
			return // err
		}
		if bresult.Status != "OK" {
			err = fmt.Errorf("%s", bresult.Status)
			return
		}
		rlog.Tracef(4, "hash %s local_hash %s \n", bresult.Block_Header.Hash, fmt.Sprintf("%x", local_hash))
		if fmt.Sprintf("%x", local_hash) == bresult.Block_Header.Hash {
			minimum = median + 1
		} else {
			maximum = median - 1
		}
	}

	if minimum >= 1 {
		minimum--
	}

	// we should start syncing from the minimum, this will help us override any soft-forks, though however deep them may be
	start_sync_at_height = uint64(minimum)
	rlog.Infof("sync height  %d\n", start_sync_at_height)

	return
}

// get the outputs from the daemon, requesting specfic outputs
// the range can be anything
// if stop is zero,
// the daemon will flush out everything it has ASAP
// the stream can be saved and used later on

func (w *Wallet) Sync_Wallet_With_Daemon() {

	w.IsDaemonOnline()
	output_lock.Lock()
	defer output_lock.Unlock()

	// only sync if both height are different
	if w.Daemon_TopoHeight == w.account.TopoHeight && w.account.TopoHeight != 0 { // wallet is already synced
		return
	}

	rlog.Infof("wallet topo height %d daemon online topo height %d\n", w.account.TopoHeight, w.Daemon_TopoHeight)

	start_height, err := w.DetectSyncPoint()
	if err != nil {
		rlog.Errorf("Error while detecting sync point err %s", err)
		return
	}

	// the safety cannot be tuned off in openbsd, see boltdb  documentation
	// if we are doing major rescanning, turn of db safety features
	// they will be activated again on resync
	if (w.Daemon_TopoHeight - int64(start_height)) > 50 { // get db into async mode
		w.Lock()
		w.db.NoSync = true
		w.Unlock()
		defer func() {
			w.Lock()
			w.db.NoSync = false
			w.db.Sync()
			w.Unlock()
		}()
	}

	rlog.Infof("requesting outputs from height %d\n", start_height)

	response, err := http.Get(fmt.Sprintf("%s/getoutputs.bin?startheight=%d",buildurl(w.Daemon_Endpoint), start_height))
	if err != nil {
		rlog.Errorf("Error while requesting outputs from daemon err %s", err)
	} else {
		defer response.Body.Close()
		gzipreader, err := gzip.NewReader(response.Body)
		if err != nil {
			rlog.Errorf("Error while decompressing output from daemon  err: %s   ", err)
			return
		}
		defer gzipreader.Close()

		// use the reader and feed the error free stream, if error occurs bailout
		decoder := msgpack.NewDecoder(gzipreader)
		rlog.Debugf("Scanning started\n")

		workers := make(chan int, runtime.GOMAXPROCS(0))
		for i := 0; i < runtime.GOMAXPROCS(0); i++ {
			workers <- i
		}
		for {
			var output globals.TX_Output_Data

			err = decoder.Decode(&output)
			if err == io.EOF { // reached eof
				break
			}
			if err != nil {
				rlog.Errorf("err while decoding msgpack stream err %s\n", err)
				break
			}

			select { // quit midway if required
			case <-w.quit:
				return
			default:
			}

			<-workers
			// try to consume all data sent by the daemon
			go func() {
				w.Add_Transaction_Record_Funds(&output) // add the funds to wallet if they are ours
				workers <- 0
			}()

		}

		rlog.Debugf("Scanning finised\n")

	}

	return
}

// triggers syncing with wallet every 5 seconds
func (w *Wallet) sync_loop() {
	for {
		select { // quit midway if required
		case <-w.quit:
			return
		case <-time.After(5 * time.Second):
		}

		if !w.wallet_online_mode { // wallet requested to be in offline mode
			return
		}

		w.Sync_Wallet_With_Daemon() // sync with the daemon
		//TODO we must sync up with pool also
	}
}

func (w *Wallet) Rescan_From_Height(startheight uint64) {
	w.Lock()
	defer w.Unlock()
	if startheight < uint64(w.account.TopoHeight) {
		w.account.TopoHeight = int64(startheight) // we will rescan from this height
	}

}

// offline file is scanned from start till finish
func (w *Wallet) Scan_Offline_File(filename string) {
	w.Lock()
	defer w.Unlock()

	f, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Cannot read offline data file=\"%s\"  err: %s   ", filename, err)
		return
	}
	bufreader := bufio.NewReader(f)
	gzipreader, err := gzip.NewReader(bufreader)
	if err != nil {
		fmt.Printf("Error while decompressing offline data file=\"%s\"  err: %s   ", filename, err)
		return
	}
	defer gzipreader.Close()

	// use the reader and feed the error free stream, if error occurs bailout
	decoder := msgpack.NewDecoder(gzipreader)
	rlog.Debugf("Scanning started")
	for {
		var output globals.TX_Output_Data

		err = decoder.Decode(&output)
		if err == io.EOF { // reached eof
			break
		}
		if err != nil {
			fmt.Printf("err while decoding msgpack stream err %s\n", err)
			break
		}
		// try to consume all data sent by the daemon
		w.Add_Transaction_Record_Funds(&output) // add the funds to wallet if they are ours

	}
	rlog.Debugf("Scanning finised")

}

// this is as simple as it gets
// single threaded communication to relay TX to daemon
// if this is successful, then daemon is in control

func (w *Wallet) SendTransaction(tx *transaction.Transaction) (err error) {

	if tx == nil {
		return fmt.Errorf("Can not send nil transaction")
	}

	var params structures.SendRawTransaction_Params
	var result structures.SendRawTransaction_Result

	params.Tx_as_hex = hex.EncodeToString(tx.Serialize())

	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(&params)
	if err != nil {
		return
	}

	// this method is NOT JSON RPC method, send raw as http request and parse response
	resp, err := http.Post(fmt.Sprintf("%s/sendrawtransaction", buildurl(w.Daemon_Endpoint)), "application/json", &buf)
	if err != nil {
		return
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&result)
	if err != nil {
		err = fmt.Errorf("err while decoding incoming sendrawtransaction response json err: %s", err)
		return
	}

	if result.Status == "OK" {
		return nil
	} else {
		err = fmt.Errorf("Err %s", result.Status)
	}

	//fmt.Printf("err in response %+v", result)

	return
}

// this is as simple as it gets
// single threaded communication  gets whether the the key image is spent in pool or in blockchain
// this can leak informtion which keyimage belongs to us
// TODO in order to stop privacy leaks we must guess this information somehow on client side itself
// maybe the server can broadcast a bloomfilter or something else from the mempool keyimages
//
func (w *Wallet) IsKeyImageSpent(keyimage crypto.Key) (spent bool) {

	defer func() {
		if r := recover(); r != nil {
			rlog.Warnf("Recovered while adding new block, Stack trace below keyimage %s", keyimage)
			rlog.Warnf("Stack trace  \n%s", debug.Stack())
			spent = false
		}
	}()

	if !w.GetMode() { // if wallet is in offline mode , we cannot do anything
		return false
	}

	spent = true // default assume the funds are spent

	rlog.Warnf("checking whether key image are spent in pool %s", keyimage)

	var params structures.Is_Key_Image_Spent_Params
	var result structures.Is_Key_Image_Spent_Result

	params.Key_images = append(params.Key_images, keyimage.String())

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(&params)
	if err != nil {
		return
	}

	// this method is NOT JSON RPC method, send raw as http request and parse response

	resp, err := http.Post(fmt.Sprintf("%s/is_key_image_spent",buildurl(w.Daemon_Endpoint)), "application/json", &buf)
	if err != nil {
		return
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&result)
	if err != nil {
		err = fmt.Errorf("err while decoding incoming sendrawtransaction response json err: %s", err)
		return
	}

	if result.Status == "OK" {
		if len(result.Spent_Status) == 1 && result.Spent_Status[0] >= 1 {
			return true
		} else {
			return false // if daemon says not spent return as available
		}

	}

	//fmt.Printf("err in response %+v", result)

	return
}

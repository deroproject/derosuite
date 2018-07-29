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

// the rpc server is an extension of walletapi and doesnot employ any global variables
// so a number can be simultaneously active ( based on resources)
package walletapi

import "io"

//import "fmt"
import "time"
import "sync"
import "log"
import "strings"
import "net/http"

//import "github.com/intel-go/fastjson"
import "github.com/osamingo/jsonrpc"

import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/structures"

// all components requiring access to wallet must use , this struct to communicate
// this structure must be update while mutex
type RPCServer struct {
	address          string
	srv              *http.Server
	mux              *http.ServeMux
	mr               *jsonrpc.MethodRepository
	Exit_Event       chan bool // wallet is shutting down and we must quit ASAP
	Exit_In_Progress bool

	w *Wallet // reference to the wallet which is open
	sync.RWMutex
}

func RPCServer_Start(w *Wallet, address string) (*RPCServer, error) {

	//var err error
	var r RPCServer

	//_ = err

	r.Exit_Event = make(chan bool)
	r.w = w
	r.address = address

	go r.Run()
	//logger.Infof("RPC server started")

	return &r, nil
}

// shutdown the rpc server component
func (r *RPCServer) RPCServer_Stop() {
	r.srv.Shutdown(nil) // shutdown the server
	r.Exit_In_Progress = true
	close(r.Exit_Event) // send signal to all connections to exit
	// TODO we  must wait for connections to kill themselves
	time.Sleep(1 * time.Second)
	//logger.Infof("RPC Shutdown")

}

func (r *RPCServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	basic_auth_enabled := false
	var parts []string

	if globals.Arguments["--rpc-login"] != nil {
		userpass := globals.Arguments["--rpc-login"].(string)
		parts = strings.SplitN(userpass, ":", 2)

		basic_auth_enabled = true
		/*if len(parts) != 2 { // these checks are done and verified during program init
		  globals.Logger.Warnf("RPC user name or password invalid")
		  return
		 }*/
		//log.Infof("RPC username \"%s\" password \"%s\" ", parts[0],parts[1])
	}

	if basic_auth_enabled {
		u, p, ok := req.BasicAuth()
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		if u != parts[0] || p != parts[1] {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

	}
	// log.Printf("basic_auth_handler") serve if everything looks okay
	r.mr.ServeHTTP(w, req)
}

// setup handlers
func (r *RPCServer) Run() {

	mr := jsonrpc.NewMethodRepository()
	r.mr = mr

	// install getbalance handler
	if err := mr.RegisterMethod("getbalance", GetBalance_Handler{r: r}, structures.GetBalance_Params{}, structures.GetBalance_Result{}); err != nil {
		log.Fatalln(err)
	}

	// install getaddress handler
	if err := mr.RegisterMethod("getaddress", GetAddress_Handler{r: r}, structures.GetAddress_Params{}, structures.GetBalance_Result{}); err != nil {
		log.Fatalln(err)
	}

	// install getheight handler
	if err := mr.RegisterMethod("getheight", GetHeight_Handler{r: r}, structures.GetHeight_Params{}, structures.GetBalance_Result{}); err != nil {
		log.Fatalln(err)
	}

	// install transfer handler
	if err := mr.RegisterMethod("transfer", Transfer_Handler{r: r}, structures.Transfer_Params{}, structures.Transfer_Result{}); err != nil {
		log.Fatalln(err)
	}
	// install transfer_split handler
	if err := mr.RegisterMethod("transfer_split", TransferSplit_Handler{r: r}, structures.TransferSplit_Params{}, structures.TransferSplit_Result{}); err != nil {
		log.Fatalln(err)
	}

	// install get_bulk_payments handler
	if err := mr.RegisterMethod("get_bulk_payments", Get_Bulk_Payments_Handler{r: r}, structures.Get_Bulk_Payments_Params{}, structures.Get_Bulk_Payments_Result{}); err != nil {
		log.Fatalln(err)
	}

	// install query_key handler
	if err := mr.RegisterMethod("query_key", Query_Key_Handler{r: r}, structures.Query_Key_Params{}, structures.Query_Key_Result{}); err != nil {
		log.Fatalln(err)
	}

	// install make_integrated_address handler
	if err := mr.RegisterMethod("make_integrated_address", Make_Integrated_Address_Handler{r: r}, structures.Make_Integrated_Address_Params{}, structures.Make_Integrated_Address_Result{}); err != nil {
		log.Fatalln(err)
	}

	// install split_integrated_address handler
	if err := mr.RegisterMethod("split_integrated_address", Split_Integrated_Address_Handler{r: r}, structures.Split_Integrated_Address_Params{}, structures.Split_Integrated_Address_Result{}); err != nil {
		log.Fatalln(err)
	}

	// install get_transfer_by_txid handler
	if err := mr.RegisterMethod("get_transfer_by_txid", Get_Transfer_By_TXID_Handler{r: r}, structures.Get_Transfer_By_TXID_Params{}, structures.Get_Transfer_By_TXID_Result{}); err != nil {
		log.Fatalln(err)
	}

	// install get_transfers
	if err := mr.RegisterMethod("get_transfers", Get_Transfers_Handler{r: r}, structures.Get_Transfers_Params{}, structures.Get_Transfers_Result{}); err != nil {
		log.Fatalln(err)
	}

	/*
		if err := mr.RegisterMethod("Main.Echo", EchoHandler{}, EchoParams{}, EchoResult{}); err != nil {
			log.Fatalln(err)
		}

		// install getblockcount handler
		if err := mr.RegisterMethod("getblockcount", GetBlockCount_Handler{}, structures.GetBlockCount_Params{}, structures.GetBlockCount_Result{}); err != nil {
			log.Fatalln(err)
		}

		// install on_getblockhash
		if err := mr.RegisterMethod("on_getblockhash", On_GetBlockHash_Handler{}, structures.On_GetBlockHash_Params{}, structures.On_GetBlockHash_Result{}); err != nil {
			log.Fatalln(err)
		}

		// install getblocktemplate handler
		if err := mr.RegisterMethod("getblocktemplate", GetBlockTemplate_Handler{}, structures.GetBlockTemplate_Params{}, structures.GetBlockTemplate_Result{}); err != nil {
			log.Fatalln(err)
		}

		// submitblock handler
		if err := mr.RegisterMethod("submitblock", SubmitBlock_Handler{}, structures.SubmitBlock_Params{}, structures.SubmitBlock_Result{}); err != nil {
			log.Fatalln(err)
		}

		if err := mr.RegisterMethod("getlastblockheader", GetLastBlockHeader_Handler{}, structures.GetLastBlockHeader_Params{}, structures.GetLastBlockHeader_Result{}); err != nil {
			log.Fatalln(err)
		}

		if err := mr.RegisterMethod("getblockheaderbyhash", GetBlockHeaderByHash_Handler{}, structures.GetBlockHeaderByHash_Params{}, structures.GetBlockHeaderByHash_Result{}); err != nil {
			log.Fatalln(err)
		}

		if err := mr.RegisterMethod("getblockheaderbyheight", GetBlockHeaderByHeight_Handler{}, structures.GetBlockHeaderByHeight_Params{}, structures.GetBlockHeaderByHeight_Result{}); err != nil {
			log.Fatalln(err)
		}
		if err := mr.RegisterMethod("getblock", GetBlock_Handler{}, structures.GetBlock_Params{}, structures.GetBlock_Result{}); err != nil {
			log.Fatalln(err)
		}

		if err := mr.RegisterMethod("get_info", GetInfo_Handler{}, structures.GetInfo_Params{}, structures.GetInfo_Result{}); err != nil {
			log.Fatalln(err)
		}

		if err := mr.RegisterMethod("gettxpool", GetTxPool_Handler{}, structures.GetTxPool_Params{}, structures.GetTxPool_Result{}); err != nil {
			log.Fatalln(err)
		}

	*/
	// create a new mux
	r.mux = http.NewServeMux()
	r.srv = &http.Server{Addr: r.address, Handler: r.mux}

	r.mux.HandleFunc("/", hello)
	r.mux.Handle("/json_rpc", r)
	/*
	   	// handle nasty http requests
	   	r.mux.HandleFunc("/getoutputs.bin", getoutputs) // stream any outputs to server, can make wallet work offline
	   	r.mux.HandleFunc("/gettransactions", gettransactions)
	           r.mux.HandleFunc("/sendrawtransaction", SendRawTransaction_Handler)
	*/
	//r.mux.HandleFunc("/json_rpc/debug", mr.ServeDebug)

	if err := r.srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("ERR listening to address err %s", err)
	}

}

func hello(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello world!")
}

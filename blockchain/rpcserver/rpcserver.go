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

package rpcserver

import "io"
import "fmt"
import "time"
import "sync"
import "sync/atomic"

//import "context"
import "net/http"

//import "github.com/intel-go/fastjson"
import "github.com/osamingo/jsonrpc"

import log "github.com/sirupsen/logrus"

import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/blockchain"

/* this file implements the rpcserver api, so as wallet and block explorer tools can work without migration */

// all components requiring access to blockchain must use , this struct to communicate
// this structure must be update while mutex
type RPCServer struct {
	Exit_Event chan bool // blockchain is shutting down and we must quit ASAP
	sync.RWMutex
}

var Exit_In_Progress bool
var chain *blockchain.Blockchain
var logger *log.Entry

func RPCServer_Start(params map[string]interface{}) (*RPCServer, error) {

	var err error
	var r RPCServer

	_ = err

	r.Exit_Event = make(chan bool)

	logger = globals.Logger.WithFields(log.Fields{"com": "RPC"}) // all components must use this logger
	chain = params["chain"].(*blockchain.Blockchain)

	// test whether chain is okay
	if chain.Get_Height() == 0 {
		return nil, fmt.Errorf("Chain DOES NOT have genesis block")
	}

	go r.Run()
	logger.Infof("RPC server started")
	atomic.AddUint32(&globals.Subsystem_Active, 1) // increment subsystem

	return &r, nil
}

// shutdown the rpc server component
func (r *RPCServer) RPCServer_Stop() {
	Exit_In_Progress = true
	close(r.Exit_Event) // send signal to all connections to exit
	// TODO we  must wait for connections to kill themselves
	time.Sleep(1 * time.Second)
	logger.Infof("RPC Shutdown")
	atomic.AddUint32(&globals.Subsystem_Active, ^uint32(0)) // this decrement 1 fom subsystem
}

func (r *RPCServer) Run() {

	mr := jsonrpc.NewMethodRepository()

	if err := mr.RegisterMethod("Main.Echo", EchoHandler{}, EchoParams{}, EchoResult{}); err != nil {
		log.Fatalln(err)
	}

	// install getblockcount handler
	if err := mr.RegisterMethod("getblockcount", GetBlockCount_Handler{}, GetBlockCount_Params{}, GetBlockCount_Result{}); err != nil {
		log.Fatalln(err)
	}

	// install on_getblockhash
	if err := mr.RegisterMethod("on_getblockhash", On_GetBlockHash_Handler{}, On_GetBlockHash_Params{}, On_GetBlockHash_Result{}); err != nil {
		log.Fatalln(err)
	}

	// install getblocktemplate handler
	if err := mr.RegisterMethod("getblocktemplate", GetBlockTemplate_Handler{}, GetBlockTemplate_Params{}, GetBlockTemplate_Result{}); err != nil {
		log.Fatalln(err)
	}

	// submitblock handler
	if err := mr.RegisterMethod("submitblock", SubmitBlock_Handler{}, SubmitBlock_Params{}, SubmitBlock_Result{}); err != nil {
		log.Fatalln(err)
	}

	if err := mr.RegisterMethod("getlastblockheader", GetLastBlockHeader_Handler{}, GetLastBlockHeader_Params{}, GetLastBlockHeader_Result{}); err != nil {
		log.Fatalln(err)
	}

	if err := mr.RegisterMethod("getblockheaderbyhash", GetBlockHeaderByHash_Handler{}, GetBlockHeaderByHash_Params{}, GetBlockHeaderByHash_Result{}); err != nil {
		log.Fatalln(err)
	}

	if err := mr.RegisterMethod("getblockheaderbyheight", GetBlockHeaderByHeight_Handler{}, GetBlockHeaderByHeight_Params{}, GetBlockHeaderByHeight_Result{}); err != nil {
		log.Fatalln(err)
	}
	if err := mr.RegisterMethod("getblock", GetBlock_Handler{}, GetBlock_Params{}, GetBlock_Result{}); err != nil {
		log.Fatalln(err)
	}

	if err := mr.RegisterMethod("get_info", GetInfo_Handler{}, GetInfo_Params{}, GetInfo_Result{}); err != nil {
		log.Fatalln(err)
	}

	if err := mr.RegisterMethod("gettxpool", GetTxPool_Handler{}, GetTxPool_Params{}, GetTxPool_Result{}); err != nil {
		log.Fatalln(err)
	}

	http.HandleFunc("/", hello)
	http.Handle("/json_rpc", mr)

	// handle nasty http requests
	http.HandleFunc("/getoutputs.bin", getoutputs) // stream any outputs to server, can make wallet work offline
	http.HandleFunc("/gettransactions", gettransactions)
	//http.HandleFunc("/json_rpc/debug", mr.ServeDebug)

	if err := http.ListenAndServe("127.0.0.1:9999", http.DefaultServeMux); err != nil {
		log.Fatalln(err)
	}

}

func hello(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello world!")
}

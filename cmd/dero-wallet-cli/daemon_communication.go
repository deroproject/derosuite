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

/* this file handles communication with the daemon
 * this includes receiving output information
 * *
 */
import "io"
import "fmt"
import "net"
import "time"
import "sync"
import "net/http"
import "compress/gzip"

//import "github.com/romana/rlog"
//import "github.com/pierrec/lz4"
import "github.com/ybbus/jsonrpc"

import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/blockchain/rpcserver"

var Wallet_Height uint64 // height of wallet
var Daemon_Height uint64 // height of daemon
var Connected bool = true

var rpcClient *jsonrpc.RPCClient
var netClient *http.Client
var endpoint string

var output_lock sync.Mutex

// this is as simple as it gets
// single threaded communication to get the daemon status and height
func Run_Communication_Engine() {

	// check if user specified  daemon address explicitly
	if globals.Arguments["--daemon-address"] != nil {
		daemon_address := globals.Arguments["--daemon-address"].(string)

		remote_end, err := net.ResolveTCPAddr("tcp", daemon_address)
		if err != nil {
			globals.Logger.Warnf("Daemon address \"%s\" is invalid. parse err %s", daemon_address, err)
		} else {

			fmt.Printf("%+v\n", remote_end.IP)
			if remote_end.IP == nil || remote_end.IP.IsUnspecified() { // user never provided an ipaddress, use loopback
				globals.Logger.Debugf("Setting loopback ip on daemon endpoint")
				remote_end.IP = net.IPv4(127, 0, 0, 1)
			}
			endpoint = remote_end.String()
		}

	}

	// if user provided endpoint has error, use default
	if endpoint == "" {
		endpoint = "127.0.0.1:9999"
		if !globals.IsMainnet() {
			endpoint = "127.0.0.1:28091"
		}
	}

	globals.Logger.Debugf("Daemon endpoint %s", endpoint)

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
	rpcClient = jsonrpc.NewRPCClient("http://" + endpoint + "/json_rpc")

	for {
		time.Sleep(1 * time.Second) // ping server every second
		// execute rpc to service
		response, err := rpcClient.Call("get_info")

		// notify user of any state change
		// if daemon connection breaks or comes live again
		if err == nil {
			if !Connected {
				globals.Logger.Infof("Connection to RPC server successful")
				Connected = true
			}
		} else {
			if Connected {
				globals.Logger.Warnf("Connection to RPC server Failed err %s", err)
				Connected = false
			}
			continue // try next time
		}
		var info rpcserver.GetInfo_Result
		err = response.GetObject(&info)
		if err != nil {
			globals.Logger.Warnf("Daemon getinfo RPC parsing error err: %s\n", err)
			continue
		}
		// detect whether both are in different modes
		//  daemon is in testnet and wallet in mainnet or
		// daemon
		if info.Testnet != !globals.IsMainnet() {
			globals.Logger.Warnf("Mainnet/TestNet  is different between wallet/daemon.Please run daemon/wallet without --testnet")
		}
		Daemon_Height = info.Height

	}

}

// get the outputs from the daemon, requesting specfic outputs
// the range can be anything
// if stop is zero,
// the daemon will flush out everything it has ASAP
// the stream can be saved and used later on
func Get_Outputs(start uint64, stop uint64) {

	output_lock.Lock()
	defer output_lock.Unlock()
	if Connected { // only execute query if we are connected

		response, err := http.Get(fmt.Sprintf("http://%s/getoutputs.bin?start=%d", endpoint, start))
		if err != nil {
			globals.Logger.Warnf("Error while requesting outputs from daemon err %s", err)
			// os.Exit(1)
		} else {
			defer response.Body.Close()

			// lz4reader := lz4.NewReader(response.Body)
			//   io.Copy(pipe_writer, lz4reader)

			gzipreader, err := gzip.NewReader(response.Body)
			if err != nil {
				globals.Logger.Warnf("Error while decompressing output from daemon  err: %s   ", err)
				return
			}
			defer gzipreader.Close()
			io.Copy(pipe_writer, gzipreader)

		}
		// contents, err := ioutil.ReadAll(response.Body)
	}

}

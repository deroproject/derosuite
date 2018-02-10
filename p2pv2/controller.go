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

package p2pv2

import "net"
import "time"
import "sync/atomic"

import log "github.com/sirupsen/logrus"

import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/blockchain"

var chain *blockchain.Blockchain // external reference to chain

var Exit_Event = make(chan bool) // causes all threads to exit
var Exit_In_Progress bool        // marks we are doing exit
var logger *log.Entry            // global logger, every logger in this package is a child of this

// Initialize P2P subsystem
func P2P_Init(params map[string]interface{}) error {
	logger = globals.Logger.WithFields(log.Fields{"com": "P2P"}) // all components must use this logger
	chain = params["chain"].(*blockchain.Blockchain)
	go P2P_engine()    // start outgoing engine
	go P2P_Server_v2() // start accepting connections
	logger.Infof("P2P started")
	atomic.AddUint32(&globals.Subsystem_Active, 1) // increment subsystem
	return nil
}

func P2P_engine() {

	// if user provided ips at command line , use them, currently we use only the first one

	var end_point_list []string

	/*
	   if _,ok := globals.Arguments["--add-exclusive-node"] ; ok { // check if parameter is supported
	       if globals.Arguments["--add-exclusive-node"] != nil {
	           tmp_list := globals.Arguments["--add-exclusive-node"].([]string)
	           for i := range tmp_list {
	               end_point_list = append(end_point_list,tmp_list[i])
	           }
	       }
	   }
	*/
	// add hard-coded seeds
	end_point_list = append(end_point_list, "127.0.0.1:18095")

	for {
		if Exit_In_Progress {
			return
		}
		//remote_addr := "localhost:18090"
		//remote_addr := "192.168.56.1:18090"
		//remote_addr := "76.74.170.128:18090"
		//remote_addr := "89.38.97.110:18090"
		remote_addr := end_point_list[0]

		remote_ip, err := net.ResolveTCPAddr("tcp", remote_addr)

		if err != nil {
			if Exit_In_Progress {
				return
			}
			logger.Debugf("Resolve address failed:", err.Error())
			time.Sleep(2 * time.Second)
			continue
		}

		// since we may be connecting through socks, grab the remote ip for our purpose rightnow
		conn, err := globals.Dialer.Dial("tcp", remote_ip.String())
		if err != nil {
			if Exit_In_Progress {
				return
			}
			logger.Debugf("Dial failed err %s", err.Error())
			time.Sleep(2 * time.Second)
			continue
		}

		logger.Debugf("Connection established to %s", remote_ip)
		Handle_Connection(conn, remote_ip, false) // handle  connection
		time.Sleep(4 * time.Second)
	}

}

func P2P_Server_v2() {

	// listen to incoming tcp connections
	l, err := net.Listen("tcp", "0.0.0.0:18095")
	if err != nil {
		logger.Fatalf("Could not listen on port 18095, errr %s", err)
	}
	defer l.Close()

	// p2p is shutting down, close the listening socket
	go func() {
		<-Exit_Event
		l.Close()
	}()

	// A common pattern is to start a loop to continously accept connections
	for {
		conn, err := l.Accept() //accept connections using Listener.Accept()
		if err != nil {
			if Exit_In_Progress { // break the loop, since we are exiting
				break
			}
			logger.Warnf("Err while accepting incoming connection errr %s", err)
			continue
		}
		raddr := conn.RemoteAddr().(*net.TCPAddr)
		go Handle_Connection(conn, raddr, true) // handle connection in a different go routine
	}

}

// shutdown the p2p component
func P2P_Shutdown() {
	Exit_In_Progress = true
	close(Exit_Event) // send signal to all connections to exit
	// TODO we  must wait for connections to kill themselves
	time.Sleep(1 * time.Second)
	logger.Infof("P2P Shutdown")
	atomic.AddUint32(&globals.Subsystem_Active, ^uint32(0)) // this decrement 1 fom subsystem

}

func Connection_ShutDown(connection *Connection) {
	connection.Conn.Close()

}

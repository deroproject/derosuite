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

package p2p

//import "os"
import "fmt"
import "net"
import "time"
import "sort"
import "strings"
import "math/big"
import "strconv"

//import "crypto/rsa"
import "crypto/ecdsa"
import "crypto/elliptic"

import "crypto/tls"
import "crypto/rand"
import "crypto/x509"
import "encoding/pem"
import "sync/atomic"
import "runtime/debug"

import "github.com/romana/rlog"
import log "github.com/sirupsen/logrus"

import "github.com/deroproject/derosuite/config"
import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/blockchain"
import "github.com/deroproject/derosuite/metrics"

var chain *blockchain.Blockchain // external reference to chain

var P2P_Port int // this will be exported while doing handshake

var Exit_Event = make(chan bool) // causes all threads to exit
var Exit_In_Progress bool        // marks we are doing exit
var logger *log.Entry            // global logger, every logger in this package is a child of this
var sync_node bool               // whether sync mode is activated

var nonbanlist []string // any ips in this list will never be banned
// the list will include seed nodes, any nodes provided at command prompt

// Initialize P2P subsystem
func P2P_Init(params map[string]interface{}) error {
	logger = globals.Logger.WithFields(log.Fields{"com": "P2P"}) // all components must use this logger

	GetPeerID() // Initialize peer id once

	// parse node tag if availble
	if _, ok := globals.Arguments["--node-tag"]; ok {
		if globals.Arguments["--node-tag"] != nil {

			node_tag = globals.Arguments["--node-tag"].(string)
		}
	}

	// permanently unban any seed nodes
	if globals.IsMainnet() {
		for i := range config.Mainnet_seed_nodes {
			nonbanlist = append(nonbanlist, strings.ToLower(config.Mainnet_seed_nodes[i]))
		}
	} else { // initial bootstrap
		for i := range config.Testnet_seed_nodes {
			nonbanlist = append(nonbanlist, strings.ToLower(config.Testnet_seed_nodes[i]))
		}
	}

	chain = params["chain"].(*blockchain.Blockchain)
	load_ban_list()  // load ban list
	load_peer_list() // load old list if availble

	// if user provided a sync node, connect with it
	if _, ok := globals.Arguments["--sync-node"]; ok { // check if parameter is supported
		if globals.Arguments["--sync-node"].(bool) {
			sync_node = true
			// disable p2p port
			globals.Arguments["--p2p-bind"] = ":0"

			// disable all connections except seed nodes
			globals.Arguments["--add-exclusive-node"] = []string{"0.0.0.0:0"}
			globals.Arguments["--add-priority-node"] = []string{"0.0.0.0:0"}

			go maintain_seed_node_connection()

			logger.Warnf("Sync mode is enabled. Please remove this option after chain syncs successfully")
		}
	}

	// register the metrics with the metrics registry
	metrics.Registry.MustRegister(block_propagation)
	metrics.Registry.MustRegister(transaction_propagation)

	go P2P_Server_v2()        // start accepting connections
	go P2P_engine()           // start outgoing engine
	go syncroniser()          // start sync engine
	go clean_up_propagation() // clean up propagation map
	logger.Infof("P2P started")
	atomic.AddUint32(&globals.Subsystem_Active, 1) // increment subsystem
	return nil
}

// TODO we need to make sure that exclusive/priority nodes are never banned
func P2P_engine() {

	var end_point_list []string
	if _, ok := globals.Arguments["--add-exclusive-node"]; ok { // check if parameter is supported
		if globals.Arguments["--add-exclusive-node"] != nil {
			tmp_list := globals.Arguments["--add-exclusive-node"].([]string)
			for i := range tmp_list {
				end_point_list = append(end_point_list, tmp_list[i])
				nonbanlist = append(nonbanlist, tmp_list[i])
			}
		}
	}

	// all prority nodes will be always connected
	if _, ok := globals.Arguments["--add-priority-node"]; ok { // check if parameter is supported
		if globals.Arguments["--add-priority-node"] != nil {
			tmp_list := globals.Arguments["--add-priority-node"].([]string)
			for i := range tmp_list {
				end_point_list = append(end_point_list, tmp_list[i])
				nonbanlist = append(nonbanlist, tmp_list[i])
			}
		}
	}

	{ // remove duplicates if any
		sort.Strings(end_point_list)
	start_again: // this list is expected to be less  than 100
		for i := range end_point_list {
			if i > 0 && end_point_list[i-1] == end_point_list[i] {
				end_point_list = append(end_point_list[:i-1], end_point_list[i:]...)
				goto start_again
			}
		}
	}

	logger.Debugf("Priority list %+v", end_point_list)

	// maintain connection to exclusive/priority nodes
	for i := range end_point_list {
		go maintain_outgoing_priority_connection(end_point_list[i], false)
	}

	// do not create connections to peers , if requested
	if _, ok := globals.Arguments["--add-exclusive-node"]; ok && len(globals.Arguments["--add-exclusive-node"].([]string)) == 0 { // check if parameter is supported
		go maintain_connection_to_peers()  // maintain certain number of  connections for peer to peers
		go maintain_seed_node_connection() // maintain connection with atleast 1 seed node

		// this code only triggers when we do not have peer list
		if find_peer_to_connect(1) == nil { // either we donot have a peer list or everyone is banned
			// trigger connection to all seed nodes hoping some will be up
			if globals.IsMainnet() { // initial boot strap should be quick
				for i := range config.Mainnet_seed_nodes {
					go connect_with_endpoint(config.Mainnet_seed_nodes[i], true)
				}
			} else { // initial bootstrap
				for i := range config.Testnet_seed_nodes {
					go connect_with_endpoint(config.Testnet_seed_nodes[i], true)
				}
			}

		}

	}

}

// will try to connect with given endpoint
// will block until the connection dies or is killed
func connect_with_endpoint(endpoint string, sync_node bool) {

	defer func() {
		if r := recover(); r != nil {
			rlog.Warnf("Recovered while handling connection, Stack trace below", r)
			rlog.Warnf("Stack trace  \n%s", debug.Stack())
		}
	}()

	remote_ip, err := net.ResolveTCPAddr("tcp", endpoint)
	if err != nil {
		rlog.Warnf("Resolve address failed:", err.Error())
		return
	}

	// check whether are already connected to this address if yes, return
	if IsAddressConnected(remote_ip.String()) {
		return
	}

	// since we may be connecting through socks, grab the remote ip for our purpose rightnow
	conn, err := globals.Dialer.Dial("tcp", remote_ip.String())

	//conn, err := tls.DialWithDialer(&globals.Dialer, "tcp", remote_ip.String(),&tls.Config{InsecureSkipVerify: true})
	//conn, err := tls.Dial("tcp", remote_ip.String(),&tls.Config{InsecureSkipVerify: true})
	if err != nil {
		rlog.Warnf("Dial failed err %s", err.Error())
		Peer_SetFail(remote_ip.String()) // update peer list as we see
		return
	}

	tcpc := conn.(*net.TCPConn)
	// detection time: tcp_keepalive_time + tcp_keepalive_probes + tcp_keepalive_intvl
	// default on linux:  30 + 8 * 30
	// default on osx:    30 + 8 * 75
	tcpc.SetKeepAlive(true)
	tcpc.SetKeepAlivePeriod(8 * time.Second)
	tcpc.SetLinger(0) // discard any pending data


	//conn.SetKeepAlive(true) // set keep alive true
	//conn.SetKeepAlivePeriod(10*time.Second) // keep alive every 10 secs

	// upgrade connection TO TLS ( tls.Dial does NOT support proxy)
	// TODO we need to choose fastest cipher here ( so both clients/servers are not loaded)
	conn = tls.Client(conn, &tls.Config{InsecureSkipVerify: true})

	// success is setup after handshake is done
	rlog.Debugf("Connection established to %s", remote_ip)
	Handle_Connection(conn, remote_ip, false, sync_node) // handle  connection
}

// maintains a persistant connection to endpoint
// if connection drops, tries again after 4 secs
func maintain_outgoing_priority_connection(endpoint string, sync_node bool) {
	for {
		select {
		case <-Exit_Event:
			return
		case <-time.After(4 * time.Second):
		}
		connect_with_endpoint(endpoint, sync_node)
	}
}

// this will maintain connection to 1 seed node randomly
func maintain_seed_node_connection() {
	for {

		select {
		case <-Exit_Event:
			return
		case <-time.After(2 * time.Second):
		}
		endpoint := ""
		if globals.IsMainnet() { // choose mainnet seed node
			r, _ := rand.Int(rand.Reader, big.NewInt(10240))
			endpoint = config.Mainnet_seed_nodes[r.Int64()%int64(len(config.Mainnet_seed_nodes))]
		} else { // choose testnet peer node
			r, _ := rand.Int(rand.Reader, big.NewInt(10240))
			endpoint = config.Testnet_seed_nodes[r.Int64()%int64(len(config.Testnet_seed_nodes))]
		}
		if endpoint != "" {
			//connect_with_endpoint(endpoint, sync_node)
			connect_with_endpoint(endpoint, true) // seed nodes always have sync mode
		}
	}
}

// keep building connections to network, we are talking outgoing connections
func maintain_connection_to_peers() {

	Min_Peers := int64(13) // we need to expose this to be modifieable at runtime without taking daemon offline
	// check how many connections are active
	if _, ok := globals.Arguments["--min-peers"]; ok && globals.Arguments["--min-peers"] != nil { // user specified a limit, use it if possible
		i, err := strconv.ParseInt(globals.Arguments["--min-peers"].(string), 10, 64)
		if err != nil {
			logger.Warnf("Error Parsing --max-peers err %s", err)
		} else {
			if i <= 1 {
				logger.Warnf("--min-peers should be positive and more than 1")
			} else {
				Min_Peers = i
			}
		}
		logger.Infof("Min outgoing peers limit %d", Min_Peers)
	}

	for {
		select {
		case <-Exit_Event:
			return
		case <-time.After(1000 * time.Millisecond):
		}

		// check number of connections, if limit is reached, trigger new connections if we have peers
		// if we have more do nothing
		_, out := Peer_Direction_Count()
		if out >= uint64(Min_Peers) { // we already have required number of peers, donot connect to more peers
			continue
		}

		peer := find_peer_to_connect(1)
		if peer != nil {
			go connect_with_endpoint(peer.Address, false)
		}
	}
}

func P2P_Server_v2() {

	default_address := "0.0.0.0:" + fmt.Sprintf("%d", config.Mainnet.P2P_Default_Port)
	P2P_Port = config.Mainnet.P2P_Default_Port
	if !globals.IsMainnet() {
		default_address = "0.0.0.0:" + fmt.Sprintf("%d", config.Testnet.P2P_Default_Port)
		P2P_Port = config.Testnet.P2P_Default_Port
	}

	if _, ok := globals.Arguments["--p2p-bind"]; ok && globals.Arguments["--p2p-bind"] != nil {
		addr, err := net.ResolveTCPAddr("tcp", globals.Arguments["--p2p-bind"].(string))
		if err != nil {
			logger.Warnf("--p2p-bind address is invalid, err = %s", err)
		} else {
			if addr.Port == 0 {
				logger.Infof("P2P server is disabled, No ports will be opened for P2P activity")
				return
			} else {
				default_address = addr.String()
				P2P_Port = addr.Port
			}
		}
	}

	logger.Infof("P2P  will listen on %s", default_address)
	tlsconfig := &tls.Config{Certificates: []tls.Certificate{generate_random_tls_cert()}}
	//l, err := tls.Listen("tcp", default_address, tlsconfig) // listen as TLS server

	// listen to incoming tcp connections tls style
	l, err := net.Listen("tcp", default_address)  // listen as simple TCP server
	if err != nil {
		logger.Fatalf("Could not listen on %s, err %s", default_address, err)
	}
	defer l.Close()

	// p2p is shutting down, close the listening socket
	go func() { <-Exit_Event; l.Close() }()

	// A common pattern is to start a loop to continously accept connections
	for {
		conn, err := l.Accept() //accept connections using Listener.Accept()
		if err != nil {
			select {
			case <-Exit_Event:
				return
			default:
			}
			logger.Warnf("Err while accepting incoming connection errr %s", err)
			continue
		}
		raddr := conn.RemoteAddr().(*net.TCPAddr)

		//if incoming IP is banned, disconnect now
		if IsAddressInBanList(raddr.IP.String()) {
			rlog.Tracef(1, "Incoming IP %s is banned, disconnecting now", raddr.IP.String())
			conn.Close()
		} else {

			tcpc := conn.(*net.TCPConn)
	// detection time: tcp_keepalive_time + tcp_keepalive_probes + tcp_keepalive_intvl
	// default on linux:  30 + 8 * 30
	// default on osx:    30 + 8 * 75
	tcpc.SetKeepAlive(true)
	tcpc.SetKeepAlivePeriod(8 * time.Second)
	tcpc.SetLinger(0) // discard any pending data

	tlsconn := tls.Server(conn,tlsconfig)
	go Handle_Connection(tlsconn, raddr, true, false) // handle connection in a different go routine
	

	//go Handle_Connection(conn, raddr, true, false) // handle connection in a different go routine
		}
	}
}

// shutdown the p2p component
func P2P_Shutdown() {
	close(Exit_Event) // send signal to all connections to exit
	save_peer_list()  // save peer list
	save_ban_list()   // save ban list

	// TODO we  must wait for connections to kill themselves
	time.Sleep(1 * time.Second)
	logger.Infof("P2P Shutdown")
	atomic.AddUint32(&globals.Subsystem_Active, ^uint32(0)) // this decrement 1 fom subsystem

}

// generate default tls cert to encrypt everything
// NOTE: this does NOT protect from individual active man-in-the-middle attacks
func generate_random_tls_cert() tls.Certificate {

	/* RSA can do only 500 exchange per second, we need to be faster
	     * reference https://github.com/golang/go/issues/20058
	    key, err := rsa.GenerateKey(rand.Reader, 512) // current using minimum size
	if err != nil {
	    log.Fatal("Private key cannot be created.", err.Error())
	}

	// Generate a pem block with the private key
	keyPem := pem.EncodeToMemory(&pem.Block{
	    Type:  "RSA PRIVATE KEY",
	    Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	*/
	// EC256 does roughly 20000 exchanges per second
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	b, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		log.Fatalf("Unable to marshal ECDSA private key: %v", err)
	}
	// Generate a pem block with the private key
	keyPem := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: b})

	tml := x509.Certificate{
		SerialNumber: big.NewInt(int64(GetPeerID()) ^ int64(time.Now().UnixNano())),

		// TODO do we need to add more parameters to make our certificate more authentic
		// and thwart traffic identification as a mass scale
		/*
		   // you can add any attr that you need
		   NotBefore:    time.Now(),
		   NotAfter:     time.Now().AddDate(5, 0, 0),
		   // you have to generate a different serial number each execution

		   Subject: pkix.Name{
		       CommonName:   "New Name",
		       Organization: []string{"New Org."},
		   },
		   BasicConstraintsValid: true,   // even basic constraints are not required
		*/
	}
	cert, err := x509.CreateCertificate(rand.Reader, &tml, &tml, &key.PublicKey, key)
	if err != nil {
		log.Fatal("Certificate cannot be created.", err.Error())
	}

	// Generate a pem block with the certificate
	certPem := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert})
	tlsCert, err := tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		log.Fatal("Cannot be loaded the certificate.", err.Error())
	}
	return tlsCert
}

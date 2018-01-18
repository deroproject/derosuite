package p2p


import "net"
import "time"
import "sync/atomic"

import log "github.com/sirupsen/logrus"

import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/blockchain"

var chain *blockchain.Blockchain // external reference to chain

var Exit_Event = make(chan bool) // causes all threads to exit
var Exit_In_Progress bool  // marks we are doing exit
var logger *log.Entry // global logger, every logger in this package is a child of this

// Initialize P2P subsystem
func P2P_Init(params map[string]interface{}) error {
	logger = globals.Logger.WithFields(log.Fields{"com": "P2P"}) // all components must use this logger
	chain = params["chain"].(*blockchain.Blockchain)
	go P2P_engine() // start outgoing engine
	//go P2P_Server_v1() // start accepting connections
        logger.Infof("P2P started")
        atomic.AddUint32(&globals.Subsystem_Active, 1) // increment subsystem
	return nil
}


func P2P_engine() {
	for {
		if Exit_In_Progress {
			return
		}
		//remote_addr := "localhost:18090"
		//remote_addr := "192.168.56.1:18090"                
		//remote_addr := "76.74.170.128:18090"
		remote_addr := "89.38.97.110:18090"

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

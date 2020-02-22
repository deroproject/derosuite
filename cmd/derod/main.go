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

import "io"
import "os"
import "time"
import "fmt"
import "bytes"
import "bufio"
import "strings"
import "strconv"
import "runtime"
import "math/big"
import "os/signal"

//import "crypto/sha1"
import "encoding/hex"
import "encoding/json"
import "path/filepath"
import "runtime/pprof"

import "golang.org/x/crypto/sha3"
import "github.com/romana/rlog"
import "github.com/chzyer/readline"
import "github.com/docopt/docopt-go"
import log "github.com/sirupsen/logrus"

//import "github.com/deroproject/derosuite/p2p"
import "github.com/deroproject/derosuite/p2p"
import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/block"
import "github.com/deroproject/derosuite/config"
//import "github.com/deroproject/derosuite/address"
import "github.com/deroproject/derosuite/blockchain"
import "github.com/deroproject/derosuite/transaction"

//import "github.com/deroproject/derosuite/checkpoints"
import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/cryptonight"

//import "github.com/deroproject/derosuite/crypto/ringct"
import "github.com/deroproject/derosuite/blockchain/rpcserver"
import "github.com/deroproject/derosuite/walletapi"

//import "github.com/deroproject/derosuite/address"

var command_line string = `derod 
DERO : A secure, private blockchain with smart-contracts

Usage:
  derod [--help] [--version] [--testnet] [--debug]  [--sync-node] [--boltdb | --badgerdb] [--disable-checkpoints] [--socks-proxy=<socks_ip:port>] [--data-dir=<directory>] [--p2p-bind=<0.0.0.0:18089>] [--add-exclusive-node=<ip:port>]... [--add-priority-node=<ip:port>]... 	 [--min-peers=<11>] [--rpc-bind=<127.0.0.1:9999>] [--lowcpuram] [--mining-address=<wallet_address>] [--mining-threads=<cpu_num>] [--node-tag=<unique name>]
  derod -h | --help
  derod --version

Options:
  -h --help     Show this screen.
  --version     Show version.
  --testnet  	Run in testnet mode.
  --debug       Debug mode enabled, print log messages
  --boltdb      Use boltdb as backend  (default on 64 bit systems)
  --badgerdb    Use Badgerdb as backend (default on 32 bit systems)
  --disable-checkpoints  Disable checkpoints, work in truly async, slow mode 1 block at a time
  --socks-proxy=<socks_ip:port>  Use a proxy to connect to network.
  --data-dir=<directory>    Store blockchain data at this location
  --rpc-bind=<127.0.0.1:9999>    RPC listens on this ip:port
  --p2p-bind=<0.0.0.0:18089>    p2p server listens on this ip:port, specify port 0 to disable listening server
  --add-exclusive-node=<ip:port>	Connect to specific peer only 
  --add-priority-node=<ip:port>	Maintain persistant connection to specified peer
  --sync-node       Sync node automatically with the seeds nodes. This option is for rare use.
  --min-peers=<11>      Number of connections the daemon tries to maintain  
  --lowcpuram          Disables some RAM consuming sections (deactivates mining/ultra compact protocol etc).
  --mining-address=<wallet_address>         This address is rewarded when a block is mined sucessfully
  --mining-threads=<cpu_num>         Number of CPU threads for mining
  --node-tag=<unique name>	Unique name of node, visible to everyone

  `

var Exit_In_Progress = make(chan bool)

func main() {
	var err error
	globals.Init_rlog()
	globals.Arguments, err = docopt.Parse(command_line, nil, true, config.Version.String(), false)

	if err != nil {
		log.Fatalf("Error while parsing options err: %s\n", err)
	}

	// We need to initialize readline first, so it changes stderr to ansi processor on windows

	l, err := readline.NewEx(&readline.Config{
		//Prompt:          "\033[92mDERO:\033[32m»\033[0m",
		Prompt:          "\033[92mDERO:\033[32m>>>\033[0m ",
		HistoryFile:     filepath.Join(os.TempDir(), "derod_readline.tmp"),
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		panic(err)
	}
	defer l.Close()

	// parse arguments and setup testnet mainnet
	globals.Initialize()     // setup network and proxy
	globals.Logger.Infof("") // a dummy write is required to fully activate logrus

	// all screen output must go through the readline
	globals.Logger.Out = l.Stdout()

	rlog.Infof("Arguments %+v", globals.Arguments)

	globals.Logger.Infof("DERO Atlantis daemon :  It is an alpha version, use it for testing/evaluations purpose only.")

	globals.Logger.Infof("Copyright 2017-2018 DERO Project. All rights reserved.")
	globals.Logger.Infof("OS:%s ARCH:%s GOMAXPROCS:%d", runtime.GOOS, runtime.GOARCH, runtime.GOMAXPROCS(0))
	globals.Logger.Infof("Version v%s", config.Version.String())

	globals.Logger.Infof("Daemon in %s mode", globals.Config.Name)
	globals.Logger.Infof("Daemon data directory %s", globals.GetDataDirectory())

	go check_update_loop ()

	params := map[string]interface{}{}

	//params["--disable-checkpoints"] = globals.Arguments["--disable-checkpoints"].(bool)
	chain, err := blockchain.Blockchain_Start(params)

	if err != nil {
		globals.Logger.Warnf("Error starting blockchain err '%s'",err)
		return
	}

	params["chain"] = chain

	// setup miner flag, before starting p2p
	//if globals.Arguments["--miner"].(bool) {
	//	params["--miner"] = true
	//}

	// since user is using a proxy, he definitely does not want to give out his IP
	/*if globals.Arguments["--socks-proxy"] != nil {
			globals.Arguments["--p2p-bind"] = ":0"
	                //globals.Logger.Warnf("Disabling P2P server since we are using socks proxy")
			if globals.Arguments["--enable-p2p-v1"].(bool) { // disable v1 p2p server
				globals.Arguments["--p2p-bind-v1"] = ":0"
			}
		}*/

	if cryptonight.HardwareAES {
		rlog.Infof("Hardware AES detected")
	}

	p2p.P2P_Init(params)

	//rpcserver.DEBUG_MODE = true
	rpc, _ := rpcserver.RPCServer_Start(params)

	// setup function pointers
	// these pointers need to fixed
	chain.Mempool.P2P_TX_Relayer = func(tx *transaction.Transaction, peerid uint64) (count int) {
		count += p2p.Broadcast_Tx(tx, peerid)
		return
	}

	chain.P2P_Block_Relayer = func(cbl *block.Complete_Block, peerid uint64) {
		p2p.Broadcast_Block(cbl, peerid)
	}

	if globals.Arguments["--lowcpuram"].(bool) == false && globals.Arguments["--sync-node"].(bool) == false { // enable v1 of protocol only if requested

		// if an address has been provided, verify that it satisfies //mainnet/testnet criteria
		if globals.Arguments["--mining-address"] != nil {

			addr, err := globals.ParseValidateAddress(globals.Arguments["--mining-address"].(string))
			if err != nil {
				globals.Logger.Fatalf("Mining address is invalid: err %s", err)
			}
			params["mining-address"] = addr

			//log.Debugf("Setting up proxy using %s", Arguments["--socks-proxy"].(string))
		}

		if globals.Arguments["--mining-threads"] != nil {
			thread_count := 0
			if s, err := strconv.Atoi(globals.Arguments["--mining-threads"].(string)); err == nil {
				//fmt.Printf("%T, %v", s, s)
				thread_count = s

			} else {
				globals.Logger.Fatalf("Mining threads argument cannot be parsed: err %s", err)
			}

			if thread_count > runtime.GOMAXPROCS(0) {
				globals.Logger.Fatalf("Mining threads (%d) is more than available CPUs (%d). This is NOT optimal", thread_count, runtime.GOMAXPROCS(0))

			}
			params["mining-threads"] = thread_count

			if _, ok := params["mining-address"]; !ok {
				globals.Logger.Fatalf("Mining threads require a valid wallet address")
			}

			globals.Logger.Infof("System will mine to %s with %d threads. Good Luck!!", globals.Arguments["--mining-address"].(string), thread_count)

	//		go start_miner(chain, params["mining-address"].(*address.Address), thread_count)
		}

	}

	go time_check_routine() // check whether server time is in sync

	// This tiny goroutine continuously updates status as required
	go func() {
		last_our_height := int64(0)
		last_best_height := int64(0)
		last_peer_count := uint64(0)
		last_topo_height := int64(0)
		last_mempool_tx_count := 0
		last_counter := uint64(0)
		last_counter_time := time.Now()
		last_mining_state := false
		for {
			select {
			case <-Exit_In_Progress:
				return
			default:
			}
			our_height := chain.Get_Height()
			best_height, best_topo_height := p2p.Best_Peer_Height()
			peer_count := p2p.Peer_Count()
			topo_height := chain.Load_TOPO_HEIGHT(nil)

			mempool_tx_count := len(chain.Mempool.Mempool_List_TX())

			// only update prompt if needed
			if last_mining_state != mining || mining || last_our_height != our_height || last_best_height != best_height || last_peer_count != peer_count || last_topo_height != topo_height || last_mempool_tx_count != mempool_tx_count {
				// choose color based on urgency
				color := "\033[32m" // default is green color
				if our_height < best_height {
					color = "\033[33m" // make prompt yellow
				} else if our_height > best_height {
					color = "\033[31m" // make prompt red
				}

				pcolor := "\033[32m" // default is green color
				if peer_count < 1 {
					pcolor = "\033[31m" // make prompt red
				} else if peer_count <= 8 {
					pcolor = "\033[33m" // make prompt yellow
				}

				mining_string := ""

				if mining {
					mining_speed := float64(counter-last_counter) / (float64(uint64(time.Since(last_counter_time))) / 1000000000.0)
					last_counter = counter
					last_counter_time = time.Now()
					switch {
					case mining_speed > 1000000:
						mining_string = fmt.Sprintf("MINING %.1f MH/s", float32(mining_speed)/1000000.0)
					case mining_speed > 1000:
						mining_string = fmt.Sprintf("MINING %.1f KH/s", float32(mining_speed)/1000.0)
					case mining_speed > 0:
						mining_string = fmt.Sprintf("MINING %.0f H/s", mining_speed)
					}
				}
				last_mining_state = mining

				hash_rate_string := ""
				hash_rate := chain.Get_Network_HashRate()
				switch {
				case hash_rate > 1000000000000:
					hash_rate_string = fmt.Sprintf("%.1f TH/s", float64(hash_rate)/1000000000000.0)
				case hash_rate > 1000000000:
					hash_rate_string = fmt.Sprintf("%.1f GH/s", float64(hash_rate)/1000000000.0)
				case hash_rate > 1000000:
					hash_rate_string = fmt.Sprintf("%.1f MH/s", float64(hash_rate)/1000000.0)
				case hash_rate > 1000:
					hash_rate_string = fmt.Sprintf("%.1f KH/s", float64(hash_rate)/1000.0)
				case hash_rate > 0:
					hash_rate_string = fmt.Sprintf("%d H/s", hash_rate)
				}

				testnet_string := ""
				if !globals.IsMainnet() {
					testnet_string = "\033[31m TESTNET"
				}

				l.SetPrompt(fmt.Sprintf("\033[1m\033[32mDERO: \033[0m"+color+"%d/%d [%d/%d] "+pcolor+"P %d TXp %d \033[32mNW %s %s>%s>>\033[0m ", our_height, topo_height, best_height, best_topo_height, peer_count, mempool_tx_count, hash_rate_string, mining_string, testnet_string))
				l.Refresh()
				last_our_height = our_height
				last_best_height = best_height
				last_peer_count = peer_count
				last_mempool_tx_count = mempool_tx_count
				last_topo_height = best_topo_height
			}
			time.Sleep(1 * time.Second)
		}
	}()

	setPasswordCfg := l.GenPasswordConfig()
	setPasswordCfg.SetListener(func(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
		l.SetPrompt(fmt.Sprintf("Enter password(%v): ", len(line)))
		l.Refresh()
		return nil, 0, false
	})
	l.Refresh() // refresh the prompt

         go func (){
            var gracefulStop = make(chan os.Signal)
            signal.Notify(gracefulStop,os.Interrupt) // listen to all signals
            for {
                sig := <-gracefulStop
                fmt.Printf("received signal %s\n", sig)
            
                if sig.String() == "interrupt" {
                    close(Exit_In_Progress)	
                }
            }
        }()

	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				fmt.Print("Ctrl-C received, Exit in progress\n")
				close(Exit_In_Progress)
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			<-Exit_In_Progress
		          break
		}

		line = strings.TrimSpace(line)
		line_parts := strings.Fields(line)

		command := ""
		if len(line_parts) >= 1 {
			command = strings.ToLower(line_parts[0])
		}

		switch {
		case line == "help":
			usage(l.Stderr())

		case strings.HasPrefix(line, "say"):
			line := strings.TrimSpace(line[3:])
			if len(line) == 0 {
				log.Println("say what?")
				break
			}
		//
		case command == "import_chain": // this migrates existing chain from DERO to DERO atlantis
			f, err := os.Open("/tmp/raw_export.txt")
			if err != nil {
				globals.Logger.Warnf("error opening  file  /tmp/raw_export.txt %s", err)
				continue
			}
			reader := bufio.NewReader(f)

			account, _ := walletapi.Generate_Keys_From_Random() // create a random address

			for {
				line, err = reader.ReadString('\n')

				if err != nil || len(line) < 10 {
					break
				}

				var txs []string

				err = json.Unmarshal([]byte(line), &txs)

				if err != nil {
					fmt.Printf("err while unmarshalling json err %s", err)
					continue
				}

				if len(txs) < 1 {
					panic("TX cannot be zero")
				}

				cbl, bl := chain.Create_new_miner_block(account.GetAddress())

				for i := range txs {

					var tx transaction.Transaction

					tx_bytes, err := hex.DecodeString(txs[i])
					if err != nil {
						globals.Logger.Warnf("TX could not be decoded")
					}

					err = tx.DeserializeHeader(tx_bytes)
					if err != nil {
						globals.Logger.Warnf("TX could not be Deserialized")
					}

					globals.Logger.Infof(" txhash  %s", tx.GetHash())
					if i == 0 {
						bl.Miner_TX = tx
						cbl.Bl.Miner_TX = tx

						if bl.Miner_TX.GetHash() != tx.GetHash() || cbl.Bl.Miner_TX.GetHash() != tx.GetHash() {
							panic("miner TX hash mismatch")
						}

					} else {
						bl.Tx_hashes = append(bl.Tx_hashes, tx.GetHash())
						cbl.Bl.Tx_hashes = append(cbl.Bl.Tx_hashes, tx.GetHash())
						cbl.Txs = append(cbl.Txs, &tx)
					}
				}

				if err, ok := chain.Add_Complete_Block(cbl); ok {
					globals.Logger.Warnf("Block Successfully accepted by chain at height %d", cbl.Bl.Miner_TX.Vin[0].(transaction.Txin_gen).Height)
				} else {
					globals.Logger.Warnf("Block rejected by chain at height %d, please investigate, err %s", cbl.Bl.Miner_TX.Vin[0].(transaction.Txin_gen).Height,err)
					globals.Logger.Warnf("Stopping import")

				}

			}

			globals.Logger.Infof("File imported Successfully")
			f.Close()

		case command == "profile": // writes cpu and memory profile
			// TODO enable profile over http rpc to enable better testing/tracking
			cpufile, err := os.Create(filepath.Join(globals.GetDataDirectory(), "cpuprofile.prof"))
			if err != nil {
				globals.Logger.Warnf("Could not start cpu profiling, err %s", err)
				continue
			}
			if err := pprof.StartCPUProfile(cpufile); err != nil {
				globals.Logger.Warnf("could not start CPU profile: %s ", err)
			}
			globals.Logger.Infof("CPU profiling will be available after program shutsdown")
			defer pprof.StopCPUProfile()

			/*
			        	memoryfile,err := os.Create(filepath.Join(globals.GetDataDirectory(), "memoryprofile.prof"))
						if err != nil{
							globals.Logger.Warnf("Could not start memory profiling, err %s", err)
							continue
						}
						if err := pprof.WriteHeapProfile(memoryfile); err != nil {
			            	globals.Logger.Warnf("could not start memory profile: ", err)
			        	}
			        	memoryfile.Close()
			*/

		case command == "print_bc":
			log.Info("printing block chain")
			// first is starting point, second is ending point
			start := int64(0)
			stop := int64(0)

			if len(line_parts) != 3 {
				log.Warnf("This function requires 2 parameters, start and endpoint\n")
				continue
			}
			if s, err := strconv.ParseInt(line_parts[1], 10, 64); err == nil {
				start = s
			} else {
				log.Warnf("Invalid start value err %s", err)
				continue
			}

			if s, err := strconv.ParseInt(line_parts[2], 10, 64); err == nil {
				stop = s
			} else {
				log.Warnf("Invalid stop value err %s", err)
				continue
			}

			if start < 0 || start > int64(chain.Load_TOPO_HEIGHT(nil)) {
				log.Warnf("Start value should be be between 0 and current height\n")
				continue
			}
			if start > stop || stop > int64(chain.Load_TOPO_HEIGHT(nil)) {
				log.Warnf("Stop value should be > start and current height\n")
				continue

			}

			log.Infof("Printing block chain from %d to %d\n", start, stop)

			for i := start; i <= stop; i++ {
				// get block id at height
				current_block_id, err := chain.Load_Block_Topological_order_at_index(nil, i)
				if err != nil {
					log.Infof("Skipping block at height %d due to error %s\n", i, err)
					continue
				}
				timestamp := chain.Load_Block_Timestamp(nil, current_block_id)

				cdiff := chain.Load_Block_Cumulative_Difficulty(nil, current_block_id)

				diff := chain.Load_Block_Difficulty(nil, current_block_id)
				//size := chain.

				log.Infof("topo height: %10d,  height %d, timestamp: %10d, difficulty: %s cdiff: %s\n", i, chain.Load_Height_for_BL_ID(nil, current_block_id), timestamp, diff.String(), cdiff.String())

				log.Infof("Block Id: %s , \n", current_block_id)
				log.Infof("")

			}
		case command == "mempool_print":
			chain.Mempool.Mempool_Print()

		case command == "mempool_flush":
			chain.Mempool.Mempool_flush()
		case command == "mempool_delete_tx":
			if len(line_parts) == 2 && len(line_parts[1]) == 64 {
				txid, err := hex.DecodeString(strings.ToLower(line_parts[1]))
				if err != nil {
					fmt.Printf("err while decoding txid err %s\n", err)
					continue
				}
				var hash crypto.Hash
				copy(hash[:32], []byte(txid))

				chain.Mempool.Mempool_Delete_TX(hash)
			} else {
				fmt.Printf("mempool_delete_tx  needs a single transaction id as arugument\n")
			}
		case command == "version":
			fmt.Printf("Version %s OS:%s ARCH:%s \n", config.Version.String(), runtime.GOOS, runtime.GOARCH)

		case command == "start_mining_legacy": // it needs 2 parameters, one dero address, second number of threads
			if mining {
				fmt.Printf("Mining is already started\n")
				continue
			}

			if globals.Arguments["--lowcpuram"].(bool) {
				globals.Logger.Warnf("Mining is deactivated since daemon is running in low cpu mode, please check program options.")
				continue
			}

			if globals.Arguments["--sync-node"].(bool) {
				globals.Logger.Warnf("Mining is deactivated since daemon is running with --sync-mode, please check program options.")
				continue
			}

			if len(line_parts) != 3 {
				fmt.Printf("This function requires 2 parameters 1) dero address  2) number of threads\n")
				continue
			}

			addr, err := globals.ParseValidateAddress(line_parts[1])
			if err != nil {
				globals.Logger.Warnf("Mining address is invalid: err %s", err)
				continue
			}

			thread_count := 0
			if s, err := strconv.Atoi(line_parts[2]); err == nil {
				//fmt.Printf("%T, %v", s, s)
				thread_count = s

			} else {
				globals.Logger.Warnf("Mining threads argument cannot be parsed: err %s", err)
				continue
			}

			if thread_count > runtime.GOMAXPROCS(0) {
				globals.Logger.Warnf("Mining threads (%d) is more than available CPUs (%d). This is NOT optimal", thread_count, runtime.GOMAXPROCS(0))

			}

			go start_miner(chain, addr, thread_count)
			fmt.Printf("Mining started for %s on %d threads", addr, thread_count)

		case command == "stop_mining":
			if mining == true {
				fmt.Printf("mining stopped\n")
			}
			mining = false

		case command == "print_tree": // prints entire block chain tree
			//WriteBlockChainTree(chain, "/tmp/graph.dot")

		case command == "print_block":
			fmt.Printf("printing block\n")
			if len(line_parts) == 2 && len(line_parts[1]) == 64 {
				bl_raw, err := hex.DecodeString(strings.ToLower(line_parts[1]))

				if err != nil {
					fmt.Printf("err while decoding txid err %s\n", err)
					continue
				}
				var hash crypto.Hash
				copy(hash[:32], []byte(bl_raw))

				bl, err := chain.Load_BL_FROM_ID(nil, hash)
				if err == nil {
					fmt.Printf("Block ID : %s\n", hash)
					fmt.Printf("Block : %x\n", bl.Serialize())
					fmt.Printf("difficulty: %s\n", chain.Load_Block_Difficulty(nil, hash).String())
					fmt.Printf("cdifficulty: %s\n", chain.Load_Block_Cumulative_Difficulty(nil, hash).String())
					fmt.Printf("PoW: %s\n", bl.GetPoWHash())
					//fmt.Printf("Orphan: %v\n",chain.Is_Block_Orphan(hash))

					json_bytes, err := json.Marshal(bl)

					fmt.Printf("%s  err : %s\n", string(prettyprint_json(json_bytes)), err)
				} else {
					fmt.Printf("Err %s\n", err)
				}
			} else if len(line_parts) == 2 {
				if s, err := strconv.ParseInt(line_parts[1], 10, 64); err == nil {
					_ = s
					// first load block id from topo height

					hash, err := chain.Load_Block_Topological_order_at_index(nil, s)
					if err != nil {
						fmt.Printf("Skipping block at topo height %d due to error %s\n", s, err)
						continue
					}
					bl, err := chain.Load_BL_FROM_ID(nil, hash)
					if err == nil {
						fmt.Printf("Block ID : %s\n", hash)
						fmt.Printf("Block : %x\n", bl.Serialize())
						fmt.Printf("difficulty: %s\n", chain.Load_Block_Difficulty(nil, hash).String())
						fmt.Printf("cdifficulty: %s\n", chain.Load_Block_Cumulative_Difficulty(nil, hash).String())
						fmt.Printf("Height: %d\n", chain.Load_Height_for_BL_ID(nil, hash))
						fmt.Printf("TopoHeight: %d\n", s)

						fmt.Printf("PoW: %s\n", bl.GetPoWHash())
						//fmt.Printf("Orphan: %v\n",chain.Is_Block_Orphan(hash))

						json_bytes, err := json.Marshal(bl)

						fmt.Printf("%s  err : %s\n", string(prettyprint_json(json_bytes)), err)
					} else {
						fmt.Printf("Err %s\n", err)
					}

				} else {
					fmt.Printf("print_block  needs a single transaction id as arugument\n")
				}
			}

		// can be used to debug/deserialize blocks
		// it can be used for blocks not in chain
		case command == "parse_block":
			if len(line_parts) != 2 {
				globals.Logger.Warnf("parse_block needs a block in hex format")
				continue
			}

			block_raw, err := hex.DecodeString(strings.ToLower(line_parts[1]))
			if err != nil {
				fmt.Printf("err while hex decoding block err %s\n", err)
				continue
			}

			var bl block.Block
			err = bl.Deserialize(block_raw)
			if err != nil {
				globals.Logger.Warnf("Error deserializing block err %s", err)
				continue
			}

			// decode and print block as much as possible
			fmt.Printf("Block ID : %s\n", bl.GetHash())
			fmt.Printf("PoW: %s\n", bl.GetPoWHash()) // block height
			fmt.Printf("Height: %d\n", bl.Miner_TX.Vin[0].(transaction.Txin_gen).Height)
			tips_found := true
			for i := range bl.Tips {
				_, err := chain.Load_BL_FROM_ID(nil, bl.Tips[i])
				if err != nil {
					fmt.Printf("Tips %s not in our DB", bl.Tips[i])
					tips_found = false
					break
				}
			}
			fmt.Printf("Tips: %d %+v\n", len(bl.Tips), bl.Tips)          // block height
			fmt.Printf("Txs: %d %+v\n", len(bl.Tx_hashes), bl.Tx_hashes) // block height
			expected_difficulty := new(big.Int).SetUint64(0)
			if tips_found { // we can solve diffculty
				expected_difficulty = chain.Get_Difficulty_At_Tips(nil, bl.Tips)
				fmt.Printf("Difficulty:  %s\n", expected_difficulty.String())

				powsuccess := chain.VerifyPoW(nil, &bl)
				fmt.Printf("PoW verification %+v\n", powsuccess)

				PoW := bl.GetPoWHash()
				for i := expected_difficulty.Uint64(); i >= 1; i-- {
					if blockchain.CheckPowHashBig(PoW, new(big.Int).SetUint64(i)) == true {
						fmt.Printf("Block actually has max Difficulty:  %d\n", i)
						break
					}
				}

			} else { // difficulty cann not solved

			}

		case command == "print_tx":

			if len(line_parts) == 2 && len(line_parts[1]) == 64 {
				txid, err := hex.DecodeString(strings.ToLower(line_parts[1]))

				if err != nil {
					fmt.Printf("err while decoding txid err %s\n", err)
					continue
				}
				var hash crypto.Hash
				copy(hash[:32], []byte(txid))

				tx, err := chain.Load_TX_FROM_ID(nil, hash)
				if err == nil {
					//s_bytes := tx.Serialize()
					//fmt.Printf("tx : %x\n", s_bytes)
					json_bytes, err := json.MarshalIndent(tx, "", "    ")
					_ = err
					fmt.Printf("%s\n", string(json_bytes))

					//tx.RctSignature.Message = ringct.Key(tx.GetPrefixHash())
					//ringct.Get_pre_mlsag_hash(tx.RctSignature)
					//chain.Expand_Transaction_v2(tx)

				} else {
					fmt.Printf("Err %s\n", err)
				}
			} else {
				fmt.Printf("print_tx  needs a single transaction id as arugument\n")
			}
		case command == "dev_verify_pool": // verifies and discards any tx which cannot be verified
			tx_list := chain.Mempool.Mempool_List_TX()
			for i := range tx_list { // check tx for nay double spend
				tx := chain.Mempool.Mempool_Get_TX(tx_list[i])
				if tx != nil {
					if !chain.Verify_Transaction_NonCoinbase_DoubleSpend_Check(nil, tx) {
						fmt.Printf("TX %s is double spended, this TX should not be in pool", tx_list[i])
						chain.Mempool.Mempool_Delete_TX(tx_list[i])
					}
				}
			}

		case strings.ToLower(line) == "diff":
			fmt.Printf("Network %s BH %d, Diff %d, NW Hashrate %0.03f MH/sec  TH %s\n", globals.Config.Name, chain.Get_Height(), chain.Get_Difficulty(), float64(chain.Get_Network_HashRate())/1000000.0, chain.Get_Top_ID())

		case strings.ToLower(line) == "status":
			// fmt.Printf("chain diff %d\n",chain.Get_Difficulty_At_Block(chain.Top_ID))
			//fmt.Printf("chain nw rate %d\n", chain.Get_Network_HashRate())
			inc, out := p2p.Peer_Direction_Count()

			supply := chain.Load_Already_Generated_Coins_for_Topo_Index(nil, chain.Load_TOPO_HEIGHT(nil))

			if supply > (1000000 * 1000000000000) {
				supply -= (1000000 * 1000000000000) // remove  premine
			}
			fmt.Printf("Network %s Height %d  NW Hashrate %0.03f MH/sec  TH %s Peers %d inc, %d out  MEMPOOL size %d  Total Supply %s DERO \n", globals.Config.Name, chain.Get_Height(), float64(chain.Get_Network_HashRate())/1000000.0, chain.Get_Top_ID(), inc, out, len(chain.Mempool.Mempool_List_TX()), globals.FormatMoney(supply))

			// print hardfork status on second line
			hf_state, _, _, threshold, version, votes, window := chain.Get_HF_info()
			switch hf_state {
			case 0: // voting in progress
				locked := false
				if window == 0 {
					window = 1
				}
				if votes >= (threshold*100)/window {
					locked = true
				}
				fmt.Printf("Hard-Fork v%d in-progress need %d/%d votes to lock in, votes: %d, LOCKED:%+v\n", version, ((threshold * 100) / window), window, votes, locked)
			case 1: // daemon is old and needs updation
				fmt.Printf("Please update this daemon to  support Hard-Fork v%d\n", version)
			case 2: // everything is okay
				fmt.Printf("Hard-Fork v%d\n", version)

			}
		case strings.ToLower(line) == "peer_list": // print peer list

			p2p.PeerList_Print()

		case strings.ToLower(line) == "sync_info": // print active connections

			p2p.Connection_Print()
		case strings.ToLower(line) == "bye":
			fallthrough
		case strings.ToLower(line) == "exit":
			fallthrough
		case strings.ToLower(line) == "quit":
			close(Exit_In_Progress)
			goto exit
		case strings.ToLower(line) == "graph":
			blockchain.WriteBlockChainTree(chain, "/tmp/graph.dot")

		case command == "pop":

			switch len(line_parts) {
			case 1:
				chain.Rewind_Chain(1)
			case 2:
				pop_count := 0
				if s, err := strconv.Atoi(line_parts[1]); err == nil {
					//fmt.Printf("%T, %v", s, s)
					pop_count = s

					if chain.Rewind_Chain(int(pop_count)) {
						globals.Logger.Infof("Rewind successful")
					} else {
						globals.Logger.Infof("Rewind failed")
					}

				} else {
					fmt.Printf("POP needs argument n to pop this many blocks from the top\n")
				}

			default:
				fmt.Printf("POP needs argument n to pop this many blocks from the top\n")
			}

		case command == "ban":
			if len(line_parts) >= 4 || len(line_parts) == 1 {
				fmt.Printf("IP address required to ban\n")
				break
			}

			if len(line_parts) == 3 { // process ban time if provided
				// if user provided a time, apply ban for specific time
				if s, err := strconv.ParseInt(line_parts[2], 10, 64); err == nil && s >= 0 {
					p2p.Ban_Address(line_parts[1], uint64(s))
					break
				} else {
					fmt.Printf("err parsing ban time (only positive number) %s", err)
					break
				}
			}

			err := p2p.Ban_Address(line_parts[1], 10*60) // default ban is 10 minutes
			if err != nil {
				fmt.Printf("err parsing address %s", err)
				break
			}

		case command == "unban":
			if len(line_parts) >= 3 || len(line_parts) == 1 {
				fmt.Printf("IP address required to unban\n")
				break
			}

			err := p2p.UnBan_Address(line_parts[1])
			if err != nil {
				fmt.Printf("err unbanning %s, err = %s", line_parts[1], err)
			} else {
				fmt.Printf("unbann %s successful", line_parts[1])
			}
		case command == "bans":
			p2p.BanList_Print() // print ban list

		case strings.ToLower(line) == "checkpoints": // save all knowns block id

			var block_id crypto.Hash
			checksums := "mainnet_checksums.dat"
			if !globals.IsMainnet() {
				checksums = "testnet_checksums.dat"
			}

			filename_checksums := filepath.Join(os.TempDir(), checksums)

			fchecksum, err := os.Create(filename_checksums)
			if err != nil {
				globals.Logger.Warnf("error creating new file %s", err)
				continue
			}

			wchecksums := bufio.NewWriter(fchecksum)

			chain.Lock() // we do not want any reorgs during this op
			height := chain.Load_TOPO_HEIGHT(nil)
			for i := int64(0); i <= height; i++ {

				block_id, err = chain.Load_Block_Topological_order_at_index(nil, i)
				if err != nil {
					break
				}

				// calculate sha1 of file
				h := sha3.New256()
				bl, err := chain.Load_BL_FROM_ID(nil, block_id)
				if err == nil {
					h.Write(bl.Serialize()) // write serialized block
				} else {
					break
				}
				for j := range bl.Tx_hashes {
					tx, err := chain.Load_TX_FROM_ID(nil, bl.Tx_hashes[j])
					if err == nil {
						h.Write(tx.Serialize()) // write serialized transaction
					} else {
						break
					}
				}
				if err != nil {
					break
				}

				wchecksums.Write(h.Sum(nil)) // write sha3 256 sum

			}
			if err != nil {
				globals.Logger.Warnf("error writing checkpoints err: %s", err)
			} else {
				globals.Logger.Infof("Successfully wrote %d checksums to file %s", height, filename_checksums)
			}

			wchecksums.Flush()
			fchecksum.Close()

			chain.Unlock()
		case line == "sleep":
			log.Println("sleep 4 second")
			time.Sleep(4 * time.Second)
		case line == "":
		default:
			log.Println("you said:", strconv.Quote(line))
		}
	}
exit:

	globals.Logger.Infof("Exit in Progress, Please wait")
	time.Sleep(100 * time.Millisecond) // give prompt update time to finish

	rpc.RPCServer_Stop()
	p2p.P2P_Shutdown() // shutdown p2p subsystem
	chain.Shutdown()   // shutdown chain subsysem

	for globals.Subsystem_Active > 0 {
		time.Sleep(100 * time.Millisecond)
	}
}

func prettyprint_json(b []byte) []byte {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	_ = err
	return out.Bytes()
}

func usage(w io.Writer) {
	io.WriteString(w, "commands:\n")
	//io.WriteString(w, completer.Tree("    "))
	io.WriteString(w, "\t\033[1mhelp\033[0m\t\tthis help\n")
	io.WriteString(w, "\t\033[1mdiff\033[0m\t\tShow difficulty\n")
	io.WriteString(w, "\t\033[1mprint_bc\033[0m\tPrint blockchain info in a given blocks range, print_bc <begin_height> <end_height>\n")
	io.WriteString(w, "\t\033[1mprint_block\033[0m\tPrint block, print_block <block_hash> or <block_height>\n")
	io.WriteString(w, "\t\033[1mprint_height\033[0m\tPrint local blockchain height\n")
	io.WriteString(w, "\t\033[1mprint_tx\033[0m\tPrint transaction, print_tx <transaction_hash>\n")
	io.WriteString(w, "\t\033[1mstatus\033[0m\t\tShow general information\n")
//	io.WriteString(w, "\t\033[1mstart_mining\033[0m\tStart mining <dero address> <number of threads>\n")
	io.WriteString(w, "\t\033[1mstop_mining\033[0m\tStop daemon mining\n")
	io.WriteString(w, "\t\033[1mpeer_list\033[0m\tPrint peer list\n")
	io.WriteString(w, "\t\033[1msync_info\033[0m\tPrint information about connected peers and their state\n")
	io.WriteString(w, "\t\033[1mbye\033[0m\t\tQuit the daemon\n")
	io.WriteString(w, "\t\033[1mban\033[0m\t\tBan specific ip from making any connections\n")
	io.WriteString(w, "\t\033[1munban\033[0m\t\tRevoke restrictions on previously banned ips\n")
	io.WriteString(w, "\t\033[1mbans\033[0m\t\tPrint current ban list\n")
	io.WriteString(w, "\t\033[1mversion\033[0m\t\tShow version\n")
	io.WriteString(w, "\t\033[1mexit\033[0m\t\tQuit the daemon\n")
	io.WriteString(w, "\t\033[1mquit\033[0m\t\tQuit the daemon\n")

}

var completer = readline.NewPrefixCompleter(
	/*	readline.PcItem("mode",
			readline.PcItem("vi"),
			readline.PcItem("emacs"),
		),
		readline.PcItem("login"),
		readline.PcItem("say",
			readline.PcItem("hello"),
			readline.PcItem("bye"),
		),
		readline.PcItem("setprompt"),
		readline.PcItem("setpassword"),
		readline.PcItem("bye"),
	*/
	readline.PcItem("help"),
	/*	readline.PcItem("go",
			readline.PcItem("build", readline.PcItem("-o"), readline.PcItem("-v")),
			readline.PcItem("install",
				readline.PcItem("-v"),
				readline.PcItem("-vv"),
				readline.PcItem("-vvv"),
			),
			readline.PcItem("test"),
		),
		readline.PcItem("sleep"),
	*/
	readline.PcItem("diff"),
//	readline.PcItem("dev_verify_pool"),
//	readline.PcItem("dev_verify_chain_doublespend"),
	readline.PcItem("mempool_flush"),
	readline.PcItem("mempool_delete_tx"),
	readline.PcItem("mempool_print"),
	readline.PcItem("peer_list"),
	readline.PcItem("print_bc"),
	readline.PcItem("print_block"),
	readline.PcItem("print_height"),
	readline.PcItem("print_tx"),
	readline.PcItem("status"),
//	readline.PcItem("start_mining"),
//	readline.PcItem("stop_mining"),
	readline.PcItem("sync_info"),
	readline.PcItem("version"),
	readline.PcItem("bye"),
	readline.PcItem("exit"),
	readline.PcItem("quit"),
)

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

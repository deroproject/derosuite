package main

import "io"
import "os"
import "fmt"
import "time"
import "crypto/rand"
import "sync"
import "runtime"
import "net"
import "net/http"
import "math/big"
import "path/filepath"
import "encoding/hex"
import "encoding/binary"
import "encoding/json"
import "os/signal"
import "sync/atomic"
import "strings"
import "strconv"

import "github.com/deroproject/derosuite/config"
import "github.com/deroproject/derosuite/globals"

import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/astrobwt"
import "github.com/deroproject/derosuite/structures"
import "github.com/deroproject/derosuite/cryptonight"

import log "github.com/sirupsen/logrus"
import "github.com/ybbus/jsonrpc"

import "github.com/romana/rlog"
import "github.com/chzyer/readline"
import "github.com/docopt/docopt-go"

var rpcClient *jsonrpc.RPCClient
var netClient *http.Client
var mutex sync.RWMutex
var job structures.GetBlockTemplate_Result
var pooljob jobreplydata
var poolID string
var m miner
var maxdelay int = 10000
var threads int
var iterations int = 100
var max_pow_size int = 819200 //astrobwt.MAX_LENGTH
var wallet_address string
var daemon_rpc_address string
var pool_address string
var pool_password string

var counter uint64
var hash_rate uint64
var Difficulty uint64
var our_height int64

var block_counter int

var command_line string = `dero-miner
DERO CPU Miner for AstroBWT.
ONE CPU, ONE VOTE.
http://wiki.dero.io

Usage:
  dero-miner --wallet-address=<wallet_address> [--daemon-rpc-address=<http://127.0.0.1:20206>] [--pool-address=<miners.dero.network:3333>] [--pool-password=<x>] [--mining-threads=<threads>] [--max-pow-size=1120] [--testnet] [--debug]
  dero-miner --bench [--max-pow-size=1120]
  dero-miner -h | --help
  dero-miner --version

Options:
  -h --help     Show this screen.
  --version     Show version.
  --bench  	    Run benchmark mode.
  --daemon-rpc-address=<http://127.0.0.1:20206>    Miner will connect to daemon RPC on this port.
  --pool-address=<miners.dero.network:3333>	Miner will connect to mining pool via this address:port
  --pool-password=<x> Defines a pool password. Some pools require password field to be defined to specify worker names or other unique attributes while connecting.
  --wallet-address=<wallet_address>    This address is rewarded when a block is mined sucessfully.
  --mining-threads=<threads>         Number of CPU threads for mining [default: ` + fmt.Sprintf("%d", runtime.GOMAXPROCS(0)) + `]
  --max-pow-size=1120          Max amount of PoW size in KiB to mine, some older/newer cpus can increase their work

Example Mainnet: ./dero-miner-linux-amd64 --wallet-address dERoXHjNHFBabzBCQbBDSqbkLURQyzmPRCLfeFtzRQA3NgVfU4HDbRpZQUKBzq59QU2QLcoAviYQ59FG4bu8T9pZ1woERqciSL --daemon-rpc-address=http://explorer.dero.io:20206 
Example Mainnet Pool Mining: ./dero-miner-linux-amd64 --wallet-address dERoXHjNHFBabzBCQbBDSqbkLURQyzmPRCLfeFtzRQA3NgVfU4HDbRpZQUKBzq59QU2QLcoAviYQ59FG4bu8T9pZ1woERqciSL --pool-address=http://miners.dero.network:3333
Example Testnet: ./dero-miner-linux-amd64 --wallet-address dEToYsDQtFoabzBCQbBDSqbkLURQyzmPRCLfeFtzRQA3NgVfU4HDbRpZQUKBzq59QU2QLcoAviYQ59FG4bu8T9pZ1woEQQstVq --daemon-rpc-address=http://127.0.0.1:30306 
Example Testnet Pool Mining: ./dero-miner-linux-amd64 --wallet-address dEToYsDQtFoabzBCQbBDSqbkLURQyzmPRCLfeFtzRQA3NgVfU4HDbRpZQUKBzq59QU2QLcoAviYQ59FG4bu8T9pZ1woEQQstVq --pool-address=http://miners.dero.network:3333
If daemon running on local machine no requirement of '--daemon-rpc-address' argument. 
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
		Prompt:          "\033[92mDERO Miner:\033[32m>>>\033[0m ",
		HistoryFile:     filepath.Join(os.TempDir(), "dero_miner_readline.tmp"),
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

	os.Setenv("RLOG_LOG_LEVEL", "INFO")
	rlog.UpdateEnv()

	rlog.Infof("Arguments %+v", globals.Arguments)

	globals.Logger.Infof("DERO Atlantis AstroBWT miner :  It is an alpha version, use it for testing/evaluations purpose only.")

	globals.Logger.Infof("Copyright 2017-2020 DERO Project. All rights reserved.")
	globals.Logger.Infof("OS:%s ARCH:%s GOMAXPROCS:%d", runtime.GOOS, runtime.GOARCH, runtime.GOMAXPROCS(0))
	globals.Logger.Infof("Version v%s", config.Version.String())

	if globals.Arguments["--wallet-address"] != nil {
		if globals.Arguments["--pool-address"] != nil {
			// Leave address parsing for the mining pool. This allows you to define extra args like @worker or .staticDiff or solo~ etc. within the wallet-address arg
			pool_address = globals.Arguments["--pool-address"].(string)
			wallet_address = globals.Arguments["--wallet-address"].(string)
		} else {
			addr, err := globals.ParseValidateAddress(globals.Arguments["--wallet-address"].(string))
			if err != nil {
				globals.Logger.Fatalf("Wallet address is invalid: err %s", err)
			}
			wallet_address = addr.String()
		}
	}

	if globals.Arguments["--pool-password"] != nil {
		pool_password = globals.Arguments["--pool-password"].(string)
	}

	if !globals.Arguments["--testnet"].(bool) {
		daemon_rpc_address = "http://127.0.0.1:20206"
	} else {
		daemon_rpc_address = "http://127.0.0.1:30306"
	}

	if globals.Arguments["--daemon-rpc-address"] != nil {
		daemon_rpc_address = globals.Arguments["--daemon-rpc-address"].(string)
	}

	threads = runtime.GOMAXPROCS(0)
	if globals.Arguments["--mining-threads"] != nil {
		if s, err := strconv.Atoi(globals.Arguments["--mining-threads"].(string)); err == nil {
			threads = s
		} else {
			globals.Logger.Fatalf("Mining threads argument cannot be parsed: err %s", err)
		}

		if threads > runtime.GOMAXPROCS(0) {
			globals.Logger.Fatalf("Mining threads (%d) is more than available CPUs (%d). This is NOT optimal", threads, runtime.GOMAXPROCS(0))
		}
	}
	if globals.Arguments["--max-pow-size"] != nil {
		if s, err := strconv.Atoi(globals.Arguments["--max-pow-size"].(string)); err == nil && s > 200 && s < 100000 {
			max_pow_size = s*1024
		} else {
			globals.Logger.Fatalf("max-pow-size argument cannot be parsed: err %s", err)
		}
	}
	globals.Logger.Infof("max-pow-size limited to %d bytes. Good Luck!!", max_pow_size)

	if globals.Arguments["--bench"].(bool) {

		var wg sync.WaitGroup

		fmt.Printf("%20s %20s %20s %20s %20s \n", "Threads", "Total Time", "Total Iterations", "Time/PoW ", "Hash Rate/Sec")
		iterations = 500
		for bench := 1; bench <= threads; bench++ {
			processor = 0
			now := time.Now()
			for i := 0; i < bench; i++ {
				wg.Add(1)
				go random_execution(&wg, iterations)
			}
			wg.Wait()
			duration := time.Now().Sub(now)

			fmt.Printf("%20s %20s %20s %20s %20s \n", fmt.Sprintf("%d", bench), fmt.Sprintf("%s", duration), fmt.Sprintf("%d", bench*iterations),
				fmt.Sprintf("%s", duration/time.Duration(bench*iterations)), fmt.Sprintf("%.1f", float32(time.Second)/(float32(duration/time.Duration(bench*iterations)))))

		}

		os.Exit(0)
	}

	globals.Logger.Infof("System will mine to \"%s\" with %d threads. Good Luck!!", wallet_address, threads)

	//threads_ptr := flag.Int("threads", runtime.NumCPU(), "No. Of threads")
	//iterations_ptr := flag.Int("iterations", 20, "No. Of DERO Stereo POW calculated/thread")
	/*bench_ptr := flag.Bool("bench", false, "run bench with params")
	daemon_ptr := flag.String("rpc-server-address", "127.0.0.1:18091", "DERO daemon RPC address to get work and submit mined blocks")
	delay_ptr := flag.Int("delay", 1, "Fetch job every this many seconds")
	wallet_address := flag.String("wallet-address", "", "Owner of this wallet will receive mining rewards")

	_ = daemon_ptr
	_ = delay_ptr
	_ = wallet_address
	*/

	if threads < 1 || iterations < 1 || threads > 2048 {
		globals.Logger.Fatalf("Invalid parameters\n")
		return
	}

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

		_ = last_mining_state
		_ = last_peer_count
		_ = last_topo_height
		_ = last_mempool_tx_count

		mining := true
		for {
			select {
			case <-Exit_In_Progress:
				return
			default:
			}

			best_height, best_topo_height := int64(0), int64(0)
			peer_count := uint64(0)

			mempool_tx_count := 0

			// only update prompt if needed
			if last_our_height != our_height || last_best_height != best_height || last_counter != counter {
				// choose color based on urgency
				color := "\033[33m" // default is green color
				/*if our_height < best_height {
					color = "\033[33m" // make prompt yellow
				} else if our_height > best_height {
					color = "\033[31m" // make prompt red
				}*/

				pcolor := "\033[32m" // default is green color
				/*if peer_count < 1 {
					pcolor = "\033[31m" // make prompt red
				} else if peer_count <= 8 {
					pcolor = "\033[33m" // make prompt yellow
				}*/

				mining_string := ""

				if mining {
					mining_speed := float64(counter-last_counter) / (float64(uint64(time.Since(last_counter_time))) / 1000000000.0)
					last_counter = counter
					last_counter_time = time.Now()
					switch {
					case mining_speed > 1000000:
						mining_string = fmt.Sprintf("MINING @ %.1f MH/s", float32(mining_speed)/1000000.0)
					case mining_speed > 1000:
						mining_string = fmt.Sprintf("MINING @ %.1f KH/s", float32(mining_speed)/1000.0)
					case mining_speed > 0:
						mining_string = fmt.Sprintf("MINING @ %.0f H/s", mining_speed)
					}
				}
				last_mining_state = mining

				hash_rate_string := ""

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

				// Remove network hashrate && block found counter for pool mining (handled by pool)
				if globals.Arguments["--pool-address"] != nil {
					l.SetPrompt(fmt.Sprintf("\033[1m\033[32mDERO Miner: \033[0m"+color+"Height %d "+pcolor+" \033[32m %s>%s>>\033[0m ", our_height, mining_string, testnet_string))
				} else {
					l.SetPrompt(fmt.Sprintf("\033[1m\033[32mDERO Miner: \033[0m"+color+"Height %d "+pcolor+" FOUND_BLOCKS %d \033[32mNW %s %s>%s>>\033[0m ", our_height, block_counter, hash_rate_string, mining_string, testnet_string))
				}
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

	l.Refresh() // refresh the prompt

	go func() {
		var gracefulStop = make(chan os.Signal)
		signal.Notify(gracefulStop, os.Interrupt) // listen to all signals
		for {
			sig := <-gracefulStop
			fmt.Printf("received signal %s\n", sig)

			if sig.String() == "interrupt" {
				close(Exit_In_Progress)
			}
		}
	}()

	if pool_address != "" {
		for i := 0; i < threads; i++ {
			go minepoolblock()
		}
		go receivework()
	} else {
		go increase_delay()
		for i := 0; i < threads; i++ {
			go mineblock()
		}
		go getwork()
	}

	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				fmt.Print("Ctrl-C received, Exit in progress\n")
				if m.conn != nil {
					globals.Logger.Infof("Closing pool connection")
					m.mu.Lock()
					m.conn.Close()
					m.mu.Unlock()
				}
				close(Exit_In_Progress)
				os.Exit(0)
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
		case command == "version":
			fmt.Printf("Version %s OS:%s ARCH:%s \n", config.Version.String(), runtime.GOOS, runtime.GOARCH)

		case strings.ToLower(line) == "bye":
			fallthrough
		case strings.ToLower(line) == "exit":
			fallthrough
		case strings.ToLower(line) == "quit":
			close(Exit_In_Progress)
			os.Exit(0)
		case line == "":
		default:
			log.Println("you said:", strconv.Quote(line))
		}
	}

	<-Exit_In_Progress

	return

}

func random_execution(wg *sync.WaitGroup, iterations int) {
	var workbuf [255]byte

	runtime.LockOSThread()
	//threadaffinity()

	for i := 0; i < iterations; i++ {
		rand.Read(workbuf[:])
		//astrobwt.POW(workbuf[:])
		//astrobwt.POW_0alloc(workbuf[:])
		_,success := astrobwt.POW_optimized_v1(workbuf[:], max_pow_size)
		if !success{
			i--
		}
	}
	wg.Done()
	runtime.UnlockOSThread()
}

func increase_delay() {
	for {
		time.Sleep(time.Second)
		maxdelay++
	}
}

// continuously get work
func getwork() {

	// create client
	rpcClient = jsonrpc.NewRPCClient(daemon_rpc_address + "/json_rpc")

	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}

	netClient = &http.Client{
		Timeout:   time.Second * 5,
		Transport: netTransport,
	}

	// execute rpc to service
	response, err := rpcClient.Call("get_info")
	if err == nil {
		globals.Logger.Infof("Connection to RPC server successful \"%s\"", daemon_rpc_address)
	} else {

		//log.Fatalf("Connection to RPC server Failed err %s", err)
		globals.Logger.Infof("Connection to RPC server \"%s\".Failed err %s", daemon_rpc_address, err)
		return
	}

	for {

		response, err = rpcClient.CallNamed("getblocktemplate", map[string]interface{}{"wallet_address": fmt.Sprintf("%s", wallet_address), "reserve_size": 10})
		if err == nil {
			var block_template structures.GetBlockTemplate_Result
			err = response.GetObject(&block_template)
			if err == nil {
				mutex.Lock()
				job = block_template
				maxdelay = 0
				mutex.Unlock()
				hash_rate = job.Difficulty / config.BLOCK_TIME_hf4
				our_height = int64(job.Height)
				Difficulty = job.Difficulty
				//fmt.Printf("block_template %+v\n", block_template)
			}

		} else {
			globals.Logger.Errorf("Error receiving block template  Failed err %s", err)
		}

		time.Sleep(300 * time.Millisecond)
	}

}

func mineblock() {
	var diff big.Int
	var powhash crypto.Hash
	var work [76]byte
	var extra_nonce [16]byte
	nonce_buf := work[39 : 39+4] //since slices are linked, it modifies parent
	runtime.LockOSThread()
	threadaffinity()

	iterations_per_loop := uint32(31.0 * float32(astrobwt.MAX_LENGTH) / float32(max_pow_size))

	var data astrobwt.Data
	for {
		mutex.RLock()
		myjob := job
		mutex.RUnlock()

		if maxdelay > 10 {
			time.Sleep(time.Second)
			continue
		}

		n, err := hex.Decode(work[:], []byte(myjob.Blockhashing_blob))
		if err != nil || n != 76 {
			time.Sleep(time.Second)
			globals.Logger.Errorf("Blockwork could not decoded successfully (%s) , err:%s n:%d %+v", myjob.Blockhashing_blob, err, n, myjob)
			continue
		}

		rand.Read(extra_nonce[:]) // fill extra nonce with random buffer
		copy(work[7+32+4:75], extra_nonce[:])

		diff.SetUint64(myjob.Difficulty)

		if work[0] <= 3 { // check major version
			for i := uint32(0); i < 20; i++ {
				atomic.AddUint64(&counter, 1)
				binary.BigEndian.PutUint32(nonce_buf, i)
				pow := cryptonight.SlowHash(work[:])
				copy(powhash[:], pow[:])

				if CheckPowHashBig(powhash, &diff) == true {
					globals.Logger.Infof("Successfully found DERO block at difficulty:%d", myjob.Difficulty)
					maxdelay = 200
					block_counter++

					response, err := rpcClient.Call("submitblock", myjob.Blocktemplate_blob, fmt.Sprintf("%x", work[:]))
					_ = response
					_ = err
					/*fmt.Printf("submitting %+v\n", []string{myjob.Blocktemplate_blob, fmt.Sprintf("%x", work[:])})
					fmt.Printf("submit err %s\n", err)
					fmt.Printf("submit response %s\n", response)
					*/
					break

				}
			}
		} else {
			for i := uint32(0); i < iterations_per_loop; i++ {
				binary.BigEndian.PutUint32(nonce_buf, i)
				//pow := astrobwt.POW_0alloc(work[:])
				pow, success := astrobwt.POW_optimized_v2(work[:],max_pow_size,&data)
				if !success {
					continue
				}
				atomic.AddUint64(&counter, 1)
				copy(powhash[:], pow[:])

				if CheckPowHashBig(powhash, &diff) == true {
					globals.Logger.Infof("Successfully found DERO block astroblock at difficulty:%d  at height %d", myjob.Difficulty, myjob.Height)
					maxdelay = 200

					block_counter++

					response, err := rpcClient.Call("submitblock", myjob.Blocktemplate_blob, fmt.Sprintf("%x", work[:]))
					_ = response
					_ = err
					/*fmt.Printf("submitting %+v\n", []string{myjob.Blocktemplate_blob, fmt.Sprintf("%x", work[:])})
					fmt.Printf("submit err %s\n", err)
					fmt.Printf("submit response %s\n", response)
					*/
					break

				}
			}
		}
	}
}

// nelbert442 - begin pool mining supporting structs/functions
type miner struct {
	mu        sync.Mutex
	conn      net.Conn
	dec       *json.Decoder
	enc       *json.Encoder
	rejectNum uint64
	acceptNum uint64
	jobDiff   uint64
}

type reply struct {
	JsonRPC    string      `json:"jsonrpc"`
	Id         int         `json:"id"`
	Method     string      `json:"method,omitempty"`
	Params     interface{} `json:"params,omitempty"`
	ReplyError interface{} `json:"error,omitempty"`
	Result     interface{} `json:"result,omitempty"`
}

type loginparams struct {
	Login string `json:"login"`
	Pass  string `json:"pass"`
	Agent string `json:"agent"`
}

type submitparams struct {
	Id     string `json:"id"`
	JobId  string `json:"job_id"`
	Nonce  string `json:"nonce"`
	Result string `json:"result"`
}

type jobreplydata struct {
	Blob   string `json:"blob"`
	JobId  string `json:"job_id"`
	Target string `json:"target"`
	Algo   string `json:"algo"`
	Height uint64 `json:"height"`
}

// generates uint64 difficulty from the targetHex sent from the pool w/ the job
func targettodiff(targetHex string) uint64 {
	var diff uint64
	strTarget := targetHex
	decoded, _ := hex.DecodeString(strTarget)
	if len(strTarget) < 16 {
		diff = uint64(binary.LittleEndian.Uint32(decoded))
		diff = 0xFFFFFFFF / diff
	} else {
		diff = binary.LittleEndian.Uint64(decoded)
	}

	return diff
}

func (m *miner) connectpool() {
	var err error
	var retryTimes = 0

Dial:
	globals.Logger.Infof("Dialing to pool: %v", pool_address)
	m.conn, err = net.Dial("tcp", pool_address)
	if err != nil {
		retryTimes++
		if retryTimes == 5 {
			retryTimes = 0
			globals.Logger.Fatalf("Failed to connect too many times, exiting...")
		}
		if err == io.EOF {
			globals.Logger.Fatalf("Failed to Login, check your username and password for pool")
		}
		globals.Logger.Errorf("retrying... %v", err)
		time.Sleep(10 * time.Second)
		goto Dial
	}

	// Message encoder (send) and decoder (receive) definitions
	m.enc = json.NewEncoder(m.conn)
	m.dec = json.NewDecoder(m.conn)

	// Send login
	login := loginparams{
		Login: wallet_address,
		Pass:  pool_password,
		Agent: "dero-miner",
	}

	loginRPC := reply{
		JsonRPC: "2.0",
		Id:      1,
		Method:  "login",
		Params:  login,
	}
	m.send(loginRPC)
}

// miner function to listen for pool jobs
func (m *miner) listen() {
	var retryTimes = 0

	for {
		var message reply
		err := m.dec.Decode(&message) // Wait for return data after a call of m.enc.Encode(...)
		if err != nil {
			if err == io.EOF {
				continue
			}

			if _, ok := err.(net.Error); ok {
				retryTimes++
				if retryTimes > 10 {
					retryTimes = 0
					globals.Logger.Errorf("Failed to connect too many times, exiting...")
				}
			}

			globals.Logger.Infof("err: %v", err)
			m.conn.Close()
			m.connectpool()
			time.Sleep(10 * time.Second)
		}

		if &message == nil {
			continue
		}

		go m.handleResponse(message)
	}
}

// func to handle received job / detail from pool
func (m *miner) handleResponse(message reply) {
	if message.ReplyError != nil {
		if err, ok := message.ReplyError.(map[string]interface{}); ok {
			if errMsg, ok := err["message"].(string); ok {
				m.rejectNum++

				if errMsg == "Invalid address used for login" {
					globals.Logger.Fatalf("errMsg: %v", errMsg)
				}
				globals.Logger.Errorf("Share rejected (%v/%v)", m.acceptNum, m.rejectNum)
				globals.Logger.Errorf("errMsg: %v", errMsg)
			}
		}
	}

	if message.Result != nil {
		results, ok := message.Result.(map[string]interface{})
		if ok && results["job"] != nil {

			// basic status check sent back by login
			if results["status"] != nil {
				if status, ok := results["status"].(string); ok && strings.ToLower(status) == "ok" {
					globals.Logger.Infof("Successful login")
				}
			}

			// receives miner ID from pool, sets poolID var for sending jobs back via submit
			if results["id"] != nil {
				globals.Logger.Infof("Setting miner ID to %v", results["id"].(string))
				poolID = results["id"].(string)
			}

			// Marshal down job from received message
			messageJobReply, _ := json.Marshal(results["job"])
			var data jobreplydata
			json.Unmarshal(messageJobReply, &data)

			// assign new job
			mutex.Lock()
			pooljob = data
			maxdelay = 0
			mutex.Unlock()

			// decode job target information to diff
			diff := targettodiff(data.Target)
			m.mu.Lock()
			m.jobDiff = diff
			m.mu.Unlock()

			globals.Logger.Infof("New job from %v at height %v on algo %v - diff %v", pool_address, data.Height, data.Algo, diff)

			our_height = int64(pooljob.Height)
		}

		if ok && results["status"] != nil && results["job"] == nil {
			if status, ok := results["status"].(string); ok && strings.ToLower(status) == "ok" {
				if strings.ToLower(status) == "ok" {
					m.acceptNum++
					globals.Logger.Infof("Share accepted at diff %v (%v/%v)", m.jobDiff, m.acceptNum, m.rejectNum)
					if extraMsg, ok := results["message"].(string); ok {
						if extraMsg != "" {
							globals.Logger.Infof("poolMsg: %v", extraMsg)
						}
					}
				}
			}
		}
	}

	if message.Params != nil && message.Method == "job" {
		// Marshal down job from received message
		messageJobReply, _ := json.Marshal(message.Params)
		var data jobreplydata
		json.Unmarshal(messageJobReply, &data)

		// assign new job
		mutex.Lock()
		pooljob = data
		maxdelay = 0
		mutex.Unlock()

		// decode job target information to diff
		diff := targettodiff(data.Target)
		m.mu.Lock()
		m.jobDiff = diff
		m.mu.Unlock()

		globals.Logger.Infof("New job from %v at height %v on algo %v - diff %v", pool_address, data.Height, data.Algo, diff)

		our_height = int64(pooljob.Height)
	}
}

// func to send information to pool, separate func for cleanliness
func (m *miner) send(message interface{}) {
	var retryTimes = 0

Enc:
	err := m.enc.Encode(message)
	if err != nil {
		globals.Logger.Errorf("err: %v", err)
		retryTimes++
		if retryTimes > 5 {
			retryTimes = 0
			globals.Logger.Fatalf("Failed to connect too many times, exiting...")
		}
		m.connectpool()
		goto Enc
	}
}

// receive work from pool
func receivework() {
	// Connect to mining pool and login
	m.connectpool()

	// Start listener for receiving messages
	m.listen()
}

// perform mining operations
func minepoolblock() {
	// Leaving comments in for changes from mineblock() for future needs if required.
	var diff big.Int
	var powhash crypto.Hash
	var work [76]byte
	nonce_buf := work[39 : 39+4] //since slices are linked, it modifies parent
	runtime.LockOSThread()
	threadaffinity()

	iterations_per_loop := uint32(31.0 * float32(astrobwt.MAX_LENGTH) / float32(max_pow_size))

	var data astrobwt.Data
	for {
		mutex.RLock()
		myjob := &pooljob
		mutex.RUnlock()

		if maxdelay > 10 {
			time.Sleep(time.Second)
			continue
		}

		n, err := hex.Decode(work[:], []byte(myjob.Blob))
		if err != nil || n != 76 {
			time.Sleep(time.Second)
			globals.Logger.Errorf("Blockwork could not decoded successfully (%s) , err:%s n:%d %+v", myjob.Blob, err, n, myjob)
			continue
		}

		// Do not need to rand read the extra nonce, it is provided by the pool within the job blob
		//rand.Read(extra_nonce[:]) // fill extra nonce with random buffer
		//copy(work[7+32+4:75], extra_nonce[:])

		if work[0] <= 3 { // check major version
			for i := uint32(0); i < 20; i++ {
				// Random read to nonce_buf, no need to inject uint32() val of i
				//binary.BigEndian.PutUint32(nonce_buf, i)

				diff.SetUint64(m.jobDiff)

				rand.Read(nonce_buf)
				pow := cryptonight.SlowHash(work[:])
				atomic.AddUint64(&counter, 1)
				copy(powhash[:], pow[:])

				if CheckPowHashBig(powhash, &diff) == true {
					result := hex.EncodeToString(powhash[:])

					// Send submit
					submit := submitparams{
						Id:     poolID,
						JobId:  myjob.JobId,
						Nonce:  hex.EncodeToString(work[39:43]),
						Result: result,
					}

					submitRPC := reply{
						JsonRPC: "2.0",
						Id:      1,
						Method:  "submit",
						Params:  submit,
					}

					m.send(submitRPC)
				}
			}
		} else {
			for i := uint32(0); i < iterations_per_loop; i++ {
				// Random read to nonce_buf, no need to inject uint32() val of i
				//binary.BigEndian.PutUint32(nonce_buf, i)

				diff.SetUint64(m.jobDiff)

				rand.Read(nonce_buf)
				pow, success := astrobwt.POW_optimized_v2(work[:], max_pow_size, &data)
				if !success {
					continue
				}
				atomic.AddUint64(&counter, 1)
				copy(powhash[:], pow[:])

				if CheckPowHashBig(powhash, &diff) == true {
					result := hex.EncodeToString(powhash[:])

					// Send submit
					submit := submitparams{
						Id:     poolID,
						JobId:  myjob.JobId,
						Nonce:  hex.EncodeToString(nonce_buf),
						Result: result,
					}

					submitRPC := reply{
						JsonRPC: "2.0",
						Id:      1,
						Method:  "submit",
						Params:  submit,
					}

					m.send(submitRPC)
				}
			}
		}
	}
}

func usage(w io.Writer) {
	io.WriteString(w, "commands:\n")
	//io.WriteString(w, completer.Tree("    "))
	io.WriteString(w, "\t\033[1mhelp\033[0m\t\tthis help\n")
	io.WriteString(w, "\t\033[1mstatus\033[0m\t\tShow general information\n")
	io.WriteString(w, "\t\033[1mbye\033[0m\t\tQuit the miner\n")
	io.WriteString(w, "\t\033[1mversion\033[0m\t\tShow version\n")
	io.WriteString(w, "\t\033[1mexit\033[0m\t\tQuit the miner\n")
	io.WriteString(w, "\t\033[1mquit\033[0m\t\tQuit the miner\n")

}

var completer = readline.NewPrefixCompleter(
	readline.PcItem("help"),
	readline.PcItem("status"),
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

package main

import "io"
import "os"
import "time"
import "fmt"
import "bytes"
import "strings"
import "strconv"
import "encoding/hex"
import "encoding/json"
import "path/filepath"

import "github.com/chzyer/readline"
import "github.com/docopt/docopt-go"
import log "github.com/sirupsen/logrus"

import "github.com/deroproject/derosuite/p2p"
import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/blockchain"
import "github.com/deroproject/derosuite/crypto"



var command_line string = `derod 
DERO : A secure, private blockchain with smart-contracts

Usage:
  derod [--help] [--version] [--testnet] [--debug] [--socks-proxy=<socks_ip:port>] [--p2p-bind-port=<18090>] [--add-exclusive-node=<ip:port>]...
  derod -h | --help
  derod --version

Options:
  -h --help     Show this screen.
  --version     Show version.
  --testnet  	Run in testnet mode.
  --debug       Debug mode enabled, print log messages
  --socks-proxy=<socks_ip:port>  Use a proxy to connect to network.
  --p2p-bind-port=<18090>    p2p server listens on this port.
  --add-exclusive-node=<ip:port>	Connect to this peer only (disabled for this version)`

func main() {
	var err error

	globals.Arguments, err = docopt.Parse(command_line, nil, true, "DERO daemon : work in progress", false)

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

	globals.Logger.Debugf("Arguments %+v", globals.Arguments)

	globals.Logger.Infof("DERO daemon :  This version is under heavy development, use it for testing/evaluations purpose only")
	globals.Logger.Infof("Daemon in %s mode", globals.Config.Name)

	chain, _ := blockchain.Blockchain_Start(nil)
	params := map[string]interface{}{}
	params["chain"] = chain
	p2p.P2P_Init(params)

	// This tiny goroutine continuously updates status as required
	go func() {
		last_our_height := uint64(0)
		last_best_height := uint64(0)
		last_peer_count := uint64(0)
		for {
			if globals.Exit_In_Progress {
				return
			}
			our_height := chain.Get_Height()
			best_height := p2p.Best_Peer_Height()
			peer_count := p2p.Peer_Count()

			// only update prompt if needed
			if last_our_height != our_height || last_best_height != best_height || last_peer_count != peer_count {
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
				l.SetPrompt(fmt.Sprintf("\033[1m\033[32mDERO: \033[0m"+color+"%d/%d "+pcolor+"P %d\033[32m>>>\033[0m ", our_height, best_height, peer_count))
				l.Refresh()
				last_our_height = our_height
				last_best_height = best_height
				last_peer_count = peer_count
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

	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {

				fmt.Print("Ctrl-C received, Exit in progress\n")
				globals.Exit_In_Progress = true
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		line_parts := strings.Fields(line)

		command := ""
		if len(line_parts) >= 1 {
			command = strings.ToLower(line_parts[0])
		}

		switch {
		case strings.HasPrefix(line, "mode "):
			switch line[5:] {
			case "vi":
				l.SetVimMode(true)
			case "emacs":
				l.SetVimMode(false)
			default:
				println("invalid mode:", line[5:])
			}
		case line == "mode":
			if l.IsVimMode() {
				println("current mode: vim")
			} else {
				println("current mode: emacs")
			}
		case line == "login":
			pswd, err := l.ReadPassword("please enter your password: ")
			if err != nil {
				break
			}
			println("you enter:", strconv.Quote(string(pswd)))
		case line == "help":
			usage(l.Stderr())
		case line == "setpassword":
			pswd, err := l.ReadPasswordWithConfig(setPasswordCfg)
			if err == nil {
				println("you set:", strconv.Quote(string(pswd)))
			}
		case strings.HasPrefix(line, "setprompt"):
			if len(line) <= 10 {
				log.Println("setprompt <prompt>")
				break
			}
			l.SetPrompt(line[10:])
		case strings.HasPrefix(line, "say"):
			line := strings.TrimSpace(line[3:])
			if len(line) == 0 {
				log.Println("say what?")
				break
			}
			go func() {
				for range time.Tick(time.Second) {
					log.Println(line)
				}
			}()
		case command == "print_bc":
			log.Info("printing block chain")
			// first is starting point, second is ending point
			start := int64(0)
			stop := int64(0)

			if len(line_parts) != 3 {
				fmt.Printf("This function requires 2 parameters, start and endpoint\n")
				continue
			}
			if s, err := strconv.ParseInt(line_parts[1], 10, 64); err == nil {
				start = s
			} else {
				fmt.Printf("Invalid start value")
				continue
			}

			if s, err := strconv.ParseInt(line_parts[2], 10, 64); err == nil {
				stop = s
			} else {
				fmt.Printf("Invalid stop value")
				continue
			}

			if start < 0 || start >= int64(chain.Get_Height()) {
				fmt.Printf("Start value should be be between 0 and current height\n")
				continue
			}
			if start > stop && stop >= int64(chain.Get_Height()) {
				fmt.Printf("Stop value should be > start and current height\n")
				continue

			}

			fmt.Printf("Printing block chain from %d to %d\n", start, stop)

			for i := start; i < stop; i++ {
				// get block id at height
				current_block_id, err := chain.Load_BL_ID_at_Height(uint64(i))
				if err != nil {
					fmt.Printf("Skipping block at height %d due to error %s\n", i, err)
					continue
				}
				timestamp := chain.Load_Block_Timestamp(current_block_id)
				parent_block_id := chain.Load_Block_Parent_ID(current_block_id)

				// calculate difficulty
				//parent_cdiff := chain.Load_Block_Cumulative_Difficulty(parent_block_id)

				//block_cdiff := chain.Load_Block_Cumulative_Difficulty(current_block_id)

				diff := chain.Get_Difficulty_At_Block(parent_block_id)
				//size := chain.

				fmt.Printf("height: %10d, timestamp: %10d, difficulty: %12d\n", i, timestamp, diff)

				fmt.Printf("Block Id: %x , prev block id:%x\n", current_block_id, parent_block_id)
				fmt.Printf("\n")

			}

		case command == "print_block":
			fmt.Printf("printing block\n")
			if len(line_parts) == 2 && len(line_parts[1]) == 64 {
				txid, err := hex.DecodeString(strings.ToLower(line_parts[1]))

				if err != nil {
					fmt.Printf("err while decoding txid err %s\n", err)
					continue
				}
				var hash crypto.Hash
				copy(hash[:32], []byte(txid))
				fmt.Printf("block id: %x\n", hash[:])

				bl, err := chain.Load_BL_FROM_ID(hash)
				if err == nil {
					fmt.Printf("Block : %x\n", bl.Serialize())
				} else {
					fmt.Printf("Err %s\n", err)
				}
			} else if len(line_parts) == 2 {
				if s, err := strconv.ParseInt(line_parts[1], 10, 64); err == nil {
					// first load block id from height
					hash, err := chain.Load_BL_ID_at_Height(uint64(s))
					if err == nil {
						bl, err := chain.Load_BL_FROM_ID(hash)
						if err == nil {
							fmt.Printf("block id: %x\n", hash[:])
							fmt.Printf("Block : %x\n", bl.Serialize())

							json_bytes, err := json.Marshal(bl)

							fmt.Printf("%s  err : %s\n", string(prettyprint_json(json_bytes)), err)
						} else {
							fmt.Printf("Err %s\n", err)
						}

					} else {
						fmt.Printf("err %s\n", err)
					}
				}
			} else {
				fmt.Printf("print_tx  needs a single transaction id as arugument\n")
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

				tx, err := chain.Load_TX_FROM_ID(hash)
				if err == nil {
					s_bytes := tx.Serialize()
					fmt.Printf("tx : %x\n", s_bytes)
					json_bytes, err := json.MarshalIndent(tx, "", "    ")
					_ = err
					fmt.Printf("%s\n", string(json_bytes))

				} else {
					fmt.Printf("Err %s\n", err)
				}
			} else {
				fmt.Printf("print_tx  needs a single transaction id as arugument\n")
			}
		case strings.ToLower(line) == "diff":
			fmt.Printf("Network %s BH %d, Diff %d, NW Hashrate %0.03f MH/sec  TH %x\n", globals.Config.Name, chain.Get_Height(), chain.Get_Difficulty(), float64(chain.Get_Network_HashRate())/1000000.0, chain.Get_Top_ID())

		case strings.ToLower(line) == "status":
			// fmt.Printf("chain diff %d\n",chain.Get_Difficulty_At_Block(chain.Top_ID))
			//fmt.Printf("chain nw rate %d\n", chain.Get_Network_HashRate())
                        inc, out := p2p.Peer_Direction_Count()
			fmt.Printf("Network %s Height %d  NW Hashrate %0.03f MH/sec  TH %x Peers %d INC, %d OUT\n", globals.Config.Name, chain.Get_Height(), float64(chain.Get_Network_HashRate())/1000000.0, chain.Get_Top_ID(), inc,out )
		case strings.ToLower(line) == "sync_info":
			p2p.Connection_Print()
		case strings.ToLower(line) == "bye":
			fallthrough
		case strings.ToLower(line) == "exit":
			fallthrough
		case strings.ToLower(line) == "quit":
			goto exit
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
        io.WriteString(w,"\t\033[1mhelp\033[0m\t\tthis help\n")
        io.WriteString(w,"\t\033[1mdiff\033[0m\t\tShow difficulty\n")
        io.WriteString(w,"\t\033[1mprint_bc\033[0m\tPrint blockchain info in a given blocks range, print_bc <begin_height> <end_height>\n")
        io.WriteString(w,"\t\033[1mprint_block\033[0m\tPrint block, print_block <block_hash> or <block_height>\n")
        io.WriteString(w,"\t\033[1mprint_height\033[0m\tPrint local blockchain height\n")
        io.WriteString(w,"\t\033[1mprint_tx\033[0m\tPrint transaction, print_tx <transaction_hash>\n")
        io.WriteString(w,"\t\033[1mstatus\033[0m\t\tShow genereal information\n")
        io.WriteString(w,"\t\033[1msync_info\033[0m\tPrint information about connected peers and their state\n")
        io.WriteString(w,"\t\033[1mbye\033[0m\t\tQuit the daemon\n")
        io.WriteString(w,"\t\033[1mexit\033[0m\t\tQuit the daemon\n")
        io.WriteString(w,"\t\033[1mquit\033[0m\t\tQuit the daemon\n")
        
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
	readline.PcItem("print_bc"),
	readline.PcItem("print_block"),
	readline.PcItem("print_height"),
	readline.PcItem("print_tx"),
        readline.PcItem("status"),
	readline.PcItem("sync_info"),
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

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

/// this file implements the wallet and rpc wallet

import "io"
import "os"
import "fmt"
import "time"
import "sync"
import "strings"
import "strconv"
import "runtime"

//import "io/ioutil"
//import "bufio"
//import "bytes"
//import "net/http"
import "encoding/hex"

import "github.com/romana/rlog"
import "github.com/chzyer/readline"
import "github.com/docopt/docopt-go"
import log "github.com/sirupsen/logrus"
import "github.com/vmihailenco/msgpack"

//import "github.com/deroproject/derosuite/address"
import "github.com/deroproject/derosuite/walletapi"
import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/walletapi/mnemonics"

var command_line string = `dero-wallet-cli 
DERO : A secure, private blockchain with smart-contracts

Usage:
  derod [--help] [--version] [--offline] [--offline_datafile=<file>] [--testnet] [--prompt] [--debug] [--daemon-address=<host:port>] [--restore-deterministic-wallet] [--electrum-seed=<recovery-seed>] [--socks-proxy=<socks_ip:port>]  
  derod -h | --help
  derod --version

Options:
  -h --help     Show this screen.
  --version     Show version.
  --offline     Run the wallet in completely offline mode 
  --offline_datafile=<file>  Use the data in offline mode default ("getoutputs.bin") in current dir
  --prompt      Disable menu and display prompt
  --testnet  	Run in testnet mode.
  --debug       Debug mode enabled, print log messages
  --restore-deterministic-wallet    Restore wallet from previously saved recovery seed
  --electrum-seed=<recovery-seed>   Seed to use while restoring wallet
  --password     Password to unlock the wallet
  --socks-proxy=<socks_ip:port>  Use a proxy to connect to Daemon.
  --daemon-address=<host:port>    Use daemon instance at <host>:<port>

  `
var menu_mode bool = true                             // default display menu mode
var account_valid bool = false                        // if an account has been opened, do not allow to create new account in this session
var offline_mode bool                                 // whether we are in offline mode
var sync_in_progress int                              //  whether sync is in progress with daemon
var account *walletapi.Account = &walletapi.Account{} // all account  data is available here
var address string
var sync_time time.Time // used to suitable update  prompt

var default_offline_datafile string = "getoutputs.bin"

// these pipes are used to feed in transaction data to recover valid amounts
var pipe_reader *io.PipeReader // any output will be read from this end point
var pipe_writer *io.PipeWriter // any data from daemon or file needs to written here

var color_black = "\033[30m"
var color_red = "\033[31m"
var color_green = "\033[32m"
var color_yellow = "\033[33m"
var color_blue = "\033[34m"
var color_magenta = "\033[35m"
var color_cyan = "\033[36m"
var color_white = "\033[37m"

var prompt_mutex sync.Mutex // prompt lock
var prompt string = "\033[92mDERO Wallet:\033[32m>>>\033[0m "

func main() {

	var err error
	globals.Arguments, err = docopt.Parse(command_line, nil, true, "DERO daemon : work in progress", false)
	if err != nil {
		log.Fatalf("Error while parsing options err: %s\n", err)
	}

	// We need to initialize readline first, so it changes stderr to ansi processor on windows
	l, err := readline.NewEx(&readline.Config{
		//Prompt:          "\033[92mDERO:\033[32m»\033[0m",
		Prompt:          prompt,
		HistoryFile:     "", // wallet never saves any history file anywhere, to prevent any leakage
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
	globals.Logger.Infof("DERO Wallet :  This version is under heavy development, use it for testing/evaluations purpose only")
	globals.Logger.Infof("Copyright 2017-2018 DERO Project. All rights reserved.")
	globals.Logger.Infof("OS:%s ARCH:%s GOMAXPROCS:%d", runtime.GOOS, runtime.GOARCH, runtime.GOMAXPROCS(0))
	globals.Logger.Infof("Wallet in %s mode", globals.Config.Name)

	// disable menu mode if requested
	if globals.Arguments["--prompt"].(bool) {
		menu_mode = false
	}
	// lets handle the arguments one by one
	if globals.Arguments["--restore-deterministic-wallet"].(bool) {
		// user wants to recover wallet, check whether seed is provided on command line, if not prompt now
		seed := ""

		if globals.Arguments["--electrum-seed"] != nil {
			seed = globals.Arguments["--electrum-seed"].(string)
		} else { // prompt user for seed
			seed = read_line_with_prompt(l, "Enter your seed (25 words) : ")
		}

		account, err = walletapi.Generate_Account_From_Recovery_Words(seed)
		if err != nil {
			globals.Logger.Warnf("Error while recovering seed err %s\n", err)
			return
		}

		account_valid = true
		globals.Logger.Debugf("Seed Language %s", account.SeedLanguage)
		globals.Logger.Infof("Successfully recovered wallet from seed")
		address = account.GetAddress().String()
	}

	// check if offline mode requested
	if globals.Arguments["--offline"].(bool) == true {
		offline_mode = true
	} else { // we are not in offline mode, start communications with the daemon
		go Run_Communication_Engine()
	}

	pipe_reader, pipe_writer = io.Pipe() // create pipes

	setPasswordCfg := l.GenPasswordConfig()
	setPasswordCfg.SetListener(func(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
		l.SetPrompt(fmt.Sprintf("Enter password(%v): ", len(line)))
		l.Refresh()
		return nil, 0, false
	})
	l.Refresh() // refresh the prompt

	// reader ready to parse any data from the file
	go blockchain_data_consumer()

	// update prompt when required
	go update_prompt(l)

	// if wallet has been opened in offline mode by commands supplied at command prompt
	// trigger the offline scan
	if account_valid {
		go trigger_offline_data_scan()
	}

	// start infinite loop processing user commands
	for {

		if globals.Exit_In_Progress { // exit if requested so
			break
		}
		if menu_mode { // display menu if requested
			if account_valid { // account is opened, display post menu
				display_easymenu_post_open_command(l)
			} else { // account has not been opened display pre open menu
				display_easymenu_pre_open_command(l)
			}
		}

		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				globals.Logger.Infof("Ctrl-C received, Exit in progress\n")
				globals.Exit_In_Progress = true
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		// pass command to suitable handler
		if menu_mode {
			if account_valid {
				handle_easymenu_post_open_command(l, line)
			} else {
				handle_easymenu_pre_open_command(l, line)
			}
		} else {
			handle_prompt_command(l, line)
		}

	}
	globals.Exit_In_Progress = true

}

// this functions reads all data transferred from daemon or from offline file
// and plays it here
// finds which transactions belong to current account
// and adds them to account for reconciliation
func blockchain_data_consumer() {
	var err error
	for {
		rlog.Tracef(1, "Discarding old pipe_reader,writer, creating new")
		// close already created pipes, discarding there data
		pipe_reader.Close()
		pipe_writer.Close()
		pipe_reader, pipe_writer = io.Pipe() // create pipes

		decoder := msgpack.NewDecoder(pipe_reader)
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

			if globals.Exit_In_Progress {
				return
			}

			sync_time = time.Now()
			if account.Index_Global < output.Index_Global { // process tx if it has not been processed earlier

				Wallet_Height = output.Height
				account.Height = output.Height
				if account.Is_Output_Ours(output.Tx_Public_Key, output.Index_within_tx, crypto.Key(output.InKey.Destination)) {
					amount, _, result := account.Decode_RingCT_Output(output.Tx_Public_Key,
						output.Index_within_tx,
						crypto.Key(output.InKey.Mask),
						output.ECDHTuple,
						output.SigType)

					if result == false {
						globals.Logger.Warnf("Internal error occurred, amount cannot be spent")
					}
					globals.Logger.Infof(color_green+"Height %d transaction %s received %s DERO"+color_white, output.Height, output.TXID, globals.FormatMoney(amount))

					// add tx to wallet
					account.Add_Transaction_Record_Funds(&output)

				}

				// check this keyimage represents our funds
				// if yes we have consumed that specific funds, mark them as such
				amount_spent := uint64(0)
				for i := range output.Key_Images {
					amount_per_keyimage, result := account.Is_Our_Fund_Consumed(output.Key_Images[i])
					if result {
						amount_spent += amount_per_keyimage
						account.Consume_Transaction_Record_Funds(&output, output.Key_Images[i]) // decrease fund from our wallet
					}
				}

				if amount_spent > 0 {
					globals.Logger.Infof(color_magenta+"Height %d transaction %s Spent %s DERO"+color_white, output.Height, output.TXID, globals.FormatMoney(amount_spent))
				}

				account.Index_Global = output.Index_Global
			}

		}
	}

}

// update prompt as and when necessary
// TODO: make this code simple, with clear direction
func update_prompt(l *readline.Instance) {

	last_wallet_height := uint64(0)
	last_daemon_height := uint64(0)

	for {
		time.Sleep(30 * time.Millisecond) // give user a smooth running number

		if globals.Exit_In_Progress {
			return
		}
		prompt_mutex.Lock() // do not update if we can not lock the mutex

		// show first 8 bytes of address
		address_trim := ""
		if len(address) > 8 {
			address_trim = address[0:8]
		} else {
			address_trim = "DERO Wallet"
		}

		if len(address) == 0 {
			last_wallet_height = 0
			Wallet_Height = 0
		}

		if !account_valid {
			l.SetPrompt(fmt.Sprintf("\033[1m\033[32m%s \033[0m"+color_green+"0/%d \033[32m>>>\033[0m ", address_trim, Daemon_Height))
			prompt_mutex.Unlock()
			continue
		}

		// only update prompt if needed, trigger resync if required
		if last_wallet_height != Wallet_Height || last_daemon_height != Daemon_Height {
			// choose color based on urgency
			color := "\033[32m" // default is green color
			if Wallet_Height < Daemon_Height {
				color = "\033[33m" // make prompt yellow
			}

			balance_string := ""

			if account_valid {
				balance_unlocked, locked_balance := account.Get_Balance()
				balance_string = fmt.Sprintf(color_green+"%s "+color_white+"| "+color_yellow+"%s", globals.FormatMoney(balance_unlocked), globals.FormatMoney(locked_balance))
			}

			l.SetPrompt(fmt.Sprintf("\033[1m\033[32m%s \033[0m"+color+"%d/%d %s\033[32m>>>\033[0m ", address_trim, Wallet_Height, Daemon_Height, balance_string))
			l.Refresh()
			last_wallet_height = Wallet_Height
			last_daemon_height = Daemon_Height

		}

		if time.Since(sync_time) > (2*time.Second) && Wallet_Height < Daemon_Height {
			if !offline_mode { // if offline mode, never connect anywhere
				go Get_Outputs(account.Index_Global, 0) // start sync
			}

			sync_time = time.Now()
		}

		prompt_mutex.Unlock()

	}

}

// create a new wallet from scratch from random numbers
func Create_New_Account(l *readline.Instance) *walletapi.Account {

	account, _ := walletapi.Generate_Keys_From_Random()
	account.SeedLanguage = choose_seed_language(l)

	// a new account has been created, append the seed to user home directory

	//usr, err := user.Current()
	/*if err != nil {
	      globals.Logger.Warnf("Cannot get current username to save recovery key and password")
	  }else{ // we have a user, get his home dir


	  }*/

	return account
}

// create a new wallet from hex seed provided
func Create_New_Account_from_seed(l *readline.Instance) *walletapi.Account {

	var account *walletapi.Account
	var seedkey crypto.Key

	seed := read_line_with_prompt(l, "Please enter your seed ( hex 64 chars): ")
	seed = strings.TrimSpace(seed)          // trim any extra space
	seed_raw, err := hex.DecodeString(seed) // hex decode
	if len(seed) != 64 || err != nil {      //sanity check
		globals.Logger.Warnf("Seed must be 64 chars hexadecimal chars")
		return account
	}

	copy(seedkey[:], seed_raw[:32])                            // copy bytes to seed
	account, _ = walletapi.Generate_Account_From_Seed(seedkey) // create a new account
	account.SeedLanguage = choose_seed_language(l)             // ask user his seed preference and set it

	account_valid = true

	return account
}

// create a new wallet from viewable seed provided
// viewable seed consists of public spend key and private view key
func Create_New_Account_from_viewable_key(l *readline.Instance) *walletapi.Account {

	var seedkey crypto.Key
	var privateview crypto.Key

	var account *walletapi.Account
	seed := read_line_with_prompt(l, "Please enter your View Only Key ( hex 128 chars): ")

	seed = strings.TrimSpace(seed) // trim any extra space

	seed_raw, err := hex.DecodeString(seed)
	if len(seed) != 128 || err != nil {
		globals.Logger.Warnf("View Only key must be 128 chars hexadecimal chars")
		return account
	}

	copy(seedkey[:], seed_raw[:32])
	copy(privateview[:], seed_raw[32:64])

	account, _ = walletapi.Generate_Account_View_Only(seedkey, privateview)

	account_valid = true

	return account
}

// helper function to let user to choose a seed in specific lanaguage
func choose_seed_language(l *readline.Instance) string {
	languages := mnemonics.Language_List()
	fmt.Printf("Language list for seeds, please enter a number (default English)\n")
	for i := range languages {
		fmt.Fprintf(l.Stderr(), "\033[1m%2d:\033[0m %s\n", i, languages[i])
	}

	language_number := read_line_with_prompt(l, "Please enter a choice: ")
	choice := 0 // 0 for english

	if s, err := strconv.Atoi(language_number); err == nil {
		choice = s
	}

	for i := range languages { // if user gave any wrong or ot of range choice, choose english
		if choice == i {
			return languages[choice]
		}
	}
	// if no match , return Englisg
	return "English"

}

// read a line from the prompt
func read_line_with_prompt(l *readline.Instance, prompt_temporary string) string {
	prompt_mutex.Lock()
	defer prompt_mutex.Unlock()
	l.SetPrompt(prompt_temporary)
	line, err := l.Readline()
	if err == readline.ErrInterrupt {
		if len(line) == 0 {
			globals.Logger.Infof("Ctrl-C received, Exiting\n")
			os.Exit(0)
		}
	} else if err == io.EOF {
		os.Exit(0)
	}
	l.SetPrompt(prompt)
	return line

}

// filter out specfic inputs from input processing
// currently we only skip CtrlZ background key
func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

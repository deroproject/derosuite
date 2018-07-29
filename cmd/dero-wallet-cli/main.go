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

import "sync/atomic"

//import "io/ioutil"
//import "bufio"
//import "bytes"
//import "net/http"
//import "encoding/hex"

import "github.com/romana/rlog"
import "github.com/chzyer/readline"
import "github.com/docopt/docopt-go"
import log "github.com/sirupsen/logrus"

//import "github.com/vmihailenco/msgpack"

//import "github.com/deroproject/derosuite/address"

import "github.com/deroproject/derosuite/config"
import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/walletapi"
import "github.com/deroproject/derosuite/walletapi/mnemonics"

var command_line string = `dero-wallet-cli 
DERO : A secure, private blockchain with smart-contracts

Usage:
  dero-wallet-cli [options] 
  dero-wallet-cli -h | --help
  dero-wallet-cli --version

  Options:
  -h --help     Show this screen.
  --version     Show version.
  --wallet-file=<file>  Use this file to restore or create new wallet
  --password=<password>  Use this password to unlock the wallet
  --offline     Run the wallet in completely offline mode 
  --offline_datafile=<file>  Use the data in offline mode default ("getoutputs.bin") in current dir
  --prompt      Disable menu and display prompt
  --testnet  	Run in testnet mode.
  --debug       Debug mode enabled, print log messages
  --unlocked    Keep wallet unlocked for cli commands (Does not confirm password before commands)
  --generate-new-wallet Generate new wallet
  --restore-deterministic-wallet    Restore wallet from previously saved recovery seed
  --electrum-seed=<recovery-seed>   Seed to use while restoring wallet
  --socks-proxy=<socks_ip:port>  Use a proxy to connect to Daemon.
  --daemon-address=<host:port>    Use daemon instance at <host>:<port>
  --rpc-server      Run rpc server, so wallet is accessible using api
  --rpc-bind=<127.0.0.1:20209>  Wallet binds on this ip address and port
  --rpc-login=<username:password>  RPC server will grant access based on these credentials
  `
var menu_mode bool = true // default display menu mode
//var account_valid bool = false                        // if an account has been opened, do not allow to create new account in this session
var offline_mode bool        // whether we are in offline mode
var sync_in_progress int     //  whether sync is in progress with daemon
var wallet *walletapi.Wallet //= &walletapi.Account{} // all account  data is available here
//var address string
var sync_time time.Time // used to suitable update  prompt

var default_offline_datafile string = "getoutputs.bin"

var color_black = "\033[30m"
var color_red = "\033[31m"
var color_green = "\033[32m"
var color_yellow = "\033[33m"
var color_blue = "\033[34m"
var color_magenta = "\033[35m"
var color_cyan = "\033[36m"
var color_white = "\033[37m"
var color_extra_white = "\033[1m"
var color_normal = "\033[0m"

var prompt_mutex sync.Mutex // prompt lock
var prompt string = "\033[92mDERO Wallet:\033[32m>>>\033[0m "

var tablock uint32

func main() {

	var err error

	globals.Init_rlog()

	globals.Arguments, err = docopt.Parse(command_line, nil, true, "DERO atlantis wallet : work in progress", false)
	//globals.Arguments, err = docopt.ParseArgs(command_line, os.Args[1:],  "DERO daemon : work in progress")
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

	// get ready to grab passwords
	setPasswordCfg := l.GenPasswordConfig()
	setPasswordCfg.SetListener(func(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
		l.SetPrompt(fmt.Sprintf("Enter password(%v): ", len(line)))
		l.Refresh()
		return nil, 0, false
	})
	l.Refresh() // refresh the prompt

	// parse arguments and setup testnet mainnet
	globals.Initialize()     // setup network and proxy
	globals.Logger.Infof("") // a dummy write is required to fully activate logrus

	// all screen output must go through the readline
	globals.Logger.Out = l.Stdout()

	rlog.Infof("Arguments %+v", globals.Arguments)
	globals.Logger.Infof("DERO Wallet : %s  This version is under heavy development, use it for testing/evaluations purpose only", config.Version.String())
	globals.Logger.Infof("Copyright 2017-2018 DERO Project. All rights reserved.")
	globals.Logger.Infof("OS:%s ARCH:%s GOMAXPROCS:%d", runtime.GOOS, runtime.GOARCH, runtime.GOMAXPROCS(0))
	globals.Logger.Infof("Wallet in %s mode", globals.Config.Name)

	// disable menu mode if requested
	if globals.Arguments["--prompt"] != nil && globals.Arguments["--prompt"].(bool) {
		menu_mode = false
	}

	wallet_file := "wallet.db" //dero.wallet"
	if globals.Arguments["--wallet-file"] != nil {
		wallet_file = globals.Arguments["--wallet-file"].(string) // override with user specified settings
	}

	wallet_password := "" // default
	if globals.Arguments["--password"] != nil {
		wallet_password = globals.Arguments["--password"].(string) // override with user specified settings
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

		account, err := walletapi.Generate_Account_From_Recovery_Words(seed)
		if err != nil {
			globals.Logger.Warnf("Error while recovering seed err %s\n", err)
			return
		}

		// ask user a pass, if not provided on command_line
		password := ""
		if wallet_password == "" {
			password = ReadConfirmedPassword(l, "Enter password", "Confirm password")
		}

		wallet, err = walletapi.Create_Encrypted_Wallet(wallet_file, password, account.Keys.Spendkey_Secret)
		if err != nil {
			globals.Logger.Warnf("Error occurred while restoring wallet. err %s", err)
			return
		}

		globals.Logger.Debugf("Seed Language %s", account.SeedLanguage)
		globals.Logger.Infof("Successfully recovered wallet from seed")

	}

	// generare new random account if requested
	if globals.Arguments["--generate-new-wallet"] != nil && globals.Arguments["--generate-new-wallet"].(bool) {
		filename := choose_file_name(l)
		// ask user a pass, if not provided on command_line
		password := ""
		if wallet_password == "" {
			password = ReadConfirmedPassword(l, "Enter password", "Confirm password")
		}

		seed_language := choose_seed_language(l)
		wallet, err = walletapi.Create_Encrypted_Wallet_Random(filename, password)
		if err != nil {
			globals.Logger.Warnf("Error occured while creating new wallet, err: %s", err)
			wallet = nil
			return

		}
		globals.Logger.Debugf("Seed Language %s", seed_language)
		display_seed(l, wallet)

	}

	if globals.Arguments["--rpc-login"] != nil {
		userpass := globals.Arguments["--rpc-login"].(string)
		parts := strings.SplitN(userpass, ":", 2)

		if len(parts) != 2 {
			globals.Logger.Warnf("RPC user name or password invalid")
			return
		}
		log.Infof("RPC username \"%s\" password \"%s\" ", parts[0], parts[1])
	}

	// if wallet is nil,  check whether the file exists, if yes, request password
	if wallet == nil {
		if _, err = os.Stat(wallet_file); err == nil {

			// if a wallet file and password  has been provide, make sure that the wallet opens in 1st attempt, othwer wise exit

			if globals.Arguments["--password"] != nil {
				wallet, err = walletapi.Open_Encrypted_Wallet(wallet_file, wallet_password)
				if err != nil {
					globals.Logger.Warnf("Error occurred while opening wallet. err %s", err)
					os.Exit(-1)
				}
			} else { // request user the password

				// ask user a password
				for i := 0; i < 3; i++ {
					wallet, err = walletapi.Open_Encrypted_Wallet(wallet_file, ReadPassword(l, wallet_file))
					if err != nil {
						globals.Logger.Warnf("Error occurred while opening wallet. err %s", err)
					} else { //  user knows the password and is db is valid
						break
					}
				}
			}

			//globals.Logger.Debugf("Seed Language %s", account.SeedLanguage)
			//globals.Logger.Infof("Successfully recovered wallet from seed")

		}
	}

	// check if offline mode requested
	if wallet != nil {
		common_processing(wallet)
	}

	//pipe_reader, pipe_writer = io.Pipe() // create pipes

	// reader ready to parse any data from the file
	//go blockchain_data_consumer()

	// update prompt when required
	prompt_mutex.Lock()
	go update_prompt(l)
	prompt_mutex.Unlock()

	// if wallet has been opened in offline mode by commands supplied at command prompt
	// trigger the offline scan

	//	go trigger_offline_data_scan()

	// start infinite loop processing user commands
	for {

		prompt_mutex.Lock()
		if globals.Exit_In_Progress { // exit if requested so
			prompt_mutex.Unlock()
			break
		}
		prompt_mutex.Unlock()

		if menu_mode { // display menu if requested
			if wallet != nil { // account is opened, display post menu
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
			if wallet != nil {
				if !handle_easymenu_post_open_command(l, line) { // if not processed , try processing as command
					handle_prompt_command(l, line)
					PressAnyKey(l, wallet)
				}
			} else {
				handle_easymenu_pre_open_command(l, line)
			}
		} else {
			handle_prompt_command(l, line)
		}

	}
	prompt_mutex.Lock()
	globals.Exit_In_Progress = true
	prompt_mutex.Unlock()

}

// update prompt as and when necessary
// TODO: make this code simple, with clear direction
func update_prompt(l *readline.Instance) {

	last_wallet_height := uint64(0)
	last_daemon_height := uint64(0)
	daemon_online := false
	last_update_time := int64(0)

	for {
		time.Sleep(30 * time.Millisecond) // give user a smooth running number

		prompt_mutex.Lock()
		if globals.Exit_In_Progress {
			prompt_mutex.Unlock()
			return
		}
		prompt_mutex.Unlock()

		if atomic.LoadUint32(&tablock) > 0 { // tab key has been presssed,  stop delivering updates to  prompt
			continue
		}

		prompt_mutex.Lock() // do not update if we can not lock the mutex

		// show first 8 bytes of address
		address_trim := ""
		if wallet != nil {
			tmp_addr := wallet.GetAddress().String()
			address_trim = tmp_addr[0:8]
		} else {
			address_trim = "DERO Wallet"
		}

		if wallet == nil {
			l.SetPrompt(fmt.Sprintf("\033[1m\033[32m%s \033[0m"+color_green+"0/%d \033[32m>>>\033[0m ", address_trim, 0))
			prompt_mutex.Unlock()
			continue
		}

		// only update prompt if needed, or update atleast once every second

		if last_wallet_height != wallet.Get_Height() || last_daemon_height != wallet.Get_Daemon_Height() ||
			daemon_online != wallet.IsDaemonOnlineCached() || (time.Now().Unix()-last_update_time) >= 1 {
			// choose color based on urgency
			color := "\033[32m" // default is green color
			if wallet.Get_Height() < wallet.Get_Daemon_Height() {
				color = "\033[33m" // make prompt yellow
			}

			dheight := wallet.Get_Daemon_Height()

			if wallet.IsDaemonOnlineCached() == false {
				color = "\033[33m" // make prompt yellow
				dheight = 0
			}

			balance_string := ""

			//balance_unlocked, locked_balance := wallet.Get_Balance_Rescan()// wallet.Get_Balance()
			balance_unlocked, locked_balance := wallet.Get_Balance()
			balance_string = fmt.Sprintf(color_green+"%s "+color_white+"| "+color_yellow+"%s", globals.FormatMoney8(balance_unlocked), globals.FormatMoney8(locked_balance))

			testnet_string := ""
			if !globals.IsMainnet() {
				testnet_string = "\033[31m TESTNET"
			}

			l.SetPrompt(fmt.Sprintf("\033[1m\033[32m%s \033[0m"+color+"%d/%d %s %s\033[32m>>>\033[0m ", address_trim, wallet.Get_Height(), dheight, balance_string, testnet_string))
			l.Refresh()
			last_wallet_height = wallet.Get_Height()
			last_daemon_height = wallet.Get_Daemon_Height()
			last_update_time = time.Now().Unix()
			daemon_online = wallet.IsDaemonOnlineCached()
			_ = last_update_time

		}

		prompt_mutex.Unlock()

	}

}

// create a new wallet from scratch from random numbers
func Create_New_Wallet(l *readline.Instance) (w *walletapi.Wallet, err error) {

	// ask user a file name to store the data

	walletpath := read_line_with_prompt(l, "Please enter wallet file name : ")
	walletpassword := ""

	account, _ := walletapi.Generate_Keys_From_Random()
	account.SeedLanguage = choose_seed_language(l)

	w, err = walletapi.Create_Encrypted_Wallet(walletpath, walletpassword, account.Keys.Spendkey_Secret)

	if err != nil {
		return
	}

	// set wallet seed language

	// a new account has been created, append the seed to user home directory

	//usr, err := user.Current()
	/*if err != nil {
	      globals.Logger.Warnf("Cannot get current username to save recovery key and password")
	  }else{ // we have a user, get his home dir


	  }*/

	return
}

/*

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
*/
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

// lets the user choose a filename or use default
func choose_file_name(l *readline.Instance) (filename string) {

	default_filename := "wallet.db"
	if globals.Arguments["--wallet-file"] != nil {
		default_filename = globals.Arguments["--wallet-file"].(string) // override with user specified settings
	}

	filename = read_line_with_prompt(l, fmt.Sprintf("Enter wallet filename (default %s): ", default_filename))

	if len(filename) < 1 {
		filename = default_filename
	}

	return
}

// read a line from the prompt
// since we cannot query existing, we can get away by using password mode with
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
	case readline.CharTab:
		atomic.StoreUint32(&tablock, 1) // lock prompt update
	case readline.CharEnter:
		atomic.StoreUint32(&tablock, 0) // enable prompt update
	}
	return r, true
}

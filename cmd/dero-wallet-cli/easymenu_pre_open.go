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
import "fmt"
import "strings"
import "encoding/hex"

import "github.com/chzyer/readline"

import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/config"
import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/walletapi"

// display menu before a wallet is opened
func display_easymenu_pre_open_command(l *readline.Instance) {
	w := l.Stderr()
	io.WriteString(w, "Menu:\n")
	io.WriteString(w, "\t\033[1m1\033[0m\tOpen existing Wallet\n")
	io.WriteString(w, "\t\033[1m2\033[0m\tCreate New Wallet\n")
	io.WriteString(w, "\t\033[1m3\033[0m\tRecover Wallet using recovery seed (25 words)\n")
	io.WriteString(w, "\t\033[1m4\033[0m\tRecover Wallet using recovery key (64 char private spend key hex)\n")
	io.WriteString(w, "\t\033[1m5\033[0m\tCreate  Watch-able Wallet (view only) using wallet view key\n")
        io.WriteString(w, "\t\033[1m6\033[0m\tRecover Non-deterministic Wallet key\n")

	io.WriteString(w, "\n\t\033[1m9\033[0m\tExit menu and start prompt\n")
	io.WriteString(w, "\t\033[1m0\033[0m\tExit Wallet\n")
}

// handle all commands
func handle_easymenu_pre_open_command(l *readline.Instance, line string) {
	var err error

	line = strings.TrimSpace(line)
	line_parts := strings.Fields(line)

	if len(line_parts) < 1 { // if no command return
		return
	}

	command := ""
	if len(line_parts) >= 1 {
		command = strings.ToLower(line_parts[0])
	}

	//account_state := account_valid
	switch command {
	case "1": // open existing wallet
		filename := choose_file_name(l)

		// ask user a password
		for i := 0; i < 3; i++ {
			wallet, err = walletapi.Open_Encrypted_Wallet(filename, ReadPassword(l, filename))
			if err != nil {
				globals.Logger.Warnf("Error occurred while opening wallet file %s. err %s", filename, err)
				wallet = nil
				break
			} else { //  user knows the password and is db is valid
				break
			}
		}
		if wallet != nil {
			globals.Logger.Infof("Successfully opened wallet")

			common_processing(wallet)
		}

	case "2": // create a new random account

		filename := choose_file_name(l)

		password := ReadConfirmedPassword(l, "Enter password", "Confirm password")

		wallet, err = walletapi.Create_Encrypted_Wallet_Random(filename, password)
		if err != nil {
			globals.Logger.Warnf("Error occured while creating new wallet, err: %s", err)
			wallet = nil
			break

		}
		err = wallet.Set_Encrypted_Wallet_Password(password)
		if err != nil {
			globals.Logger.Warnf("Error changing password")
		}
		seed_language := choose_seed_language(l)
		wallet.SetSeedLanguage(seed_language)
		globals.Logger.Debugf("Seed Language %s", seed_language)

		display_seed(l, wallet)

		common_processing(wallet)

	case "3": // create wallet from recovery words

		filename := choose_file_name(l)
		password := ReadConfirmedPassword(l, "Enter password", "Confirm password")
		electrum_words := read_line_with_prompt(l, "Enter seed (25 words) : ")

		wallet, err = walletapi.Create_Encrypted_Wallet_From_Recovery_Words(filename, password, electrum_words)
		if err != nil {
			globals.Logger.Warnf("Error while recovering wallet using seed err %s\n", err)
			break
		}
		//globals.Logger.Debugf("Seed Language %s", account.SeedLanguage)
		globals.Logger.Infof("Successfully recovered wallet from seed")
		common_processing(wallet)

	case "4": // create wallet from  hex seed

		var seedkey crypto.Key
		filename := choose_file_name(l)
		password := ReadConfirmedPassword(l, "Enter password", "Confirm password")

		seed_key_string := read_line_with_prompt(l, "Please enter your seed ( hex 64 chars): ")

		seed_raw, err := hex.DecodeString(seed_key_string) // hex decode
		if len(seed_key_string) != 64 || err != nil {      //sanity check
			globals.Logger.Warnf("Seed must be 64 chars hexadecimal chars")
			break
		}

		copy(seedkey[:], seed_raw[:32])

		wallet, err = walletapi.Create_Encrypted_Wallet(filename, password, seedkey)
		if err != nil {
			globals.Logger.Warnf("Error while recovering wallet using seed key err %s\n", err)
			break
		}
		globals.Logger.Infof("Successfully recovered wallet from hex seed")
		seed_language := choose_seed_language(l)
		wallet.SetSeedLanguage(seed_language)
		globals.Logger.Debugf("Seed Language %s", seed_language)

		display_seed(l, wallet)
		common_processing(wallet)

	case "5": // create new view only wallet // TODO user providing wrong key is not being validated, do it ASAP

		filename := choose_file_name(l)
		view_key_string := read_line_with_prompt(l, "Please enter your View Only Key ( hex 128 chars): ")

		password := ReadConfirmedPassword(l, "Enter password", "Confirm password")
		wallet, err = walletapi.Create_Encrypted_Wallet_ViewOnly(filename, password, view_key_string)

		if err != nil {
			globals.Logger.Warnf("Error while reconstructing view only wallet using view key err %s\n", err)
			break
		}

		if globals.Arguments["--offline"].(bool) == true {
			//offline_mode = true
		} else {
			wallet.SetOnlineMode()
		}
        case "6": // create non deterministic wallet // TODO user providing wrong key is not being validated, do it ASAP

		filename := choose_file_name(l)
		spend_key_string := read_line_with_prompt(l, "Please enter your Secret spend key ( hex 64 chars): ")
                view_key_string := read_line_with_prompt(l, "Please enter your Secret view key ( hex 64 chars): ")

		password := ReadConfirmedPassword(l, "Enter password", "Confirm password")
		wallet, err = walletapi.Create_Encrypted_Wallet_NonDeterministic(filename, password, spend_key_string,view_key_string)

		if err != nil {
			globals.Logger.Warnf("Error while reconstructing view only wallet using view key err %s\n", err)
			break
		}

		if globals.Arguments["--offline"].(bool) == true {
			//offline_mode = true
		} else {
			wallet.SetOnlineMode()
		}

	case "9":
		menu_mode = false
		globals.Logger.Infof("Prompt mode enabled")
	case "0", "bye", "exit", "quit":
		globals.Exit_In_Progress = true
	default: // just loop

	}
	//_ = account_state

	// NOTE: if we are in online mode, it is handled automatically
	// user opened or created a new account
	// rescan blockchain in offline mode
	//if account_state == false && account_valid && offline_mode {
	//	go trigger_offline_data_scan()
	//}

}

// sets online mode, starts RPC server etc
func common_processing(wallet *walletapi.Wallet) {
	if globals.Arguments["--offline"].(bool) == true {
		//offline_mode = true
	} else {
		wallet.SetOnlineMode()
	}

	// start rpc server if requested
	if globals.Arguments["--rpc-server"].(bool) == true {
		rpc_address := "127.0.0.1:" + fmt.Sprintf("%d", config.Mainnet.Wallet_RPC_Default_Port)

		if !globals.IsMainnet() {
			rpc_address = "127.0.0.1:" + fmt.Sprintf("%d", config.Testnet.Wallet_RPC_Default_Port)
		}

		if globals.Arguments["--rpc-bind"] != nil {
			rpc_address = globals.Arguments["--rpc-bind"].(string)
		}
		globals.Logger.Infof("Starting RPC server at %s", rpc_address)
		err := wallet.Start_RPC_Server(rpc_address)
		if err != nil {
			globals.Logger.Warnf("Error starting rpc server err %s", err)

		}

	}

}

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
import "strings"

import "github.com/chzyer/readline"

import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/walletapi"

// display menu before a wallet is opened
func display_easymenu_pre_open_command(l *readline.Instance) {
	w := l.Stderr()
	io.WriteString(w, "Menu:\n")
	io.WriteString(w, "\t\033[1m1\033[0m\tCreate New Wallet\n")
	io.WriteString(w, "\t\033[1m2\033[0m\tRecover Wallet using recovery seed (25 words)\n")
	io.WriteString(w, "\t\033[1m3\033[0m\tRecover Wallet using recovery key (64 char private spend key hex)\n")
	io.WriteString(w, "\t\033[1m4\033[0m\tCreate  Watch-able Wallet (view only) using wallet view key\n")

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

	account_state := account_valid
	switch command {
	case "1": // create a new random account
		if account_valid {
			globals.Logger.Warnf("Account already exists. Cannot create new account")
			break
		}
		account = Create_New_Account(l)
		account_valid = true
		globals.Logger.Debugf("Seed Language %s", account.SeedLanguage)
		address = account.GetAddress().String()
		display_seed(l)

		if offline_mode {
			go trigger_offline_data_scan()
		}

	case "2": // create wallet from recovery seed
		if account_valid {
			globals.Logger.Warnf("Account already exists. Cannot recover account")
			break
		}
		seed := read_line_with_prompt(l, "Enter your seed (25 words) : ")

		account, err = walletapi.Generate_Account_From_Recovery_Words(seed)
		if err != nil {
			globals.Logger.Warnf("Error while recovering seed err %s\n", err)
			break
		}
		account_valid = true
		globals.Logger.Debugf("Seed Language %s", account.SeedLanguage)
		globals.Logger.Infof("Successfully recovered wallet from seed")
		address = account.GetAddress().String()
		if offline_mode {
			go trigger_offline_data_scan()
		}
	case "3": // create wallet from  hex seed
		account = Create_New_Account_from_seed(l)
		if account_valid {
			globals.Logger.Infof("Successfully recovered wallet from hex seed")
			display_seed(l)
			address = account.GetAddress().String()
			if offline_mode {
				go trigger_offline_data_scan()
			}

		}

	case "4": // create new view only wallet // TODO user providing wrong key is not being validated, do it ASAP
		account = Create_New_Account_from_viewable_key(l)
		if account_valid {
			globals.Logger.Infof("Successfully created view only wallet from viewable keys")
			address = account.GetAddress().String()
			account_valid = true

			if offline_mode {
				go trigger_offline_data_scan()
			}
		}

	case "9":
		menu_mode = false
		globals.Logger.Infof("Prompt mode enabled")
	case "0", "bye", "exit", "quit":
		globals.Exit_In_Progress = true
	default: // just loop

	}
	_ = account_state

	// NOTE: if we are in online mode, it is handled automatically
	// user opened or created a new account
	// rescan blockchain in offline mode
	//if account_state == false && account_valid && offline_mode {
	//	go trigger_offline_data_scan()
	//}

}

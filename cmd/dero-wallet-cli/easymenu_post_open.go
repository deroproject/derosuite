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

import "github.com/chzyer/readline"

import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/walletapi"

// handle menu if a wallet is currently opened
func display_easymenu_post_open_command(l *readline.Instance) {
	w := l.Stderr()
	io.WriteString(w, "Menu:\n")

	io.WriteString(w, "\t\033[1m1\033[0m\tDisplay account Address \n")
	if !account.ViewOnly { // hide some commands, if view only wallet
		io.WriteString(w, "\t\033[1m2\033[0m\tDisplay Seed "+color_red+"(Please save seed otherwise your wallet is LOST)\n\033[0m")
	}
	io.WriteString(w, "\t\033[1m3\033[0m\tDisplay Keys (hex)\n")
	io.WriteString(w, "\t\033[1m4\033[0m\tDisplay Watch-able View only wallet key ( cannot spend any funds from this wallet) (hex)\n")

	if !account.ViewOnly { // hide some commands, if view only wallet
		io.WriteString(w, "\t\033[1m5\033[0m\tTransfer ( send  DERO) To Another Wallet\n")
		io.WriteString(w, "\t\033[1m6\033[0m\tTransfer ( send  DERO) To Exchange (payment id mandatory\n")
		io.WriteString(w, "\t\033[1m7\033[0m\tCreate Transaction in offline mode\n")
	}

	io.WriteString(w, "\t\033[1m8\033[0m\tClose Wallet\n")

	io.WriteString(w, "\n\t\033[1m9\033[0m\tExit menu and start prompt\n")
	io.WriteString(w, "\t\033[1m0\033[0m\tExit Wallet\n")

}

// this handles all the commands if wallet in menu mode  and a wallet is opened
func handle_easymenu_post_open_command(l *readline.Instance, line string) {

	var err error
	_ = err
	line = strings.TrimSpace(line)
	line_parts := strings.Fields(line)

	if len(line_parts) < 1 { // if no command return
		return
	}

	command := ""
	if len(line_parts) >= 1 {
		command = strings.ToLower(line_parts[0])
	}

	switch command {
	case "1":
		if account_valid {
			fmt.Fprintf(l.Stderr(), "%s\n", account.GetAddress())
		}
	case "2": // give user his seed
		if !account.ViewOnly {
			display_seed(l)
		}

	case "3": // give user his keys in hex form
		display_spend_key(l)
		display_view_key(l)

	case "4": // display user keys to create view only wallet
		display_viewwallet_key(l)
	case "5", "6", "7":
		if !account.ViewOnly {
			globals.Logger.Warnf("This command is NOT yet implemented")
		}

	case "8": // close and discard user key
		account_valid = false
		account = &walletapi.Account{} // overwrite previous instance
		address = ""                   // empty the address
		Wallet_Height = 0

	case "9": // enable prompt mode
		menu_mode = false
		globals.Logger.Infof("Prompt mode enabled")
	case "0", "bye", "exit", "quit":
		globals.Exit_In_Progress = true

	default: // just loop

	}

}

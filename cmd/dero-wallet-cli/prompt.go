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

import "os"
import "io"
import "fmt"
import "bufio"
import "strings"
import "strconv"
import "compress/gzip"

import "github.com/chzyer/readline"

import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/walletapi"

// handle all commands while  in prompt mode
func handle_prompt_command(l *readline.Instance, line string) {

	var err error
	line = strings.TrimSpace(line)
	line_parts := strings.Fields(line)

	if len(line_parts) < 1 { // if no command return
		return
	}

	_ = err
	command := ""
	if len(line_parts) >= 1 {
		command = strings.ToLower(line_parts[0])
	}

	switch command {
	case "help":
		usage(l.Stderr())
	case "address": // give user his account address
		if account_valid {
			fmt.Fprintf(l.Stderr(), "%s\n", account.GetAddress())
		}
	case "rescan_bc": // rescan from 0
		fallthrough
	case "rescan_spent": // rescan from 0
		if offline_mode {
			go trigger_offline_data_scan()
		} else {
			globals.Logger.Warnf("This command is NOT yet implemented")
		}

	case "seed": // give user his seed
		display_seed(l)
	case "spendkey": // give user his spend key
		display_spend_key(l)
	case "viewkey": // give user his viewkey
		display_view_key(l)
	case "walletviewkey":
		display_viewwallet_key(l)

	case "set": // set different settings
	case "close": // close the account
		account_valid = false
		tmp := &walletapi.Account{}
		address = ""
		account = tmp // overwrite previous instance

	case "menu": // enable menu mode
		menu_mode = true
		globals.Logger.Infof("Menu mode enabled")
	case "bye", "exit", "quit":
		globals.Exit_In_Progress = true

	case "": // blank enter key just loop
	default:
		fmt.Fprintf(l.Stderr(), "you said: %s", strconv.Quote(line))
	}

}

// if we are in offline, scan default or user provided file
// this function will replay the blockchain data in offline mode
func trigger_offline_data_scan() {
	filename := default_offline_datafile

	if globals.Arguments["--offline_datafile"] != nil {
		filename = globals.Arguments["--offline_datafile"].(string)
	}

	f, err := os.Open(filename)
	if err != nil {
		globals.Logger.Warnf("Cannot read offline data file=\"%s\"  err: %s   ", filename, err)
		return
	}
	w := bufio.NewReader(f)
	gzipreader, err := gzip.NewReader(w)
	if err != nil {
		globals.Logger.Warnf("Error while decompressing offline data file=\"%s\"  err: %s   ", filename, err)
		return
	}
	defer gzipreader.Close()
	io.Copy(pipe_writer, gzipreader)
}

// this completer is used to complete the commands at the prompt
// BUG, this needs to be disabled in menu mode
var completer = readline.NewPrefixCompleter(
	readline.PcItem("help"),
	readline.PcItem("address"),
	readline.PcItem("rescan_bc"),
	readline.PcItem("rescan_spent"),
	readline.PcItem("print_height"),
	readline.PcItem("seed"),
	readline.PcItem("menu"),
	readline.PcItem("set",
		readline.PcItem("priority",
			readline.PcItem("lowest x1"),
			readline.PcItem("low x4"),
			readline.PcItem("normal x8"),
			readline.PcItem("high x13"),
			readline.PcItem("veryhigh x20"),
		),
		readline.PcItem("default-ring-size"),
		readline.PcItem("store-tx-info"),
		readline.PcItem("ask-password"),
	),
	readline.PcItem("spendkey"),
	readline.PcItem("viewkey"),
	readline.PcItem("walletviewkey"),
	readline.PcItem("bye"),
	readline.PcItem("exit"),
	readline.PcItem("quit"),
)

// help command screen
func usage(w io.Writer) {
	io.WriteString(w, "commands:\n")
	io.WriteString(w, "\t\033[1mhelp\033[0m\t\tthis help\n")
	io.WriteString(w, "\t\033[1maddress\033[0m\t\tDisplay user address\n")
	io.WriteString(w, "\t\033[1mmenu\033[0m\t\tEnable menu mode\n")
	io.WriteString(w, "\t\033[1mrescan_bc\033[0m\tRescan blockchain again from 0 height\n")
	io.WriteString(w, "\t\033[1mprint_block\033[0m\tPrint block, print_block <block_hash> or <block_height>\n")
	io.WriteString(w, "\t\033[1mseed\033[0m\tDisplay seed\n")
	io.WriteString(w, "\t\033[1mset\033[0m\tSet various settings\n")
	io.WriteString(w, "\t\033[1mstatus\033[0m\t\tShow genereal information\n")
	io.WriteString(w, "\t\033[1mspendkey\033[0m\tView secret key\n")
	io.WriteString(w, "\t\033[1mviewkey\033[0m\tView view key\n")
	io.WriteString(w, "\t\033[1mwalletviewkey\033[0m\tWallet view key, used to create watchable view only wallet\n")
	io.WriteString(w, "\t\033[1mbye\033[0m\t\tQuit wallet\n")
	io.WriteString(w, "\t\033[1mexit\033[0m\t\tQuit wallet\n")
	io.WriteString(w, "\t\033[1mquit\033[0m\t\tQuit wallet\n")

}

// display seed to the user in his preferred language
func display_seed(l *readline.Instance) {
	if account_valid {
		seed := account.GetSeed()
		fmt.Fprintf(l.Stderr(), color_green+"PLEASE NOTE: the following 25 words can be used to recover access to your wallet. Please write them down and store them somewhere safe and secure. Please do not store them in your email or on file storage services outside of your immediate control."+color_white+"\n")
		fmt.Fprintf(os.Stderr, color_red+"%s"+color_white+"\n", seed)
	}

}

// display spend key
// viewable wallet do not have spend secret key
// TODO wee need to give user a warning if we are printing secret
func display_spend_key(l *readline.Instance) {
	if account_valid {
		if !account.ViewOnly {
			fmt.Fprintf(os.Stderr, "spend key secret : "+color_red+"%s"+color_white+"\n", account.Keys.Spendkey_Secret)
		}
		fmt.Fprintf(os.Stderr, "spend key public : %s\n", account.Keys.Spendkey_Public)
	}

}

//display view key
func display_view_key(l *readline.Instance) {
	if account_valid {
		fmt.Fprintf(os.Stderr, "view key secret : "+color_yellow+"%s"+color_white+"\n", account.Keys.Viewkey_Secret)
		fmt.Fprintf(os.Stderr, "view key public : %s\n", account.Keys.Viewkey_Public)
	}

}

// display wallet view only Keys
// this will create a watchable view only mode
func display_viewwallet_key(l *readline.Instance) {
	if account_valid {
		io.WriteString(l.Stderr(), fmt.Sprintf("This Key can used to create a watch only wallet. This wallet can only see incoming funds but cannot spend them.\nThe key is below.\n%s\n\n", account.GetViewWalletKey()))
	}

}

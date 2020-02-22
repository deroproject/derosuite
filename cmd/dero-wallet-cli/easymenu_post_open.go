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
import "fmt"
import "io/ioutil"
import "strings"
import "path/filepath"
import "encoding/hex"

import "github.com/chzyer/readline"

import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/address"
import "github.com/deroproject/derosuite/walletapi"
import "github.com/deroproject/derosuite/transaction"

// handle menu if a wallet is currently opened
func display_easymenu_post_open_command(l *readline.Instance) {
	w := l.Stderr()
	io.WriteString(w, "Menu:\n")

	io.WriteString(w, "\t\033[1m1\033[0m\tDisplay account Address \n")
	if !wallet.Is_View_Only() { // hide some commands, if view only wallet
		io.WriteString(w, "\t\033[1m2\033[0m\tDisplay Seed "+color_red+"(Please save seed in safe location)\n\033[0m")
	}
	io.WriteString(w, "\t\033[1m3\033[0m\tDisplay Keys (hex)\n")
	io.WriteString(w, "\t\033[1m4\033[0m\tDisplay Watch-able View only wallet key ( cannot spend any funds from this wallet) (hex)\n")

	if !wallet.Is_View_Only() { // hide some commands, if view only wallet
		io.WriteString(w, "\t\033[1m5\033[0m\tTransfer (send  DERO) To Another Wallet\n")
		io.WriteString(w, "\t\033[1m6\033[0m\tCreate Transaction in offline mode\n")

	}

	io.WriteString(w, "\t\033[1m7\033[0m\tChange wallet password\n")
	io.WriteString(w, "\t\033[1m8\033[0m\tClose Wallet\n")
	io.WriteString(w, "\t\033[1m11\033[0m\tRescan Blockchain\n")
	if !wallet.Is_View_Only() {
		io.WriteString(w, "\t\033[1m12\033[0m\tTransfer all balance (send  DERO) To Another Wallet\n")
		io.WriteString(w, "\t\033[1m13\033[0m\tShow transaction history\n")
	}
	io.WriteString(w, "\t\033[1m14\033[0m\tRescan Balance\n")

	io.WriteString(w, "\n\t\033[1m9\033[0m\tExit menu and start prompt\n")
	io.WriteString(w, "\t\033[1m0\033[0m\tExit Wallet\n")

}

// this handles all the commands if wallet in menu mode  and a wallet is opened
func handle_easymenu_post_open_command(l *readline.Instance, line string) (processed bool) {

	var err error
	_ = err
	line = strings.TrimSpace(line)
	line_parts := strings.Fields(line)
	processed = true

	if len(line_parts) < 1 { // if no command return
		return
	}

	command := ""
	if len(line_parts) >= 1 {
		command = strings.ToLower(line_parts[0])
	}

	offline_tx := false
	switch command {
	case "1":
		fmt.Fprintf(l.Stderr(), "Wallet address : "+color_green+"%s"+color_white+"\n", wallet.GetAddress())
		PressAnyKey(l, wallet)

	case "2": // give user his seed
		if !wallet.Is_View_Only() {
			if !ValidateCurrentPassword(l, wallet) {
				globals.Logger.Warnf("Invalid password")
				PressAnyKey(l, wallet)
				break
			}
			display_seed(l, wallet) // seed should be given only to authenticated users
		} else {
			fmt.Fprintf(l.Stderr(), color_red+" View wallet do not have seeds"+color_white)
		}
		PressAnyKey(l, wallet)

	case "3": // give user his keys in hex form

		if !ValidateCurrentPassword(l, wallet) {
			globals.Logger.Warnf("Invalid password")
			PressAnyKey(l, wallet)
			break
		}

		display_spend_key(l, wallet)
		display_view_key(l, wallet)
		PressAnyKey(l, wallet)

	case "4": // display user keys to create view only wallet
		if !ValidateCurrentPassword(l, wallet) {
			globals.Logger.Warnf("Invalid password")
			break
		}
		display_viewwallet_key(l, wallet)
		PressAnyKey(l, wallet)
	case "6":
		offline_tx = true
		fallthrough
	case "5":
		if wallet.Is_View_Only() {
			fmt.Fprintf(l.Stderr(), color_yellow+"View Only wallet cannot transfer."+color_white)
		}

		if !ValidateCurrentPassword(l, wallet) {
			globals.Logger.Warnf("Invalid password")
			break
		}

		// a , amount_to_transfer, err := collect_transfer_info(l,wallet)
		a, err := ReadAddress(l)
		if err != nil {
			globals.Logger.Warnf("Err :%s", err)
			break
		}

		amount_str := read_line_with_prompt(l, fmt.Sprintf("Enter amount to transfer in DERO (max TODO): "))
		amount_to_transfer, err := globals.ParseAmount(amount_str)
		if err != nil {
			globals.Logger.Warnf("Err :%s", err)
			break // invalid amount provided, bail out
		}

		var payment_id []byte
		// if user provided an integrated address donot ask him payment id
		if !a.IsIntegratedAddress() {
			payment_id, err = ReadPaymentID(l)
			if err != nil {
				globals.Logger.Warnf("Err :%s", err)
				break
			}
		} else {
			globals.Logger.Infof("Payment ID is integreted in address ID:%x", a.PaymentID)
		}

		addr_list := []address.Address{*a}
		amount_list := []uint64{amount_to_transfer} // transfer 50 dero, 2 dero
		fees_per_kb := uint64(0)                    // fees  must be calculated by walletapi
		tx, inputs, input_sum, change, err := wallet.Transfer(addr_list, amount_list, 0, hex.EncodeToString(payment_id), fees_per_kb, 0)
		_ = inputs
		if err != nil {
			globals.Logger.Warnf("Error while building Transaction err %s\n", err)
			break
		}

		build_relay_transaction(l, tx, inputs, input_sum, change, err, offline_tx, amount_list)

	case "12":
		Transfer_Everything(l)
		PressAnyKey(l, wallet) // wait for a key press

	case "7": // change password
		if ConfirmYesNoDefaultNo(l, "Change wallet password (y/N)") &&
			ValidateCurrentPassword(l, wallet) {

			new_password := ReadConfirmedPassword(l, "Enter new password", "Confirm password")
			err = wallet.Set_Encrypted_Wallet_Password(new_password)
			if err == nil {
				globals.Logger.Infof("Wallet password successfully changed")
			} else {
				globals.Logger.Warnf("Wallet password could not be changed err %s", err)
			}
		}

	case "8": // close and discard user key

		wallet.Close_Encrypted_Wallet()
		prompt_mutex.Lock()
		wallet = nil // overwrite previous instance
		prompt_mutex.Unlock()

		fmt.Fprintf(l.Stderr(), color_yellow+"Wallet closed"+color_white)

	case "9": // enable prompt mode
		menu_mode = false
		globals.Logger.Infof("Prompt mode enabled, type \"menu\" command to start menu mode")

	case "0", "bye", "exit", "quit":
		wallet.Close_Encrypted_Wallet() // save the wallet
		prompt_mutex.Lock()
		wallet = nil
		globals.Exit_In_Progress = true
		prompt_mutex.Unlock()
		fmt.Fprintf(l.Stderr(), color_yellow+"Wallet closed"+color_white)
		fmt.Fprintf(l.Stderr(), color_yellow+"Exiting"+color_white)

	case "11":
		rescan_bc(wallet)
	case "13":
		show_transfers(l, wallet, 100)

	case "14":
		globals.Logger.Infof("Rescanning wallet")
		balance_unlocked, locked_balance := wallet.Get_Balance_Rescan()
		fmt.Fprintf(l.Stderr(), "Locked balance   : "+color_green+"%s"+color_white+"\n", globals.FormatMoney12(locked_balance))
		fmt.Fprintf(l.Stderr(), "Unlocked balance : "+color_green+"%s"+color_white+"\n", globals.FormatMoney12(balance_unlocked))
		fmt.Fprintf(l.Stderr(), "Total balance    : "+color_green+"%s"+color_white+"\n\n", globals.FormatMoney12(locked_balance+balance_unlocked))

	default:
		processed = false // just loop

	}
	return
}

// handles the output after building tx, takes feedback, confirms or relays tx
func build_relay_transaction(l *readline.Instance, tx *transaction.Transaction, inputs []uint64, input_sum uint64, change uint64, err error, offline_tx bool, amount_list []uint64) {

	if err != nil {
		globals.Logger.Warnf("Error while building Transaction err %s\n", err)
		return
	}
	globals.Logger.Infof("%d Inputs Selected for %s DERO", len(inputs), globals.FormatMoney12(input_sum))
	amount := uint64(0)
	for i := range amount_list {
		amount += amount_list[i]
	}
	globals.Logger.Infof("Transfering total amount %s DERO", globals.FormatMoney12(amount))
	globals.Logger.Infof("change amount ( will come back ) %s DERO", globals.FormatMoney12(change))

	globals.Logger.Infof("fees %s DERO", globals.FormatMoney12(tx.RctSignature.Get_TX_Fee()))
	globals.Logger.Infof("TX Size %0.1f KiB", float32(len(tx.Serialize()))/1024.0)

	if input_sum != (amount + change + tx.RctSignature.Get_TX_Fee()) {
		panic(fmt.Sprintf("Inputs %d != outputs ( %d + %d + %d )", input_sum, amount, change, tx.RctSignature.Get_TX_Fee()))
	}

	if ConfirmYesNoDefaultNo(l, "Confirm Transaction (y/N)") {

		if offline_tx { // if its an offline tx, dump it to a file
			cur_dir, err := os.Getwd()
			if err != nil {
				globals.Logger.Warnf("Cannot obtain current directory to save tx")
				globals.Logger.Infof("Transaction discarded")
				return
			}
			filename := filepath.Join(cur_dir, tx.GetHash().String()+".tx")
			err = ioutil.WriteFile(filename, []byte(hex.EncodeToString(tx.Serialize())), 0600)
			if err == nil {
				if err == nil {
					globals.Logger.Infof("Transaction saved successfully. txid = %s", tx.GetHash())
					globals.Logger.Infof("Saved to %s", filename)
				} else {
					globals.Logger.Warnf("Error saving tx to %s, err %s", filename, err)
				}
			}

		} else {

			err = wallet.SendTransaction(tx) // relay tx to daemon/network
			if err == nil {
				globals.Logger.Infof("Transaction sent successfully. txid = %s", tx.GetHash())
			} else {
				globals.Logger.Warnf("Transaction sending failed txid = %s, err %s", tx.GetHash(), err)
			}

		}

		PressAnyKey(l, wallet) // wait for a key press
	} else {
		globals.Logger.Infof("Transaction discarded")
	}

}

// collects inputs from user to send some dero
func collect_transfer_info(l *readline.Instance, wallet *walletapi.Wallet) (a *address.Address, amount uint64, err error) {

	address_str := read_line_with_prompt(l, "Please enter destination address: ")
	// we have an address parse and verify, whethers is mainnet, testnet, belongs to our network
	a, err = globals.ParseValidateAddress(address_str)
	if err != nil {
		return // invalid address provided, bail out
	}

	amount_str := read_line_with_prompt(l, fmt.Sprintf("Please enter amount to transfer in DERO (max %s): ", "0"))
	amount, err = globals.ParseAmount(amount_str)
	if err != nil {
		return // invalid amount provided, bail out
	}

	//TODO check whether amount provided is less than balance, similiar checks are done within wallet also
	return // everything is okay return

	//err = fmt.Errorf("Account is invalid")
	//return
}

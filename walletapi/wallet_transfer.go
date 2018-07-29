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

package walletapi

import "fmt"
import "sort"
import "math/rand"
import cryptorand "crypto/rand"
import "encoding/binary"
import "encoding/hex"
import "encoding/json"

import "github.com/romana/rlog"
import "github.com/vmihailenco/msgpack"

import "github.com/deroproject/derosuite/config"
import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/crypto/ringct"
import "github.com/deroproject/derosuite/transaction"
import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/address"
import "github.com/deroproject/derosuite/structures"
import "github.com/deroproject/derosuite/blockchain/inputmaturity"




// send amount to specific addresses
func (w *Wallet) Transfer(addr []address.Address, amount []uint64, unlock_time uint64, payment_id_hex string, fees_per_kb uint64, mixin uint64) (tx *transaction.Transaction, inputs_selected []uint64, inputs_sum uint64, change_amount uint64, err error) {

    
        var  transfer_details structures.Outgoing_Transfer_Details
	w.transfer_mutex.Lock()
	defer w.transfer_mutex.Unlock()
	if mixin == 0 {
		mixin = uint64(w.account.Mixin) // use wallet mixin, if mixin not provided
	}
	if mixin < 5 { // enforce minimum mixin
		mixin = 5
	}

	// if wallet is online,take the fees from the network itself
	// otherwise use whatever user has provided
	//if w.GetMode()  {
	fees_per_kb = w.dynamic_fees_per_kb // TODO disabled as protection while lots more testing is going on
	rlog.Infof("Fees per KB %d\n", fees_per_kb)
	//}

	if fees_per_kb == 0 {
		fees_per_kb = config.FEE_PER_KB
	}

	var txw *TX_Wallet_Data
	if len(addr) != len(amount) {
		err = fmt.Errorf("Count of address and amounts mismatch")
		return
	}

	if len(addr) < 1 {
		err = fmt.Errorf("Destination address missing")
		return
	}

	var payment_id []byte // we later on find WHETHER to include it, encrypt it depending on length

	// if payment  ID is provided explicity, use it
	if payment_id_hex != "" {
		payment_id, err = hex.DecodeString(payment_id_hex) // payment_id in hex
		if err != nil {
			return
		}

		if len(payment_id) == 32 || len(payment_id) == 8 {

		} else {
			err = fmt.Errorf("Payment ID must be atleast 64 hex chars (32 bytes) or 16 hex chars 8 byte")
			return
		}

	}

	// only only single payment id
	for i := range addr {
		if addr[i].IsIntegratedAddress() && payment_id_hex != "" {
			err = fmt.Errorf("Payment ID provided in both integrated address and separately")
			return
		}
	}

	// if integrated address payment id present , normal payment id must not be provided
	for i := range addr {
		if addr[i].IsIntegratedAddress() {
			if len(payment_id) > 0 { // a transaction can have only single encrypted payment ID
				err = fmt.Errorf("More than 1 integrated address provided")
				return
			}
			payment_id = addr[i].PaymentID
		}
	}

	fees := uint64(0) // start with zero fees
	expected_fee := uint64(0)
	total_amount_required := uint64(0)

	for i := range amount {
		if amount[i] == 0 { // cannot send 0  amount
			err = fmt.Errorf("Sending 0 amount to destination NOT possible")
			return
		}
		total_amount_required += amount[i]
	}

	// infinite tries to build a transaction
	for {

		// we need to make sure that account has sufficient unlocked balance ( to send amount ) + required amount of fees
		unlocked, _ := w.Get_Balance()

		if total_amount_required >= unlocked {
			err = fmt.Errorf("Insufficent unlocked balance")
			return
		}

		// now we need to select outputs with sufficient balance
		//total_amount_required += fees
		// select few outputs randomly
		inputs_selected, inputs_sum = w.select_outputs_for_transfer(total_amount_required, fees+expected_fee, false)

		if inputs_sum < (total_amount_required + fees) {
                        err = fmt.Errorf("Insufficent unlocked balance")
			return
		}

		rlog.Infof("Selected %d (%+v) iinputs to transfer  %s DERO\n", len(inputs_selected), inputs_selected, globals.FormatMoney(inputs_sum))

		/* for i := range user.Outputs_Ready {
		       if i == 739 || i == 752  {

		       fmt.Printf("selecting inuput %d  data %+v \n",i, user.Outputs_Ready[i])
		   }

		   }*/

		// lets prepare the inputs for ringct, so as we can used them
		var inputs []ringct.Input_info
		for i := range inputs_selected {

			txw, err = w.load_funds_data(inputs_selected[i], FUNDS_BUCKET)
			if err != nil {
				err = fmt.Errorf("Error while reading available funds index( it was just selected ) index %d err %s", inputs_selected[i], err)
				return
			}

			rlog.Infof("current input  %d %d \n", i, inputs_selected[i])
			var current_input ringct.Input_info
			current_input.Amount = txw.WAmount
			current_input.Key_image = crypto.Hash(txw.WKimage)
			current_input.Sk = txw.WKey

			//current_input.Index = i   is calculated after sorting of ring members
			current_input.Index_Global = txw.TXdata.Index_Global

			// add ring members here

			// ring_size = 4
			// TODO force random ring members

			//  mandatory add ourselves as ring member, otherwise there is no point in building the tx
			current_input.Ring_Members = append(current_input.Ring_Members, current_input.Index_Global)
			current_input.Pubs = append(current_input.Pubs, txw.TXdata.InKey)

			// add necessary amount  of random ring members
			// TODO we need to make sure ring members are mature, otherwise tx will fail because o immature inputs
			// This can cause certain TX to fail
			for {

				var buf [8]byte
				cryptorand.Read(buf[:])
				r, err := w.load_ring_member(binary.LittleEndian.Uint64(buf[:]) % w.account.Index_Global)
				if err == nil {

					// make sure ring member are not repeated
					new_ring_member := true
					for j := range current_input.Ring_Members { // TODO we donot need the loop
						if r.Index_Global == current_input.Ring_Members[j] {
							new_ring_member = false // we should not use this ring member
						}
						// if ring member is not mature, choose another one
						if !inputmaturity.Is_Input_Mature(w.Get_Height(),
							r.Height,
							r.Unlock_Height,
							r.Sigtype) {
							new_ring_member = false // we should not use this ring member

						}
					}

					if !new_ring_member {
						continue
					}
					current_input.Ring_Members = append(current_input.Ring_Members, r.Index_Global)
					current_input.Pubs = append(current_input.Pubs, r.InKey)
				}
				if uint64(len(current_input.Ring_Members)) == mixin { // atleast 5 ring members
					break
				}
			}

			// sort ring members and setup index

			/*
			   if i == 0 {
			       current_input.Ring_Members = []uint64{2,3,739,1158}
			       current_input.Pubs = append(current_input.Pubs,user.Random_Outputs[2].TXdata.InKey)
			       current_input.Pubs = append(current_input.Pubs,user.Random_Outputs[3].TXdata.InKey)
			       current_input.Pubs = append(current_input.Pubs,user.Random_Outputs[739].TXdata.InKey)
			       current_input.Pubs = append(current_input.Pubs,user.Random_Outputs[1158].TXdata.InKey)
			       current_input.Index = 2
			   }
			   if i == 1 {
			       continue
			       current_input.Ring_Members = []uint64{2,3,752,1158}
			       current_input.Pubs = append(current_input.Pubs,user.Random_Outputs[2].TXdata.InKey)
			       current_input.Pubs = append(current_input.Pubs,user.Random_Outputs[3].TXdata.InKey)
			       current_input.Pubs = append(current_input.Pubs,user.Random_Outputs[752].TXdata.InKey)
			       current_input.Pubs = append(current_input.Pubs,user.Random_Outputs[1158].TXdata.InKey)
			       current_input.Index = 2

			   }
			*/
			/*   current_input.Ring_Members = []uint64{2,3,4,inputs_selected[i]}
			     current_input.Pubs = append(current_input.Pubs,user.Random_Outputs[2].TXdata.InKey)
			     current_input.Pubs = append(current_input.Pubs,user.Random_Outputs[3].TXdata.InKey)
			     current_input.Pubs = append(current_input.Pubs,user.Random_Outputs[4].TXdata.InKey)
			     current_input.Pubs = append(current_input.Pubs,user.Random_Outputs[inputs_selected[i]].TXdata.InKey)
			     current_input.Index = 3

			*/

			rlog.Infof(" current input before sorting %+v \n", current_input.Ring_Members)

			current_input = sort_ring_members(current_input)
			rlog.Infof(" current input after sorting  %+v \n", current_input.Ring_Members)
			inputs = append(inputs, current_input)
		}

		//  fees = uint64(0)
		// fees = uint64(66248560000)

		// fill in the outputs

		var outputs []ringct.Output_info

	rebuild_tx_with_correct_fee:
		outputs = outputs[:0]
		
		transfer_details.Fees = fees
		transfer_details.Amount = transfer_details.Amount[:0]
		transfer_details.Daddress = transfer_details.Daddress[:0]

		for i := range addr {
			var output ringct.Output_info
			output.Amount = amount[i]
			output.Public_Spend_Key = addr[i].SpendKey
			output.Public_View_Key = addr[i].ViewKey
			
			transfer_details.Amount = append(transfer_details.Amount,amount[i]) 
                        transfer_details.Daddress = append(transfer_details.Daddress,addr[i].String()) 

			outputs = append(outputs, output)
		}

		// get ready to receive change
		var change ringct.Output_info
		change.Amount = inputs_sum - total_amount_required - fees // we must have atleast change >= fees
		change.Public_Spend_Key = w.account.Keys.Spendkey_Public  /// fill our public spend key
		change.Public_View_Key = w.account.Keys.Viewkey_Public    // fill our public view key

		if change.Amount > 0 { // include change only if required
                    
                    transfer_details.Amount = append(transfer_details.Amount,change.Amount) 
                    transfer_details.Daddress = append(transfer_details.Daddress,w.account.GetAddress().String()) 

                        
                    outputs = append(outputs, change)
		}

		change_amount = change.Amount
		
		// if encrypted payment ids are used, they are encrypted against first output
		// if we shuffle outputs encrypted ids will break
		if unlock_time == 0  { // shuffle output and change randomly
			if  len(payment_id) == 8{ // do not shuffle if encrypted payment IDs are used

			}else{
                    globals.Global_Random.Shuffle(len(outputs), func(i, j int) {
					outputs[i], outputs[j] = outputs[j], outputs[i]
			})
                    
                }
            }

		// outputs = append(outputs, change)
		tx = w.Create_TX_v2(inputs, outputs, fees, unlock_time, payment_id, true)

		tx_size := uint64(len(tx.Serialize()))
		size_in_kb := tx_size / 1024

		if (tx_size % 1024) != 0 { // for any part there of, use a full KB fee
			size_in_kb += 1
		}

		minimum_fee := size_in_kb * fees_per_kb

		needed_fee := w.getfees(minimum_fee) // multiply minimum fees by multiplier

		rlog.Infof("minimum fee %s required fees %s provided fee %s size %d fee/kb %s\n", globals.FormatMoney(minimum_fee), globals.FormatMoney(needed_fee), globals.FormatMoney(fees), size_in_kb, globals.FormatMoney(fees_per_kb))

		if fees > needed_fee { // transaction was built up successfully
			fees = needed_fee // setup fees parameter exactly as much required
			goto rebuild_tx_with_correct_fee
		}

		// keep trying until we are successfull or funds become Insufficent
		if fees == needed_fee { // transaction was built up successfully
			break
		}

		// we need to try again
		fees = needed_fee               // setup estimated parameter
		expected_fee = expected_fee * 2 // double the estimated fee

	}
	
	// log enough information to wallet to display it again to users
	transfer_details.PaymentID =  hex.EncodeToString(payment_id) 
        
        // get the tx secret key and store it
        txhash := tx.GetHash()
        transfer_details.TXsecretkey = w.GetTXKey(tx.GetHash())
        transfer_details.TXID = txhash.String()
        
        // lets marshal the structure and store it in in DB
        
        details_serialized, err := json.Marshal(transfer_details)
	if err != nil {
                rlog.Warnf("Err marshalling details err %s", err) 
	}
        
        w.store_key_value(BLOCKCHAIN_UNIVERSE, []byte(TX_OUT_DETAILS_BUCKET), txhash[:], details_serialized[:])
        
//        fmt.Printf("%+v\n",transfer_details)
//        fmt.Printf("%+v\n",transfer_details,w.GetTXOutDetails(tx.GetHash()))

        

	// log enough information in log file to validate sum(inputs) = sum(outputs) + fees

	{

		rlog.Infof("Transfering total amount %s \n", globals.FormatMoneyPrecision(inputs_sum, 12))
		rlog.Infof("total amount (output) %s \n", globals.FormatMoneyPrecision(total_amount_required, 12))
		rlog.Infof("change amount ( will come back ) %s \n", globals.FormatMoneyPrecision(change_amount, 12))
		rlog.Infof("fees %s \n", globals.FormatMoneyPrecision(tx.RctSignature.Get_TX_Fee(), 12))
		rlog.Infof("Inputs %d == outputs %d ( %d + %d + %d )", inputs_sum, (total_amount_required + change_amount + tx.RctSignature.Get_TX_Fee()), total_amount_required, change_amount, tx.RctSignature.Get_TX_Fee())
		if inputs_sum != (total_amount_required + change_amount + tx.RctSignature.Get_TX_Fee()) {
			rlog.Warnf("INPUTS != OUTPUTS, please check")
			panic(fmt.Sprintf("Inputs %d != outputs ( %d + %d + %d )", inputs_sum, total_amount_required, change_amount, tx.RctSignature.Get_TX_Fee()))
		}
	}

	return
}

// send all unlocked balance amount to specific address
func (w *Wallet) Transfer_Everything(addr address.Address, payment_id_hex string, unlock_time uint64, fees_per_kb uint64, mixin uint64) (tx *transaction.Transaction, inputs_selected []uint64, inputs_sum uint64, err error) {

    
        var  transfer_details structures.Outgoing_Transfer_Details
	
	
	w.transfer_mutex.Lock()
	defer w.transfer_mutex.Unlock()

	if mixin < 5 { // enforce minimum mixin
		mixin = 5
	}

	// if wallet is online,take the fees from the network itself
	// otherwise use whatever user has provided
	//if w.GetMode()  {
	fees_per_kb = w.dynamic_fees_per_kb // TODO disabled as protection while lots more testing is going on
	rlog.Infof("Fees per KB %d\n", fees_per_kb)
	//}

	if fees_per_kb == 0 { // hard coded at compile time
		fees_per_kb = config.FEE_PER_KB
	}

	var txw *TX_Wallet_Data

	var payment_id []byte // we later on find WHETHER to include it, encrypt it depending on length

	// if payment  ID is provided explicity, use it
	if payment_id_hex != "" {
		payment_id, err = hex.DecodeString(payment_id_hex) // payment_id in hex
		if err != nil {
			return
		}

		if len(payment_id) == 32 || len(payment_id) == 8 {

		} else {
			err = fmt.Errorf("Payment ID must be atleast 64 hex chars (32 bytes) or 16 hex chars 8 byte")
			return
		}

	}

	// only only single payment id
	if addr.IsIntegratedAddress() && payment_id_hex != "" {
		err = fmt.Errorf("Payment ID provided in both integrated address and separately")
		return
	}
	// if integrated address payment id present , normal payment id must not be provided
	if addr.IsIntegratedAddress() {
		payment_id = addr.PaymentID
	}

	fees := uint64(0) // start with zero fees
	expected_fee := uint64(0)

	// infinite tries to build a transaction
	for {

		// now we need to select all outputs with sufficient balance
		inputs_selected, inputs_sum = w.select_outputs_for_transfer(0, fees+expected_fee, true)

		if len(inputs_selected) < 1 {
			err = fmt.Errorf("Insufficent unlocked balance")
			return
		}

		rlog.Infof("Selected %d (%+v) iinputs to transfer  %s DERO\n", len(inputs_selected), inputs_selected, globals.FormatMoney(inputs_sum))

		// lets prepare the inputs for ringct, so as we can used them
		var inputs []ringct.Input_info
		for i := range inputs_selected {

			txw, err = w.load_funds_data(inputs_selected[i], FUNDS_BUCKET)
			if err != nil {
				err = fmt.Errorf("Error while reading available funds index( it was just selected ) index %d err %s", inputs_selected[i], err)
				return
			}

			rlog.Infof("current input  %d %d \n", i, inputs_selected[i])
			var current_input ringct.Input_info
			current_input.Amount = txw.WAmount
			current_input.Key_image = crypto.Hash(txw.WKimage)
			current_input.Sk = txw.WKey

			//current_input.Index = i   is calculated after sorting of ring members
			current_input.Index_Global = txw.TXdata.Index_Global

			// add ring members here

			// ring_size = 4
			// TODO force random ring members

			//  mandatory add ourselves as ring member, otherwise there is no point in building the tx
			current_input.Ring_Members = append(current_input.Ring_Members, current_input.Index_Global)
			current_input.Pubs = append(current_input.Pubs, txw.TXdata.InKey)

			// add necessary amount  of random ring members
			// TODO we need to make sure ring members are mature, otherwise tx will fail because o immature inputs
			// This can cause certain TX to fail
			for {

				var buf [8]byte
				cryptorand.Read(buf[:])
				r, err := w.load_ring_member(binary.LittleEndian.Uint64(buf[:]) % w.account.Index_Global)
				if err == nil {

					// make sure ring member are not repeated
					new_ring_member := true
					for j := range current_input.Ring_Members { // TODO we donot need the loop
						if r.Index_Global == current_input.Ring_Members[j] {
							new_ring_member = false // we should not use this ring member
						}
						// if ring member is not mature, choose another one
						if !inputmaturity.Is_Input_Mature(w.Get_Height(),
							r.Height,
							r.Unlock_Height,
							r.Sigtype) {
							new_ring_member = false // we should not use this ring member

						}
					}

					if !new_ring_member {
						continue
					}
					current_input.Ring_Members = append(current_input.Ring_Members, r.Index_Global)
					current_input.Pubs = append(current_input.Pubs, r.InKey)
				}
				if uint64(len(current_input.Ring_Members)) == mixin { // atleast 5 ring members
					break
				}
			}

			// sort ring members and setup index

			/*
			   if i == 0 {
			       current_input.Ring_Members = []uint64{2,3,739,1158}
			       current_input.Pubs = append(current_input.Pubs,user.Random_Outputs[2].TXdata.InKey)
			       current_input.Pubs = append(current_input.Pubs,user.Random_Outputs[3].TXdata.InKey)
			       current_input.Pubs = append(current_input.Pubs,user.Random_Outputs[739].TXdata.InKey)
			       current_input.Pubs = append(current_input.Pubs,user.Random_Outputs[1158].TXdata.InKey)
			       current_input.Index = 2
			   }
			   if i == 1 {
			       continue
			       current_input.Ring_Members = []uint64{2,3,752,1158}
			       current_input.Pubs = append(current_input.Pubs,user.Random_Outputs[2].TXdata.InKey)
			       current_input.Pubs = append(current_input.Pubs,user.Random_Outputs[3].TXdata.InKey)
			       current_input.Pubs = append(current_input.Pubs,user.Random_Outputs[752].TXdata.InKey)
			       current_input.Pubs = append(current_input.Pubs,user.Random_Outputs[1158].TXdata.InKey)
			       current_input.Index = 2

			   }
			*/
			/*   current_input.Ring_Members = []uint64{2,3,4,inputs_selected[i]}
			     current_input.Pubs = append(current_input.Pubs,user.Random_Outputs[2].TXdata.InKey)
			     current_input.Pubs = append(current_input.Pubs,user.Random_Outputs[3].TXdata.InKey)
			     current_input.Pubs = append(current_input.Pubs,user.Random_Outputs[4].TXdata.InKey)
			     current_input.Pubs = append(current_input.Pubs,user.Random_Outputs[inputs_selected[i]].TXdata.InKey)
			     current_input.Index = 3

			*/

			rlog.Infof(" current input before sorting %+v \n", current_input.Ring_Members)

			current_input = sort_ring_members(current_input)
			rlog.Infof(" current input after sorting  %+v \n", current_input.Ring_Members)
			inputs = append(inputs, current_input)
		}

		//  fees = uint64(0)
		// fees = uint64(66248560000)

		// fill in the outputs

		var outputs []ringct.Output_info

	rebuild_tx_with_correct_fee:
		outputs = outputs[:0]

		var output ringct.Output_info
		output.Amount = inputs_sum - fees
		output.Public_Spend_Key = addr.SpendKey
		output.Public_View_Key = addr.ViewKey
		
		
		transfer_details.Fees = fees
		transfer_details.Amount = transfer_details.Amount[:0]
		transfer_details.Daddress = transfer_details.Daddress[:0]

		
		transfer_details.Amount = append(transfer_details.Amount,output.Amount) 
                transfer_details.Daddress = append(transfer_details.Daddress,addr.String()) 

                    

		outputs = append(outputs, output)

		// outputs = append(outputs, change)
		tx = w.Create_TX_v2(inputs, outputs, fees, unlock_time, payment_id, true)

		tx_size := uint64(len(tx.Serialize()))
		size_in_kb := tx_size / 1024

		if (tx_size % 1024) != 0 { // for any part there of, use a full KB fee
			size_in_kb += 1
		}

		minimum_fee := size_in_kb * fees_per_kb

		needed_fee := w.getfees(minimum_fee) // multiply minimum fees by multiplier

		rlog.Infof("required fees %s provided fee %s size %d fee/kb %s\n", globals.FormatMoney(needed_fee), globals.FormatMoney(fees), size_in_kb, globals.FormatMoney(fees_per_kb))

		if inputs_sum <= fees {
			err = fmt.Errorf("Insufficent unlocked balance to cover fees")
			return
		}

		if fees > needed_fee { // transaction was built up successfully
			fees = needed_fee // setup fees parameter exactly as much required
			goto rebuild_tx_with_correct_fee
		}

		// keep trying until we are successfull or funds become Insufficent
		if fees == needed_fee { // transaction was built up successfully
			break
		}

		// we need to try again
		fees = needed_fee               // setup estimated parameter
		expected_fee = expected_fee * 2 // double the estimated fee

	}
	
	
	
	// log enough information to wallet to display it again to users
	transfer_details.PaymentID =  hex.EncodeToString(payment_id) 
        
        // get the tx secret key and store it
        txhash := tx.GetHash()
        transfer_details.TXsecretkey = w.GetTXKey(tx.GetHash())
        transfer_details.TXID = txhash.String()
        
        // lets marshal the structure and store it in in DB
        
        details_serialized, err := json.Marshal(transfer_details)
	if err != nil {
                rlog.Warnf("Err marshalling details err %s", err) 
	}
        
        w.store_key_value(BLOCKCHAIN_UNIVERSE, []byte(TX_OUT_DETAILS_BUCKET), txhash[:], details_serialized[:])
        
       // fmt.Printf("%+v\n",transfer_details)
      //  fmt.Printf("%+v\n",transfer_details,w.GetTXOutDetails(tx.GetHash()))

        

	return
}

type member struct {
	index uint64
	key   ringct.CtKey
}

type members []member

func (s members) Len() int {
	return len(s)
}
func (s members) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s members) Less(i, j int) bool {
	return s[i].index < s[j].index
}

// sort ring members
func sort_ring_members(input ringct.Input_info) ringct.Input_info {

	if len(input.Ring_Members) != len(input.Pubs) {
		panic(fmt.Sprintf("Internal error !!!, ring member count %d != pubs count %d", len(input.Ring_Members), len(input.Pubs)))
	}
	var data_set members
	for i := range input.Pubs {
		data_set = append(data_set, member{input.Ring_Members[i], input.Pubs[i]})
	}
	sort.Sort(data_set)

	for i := range input.Pubs {
		input.Ring_Members[i] = data_set[i].index
		input.Pubs[i] = data_set[i].key
		if data_set[i].index == input.Index_Global {
			input.Index = i
		}
	}

	return input

}

func (w *Wallet) select_outputs_for_transfer(needed_amount uint64, fees uint64, all bool) (selected_output_index []uint64, sum uint64) {

	// return []uint64{739,752}, 6000000000000
	// return []uint64{4184}, user.Outputs_Ready[4184].WAmount

	index_list := w.load_all_values_from_bucket(BLOCKCHAIN_UNIVERSE, []byte(FUNDS_AVAILABLE))

	// shuffle the index_list
	// see https://stackoverflow.com/questions/12264789/shuffle-array-in-go
	for i := len(index_list) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		index_list[i], index_list[j] = index_list[j], index_list[i]
	}

	for i := range index_list { // load index
		current_index := binary.BigEndian.Uint64(index_list[i])

		tx, err := w.load_funds_data(current_index, FUNDS_BUCKET)
		if err != nil {
			fmt.Printf("Error while reading available funds index index %d err %s", current_index, err)
			continue
		}

		if inputmaturity.Is_Input_Mature(w.Get_Height(),
			tx.TXdata.Height,
			tx.TXdata.Unlock_Height,
			tx.TXdata.SigType) && !w.IsKeyImageSpent(tx.WKimage) {

			sum += tx.WAmount

			selected_output_index = append(selected_output_index, current_index) // select this output

			if !all { // user requested all inputs
				if sum > (needed_amount + fees) {
					return
				}
			}
		}

	}

	return

}

// load funds data structure from DB
func (w *Wallet) load_funds_data(index uint64, bucket string) (tx_wallet *TX_Wallet_Data, err error) {

	value_bytes, err := w.load_key_value(BLOCKCHAIN_UNIVERSE, []byte(bucket), itob(index))
	if err != nil {
		err = fmt.Errorf("Error while reading available funds index index %d err %s", index, err)
		return
	}

	tx_wallet = &TX_Wallet_Data{}
	err = msgpack.Unmarshal(value_bytes, &tx_wallet)
	if err != nil {
		err = fmt.Errorf("Error while decoding availble funds data index %d err %s", index, err)
		tx_wallet = nil
		return
	}
	return // everything was success
}

// this will create ringct simple 2 transaction to transfer x amount
func (w *Wallet) Create_TX_v2(inputs []ringct.Input_info, outputs []ringct.Output_info, fees uint64, unlock_time uint64, payment_id []byte, bulletproof bool) (txout *transaction.Transaction) {
	var tx transaction.Transaction
	tx.Version = 2
	tx.Unlock_Time = unlock_time // for the first input

	// setup the vins as they should be , setup key image
	for i := range inputs {

		/* for j := range inputs[i].Pubs {
		    fmt.Printf("dest key %d %s\n",j,inputs[i].Pubs[j].Destination )
		    fmt.Printf("dest key %d %s\n",j,inputs[i].Pubs[j].Mask )
		}*/

		//fmt.Printf("Key image %d %s\n",i, inputs[i].Key_image )
		txin := transaction.Txin_to_key{Amount: 0, K_image: inputs[i].Key_image} //amount is always zero in ringct and later

		if len(inputs[i].Ring_Members) != len(inputs[i].Pubs) {
			panic(fmt.Sprintf("Ring members and public keys should be equal %d %d", len(inputs[i].Ring_Members), len(inputs[i].Pubs)))
		}
		// fill in the ring members coded as offsets
		last_member := uint64(0)
		for j := range inputs[i].Ring_Members {
			current_offset := inputs[i].Ring_Members[j] - last_member
			last_member = inputs[i].Ring_Members[j]
			txin.Key_offsets = append(txin.Key_offsets, current_offset)
		}

		//fmt.Printf("Offsets %d = %+v\n",i, txin.Key_offsets)
		tx.Vin = append(tx.Vin, txin)
	}

	// input setup is completed, now we need to setup outputs

	// generate transaction wide unique key
	tx_secret_key, tx_public_key := crypto.NewKeyPair() // create new tx key pair

	/*
	   // these 3 lines are temporary
	    crypto.Sc_0(tx_secret_key);  // temporary for debugging puprpose, make it zero
	    crypto.ScReduce32(tx_secret_key) // reduce it for crypto purpose
	    tx_public_key = tx_secret_key.PublicKey()
	*/

	//tx_public_key is added to extra and serialized
	tx.Extra_map = map[transaction.EXTRA_TAG]interface{}{}
	tx.Extra_map[transaction.TX_PUBLIC_KEY] = *tx_public_key
	tx.PaymentID_map = map[transaction.EXTRA_TAG]interface{}{}

	if len(payment_id) == 32 {
		tx.PaymentID_map[transaction.TX_EXTRA_NONCE_PAYMENT_ID] = payment_id
	}

	tx.Extra = tx.Serialize_Extra() // serialize the extra

	for i := range outputs {

		//fmt.Printf("%d amount %d\n",i,outputs[i].Amount )
		derivation := crypto.KeyDerivation(&outputs[i].Public_View_Key, tx_secret_key) // keyderivation using output address view key

		// payment id if encrypted are encrypted against first receipient
		if i == 0 { // encrypt it now for the first output
			if len(payment_id) == 8 { // it is an encrypted payment ID,
				tx.PaymentID_map[transaction.TX_EXTRA_NONCE_ENCRYPTED_PAYMENT_ID] = EncryptDecryptPaymentID(derivation, *tx_public_key, payment_id)
			}
			tx.Extra = tx.Serialize_Extra() // serialize the extra

		}

		// this becomes the key within Vout
		index_within_tx := i
		ehphermal_public_key := derivation.KeyDerivation_To_PublicKey(uint64(index_within_tx), outputs[i].Public_Spend_Key)

		// added the amount and key in vout
		tx.Vout = append(tx.Vout, transaction.Tx_out{Amount: 0, Target: transaction.Txout_to_key{Key: ehphermal_public_key}})

		// setup key so as output amount can be encrypted, this will be passed later on to ringct package to encrypt amount
		outputs[i].Scalar_Key = *(derivation.KeyDerivationToScalar(uint64(index_within_tx)))
		// outputs[i].Destination = ehphermal_public_key

	}

	// now comes the ringct part, we always generate rinct simple, they are a bit larger (~1KB) if only single input is used
	// but that is okay as soon we will migrate to bulletproof
	tx.RctSignature = &ringct.RctSig{} // we always generate ringct simple

	// fmt.Printf("txprefix hash %s\n",tx.GetPrefixHash() )

	if bulletproof {
		tx.RctSignature.Gen_RingCT_Simple_BulletProof(tx.GetPrefixHash(), inputs, outputs, fees)
	} else {
		tx.RctSignature.Gen_RingCT_Simple(tx.GetPrefixHash(), inputs, outputs, fees)
	}

	// store the tx key to db, always, since we will never no since the tx may be sent offline
	txhash := tx.GetHash()
	w.store_key_value(BLOCKCHAIN_UNIVERSE, []byte(SECRET_KEY_BUCKET), txhash[:], tx_secret_key[:])

	return &tx
}

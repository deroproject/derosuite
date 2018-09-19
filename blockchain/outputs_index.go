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

package blockchain

// NOTE: this is extremely critical code ( as a single error or typo here will lead to invalid transactions )
//
// thhis file implements code which controls output indexes
// rewrites them during chain reorganisation
import "fmt"

//import "os"
//import "io/ioutil"
//import "sync"
//import "encoding/binary"

import "github.com/romana/rlog"
import "github.com/vmihailenco/msgpack"

import "github.com/deroproject/derosuite/config"
import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/storage"
import "github.com/deroproject/derosuite/crypto/ringct"
import "github.com/deroproject/derosuite/transaction"

//import "github.com/deroproject/derosuite/walletapi"

type Index_Data struct {
	InKey     ringct.CtKey
	ECDHTuple ringct.ECdhTuple // encrypted Amounts
	// Key crypto.Hash  // stealth address key
	// Commitment crypto.Hash // commitment public key
	Height        uint64 // height to which this belongs
	Unlock_Height uint64 // height at which it will unlock
}

/*
func (o *Index_Data) Serialize() (result []byte) {
	result = append(o.InKey.Destination[:], o.InKey.Mask[:]...)
        result = append(result, o.ECDHTuple.Mask[:]...)
        result = append(result, o.ECDHTuple.Amount[:]...)
       result = append(result, itob(o.Height)...)
	return
}

func (o *Index_Data) Deserialize(buf []byte) (err error) {
        if len(buf) != ( 32 + 32 + 32+ 32+8){
            return fmt.Errorf("Output index needs to be 72 bytes in size but found to be %d bytes", len(buf))
        }
        copy(o.InKey.Destination[:],buf[:32])
        copy(o.InKey.Mask[:],buf[32:64])
        copy(o.ECDHTuple.Mask[:],buf[64:96])
        copy(o.ECDHTuple.Amount[:],buf[96:128])
        o.Height = binary.BigEndian.Uint64(buf[64:])
	return
}
*/

//var account walletapi.Account

/*
func init() {

    var err error
    account , err =  wallet.Generate_Account_From_Recovery_Words("PLACE RECOVERY SEED here to test tx evaluation from within daemon")

    if err != nil {
        fmt.Printf("err %s\n",err)
        return
    }

    fmt.Printf("%+v\n", account)
}
*/

// this function writes or overwrites the data related to outputs
// the following data is collected from each output
// the secret key,
// the commitment  ( for miner tx the commitment is created from scratch
// 8 bytes blockheight to which this output belongs
// this function should always succeed or panic showing something is not correct
// NOTE: this function should only be called after all the tx and the block has been stored to DB
func (chain *Blockchain) write_output_index(dbtx storage.DBTX, block_id crypto.Hash, index_start int64, hard_fork_version_current int64) (result bool) {

	// load the block
	bl, err := chain.Load_BL_FROM_ID(dbtx, block_id)
	if err != nil {
		logger.Warnf("No such block %s for writing output index", block_id)
		return
	}

	//index_start := chain.Get_Block_Output_Index(dbtx,block_id) // get index position
	// load topo height
	height := chain.Load_Height_for_BL_ID(dbtx, block_id)

	// this was for quick tetsing of wallet
	//index_start =  uint64(height) //

	rlog.Debugf("Writing Output Index for block %s height %d output index %d", block_id, height, index_start)

	dbtx.StoreUint64(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, block_id[:], PLANET_OUTPUT_INDEX, uint64(index_start))

	// ads miner tx separately as a special case
	var o globals.TX_Output_Data
	var d Index_Data

	// extract key and commitment mask from for miner tx
	//	d.InKey.Destination = ringct.Key(bl.Miner_TX.Vout[0].Target.(transaction.Txout_to_key).Key)

	// mask can be calculated for miner tx on the wallet side as below
	//	d.InKey.Mask = ringct.ZeroCommitment_From_Amount(bl.Miner_TX.Vout[0].Amount)
	//	d.Height = uint64(height)
	//	d.Unlock_Height = uint64(height) + config.MINER_TX_AMOUNT_UNLOCK

	minertx_reward, err := dbtx.LoadUint64(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, block_id[:], PLANET_MINERTX_REWARD)
	if err != nil {
		logger.Fatalf("Base Reward is not stored in DB block %s", block_id)

		return
	}
	o.BLID = block_id // store block id
	o.TXID = bl.Miner_TX.GetHash()
	o.InKey.Destination = crypto.Key(bl.Miner_TX.Vout[0].Target.(transaction.Txout_to_key).Key)

	// FIXME miner tx amount should be what we calculated
	//	o.InKey.Mask = ringct.ZeroCommitment_From_Amount(bl.Miner_TX.Vout[0].Amount)
	o.InKey.Mask = ringct.ZeroCommitment_From_Amount(minertx_reward)
	o.Height = uint64(height)
	o.Unlock_Height = 0 // miner tx cannot be locked
	o.Index_within_tx = 0
	o.Index_Global = uint64(index_start)
	//	o.Amount = bl.Miner_TX.Vout[0].Amount
	o.Amount = minertx_reward
	o.SigType = 0
	o.Block_Time = bl.Timestamp
	o.TopoHeight = chain.Load_Block_Topological_order(dbtx, block_id)

	//ECDHTuple & sender pk is not available for miner tx

	if bl.Miner_TX.Parse_Extra() {

		// store public key if present
		if _, ok := bl.Miner_TX.Extra_map[transaction.TX_PUBLIC_KEY]; ok {
			o.Tx_Public_Key = bl.Miner_TX.Extra_map[transaction.TX_PUBLIC_KEY].(crypto.Key)
		}

		//o.Derivation_Public_Key_From_Vout = bl.Miner_tx.Vout[0].Target.(transaction.Txout_to_key).Key

	}

	serialized, err := msgpack.Marshal(&o)
	if err != nil {
		panic(err)
	}

	//fmt.Printf("index %d  %x\n",index_start,d.InKey.Destination)

	// store the index and relevant keys together in compact form
	dbtx.StoreObject(BLOCKCHAIN_UNIVERSE, GALAXY_OUTPUT_INDEX, GALAXY_OUTPUT_INDEX, itob(uint64(index_start)), serialized)

	index_start++

	// now loops through all the transactions, and store there ouutputs also
	// however as per client protocol, only process accepted transactions
	for i := 0; i < len(bl.Tx_hashes); i++ { // load all tx one by one

		if !chain.IS_TX_Valid(dbtx, block_id, bl.Tx_hashes[i]) { // skip invalid TX
			rlog.Tracef(1, "bl %s tx %s ignored while building outputs index as per client protocol", block_id, bl.Tx_hashes[i])
			continue
		}
		rlog.Tracef(1, "bl %s tx %s is being used while building outputs index as per client protocol", block_id, bl.Tx_hashes[i])

		tx, err := chain.Load_TX_FROM_ID(dbtx, bl.Tx_hashes[i])
		if err != nil {
			panic(fmt.Errorf("Cannot load  tx for %x err %s", bl.Tx_hashes[i], err))
		}

		//fmt.Printf("tx %s",bl.Tx_hashes[i])
		index_within_tx := uint64(0)

		o.BLID = block_id // store block id
		o.TXID = bl.Tx_hashes[i]
		o.Height = uint64(height)
		o.SigType = uint64(tx.RctSignature.Get_Sig_Type())

		// TODO unlock specific outputs on specific height
		o.Unlock_Height = uint64(height) + config.NORMAL_TX_AMOUNT_UNLOCK

		// build the key image list and pack it
		for j := 0; j < len(tx.Vin); j++ {
			k_image := crypto.Key(tx.Vin[j].(transaction.Txin_to_key).K_image)
			o.Key_Images = append(o.Key_Images, crypto.Key(k_image))
		}

		// zero out fields between tx
		o.Tx_Public_Key = crypto.Key(ZERO_HASH)
		o.PaymentID = o.PaymentID[:0]

		extra_parsed := tx.Parse_Extra()

		// tx has been loaded, now lets get the vout
		for j := uint64(0); j < uint64(len(tx.Vout)); j++ {

			//fmt.Printf("Processing vout %d\n", j)
			d.InKey.Destination = crypto.Key(tx.Vout[j].Target.(transaction.Txout_to_key).Key)
			d.InKey.Mask = crypto.Key(tx.RctSignature.OutPk[j].Mask)

			o.InKey.Destination = crypto.Key(tx.Vout[j].Target.(transaction.Txout_to_key).Key)
			o.InKey.Mask = crypto.Key(tx.RctSignature.OutPk[j].Mask)

			o.ECDHTuple = tx.RctSignature.ECdhInfo[j]

			o.Index_within_tx = index_within_tx
			o.Index_Global = uint64(index_start)
			o.Amount = tx.Vout[j].Amount
			o.Unlock_Height = 0

			if j == 0 && tx.Unlock_Time != 0 { // only first output of a TX can be locked
				o.Unlock_Height = tx.Unlock_Time
			}

			if hard_fork_version_current >= 3 && o.Unlock_Height  != 0 {
				if o.Unlock_Height < config.CRYPTONOTE_MAX_BLOCK_NUMBER {
					if o.Unlock_Height  < (o.Height + 1000) {
						o.Unlock_Height = o.Height + 1000
					}
				}else{
					if o.Unlock_Height < (o.Block_Time + 12000) {
						 o.Unlock_Height = o.Block_Time + 12000
					}
				}
			}

			// include the key image list in the first output itself
			// rest all the outputs donot contain the keyimage
			if j != 0 && len(o.Key_Images) > 0 {
				o.Key_Images = o.Key_Images[:0]
			}

			if extra_parsed {
				// store public key if present
				if _, ok := tx.Extra_map[transaction.TX_PUBLIC_KEY]; ok {
					o.Tx_Public_Key = tx.Extra_map[transaction.TX_PUBLIC_KEY].(crypto.Key)
				}

				// store payment IDs if present
				if _, ok := tx.PaymentID_map[transaction.TX_EXTRA_NONCE_ENCRYPTED_PAYMENT_ID]; ok {
					o.PaymentID = tx.PaymentID_map[transaction.TX_EXTRA_NONCE_ENCRYPTED_PAYMENT_ID].([]byte)
				} else if _, ok := tx.PaymentID_map[transaction.TX_EXTRA_NONCE_PAYMENT_ID]; ok {
					o.PaymentID = tx.PaymentID_map[transaction.TX_EXTRA_NONCE_PAYMENT_ID].([]byte)
				}

				/*   during emergency, for debugging purpose only
				     NOTE: remove this before rekeasing code

				                        if account.Is_Output_Ours(tx.Extra_map[transaction.TX_PUBLIC_KEY].(crypto.Key),index_within_tx, tx.Vout[j].Target.(transaction.Txout_to_key).Key){
				                            logger.Warnf("MG/simple Output is ours in tx %s at index %d height %d  global index %d",bl.Tx_hashes[i],index_within_tx,height, o.Index_Global)

				                            account.Decode_RingCT_Output(tx.Extra_map[transaction.TX_PUBLIC_KEY].(crypto.Key),
				                                                 j,
				                                                 crypto.Key(tx.RctSignature.OutPk[j].Mask),
				                                                 tx.RctSignature.ECdhInfo[j],
				                                                 2)
				                    }
				*/
			}

			serialized, err := msgpack.Marshal(&o)
			if err != nil {
				panic(err)
			}

			dbtx.StoreObject(BLOCKCHAIN_UNIVERSE, GALAXY_OUTPUT_INDEX, GALAXY_OUTPUT_INDEX, itob(uint64(index_start)), serialized)

			// fmt.Printf("index %d  %x\n",index_start,d.InKey.Destination)
			index_start++
			index_within_tx++
		}

	}

	// store where the look up ends
	dbtx.StoreUint64(BLOCKCHAIN_UNIVERSE, GALAXY_BLOCK, block_id[:], PLANET_OUTPUT_INDEX_END, uint64(index_start))

	return true
}

// this will load the index  data for specific index
// this should be done while holding the chain lock,
// since during reorganisation we might  give out wrong keys,
// to avoid that pitfall take the chain lock
// NOTE: this function is now for internal use only by the blockchain itself
//
func (chain *Blockchain) load_output_index(dbtx storage.DBTX, index uint64) (idata globals.TX_Output_Data, success bool) {
	// chain.Lock()
	// defer chain.Unlock()

	success = false
	data_bytes, err := dbtx.LoadObject(BLOCKCHAIN_UNIVERSE, GALAXY_OUTPUT_INDEX, GALAXY_OUTPUT_INDEX, itob(index))

	if err != nil {
		logger.Warnf("err while loading output index data index = %d err %s", index, err)
		success = false
		return
	}

	err = msgpack.Unmarshal(data_bytes, &idata)
	if err != nil {
		rlog.Warnf("err while unmarshallin output index data index = %d  data_len %d err %s", index, len(data_bytes), err)
		success = false
		return
	}

	success = true
	return
}

// this will read the output index data but will not deserialize it
// this is exposed for rpcserver giving access to wallet
func (chain *Blockchain) Read_output_index(dbtx storage.DBTX, index uint64) (data_bytes []byte, err error) {

	if dbtx == nil {
		dbtx, err = chain.store.BeginTX(false)
		if err != nil {
			rlog.Warnf("Error obtaining read-only tx. Error opening writable TX, err %s", err)
			return
		}

		defer dbtx.Rollback()

	}

	data_bytes, err = dbtx.LoadObject(BLOCKCHAIN_UNIVERSE, GALAXY_OUTPUT_INDEX, GALAXY_OUTPUT_INDEX, itob(index))

	if err != nil {
		rlog.Warnf("err while loading output index data index = %d err %s", index, err)
		return
	}
	return data_bytes, err
}

// this function finds output index for the tx
// first find a block index , and get the start offset
// then loop the index till you find the key in the result
// if something is not right, we return 0

func (chain *Blockchain) Find_TX_Output_Index(tx_hash crypto.Hash) (offset int64) {
	topo_Height := chain.Load_TX_Height(nil, tx_hash) // get height at which it's mined

	block_id, err := chain.Load_Block_Topological_order_at_index(nil, topo_Height)

	if err != nil {
		rlog.Warnf("error while finding tx_output_index %s", tx_hash)
		return 0
	}

	block_index_start, _ := chain.Get_Block_Output_Index(nil, block_id)

	// output_max_count := chain.Block_Count_Vout(block_id)  // this function will load/serdes all tx contained within block

	bl, err := chain.Load_BL_FROM_ID(nil, block_id)
	if err != nil {
		rlog.Warnf("Cannot load  block for %s err %s", block_id, err)
		return
	}

	if tx_hash == bl.Miner_TX.GetHash() { // miner tx is the beginning point
		return block_index_start
	}

	offset = block_index_start + 1 // shift by 1

	for i := 0; i < len(bl.Tx_hashes); i++ { // load all tx one by one

		// follow client protocol and skip some transactions
		if !chain.IS_TX_Valid(nil, block_id, bl.Tx_hashes[i]) { // skip invalid TX
			continue
		}

		if bl.Tx_hashes[i] == tx_hash {
			return offset
		}
		tx, err := chain.Load_TX_FROM_ID(nil, bl.Tx_hashes[i])
		if err != nil {
			rlog.Warnf("Cannot load  tx for %s err %s", bl.Tx_hashes[i], err)
		}

		// tx has been loaded, now lets get the vout
		vout_count := int64(len(tx.Vout))
		offset += vout_count
	}

	// we will reach here only if tx is linked to wrong block
	// this may be possible during reorganisation
	// return 0
	//logger.Warnf("Index Position must never reach here")
	return -1
}

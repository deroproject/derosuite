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

import "github.com/vmihailenco/msgpack"

import "github.com/deroproject/derosuite/config"
import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/crypto/ringct"
import "github.com/deroproject/derosuite/transaction"

import "github.com/deroproject/derosuite/walletapi"

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

var account walletapi.Account

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
func (chain *Blockchain) write_output_index(block_id crypto.Hash) {

	// load the block
	bl, err := chain.Load_BL_FROM_ID(block_id)
	if err != nil {
		logger.Warnf("No such block %s for writing output index", block_id)
		return
	}

	index_start := chain.Get_Block_Output_Index(block_id) // get index position
	height := chain.Load_Height_for_BL_ID(block_id)

	logger.Debugf("Writing Output Index for block %s height %d output index %d", block_id, height, index_start)

	// ads miner tx separately as a special case
	var o globals.TX_Output_Data
	var d Index_Data

	// extract key and commitment mask from for miner tx
	d.InKey.Destination = ringct.Key(bl.Miner_tx.Vout[0].Target.(transaction.Txout_to_key).Key)

	// mask can be calculated for miner tx on the wallet side as below
	d.InKey.Mask = ringct.ZeroCommitment_From_Amount(bl.Miner_tx.Vout[0].Amount)
	d.Height = height
	d.Unlock_Height = height + config.MINER_TX_AMOUNT_UNLOCK

	o.TXID = bl.Miner_tx.GetHash()
	o.InKey.Destination = ringct.Key(bl.Miner_tx.Vout[0].Target.(transaction.Txout_to_key).Key)
	o.InKey.Mask = ringct.ZeroCommitment_From_Amount(bl.Miner_tx.Vout[0].Amount)
	o.Height = height
	o.Unlock_Height = 0 // miner tx caannot be locked
	o.Index_within_tx = 0
	o.Index_Global = index_start
	o.Amount = bl.Miner_tx.Vout[0].Amount
	o.SigType = 0
	o.Block_Time = bl.Timestamp

	//ECDHTuple & sender pk is not available for miner tx

	if bl.Miner_tx.Parse_Extra() {

		o.Tx_Public_Key = bl.Miner_tx.Extra_map[transaction.TX_PUBLIC_KEY].(crypto.Key)
		//o.Derivation_Public_Key_From_Vout = bl.Miner_tx.Vout[0].Target.(transaction.Txout_to_key).Key
		/*
		   *   PRE-WALLET code, can be  used to track down bugs in wallet
		          if account.Is_Output_Ours(bl.Miner_tx.Extra_map[transaction.TX_PUBLIC_KEY].(crypto.Key),0, bl.Miner_tx.Vout[0].Target.(transaction.Txout_to_key).Key){
		              logger.Warnf("Miner Output is ours in tx %s height %d",bl.Miner_tx.GetHash(),height)
		          }

		*/
	}

	serialized, err := msgpack.Marshal(&o)
	if err != nil {
		panic(err)
	}

	//fmt.Printf("index %d  %x\n",index_start,d.InKey.Destination)

	// store the index and relevant keys together in compact form
	chain.store.StoreObject(BLOCKCHAIN_UNIVERSE, GALAXY_OUTPUT_INDEX, GALAXY_OUTPUT_INDEX, itob(index_start), serialized)

	index_start++

	// now loops through all the transactions, and store there ouutputs also

	for i := 0; i < len(bl.Tx_hashes); i++ { // load all tx one by one
		tx, err := chain.Load_TX_FROM_ID(bl.Tx_hashes[i])
		if err != nil {
			panic(fmt.Errorf("Cannot load  tx for %x err %s", bl.Tx_hashes[i], err))
		}

		//fmt.Printf("tx %s",bl.Tx_hashes[i])
		index_within_tx := uint64(0)

		o.TXID = bl.Tx_hashes[i]
		o.Height = height
		o.SigType = uint64(tx.RctSignature.Get_Sig_Type())

		// TODO unlock specific outputs on specific height
		o.Unlock_Height = height + config.NORMAL_TX_AMOUNT_UNLOCK

		// build the key image list and pack it
		for j := 0; j < len(tx.Vin); j++ {
			k_image := ringct.Key(tx.Vin[j].(transaction.Txin_to_key).K_image)
			o.Key_Images = append(o.Key_Images, crypto.Key(k_image))
		}

		extra_parsed := tx.Parse_Extra()

		// tx has been loaded, now lets get the vout
		for j := uint64(0); j < uint64(len(tx.Vout)); j++ {

			//fmt.Printf("Processing vout %d\n", j)
			d.InKey.Destination = ringct.Key(tx.Vout[j].Target.(transaction.Txout_to_key).Key)
			d.InKey.Mask = ringct.Key(tx.RctSignature.OutPk[j].Mask)

			o.InKey.Destination = ringct.Key(tx.Vout[j].Target.(transaction.Txout_to_key).Key)
			o.InKey.Mask = ringct.Key(tx.RctSignature.OutPk[j].Mask)

			o.ECDHTuple = tx.RctSignature.ECdhInfo[j]

			o.Index_within_tx = index_within_tx
			o.Index_Global = index_start
			o.Amount = tx.Vout[j].Amount
			o.Unlock_Height = 0

			if j == 0 && tx.Unlock_Time != 0 { // only first output of a TX can be locked
				o.Unlock_Height = tx.Unlock_Time
			}

			// include the key image list in the first output itself
			// rest all the outputs donot contain the keyimage
			if j != 0 && len(o.Key_Images) > 0 {
				o.Key_Images = o.Key_Images[:0]
			}

			if extra_parsed {
				o.Tx_Public_Key = tx.Extra_map[transaction.TX_PUBLIC_KEY].(crypto.Key)
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
			chain.store.StoreObject(BLOCKCHAIN_UNIVERSE, GALAXY_OUTPUT_INDEX, GALAXY_OUTPUT_INDEX, itob(index_start), serialized)

			// fmt.Printf("index %d  %x\n",index_start,d.InKey.Destination)
			index_start++
			index_within_tx++
		}

	}

}

// this will load the index  data for specific index
// this should be done while holding the chain lock,
// since during reorganisation we might  give out wrong keys,
// to avoid that pitfall take the chain lock
// NOTE: this function is now for internal use only by the blockchain itself
//
func (chain *Blockchain) load_output_index(index uint64) (idata globals.TX_Output_Data) {
	// chain.Lock()
	// defer chain.Unlock()
	data_bytes, err := chain.store.LoadObject(BLOCKCHAIN_UNIVERSE, GALAXY_OUTPUT_INDEX, GALAXY_OUTPUT_INDEX, itob(index))

	if err != nil {
		logger.Warnf("err while loading output index data index = %d err %s", index, err)
		return
	}

	err = msgpack.Unmarshal(data_bytes, &idata)
	if err != nil {
		logger.Warnf("err while unmarshallin output index data index = %d  data_len %d err %s", index, len(data_bytes), err)
		return
	}

	return
}

// this will read the output index data but will not deserialize it
// this is exposed for rpcserver giving access to wallet
func (chain *Blockchain) Read_output_index(index uint64) (data_bytes []byte, err error) {
	chain.Lock()
	defer chain.Unlock()
	data_bytes, err = chain.store.LoadObject(BLOCKCHAIN_UNIVERSE, GALAXY_OUTPUT_INDEX, GALAXY_OUTPUT_INDEX, itob(index))

	if err != nil {
		logger.Warnf("err while loading output index data index = %d err %s", index, err)
		return
	}
	return data_bytes, err
}

// this function finds output index for the tx
// first find a block index , and get the start offset
// then loop the index till you find the key in the result
// if something is not right, we return 0
func (chain *Blockchain) Find_TX_Output_Index(tx_hash crypto.Hash) (offset uint64) {
	Block_Height := chain.Load_TX_Height(tx_hash) // get height

	block_id, err := chain.Load_BL_ID_at_Height(Block_Height)
	if err != nil {
		logger.Warnf("error while finding tx_output_index %s", tx_hash)
		return 0
	}

	block_index_start := chain.Get_Block_Output_Index(block_id)
	// output_max_count := chain.Block_Count_Vout(block_id)  // this function will load/serdes all tx contained within block
	/*for index_start:= block_index_start; index_start < (block_index_start+output_max_count); index_start++{


	  }
	*/

	bl, err := chain.Load_BL_FROM_ID(block_id)

	if err != nil {
		logger.Warnf("Cannot load  block for %s err %s", block_id, err)
		return
	}
	if tx_hash == bl.Miner_tx.GetHash() {
		return block_index_start
	}

	offset = block_index_start + 1 // shift by 1

	for i := 0; i < len(bl.Tx_hashes); i++ { // load all tx one by one

		if bl.Tx_hashes[i] == tx_hash {
			return offset
		}
		tx, err := chain.Load_TX_FROM_ID(bl.Tx_hashes[i])
		if err != nil {
			logger.Warnf("Cannot load  tx for %s err %s", bl.Tx_hashes[i], err)
		}

		// tx has been loaded, now lets get the vout
		vout_count := uint64(len(tx.Vout))
		offset += vout_count
	}

	// we will reach here only if tx is linked to wrong block
	// this may be possible during reorganisation
	// return 0
	logger.Warnf("Index Position must never reach here")
	return 0
}

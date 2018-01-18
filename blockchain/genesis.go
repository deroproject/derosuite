package blockchain


import "encoding/hex"

import "github.com/romana/rlog"

import "github.com/deroproject/derosuite/address"
import "github.com/deroproject/derosuite/globals"



func Create_Miner_Transaction(height uint64, median_size uint64, already_generated_coins uint64,
	current_block_size uint64, fee uint64,
	miner_address address.Address, nonce []byte,
	max_outs uint64, hard_fork uint64) (tx *Transaction, err error) {

	return nil, nil
}

// genesis transaction  hash  5a18d9489bcd353aeaf4a19323d04e90353f98f0d7cc2a030cfd76e19495547d
// genesis amount 35184372088831
func Generate_Genesis_Block() (bl Block) {

	genesis_tx_blob, err := hex.DecodeString(globals.Config.Genesis_Tx)
	if err != nil {
		panic("Failed to hex decode genesis Tx")
	}
	err = bl.Miner_tx.DeserializeHeader(genesis_tx_blob)

	if err != nil {
		panic("Failed to parse genesis tx ")
	}
	rlog.Tracef(2, "Hash of Genesis Tx %x\n", bl.Miner_tx.GetHash())

	// verify whether tx is coinbase and valid

	// setup genesis block header
	bl.Major_Version = 1
	bl.Minor_Version = 0
	bl.Timestamp = 0 // first block timestamp
	//bl.Prev_hash is automatic zero
	bl.Nonce = globals.Config.Genesis_Nonce

	rlog.Tracef(2, "Hash of genesis block is %x", bl.GetHash())

	rlog.Tracef(2, "Genesis Block PoW %x\n", bl.GetPoWHash())

	return
}

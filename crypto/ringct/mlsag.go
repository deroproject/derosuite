package ringct

import "fmt"
import "github.com/deroproject/derosuite/crypto"

/* This file implements MLSAG signatures for the transactions */

// get the hash of the transaction which is used to create the mlsag later on, this hash is input to MLSAG 
// the hash is = hash( message + hash(basehash) + hash(pederson and borromean data))
func Get_pre_mlsag_hash(sig *RctSig)  (crypto.Hash) {
    
    message_hash := sig.Message
    base_hash := crypto.Keccak256(sig.SerializeBase())
    
    fmt.Printf("Message hash %x\n", message_hash)
    fmt.Printf("Base hash %x\n", base_hash)
    
    // now join the borromean signature and extract a sig
    var other_data []byte
    for i := range sig.rangeSigs {
        other_data = append(other_data,sig.rangeSigs[i].asig.s0.Serialize()...)
        other_data = append(other_data,sig.rangeSigs[i].asig.s1.Serialize()...)
        other_data = append(other_data, sig.rangeSigs[i].ci.Serialize()...)
    }
    other_data_hash := crypto.Keccak256(other_data)
    
    fmt.Printf("other  hash %x\n", other_data_hash)
    
    
    // join all 3 hashes and hash them again to get the data
    final_data := append(message_hash[:], base_hash[:]...)
    final_data = append(final_data, other_data_hash[:]...)
    final_data_hash :=  crypto.Keccak256(final_data)
    
    fmt.Printf("final_data_hash  hash %x\n", other_data_hash)
    
    
    return final_data_hash
}

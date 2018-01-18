package blockchain


// thhis file implements code which controls output indexes
// rewrites them during chain reorganisation
import "fmt"
import "encoding/binary"

import "github.com/deroproject/derosuite/crypto"
//import "github.com/deroproject/derosuite/crypto/ringct"


type Output_index struct {
    Key crypto.Hash  // stealth address key
    Commitment crypto.Hash // commitment public key
    Height  uint64  // height to which this belongs
}


func (o *Output_index) Serialize() (result []byte) {
	result = append(o.Key[:], o.Commitment[:]...)
        result = append(result, itob(o.Height)...)
	return
}

func (o *Output_index) Deserialize(buf []byte) (err error) {
        if len(buf) != ( 32 + 32 + 8){
            return fmt.Errorf("Output index needs to be 72 bytes in size but found to be %d bytes", len(buf))
        }
        copy(o.Key[:],buf[:32])
        copy(o.Commitment[:],buf[32:64])
        o.Height = binary.BigEndian.Uint64(buf[64:])
	return
}

// this function writes or overwrites the data related to outputs
// the following data is collected from each output
// the secret key,
// the commitment  ( for miner tx the commitment is created from scratch
// 8 bytes blockheight to which this output belongs
// this function should always succeed or panic showing something is not correct
// NOTE: this function should only be called after all the tx and the block has been stored to DB
func (chain *Blockchain)write_output_index(block_id  crypto.Hash){
    
    // load the block
    bl, err := chain.Load_BL_FROM_ID(block_id)
    if err != nil {
        panic(fmt.Sprintf("No such block %x\n", block_id))        
    }
    _ = bl
    
    
    
    
    
}

package blockchain

import "fmt"
import "encoding/hex"

import log "github.com/sirupsen/logrus"

import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/globals"


var mainnet_static_checkpoints = map[crypto.Hash]uint64{}
var testnet_static_checkpoints = map[crypto.Hash]uint64{}

// initialize the checkpoints
func init_checkpoints() {

	switch globals.Config.Name {
	case "mainnet":
		ADD_CHECKPOINT(mainnet_static_checkpoints, 1, "cea3eb82889a1d063aa6205011a27a108e727f2ac66721a46bec8b7ee9e83d7e")
		ADD_CHECKPOINT(mainnet_static_checkpoints, 10, "a8feb533ad2ad021f356b964957f4880445f89d6c6658a9407d63ce5144fe8ea")
		ADD_CHECKPOINT(mainnet_static_checkpoints, 100, "462ef1347bd00511ad6e7be463cba4d44a69dbf8b7d1d478ff1fd68507dfc9e2")

		log.Debugf("Added %d static checkpoints to mainnnet\n", len(mainnet_static_checkpoints))
	case "testnet":

	default:
		panic(fmt.Sprintf("Unknown Network \"%s\"", globals.Config.Name))
	}

}

// add a checkpoint to specific network
func ADD_CHECKPOINT(x map[crypto.Hash]uint64, Height uint64, hash_hex string) {

	var hash crypto.Hash
	hash_raw, err := hex.DecodeString(hash_hex)

	if err != nil {
		panic(fmt.Sprintf("Cannot hex decode checkpint hash \"%s\"", hash_hex))
	}

	if len(hash_raw) != 32 {
		panic(fmt.Sprintf(" hash not 32 byte size Cannot hex decode checkpint hash \"%s\"", hash_hex))
	}

	copy(hash[:], hash_raw)
	x[hash] = Height
}

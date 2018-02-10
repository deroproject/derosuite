package ringct

import "fmt"
import "github.com/deroproject/derosuite/crypto"

// this file has license pending  since it triggers a hard to find golang bug TODO add license after the golang bug is fixed
/* This file implements MLSAG signatures for the transactions */

// get the hash of the transaction which is used to create the mlsag later on, this hash is input to MLSAG
// the hash is = hash( message + hash(basehash) + hash(pederson and borromean data))
func Get_pre_mlsag_hash(sig *RctSig) crypto.Hash {

	message_hash := sig.Message
	base_hash := crypto.Keccak256(sig.SerializeBase())

	//fmt.Printf("Message hash %s\n", message_hash)
	//fmt.Printf("Base hash %s\n", base_hash)

	// now join the borromean signature and extract a sig
	var other_data []byte
	for i := range sig.rangeSigs {
		//other_data = append(other_data,sig.rangeSigs[i].asig.s0.Serialize()...)
		//other_data = append(other_data,sig.rangeSigs[i].asig.s1.Serialize()...)
		//other_data = append(other_data,sig.rangeSigs[i].asig.ee[:]...)
		//OR
		// other_data = append(other_data, sig.rangeSigs[i].asig.Serialize()...) // serialise borrosig
		//  other_data = append(other_data, sig.rangeSigs[i].ci.Serialize()...)
		// OR
		other_data = append(other_data, sig.rangeSigs[i].Serialize()...) // range sig serialise
	}
	other_data_hash := crypto.Keccak256(other_data)

	//fmt.Printf("other  hash %s\n", other_data_hash)

	// join all 3 hashes and hash them again to get the data
	final_data := append(message_hash[:], base_hash[:]...)
	final_data = append(final_data, other_data_hash[:]...)
	final_data_hash := crypto.Keccak256(final_data)

	if DEBUGGING_MODE {
		fmt.Printf("final_data_hash  hash %s\n", final_data_hash)
	}

	return final_data_hash
}

//Multilayered Spontaneous Anonymous Group Signatures (MLSAG signatures)
//This is a just slghtly more efficient version than the ones described below
//(will be explained in more detail in Ring Multisig paper
//These are aka MG signatutes in earlier drafts of the ring ct paper
// c.f. http://eprint.iacr.org/2015/1098 section 2.
// keyImageV just does I[i] = xx[i] * Hash(xx[i] * G) for each i
// Gen creates a signature which proves that for some column in the keymatrix "pk"
//   the signer knows a secret key for each row in that column
// Ver verifies that the MG sig was created correctly
func MLSAG_Ver(message Key, pk [][]Key, rv *MlsagSig, dsRows int, r *RctSig) (result bool) {
	result = false
	cols := len(pk)
	if cols < 2 {
		if DEBUGGING_MODE {
			fmt.Printf("RingCT MLSAG_Ver  must have cols > 1\n")
		}
		result = false
		return
	}

	rows := len(pk[0])
	if rows < 1 {
		if DEBUGGING_MODE {
			fmt.Printf("RingCT MLSAG_Ver  must have rows > 0\n")
		}
		result = false
		return
	}

	for i := 0; i < cols; i++ {
		if len(pk[i]) != rows {
			if DEBUGGING_MODE {
				fmt.Printf("RingCT MLSAG_Ver  pk matrix not rectangular\n")
			}
			result = false
			return
		}
	}

	if len(rv.II) != dsRows {
		if DEBUGGING_MODE {
			fmt.Printf("RingCT MLSAG_Ver Bad II size\n")
		}
		result = false
		return
	}

	if len(rv.ss) != cols {
		if DEBUGGING_MODE {
			fmt.Printf("RingCT MLSAG_Ver Bad rv.ss size   len(rv.ss) = %d cols = %d\n", len(rv.ss), cols)
		}
		result = false
		return
	}

	for i := 0; i < cols; i++ {
		if len(rv.ss[i]) != rows {
			if DEBUGGING_MODE {
				fmt.Printf("RingCT MLSAG_Ver rv.ss is not rectangular\n")
			}
			result = false
			return
		}
	}

	if dsRows > rows {
		if DEBUGGING_MODE {
			fmt.Printf("RingCT MLSAG_Ver Bad dsRows value\n")
		}
		result = false
		return
	}

	for i := 0; i < len(rv.ss); i++ {
		for j := 0; j < len(rv.ss[i]); j++ {
			if !ScValid(&rv.ss[i][j]) {
				if DEBUGGING_MODE {
					fmt.Printf("RingCT MLSAG_Ver Bad ss slot\n")
				}
				result = false
				return
			}
		}
	}

	if !ScValid(&rv.cc) {
		if DEBUGGING_MODE {
			fmt.Printf("RingCT MLSAG_Ver Bad r.cc slot\n")
		}
		result = false
		return
	}

	Ip := make([][8]CachedGroupElement, dsRows, dsRows) // do pre computation of key keyImage
	for i := 0; i < dsRows; i++ {
		key_image_point := new(ExtendedGroupElement)
		key_image_point.FromBytes(&rv.II[i])
		GePrecompute(&Ip[i], key_image_point)
	}

	ndsRows := 3 * dsRows //non Double Spendable Rows (see identity chains paper
	toHash := make([]Key, 1+3*dsRows+2*(rows-dsRows), 1+3*dsRows+2*(rows-dsRows))
	toHash[0] = message

	// golang does NOT allow to use casts without using unsafe, so we can be slow but safe
	toHash_bytes := make([]byte, 0, (1+3*dsRows+2*(rows-dsRows))*len(message))

	var c Key
	c_old := rv.cc
	for i := 0; i < cols; i++ {
		Sc_0(&c) // zero out c

		// BUG BUG BUG
		// DERO project has found a go bug
		// This bug affects all known golang supported platforms and architectures (arm/386/amd64/...?)
		// even the golang 1.10 release candidates are affected ( though tip has not been tested )
		//
		// if you  comment the line below and uncomment similiar line in first loop, the bug triggers for 1 in 20000 RCT full transactions, this means the the entire block chain cannot be verified before the bug fails the crypto
		// this bug triggers at high probability when the affected code is executed by multiple  goroutines simultaneously
		// since the bug is notoriously hard to trigger  while execution
		// we have figured out an easy way to visibly demonstrate that the bug is present
		//  if the variables are declared within first loop,  second loop is able to access them without declaring them
		//  this causes random memory to be used during CPU load, causing the transaction to fail since crypto checksums mark it as failure ( TODO detailed analysis)

		// this bug may be the source of several other random crash bugs and need to be given a detailed looked
		// this may be  the source of the follown known but not-understood (cause not found) bugs
		// https://github.com/golang/go/issues/15658
		// https://github.com/golang/go/issues/20427
		// https://github.com/golang/go/issues/22988
		// https://github.com/golang/go/issues/20300 and even more
		// NOTE: for golang developers, this bug needs to studied and fixed correctly. Since, another bug seems to exists
		// which causes constants to flip ( yes consts ). However, we cannot be certain if its the same bug, once this gets quashed, we will test the other one too.
		var L, R, Hi Key // comment this line and uncomment similiar line in first loop to trigger BUG

		// first loop
		for j := 0; j < dsRows; j++ {
			//var L, R, Hi  Key // uncomment this line to trigger golang compiler BUG ( atleast on linux amd64)
			AddKeys2(&L, &rv.ss[i][j], &c_old, &pk[i][j])
			Hi = pk[i][j].HashToPoint()
			AddKeys3(&R, &rv.ss[i][j], &Hi, &c_old, &Ip[j])

			toHash[3*j+1] = pk[i][j]
			toHash[3*j+2] = L
			toHash[3*j+3] = R
		}

		//second loop
		for j, ii := dsRows, 0; j < rows; j, ii = j+1, ii+1 {
			AddKeys2(&L, &rv.ss[i][j], &c_old, &pk[i][j]) // here L is getting used again
			toHash[ndsRows+2*ii+1] = pk[i][j]
			toHash[ndsRows+2*ii+2] = L
		}

		toHash_bytes = toHash_bytes[:0] // zero out everything
		for k := range toHash {
			toHash_bytes = append(toHash_bytes, toHash[k][:]...)
		}

		c = *(HashToScalar(toHash_bytes)) // hash_to_scalar(toHash);
		copy(c_old[:], c[:])              // flipping the args here, will cause all transactions to become valid

	}

	if DEBUGGING_MODE {
		//fmt.Printf("c     %x\n",c)
		fmt.Printf("c_old %s\n", c_old)
		fmt.Printf("rv.ss %s\n", rv.cc)
	}

	// c = c_old-rv.cc
	ScSub(&c, &c_old, &rv.cc)

	// if 0 checksum verified, otherwise checksum failed
	result = ScIsZero(&c)

	if DEBUGGING_MODE {

		if result {
			fmt.Printf("RingCT MLSAG_Ver Success\n")
		} else {
			fmt.Printf("RingCT MLSAG_Ver  verification failed\n")
		}
	}

	return

}

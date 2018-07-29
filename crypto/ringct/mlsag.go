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

	if sig.sigType == RCTTypeSimpleBulletproof || sig.sigType == RCTTypeFullBulletproof {
		for i := range sig.BulletSigs {
			//for j := range sig.BulletSigs[i].V{
			//	other_data= append(other_data,sig.BulletSigs[i].V[j][:]...)
			//}
			other_data = append(other_data, sig.BulletSigs[i].A[:]...)
			other_data = append(other_data, sig.BulletSigs[i].S[:]...)
			other_data = append(other_data, sig.BulletSigs[i].T1[:]...)
			other_data = append(other_data, sig.BulletSigs[i].T2[:]...)
			other_data = append(other_data, sig.BulletSigs[i].taux[:]...)
			other_data = append(other_data, sig.BulletSigs[i].mu[:]...)
			for j := range sig.BulletSigs[i].L {
				other_data = append(other_data, sig.BulletSigs[i].L[j][:]...)
			}
			for j := range sig.BulletSigs[i].R {
				other_data = append(other_data, sig.BulletSigs[i].R[j][:]...)
			}
			other_data = append(other_data, sig.BulletSigs[i].a[:]...)
			other_data = append(other_data, sig.BulletSigs[i].b[:]...)
			other_data = append(other_data, sig.BulletSigs[i].t[:]...)
		}

	} else if sig.sigType == RCTTypeSimple || sig.sigType == RCTTypeFull {
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
func MLSAG_Ver(message crypto.Key, pk [][]crypto.Key, rv *MlsagSig, dsRows int, r *RctSig) (result bool) {

	defer func() { // safety so if anything wrong happens, verification fails
		if r := recover(); r != nil {
			result = false
		}
	}()

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
			if !crypto.ScValid(&rv.ss[i][j]) {
				if DEBUGGING_MODE {
					fmt.Printf("RingCT MLSAG_Ver Bad ss slot\n")
				}
				result = false
				return
			}
		}
	}

	if !crypto.ScValid(&rv.cc) {
		if DEBUGGING_MODE {
			fmt.Printf("RingCT MLSAG_Ver Bad r.cc slot\n")
		}
		result = false
		return
	}

	//fmt.Printf("cc ver %s\n",rv.cc)

	Ip := make([][8]crypto.CachedGroupElement, dsRows, dsRows) // do pre computation of key keyImage
	for i := 0; i < dsRows; i++ {
		key_image_point := new(crypto.ExtendedGroupElement)
		key_image_point.FromBytes(&rv.II[i])
		crypto.GePrecompute(&Ip[i], key_image_point)
	}

	ndsRows := 3 * dsRows //non Double Spendable Rows (see identity chains paper
	toHash := make([]crypto.Key, 1+3*dsRows+2*(rows-dsRows), 1+3*dsRows+2*(rows-dsRows))
	toHash[0] = message

	// golang does NOT allow to use casts without using unsafe, so we can be slow but safe
	toHash_bytes := make([]byte, 0, (1+3*dsRows+2*(rows-dsRows))*len(message))

	var c crypto.Key
	c_old := rv.cc
	for i := 0; i < cols; i++ {
		crypto.Sc_0(&c) // zero out c

		var L, R, Hi crypto.Key

		// first loop
		for j := 0; j < dsRows; j++ {
			crypto.AddKeys2(&L, &rv.ss[i][j], &c_old, &pk[i][j])
			Hi = pk[i][j].HashToPoint()
			crypto.AddKeys3(&R, &rv.ss[i][j], &Hi, &c_old, &Ip[j])

			toHash[3*j+1] = pk[i][j]
			toHash[3*j+2] = L
			toHash[3*j+3] = R
		}

		//second loop
		for j, ii := dsRows, 0; j < rows; j, ii = j+1, ii+1 {
			crypto.AddKeys2(&L, &rv.ss[i][j], &c_old, &pk[i][j]) // here L is getting used again
			toHash[ndsRows+2*ii+1] = pk[i][j]
			toHash[ndsRows+2*ii+2] = L
		}

		toHash_bytes = toHash_bytes[:0] // zero out everything
		for k := range toHash {
			//fmt.Printf("vhash %d %s\n",k,toHash[k])
			toHash_bytes = append(toHash_bytes, toHash[k][:]...)
		}

		c = *(crypto.HashToScalar(toHash_bytes)) // hash_to_scalar(toHash);
		copy(c_old[:], c[:])                     // flipping the args here, will cause all transactions to become valid

	}

	if DEBUGGING_MODE {
		//fmt.Printf("c     %x\n",c)
		fmt.Printf("c_old %s\n", c_old)
		fmt.Printf("rv.ss %s\n", rv.cc)
	}

	// c = c_old-rv.cc
	crypto.ScSub(&c, &c_old, &rv.cc)

	// if 0 checksum verified, otherwise checksum failed
	result = crypto.ScIsZero(&c)

	if DEBUGGING_MODE {

		if result {
			fmt.Printf("RingCT MLSAG_Ver Success\n")
		} else {
			fmt.Printf("RingCT MLSAG_Ver  verification failed\n")
		}
	}

	return

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

func MLSAG_Gen(message crypto.Key, pk [][]crypto.Key, xx []crypto.Key, index int, dsRows int) (rv MlsagSig) {

	result := false
	_ = result
	cols := len(pk)
	if cols < 2 {
		if DEBUGGING_MODE {
			fmt.Printf("RingCT MLSAG_Gen  must have cols > 1\n")
		}
		result = false
		return
	}
	if index >= cols {
		panic("RingCT MLSAG_Gen Index out of range")
	}
	rows := len(pk[0])
	if rows < 1 {
		if DEBUGGING_MODE {
			fmt.Printf("RingCT MLSAG_Gen  must have rows > 0\n")
		}
		result = false
		return
	}

	for i := 0; i < cols; i++ {
		if len(pk[i]) != rows {
			if DEBUGGING_MODE {
				fmt.Printf("RingCT MLSAG_Gen pk is not rectangular\n")
			}
			result = false
			return
		}
	}

	if len(xx) != rows {
		if DEBUGGING_MODE {
			fmt.Printf("RingCT MLSAG_Gen Bad xx size\n")
		}
		result = false
		return

	}

	if dsRows > rows {
		if DEBUGGING_MODE {
			fmt.Printf("RingCT MLSAG_Gen Bad dsRows value\n")
		}
		result = false
		return
	}
	var i, j int
	var c, c_old, L, R, Hi crypto.Key

	crypto.Sc_0(&c_old)

	Ip := make([][8]crypto.CachedGroupElement, dsRows, dsRows) // do pre computation of key keyImage

	rv.II = make([]crypto.Key, dsRows, dsRows)
	alpha := make([]crypto.Key, rows, rows)
	aG := make([]crypto.Key, rows, rows)
	aHP := make([]crypto.Key, dsRows, dsRows)

	//rv.ss = keyM(cols, aG); // TODO

	// next 5 lines are quite common
	M := make([][]crypto.Key, cols)
	for i := 0; i < (cols); i++ {
		M[i] = make([]crypto.Key, rows+0, rows+0)
		for j := 0; j < (rows + 0); j++ { // yes there is an extra column
			M[i][j] = Identity // fill it with identity
		}
	}
	rv.ss = M

	toHash := make([]crypto.Key, 1+3*dsRows+2*(rows-dsRows), 1+3*dsRows+2*(rows-dsRows))
	toHash[0] = message

	// golang does NOT allow to use casts without using unsafe, so we can be slow but safe
	toHash_bytes := make([]byte, 0, (1+3*dsRows+2*(rows-dsRows))*len(message))

	for i := 0; i < dsRows; i++ {

		alpha[i] = crypto.SkGen()

		// Sc_0(&alpha[i]); // make random key zero  for tesing puprpose // BUG if line is uncommented

		//fmt.Printf("alpha[i] %s\n",alpha[i])
		// ScReduce32(&alpha[i]) // reduce it

		aG[i] = crypto.ScalarmultBase(alpha[i]) // *(alpha[i].PubKey()) //need to save alphas for later..
		//skpkGen(alpha[i], aG[i]);
		// Hi = hashToPoint(pk[index][i]);
		Hi = pk[index][i].HashToPoint()
		//aHP[i] = scalarmultKey(Hi, alpha[i]);
		aHP[i] = *crypto.ScalarMultKey(&Hi, &alpha[i])
		toHash[3*i+1] = pk[index][i]
		toHash[3*i+2] = aG[i]
		toHash[3*i+3] = aHP[i]
		rv.II[i] = *crypto.ScalarMultKey(&Hi, &xx[i])
		//precomp(Ip[i].k, rv.II[i]);
		//fmt.Printf("secret key %s\n", xx[i])

		key_image_point := new(crypto.ExtendedGroupElement)
		key_image_point.FromBytes(&rv.II[i])
		crypto.GePrecompute(&Ip[i], key_image_point)

	}

	ndsRows := 3 * dsRows //non Double Spendable Rows (see identity chains paper)
	for i, ii := dsRows, 0; i < rows; i, ii = i+1, ii+1 {
		//skpkGen(alpha[i], aG[i]); //need to save alphas for later..
		alpha[i] = crypto.SkGen()

		//Sc_0(&alpha[i]); // make random key zero  for tesing puprpose // BUG if line is uncommented
		//ScReduce32(&alpha[i]) // reduce it

		//aG[i] = *(alpha[i].PubKey()) //need to save alphas for later..
		aG[i] = *(alpha[i].PublicKey())

		toHash[ndsRows+2*ii+1] = pk[index][i]
		toHash[ndsRows+2*ii+2] = aG[i]
	}

	toHash_bytes = toHash_bytes[:0] // zero out everything
	for k := range toHash {
		//fmt.Printf("gen hash %d %s\n",k,toHash[k])
		toHash_bytes = append(toHash_bytes, toHash[k][:]...)
	}

	//c_old = hash_to_scalar(toHash);
	c_old = *(crypto.HashToScalar(toHash_bytes)) // hash_to_scalar(toHash);

	i = (index + 1) % cols

	// fmt.Printf("hash to scalar calculated %s  index = %d\n", c_old,i)

	if i == 0 {
		rv.cc = c_old
	}
	for i != index {

		//rv.ss[i] = skvGen(rows);
		for j := 0; j < rows; j++ {
			rv.ss[i][j] = crypto.SkGen()

			//  Sc_0(&rv.ss[i][j]); // make random key zero  for tesing puprpose // BUG if line is uncommented
			//ScReduce32(&rv.ss[i][j]) // reduce it

		}
		crypto.Sc_0(&c)

		for j := 0; j < dsRows; j++ {
			crypto.AddKeys2(&L, &rv.ss[i][j], &c_old, &pk[i][j])
			Hi = pk[i][j].HashToPoint()
			crypto.AddKeys3(&R, &rv.ss[i][j], &Hi, &c_old, &Ip[j])
			/*fmt.Printf("R = %s\n",R)
			fmt.Printf("rv.ss[i][j] = %s\n",rv.ss[i][j])
			fmt.Printf("Hi = %s\n",Hi)
			fmt.Printf("c_old = %s\n",c_old)*/

			toHash[3*j+1] = pk[i][j]
			toHash[3*j+2] = L
			toHash[3*j+3] = R
		}

		//second loop
		for j, ii := dsRows, 0; j < rows; j, ii = j+1, ii+1 {
			crypto.AddKeys2(&L, &rv.ss[i][j], &c_old, &pk[i][j]) // here L is getting used again
			toHash[ndsRows+2*ii+1] = pk[i][j]
			toHash[ndsRows+2*ii+2] = L
		}

		toHash_bytes = toHash_bytes[:0] // zero out everything
		for k := range toHash {
			//fmt.Printf("gen hash cc index %d hashindex %d %s\n",i,k,toHash[k])
			toHash_bytes = append(toHash_bytes, toHash[k][:]...)
		}

		c = *(crypto.HashToScalar(toHash_bytes)) // hash_to_scalar(toHash);

		copy(c_old[:], c[:])

		i = (i + 1) % cols

		//fmt.Printf("cc index %d %s\n", i, c)

		if i == 0 {
			rv.cc = c_old
		}
	}
	for j = 0; j < rows; j++ {
		crypto.ScMulSub(&rv.ss[index][j], &c, &xx[j], &alpha[j])
	}

	//fmt.Printf("cc gen %s\n", rv.cc)
	return rv
}

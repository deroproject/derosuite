package astrobwt

//import "os"
//import "fmt"

import "sync"
import "encoding/binary"
import "golang.org/x/crypto/sha3"

import "golang.org/x/crypto/salsa20/salsa"

// see here to improve the algorithms more https://github.com/y-256/libdivsufsort/blob/wiki/SACA_Benchmarks.md

// Original implementation was in xmrig miner, however it had a flaw which has been fixed
// this optimized algorithm is used only  in the miner and not in the blockchain

//const stage1_length int = 147253 // it is a prime
//const max_length int = 1024*1024 + stage1_length + 1024

type Data struct {
	stage1        [stage1_length + 64]byte // stages are taken from it
	stage1_result [stage1_length + 1]byte
	stage2        [1024*1024 + stage1_length + 1 + 64]byte
	stage2_result [1024*1024 + stage1_length + 1]byte
	indices       [ALLOCATION_SIZE]uint64
	tmp_indices   [ALLOCATION_SIZE]uint64
}

var pool = sync.Pool{New: func() interface{} { return &Data{} }}

func POW_optimized_v1(inputdata []byte, max_limit int) (outputhash [32]byte, success bool) {
	data := pool.Get().(*Data)
	outputhash, success = POW_optimized_v2(inputdata,max_limit,data)
	pool.Put(data)
	return
}
func POW_optimized_v2(inputdata []byte, max_limit int, data *Data) (outputhash [32]byte, success bool) {

	var counter [16]byte


	for i := range data.stage1 {
		data.stage1[i] = 0
	}
	/* for i := range data.stage1_result{
	    data.stage1_result[i] =0
	}*/

	key := sha3.Sum256(inputdata)
	salsa.XORKeyStream(data.stage1[1:stage1_length+1], data.stage1[1:stage1_length+1], &counter, &key)
	sort_indices(stage1_length+1, data.stage1[:], data.stage1_result[:], data)
	key = sha3.Sum256(data.stage1_result[:])
	stage2_length := stage1_length + int(binary.LittleEndian.Uint32(key[:])&0xfffff)

    if stage2_length > max_limit {
        for i := range outputhash { // will be optimized by compiler
		    outputhash[i] = 0xff
	    }
        success = false
      return
    }

	for i := range counter { // will be optimized by compiler
		counter[i] = 0
	}

	salsa.XORKeyStream(data.stage2[1:stage2_length+1], data.stage2[1:stage2_length+1], &counter, &key)
	sort_indices(stage2_length+1, data.stage2[:], data.stage2_result[:], data)
	key = sha3.Sum256(data.stage2_result[:stage2_length+1])
	for i := range data.stage2{
	    data.stage2[i] =0
	}

	copy(outputhash[:], key[:])
    success = true
	return
}

const COUNTING_SORT_BITS uint64 = 10
const COUNTING_SORT_SIZE uint64 = 1 << COUNTING_SORT_BITS

const ALLOCATION_SIZE = MAX_LENGTH

func BigEndian_Uint64(b []byte) uint64 {
	_ = b[7] // bounds check hint to compiler; see golang.org/issue/14808
	return uint64(b[7]) | uint64(b[6])<<8 | uint64(b[5])<<16 | uint64(b[4])<<24 |
		uint64(b[3])<<32 | uint64(b[2])<<40 | uint64(b[1])<<48 | uint64(b[0])<<56
}

func smaller(input []uint8, a, b uint64) bool {
	value_a := a >> 21
	value_b := b >> 21

	if value_a < value_b {
		return true
	}

	if value_a > value_b {
		return false
	}

	data_a := BigEndian_Uint64(input[(a%(1<<21))+5:])
	data_b := BigEndian_Uint64(input[(b%(1<<21))+5:])
	return data_a < data_b
}

// basically
func sort_indices(N int, input_extra []byte, output []byte, d *Data) {

	var counters [2][COUNTING_SORT_SIZE]uint32
	indices := d.indices[:]
	tmp_indices := d.tmp_indices[:]

	input := input_extra[1:]

	loop3 := N / 3 * 3
	for i := 0; i < loop3; i += 3 {
		k0 := BigEndian_Uint64(input[i:])
		counters[0][(k0>>(64-COUNTING_SORT_BITS*2))&(COUNTING_SORT_SIZE-1)]++
		counters[1][k0>>(64-COUNTING_SORT_BITS)]++
		k1 := k0 << 8
		counters[0][(k1>>(64-COUNTING_SORT_BITS*2))&(COUNTING_SORT_SIZE-1)]++
		counters[1][k1>>(64-COUNTING_SORT_BITS)]++
		k2 := k0 << 16
		counters[0][(k2>>(64-COUNTING_SORT_BITS*2))&(COUNTING_SORT_SIZE-1)]++
		counters[1][k2>>(64-COUNTING_SORT_BITS)]++
	}

	if N%3 != 0 {
		for i := loop3; i < N; i++ {
			k := BigEndian_Uint64(input[i:])
			counters[0][(k>>(64-COUNTING_SORT_BITS*2))&(COUNTING_SORT_SIZE-1)]++
			counters[1][k>>(64-COUNTING_SORT_BITS)]++
		}
	}

	/*
	   	for i := 0; i < N ; i++{
	   		k := BigEndian_Uint64(input[i:])
	     		counters[0][(k >> (64 - COUNTING_SORT_BITS * 2)) & (COUNTING_SORT_SIZE - 1)]++
	   		counters[1][k >> (64 - COUNTING_SORT_BITS)]++
	   	}
	*/

	prev := [2]uint32{counters[0][0], counters[1][0]}
	counters[0][0] = prev[0] - 1
	counters[1][0] = prev[1] - 1
	var cur [2]uint32
	for i := uint64(1); i < COUNTING_SORT_SIZE; i++ {
		cur[0], cur[1] = counters[0][i]+prev[0], counters[1][i]+prev[1]
		counters[0][i] = cur[0] - 1
		counters[1][i] = cur[1] - 1
		prev[0] = cur[0]
		prev[1] = cur[1]
	}

	for i := N - 1; i >= 0; i-- {
		k := BigEndian_Uint64(input[i:])
		// FFFFFFFFFFE00000 =  (0xFFFFFFFFFFFFFFF<< 21)  // to clear bottom 21 bits
		tmp := counters[0][(k>>(64-COUNTING_SORT_BITS*2))&(COUNTING_SORT_SIZE-1)]
		counters[0][(k>>(64-COUNTING_SORT_BITS*2))&(COUNTING_SORT_SIZE-1)]--

		tmp_indices[tmp] = (k & 0xFFFFFFFFFFE00000) | uint64(i)
	}

	for i := N - 1; i >= 0; i-- {
		data := tmp_indices[i]
		tmp := counters[1][data>>(64-COUNTING_SORT_BITS)]
		counters[1][data>>(64-COUNTING_SORT_BITS)]--
		indices[tmp] = data
	}

	prev_t := indices[0]
	for i := 1; i < N; i++ {
		t := indices[i]
		if smaller(input, t, prev_t) {
			t2 := prev_t
			j := i - 1
			for {
				indices[j+1] = prev_t
				j--
				if j < 0 {
					break
				}
				prev_t = indices[j]
				if !smaller(input, t, prev_t) {
					break
				}
			}
			indices[j+1] = t
			t = t2
		}
		prev_t = t
	}

	// optimized unrolled code below this comment
	/*for i := 0; i < N;i++{
		output[i] =  input_extra[indices[i] & ((1 << 21) - 1) ]
	}*/

	loop4 := ((N + 1) / 4) * 4
	for i := 0; i < loop4; i += 4 {
		output[i+0] = input_extra[indices[i+0]&((1<<21)-1)]
		output[i+1] = input_extra[indices[i+1]&((1<<21)-1)]
		output[i+2] = input_extra[indices[i+2]&((1<<21)-1)]
		output[i+3] = input_extra[indices[i+3]&((1<<21)-1)]
	}
	for i := loop4; i < N; i++ {
		output[i] = input_extra[indices[i]&((1<<21)-1)]
	}

	// there is an issue above, if the last byte of input is 0x00, initialbytes are wrong, this fix may not be complete
	if N > 3 && input[N-2] == 0 {
		backup_byte := output[0]
		output[0] = 0
		for i := 1; i < N; i++ {
			if output[i] != 0 {
				output[i-1] = backup_byte
				break
			}
		}
	}

}

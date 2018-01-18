package difficulty


import "fmt"
import "math/big"

import "github.com/deroproject/derosuite/config"
import "github.com/deroproject/derosuite/crypto"

var (
	// bigZero is 0 represented as a big.Int.  It is defined here to avoid
	// the overhead of creating it multiple times.
	bigZero = big.NewInt(0)

	// bigOne is 1 represented as a big.Int.  It is defined here to avoid
	// the overhead of creating it multiple times.
	bigOne = big.NewInt(1)

	// oneLsh256 is 1 shifted left 256 bits.  It is defined here to avoid
	// the overhead of creating it multiple times.
	oneLsh256 = new(big.Int).Lsh(bigOne, 256)
)

// HashToBig converts a PoW has into a big.Int that can be used to
// perform math comparisons.
func HashToBig(buf crypto.Hash) *big.Int {
	// A Hash is in little-endian, but the big package wants the bytes in
	// big-endian, so reverse them.
	blen := len(buf) // its hardcoded 32 bytes, so why do len but lets do it
	for i := 0; i < blen/2; i++ {
		buf[i], buf[blen-1-i] = buf[blen-1-i], buf[i]
	}

	return new(big.Int).SetBytes(buf[:])
}

// this function calculates the difficulty in big num form
func ConvertDifficultyToBig(difficultyi uint64) *big.Int {
	if difficultyi == 0 {
		panic("difficulty can never be zero")
	}
	// (1 << 256) / (difficultyNum )
	difficulty := new(big.Int).SetUint64(difficultyi)
	denominator := new(big.Int).Add(difficulty, bigZero) // above 2 lines can be merged
	return new(big.Int).Div(oneLsh256, denominator)
}

// this function check whether the pow hash meets difficulty criteria
func CheckPowHash(pow_hash crypto.Hash, difficulty uint64) bool {
	big_difficulty := ConvertDifficultyToBig(difficulty)
	big_pow_hash := HashToBig(pow_hash)

	if big_pow_hash.Cmp(big_difficulty) <= 0 { // if work_pow is less than difficulty
		return true
	}
	return false
}

/* this function calculates difficulty on the basis of previous timestamps and cumulative_difficulty */
func Next_Difficulty(timestamps []uint64, cumulative_difficulty []uint64, target_seconds uint64) (difficulty uint64) {

	difficulty = 1 // default difficulty is 1 // for genesis block

	if len(timestamps) > config.DIFFICULTY_BLOCKS_COUNT_V2 {
		panic("More timestamps provided than required")
	}
	if len(timestamps) != len(cumulative_difficulty) {
		panic("Number of timestamps != Number of cumulative_difficulty")
	}

	if len(timestamps) <= 1 {
		return difficulty // return 1
	}

	length := uint64(len(timestamps))

	weighted_timespans := uint64(0)
	for i := uint64(1); i < length; i++ {
		timespan := uint64(0)
		if timestamps[i-1] >= timestamps[i] {
			timespan = 1
		} else {
			timespan = timestamps[i] - timestamps[i-1]
		}
		if timespan > (10 * target_seconds) {
			timespan = 10 * target_seconds
		}
		weighted_timespans += i * timespan
	}

	minimum_timespan := (target_seconds * length) / 2
	if weighted_timespans < minimum_timespan { // fix startup weirdness
		weighted_timespans = minimum_timespan
	}

	total_work := cumulative_difficulty[length-1] - cumulative_difficulty[0]

	// convert input for 128 bit multiply
	var big_total_work, big_target, big_result big.Int
	big_total_work.SetUint64(total_work)

	target := (((length + 1) / 2) * target_seconds * 3) / 2
	big_target.SetUint64(target)

	big_result.Mul(&big_total_work, &big_target)

	if big_result.IsUint64() {
		if big_result.Uint64() > 0x000fffffffffffff { // this will give us atleast 1 year to fix the difficulty algorithm
			fmt.Printf("Total work per target_time  will soon cross 2^64, please fix the difficulty algorithm\n")
		}
		difficulty = big_result.Uint64() / weighted_timespans
	} else {
		panic("Total work per target_time crossing 2^64 , please fix the above")
	}

	return difficulty
}

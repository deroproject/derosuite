package emission

//import "encoding/hex"
//import "bytes"
import "testing"

func Test_Emission_Rewards(t *testing.T) {

	bl_median_size := uint64(0)
	bl_current_size := uint64(0)
	already_generated_coins := uint64(0)
	hard_fork_version := uint64(6)
	fee := uint64(0)
	if GetBlockReward(bl_median_size, bl_current_size, already_generated_coins, hard_fork_version, fee) != uint64(1000000000000) {
		t.Error("Block reward failed for genesis block\n")
	}

	already_generated_coins = uint64(1000000000000)

	if GetBlockReward(bl_median_size, bl_current_size, already_generated_coins, hard_fork_version, fee) != uint64(2000000000000000000) {
		t.Error("Block reward failed for pre-mine\n")
	}

	already_generated_coins += uint64(2000000000000000000)

	calculated := GetBlockReward(bl_median_size, bl_current_size, already_generated_coins, hard_fork_version, fee)
	expected := uint64(31369672915858)
	if calculated != expected {
		t.Errorf("Block reward failed for 2nd block estm %d calculated %d\n", expected, calculated)
	}

	already_generated_coins += uint64(31369672915858)

	// calculated block reward for 3rd block
	calculated = GetBlockReward(bl_median_size, bl_current_size, already_generated_coins, hard_fork_version, fee)
	expected = uint64(31369613082955)

	if calculated != expected {
		t.Errorf("Block reward failed for 3rd block estm %d calculated %d\n", expected, calculated)
	}

}

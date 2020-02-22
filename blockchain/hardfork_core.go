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

import "github.com/deroproject/derosuite/block"
import "github.com/deroproject/derosuite/config"
import "github.com/deroproject/derosuite/storage"
import "github.com/deroproject/derosuite/globals"

// the voting for hard fork works as follows
// block major version remains contant, while block minor version contains the next hard fork number,
// at trigger height, the last window_size blocks are counted as folllows
// if Major_Version == minor version, it is a negative vote
// if minor_version > major_version, it is positive vote
// if threshold votes are positive, the next hard fork triggers

// this is work in progress
// hard forking is integrated deep within the the blockchain as almost anything can be replaced in DERO without disruption
const default_voting_window_size = 6000 // this many votes will counted
const default_vote_percent = 62         // 62 percent votes means the hard fork is locked in

type Hard_fork struct {
	Version     int64 // which version will trigger
	Height      int64 // at what height hard fork will come into effect, trigger block
	Window_size int64 // how many votes to count (x number of votes)
	Threshold   int64 //  between 0 and 99  // percent number of votes required to lock in  hardfork, 0 = mandatory
	Votes       int64 // number of votes in favor
	Voted       bool  // whether voting resulted in hardfork
}

// current mainnet_hard_forks
var mainnet_hard_forks = []Hard_fork{
	// {0, 0,0,0,0,true}, // dummy entry so as we can directly use the fork index into this entry
	{1,      0, 0, 0, 0, true}, // version 1 hard fork where genesis block landed and chain migration occurs
	// version 1 has difficulty hardcoded to 1
	{2, 95551,  0, 0, 0, true}, // version 2 hard fork where Atlantis bootstraps , it's mandatory
        {3, 721000, 0, 0, 0, true}, // version 3 hard fork emission fix, it's mandatory
        {4, 4550555, 0, 0, 0, true}, // version 4 hard fork AstroBWT CPU Mining enabled. It's mandatory
}

// current testnet_hard_forks
var testnet_hard_forks = []Hard_fork{
	//{1, 0, 0, 0, 0, true},    // version 1 hard fork where genesis block landed
	{3, 0, 0, 0, 0, true}, // version 3 hard fork where we started , it's mandatory
  	{4, 3, 0, 0, 0, true}, // version 4 hard fork where we change mining algorithm it's mandatory
}

// current simulation_hard_forks
// these can be tampered with for testing and other purposes
// this variable is exported so as simulation can play/test hard fork code
var Simulation_hard_forks = []Hard_fork{
	{1, 0, 0, 0, 0, true}, // version 1 hard fork where genesis block landed
	{2, 1, 0, 0, 0, true}, // version 2 hard fork where we started , it's mandatory
}

// at init time, suitable versions are selected
var current_hard_forks []Hard_fork

// init suitable structure based on mainnet/testnet selection at runtime
func init_hard_forks(params map[string]interface{}) {

	// if simulation , load simulation features
	if params["--simulator"] == true {
		current_hard_forks = Simulation_hard_forks // enable simulator mode hard forks
		logger.Debugf("simulator hardforks are online")
	} else {
		if globals.IsMainnet() {
			current_hard_forks = mainnet_hard_forks
			logger.Debugf("mainnet hardforks are online")
		} else {
			current_hard_forks = testnet_hard_forks
			logger.Debugf("testnet hardforks are online")
		}

	}

	// if voting in progress, load all votes from db, since we do not store votes in disk,
	// we will load all necessary blocks, counting votes
}

// check block version validity at specific height according to current network
func (chain *Blockchain) Check_Block_Version(dbtx storage.DBTX, bl *block.Block) (result bool) {

	height := chain.Calculate_Height_At_Tips(dbtx, bl.Tips)

	if height == 0 && bl.Major_Version == 1 { // handle genesis block as exception
		return true
	}
	// all blocks except genesis block land here
	if bl.Major_Version == uint64(chain.Get_Current_Version_at_Height(height)) {
		return true
	}

	return
}

// this func will recount votes, set whether the version is voted in
// only the main chain blocks are counted in
// this func must be called with chain in lock state
/*
func (chain *Blockchain) Recount_Votes() {
	height := chain.Load_Height_for_BL_ID(chain.Get_Top_ID())

	for i := len(current_hard_forks) - 1; i > 0; i-- {
		// count votes only if voting is in progress
		if 0 != current_hard_forks[i].Window_size && // if window_size > 0
			height <= current_hard_forks[i].Height &&
			height >= (current_hard_forks[i].Height-current_hard_forks[i].Window_size) { // start voting when required

			hard_fork_locked := false
			current_hard_forks[i].Votes = 0 // make votes zero, before counting
			for j := height; j >= (current_hard_forks[i].Height - current_hard_forks[i].Window_size); j-- {
				// load each block, and count the votes

				hash, err := chain.Load_BL_ID_at_Height(j)
				if err == nil {
					bl, err := chain.Load_BL_FROM_ID(hash)
					if err == nil {
						if bl.Minor_Version == uint64(current_hard_forks[i].Version) {
							current_hard_forks[i].Votes++
						}
					} else {
						logger.Warnf("err loading block (%s) at height %d,  chain height %d err %s", hash, j, height, err)
					}

				} else {
					logger.Warnf("err loading block at height %d,  chain height %d err %s", j, height, err)
				}

			}

			// if necessary votes have been accumulated , lock in the hard fork
			if ((current_hard_forks[i].Votes * 100) / current_hard_forks[i].Window_size) >= current_hard_forks[i].Threshold {
				hard_fork_locked = true
			}
			current_hard_forks[i].Voted = hard_fork_locked // keep it as per status
		}

	}

}
*/
// this function returns  number of information whether hf is going on scheduled, everything is okay etc
func (chain *Blockchain) Get_HF_info() (state int, enabled bool, earliest_height, threshold, version, votes, window int64) {

	state = 2 // default is everything is okay
	enabled = true

	topoheight := chain.Load_TOPO_HEIGHT(nil)
	block_id, err := chain.Load_Block_Topological_order_at_index(nil, topoheight)
	if err != nil {
		return
	}

	bl, err := chain.Load_BL_FROM_ID(nil, block_id)
	if err != nil {
		logger.Warnf("err loading block (%s) at topo height %d err %s", block_id, topoheight, err)
	}

	height := chain.Load_Height_for_BL_ID(nil, block_id)

	version = chain.Get_Current_Version_at_Height(height)

	// check top block to see if the network is going through a hard fork
	if bl.Major_Version != bl.Minor_Version { // network is going through voting
		state = 0
		enabled = false
	}

	if bl.Minor_Version != uint64(chain.Get_Ideal_Version_at_Height(height)) {
		// we are NOT voting for the hard fork ( or we are already broken), issue warning to user, that we need an upgrade NOW
		state = 1
		enabled = false
		version = int64(bl.Minor_Version)
	}
	if state == 0 { // we know our state is good, report back, good info
		for i := range current_hard_forks {
			if version == current_hard_forks[i].Version {
				earliest_height = current_hard_forks[i].Height
				threshold = current_hard_forks[i].Threshold
				version = current_hard_forks[i].Version
				votes = current_hard_forks[i].Votes
				window = current_hard_forks[i].Window_size
			}
		}
	}

	return
}

// current hard fork version , block major version
// we may be at genesis block height
func (chain *Blockchain) Get_Current_Version() int64 { // it is last version voted or mandatory update
	return chain.Get_Current_Version_at_Height(chain.Get_Height())
}

func (chain *Blockchain) Get_Current_BlockTime() uint64 { // it is last version voted or mandatory update
   block_time:= config.BLOCK_TIME
   if chain.Get_Current_Version() >= 4 {
        block_time= config.BLOCK_TIME_hf4
    }
	return block_time
}


func (chain *Blockchain) Get_Current_Version_at_Height(height int64) int64 {
	for i := len(current_hard_forks) - 1; i >= 0; i-- {
		//logger.Infof("i %d height %d  hf height %d",i, height,current_hard_forks[i].Height )
		if height >= current_hard_forks[i].Height {

			// if it was a mandatory fork handle it directly
			if current_hard_forks[i].Threshold == 0 {
				return current_hard_forks[i].Version
			}

			if current_hard_forks[i].Voted { // if the version was voted in, select it, other wise try lower
				return current_hard_forks[i].Version
			}
		}
	}
	return 0
}

// if we are voting, return the next expected version
func (chain *Blockchain) Get_Ideal_Version() int64 {
	return chain.Get_Ideal_Version_at_Height(chain.Get_Height())
}

// used to cast vote
func (chain *Blockchain) Get_Ideal_Version_at_Height(height int64) int64 {
	for i := len(current_hard_forks) - 1; i > 0; i-- {
		// only voted during the period required
		if height <= current_hard_forks[i].Height &&
			height >= (current_hard_forks[i].Height-current_hard_forks[i].Window_size) { // start voting when required
			return current_hard_forks[i].Version
		}
	}

	// if we are not voting, return current version
	return chain.Get_Current_Version_at_Height(height)
}

/*

//  if the block major version is more than what we have in our index, display warning to user
func (chain *Blockchain) Display_Warning_If_Blocks_are_New(bl *block.Block)  {
        // check the biggest fork
        if current_hard_forks[len(current_hard_forks )-1].version < bl.Major_Version {
            logger.Warnf("We have seen new blocks floating with version number bigger than ours, please update the software")
        }
 return
}
*/

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

package main

/*
import "io"
import "os"

import "fmt"
import "bytes"
import "bufio"
import "strings"
import "strconv"
import "runtime"
import "crypto/sha1"
import "encoding/hex"
import "encoding/json"
import "path/filepath"

import "github.com/romana/rlog"
import "github.com/chzyer/readline"
import "github.com/docopt/docopt-go"
import log "github.com/sirupsen/logrus"

import "github.com/deroproject/derosuite/address"
import "github.com/deroproject/derosuite/p2pv2"


import "github.com/deroproject/derosuite/config"

import "github.com/deroproject/derosuite/transaction"

//import "github.com/deroproject/derosuite/checkpoints"
import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/crypto/ringct"
import "github.com/deroproject/derosuite/blockchain/rpcserver"
*/

import "time"
import "sync"
import "math/big"
import "crypto/rand"
import "sync/atomic"

//import "encoding/hex"
import "encoding/binary"

import "github.com/deroproject/derosuite/block"
import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/globals"
import "github.com/deroproject/derosuite/address"
import "github.com/deroproject/derosuite/blockchain"
import "github.com/deroproject/derosuite/cryptonight"

// p2p needs to export a varible declaring whether the chain is in syncronising mode

var counter uint64 = 0 // used to track speeds of current miner

var mining bool // whether system is mining

// request block chain template, see if the tip changes, then continously mine
func start_miner(chain *blockchain.Blockchain, addr *address.Address, threads int) {

	mining = true
	counter = 0
	//tip_counter := 0

	for {
		//time.Sleep(50 * time.Millisecond)

		if !mining {
			break
		}

		if chain.MINING_BLOCK == true {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		cbl, bl := chain.Create_new_miner_block(*addr)

		difficulty := chain.Get_Difficulty_At_Tips(nil, bl.Tips)

		//globals.Logger.Infof("Difficulty of new block is %s", difficulty.String())
		// calculate difficulty once
		// update job from chain
		wg := sync.WaitGroup{}
		wg.Add(threads) // add total number of tx as work

		for i := 0; i < threads; i++ {
			go generate_valid_PoW(chain, 0, cbl, cbl.Bl, difficulty, &wg) // work should be complete in approx 100 ms, on a 12 cpu system, this would add cost of launching 12 g routine per second
		}
		wg.Wait()
	}

	// g
}

// each invoke will be take atleast 250 milliseconds
func generate_valid_PoW(chain *blockchain.Blockchain, hf_version uint64, cbl *block.Complete_Block, bl *block.Block, current_difficulty *big.Int, wg *sync.WaitGroup) {
	var powhash crypto.Hash
	block_work := bl.GetBlockWork()

	// extra nonce is always at offset 36 and is of length 32
	var extra_nonce [16]byte
	rand.Read(extra_nonce[:]) // fill extra nonce with random buffer

	bl.SetExtraNonce(extra_nonce[:])

	// TODO  this time must be replaced  by detecting TIP change
	start := time.Now()
	//deadline := time.Now().Add(250*time.Millisecond)
	i := uint32(0)

	nonce_buf := block_work[39 : 39+4] // take last 8 bytes as nonce counter and bruteforce it, since slices are linked, it modifies parent
	for {
		//time.Sleep(1000 * time.Millisecond)

		atomic.AddUint64(&counter, 1)

		binary.BigEndian.PutUint32(nonce_buf, i)

		//PoW := crypto.Scrypt_1024_1_1_256(block_work)
		//PoW := crypto.Keccak256(block_work)
		PoW := cryptonight.SlowHash(block_work)
		copy(powhash[:], PoW[:])

		if blockchain.CheckPowHashBig(powhash, current_difficulty) == true {

			bl.CopyNonceFromBlockWork(block_work)
			//globals.Logger.Infof("Pow Successfully solved, Submitting block")

			if _, ok := chain.Add_Complete_Block(cbl); ok {
				globals.Logger.Infof("Block %s successfully accepted diff %s", bl.GetHash(), current_difficulty.String())
				chain.P2P_Block_Relayer(cbl, 0) // broadcast block to network ASAP

				break
			}

		}

		if time.Now().Sub(start) > 250*time.Millisecond {
			break
		}
		i++

	}

	wg.Done()

}

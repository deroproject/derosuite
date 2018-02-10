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

/*
import log "github.com/sirupsen/logrus"

import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/block"
import "github.com/deroproject/derosuite/transaction"
//import "github.com/deroproject/derosuite/blockchain/mempool"

*/
// DERO blockchain has been designed/developed as a state machine ( single-threaded)
// The state machine  cannot change until there is external input
// the blockchain has 2 events as input
//    1) new block
//    2) new transaction
// So, the blockchain just waits for events from 2 sources
//    1) p2p layer  ( network did a trasaction or found new block after mining)
//    2) user side ( user did  a transaction or found new block after mining)
// the design has been simplified so as smart contracts can be integrated easily
// NOTE that adding a block is an atomic event in DERO blockchain

//
//
// This is the global event handler for block
// any block whether mined locally or via network must be dispatched using this channel
//var Incoming_Block_Channel = make(chan *block.Complete_Block, 512) // upto 500 blocks can be queued
//var Incoming_Transaction_Channel = make(chan *transaction.Transaction, 512) // upto 500 transactions can be queued

/*

// infinite looping function to process incoming block
// below event loops are never terminated, not even while exiting
// but we take the lock of chain, so as state/db cannot be changed
// also note that p2p layer is stopped earlier, so input cannot appear
func (chain *Blockchain)  Handle_Block_Event_Loop(){
	for{

		  select{
		  	case <-chain.Exit_Event:
					logger.Debugf("Exiting Block event loop")
					return
			default:
			}

		  select{
		  	case <-chain.Exit_Event:
					logger.Debugf("Exiting Block event loop")
					return

			case complete_bl := <- Incoming_Block_Channel:
				     logger.Debugf("Incoming New Block")
				       func (){
				       	var blid crypto.Hash
				       	_ = blid

				       //	defer func() {  // safety so if anything wrong happens, verification fails
        			///	if r := recover(); r != nil {
            		//		logger.WithFields( log.Fields{"blid": blid}).Warnf("Recovered while processing incoming block")
        			//	}}()

        				blid = complete_bl.Bl.GetHash()


        				chain.add_Complete_Block(complete_bl)

				       }()
			//default:

			//case <- Exit_Event:
		}

	}
}



// infinite looping function to process incoming block
func (chain *Blockchain) Handle_Transaction_Event_Loop(){
	for{
		select {
			case <-chain.Exit_Event:
					logger.Debugf("Exiting Block event loop")
					return

			case tx := <- Incoming_Transaction_Channel:

				logger.Debugf("Incoming New Transaction")
				func() {
					var txid crypto.Hash
				defer func() {  // safety so if anything wrong happens, verification fails

        if r := recover(); r != nil {
            logger.WithFields( log.Fields{"txid": txid}).Warnf("Recovered while Verifying transaction, failed verification")
            //result = false
        }}()

         txid = tx.GetHash()

        }()

		//	case <- Exit_Event:
		}

	}

}
*/

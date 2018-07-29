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

import "testing"

import "github.com/deroproject/derosuite/crypto"

/*
 * raw data from daemon
 *
 * /*
 * 2018-01-07 20:18:21.157 [P2P4]  INFO    global  src/cryptonote_core/blockchain.cpp:1436        ----- BLOCK ADDED AS ALTERNATIVE ON HEIGHT 16163
id:     <f9a3faa33054a4a1fa349321c546ee5f42cc416f13a991152c64fcbef994518b>
PoW:    <28b6fdf6655c45631c28be02bc528342b25fe913696911b788a01e0b0a000000>
difficulty:     77897895
2018-01-07 21:17:40.182 [P2P9]  INFO    global  src/cryptonote_protocol/cryptonote_protocol_handler.inl:1521   SYNCHRONIZED OK
2018-01-07 22:04:14.368 [P2P2]  INFO    global  src/p2p/net_node.inl:258      Host 125.161.128.47 blocked.
status
Height: 16294/16294 (100.0%) on mainnet, not mining, net hash 780.13 kH/s, v6, up to date, 0(out)+15(in) connections, uptime 0d 10h 29m 2s
2018-01-08 03:14:37.490 [P2P1]  INFO    global  src/cryptonote_core/blockchain.cpp:1436        ----- BLOCK ADDED AS ALTERNATIVE ON HEIGHT 13618
id:     <a3918ac81a08e8740f99f79ff788d9e147ceb7e530ed590ac1e0f5d1cbba28c5>
PoW:    <b34caa51543b82efee0336677dd825e3236220e69d2f090c58df0b3e05000000>
difficulty:     90940906
*/
func Test_CheckPowHash(t *testing.T) {

	hash := crypto.Hash{0x28, 0xb6, 0xfd, 0xf6, 0x65, 0x5c, 0x45, 0x63, 0x1c, 0x28, 0xbe,
		0x02, 0xbc, 0x52, 0x83, 0x42, 0xb2, 0x5f, 0xe9, 0x13, 0x69, 0x69,
		0x11, 0xb7, 0x88, 0xa0, 0x1e, 0x0b, 0x0a, 0x00, 0x00, 0x00}

	difficulty := uint64(77897895)

	if !CheckPowHash(hash, difficulty) {
		t.Errorf("POW  check failedm, severe BUG\n")
	}

	hash = crypto.Hash{0xb3, 0x4c, 0xaa, 0x51, 0x54, 0x3b, 0x82, 0xef, 0xee, 0x03, 0x36, 0x67,
		0x7d, 0xd8, 0x25, 0xe3, 0x23, 0x62, 0x20, 0xe6, 0x9d, 0x2f, 0x09,
		0x0c, 0x58, 0xdf, 0x0b, 0x3e, 0x05, 0x00, 0x00, 0x00}

	difficulty = uint64(77897895)

	if !CheckPowHash(hash, difficulty) {
		t.Errorf("POW  check 2 failed, severe BUG\n")
	}

	difficulty = uint64(77897895000)

	if CheckPowHash(hash, difficulty) {
		t.Errorf("POW  check 3 failed, severe BUG\n")
	}

}

// test if difficulty is ever passed as zero
func Test_Panic_Invalid_Difficulty(t *testing.T) {

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Difficulty did not panic on 0 difficulty")
		}
	}()
	ConvertDifficultyToBig(0)
}

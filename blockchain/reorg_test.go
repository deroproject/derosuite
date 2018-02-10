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

import "sort"
import "testing"

import "github.com/deroproject/derosuite/crypto"

/* this function tests the core chain selection algorithm for reorganisation purpose */
func Test_chain_sort(t *testing.T) {

	var first_chain crypto.Hash  // all zero
	var second_chain crypto.Hash // first bytes 1
	var third_chain crypto.Hash  // first bytes 2
	first_chain[0] = 0
	second_chain[0] = 1
	third_chain[0] = 2

	{
		var chain bestChain

		chain = append(chain, chain_data{
			hash:        first_chain, // this is the best chain based , if cumulative difficulty and time are same
			cdifficulty: 99,
			foundat:     99})

		chain = append(chain, chain_data{ // this is the best chain based on cumulative difficulty
			hash:        second_chain,
			cdifficulty: 123,
			foundat:     96})

		chain = append(chain, chain_data{
			hash:        third_chain,
			cdifficulty: 80,
			foundat:     94}) //if cumulative diff are same , this is the best chain

		sort.Sort(chain)

		if chain[0].hash != second_chain {
			t.Errorf("Best chain not selected   instead selected %d\n", chain[0].hash[0])
		}

		if chain[1].hash != first_chain || chain[2].hash != third_chain {
			t.Errorf("Best chain not selected fault in sorting\n")
		}

		// lets trigger the case when cumulative difficulty are same
		chain[0].cdifficulty = 5000
		chain[1].cdifficulty = 5000
		chain[2].cdifficulty = 5000

		sort.Sort(chain)

		if chain[0].hash != third_chain {
			t.Errorf("Best chain not selected   instead selected %d\n", chain[0].hash[0])
		}

		if chain[1].hash != second_chain || chain[2].hash != first_chain {
			t.Errorf("Best chain not selected fault in sorting\n")
		}

		// lets trigger the case when cumulative difficulty are same and also find time are same
		chain[0].foundat = 5000
		chain[1].foundat = 5000
		chain[2].foundat = 5000

		sort.Sort(chain)

		if chain[0].hash != first_chain {
			t.Errorf("Best chain not selected   instead selected %d\n", chain[0].hash[0])
		}

		if chain[1].hash != second_chain || chain[2].hash != third_chain {
			t.Errorf("Best chain not selected fault in sorting\n")
		}

	}

}

/* since reorganisation based consensus is tough to test at this point this time
 * we have extracted a snapshot based on dero chain wirh details below,
 * an altchain 13 long
 * 13 blocks long, from height 16996 (816 deep), diff 538095141258 :77d6b92961146232ba36ced359f6dbbf3170299b530d2c7d0da5c398f7713d23
 *
1 blocks long, from height 16995 (817 deep), diff 537112495290:597ff30f0913834ecbc59e4273619e930c2166dc07c1362c6ed6387265b2ddd8
1 blocks long, from height 16859 (953 deep), diff 525062172399:5153884d9cf66bd6365b99fa7aad410a6fa219c4d749f186471f1166e76b4082
2 blocks long, from height 13619 (4193 deep), diff 288397472393:ff36451029ad3a021629f38f5f88a3d76bf4208589530b2a4b57a535f16aaf89
1 blocks long, from height 16163 (1649 deep), diff 468618212179:f9a3faa33054a4a1fa349321c546ee5f42cc416f13a991152c64fcbef994518b
1 blocks long, from height 15874 (1938 deep), diff 455084019052:143cc96808483da603dc704b75b6114699d99bdfb9c415a5be20d42e63b1cc4f
1 blocks long, from height 17397 (415 deep), diff 578283212318:6c3106b25a858c178a6e3eb7614438cf6d9faaecc1ea8fe416260d587867deb3
1 blocks long, from height 13618 (4194 deep), diff 288208474072:a3918ac81a08e8740f99f79ff788d9e147ceb7e530ed590ac1e0f5d1cbba28c5
13 blocks long, from height 16996 (816 deep), diff 538095141258:77d6b92961146232ba36ced359f6dbbf3170299b530d2c7d0da5c398f7713d23

 *
2018-01-11 09:13:08.250 [P2P4]  INFO    global  src/cryptonote_protocol/cryptonote_protocol_handler.inl:305     [78.131.18.216:
57948 INC] Sync data returned a new top block can
didate: 17822 -> 17022 [Your node is 800 blocks (0 days) ahead]
SYNCHRONIZATION started
2018-01-11 09:13:10.089 [P2P9]  INFO    global  src/cryptonote_core/blockchain.cpp:1436 ----- BLOCK ADDED AS ALTERNATIVE ON HEI
GHT 16996
id:     <1ae49ffd0caa6ffb6dc5617a06ae75e7d140303d69e06f4b6b0dcc4bc98383bc>
PoW:    <ea2a23ce73c81d1a868749165264dda8b5bbf8016432fa9aed4aacb608000000>
difficulty:     96764713
2018-01-11 09:13:10.122 [P2P9]  INFO    global  src/cryptonote_core/blockchain.cpp:1436 ----- BLOCK ADDED AS ALTERNATIVE ON HEI
GHT 16997
id:     <78581a7ea3a71082ead85cfde06ebb2a4d7bbfcfef9a0fe69e373e78b807fa7f>
PoW:    <cd989a29103f7d06b1b135ff196764bf490927f5a1dd81421ce45f3417000000>
difficulty:     105117242
2018-01-11 09:13:10.161 [P2P9]  INFO    global  src/cryptonote_core/blockchain.cpp:1436 ----- BLOCK ADDED AS ALTERNATIVE ON HEI
GHT 16998
id:     <03ca43b2dbfefe2b1854c516a4d30a7feeb5386dfe1169c3caf33ec3ff898582>
PoW:    <4cf0c25ca3760cb69e5c664716ceb6bf2aaf5f5a399f8c6d9ca8c1f01d000000>
difficulty:     110587348
2018-01-11 09:13:10.200 [P2P9]  INFO    global  src/cryptonote_core/blockchain.cpp:1436 ----- BLOCK ADDED AS ALTERNATIVE ON HEI
GHT 16999
id:     <0eb61f1b78d810d314769e5e6157bc22894127b5d8e5bc417e99091005a7b858>
PoW:    <bcb8cec71e4d6b000bc34deeaabba36605f759eab3ad2d0cb45d5c692a000000>
difficulty:     93615476
2018-01-11 09:13:10.246 [P2P9]  INFO    global  src/cryptonote_core/blockchain.cpp:1436 ----- BLOCK ADDED AS ALTERNATIVE ON HEI
GHT 17000
id:     <a39b5f88e39231ddebf80f8023724f9b7ee77dbb81e91f19b96933a8cf118313>
PoW:    <cdd57823371698e564c30bf2e6e4a1740c08874baaba317e42b022b31c000000>
difficulty:     67324783
2018-01-11 09:13:10.278 [P2P9]  INFO    global  src/cryptonote_core/blockchain.cpp:1436 ----- BLOCK ADDED AS ALTERNATIVE ON HEI
GHT 17001
id:     <f63b50e5962bb7fec38a224f09935177a103c20bcd272a0df24c36559bdc2106>
PoW:    <1186387bc636a097188298d9acccd73e4b7bb0722bc0c24eb2371cd70a000000>
difficulty:     62665212
2018-01-11 09:13:10.327 [P2P9]  INFO    global  src/cryptonote_core/blockchain.cpp:1436 ----- BLOCK ADDED AS ALTERNATIVE ON HEI
GHT 17002

id:     <f87dd7136998a2e56a963ca1387a9738a98c7aa72a6f8a570f9c71397af383ed>
PoW:    <75144bd06a1acb7b1ecb90862a6d6044e89756816c18e9eea135452b33000000>
difficulty:     58400488
2018-01-11 09:13:10.363 [P2P9]  INFO    global  src/cryptonote_core/blockchain.cpp:1436 ----- BLOCK ADDED AS ALTERNATIVE ON HEI
GHT 17003
id:     <8e7a86de29068834e3dd3a498c4717159cceec7b6f0753b65cbf889d564a6982>
PoW:    <84211ee2f3d370fb2ecab21158856d8ccb5f1f183d61a5d21bc84eae36000000>
difficulty:     61138607
2018-01-11 09:13:10.405 [P2P9]  INFO    global  src/cryptonote_core/blockchain.cpp:1436 ----- BLOCK ADDED AS ALTERNATIVE ON HEI
GHT 17004
id:     <038e2b878d3c204b8726a4159d8e2c113900a7180015ea30a48eda138ed043c7>
PoW:    <b38f4d2a670f117cfdeb251fe82cc8b5d8fc936bcc291df5f7812ae841000000>
difficulty:     63855261
2018-01-11 09:13:10.435 [P2P9]  INFO    global  src/cryptonote_core/blockchain.cpp:1436 ----- BLOCK ADDED AS ALTERNATIVE ON HEI
GHT 17005
id:     <9249ac65bef9acb1b4375a50d9857627ef9f78961fe8fdd6068621572a174bd6>
PoW:    <dff7efc9cd65b0867be8f67c18368bd190bf286e3d33da4fb3a03bd212000000>
difficulty:     67320973
2018-01-11 09:13:10.467 [P2P9]  INFO    global  src/cryptonote_core/blockchain.cpp:1436 ----- BLOCK ADDED AS ALTERNATIVE ON HEI
GHT 17006
id:     <1d9687780f854ea310bf5d7ea5e756be167fa043d81a58c3a3e602e08e7ad920>
PoW:    <6126e42f0b6591db74345554fac27ef2fc5f115c5d680889485ebc3d0e000000>
difficulty:     70401048
2018-01-11 09:13:10.514 [P2P9]  INFO    global  src/cryptonote_core/blockchain.cpp:1436 ----- BLOCK ADDED AS ALTERNATIVE ON HEI
GHT 17007
id:     <6360ac8eba51cf1215461f304635e01935345f7fb7fa0c1ce3263ac131b5cb35>
PoW:    <65b535014f8befdd9b503e1aa0b362e81f38b8fdad6d9751c2302f6e10000000>
difficulty:     74010875
2018-01-11 09:13:10.550 [P2P9]  INFO    global  src/cryptonote_core/blockchain.cpp:1436 ----- BLOCK ADDED AS ALTERNATIVE ON HEI
GHT 17008
id:     <77d6b92961146232ba36ced359f6dbbf3170299b530d2c7d0da5c398f7713d23>
PoW:    <af18c3b12f0455a0b0fd88cc42be6ff9b877126e524f9ef8bb10fc1a2d000000>
difficulty:     67999998
2018-01-11 09:13:13.604 [P2P7]  INFO    global  src/cryptonote_core/blockchain.cpp:1436 ----- BLOCK ADDED AS ALTERNATIVE ON HEI
GHT 16996

*/

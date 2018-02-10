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

import "fmt"
import "encoding/hex"

import "github.com/deroproject/derosuite/block"

// this file only contains an altchain, 13 long
// which can be used to test out consensus on live data in house

var block_16996 []byte
var block_16997 []byte
var block_16998 []byte
var block_16999 []byte
var block_17000 []byte
var block_17001 []byte
var block_17002 []byte
var block_17003 []byte
var block_17004 []byte
var block_17005 []byte
var block_17006 []byte
var block_17007 []byte
var block_17008 []byte

func init() {
	block_16996, _ = hex.DecodeString("06069dacd3d20516417000452bb3c27c028024a1a21b2d8264a13967cc2b070563b7cb2ecc8236240215db02a0850101ffe484010192ca908deef30602b7840aca58a897165d0c4c90a6e1519d635808f5f32d0dfd8cec545751f9719a2b0185ab9544cf529130c7557fa1a7d3ebc0a25f832ca909bca1e67ca1fb38b472fe02080000000007472ac70000")

	block_16997, _ = hex.DecodeString("0606e1acd3d2051ae49ffd0caa6ffb6dc5617a06ae75e7d140303d69e06f4b6b0dcc4bc98383bc0fb1aa6002a1850101ffe5840101dd91c1f1edf30602ec337478b685134eff525f3cfbf18dd46de2824fa71a2b822d855376ec04b7f82b01f6f9bcac1f8bc3358c0279588a46433e1fe6a11e97eaaebfaee98dff85a621a102080000000001f682740000")

	block_16998, _ = hex.DecodeString("0606c8b2d3d20578581a7ea3a71082ead85cfde06ebb2a4d7bbfcfef9a0fe69e373e78b807fa7f9662013d02a2850101ffe6840101def3eaad9ba00702bb5bb6f749afdd7d64d8b3b4fa2b8a44c77bf637d0c13d3a994b1d269d8780f82b013eeb27eeff0bd10d40f5892504ec95d81399d150cefa504a0fbacfe6fdc8a563020800000000a0472ac70017abda0ac60520cf99e8b28612349ae639fcbdcc4b1ee256bd064660e96029ba63b0450e0e92ee7af62ca755e2f0522e68a4c11497c1179e09dad0a807729455abc27d43eefa50e1714eb0714e45742c7c6c4dd718b494d3290177dd18172a5acc03d87bf62b5c7d6980a78a238061894f2baf3e3929dc006abcd3622f80adc7dcd900d779234f61b29f6f98ba0956a08082607d6c95ead01c252a5086b413a9fd3a0dd65e7e2005395c69a0a676af4048a69e4191b5948d1cae9cb548ef8bc79dd8e247ed59b2b67e7ff1d7129de06038c388bc55581585b178c440eb9d8e60ab843859096066f80ca797d71db0b481d27f64791dc4cfc621ab462d455c94f7424ab37e5a8ca54eca36d5c0dc096cf85b3022f5c5461f704eea1b48b625189e13575dc5528ba9321dea609fa0574b36a4225c7c850f6b78ce68b2a6d6f001bf9874400f342178ef174768af48d2f3bab1cf647684315321fa09ee984b18667712a37f8767923e0f43524a27b4744a2a189ab42b0eee3505aad99afa25952a9debf3fb8f643488edc07d19bd4ce931e82405a22f3f65aa2500146eeafe7aa2785e9e25f5dc5283b684e29582242db611927cdb6a33636b916a6923f56afe2233bc6cd9e3fb4c999eda97345ce93f6af87d917a110c24b885e40b1bab0213a3dc13ea61c256f8cae39eda04a7d4f234a4b8a664834135bd01169b5ae91683a0b11e245b7e3ac9088f863f260c9697b420a80b9e1ec0713b230a2ae4b2edb9f8fad5616c670ff266a41b2006fdfbe835682a6f97486c2df19c3a1052202b74e0d65a59c2fada5e48344a33d0b1d97d1b7fd261a3f8226644c8ca8127fd3a4347ae5bd5b09a9bbbfcb1026da1a2ef4521247d53c0d71abcec48f9d6112f6dd43627aa882a53d4f7e97b27e52c57b66eefa55d2e855459fb46b2f273f564f390fa3f005281716c770f50afbbc1946e1516e6bddd2ce0ea826a4a6f433cadfeb962e4b97964737a3bada86e6ec67b36b94873402dbf72fd53f7f7f5c8804ddb82069f41")

	block_16999, _ = hex.DecodeString("0606d0c9d3d20503ca43b2dbfefe2b1854c516a4d30a7feeb5386dfe1169c3caf33ec3ff898582ed0300d702a3850101ffe78401019fd4ef8592a007025a5d066ac8f5772be60aad59ac33813937efb7edd090006d83f7b625d56deeb72b012c28df62dfe6eda5c5d3113f8d1f3de463f99937e57369ce44b26580444dbf1b0208000000000cf68274001774149ceb2f51f2b8a71536e1e1ddd57a141cdb19d9dea33bf10b09fd336617b0b0f82f931a8147cdc44dbaa9d95425c9980dcd5ab9065892fbeed204902e339f31ca3214fdecac929eb7a4d09fcd799bbe12c5adda9740fa20e4a58178bf32fa014252ccdfb7f5e97838ec2e142eb7ba77115ad0ce34f875e8d0e7d15a2aaa4ddd07a95dd7b66f7b3354c455b712014acf488051700d91869abd0c386ebb40c8c958b296459577ac81d472c1b7c026cfbf355bf86c603fc98615b8f3ae17d17cb4f8edbbdf0bdaf1f269651ef74c1e653d84ff5c597031a008e58f68a04db944da76a0ee549dbadbf47bffa9d729a7747226496b7e213f0729639eebeb6f48108a82a1ca79591e9efd8c6c020be08b02e76d2ebf61aa3f21b84f90bf2a618b477bb01c4b62bd317b4cf92b846d81f8eee205689e4f084e189f3124e697877f67d857f478803f81c1f8b7b893f3f8af830833600d75c43e61733c648cc4a524b50e32232cf5e0e6d9b5c4e539e3d33578aa9c82b54e549e92b1e04d16f3f9dbd156163b6fca1ce4c1572482f2d7dc16634043a7b828488f0b8e5178ddedab49bba1b6646110f2437e83782f50dcffedc67ea4f59971681a3e0f552de091bebbd2c0be66fd022b64d3c61e450513bb4edd1123ef43b1da771bcd94141b29ca933db89499001acb38098e666a8c8d426692ab16547009db6f55a1badc84ea714fbe9c6d63cbfdba522c4d9bac212fdabce699691c646e50fe77e6b6110174d4ff592de728b309cd842af835b80b5966df29fa10d5341c7c3637922b7dfa7ec00937b17522a0136067fe7823bd1808ff8850556584f39b61cc774871a4ff3adedab05cad0e401308b0e9ecd231a4d50f85bc50babfd3701256c7dfaa5b86f37cbe7dc71ce59eb3fbfc0afc6a80b2a4450673f9a44772eee79df273d1bdd43faf13f813dc222c7c88d371504d721be75332b20e3278afefcdb1b4c5a36aee6ae4a5fd7aec7e4f79daab35affd5f8fa2edbd40b83e761b8d59a74cf87e77944244539b")

	block_17000, _ = hex.DecodeString("0606a1cfd3d2050eb61f1b78d810d314769e5e6157bc22894127b5d8e5bc417e99091005a7b8588f81029602a4850101ffe8840101ebccbfbbe5f906023c944822f98c9f59fdef57a196bf1987d6f04f8e0721034f974ceb3e4d88bb2d2b019c25765d6603202013f4a89d8ab60e3149ef064aee4a8feab0f936115a716a5502080000000004f64c0300036c69e35c2acfc877af5d1eaa74c70606dbc0b61d84f7e9d497fffaa1af1b6d169027c337be0ddbe65cbe3e21586915d4f39627a96397a40e445be458ba739200ef54f28c61a031912cd49d1a8c8ef1ccba64e23854dbb3328a2551de2290cfc8")

	block_17001, _ = hex.DecodeString("0606f5d7d3d205a39b5f88e39231ddebf80f8023724f9b7ee77dbb81e91f19b96933a8cf118313ac15007902a5850101ffe9840101f1dc8d83edf30602b41caa95b2a8f18f4a776d13b501d9c75690cc8eb08214a2f56ade005667b1c32b01b8ebd18f15c4d2028b057b25561c5802455cb0e8628036448d87f8aa917a866302080000000003472ac70000")

	block_17002, _ = hex.DecodeString("060684dbd3d205f63b50e5962bb7fec38a224f09935177a103c20bcd272a0df24c36559bdc21068e8e029502a6850101ffea840101e4a8bee7ecf30602613f808c5bbb29eb77f807aaf7a4f155a40b60c858a1eac744f110f182ff236c2b0179be02b5802845c93eb47dec53c476b3ab6c9d416a098ce0c94a0a108f87b6c102080000000003f64c030000")

	block_17003, _ = hex.DecodeString("0606a0ddd3d205f87dd7136998a2e56a963ca1387a9738a98c7aa72a6f8a570f9c71397af383ed76080a1502a7850101ffeb840101c6c7edffe9f506021c41a23fd5a320a69fa4cb7d2ed436ccbab272e00a87a8e5e62773d2ba068af02b01f760a7b6c9ab5c7a93980930048314a0d030453bff3f17d480e2a47db4423d6202080000000002f682740001a3da5b87603312186d940ef790c3cd3d46ae5533a4118b40949d902d3a190583")

	block_17004, _ = hex.DecodeString("0606a8dfd3d2058e7a86de29068834e3dd3a498c4717159cceec7b6f0753b65cbf889d564a69823c7c008a02a8850101ffec84010197c39fb0ecf3060235527c1f4ad20b7003a4a714801d5e0fc5c44538256b2c80e4c98a75be7424ed2b01296ce3ef0900a4d99c574f918e11b1e72e536b8f3a69702bf307800b8852037a02080000000002f64c030000")

	block_17005, _ = hex.DecodeString("0606c5e1d3d205038e2b878d3c204b8726a4159d8e2c113900a7180015ea30a48eda138ed043c7714e05ec02a9850101ffed840101d691d094ecf306020b9449d7678bc9eeaee5d73f2d56b87e07c7d44172c8b994e009bf2a4c4af6752b01fb004518d60291251a1a576677f9e2bcf5233079c7a0b2b42190b7cb8fb6ff0502080000000002472ac70000")

	block_17006, _ = hex.DecodeString("0606d9e7d3d2059249ac65bef9acb1b4375a50d9857627ef9f78961fe8fdd6068621572a174bd61c1e006802aa850101ffee84010183e180f9ebf30602edc679784518be061a918062d7b11544948d091e5f06d389d849156100d47b5e2b01a9128d72db29bc746278350850ba4d12a904dc565a03265c599c30b5b34e921a02080000000004472ac70000")

	block_17007, _ = hex.DecodeString("0606a9e9d3d2051d9687780f854ea310bf5d7ea5e756be167fa043d81a58c3a3e602e08e7ad9209929090602ab850101ffef8401019fb1b1ddebf30602c7b8a1da31ed781b3153245337e5c34aab53055ebc64a0de8a7c336af1a233ce2b016cbf39ec1ad5079ec981f8ce5b0e59b2e014199c0476496cf404a1323adb6d4d02080000000001bcd9f60000")

	block_17008, _ = hex.DecodeString("0606f3efd3d2056360ac8eba51cf1215461f304635e01935345f7fb7fa0c1ce3263ac131b5cb35ff94049c02ac850101fff0840101aa9fddd9e3f90602a96b487a44e654d507ce0874bf5464b6bc7e6ca41af6616b2e3382a95c6ca0c82b01ba0c32ea6c5ca20178026281d05527d6bb0c55b744df551a7dca1bb836f69cb802080000000012f64c030003c58f0769da85da444823d8d6839b16253df8f09ffdddc13dfdb9a2b6fbc8628e83cbaf489cf66fb1c6b0c1b8999fba800cf9fce5750dbc167e359c9b0dee3fdc1492c83a04df176ebc184492194ebe64191d471f377132dcafb09724b8176b4a")

}

func (chain *Blockchain) Inject_Alt_Chain() {
	if chain.Get_Height() < 16997 {
		fmt.Printf("cannot inject alt-chain for testing purpose\n")
	}

	chain.inject_block(block_16996)
	/**/ /*      chain.inject_block(block_16997)
	        chain.inject_block(block_16998)
	 /**/ /*      chain.inject_block(block_16999)
	 chain.inject_block(block_17000)
	 chain.inject_block(block_17001)
	 chain.inject_block(block_17002)
	 chain.inject_block(block_17003)
	 chain.inject_block(block_17004)
	 chain.inject_block(block_17005)
	 chain.inject_block(block_17006)
	 chain.inject_block(block_17007)
	/* chain.inject_block(block_17008)
	*/

}

func (chain *Blockchain) inject_block(x []byte) {
	var bl block.Block

	if len(x) < 20 {
		fmt.Printf("block cannot be injected for sure\n")
		return
	}
	err := bl.Deserialize(x)

	if err != nil {
		fmt.Printf("cannor deserialize hardcoded block\n")
		return
	}

	//chain.Chain_Add(&bl)
}
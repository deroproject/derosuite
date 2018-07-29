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

package transaction

//import "fmt"
import _ "unsafe"
import "testing"
import "encoding/hex"

import "github.com/deroproject/derosuite/crypto"
import "github.com/deroproject/derosuite/crypto/ringct"

// TODO we need to figure out a way to control the variable with exporting it
var bulletproof_compatibility bool

//go:linkname setcompatibility crypto.setcompatibility
//func setcompatibility(value bool)

// test transaction from XMR testnet 0a65ce0098f48085fe2dce36773d696f98b334d1ca2a2510617843bd40600a8d
func Test_Transaction_BP_Simple_Full_test(t *testing.T) {

	//t.Logf("Compatibility mode %+v", bulletproof_compatibility)

	//bulletproof_compatibility = true
	ringct.XMR_COMPATIBILITY = true

	defer func() {
		//bulletproof_compatibility = false
		ringct.XMR_COMPATIBILITY = false
	}()

	// transaction from height txid 0a65ce0098f48085fe2dce36773d696f98b334d1ca2a2510617843bd40600a8d
	// prefix hash b7588a1d92eca5750bd813c14642a1987a7e2fb967e6f06cb3099d90f358ba4e
	tx_data_hex := "0200020200079de610adbc02d0ef01f1b602a83dcd064eb8868ce416d45559dd43089e60ae8dfde2cb09e08d6847c44a6dcd37ca2f17fc020007b9c307f9870dfd5f962a94b5029a0261726608989857fb47e3c3d98a52266a8f76acf61edfb856b17ea7817e46dddfcd020002585a60ecb7323673ca9309e6a6ef8252ea1a829c60d52824088b47f69c736c200002d6d9b4434eb65f92673094736e024ccadc8c3ef9c5d8e1024602b307d119efbb2101b5f382752dbcee162b9c10302eaeac48559e129b8dacf12f54f79945be37ad7b04d094d6bb039e5e72e0cddb022998b7337225ec926f1a189c78d5bc35a7f5d1e46eff1a7c0653dc8cb55e78ed46d191fd74c9c353b3928c9abcb0e840e5897a3f823f227e08461eeb933cc69490387f5874b29af3746a38536554480258dce9644c8b18c40602d8fdf63985ee192cb6b381b231eacfe059495cc92bca586629199a3c897205717781fb37cfb1bcd97046fdaf821d119a0ada002b0cb6d06cc604b46a58c367d46cfa457eb54581405c9a7246034cf28fbf798d87dba92e281c4afed089f2f12c2e9092ed54cacfab0378e9af3216002c230693b1d5ea1ca5436038bf5f0ea57296a74a20fcc7502fbcc3ad84747e6d2aaf71aa7f04cc6858afb0710e9f1fb8b077c9dae8c2292db654f797be906c0a66aa58368ffe8787d099618e91535a826cc877ad4f02b9501c10a2341d78320a7332e0a75563d1e919143b0913fcc5e66f2d3f9332ce14de1fc56436e5d117cb82fd30f30f828e690eb272b9f080820f111a4a3e2aabcc44c619711a88d5c9842cb43a6b51535e7a7e8a13587db5b70106a4225f609f75c92d94e5e7cc687f81828ab3ec0a0fd68803d1d41c88c719ed4d7d5eb2ee9a4f3918bfe462d6d433395c17a51b01c92d6887c532b8d16501336a4f0a11704e8fc0eca9bc5890bd8caf98e20eb229cebc2b964734def80a0956705d1c97f83167dfef5ad6b54b734ddf9242c19a482dc67f461fd52fae863b5b28c3277e60c5931572f81bc10e4ec09008947da0f786635dd1ab0879715d92112edd1bbdaf54fe52c62f6dddb7e5e1f81cf54bc408782f0648b62304e65030d6d30604b4d36a81d277081b467a8c9d0c8cecd4f6edced8a6d9714b62a895c2458875da29292e82ebe18da4dc3ca32523c9e6decb3cce9e5b711f9aa6d23a3f38771355a2c4b76a0f16014d19f5d9993fedae6520e18d3590749706387a676d712c9b33759a17f2afc51790246413d2d1fd2fdab1e914dae9fa0e3cb36fca70fef5b141a3cd3b7e6789fba8caabc078da8d1d686b46badc3abc5dd6ebeb1c0c5b4a3db3121ec2820d4176ba85a181fa24d8ae50d293eb518191c205897e93edf68b2ffe7e5ee9b2c5ee7d447f7fdf5071594837aa37e9836689d705c2d6187449960e221da4d39a4dcae27a2e1b3a6e535bcb973a66e6e06ed1da200f7dd345335105b61b6657b9a9b6aeea1c6c186421b881baf7d93fdda2e390f885428c7685a703ad36cc6114612b978df4db45022bf26af2e75bee28fce899a1175cf9bf3d37962f1a00929ef02043369d63e5dd7e3126201107f3c10d4db91bae3b4dd4983eec873cd82c4f85d95b22453a150d4d061b0497e7428a510237107eb8ac2705ff92a7b5c8341a49f4bef9a9996d74aab10d6799bd5add51c7b56ced8faabf2a1c66bccb41479d162550e89b7a6bd9c18a7097f317c7578f67ceb11ba22ee39a1b0c48d4b375a0a75f8534760300c1208b524dfe0d4ef7937a12eb796e985e32200706fe48e89424623ec335054d92109739be397516383fdee1abb11189acfb165aa020e7fb51d55a055ae47b43d5222d0905fce04e36506c0295cd56374974aee2985d0babdb24f148a5d74ad5e93dee99efbfc94cac4bfe41eda83205df4b93ece17c7853f185d3849db9886a650cfdf639903fccfc3e75b0c08e39eebfa844ba730179788a6adc04fb18ee04dbf531ef404d33b6ed9b65e9c200f83c79c0030ecbc947bad211ae7a46214c93a5cd5ea9d14a721ceb3a6620abb3f42ab69847cece06d822e744ba3c02f1f3cf6f2a69b48bc67909404104045baae247d7f5cf0ff2dd5edc591d79d449acf8160af9bf6a5c8676928cee0745dd349043ee19b3e64382a0fc0e6e669fa66623bd3172be693eeea54eeaa73bbe61dfc32924ed4e0c59253b54c3a5f8bdcab5d883bccce57cc1ff42dce4747870dfc7b097e5028eb36d1fe4506b12dbddbccf20ee28e398e31769134945d77f237269e2868debd406cce67af5e6992d9e11ad81e019a64d096c8e808ed9c01046b08ab254ecfd0ab7ca3acfa6a75a80098eda1ceaaaa01bbdbee71f40593a0a54cb32c0741bab22417d08b791fbc5ed7299269c6d582cf87d01edd803408c4dda6aa4b86e5bc60c7c7f0bdfbe985d2c9c29625949292b61a9a5068d8a3e38778cee4d792b6a6fe1eea40a3b79dc967ec8f4f1c1487701903b17d370186a3820d6dd363e2fc6799ce6110b86ae165de65b5f192ea6707692a09d2543f73a9f9b8250e2e435371ac098aa030a6a9cbb09186cd4c2acda8cceedd57e35afd78f061354f924bba25896a8740adf9989cd9d2e6445c013adc102717dfa6b0adbff7795013083120def1bd55f0a3386dad51f0ceb027b45e1d78355e42f24da4eec5ed68c3c836aaf7e3721f7032ce69c70412c546a65b47ca8f1c1b7ae9b9e244383ba672b13ac2b3543dba6060ff7785637aefa42685e993e6b84586f4900f3553a4c00481be601db8fcab309990be96c3643b23d7172f59353f88d58218d39f2edc865422c001b6bbc28c30c47aa8fa82a50989d5b8db934cca740832a9f3e4cf7a2a12547a5f0e9a48a6404acf9896abf89b744762f97c4a325d855f446e04fe4abe55c726f2b4fdb71c0033667181920c8f6c5252c53ebc6bebeaefc2b6bd00bed41f9600bce0f5b3cc70798735aa970ab9608e441db26cfa901bf8e13bd278240920f188783297b9f4b06ef943de0d733f46aa186c9270746f0a2a6f3832178ec1f49866b980e5a228e0f2599663d3a7af437c3363e55a0c3b97ca851e4b66e0a0f746dca0dada3ca590f6d54432d7a064e30a641aabe6547d67cc196b1131c7621495213b57df7832b0c5604b4a466f479dd49126359b9d1366f090e317e5e086c0dce7e8d0a5bf02b0080189fabe93581fef06a26bdd95fe32f7916770d63b8d89a41a7f61fc8337b0f785683f38ee553a17201a5742d9b6291e9183847b9a69d00e222041783473d0cdb732cddcd83351f1b8f9eb03008e1ac04a2f50ec002674eb7451b93643b700a8a4770d42b412eedfc405c606e9a3eb046cf954e6b9397bece25282a84d89c0efcdcbe37dcee4c15b5e95d7dae086e7f98ec3a724805e701890a6289d6046d0db231d694b5aaa983af61b055a7edc427ac7146633a26f09021cdd538c452ff0a5006d41349f53f252a563cb0e1d6f34d0d34186520b2c06e4f61fa08fc62a608096eee437a02b5eb19b34ba8f9036970198f70868a6bed011c7082b69980a30865e51b1a5f81e1faf6875c9ceb74e54e1cc6a35b9b9df8ab580f0f1cdd2a9d007698547a8fcecdbe434209cfe32ad7d473e08594d6f392a0413bd203bbee38042f1a2c4716dca5c56fcd70334ca62926352478eee01857f72f934fb22478db0300ce38f0f32b21e19f618f8c390e28e12ce8f744721b0d7d7c301f246a0872097092d1b4558978c293d188b3c34429c418a323db012fd20f292b0ca81e7e5f07f36e420516dffef4bd225b7d80498948134ac510c086406b5bf6443c8ac8660864bda4029734fe3637563838da565489baf82a51d321b2311294aea05bf1fedcb19c0f020ee3e9efb1af7792438a63a8c6eeff081e23270c0d45f786de85fe31"
	tx_data_blob, _ := hex.DecodeString(tx_data_hex)

	var tx Transaction

	err := tx.DeserializeHeader(tx_data_blob)

	if err != nil {
		t.Errorf("Deserialize  transaction  %s\n", err)
		return
	}

	// verify prefix hash
	calculated_prefix_hash := tx.GetPrefixHash()

	if hex.EncodeToString([]byte(calculated_prefix_hash[:])) != "b7588a1d92eca5750bd813c14642a1987a7e2fb967e6f06cb3099d90f358ba4e" {
		t.Error("Prefix hash mismatched")
	}

	// 972f9dfb84c5b2b42db244bcea630f6545a9b49867ae1b0cfc4ec760becdca92
	calculated_hash := tx.GetHash()

	if hex.EncodeToString([]byte(calculated_hash[:])) != "972f9dfb84c5b2b42db244bcea630f6545a9b49867ae1b0cfc4ec760becdca92" {
		t.Error("transaction hash mismatched")
	}

	// check full tx serialisation
	tx.Clear()
	err = tx.DeserializeHeader(tx_data_blob)

	if err != nil {
		t.Errorf("Bulletproof deserialisation failed")

	}

	serialised := tx.Serialize()
	if hex.EncodeToString(serialised) != tx_data_hex {

		t.Logf("serialized  %s\n", hex.EncodeToString(serialised))
		//t.Logf("reserialized %s\n", hex.EncodeToString(tx.Serialize()))
		t.Error("Serialize TX  Failed")
	}

	// now lets run the tx through the entire verification test with blockchain data
	tx.RctSignature.Message = crypto.Key(tx.GetPrefixHash())

	tx.RctSignature.MlsagSigs[0].II = make([]crypto.Key, 1, 1)
	tx.RctSignature.MlsagSigs[0].II[0] = crypto.HexToKey("b8868ce416d45559dd43089e60ae8dfde2cb09e08d6847c44a6dcd37ca2f17fc")

	tx.RctSignature.MlsagSigs[1].II = make([]crypto.Key, 1, 1)
	tx.RctSignature.MlsagSigs[1].II[0] = crypto.HexToKey("726608989857fb47e3c3d98a52266a8f76acf61edfb856b17ea7817e46dddfcd")

	// the mixin information is pulled from the chain live
	//mixin := 7 // we know the mixin is 7
	tx.RctSignature.MixRing = make([][]ringct.CtKey, 2, 2) // there are 2 inputs

	tx.RctSignature.MixRing[0] = make([]ringct.CtKey, 7, 7)
	tx.RctSignature.MixRing[1] = make([]ringct.CtKey, 7, 7)

	// mixring for first input
	tx.RctSignature.MixRing[0][0].Destination = crypto.HexToKey("2f8bf5c9e5cd25999a34589bbc59f13611e21157ac05daada73f7e0b376bc078")
	tx.RctSignature.MixRing[0][0].Mask = crypto.HexToKey("146a5c81f21de421983aaf8c14e1ead7aa2c80774e2c71886b251017b1461bb4")
	tx.RctSignature.MixRing[0][1].Destination = crypto.HexToKey("d88caa2f329c0b6f1d67be041f510ea1f0d0b15cece4c37a6284a7ea383afeaa")
	tx.RctSignature.MixRing[0][1].Mask = crypto.HexToKey("42f758b7e05844157e3b18ad3839432faf4f2339439f4852d1de0993cdf62fb9")
	tx.RctSignature.MixRing[0][2].Destination = crypto.HexToKey("9fa324d13b62381ddc141b5983b3dbb25b7dc0677bda80ff54e7113bff891616")
	tx.RctSignature.MixRing[0][2].Mask = crypto.HexToKey("d2d16dd77140b9f643fb1acc8b8d7307e4b2159da9ff19463da18c97de70f5d7")

	tx.RctSignature.MixRing[0][3].Destination = crypto.HexToKey("ed87ea02ec7f40e7359c20d91f9b7029e640fa64a0645968fc299a33cd3c0d44")
	tx.RctSignature.MixRing[0][3].Mask = crypto.HexToKey("8d132a45b65e7bd2c101db5d84b50d90daa1a6121c96563d35f11449e9b21e02")
	tx.RctSignature.MixRing[0][4].Destination = crypto.HexToKey("80b283feddfa6a23afafd595279a69c2a082737c88b268de3ade320da14e5adc")
	tx.RctSignature.MixRing[0][4].Mask = crypto.HexToKey("c0cf6105e56030e241307582d3cd853bd639c09371a9cba58141082c4d855629")
	tx.RctSignature.MixRing[0][5].Destination = crypto.HexToKey("d29b9c90c7ecd9d324292507dc0386a217355b831906bcab64140234023c9478")
	tx.RctSignature.MixRing[0][5].Mask = crypto.HexToKey("06e2cf5aa63c1a4cd2bc2cac4475550a488d6999749e604b5e3cea6663783811")

	tx.RctSignature.MixRing[0][6].Destination = crypto.HexToKey("3fc7f7c1305b2fda3f292698b996531a5e99564b46c7a2a9cac73538d63356dc")
	tx.RctSignature.MixRing[0][6].Mask = crypto.HexToKey("2a39243bf9d4cbb682ff8abd1971e7e87e844c4ef34af6c180e723be848d0fb3")

	// mixring for second input
	tx.RctSignature.MixRing[1][0].Destination = crypto.HexToKey("6318f91b61b6b24a93c2d77cfc0bc99b38118aaafcfbb85419a4597b0d55ddf7")
	tx.RctSignature.MixRing[1][0].Mask = crypto.HexToKey("0f60acc9da5a3eeda600c74316be3e952db2d5d095e8099f42a56f2c563f83cd")
	tx.RctSignature.MixRing[1][1].Destination = crypto.HexToKey("26cec6e727f65adda84f4a5c85289865b0288b0d23e72d4a382907c6238b0013")
	tx.RctSignature.MixRing[1][1].Mask = crypto.HexToKey("30eeeb19d7c65efc19937314fcd2172d01a9e66280f287df0dd2e36711047315")

	tx.RctSignature.MixRing[1][2].Destination = crypto.HexToKey("25d09bc73dd49c7820d07555020cda2093e0679b3f3edeef366388c9bc06da08")
	tx.RctSignature.MixRing[1][2].Mask = crypto.HexToKey("bd200195ae693c10b7773f5b9a088ad0b4df209e370ca2d8a13de74c01ed14f4")
	tx.RctSignature.MixRing[1][3].Destination = crypto.HexToKey("7c8ea4f6b3c50df0e5bcb15fede9a80ea4a88f64b842c82b36c1f5e036b3cd63")
	tx.RctSignature.MixRing[1][3].Mask = crypto.HexToKey("00133ffd1f3ff1f45165f6f83fb10e43014e5dbdb8035232b00db3c562172c33")

	tx.RctSignature.MixRing[1][4].Destination = crypto.HexToKey("ec6fb8246cc980d20141ff16ab8e9a63bb3a0fd015f7fb4e7dbe14887fbdc8f1")
	tx.RctSignature.MixRing[1][4].Mask = crypto.HexToKey("8deb6e3b614f8fde03882cff681d6994de24d61440b45383a4be206d73988191")
	tx.RctSignature.MixRing[1][5].Destination = crypto.HexToKey("0cbb84cf4ea42fa631a4dd9bfa6c85aab4c123be4d6f54183648dc9b98c2467d")
	tx.RctSignature.MixRing[1][5].Mask = crypto.HexToKey("d8b6c4156b4442d44db860aa76083bde08f5d70b5c5fb917f400c6a3134c564a")

	tx.RctSignature.MixRing[1][6].Destination = crypto.HexToKey("84ebb924918ec9233766ce8e1c7ce206f7fc0e3e9180706b888716643b370105")
	tx.RctSignature.MixRing[1][6].Mask = crypto.HexToKey("c6d2a977b6503fc5e49a1a41b6bc6cc7d6405478d3933b4d89fcaf982bdbffe4")

	// test whether it passes range proof
	if tx.RctSignature.VerifyRctSimpleBulletProof() != true || tx.RctSignature.Verify() != true {
		t.Fatalf("Tx crypto full test failed")
	}
	
	// check whether key image test fails
	{
		
		tx.RctSignature.MlsagSigs[0].II[0][0] = 0 // patch first keyimage byte
		if tx.RctSignature.VerifyRctSimpleBulletProof() == true || tx.RctSignature.Verify() == true {
		t.Fatalf("Tx Ringct should have failed but it passed")
		}
		tx.RctSignature.MlsagSigs[0].II[0][0] = 0xb8 // restore for another test
	}
	
	{
		
		tx.RctSignature.MlsagSigs[1].II[0][0] = 0 // patch second keyimage byte
		if tx.RctSignature.VerifyRctSimpleBulletProof() == true || tx.RctSignature.Verify() == true {
		t.Fatalf("Tx Ringct should have failed but it passed")
		}
		tx.RctSignature.MlsagSigs[1].II[0][0] = 0x72 // restore for another test
	}

	{
		tx.RctSignature.MixRing[0][0].Destination[0] = 0 // patch Destination byte
		if tx.RctSignature.VerifyRctSimpleBulletProof() == true || tx.RctSignature.Verify() == true {
		t.Fatalf("Tx Ringct should have failed but it passed")
		}
		tx.RctSignature.MixRing[0][0].Destination[0] = 0x2f
	}
	
	
	{
		tx.RctSignature.MixRing[0][0].Mask[0] = 0 // patch Mask byte
		if tx.RctSignature.VerifyRctSimpleBulletProof() == true || tx.RctSignature.Verify() == true {
		t.Fatalf("Tx Ringct should have failed but it passed")
		}
		tx.RctSignature.MixRing[0][0].Mask[0] = 0x14
	}
	
	
		{
		tx.RctSignature.MixRing[1][0].Destination[0] = 0 // patch Destination byte
		if tx.RctSignature.VerifyRctSimpleBulletProof() == true || tx.RctSignature.Verify() == true {
		t.Fatalf("Tx Ringct should have failed but it passed")
		}
		tx.RctSignature.MixRing[1][0].Destination[0] = 0x63
	}
	
	
	{
		tx.RctSignature.MixRing[1][0].Mask[0] = 0 // patch Mask byte
		if tx.RctSignature.VerifyRctSimpleBulletProof() == true || tx.RctSignature.Verify() == true {
		t.Fatalf("Tx Ringct should have failed but it passed")
		}
		tx.RctSignature.MixRing[1][0].Mask[0] = 0x0f
	}

	// check dependency of prefix hash

	{
		tx.RctSignature.Message[0]= 0
		if tx.RctSignature.VerifyRctSimpleBulletProof() == true || tx.RctSignature.Verify() == true {
		t.Fatalf("Tx Ringct should have failed but it passed")
		}
		tx.RctSignature.Message = crypto.Key(tx.GetPrefixHash())

	}
	
	// test whether it passes range proof
	if tx.RctSignature.VerifyRctSimpleBulletProof() != true || tx.RctSignature.Verify() != true {
		t.Fatalf("Tx crypto full test failed")
	}


}

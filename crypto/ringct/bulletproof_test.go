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

package ringct

//import "fmt"
import "os"
import "testing"
import "runtime/pprof"

//import "math/big"
//import "encoding/binary"

import "github.com/deroproject/derosuite/crypto"

func TestInverse(t *testing.T) {

	{ // inverse of identity is identity
		x := crypto.Identity

		inverse := invert_scalar(x)
		var result crypto.Key
		crypto.ScMul(&result, &inverse, &x)

		if result != crypto.Identity {
			t.Fatalf("Inverse failed on identity actual %s", result)
		}
	}

	{
		x := crypto.HexToKey("0200000000000000000000000000000000000000000000000000000000000000")

		inverse := invert_scalar(x)

		var result crypto.Key

		crypto.ScMul(&result, &inverse, &x)

		if result != crypto.Identity || inverse != crypto.HexToKey("f7e97a2e8d31092c6bce7b51ef7c6f0a00000000000000000000000000000008") {
			t.Fatalf("Inverse failed on Ed25519 actual %s", result)
		}
	}

	{
		x := crypto.HexToKey("ffffffffffffffff000000000000000000000000000000000000000000000000")
		inverse := invert_scalar(x)
		var result crypto.Key
		crypto.ScMul(&result, &inverse, &x)

		if result != crypto.Identity || inverse != crypto.HexToKey("f5b1c63959b9d0a5c5e29780bb3961288f3d440923e9a2238f3d440923e9a203") {
			t.Fatalf("Inverse failed on ip12 actual %s", result)
		}
	}

}

func TestBulletProofInnerProduct(t *testing.T) {

	if crypto.Key(ip12) != crypto.HexToKey("ffffffffffffffff000000000000000000000000000000000000000000000000") {

		t.Fatalf("Inner product failed expected ffffffffffffffff000000000000000000000000000000000000000000000000 actual %s", ip12)
	}

}

// this package needs to be verified for bug,
// just in case, the top bit is set, it is impossible to do varint 64 bit number into 8 bytes, if the number is too big
// in that case go needs 9 bytes, we should verify whether the number can ever reach there and thus place
// suitable checks to avoid falling into the trap later on
func TestBulletProofHiGi(t *testing.T) {
	tests := []struct {
		tHi        crypto.Key
		tGi        crypto.Key
		tHiInverse crypto.Key
		tGiInverse crypto.Key
	}{
		{ // 0
			tHi:        crypto.HexToKey("42ba668a007d0fcd6fea4009de8a6437248f2d445230af004a89fd04279bc297"),
			tGi:        crypto.HexToKey("0b48be50e49cad13fb3e014f3fa7d68baca7c8a91083dc9c59b379aaab218f15"),
			tHiInverse: crypto.HexToKey("15c3fd7aadf7bf69548131398470870494cfba1faca441ba7281e590f33f450a"),
			tGiInverse: crypto.HexToKey("6b27f18b95315a65bc19b3446b618aecdb6df138cc9c9579d8d07f3489e8070b"),
		},
		{ // 1
			tHi:        crypto.HexToKey("e5224ef871eeb87211511d2a5cb81eeaa160a8a5408eab5deaeb9d4558780947"),
			tGi:        crypto.HexToKey("df01a5d63b3e3a38382afbd7bc685f343d6192da16ed4b451f15fddab170e22d"),
			tHiInverse: crypto.HexToKey("594cd3a833a2bcd4ea6eeb61644910f6e57f969b17a121e01076a87101a4a802"),
			tGiInverse: crypto.HexToKey("c818286e1e6ada27390332873057cabf7e52af27c215e37aa42c25f572f25a0a"),
		},
		{ // 2
			tHi:        crypto.HexToKey("8fc547c0c52e90e01ecd2ce41bfc624086f0ecdc260cf30e1b9cae3b18ed6b2c"),
			tGi:        crypto.HexToKey("7369c8d5a745423d260623a1f75fae1fb1f81b169d422acd8558e9d5742548bd"),
			tHiInverse: crypto.HexToKey("5d8961bbc98dda1ef118022c3145eeb965b12ce6df3d4e39533bc3ae1f1c5f01"),
			tGiInverse: crypto.HexToKey("bc9508b8c3a66c3e88403bef90bfdbb8003b948811d520649febaa355a3a2009"),
		},
		{ // 3
			tHi:        crypto.HexToKey("9f11044145da98e3111b40a1078ea90457b28b01462c90e3d847949ed8c1d31d"),
			tGi:        crypto.HexToKey("81c07d2bd8771eb4bd84155d38d70531fe662b78f0c44a9aeaea2ed2d6f0ebe1"),
			tHiInverse: crypto.HexToKey("3331e9b31bfdacb0e6f27477134ce6ccf84389a053a4ae16a2043a24c9108d05"),
			tGiInverse: crypto.HexToKey("4e3aa08d19830405d3accf14080945297ae3df6bf72ab5f03e7ecdc8b707dd03"),
		},

		{ //4
			tHi:        crypto.HexToKey("179637ec7565f76fa20acc471b1694b795ca44618e4cc68e0a46b20f91e86777"),
			tGi:        crypto.HexToKey("0896c5c22f0070ebf055dfe8dc1cb20542ef29151aa0771e581e68fe7818ef42"),
			tHiInverse: crypto.HexToKey("63532b356cdf25f390660eaa8c033d1595dcd6a7ecd32ad2833b76329b420f06"),
			tGiInverse: crypto.HexToKey("1de6ccd0e63ca6be13b4489eff77add6e0a249f7d1984efc030b23b973457c03"),
		},
		{ // 5
			tHi: crypto.HexToKey("251dad91f0d5d451d7e94bfcd413934c1da173a92ddc0d5e0e4c2cfbe5925b0b"),
			tGi: crypto.HexToKey("35c8df1a32aeceedefcbdf6d91d524929b8402a026cb8574e0e3a3342ce211bc"),

			tHiInverse: crypto.HexToKey("fa6b6b3ee8e6a17411cf9402b3c3721912076cf97cdf4cfdf7c41bc225e46805"),
			tGiInverse: crypto.HexToKey("f164117239342c55bf11993a30beed93e33fb242b2769ac526e1e735d5dcde09"),
		},
		{ // 6
			tHi:        crypto.HexToKey("889c8022f3a7e42fcfd4eacd06316315c8c06cb667176e8fd675e18a2296100a"),
			tGi:        crypto.HexToKey("d967bc14e7abda6c17c2f22a381b84c249757852e99d62c45f160e8915ec21d4"),
			tHiInverse: crypto.HexToKey("cab5b9ec75a161fa3d11c44735655495ab11e9406037b3d1bb217a147f6efd06"),
			tGiInverse: crypto.HexToKey("e9714278818e7cf4b2d9c6b4d1cc8bccf553887648e1e3657caddbe39c328505"),
		},

		{ // 7
			tHi:        crypto.HexToKey("d34206fcf444357be1e9872f59d71c4e66afdf7c196b6a596be2890c0aea928a"),
			tGi:        crypto.HexToKey("c8a3831d7c2f24581ec9d15013dfccb5eba69df691a08002b33d4f2fb06ca9f2"),
			tHiInverse: crypto.HexToKey("a86a0a12d9def1320a512852fa994d61ba09978fc6de3e465a2966fceb260b04"),
			tGiInverse: crypto.HexToKey("ba0b9046f2f1f9d1e08fe2afba05fa4670e1af10e7ddcfbbf6c0afcda3028a0e"),
		},
		{ // 8
			tHi:        crypto.HexToKey("9c69d2c4df3b9c528bce2c0c306b6291dea28de1c023328719e9a1ba1d849c1b"),
			tGi:        crypto.HexToKey("9cfbc70db023a48e4535f5838f5ea27f70980d11ecd935b478258e2a4f1006b3"),
			tHiInverse: crypto.HexToKey("77a6ec9bdaaad86b839b3637d5b259ef8e1b27dd4b3747be282e9b5952b3ef0b"),
			tGiInverse: crypto.HexToKey("e1477bb03d1c603ffb0aeea5f3c341a2a03dc4fbb2bdc2d516858e6ad08edd04"),
		},
		{ // 9
			tHi:        crypto.HexToKey("b446bc0b0d3776250dd66d9727c25d0efeb0f931fc537ab2bd9f8978216f6eb6"),
			tGi:        crypto.HexToKey("2da6387292259e69ac0a829ef347699896728c0cc0cadc746dae46fb31864a59"),
			tHiInverse: crypto.HexToKey("0c47d3d83a7043a2e61257dc91c851c2162319af19f104279f357ec83b31b30e"),
			tGiInverse: crypto.HexToKey("750b0146dc5f5f5d398349c3af33419afc053a2516e09baecc708d46dcbdfb09"),
		},
		{ // 10
			tHi:        crypto.HexToKey("e423fae0d374d34a20694e397a70b84b75e3be14b2cf5301c7cbc662509671a5"),
			tGi:        crypto.HexToKey("a5b9a1549c77e4cf8ab8b255a3a0aefaa4cad125d219949c0aeff0c3560ab158"),
			tHiInverse: crypto.HexToKey("4f7077a917981a304ad1bbdbfcef6e93e411c63f30c69220a4ff1f079e5e790c"),
			tGiInverse: crypto.HexToKey("7ef1f7fe1c3b391dc07746a59de37d9759f9216b86db8dcc643180c73737bc08"),
		},
		{ // 11
			tHi:        crypto.HexToKey("e593736f6113c3f288ec00a1cc2fc7156f4fffa1748e9b2c2ddf2f4303bbfe7f"),
			tGi:        crypto.HexToKey("ed671748a17556419ec942e16b901dbb2fc6df9660324fcbcd6e40f235d75b76"),
			tHiInverse: crypto.HexToKey("0c7737eee7a2f670ac8d4425b6935f59b53b8544d8b68404ddb80afd9d26d702"),
			tGiInverse: crypto.HexToKey("f75ac3ed2ea9bd512a26762d1affb3cc028e087ada0e5b20880ce9c52538b80e"),
		},
		{ // 12
			tHi:        crypto.HexToKey("fcee5e57b3b84206a91bcf32f712c75e5fa5108785b8cc2447998312ca31ab85"),
			tGi:        crypto.HexToKey("4faff61c1905222baf87d51d45f3558138c87ce54c464cc640b955e7fa3310f8"),
			tHiInverse: crypto.HexToKey("0a612e558e7fb5a9e59e9ff244c45e00eff3da506d39a6c802ef76725755c709"),
			tGiInverse: crypto.HexToKey("5149e73f8a344e271e9c0f599b489c84e440ec9bd4491a5c09f94d8bb0c71604"),
		},
		{ // 13
			tHi:        crypto.HexToKey("00c82c62684539a27001fb17f2a5649db2e2d64b6b88f0d681009ae78eaece9c"),
			tGi:        crypto.HexToKey("3b13dd7b247319e13ce61995bc771ee1ede7363599f08fc5cfda890ea803e0ec"),
			tHiInverse: crypto.HexToKey("6e4fa0970f177b5b034f51fdb43b2b71f7b1139a4fd21a18faa5f16672bcce01"),
			tGiInverse: crypto.HexToKey("3a7c0c2efa72ac4d133028c37b9e138ebc0f83d2256f9dd054f86dbcba31eb0e"),
		},
		{ // 14
			tHi:        crypto.HexToKey("7357802c6c1cd81ef6248689854089aad694473391bad618ef01dfd680981a78"),
			tGi:        crypto.HexToKey("a70a97707e905629a5e06d186a964f322fffbaa7ed2e781d4d3fede07461f44b"),
			tHiInverse: crypto.HexToKey("f7b521897d6d7755ab8b521946609a152b70e9fae7059d5e0d4d41a7bf83aa0b"),
			tGiInverse: crypto.HexToKey("7f3f8601425eaa2a62c4d9ae515ed6978ac05946253fa17abd67102ae5515d0d"),
		},

		{ // 15
			tHi:        crypto.HexToKey("9718e9d7caef063deb2d675fe843ea634dcf9677c1d3ee92513971b724c788e4"),
			tGi:        crypto.HexToKey("2d98dbcc0caa2055146e13f50ecf75491dadd36ad2baac56bc08562ec66ce110"),
			tHiInverse: crypto.HexToKey("15514af18372e46f69fffe1f883e50c243c7c548f0c9db44e3171e1d76f18c0f"),
			tGiInverse: crypto.HexToKey("a8874a626a02903569a46c9ae3b9fe7ab6c45d559c1b9cf5c3e542af3ab93f04"),
		},
		{ // 16
			tHi:        crypto.HexToKey("107a4240fe26e5fb36cc007e7658964882f769f18c786ab152f25c5d2ae472f7"),
			tGi:        crypto.HexToKey("b544831dbd34c6c252958151c49a734c6e625e42608c005e797edb6d0a8934b3"),
			tHiInverse: crypto.HexToKey("b24f2c767d16499fc0a379d81da087c7e6b55d5bd290f27547b5f0ed0a556704"),
			tGiInverse: crypto.HexToKey("bd4d79bb42c9818799f0044081278a4bbf4a651815ed33a961d42267dc8d2c04"),
		},
		{ // 17
			tHi:        crypto.HexToKey("1e4013c4b0c5787dc1d78bdc8d52331039af4124112ee9346f110a4e8118e864"),
			tGi:        crypto.HexToKey("24a0e4d31cba015783501ecdfa7a8ebae3a6bfd32e6d1a3614b11183c80980d4"),
			tHiInverse: crypto.HexToKey("35bea42dd7531a507e21e619bb28f789eec8b2f4ddfff3f11acaaf631ed9b20c"),
			tGiInverse: crypto.HexToKey("24bd0f9722793656f3ef24b430057ca291a47ecd1a0c9dbc212e82271ed2580a"),
		},
		{ // 18
			tHi:        crypto.HexToKey("115d49b082c83851d4d5e110a4abdaddbda9b0227f5b26bf52d5a22525235972"),
			tGi:        crypto.HexToKey("546cc3ee5db47bfe9705aa95e2da29f228230353917e5d2b1932fe482fbcfed7"),
			tHiInverse: crypto.HexToKey("b4c442dc88edb2f7acf18595f43aa3b401f6582e25266335ad7885dd6008770a"),
			tGiInverse: crypto.HexToKey("af236b59512825382fdafebcdf6242fa734abbe5a227d308650163bdeb145802"),
		},

		{ // 19
			tHi:        crypto.HexToKey("843de91d99d0091f17f4782d4feb2b760cd58b6f2476e8b02d908a1515078aa8"),
			tGi:        crypto.HexToKey("134d556d0c27f6cc6bf3015c06611625739d889c5789fa75b3c83969cb88b1df"),
			tHiInverse: crypto.HexToKey("10551c2ed6d7a9d9d7c3ddebb0b8e4252c19bae51958c581ecfbc9d17ba0f800"),
			tGiInverse: crypto.HexToKey("c3aef5dae154085377364f6fe2e3afd678cfdd65df439541488beedbde923003"),
		},

		{ // 20
			tHi:        crypto.HexToKey("08aa3a565efcb7169fe0cbf72c12ce1750f2861fb6c6851613cbe974efc1684a"),
			tGi:        crypto.HexToKey("01c0aca470f665eb7182e072bca89bc669ffe5b0296fe21343a8c327c8a84175"),
			tHiInverse: crypto.HexToKey("464e2ec999ff6c74af9f3571d58363bec03597681f8711ee61a936717b72fe07"),
			tGiInverse: crypto.HexToKey("df6109285fdc6c155341da61831650c29d1c8efc0eb6165d89e42830b6b08d0b"),
		},
		{ // 21
			tHi:        crypto.HexToKey("ebbe8b8a522abbe78277d0daa7892d9da87c27becd3ec03895233ad466318c44"),
			tGi:        crypto.HexToKey("02855a25ccb75b2f8eeac5d1db25044b0aead2cf77021ed94f79f3001e7b8e9d"),
			tHiInverse: crypto.HexToKey("1adca9faf1b7baf54387f0da8864ebbee27d8aa4d87585ef93d18960e4cf110e"),
			tGiInverse: crypto.HexToKey("1189568ed225a64d802e94aaaab75d07978bc4bbef1a8aa2b21ad00dd342f404"),
		},
		{ // 22
			tHi:        crypto.HexToKey("3c4d6d5cf12eba7dbd3e84329df61afc9b7e08fc1332a682344273396ec7dcdc"),
			tGi:        crypto.HexToKey("b7311db28c45c90d80a1e3d5b27b43f8e380214d6a2c4046c8d40f524d478353"),
			tHiInverse: crypto.HexToKey("64a4a2164cc01c3cb8efc0b3562e4f2c76ff147da0fe32f461957bac23306f03"),
			tGiInverse: crypto.HexToKey("3427db45f598c4f9053607dc10aa068f6b520b5a501d75a43817436c0c55d40f"),
		},
		{ // 23
			tHi:        crypto.HexToKey("beae48ff70a19a31d662443cce57f77afe050b81224860255bcbc8f480c43cfd"),
			tGi:        crypto.HexToKey("204d01a17c4fb7b18c2f48270150db67d4b0b9ce8786e03c9550c547fb18029e"),
			tHiInverse: crypto.HexToKey("1ff56806c6e353befc540eec75746340588e67c4f02c36193736b7f81901b20c"),
			tGiInverse: crypto.HexToKey("922f6c5872a48a518f8391daec52745b4fbca0378eb479f5a09c37b916851a05"),
		},
		{ // 24
			tHi:        crypto.HexToKey("ebb1b2a68972b7d3323b0361f3a1142f8b452e9298773def5635c2e2efa3700e"),
			tGi:        crypto.HexToKey("f16e5629e9a1c668e1aa79c7887355f5f51b0cbb1f0835e04e7acc53ac55a357"),
			tHiInverse: crypto.HexToKey("7ac4b1ade3a8eb53e8bf74c87e13aaeff2042b191ca8eafba8a6e96dd7df0b04"),
			tGiInverse: crypto.HexToKey("b8bd98820bd497dadebfe463f8747e787dc47636fb16d6e1f162ceb3bce7e50a"),
		},
		{ // 25
			tHi:        crypto.HexToKey("4cc9e5d8de78967e573582cf7c74977c30b5469b2c0bace8ec259f71ba25c8dd"),
			tGi:        crypto.HexToKey("4197b54c5aaaad47be24dbbc11c1bd3eeb6246542d2f5ae5f4398dd4a7601703"),
			tHiInverse: crypto.HexToKey("6c1b2d214a9db423e096b7e67aaf9742947a1b93f5c3a478fe667df40d7b4708"),
			tGiInverse: crypto.HexToKey("6621a430ca87b86a31091a9d1461af7ea6f04e670487d45bf5d88a3ff8574007"),
		},
		{ // 26
			tHi:        crypto.HexToKey("1c51e5b0241cca7c86f718b7d2c3d457a6e5e0b39f1f39ebafbb0883d427d936"),
			tGi:        crypto.HexToKey("cbbfd59baddd3a7ce6e375e7d90050e271b13f132df85e1c12be54fe66de81f6"),
			tHiInverse: crypto.HexToKey("61cbdffb01aa347be1b04a072c91eb9911e57eb7ef00094f95ecc52156909800"),
			tGiInverse: crypto.HexToKey("fc1cb00192a4f8ffd79040fae6d8ccd20b799195dd4ba6c2291e71bbe9e7d506"),
		},
		{ // 27
			tHi:        crypto.HexToKey("476015ad88b792a031e4dd983757c99aea3912e8f8c2f659de4bc1a2204cea13"),
			tGi:        crypto.HexToKey("8a1c8f696f3e773c7eef57ac1389bd0280d558ea7862f01b641ec6da0efefbee"),
			tHiInverse: crypto.HexToKey("ff427d386a1c9670b0c5a74c127d85dfb4340ed05de57d586633cb0e52347104"),
			tGiInverse: crypto.HexToKey("6a41455a475312ef3d4b8188a7a6557b10d36c85809de0e0d2981335e2893009"),
		},
		{ // 28
			tHi:        crypto.HexToKey("2e4f9ef71777119153639a71ff2417f522fe41b87e9c1cb7669f40f9d685887d"),
			tGi:        crypto.HexToKey("d0509c538a8c3616681d761ae5c6f9d2aaded71890da2496156043082182ec85"),
			tHiInverse: crypto.HexToKey("072546fe3ee867ff88385f67896ce1526ea0a9d04b0378c8087a93132b70c20e"),
			tGiInverse: crypto.HexToKey("03299f0d4ed25e380bf9a5d7f9615212c60f59e3212c1f04fc6caf47df142104"),
		},
		{ // 29
			tHi:        crypto.HexToKey("ff81927aa42eda7f2a696789091033cf5be2fc1f5f3a2de22715eb33d6282892"),
			tGi:        crypto.HexToKey("9c3ae48693f91343d0a5f0ecbb7dec9b973bf213678a653b0d9df510652a23c0"),
			tHiInverse: crypto.HexToKey("5fe6de540f347dc9d2d40890c78195a9f5ee7878b236bb2f2dfc684999d60e05"),
			tGiInverse: crypto.HexToKey("78cc5c2b44f06241139bd1a77084bda5b43bae00fb1a586b30dfaf90ed2b5b04"),
		},

		{ // 30
			tHi:        crypto.HexToKey("2dac862efc7fc6d54c99e6ec6e58c0b64da957e736d30093c867a120d5dbfc55"),
			tGi:        crypto.HexToKey("b8065367924a4cfc786036c066caa738349cf1cda70dbfa85cceb4a09f85039b"),
			tHiInverse: crypto.HexToKey("d0326499aedd5aab24e90eecbb1cee3e32461072213adbcb0272123f03a33905"),
			tGiInverse: crypto.HexToKey("992f0459d57b613966e85e7217f9e2a5520252038efcf80d1d53ce7351c2b106"),
		},
		{ // 31
			tHi:        crypto.HexToKey("03ca276405df4b2dbe6cfe7c2c56bcd2669f1b7d82c9f92991bf4102af6110bf"),
			tGi:        crypto.HexToKey("6f77274fa6e27935bf89ae373a3b5ada5824bd4b2aec222aebd7fee7a482e9c1"),
			tHiInverse: crypto.HexToKey("56d5bda27b89256e51c103fa9f071fde53e4fc197139bc2e1fed6d41945c700f"),
			tGiInverse: crypto.HexToKey("cafa412988c445a9d15d5ede828613c6f2a7e5f048128517da56a33515055b04"),
		},
		{ // 32
			tHi:        crypto.HexToKey("1bf5bdae897f9a064209cf31299653137e865f905c8929449139545ac8253c32"),
			tGi:        crypto.HexToKey("3358eab25f942236f3f4b6ebafe1c3eeeef793836680667c669464c3d4a0847d"),
			tHiInverse: crypto.HexToKey("273cb042f4f86a459d260689d880a7b3b86d8074766fd09d6dccbb0449755003"),
			tGiInverse: crypto.HexToKey("d5f0dcba66ff5ff15327694643e0309d831572699f60c96e5a52e22ac3eb6509"),
		},
		{ // 33
			tHi:        crypto.HexToKey("be19cc8bd854ca7cdb07c2aeba12a14ccfa3085f9ffd9f753980c9d45b7b4e0f"),
			tGi:        crypto.HexToKey("f3024bd5df2aa4aa4d19e551ede93dd075f7953acae53f0f9e8a384e496c5250"),
			tHiInverse: crypto.HexToKey("7bb153c8dec1ac63776804ec354e78b656296f2510a0997618b058f25bf1960c"),
			tGiInverse: crypto.HexToKey("df73e2af7898681697a6ec83a5b6e241279f8ca195caf02de4f5e7dc80a28b0b"),
		},
		{ // 34
			tHi:        crypto.HexToKey("5be46df3ae5c10c189f1dc9ed2592e246bd2449aa0daae458ae8bfbd52f983c3"),
			tGi:        crypto.HexToKey("b07e7617e89e28f953d096ec2987ebd8f3e74d933963b82773d37ab1b7a3601d"),
			tHiInverse: crypto.HexToKey("529ec228577dd9c11a81cff064411893eb0c4fbd8a3606cdf3042dc1bbcead08"),
			tGiInverse: crypto.HexToKey("4c9b696409269dca823a21512ff2e7fa1330cd5bd3c385af9c8d2e5a7d604e0c"),
		},

		{ // 35
			tHi:        crypto.HexToKey("de44123726719c08d4c37c8c9b0be17b6b49826136aa7b908531bc91732b087a"),
			tGi:        crypto.HexToKey("c8971334825dd1d67e4c48297292a07a40629675b3e8788efc687385300481ae"),
			tHiInverse: crypto.HexToKey("a650e01d536ecafca1f598b6dc2944c61fffcc510309aebe70210047850fce01"),
			tGiInverse: crypto.HexToKey("ca01a6001a90a4b4ccad3e5080d74d85b9bb4c2fa9c975fcd695a0eee46ea604"),
		},
		{ // 36
			tHi:        crypto.HexToKey("4136030bad7b5b1cfa7d9c98a9dc347a92651f29c2e110aff8897f267c042210"),
			tGi:        crypto.HexToKey("697406d24ef88ebf9ca1972c1d528478858ead85782ed410ebbc1f3da48ba807"),
			tHiInverse: crypto.HexToKey("f97caa07734a11166087d44cad16d16dda5568fb868e88c8faa28eba43b5ed0f"),
			tGiInverse: crypto.HexToKey("c3a3d99951a9ca0ace0cd23e738c0dcf6422a7857b0e122fa0d6157f03466705"),
		},
		{ // 37
			tHi:        crypto.HexToKey("a6b70a313cc06afa2bd9c2911537d609d68bec9432e84b9679527d6abb588ba7"),
			tGi:        crypto.HexToKey("836236aac0a8f08a5029115d57e7ef18cb27cce8d2c157a9f4f5615dcc348aea"),
			tHiInverse: crypto.HexToKey("6160d4e01c33d35dc40dcf1a9dd06e24befb33fd5b19f39d4fad727c4d8d000e"),
			tGiInverse: crypto.HexToKey("45cf9d83e3523354163afb91467030cb5423256c4afc5c5b233dacfdd4d62701"),
		},
		{ // 38
			tHi:        crypto.HexToKey("2bb214987069d80b0abc2bbd68eba0331e3ae5f4106f7fc1e2e7b8d6e5370e32"),
			tGi:        crypto.HexToKey("c80d0f28df33babe39f6ecbd19a4a6afa853aa4da03b6bd7a806229ded76d2c5"),
			tHiInverse: crypto.HexToKey("0779c3914b47bfac6d8028078e169648a7a65d97255a053974ada2edebeeb007"),
			tGiInverse: crypto.HexToKey("faeee7c6105372a2751ecae112dce2d84ef32aebe4d3bab0f54564abf8bb650c"),
		},
		{ // 39
			tHi:        crypto.HexToKey("01cce2a036b68ed354316339f092dec7662bcebdd2066111d16ce55a937e2c61"),
			tGi:        crypto.HexToKey("b9de1176d519a793946792b5417eaf7d2d5126977c5704fc0fcd8e1b2f589b1d"),
			tHiInverse: crypto.HexToKey("b7074e5fbd8b4011021dae44f478b836723a109edb5d688e9b44c9fb897b0e05"),
			tGiInverse: crypto.HexToKey("2fcc6af1de54c50ec228eada93b052c00cb0495c23bb64a5971f629ff92dfa08"),
		},

		{ // 40
			tHi:        crypto.HexToKey("907bc366c885daa37495be671ef6c2f2e554ede3b53ce280cbe88a48b9d9740e"),
			tGi:        crypto.HexToKey("418d19dd28f7e94c51a1782d322e03cba478857424497b4a373fde0fbae4ccd9"),
			tHiInverse: crypto.HexToKey("5a3b990a815a383901f9a11603b630b6345df019fe14fc19f2addbaca4d44e0b"),
			tGiInverse: crypto.HexToKey("d4a76ae3e285f7a8d90a2377381bb07735a5aa200998200e6ebca482ab512703"),
		},
		{ // 41
			tHi:        crypto.HexToKey("980ceaf804edcd8c96858193e6d5178bf604cc73bd8faad50d531549993197cc"),
			tGi:        crypto.HexToKey("38cbbfa0f4ad2397eed7f76dc3cdb6b06a36660c0775d391ca47213341f659e9"),
			tHiInverse: crypto.HexToKey("9aee952ecf3c36eac409c9e498322553f67dc92095932e2e72f66ad32df3ec0e"),
			tGiInverse: crypto.HexToKey("3046cd473cb6e5c2d808d1c34c7e8eaed97851124a524908b35a5abd769a430d"),
		},
		{ // 42
			tHi:        crypto.HexToKey("272827216d1af9dcc6e9862a6e53a0a2c73298e1fadc0f9148cbc85ec0567c38"),
			tGi:        crypto.HexToKey("014f70284efaa5faaba4bb8379ce0204f5aedc28268d82438b5b881fdf2dee4a"),
			tHiInverse: crypto.HexToKey("98918a8320d6a86d7e79b9371ec390a72b72918f42c0439ceb92315aa54a5105"),
			tGiInverse: crypto.HexToKey("9c242a37728df7dbe3a45289ffd363f46e637a1426427564396beb30eae76402"),
		},
		{ // 43
			tHi:        crypto.HexToKey("769c2765d654c4269f6ef13947f13c239cbb08b7cf67a25bac030ad1b892c434"),
			tGi:        crypto.HexToKey("d7d40ed13dad57ca929614a63a00fe3a78f33b30b6fd5f39e4437036dced8d87"),
			tHiInverse: crypto.HexToKey("0720595dfcd455a937f74631ff3c622ef4888b364806eb1fa71b9f7c395e2c05"),
			tGiInverse: crypto.HexToKey("0b5d243836d5e7911a50f88ea39029251177285e6cca2a0256cf2ab91d6c8408"),
		},
		{ // 44
			tHi:        crypto.HexToKey("79246449f5328dac3141d3d7c8a9a2540dcac2cbc98e27843143e7d4b96dde75"),
			tGi:        crypto.HexToKey("af43282f43fa14abaf6c8415fc05ee1ad171d81faa467ddfe5e02eb6895e5688"),
			tHiInverse: crypto.HexToKey("05386af72fcbc20a350b492d6e6ac6884f609cb07d5f5aaa2049b9f6e06d5e0d"),
			tGiInverse: crypto.HexToKey("32ae85e5b32d14a54e1a3bfe6e7f491e179206c9b54962858d33bb20fc122804"),
		},
		{ // 45
			tHi:        crypto.HexToKey("21fc70b3280a2a4c5f39287f5d24d7a759ea037b11448739ee2a28fc4b160eac"),
			tGi:        crypto.HexToKey("dec048f6660e3a2fd8bdec602af59590ec4c6eab834cc0dec8621eb510fba6f7"),
			tHiInverse: crypto.HexToKey("76eb27c11e5690d48be5c847be8d8b24cb6cdeb872bfa44736bdb3bd95af3d0a"),
			tGiInverse: crypto.HexToKey("27bca2d58c65819f0239178176b2d90a7829fa58c0da8a704fb929e73866a80b"),
		},
		{ // 46
			tHi:        crypto.HexToKey("406108aee6b580621311fe030bf08b4f6eed3d7d3d8693d3ac524da2b4ebf19e"),
			tGi:        crypto.HexToKey("adf47693c2fd574d8220a2e70e73ad68e4c332488eb8e731fe600d1e9f6b8f5c"),
			tHiInverse: crypto.HexToKey("cfd4f230338ea0809f5f130076c3bcab696e7b018f9e82c74298ff35ddc4980f"),
			tGiInverse: crypto.HexToKey("1e801d444ab40a5b6938a22d72aac050cffba384e682df1eef6c210a81aa860e"),
		},
		{ // 47
			tHi:        crypto.HexToKey("2559dc50ff35e62da620dc0a02edcbe4f398b1bd86ea154b6a9400579e3f1cd5"),
			tGi:        crypto.HexToKey("bf699c18d06bcd73b7cfcef42e68af7ae67fea46e946de6a61faa42c535cfcae"),
			tHiInverse: crypto.HexToKey("9de67050acfd818ab1e2d539f15d3f1ad998f53848fc2eb0d89aecab0c6d6f00"),
			tGiInverse: crypto.HexToKey("20d10f935c4badcfc970674fd6ac3a733c804c36832a5e62523b2058e917cd0f"),
		},
		{ // 48
			tHi:        crypto.HexToKey("7fdc2f10bd8cdb167c0b283f9007e620d9ca28067fe2b015ed657c9153b8443d"),
			tGi:        crypto.HexToKey("aad5334fc1a9bad4a53e57d11c6accfcefd2e8ab44cb12fb2e664fcbdf5c82b2"),
			tHiInverse: crypto.HexToKey("523e2c991a4e5a365781d584d1a52678dfe4362cd2b245c4a81696f521a8e60c"),
			tGiInverse: crypto.HexToKey("22b641e49a0d5f8b7716fd4882c9cc0c7688e139cc6637d80be4a1833f1d210a"),
		},
		{ // 49
			tHi:        crypto.HexToKey("77e8e25ff348f4df78bbc1ce20a7babde40ed2bdbeaf2b5cd98e5202baf7e3dc"),
			tGi:        crypto.HexToKey("1289626ac2a1402bde7a869eb9ed7807338dd3b2ba8237845db96771cc988008"),
			tHiInverse: crypto.HexToKey("f673297255168763de7eeb857e99daa5b2f4153be6d939a29f8c18c3da75c606"),
			tGiInverse: crypto.HexToKey("7a90adab8c987747a23d0d874d69d22fad4a3e723aa0dd057c9891a5ecfbf801"),
		},
		{ // 50
			tHi:        crypto.HexToKey("f18ba115620c51ae8b58b4923b9a8694c93df64b178c4cd2f9f6efc51f458b0c"),
			tGi:        crypto.HexToKey("1acf053d9bd51c0101941c4c26f66aa5dbad3f5354608577f9e51afe743add50"),
			tHiInverse: crypto.HexToKey("a4f882b6190abf2f23505a40b4133f24144cd7800bab4f0bd4d1b1fe9ee9690d"),
			tGiInverse: crypto.HexToKey("b2d5bb3400414f7aa879d1ac179cfc02fa558e85e868c1f126eee652e72ed703"),
		},
		{ // 51
			tHi:        crypto.HexToKey("5ee860a40ac8cec3506ec85b99dc716b95cbb342db91ade4b61e177f60f9fabb"),
			tGi:        crypto.HexToKey("f1b5901bea7beb5ae780b6ece977f65b9c628e1dce0ad1e078c746c2f38d0e7f"),
			tHiInverse: crypto.HexToKey("fb616c745050445e052de23a5242886e6f83e364b99d41b96bbf1a1b70cafa0a"),
			tGiInverse: crypto.HexToKey("69a0d093a0c4757152cc1fd0173b5081e96ed88adfd9a8a9b78600f3f2c4bb0b"),
		},
		{ // 52
			tHi:        crypto.HexToKey("ff2c9badee04cfd741d66d2f26321e2cf50a3cd021f6288863de2dadf8d52d1f"),
			tGi:        crypto.HexToKey("06b088708ae9ac1117e3a37999c1d75a62e9c9e017018e088aebfb378de29c78"),
			tHiInverse: crypto.HexToKey("dfdb55c7bc49b4f270bafbbfe5a0bd6dc72a2b55b53caa22614ace15ede19400"),
			tGiInverse: crypto.HexToKey("fe8a8432ac80f6b723ad3a44564da7322f83eb7d81581474351d7ad0b97e9a07"),
		},
		{ // 53
			tHi:        crypto.HexToKey("8b9f51424305a3d407962963c1d0beeb8113f80307ecc21923947fe8cbaf5c2c"),
			tGi:        crypto.HexToKey("93acf10942584bf558a2d02d751e34f3f484b001e31924cc21848bf0ddaf1f3d"),
			tHiInverse: crypto.HexToKey("c66e744763e5afa5a1b101c28f7517822406fa700979aa82b968bdeb135b5c0f"),
			tGiInverse: crypto.HexToKey("9922d87869e7e2bbc3fa9b331ea3701e049055fc591f7f1e027f78e5396c530b"),
		},
		{ // 54
			tHi:        crypto.HexToKey("05ae6369852199c52a1797b9aff2a9245d7a8b9172d572b4432f63441ff51c4a"),
			tGi:        crypto.HexToKey("8a310049736ff7f049294d8a595f2ca7263a3613840c14b33ef483cdca5bbb8a"),
			tHiInverse: crypto.HexToKey("e16e59b5071da3833cfb0dd11164101a05179198de87febcaaa3c8f91a473e00"),
			tGiInverse: crypto.HexToKey("d6ce5a7ce6d56f5dc6089bedc6b5f778b7163a6a91fc32e81acf8f29b83ab300"),
		},

		{ // 55
			tHi:        crypto.HexToKey("4e270e3b61eae6e13eefe35e85427bc758ef4af4c00f9c77521c0361d299431f"),
			tGi:        crypto.HexToKey("4c7004ccb8f67156267ee35f280db12645de8e552a9312df5769a030a6b46d80"),
			tHiInverse: crypto.HexToKey("4d490dd312c0975fad9672c9415e36b15babb284311b889f2d7151d7ed368f0d"),
			tGiInverse: crypto.HexToKey("8b1146d960a953d027df69c40a3f7a4687abe0db60f5909fa98d89ec6eae420b"),
		},
		{ // 56
			tHi:        crypto.HexToKey("9d8e298c13414c46170a1d82a1380fbafe531ca70184ab8965c4c807060e8039"),
			tGi:        crypto.HexToKey("db2e6c06b3c76c1ada42373b29a0591f39856749dfdfb26681166a286fb4f209"),
			tHiInverse: crypto.HexToKey("94c2f9d861387f087377c8f153034244fc040f839bacf16b52ff8a3fb7a0bc04"),
			tGiInverse: crypto.HexToKey("6bd1cf437e6c56773f4a88ce3f032004603f2da976bd9ccec1093cee5278b409"),
		},
		{ // 57
			tHi:        crypto.HexToKey("fec4615e5909d27ac5ca8041e3f95b27f1c3d4d406a2048b1e6ce1e637cb87c0"),
			tGi:        crypto.HexToKey("7a3b6f8febdbe4413b67b558689c2e7c1d6d6408f46a6094c74b2281e796e1d9"),
			tHiInverse: crypto.HexToKey("62541599436e936cf5139abb7adedef8053951ff61538ad4bb90e52666cfbf00"),
			tGiInverse: crypto.HexToKey("cada930635a917aaa121b1195596e5e1682f628e7b1f38c9bdfe5bdb0d2ba907"),
		},
		{ // 58
			tHi:        crypto.HexToKey("f97d3617d46aeffdd1e813c255fb8b3ef939a2c5fad4d10973c08c055f7913c5"),
			tGi:        crypto.HexToKey("00cc835337a31b5350caa9c444c670f78f866e03ef6ec2cbcbc179974145b239"),
			tHiInverse: crypto.HexToKey("538dfca4affad88db753d066ebe1ec257ff7d6c5f2ff89ad1a63f134d4d40909"),
			tGiInverse: crypto.HexToKey("902fabe3b31e2f26b9cdfa59e6710f1a46e74e972309ecc64b025ea04d708008"),
		},
		{ // 59
			tHi:        crypto.HexToKey("1664589da5145a9c5972f4b212ebf51171d92343833a08953cd80cd0d908904c"),
			tGi:        crypto.HexToKey("b90912bbeef8f576961b5efc69641f7a7151708775b67c9e65ed9bb9f5a87bb7"),
			tHiInverse: crypto.HexToKey("6b26496ddb27ce7b7df71e2923fc823950ba0630460155f294f1185e37e14208"),
			tGiInverse: crypto.HexToKey("210afd245b6ba719276c19df7fd05327ec7867215b7bb5c32fa3434577f0f105"),
		},
		{ // 60
			tHi:        crypto.HexToKey("563edc34294221865633d8cf6ff50444b9d29beb05a47b8bb121cb118d6cb16b"),
			tGi:        crypto.HexToKey("90da203557bed2674055e8a6ab3646c4e1a845ea53d8614ae490065def757615"),
			tHiInverse: crypto.HexToKey("71f13f1892191528bcf18c3d3565ddd02593618d683bd22a4079584fa969b20c"),
			tGiInverse: crypto.HexToKey("87b24583b86a0df91d72070d6c6ae414b5df00b17ffb15566484495a8c186b06"),
		},
		{ // 61
			tHi:        crypto.HexToKey("24c445098aa90e6d5a10eae0a0f3977a2808f79cafe8f8705297bd91ebbf2792"),
			tGi:        crypto.HexToKey("a265f2ab98388029aec3afb5cca3a666ab29b6d2c002979c636a3b41b8837a43"),
			tHiInverse: crypto.HexToKey("10e7e767d0f9f1ddc74f5d36c7b6dc7155c0fa1700c57157fcea8adfd711e905"),
			tGiInverse: crypto.HexToKey("c390a5605e069d19273686bd29d35501f711d0b5376ddf7ba47e49a7b70f4b0a"),
		},
		{ // 62
			tHi:        crypto.HexToKey("a1892cb009db0b7ac351d0353f43fe3aa97192e8b9d7fef5baec415b0ca48c92"),
			tGi:        crypto.HexToKey("2a81d6db55cf406b1f5842b0a887fe6b2bd88e46298ed3ecc3874c9837734633"),
			tHiInverse: crypto.HexToKey("8c2e05720f94692fd89756c91f0c0fba8cfa40c5d44858f896d1cb8cfafe1b08"),
			tGiInverse: crypto.HexToKey("ff0b0e52b8b665bb9c7faff12621c9de667b91540ca154b2b15e2b0b6f8a080f"),
		},
		{ // 63
			tHi:        crypto.HexToKey("0e7cdd78f9246ad254e87ee1b06584b860b0b8800aaee17896f0290cb789b0d7"),
			tGi:        crypto.HexToKey("1fde7a2ff7f104265bbd2d0274c033c7583851001dcdb3ded90a9c0977c1f86d"),
			tHiInverse: crypto.HexToKey("4e0f864794a4f870044dafd3f1a6cb44c9bb84d743f0f8559a4ce61e326ac605"),
			tGiInverse: crypto.HexToKey("ad18f5cd9d1888b0b941d845cb8139007b540b1540d392da239aaf880ea26d02"),
		},
	}

	/* constants from implementation
	   0 <42ba668a007d0fcd6fea4009de8a6437248f2d445230af004a89fd04279bc297> <0b48be50e49cad13fb3e014f3fa7d68baca7c8a91083dc9c59b379aaab218f15>
	   1 <e5224ef871eeb87211511d2a5cb81eeaa160a8a5408eab5deaeb9d4558780947> <df01a5d63b3e3a38382afbd7bc685f343d6192da16ed4b451f15fddab170e22d>
	   2 <8fc547c0c52e90e01ecd2ce41bfc624086f0ecdc260cf30e1b9cae3b18ed6b2c> <7369c8d5a745423d260623a1f75fae1fb1f81b169d422acd8558e9d5742548bd>
	   3 <9f11044145da98e3111b40a1078ea90457b28b01462c90e3d847949ed8c1d31d> <81c07d2bd8771eb4bd84155d38d70531fe662b78f0c44a9aeaea2ed2d6f0ebe1>
	   4 <179637ec7565f76fa20acc471b1694b795ca44618e4cc68e0a46b20f91e86777> <0896c5c22f0070ebf055dfe8dc1cb20542ef29151aa0771e581e68fe7818ef42>
	   5 <251dad91f0d5d451d7e94bfcd413934c1da173a92ddc0d5e0e4c2cfbe5925b0b> <35c8df1a32aeceedefcbdf6d91d524929b8402a026cb8574e0e3a3342ce211bc>
	   6 <889c8022f3a7e42fcfd4eacd06316315c8c06cb667176e8fd675e18a2296100a> <d967bc14e7abda6c17c2f22a381b84c249757852e99d62c45f160e8915ec21d4>
	   7 <d34206fcf444357be1e9872f59d71c4e66afdf7c196b6a596be2890c0aea928a> <c8a3831d7c2f24581ec9d15013dfccb5eba69df691a08002b33d4f2fb06ca9f2>
	   8 <9c69d2c4df3b9c528bce2c0c306b6291dea28de1c023328719e9a1ba1d849c1b> <9cfbc70db023a48e4535f5838f5ea27f70980d11ecd935b478258e2a4f1006b3>
	   9 <b446bc0b0d3776250dd66d9727c25d0efeb0f931fc537ab2bd9f8978216f6eb6> <2da6387292259e69ac0a829ef347699896728c0cc0cadc746dae46fb31864a59>
	   10 <e423fae0d374d34a20694e397a70b84b75e3be14b2cf5301c7cbc662509671a5> <a5b9a1549c77e4cf8ab8b255a3a0aefaa4cad125d219949c0aeff0c3560ab158>
	   11 <e593736f6113c3f288ec00a1cc2fc7156f4fffa1748e9b2c2ddf2f4303bbfe7f> <ed671748a17556419ec942e16b901dbb2fc6df9660324fcbcd6e40f235d75b76>
	   12 <fcee5e57b3b84206a91bcf32f712c75e5fa5108785b8cc2447998312ca31ab85> <4faff61c1905222baf87d51d45f3558138c87ce54c464cc640b955e7fa3310f8>
	   13 <00c82c62684539a27001fb17f2a5649db2e2d64b6b88f0d681009ae78eaece9c> <3b13dd7b247319e13ce61995bc771ee1ede7363599f08fc5cfda890ea803e0ec>
	   14 <7357802c6c1cd81ef6248689854089aad694473391bad618ef01dfd680981a78> <a70a97707e905629a5e06d186a964f322fffbaa7ed2e781d4d3fede07461f44b>
	   15 <9718e9d7caef063deb2d675fe843ea634dcf9677c1d3ee92513971b724c788e4> <2d98dbcc0caa2055146e13f50ecf75491dadd36ad2baac56bc08562ec66ce110>
	   16 <107a4240fe26e5fb36cc007e7658964882f769f18c786ab152f25c5d2ae472f7> <b544831dbd34c6c252958151c49a734c6e625e42608c005e797edb6d0a8934b3>
	   17 <1e4013c4b0c5787dc1d78bdc8d52331039af4124112ee9346f110a4e8118e864> <24a0e4d31cba015783501ecdfa7a8ebae3a6bfd32e6d1a3614b11183c80980d4>
	   18 <115d49b082c83851d4d5e110a4abdaddbda9b0227f5b26bf52d5a22525235972> <546cc3ee5db47bfe9705aa95e2da29f228230353917e5d2b1932fe482fbcfed7>
	   19 <843de91d99d0091f17f4782d4feb2b760cd58b6f2476e8b02d908a1515078aa8> <134d556d0c27f6cc6bf3015c06611625739d889c5789fa75b3c83969cb88b1df>
	   20 <08aa3a565efcb7169fe0cbf72c12ce1750f2861fb6c6851613cbe974efc1684a> <01c0aca470f665eb7182e072bca89bc669ffe5b0296fe21343a8c327c8a84175>
	   21 <ebbe8b8a522abbe78277d0daa7892d9da87c27becd3ec03895233ad466318c44> <02855a25ccb75b2f8eeac5d1db25044b0aead2cf77021ed94f79f3001e7b8e9d>
	   22 <3c4d6d5cf12eba7dbd3e84329df61afc9b7e08fc1332a682344273396ec7dcdc> <b7311db28c45c90d80a1e3d5b27b43f8e380214d6a2c4046c8d40f524d478353>
	   23 <beae48ff70a19a31d662443cce57f77afe050b81224860255bcbc8f480c43cfd> <204d01a17c4fb7b18c2f48270150db67d4b0b9ce8786e03c9550c547fb18029e>
	   24 <ebb1b2a68972b7d3323b0361f3a1142f8b452e9298773def5635c2e2efa3700e> <f16e5629e9a1c668e1aa79c7887355f5f51b0cbb1f0835e04e7acc53ac55a357>
	   25 <4cc9e5d8de78967e573582cf7c74977c30b5469b2c0bace8ec259f71ba25c8dd> <4197b54c5aaaad47be24dbbc11c1bd3eeb6246542d2f5ae5f4398dd4a7601703>
	   26 <1c51e5b0241cca7c86f718b7d2c3d457a6e5e0b39f1f39ebafbb0883d427d936> <cbbfd59baddd3a7ce6e375e7d90050e271b13f132df85e1c12be54fe66de81f6>
	   27 <476015ad88b792a031e4dd983757c99aea3912e8f8c2f659de4bc1a2204cea13> <8a1c8f696f3e773c7eef57ac1389bd0280d558ea7862f01b641ec6da0efefbee>
	   28 <2e4f9ef71777119153639a71ff2417f522fe41b87e9c1cb7669f40f9d685887d> <d0509c538a8c3616681d761ae5c6f9d2aaded71890da2496156043082182ec85>
	   29 <ff81927aa42eda7f2a696789091033cf5be2fc1f5f3a2de22715eb33d6282892> <9c3ae48693f91343d0a5f0ecbb7dec9b973bf213678a653b0d9df510652a23c0>
	   30 <2dac862efc7fc6d54c99e6ec6e58c0b64da957e736d30093c867a120d5dbfc55> <b8065367924a4cfc786036c066caa738349cf1cda70dbfa85cceb4a09f85039b>
	   31 <03ca276405df4b2dbe6cfe7c2c56bcd2669f1b7d82c9f92991bf4102af6110bf> <6f77274fa6e27935bf89ae373a3b5ada5824bd4b2aec222aebd7fee7a482e9c1>
	   32 <1bf5bdae897f9a064209cf31299653137e865f905c8929449139545ac8253c32> <3358eab25f942236f3f4b6ebafe1c3eeeef793836680667c669464c3d4a0847d>
	   33 <be19cc8bd854ca7cdb07c2aeba12a14ccfa3085f9ffd9f753980c9d45b7b4e0f> <f3024bd5df2aa4aa4d19e551ede93dd075f7953acae53f0f9e8a384e496c5250>
	   34 <5be46df3ae5c10c189f1dc9ed2592e246bd2449aa0daae458ae8bfbd52f983c3> <b07e7617e89e28f953d096ec2987ebd8f3e74d933963b82773d37ab1b7a3601d>
	   35 <de44123726719c08d4c37c8c9b0be17b6b49826136aa7b908531bc91732b087a> <c8971334825dd1d67e4c48297292a07a40629675b3e8788efc687385300481ae>
	   36 <4136030bad7b5b1cfa7d9c98a9dc347a92651f29c2e110aff8897f267c042210> <697406d24ef88ebf9ca1972c1d528478858ead85782ed410ebbc1f3da48ba807>
	   37 <a6b70a313cc06afa2bd9c2911537d609d68bec9432e84b9679527d6abb588ba7> <836236aac0a8f08a5029115d57e7ef18cb27cce8d2c157a9f4f5615dcc348aea>
	   38 <2bb214987069d80b0abc2bbd68eba0331e3ae5f4106f7fc1e2e7b8d6e5370e32> <c80d0f28df33babe39f6ecbd19a4a6afa853aa4da03b6bd7a806229ded76d2c5>
	   39 <01cce2a036b68ed354316339f092dec7662bcebdd2066111d16ce55a937e2c61> <b9de1176d519a793946792b5417eaf7d2d5126977c5704fc0fcd8e1b2f589b1d>
	   40 <907bc366c885daa37495be671ef6c2f2e554ede3b53ce280cbe88a48b9d9740e> <418d19dd28f7e94c51a1782d322e03cba478857424497b4a373fde0fbae4ccd9>
	   41 <980ceaf804edcd8c96858193e6d5178bf604cc73bd8faad50d531549993197cc> <38cbbfa0f4ad2397eed7f76dc3cdb6b06a36660c0775d391ca47213341f659e9>
	   42 <272827216d1af9dcc6e9862a6e53a0a2c73298e1fadc0f9148cbc85ec0567c38> <014f70284efaa5faaba4bb8379ce0204f5aedc28268d82438b5b881fdf2dee4a>
	   43 <769c2765d654c4269f6ef13947f13c239cbb08b7cf67a25bac030ad1b892c434> <d7d40ed13dad57ca929614a63a00fe3a78f33b30b6fd5f39e4437036dced8d87>
	   44 <79246449f5328dac3141d3d7c8a9a2540dcac2cbc98e27843143e7d4b96dde75> <af43282f43fa14abaf6c8415fc05ee1ad171d81faa467ddfe5e02eb6895e5688>
	   45 <21fc70b3280a2a4c5f39287f5d24d7a759ea037b11448739ee2a28fc4b160eac> <dec048f6660e3a2fd8bdec602af59590ec4c6eab834cc0dec8621eb510fba6f7>
	   46 <406108aee6b580621311fe030bf08b4f6eed3d7d3d8693d3ac524da2b4ebf19e> <adf47693c2fd574d8220a2e70e73ad68e4c332488eb8e731fe600d1e9f6b8f5c>
	   47 <2559dc50ff35e62da620dc0a02edcbe4f398b1bd86ea154b6a9400579e3f1cd5> <bf699c18d06bcd73b7cfcef42e68af7ae67fea46e946de6a61faa42c535cfcae>
	   48 <7fdc2f10bd8cdb167c0b283f9007e620d9ca28067fe2b015ed657c9153b8443d> <aad5334fc1a9bad4a53e57d11c6accfcefd2e8ab44cb12fb2e664fcbdf5c82b2>
	   49 <77e8e25ff348f4df78bbc1ce20a7babde40ed2bdbeaf2b5cd98e5202baf7e3dc> <1289626ac2a1402bde7a869eb9ed7807338dd3b2ba8237845db96771cc988008>
	   50 <f18ba115620c51ae8b58b4923b9a8694c93df64b178c4cd2f9f6efc51f458b0c> <1acf053d9bd51c0101941c4c26f66aa5dbad3f5354608577f9e51afe743add50>
	   51 <5ee860a40ac8cec3506ec85b99dc716b95cbb342db91ade4b61e177f60f9fabb> <f1b5901bea7beb5ae780b6ece977f65b9c628e1dce0ad1e078c746c2f38d0e7f>
	   52 <ff2c9badee04cfd741d66d2f26321e2cf50a3cd021f6288863de2dadf8d52d1f> <06b088708ae9ac1117e3a37999c1d75a62e9c9e017018e088aebfb378de29c78>
	   53 <8b9f51424305a3d407962963c1d0beeb8113f80307ecc21923947fe8cbaf5c2c> <93acf10942584bf558a2d02d751e34f3f484b001e31924cc21848bf0ddaf1f3d>
	   54 <05ae6369852199c52a1797b9aff2a9245d7a8b9172d572b4432f63441ff51c4a> <8a310049736ff7f049294d8a595f2ca7263a3613840c14b33ef483cdca5bbb8a>
	   55 <4e270e3b61eae6e13eefe35e85427bc758ef4af4c00f9c77521c0361d299431f> <4c7004ccb8f67156267ee35f280db12645de8e552a9312df5769a030a6b46d80>
	   56 <9d8e298c13414c46170a1d82a1380fbafe531ca70184ab8965c4c807060e8039> <db2e6c06b3c76c1ada42373b29a0591f39856749dfdfb26681166a286fb4f209>
	   57 <fec4615e5909d27ac5ca8041e3f95b27f1c3d4d406a2048b1e6ce1e637cb87c0> <7a3b6f8febdbe4413b67b558689c2e7c1d6d6408f46a6094c74b2281e796e1d9>
	   58 <f97d3617d46aeffdd1e813c255fb8b3ef939a2c5fad4d10973c08c055f7913c5> <00cc835337a31b5350caa9c444c670f78f866e03ef6ec2cbcbc179974145b239>
	   59 <1664589da5145a9c5972f4b212ebf51171d92343833a08953cd80cd0d908904c> <b90912bbeef8f576961b5efc69641f7a7151708775b67c9e65ed9bb9f5a87bb7>
	   60 <563edc34294221865633d8cf6ff50444b9d29beb05a47b8bb121cb118d6cb16b> <90da203557bed2674055e8a6ab3646c4e1a845ea53d8614ae490065def757615>
	   61 <24c445098aa90e6d5a10eae0a0f3977a2808f79cafe8f8705297bd91ebbf2792> <a265f2ab98388029aec3afb5cca3a666ab29b6d2c002979c636a3b41b8837a43>
	   62 <a1892cb009db0b7ac351d0353f43fe3aa97192e8b9d7fef5baec415b0ca48c92> <2a81d6db55cf406b1f5842b0a887fe6b2bd88e46298ed3ecc3874c9837734633>
	   63 <0e7cdd78f9246ad254e87ee1b06584b860b0b8800aaee17896f0290cb789b0d7> <1fde7a2ff7f104265bbd2d0274c033c7583851001dcdb3ded90a9c0977c1f86d>
	*/

	for i, _ := range tests {

		if Hi[i] != tests[i].tHi {
			t.Fatalf("TestBulletProofHiGi: want %d Hi %s, got %s", i, Hi[i], tests[i].tHi)
		}
		if Gi[i] != tests[i].tGi {
			t.Fatalf("TestBulletProofHiGi: want %d Gi %s, got %s", i, Gi[i], tests[i].tGi)
		}

		{ // test inverse in Hi
			inverse := invert_scalar(Hi[i])
			var result crypto.Key
			crypto.ScMul(&result, &inverse, &Hi[i])

			if result != crypto.Identity || inverse != tests[i].tHiInverse {
				t.Fatalf("Inverse failed on Ed25519  %d actual %s expected  %s", i, inverse, tests[i].tHiInverse)
			}
		}

		{ // test inverse in Gi
			inverse := invert_scalar(Gi[i])
			var result crypto.Key
			crypto.ScMul(&result, &inverse, &Gi[i])

			if result != crypto.Identity || inverse != tests[i].tGiInverse {
				t.Fatalf("Inverse failed on Ed25519  %d actual %s expected  %s", i, inverse, tests[i].tGiInverse)
			}
		}
	}
}

func TestCompleteBulletProof(t *testing.T) {
	for i := uint64(0); i < 1; i++ {
		b := BULLETPROOF_Prove_Amount(0, &crypto.Identity)

		if !b.BULLETPROOF_Verify() {
			t.Fatalf("BulletProof random test failed")
		}

		if !b.BULLETPROOF_Verify_fast() {
			t.Fatalf("BulletProof fast random test failed")
		}
		if !b.BULLETPROOF_Verify_ultrafast() {
			t.Fatalf("BulletProof fast random test failed")
		}
	}
}

// test few edge cases
func TestEdgeBulletProof(t *testing.T) {
	random_gamma := crypto.SkGen()
	b0 := BULLETPROOF_Prove_Amount(0, &random_gamma)

	if !b0.BULLETPROOF_Verify() {
		t.Fatalf("BulletProof 0 amount test failed")
	}
	if !b0.BULLETPROOF_Verify_fast() {
		t.Fatalf("BulletProof fast 0 amount test failed")
	}

	if !b0.BULLETPROOF_Verify_ultrafast() {
		t.Fatalf("BulletProof ultra fast 0 amount test failed")
	}

	bmax := BULLETPROOF_Prove_Amount(0xffffffffffffffff, &random_gamma)

	if !bmax.BULLETPROOF_Verify() {
		t.Fatalf("BulletProof 0xffffffffffffffff max amount test failed")
	}
	if !bmax.BULLETPROOF_Verify_fast() {
		t.Fatalf("BulletProof fast 0xffffffffffffffff max amount test failed")
	}
	if !bmax.BULLETPROOF_Verify_ultrafast() {
		t.Fatalf("BulletProof ultrafast 0xffffffffffffffff max amount test failed")
	}

	invalid_8 := crypto.Zero
	invalid_8[8] = 1

	binvalid8 := BULLETPROOF_Prove(&invalid_8, &random_gamma)

	if binvalid8.BULLETPROOF_Verify() {
		t.Fatalf("BulletProof invalid 8 test failed")
	}

	if binvalid8.BULLETPROOF_Verify_fast() {
		t.Fatalf("BulletProof invalid 8 test failed")
	}
	if binvalid8.BULLETPROOF_Verify_ultrafast() {
		t.Fatalf("BulletProof invalid 8 test failed")
	}

	invalid_31 := crypto.Zero
	invalid_31[31] = 1

	binvalid31 := BULLETPROOF_Prove(&invalid_31, &random_gamma)

	if binvalid31.BULLETPROOF_Verify() {
		t.Fatalf("BulletProof invalid 31 test failed")
	}
	if binvalid31.BULLETPROOF_Verify_fast() {
		t.Fatalf("BulletProof invalid 31 fast test failed")
	}
	if binvalid31.BULLETPROOF_Verify_ultrafast() {
		t.Fatalf("BulletProof invalid 31 fast test failed")
	}

}

// test few edge cases
func BenchmarkDoubleScalar(b *testing.B) {

	s1 := *(crypto.RandomScalar())
	s2 := *(crypto.RandomScalar())
	var output crypto.Key
	for n := 0; n < b.N; n++ {
		crypto.AddKeys3_3(&output, &s1, &Hi_Precomputed[0], &s2, &Gi_Precomputed[0])

	}

}

// test few edge cases
func BenchmarkBulletproofVerify(b *testing.B) {

	cpufile, err := os.Create("/tmp/bp_cpuprofile.prof")
	if err != nil {

	}
	if err := pprof.StartCPUProfile(cpufile); err != nil {
	}
	defer pprof.StopCPUProfile()

	s1 := *(crypto.RandomScalar())
	bp := BULLETPROOF_Prove_Amount(0, &s1)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if !bp.BULLETPROOF_Verify() {
			b.Fatalf("BulletProof verification failed")
		}

	}

}

// test few edge cases
func BenchmarkBulletproofVerifyfast(b *testing.B) {

	cpufile, err := os.Create("/tmp/bp_cpuprofile_fast.prof")
	if err != nil {

	}
	if err := pprof.StartCPUProfile(cpufile); err != nil {
	}
	defer pprof.StopCPUProfile()

	s1 := *(crypto.RandomScalar())
	bp := BULLETPROOF_Prove_Amount(0, &s1)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if !bp.BULLETPROOF_Verify_fast() {
			b.Fatalf("BulletProof verification failed")
		}
	}

}

// test few edge cases
func BenchmarkBulletproofVerifyultrafast(b *testing.B) {

	cpufile, err := os.Create("/tmp/bp_cpuprofile_fast.prof")
	if err != nil {

	}
	if err := pprof.StartCPUProfile(cpufile); err != nil {
	}
	defer pprof.StopCPUProfile()

	s1 := *(crypto.RandomScalar())
	bp := BULLETPROOF_Prove_Amount(0, &s1)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if !bp.BULLETPROOF_Verify_ultrafast() {
			b.Fatalf("BulletProof verification failed")
		}
	}

}

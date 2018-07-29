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

package crypto

import "os"
import "runtime/pprof"
import "crypto/rand"
import "testing"

func TestScMulSub(t *testing.T) {
	tests := []struct {
		name    string
		aHex    string
		bHex    string
		cHex    string
		wantHex string
	}{
		{
			name:    "simple",
			aHex:    "0100000000000000000000000000000000000000000000000000000000000000",
			bHex:    "0100000000000000000000000000000000000000000000000000000000000000",
			cHex:    "0200000000000000000000000000000000000000000000000000000000000000",
			wantHex: "0100000000000000000000000000000000000000000000000000000000000000",
		},
		{
			name:    "more complex",
			aHex:    "1000000000000000000000000000000000000000000000000000000000000000",
			bHex:    "1000000000000000000000000000000000000000000000000000000000000000",
			cHex:    "0002000000000000000000000000000000000000000000000000000000000000",
			wantHex: "0001000000000000000000000000000000000000000000000000000000000000",
		},
		{
			name:    "more complex",
			aHex:    "0000000000000000000000000000000000000000000000000000000000000010",
			bHex:    "0000000000000000000000000000000000000000000000000000000000000010",
			cHex:    "0000000000000000000000000000000000000000000000000000000000000000",
			wantHex: "844ae3b1946c2475b8f95e806867dbac410ae82d8c1331c265cf83e4be664c0e",
		},
	}
	for _, test := range tests {
		a := HexToKey(test.aHex)
		b := HexToKey(test.bHex)
		c := HexToKey(test.cHex)
		want := HexToKey(test.wantHex)
		var got Key
		ScMulSub(&got, &a, &b, &c)
		if want != got {
			t.Errorf("%s: want %x, got %x", test.name, want, got)
		}
	}
}

func TestScalarMult(t *testing.T) {
	tests := []struct {
		name      string
		scalarHex string
		pointHex  string
		wantHex   string
	}{
		{
			name:      "zero",
			scalarHex: "0000000000000000000000000000000000000000000000000000000000000000",
			pointHex:  "0100000000000000000000000000000000000000000000000000000000000000",
			wantHex:   "0100000000000000000000000000000000000000000000000000000000000000",
		},
		{
			name:      "basepoint * 1",
			scalarHex: "0100000000000000000000000000000000000000000000000000000000000000",
			pointHex:  "5866666666666666666666666666666666666666666666666666666666666666",
			wantHex:   "5866666666666666666666666666666666666666666666666666666666666666",
		},
		{
			name:      "basepoint * 8",
			scalarHex: "0800000000000000000000000000000000000000000000000000000000000000",
			pointHex:  "5866666666666666666666666666666666666666666666666666666666666666",
			wantHex:   "b4b937fca95b2f1e93e41e62fc3c78818ff38a66096fad6e7973e5c90006d321",
		},
		{
			name:      "basepoint * 2",
			scalarHex: "0200000000000000000000000000000000000000000000000000000000000000",
			pointHex:  "2f1132ca61ab38dff00f2fea3228f24c6c71d58085b80e47e19515cb27e8d047",
			wantHex:   "b4b937fca95b2f1e93e41e62fc3c78818ff38a66096fad6e7973e5c90006d321",
		},
	}
	for _, test := range tests {
		scalarBytes := HexToKey(test.scalarHex)
		pointBytes := HexToKey(test.pointHex)
		want := HexToKey(test.wantHex)
		point := new(ExtendedGroupElement)
		point.FromBytes(&pointBytes)
		result := new(ProjectiveGroupElement)
		GeScalarMult(result, &scalarBytes, point)
		var got Key
		point.ToBytes(&got)

		if got != pointBytes {
			t.Fatalf("%s: want %s, got %s point testing failed", test.name, pointBytes, got)
		}

		result.ToBytes(&got)

		if want != got {
			t.Fatalf("%s: want %s, got %s", test.name, want, got)
		}
	}
}

func TestGeMul8(t *testing.T) {
	tests := []struct {
		name     string
		pointHex string
		wantHex  string
	}{
		{
			name:     "zero",
			pointHex: "0100000000000000000000000000000000000000000000000000000000000000",
			wantHex:  "0100000000000000000000000000000000000000000000000000000000000000",
		},
		{
			name:     "basepoint",
			pointHex: "5866666666666666666666666666666666666666666666666666666666666666",
			wantHex:  "b4b937fca95b2f1e93e41e62fc3c78818ff38a66096fad6e7973e5c90006d321",
		},
	}
	for _, test := range tests {
		pointBytes := HexToKey(test.pointHex)
		want := HexToKey(test.wantHex)
		tmp := new(ExtendedGroupElement)
		tmp.FromBytes(&pointBytes)
		point := new(ProjectiveGroupElement)
		tmp.ToProjective(point)
		tmp2 := new(CompletedGroupElement)
		result := new(ExtendedGroupElement)
		var got Key
		GeMul8(tmp2, point)
		tmp2.ToExtended(result)
		result.ToBytes(&got)
		if want != got {
			t.Errorf("%s: want %x, got %x", test.name, want, got)
		}
	}
}

func TestGeDoubleScalarMultVartime(t *testing.T) {
	tests := []struct {
		name       string
		pointHex   string
		scalar1Hex string
		scalar2Hex string
		wantHex    string
	}{
		{
			name:       "zero",
			pointHex:   "0100000000000000000000000000000000000000000000000000000000000000",
			scalar1Hex: "0000000000000000000000000000000000000000000000000000000000000000",
			scalar2Hex: "0000000000000000000000000000000000000000000000000000000000000000",
			wantHex:    "0100000000000000000000000000000000000000000000000000000000000000",
		},
		{
			name:       "8 times base point only",
			pointHex:   "0100000000000000000000000000000000000000000000000000000000000000",
			scalar1Hex: "0000000000000000000000000000000000000000000000000000000000000000",
			scalar2Hex: "0800000000000000000000000000000000000000000000000000000000000000",
			wantHex:    "b4b937fca95b2f1e93e41e62fc3c78818ff38a66096fad6e7973e5c90006d321",
		},
		{
			name:       "2 times non-base-point",
			pointHex:   "2f1132ca61ab38dff00f2fea3228f24c6c71d58085b80e47e19515cb27e8d047",
			scalar1Hex: "0200000000000000000000000000000000000000000000000000000000000000",
			scalar2Hex: "0000000000000000000000000000000000000000000000000000000000000000",
			wantHex:    "b4b937fca95b2f1e93e41e62fc3c78818ff38a66096fad6e7973e5c90006d321",
		},
		{
			name:       "Combination",
			pointHex:   "2f1132ca61ab38dff00f2fea3228f24c6c71d58085b80e47e19515cb27e8d047",
			scalar1Hex: "0100000000000000000000000000000000000000000000000000000000000000",
			scalar2Hex: "0400000000000000000000000000000000000000000000000000000000000000",
			wantHex:    "b4b937fca95b2f1e93e41e62fc3c78818ff38a66096fad6e7973e5c90006d321",
		},
	}
	for _, test := range tests {
		pointBytes := HexToKey(test.pointHex)
		a := HexToKey(test.scalar1Hex)
		b := HexToKey(test.scalar2Hex)
		want := HexToKey(test.wantHex)
		point := new(ExtendedGroupElement)
		point.FromBytes(&pointBytes)
		result := new(ProjectiveGroupElement)
		GeDoubleScalarMultVartime(result, &a, point, &b)
		var got Key
		result.ToBytes(&got)
		if want != got {
			t.Errorf("%s: want %x, got %x", test.name, want, got)
		}
	}
}

func TestGeDoubleScalarMultPrecompVartime(t *testing.T) {
	tests := []struct {
		name       string
		point1Hex  string
		point2Hex  string
		scalar1Hex string
		scalar2Hex string
		wantHex    string
	}{
		{
			name:       "zero",
			point1Hex:  "0100000000000000000000000000000000000000000000000000000000000000",
			point2Hex:  "0100000000000000000000000000000000000000000000000000000000000000",
			scalar1Hex: "0000000000000000000000000000000000000000000000000000000000000000",
			scalar2Hex: "0000000000000000000000000000000000000000000000000000000000000000",
			wantHex:    "0100000000000000000000000000000000000000000000000000000000000000",
		},
		{
			name:       "scalar 1 only",
			point1Hex:  "5866666666666666666666666666666666666666666666666666666666666666",
			point2Hex:  "0100000000000000000000000000000000000000000000000000000000000000",
			scalar1Hex: "0800000000000000000000000000000000000000000000000000000000000000",
			scalar2Hex: "0000000000000000000000000000000000000000000000000000000000000000",
			wantHex:    "b4b937fca95b2f1e93e41e62fc3c78818ff38a66096fad6e7973e5c90006d321",
		},
		{
			name:       "scalar 2 only",
			point1Hex:  "0100000000000000000000000000000000000000000000000000000000000000",
			point2Hex:  "5866666666666666666666666666666666666666666666666666666666666666",
			scalar1Hex: "0000000000000000000000000000000000000000000000000000000000000000",
			scalar2Hex: "0800000000000000000000000000000000000000000000000000000000000000",
			wantHex:    "b4b937fca95b2f1e93e41e62fc3c78818ff38a66096fad6e7973e5c90006d321",
		},
		{
			name:       "Combination",
			point1Hex:  "2f1132ca61ab38dff00f2fea3228f24c6c71d58085b80e47e19515cb27e8d047",
			point2Hex:  "5866666666666666666666666666666666666666666666666666666666666666",
			scalar1Hex: "0100000000000000000000000000000000000000000000000000000000000000",
			scalar2Hex: "0400000000000000000000000000000000000000000000000000000000000000",
			wantHex:    "b4b937fca95b2f1e93e41e62fc3c78818ff38a66096fad6e7973e5c90006d321",
		},
	}
	for _, test := range tests {
		point1Bytes := HexToKey(test.point1Hex)
		point2Bytes := HexToKey(test.point2Hex)
		a := HexToKey(test.scalar1Hex)
		b := HexToKey(test.scalar2Hex)
		want := HexToKey(test.wantHex)
		point1 := new(ExtendedGroupElement)
		point1.FromBytes(&point1Bytes)
		point2 := new(ExtendedGroupElement)
		point2.FromBytes(&point2Bytes)
		var point2Precomp [8]CachedGroupElement
		GePrecompute(&point2Precomp, point2)
		result := new(ProjectiveGroupElement)
		GeDoubleScalarMultPrecompVartime(result, &a, point1, &b, &point2Precomp)
		var got Key
		result.ToBytes(&got)
		if want != got {
			t.Errorf("%s: want %x, got %x", test.name, want, got)
		}
	}
}

func TestScValid(t *testing.T) {
	// All tests from github.com/monero-project/monero/tests/crypto/tests.txt
	tests := []struct {
		scalarHex string
		valid     bool
	}{
		{
			scalarHex: "ac10e070c8574ef374bdd1c5dbe9bacfd927f9ae0705cf08018ff865f6092d0f",
			valid:     true,
		},
		{
			scalarHex: "fa939388e8cb0ffc5c776cc517edc2a9457c11a89820a7bee91654ce2e2fb300",
			valid:     true,
		},
		{
			scalarHex: "6d728f89e567522dffe81cf641c2e483e2cee7051c1b3cd9cd38cba09d2dc900",
			valid:     true,
		},
		{
			scalarHex: "18fd66f7a0874de792f12a1b2add7d294100ea454537ae5794d0abc91dbf098a",
			valid:     false,
		},
		{
			scalarHex: "fdb4cee00230c448aeaa4790dd12e855eca6223d81dab6bfe3fe99ce5e702325",
			valid:     false,
		},
		{
			scalarHex: "f3ab5c4a64f9afbd48c75ad25700996a69504d31c75a6aa8beb24d178cfc32ba",
			valid:     false,
		},
		{
			scalarHex: "167e88e9298926ecd66d32a890148a9adcfb80a7ecc69396edc5f5ee8848ce91",
			valid:     false,
		},
		{
			scalarHex: "a9172ced17605ab2ba82caec382f1176fe6a1e0b92a9c95679978b3999d22605",
			valid:     true,
		},
		{
			scalarHex: "44f65ce00e64cf29e07bf3843e8e43e83b8b26d3dfcd83160d9a3fedda3fc908",
			valid:     true,
		},
		{
			scalarHex: "69951c807647f4dea33f8a40b0ddef999a9e8efd1b1b176f45fb92aae6287000",
			valid:     true,
		},
		{
			scalarHex: "dcb6f94312b4b815101af10775d189301d063816b0aa9b9ac028c0eb9b17dd0e",
			valid:     true,
		},
		{
			scalarHex: "a1a9574a90bed7f58bfd4c3295e181e152a4a39f9ad56520947e94fea40dfe2e",
			valid:     false,
		},
		{
			scalarHex: "757ebb27b14433f3ed696cd291825eba555af0863862216a98cf3080432bda01",
			valid:     true,
		},
		{
			scalarHex: "a7401d68c7956d41a05d8c6c1a3cc692cc269ad8a5837f93ac2c7654abaebf0e",
			valid:     true,
		},
		{
			scalarHex: "7c84ef4d9627529723f5e171a88bba91118192e6e19929204a6d23f5a541850b",
			valid:     true,
		},
		{
			scalarHex: "ebe4b2f9eadd52271e1115aa3b89b7bb3f68d58bf74d75486b77e6ae8216f609",
			valid:     true,
		},
		{
			scalarHex: "6221caf5dfd623587ca7aac45f5312faf1c2847d7bfbc5c46f88fd933970c866",
			valid:     false,
		},
		{
			scalarHex: "b38334b41f0b829bf3a2ea73b58c16afb56ccba5144668e78257ed4aa9a33a5e",
			valid:     false,
		},
		{
			scalarHex: "92e5df37f5412d9e9910cd28d27138ba46bdd4dee5dab1125f200f0a4e66f01c",
			valid:     false,
		},
		{
			scalarHex: "2d1717fe2965133009bdcf0c13b02ca85df4dc23d969d7b16fdab94341c2b90e",
			valid:     true,
		},
		{
			scalarHex: "82d771565024cf4b5c9add6d10a4114a2fbdbf7626f03749c8bddcf6c6feee03",
			valid:     true,
		},
		{
			scalarHex: "2db0f3d9bf47666bb40ae25ef7508296af2ef665a5333efff92c9daee5ecd986",
			valid:     false,
		},
		{
			scalarHex: "47883a3f0b0e624512887f3dae6f401b2f02eb8950eb8e741f07347108fbea14",
			valid:     false,
		},
		{
			scalarHex: "dae25f27f2756b8270f78015b1beafbd03a8d67a4dd940ee9550cecae9f3560f",
			valid:     true,
		},
		{
			scalarHex: "bab79d355572361259484ca77308943b7c71f5831be5dde11b3c7a26ac6e9a01",
			valid:     true,
		},
		{
			scalarHex: "f0bf30a9862d51fd8d1ca3a4e5a0e4330a5beab751af075b2e46f3e2d0297101",
			valid:     true,
		},
		{
			scalarHex: "a94270659333c985ab1461e9820f219416ad76f551a34f4c983ca967aa6b618c",
			valid:     false,
		},
		{
			scalarHex: "98e582e9d0e485e6b756e44b1cb187e9f2d4e16c02d2713a41dcee6d7fac0205",
			valid:     true,
		},
		{
			scalarHex: "62b57869238e72f05992e01f07c7a26c71ef9e3078ee7390cd4fc6406d3b37e3",
			valid:     false,
		},
		{
			scalarHex: "c751566f5420232cd626d2bd8073debedf219a0a28a7390200a2df9df8981647",
			valid:     false,
		},
		{
			scalarHex: "cfafede7d76def164bf15047a388365d712a600f1369bf692cb2dcb249f66508",
			valid:     true,
		},
		{
			scalarHex: "c0eb03c2c46a77b0cf67bea7d547b9b4d02116f70702790398bdabbfa775100a",
			valid:     true,
		},
		{
			scalarHex: "58c5afbb7b334e921bb46eeee1dfebadfbbc3a77c420b0b325b9470eff94cc0e",
			valid:     true,
		},
		{
			scalarHex: "b31b178b4dc8f7a7c24f6eb0987e32b8a9446ce8207186259074cc3d925c530d",
			valid:     true,
		},
		{
			scalarHex: "467b2c5f63120744b5aee7ec255d170aba83babdec9e76c2690300540eb17483",
			valid:     false,
		},
		{
			scalarHex: "9bc3be487bea81da020af0353660878e7deca693ac81278fac343b8b1809bd9c",
			valid:     false,
		},
		{
			scalarHex: "162cdb42f9d9147118537cc4d5eadb8f589f6de3929188f6c41428bdab2c650e",
			valid:     true,
		},
		{
			scalarHex: "6eeb88f685d1b5c99f8c27dcadf95f5bea9082e2735a4f8aafc3c3a35f749002",
			valid:     true,
		},
		{
			scalarHex: "e03bb5c9bedc41d324ca6d988e50b7e7c6cd6d2b1e471b752a00a10a19c9a604",
			valid:     true,
		},
		{
			scalarHex: "4fa064dcef2f14fc453c30865583468ad35af81c460a643b11285c15f2babb04",
			valid:     true,
		},
		{
			scalarHex: "bf067a7b53dab7fcd6a1f5a2c8b9b54c262e9be5ea5892dba5e4ad25585be702",
			valid:     true,
		},
		{
			scalarHex: "14d8d255c2156bb12133d5ee792061a21d3fbda2f14310753f3a90c2ad4ccbc8",
			valid:     false,
		},
		{
			scalarHex: "81e78a37b7e2dfed35ad8228a6c4a6f0aa34cc1de43da23895c5c86790804e01",
			valid:     true,
		},
		{
			scalarHex: "62366805970fb07b877bf1705d9c10517c6667ab7586d9afa82c30236a72f50f",
			valid:     true,
		},
		{
			scalarHex: "bb3b0e27b7c57052e5ac20bc6390229f57c6dbc8213dbc4ee1cbb9e174e389f9",
			valid:     false,
		},
		{
			scalarHex: "5c4039f162531ce9129667ed424446ef375a18da246a72c0dab00321fa04e929",
			valid:     false,
		},
		{
			scalarHex: "8d41a450d98a6f354266e36510921649e5229412976f8f318a3090d390a46807",
			valid:     true,
		},
		{
			scalarHex: "2fbce4441723379b8b6ef417803eb9220c8fee42f72da7e36e2c64b51628a10c",
			valid:     true,
		},
		{
			scalarHex: "5bbf0e5b58e3df89a5f4c607e3033fff8b9643fc900813f4ea50b889260bb004",
			valid:     true,
		},
		{
			scalarHex: "e5acfcb2ecb7175f90aa24db107cea1be7a45bc471c4c65b9f46535768545de5",
			valid:     false,
		},
		{
			scalarHex: "fae6e8d7226e6dc25faafdd8b5c63b68587980f6e76ea59b8595f18f64a96bd5",
			valid:     false,
		},
		{
			scalarHex: "6736f7939157ccee066a40056c13d96230ca5ffd014b0ebb0a454e3d8ea6830f",
			valid:     true,
		},
		{
			scalarHex: "b2b490d52f6e4d9ef14fc1ab3e1e97160693ba5d1220e4d63956eca316693504",
			valid:     true,
		},
		{
			scalarHex: "dca490cb4e21589e045150e8c0876871f08e6b3886c6393c6348bfb432799079",
			valid:     false,
		},
		{
			scalarHex: "ca2663d84aa7d7581ac4f0a270a685019178f5288013d163028d321e92b04901",
			valid:     true,
		},
		{
			scalarHex: "7dd277472249a53620d6b21c1ab0424024581b5ae3e6ecb9a46af8511b55a408",
			valid:     true,
		},
		{
			scalarHex: "30f07adb405de73987d81fcba3d523d420d130fe753afbf3dd19d5f08e8a1356",
			valid:     false,
		},
		{
			scalarHex: "cc31de9889440c695df43a9c621a7156bc064b161be2fdef35dd8ae0e1017fe2",
			valid:     false,
		},
		{
			scalarHex: "962ac1be6677cf0d298dd1244501e6a164a915df61b83c2837f80a84541e2e03",
			valid:     true,
		},
		{
			scalarHex: "e0a1e1c48f5dfb7c0faddd9186214ec2f81fb0bc3aca423ccf0e538cf43a2bd5",
			valid:     false,
		},
		{
			scalarHex: "8ed93780cbd46092acce856a8c691edbb8419deec5acfb26fcf82419b4a964b9",
			valid:     false,
		},
		{
			scalarHex: "f62cf00cac227d2b96e7f96268f0d606ad5924b0efb64e623e25584f993c93a8",
			valid:     false,
		},
		{
			scalarHex: "e49eacd6858efad9f3bedba8b9edff5ff3d5bd622926d5e96e6e8c8752a43d25",
			valid:     false,
		},
		{
			scalarHex: "ed424965b9a79c2f6611d7d4bcbe5bae4dbf2c8130141c714acc95c492061411",
			valid:     false,
		},
		{
			scalarHex: "7ea2bb2a4dab662c3adf71eb959314208ff9dbf50b3cb1a7940e4a4d6d797e0e",
			valid:     true,
		},
		{
			scalarHex: "96fc939742b8ed518bb7aa4bb6c9bbe3fddded6b9441d9ea27712c377e0bdff1",
			valid:     false,
		},
		{
			scalarHex: "c5122ad5575b54ebce3f1d1d97c3bb0cfeb9e5ea3bb223ab63189ced40579cf4",
			valid:     false,
		},
		{
			scalarHex: "c9147f1b42890b089ae2dfdc27f69b4a29c67b8bc0264259fcf71c39e31b893c",
			valid:     false,
		},
		{
			scalarHex: "9292516356f70e84852c3a563970bff369dfc39325c64432610baa0f31f2320e",
			valid:     true,
		},
		{
			scalarHex: "85f8b6251f177701b319c5db4bc372a5cf05cb07d86b2b2d5c938b0faf6d670f",
			valid:     true,
		},
		{
			scalarHex: "de88df59b4e8e3fe2beab575b4c6ffccbd24cf294f890fa61cbc231b0c976f0d",
			valid:     true,
		},
		{
			scalarHex: "2922bd9b50e09fe75a3745c62afd5bd1813dcffcf4503668e71d8f5d37c0090d",
			valid:     true,
		},
		{
			scalarHex: "f090b3ed9355221e4fb40397c367fb31a2c7ce308ba96b9c4fc9610c14cbe305",
			valid:     true,
		},
		{
			scalarHex: "1c545854bd134a995f2138d6d088032c33a9a23a30748615652e6675d56e670f",
			valid:     true,
		},
		{
			scalarHex: "588050d78a3079a12bc88d6bcaf34a95bc6928c9233f14056277bf36bb355d26",
			valid:     false,
		},
		{
			scalarHex: "3956481800462e5c34aa2178dde33d5096c23185d61a5839c212d11e61adc901",
			valid:     true,
		},
		{
			scalarHex: "d0d61e471d5d503a6efcd549694b61851a58b0c8f8fb400227e2a48c1cbe4a0d",
			valid:     true,
		},
		{
			scalarHex: "1eed64f825d3f1bba4a226425dbef91f102ec01bcf4a0ac0a00baef8de20c405",
			valid:     true,
		},
		{
			scalarHex: "86764f74c855d580581ff17416a2c394de55bb3cb6e676756e86f3893ee61a09",
			valid:     true,
		},
		{
			scalarHex: "a1cc5ecb2a528ab8fe3efac17a2859c2ef1b81b51d18643bb3ee3fcc8794eb80",
			valid:     false,
		},
		{
			scalarHex: "740464b60e54da6a3541057742d327330ae5ac190e2b86708f5af7a1ff89fd0a",
			valid:     true,
		},
		{
			scalarHex: "ada63ca80e15912177b738c7c7618917827893b57ad20688a0a9c93c6c7c4e0a",
			valid:     true,
		},
		{
			scalarHex: "a55a6f1f2cab41704b251d3c64f3123f1635e3ee24fcce7e88cdaaf148636500",
			valid:     true,
		},
		{
			scalarHex: "515bfea540021ffb6d8f2324952a1ffa02f40adb860cff143ff0d797bd63fb09",
			valid:     true,
		},
		{
			scalarHex: "2784fbff1cad8d63b47359691caf8c7972219d7f9a64ee25fdbfe970981b2e87",
			valid:     false,
		},
		{
			scalarHex: "596071a52ac2e1b55001170c7ae5ee8a66a52e4b14db270a66c4828eccfca8a5",
			valid:     false,
		},
		{
			scalarHex: "afee887ccf82dd3406988f4afafd0d779e43d1f0f1b5107b1c3aed15d4ab57ed",
			valid:     false,
		},
		{
			scalarHex: "fb38ec6e693ee2aed1364c29dc7a60a2a7cb10601582a0ca4f4a0ca2c0499e08",
			valid:     true,
		},
		{
			scalarHex: "7f890e7539b8b3e6510a3f523a764f582f9e2c791400678eee85bd5842978c06",
			valid:     true,
		},
		{
			scalarHex: "26efbf6e436e974860ee3b6c197afa55c29b115e68b488ba46c1e865fe37f894",
			valid:     false,
		},
		{
			scalarHex: "6f6d9f15f262dde42c60af6880e8c6b6f07cab7cad785c19866f9a5c37419240",
			valid:     false,
		},
		{
			scalarHex: "f90673b823b480de9e560948307e1ada8ff1f4380e74e89ee45aa7fecae1d308",
			valid:     true,
		},
		{
			scalarHex: "269abc577fd32b1625895a1117dbc367822f9849441d4dbf7f56ca633fbf50b8",
			valid:     false,
		},
		{
			scalarHex: "26fcbe1690ff0bc306a9ec3cef75bee1a5e39710d5c02d46334fb7d89e75dc0c",
			valid:     true,
		},
		{
			scalarHex: "48c827ac48cda0f2ccd29a37884a14fa9cebe2ddf76bcc4fdf827245835d1daa",
			valid:     false,
		},
		{
			scalarHex: "9c82c6403ec29e60ad812c1762610e00c05891c6eb0fbf2393b886f298397e1a",
			valid:     false,
		},
		{
			scalarHex: "260fb700bf36504bfc5bea76c115929746f4aa709ba831fddb64b59af9bfd704",
			valid:     true,
		},
		{
			scalarHex: "aaff292106a5dd831c507603f72fb7e0b9aa5ad990575a7a3dbe83744668c552",
			valid:     false,
		},
		{
			scalarHex: "aca3f81877ac78519ce2944063fd7c50530d0ba7de2fa20ebdb1107872b6f302",
			valid:     true,
		},
		{
			scalarHex: "1086efacf20320ba25eedd1cf8d3b28348b0bf3e2191b7585e16dc9e4f7f5803",
			valid:     true,
		},
		{
			scalarHex: "66905a80d3d06c38faa5212c7455e5f48cdd5f15d816025b6466c085ef6f2d51",
			valid:     false,
		},
		{
			scalarHex: "a9fb08088371cc78c2b514608cee6b43c534e657d4e1f7d0cfd9fd4173aca769",
			valid:     false,
		},
		{
			scalarHex: "590e0a35ac44640ecf0f1f89eba200c6d98245b87360496ef408297af0011673",
			valid:     false,
		},
		{
			scalarHex: "9a83a1ab418d8d9e92e0ecde0cb1246e4b25f407c5421a3997cda43ede83930b",
			valid:     true,
		},
		{
			scalarHex: "bd3b91b68594291e859ace170fedcc64adfff013ed9f066bc625326e452ebc54",
			valid:     false,
		},
		{
			scalarHex: "ef68e423644b94c8549a9a8ca49e396d642ef0fc28448b77b2bdee1487787a0f",
			valid:     true,
		},
		{
			scalarHex: "6a6fe13345ffd78567560c8dfb511f20edfc15067212aaee69bddb77d5b4860e",
			valid:     true,
		},
		{
			scalarHex: "a1fb1f4e3e88997e4e4017a8346d37671cad86cd5d795d2550b1b9c805dfdb0f",
			valid:     true,
		},
		{
			scalarHex: "99e61e19a4c6c94354a5b7a6e206663f677223a6d3fc78659d9eaec21ef38402",
			valid:     true,
		},
		{
			scalarHex: "d4bf37268f05041950d82420a78753d168b8b808628b48d6ce7b0ed330701d02",
			valid:     true,
		},
		{
			scalarHex: "fee112a3e2de6313bc8809aba4466720c552d3615538b0523163fb01cbd1258a",
			valid:     false,
		},
		{
			scalarHex: "b9692112ef467088d7f32a873719c7f2179be1445aef71eab73f218e6d6336b9",
			valid:     false,
		},
		{
			scalarHex: "9aa00cad08702566bc88d11a2e8a99b188efd0d5fb164a1b4f3fc7378b962800",
			valid:     true,
		},
		{
			scalarHex: "b4301f1630b288b9d6a8afed5c72c10c3c3ce8c935a7aed218e8ff3f70b4a539",
			valid:     false,
		},
		{
			scalarHex: "f7f4d3bd2a84cf95d8793bd5e792510d5e95b8ce833775572e901795b5e1bd01",
			valid:     true,
		},
		{
			scalarHex: "3b7d10a8b1a2695dc00c7f12fc12c0876b27509b1ffa452c8b80643f3cd26315",
			valid:     false,
		},
		{
			scalarHex: "be2fa0a79bfa9fad3d7e9be1e5961c4849a38a1f4243a0c55d65d019a4fc5c0b",
			valid:     true,
		},
		{
			scalarHex: "a03687639bb3d2e8506fcf8d062fbc342b499dfc26d36baea80edc87674d6b0f",
			valid:     true,
		},
		{
			scalarHex: "c995ca23c04d581d52f7e550a4531b837f185ef58362b13f237964c0663f0139",
			valid:     false,
		},
		{
			scalarHex: "32865bf3762338c5e13fba1ffcc08629498a31af695813f221ebfba72ff7714b",
			valid:     false,
		},
		{
			scalarHex: "a8f37b565fe66aa2e851a74904e98c3059f92a037979da3bd4a52a10f719bf09",
			valid:     true,
		},
		{
			scalarHex: "f7474c236309d9888f0d8a53ab5a22bb4d600f3687a4e1d2815b89f866523202",
			valid:     true,
		},
		{
			scalarHex: "090d4f638d250e4588d1a25888c0b9ac6c207e4a98a9ef5a69ad0b6b1ed7ce1f",
			valid:     false,
		},
		{
			scalarHex: "6b6daf404e2b6f4e2e0398baf8627fed108d0a0bc38a364e1b5d55bfd63f6d21",
			valid:     false,
		},
		{
			scalarHex: "7e83a7bbf587be0694738f36bf52b24577a3489d72f7426ec0886a37ece4f10d",
			valid:     true,
		},
		{
			scalarHex: "5f7abc23dcf726378db619bc2ce1e86ae209daa6d1909117ce7270829c4d3da3",
			valid:     false,
		},
		{
			scalarHex: "f9ffc537ad782ce9f00f80be30e8b822c901bf62b2433c365a2f856e65f88706",
			valid:     true,
		},
		{
			scalarHex: "9d3374f71e9ad3dbad360d8a80bc8b5c0d814eb09cade0c62bc094b858849004",
			valid:     true,
		},
		{
			scalarHex: "f97097fe5d46b78e030a5207f1db92df9c8efcb0051b6e3b65cb1a58dfab660d",
			valid:     true,
		},
		{
			scalarHex: "ead8314b9663a10344eaad550137ffb0906500057efd4a28b69def67bed8a10f",
			valid:     true,
		},
		{
			scalarHex: "38ca4c5eefa2a159c05eac5e05466db588c8fbbac7df6717e1c2fc1cd15e0956",
			valid:     false,
		},
		{
			scalarHex: "9b43159b9077fb2a049aa067e1d5c1a0d16da8515934b7588f4670483050ad0e",
			valid:     true,
		},
		{
			scalarHex: "a43c8bb23b53fcd288d71d301e9cdd225bdf88e5d58ef3b485d8318c19fb9b9d",
			valid:     false,
		},
		{
			scalarHex: "2278fffc9472031130714cd01599f4a476fe5b94cd092fdc5568917fb7a479f9",
			valid:     false,
		},
		{
			scalarHex: "b78718c39d1c90448ef8787350c8baaec08de36ecde9ac234eb570397761a703",
			valid:     true,
		},
		{
			scalarHex: "48a8506dd388eb99eab0d587bf4a7bd34e68f001eb88999a76f1878c32b8e478",
			valid:     false,
		},
		{
			scalarHex: "659eb800501cccb504ae9bdcf5779ba8b2d24e4f63824c84124c453a34dd3540",
			valid:     false,
		},
		{
			scalarHex: "1ce40f32aee47060a0dadf34ce04f81c48af3218051770fb86976785f981b05e",
			valid:     false,
		},
		{
			scalarHex: "e7360d88d6dbbf3c352b80811110235a213aa5a3ea182d36c8c670ba9610a00f",
			valid:     true,
		},
		{
			scalarHex: "82934fcaed4804020e1ce4ff798c30312aff862b1371f7f74d71b4316b413806",
			valid:     true,
		},
		{
			scalarHex: "5bb1383ae3547b8caeba767eaa73a8fc08ac5972f982c736568424f8665770ac",
			valid:     false,
		},
		{
			scalarHex: "533b7c9bde705efe0c08a8a0605ea2d4c3ec21919aeca99b40e10e29ca7dc255",
			valid:     false,
		},
		{
			scalarHex: "93c1411f47251750b1b95c481d0ce55b401722273932914e2d1e78924e08fb06",
			valid:     true,
		},
		{
			scalarHex: "5b90a4fd7d50881168802f7a8da19437305e38203b6bb6ac68b114b59dfbfa0e",
			valid:     true,
		},
		{
			scalarHex: "e98125b5ad8262054c811f92d0596863ef876cb86c7f6b5783aee2e8cd6a5300",
			valid:     true,
		},
		{
			scalarHex: "3b47eac89bc19bc46dde8fb091bb1128ddfb070cbac09b65d37416d62b1d9d3a",
			valid:     false,
		},
		{
			scalarHex: "4710f6af3f144eaa2a8f1f2af329bebe752d3594c73008bad237e5e0be0cf07f",
			valid:     false,
		},
		{
			scalarHex: "a6a654a6e2d2ecd7b1daea341897658b01ab34f7e84d9fc1b3ebf8214b39a840",
			valid:     false,
		},
		{
			scalarHex: "006ba25fa6dc45938b6ab7bf6fe7096bb0eee2fc17688995a3a01567739503d2",
			valid:     false,
		},
		{
			scalarHex: "ec2a2a700045b57c85c73770fbca5896aba8941ae39035a29641891ddb309a08",
			valid:     true,
		},
		{
			scalarHex: "a6ea18e936f99d7f284553e83545cf324784e5a263f47cdcd78cfc4da10cf80b",
			valid:     true,
		},
		{
			scalarHex: "86853a4044faaa87e2a949d9623d523eae1022467816a713488ef18de5db590a",
			valid:     true,
		},
		{
			scalarHex: "0f79400a90af899e64dc6c718cb2df4f8426cc77e1745e64af2dfc4743f4ef08",
			valid:     true,
		},
		{
			scalarHex: "4b9727146720b7b9fc2bb00b7b65b18909b53558fc5c4777e2efcb336ce9f103",
			valid:     true,
		},
		{
			scalarHex: "4a71e9840fb19d79255f06456bceda85029950a9d7aebd80f19da16672ea2e4f",
			valid:     false,
		},
		{
			scalarHex: "a76d81141e740c48642aeecb93f87edbe74c94d7795c35938979913f62355350",
			valid:     false,
		},
		{
			scalarHex: "fe82ea864b773ba785ae30742142fc95a6311f1caeacba19ed019a11b3ddee06",
			valid:     true,
		},
		{
			scalarHex: "c43a927dd10a6fc47bd3888aa9d9613ba258ad838832b923251a2863691c0600",
			valid:     true,
		},
		{
			scalarHex: "55d66194e087988015066ddc0237d061163797305034f8b54524e935f832ed08",
			valid:     true,
		},
		{
			scalarHex: "650c717392bdf31373dbeda48d99fd10b76080f0319b6039f48eedc8de8daa01",
			valid:     true,
		},
		{
			scalarHex: "ef8a542e46802078598d8f41489c0d618cc19ac897719f3b298ea22cca35030c",
			valid:     true,
		},
		{
			scalarHex: "6f8be2599f25b3645571a0d3edda8d354218f1bac4e22fed861fb472ff615a04",
			valid:     true,
		},
		{
			scalarHex: "78b99806f01c239ea3e717e15a509ffc2f3c7481b5eceb365f7eb8e2ab13d907",
			valid:     true,
		},
		{
			scalarHex: "2bae1ebc0fb83cade154f1daaed5d1f03581f350e733f2f0d56a73fa3c97cf2c",
			valid:     false,
		},
		{
			scalarHex: "4c9f5a53e544ae46926b1504af11a7500f57487e62836c3f37b445c2b701dc0e",
			valid:     true,
		},
		{
			scalarHex: "162003195972b318297a38632d9859bc86f1677154f94996970a320e0d23aa05",
			valid:     true,
		},
		{
			scalarHex: "9563b33d7a9a751d3c8ae2b4536bc2ed7cf5231051191653d7c08d502357ed0b",
			valid:     true,
		},
		{
			scalarHex: "2ffc2e41884b11129e1cc83aa95cd0b44dacb040a9ff1809a353597af8af0736",
			valid:     false,
		},
		{
			scalarHex: "ebfb376fb3a20871f35a54ecc2bf0003a2f2d00838196f0860e229c1db21e90c",
			valid:     true,
		},
		{
			scalarHex: "3da00eff81fd6771a23481fa8719fbe674c0dffa8db583e5fc0da7c9a0713006",
			valid:     true,
		},
		{
			scalarHex: "2d492ad9915f6274a6b62cdc245204527a1216b4ef0e7e1e1d3aa6d81f0c3905",
			valid:     true,
		},
		{
			scalarHex: "1a6d227bc48af4cff57ab11dc7ba3b7da21a9998b05dbeea4c05ca8bcffe2277",
			valid:     false,
		},
		{
			scalarHex: "8db67435b9e243dfbb58f252061ca3a005d197b941fa1f9cbbd29fbeafd507d4",
			valid:     false,
		},
		{
			scalarHex: "b2f744640c89049ebbbf94bd3013d3103ccf0e750a810543b39601116dc3300d",
			valid:     true,
		},
		{
			scalarHex: "8177e87433900f41dbdc22db054c79995f6f610647bcee7527f51091f5198274",
			valid:     false,
		},
		{
			scalarHex: "2c7cbaa6142473cecdb1c10bf9fc2430984a4b6cd34afc607668c6d88bf24f06",
			valid:     true,
		},
		{
			scalarHex: "a0a2ccb60bcbd0eda5dd11540e3517462bd841d58d4aed185beb5cfbda9e5f34",
			valid:     false,
		},
		{
			scalarHex: "767219f2581c07b1bbccd62fe4af10b4e202538fe918cedefbba3297bc89b604",
			valid:     true,
		},
		{
			scalarHex: "ed8828a708c013d80fe558be718b1e738063c4eca7528c06067b73d9dd2adccc",
			valid:     false,
		},
		{
			scalarHex: "4e81f8c4a0166a40db4063b31a5a4e511648b0f094d8930d956f74a7bec44f03",
			valid:     true,
		},
		{
			scalarHex: "09d5a71ca33bbc186f06b5ece548f25b0d008ad71c04e55d3d4c5fe0ef960b19",
			valid:     false,
		},
		{
			scalarHex: "51aa0d12359bc5819055618be5955824d81bd6e5ef52d51ee7e75aa9ca09edc4",
			valid:     false,
		},
		{
			scalarHex: "94927b83b0f167631a5c8eae13a9290c027f795dbd0c3d0f2d1a270b65db05e4",
			valid:     false,
		},
		{
			scalarHex: "e30c00666b11f332690267f3a2b8caa93a9a6e3961522a4ef886f3768f4c4166",
			valid:     false,
		},
		{
			scalarHex: "6ec7e22c9921501879a85c89c2833e04526538c6cbdecd8a040fe6c57c881c06",
			valid:     true,
		},
		{
			scalarHex: "54a4299c9d8572857247843d38842926fc88064905d8041974a1d9d407313a0f",
			valid:     true,
		},
		{
			scalarHex: "d4604b16400e77f37e8f6f5cc4ced1ef0da8f2a832bf6a0f51321d221111ce83",
			valid:     false,
		},
		{
			scalarHex: "a545d4cb99de9fc8f976356996351009628be30201dd42a11f7b379a5099c40b",
			valid:     true,
		},
		{
			scalarHex: "7db6f2c324cda1f3c9b39d659e5b5e39ad8e4b65e31e3afc10e91f890d6c0b9f",
			valid:     false,
		},
		{
			scalarHex: "afee4fe9a50c0d0aba669ca7f866ac4eb16e83afd843f142caf1132b9c5bcd11",
			valid:     false,
		},
		{
			scalarHex: "03cf1b6414acc2990bf4df95a9afa40aefcbc55946f8f40298cec7b45e942b09",
			valid:     true,
		},
		{
			scalarHex: "82c67f0cc06d1f5d878e710dc3734dc9b698fbd43b9cfaa50bc0734240631eee",
			valid:     false,
		},
		{
			scalarHex: "ef10fc4f79bdfade5af636fd41a7e5c48e966b5562693fd7a95568f52d472d06",
			valid:     true,
		},
		{
			scalarHex: "0ce4ab8a2f342f7ff2fd10b1b734d5333c7f8240075b2ae8522313d192d8a01d",
			valid:     false,
		},
		{
			scalarHex: "745a1228beb514bac4521c342b345f301fb28e7bfebb85d3c22497d4474fe904",
			valid:     true,
		},
		{
			scalarHex: "684cfecbb321a4236354e94fd8bc74c22a28f1a3ee7e21193b0cf264204fce51",
			valid:     false,
		},
		{
			scalarHex: "14a31b0cf4630e8e74730979d661c1902ab2f47823244ff7809be751b273b00c",
			valid:     true,
		},
		{
			scalarHex: "6c0c31a9e40faf4ead079582dae043bee80f27fbb8b76ffa5c960083d578ac0f",
			valid:     true,
		},
		{
			scalarHex: "ef49cd4030926ad3dbc08425fedfed7de48a1ca119259abec5600be5b60deb0f",
			valid:     true,
		},
		{
			scalarHex: "b432db4174848e42b9a680c74d66e746a48cb397b4008815064178cf7acec91c",
			valid:     false,
		},
		{
			scalarHex: "8c041643a91de5b9a652c3edb1b2512656daf33521197f7ab01056aa04d33708",
			valid:     true,
		},
		{
			scalarHex: "d683e2784298e070cd686eddc848e1ce3205c42822650cf9bc2301fb9a330564",
			valid:     false,
		},
		{
			scalarHex: "48ad5703dbf88a809d1bde1d295c4e51e3cf24215a3136c251c165505f80e84e",
			valid:     false,
		},
		{
			scalarHex: "19d22f694db4b90b5500af61b7d6a26d5568b4a3048ddef5709698ed6df8820b",
			valid:     true,
		},
		{
			scalarHex: "f8d5164c397f22c6b8cf3ae26ebfa4a9fe893afc2d7aec1f7c4f1e46c9eca306",
			valid:     true,
		},
		{
			scalarHex: "64a5a328e130e416cf1b3fb3accd3bd032f66df404f5364e0ec9d18d652196c2",
			valid:     false,
		},
		{
			scalarHex: "300931fb2c3670af36168883af39e10e1593945114e11c4eee9547d253f4f504",
			valid:     true,
		},
		{
			scalarHex: "0bbe9316029ea6d59883574d39d1fcd3af3344d73033e2b615f0883361a11602",
			valid:     true,
		},
		{
			scalarHex: "75609b2f7a8c04053ba7c1dbfc1b7c2a8f00720817d1abd2dc1db11645bbaaa1",
			valid:     false,
		},
		{
			scalarHex: "fa98e50dd3e3fa8308ffe861194920ec6b4f431eb34c67d7332a92830a7cb20d",
			valid:     true,
		},
		{
			scalarHex: "50b48576275008950b552a94e32d25c0835cb6dd44e146519848c9a7ee003707",
			valid:     true,
		},
		{
			scalarHex: "032487e8b30ddec54b5cf90cbc38dde71104169e48228c2336916f1575c1c005",
			valid:     true,
		},
		{
			scalarHex: "574fe8ee31b42bcb9224c34c34ddca2908db4b5f4f630ce76244d1257f81e504",
			valid:     true,
		},
		{
			scalarHex: "a4bba7f4765850dd77a67e56cf89e3c8967ba8f3c72ba14c7ea014202f81780b",
			valid:     true,
		},
		{
			scalarHex: "90eca818a69f6b8c52bddee273f37eaa96c58f2322fa5f09de36d606edc2ef04",
			valid:     true,
		},
		{
			scalarHex: "ac689dc28cb94239b78e6dd66249feec0519686e0a3f974ef8d4760d320de842",
			valid:     false,
		},
		{
			scalarHex: "a578412ffcd608f1fc4d77b954b5650d5016a24747e08664373a6c1c8d61c71a",
			valid:     false,
		},
		{
			scalarHex: "d5b2f9a4b3ae8dcfc1a2ade16b320d7d00180ad72e9ba500cfeceb4e45335c00",
			valid:     true,
		},
		{
			scalarHex: "ba26b0105b5d74df01e27ac128303cf4f139b2d164772ce56a31b5787b5d2207",
			valid:     true,
		},
		{
			scalarHex: "fb9e9e404f5901ffc02ba848ed33c98a88365ff161b4d484f67f8d4d2bc9fe85",
			valid:     false,
		},
		{
			scalarHex: "e326865f9ee2b68fdff8c36d39672f897669d9efb52bd4b4d037f4e219f72887",
			valid:     false,
		},
		{
			scalarHex: "8d911a2a666472a69614bbdd887169936e4e4384992cffac6f5c8ad3f81da708",
			valid:     true,
		},
		{
			scalarHex: "a9c04591640432f35dcdcdbe5027080c31613253f01a8e4fb1c2ed04f3e8c2a1",
			valid:     false,
		},
		{
			scalarHex: "e5e18376596aa363f1de26fb318eadaee85792bb7e51af164ba80da6e358ac03",
			valid:     true,
		},
		{
			scalarHex: "2632ef6ab45b46f2a48b5691c3e82e54e50ba74f23751aa602ec77a39596a10d",
			valid:     true,
		},
		{
			scalarHex: "f110202b14458789eb17321f757032e3515012143e8dd156a401d0b48ebb6287",
			valid:     false,
		},
		{
			scalarHex: "82d4274f7518f0cf92637657c69407a5fccb685ce9de8fa617a6f6f6b86c290d",
			valid:     true,
		},
		{
			scalarHex: "e1d9512e60b4dab6b55f14f56c320ad8eda6b8e378c5f873ce5d89c622c76c0c",
			valid:     true,
		},
		{
			scalarHex: "adf9a384148ff31f52f4e2365c5234b93edd574df1b5b664650859a764130f07",
			valid:     true,
		},
		{
			scalarHex: "f5c8043c4b5a3a1ded6d02af7e048d298872e8fedbdfdb544b881215313956f9",
			valid:     false,
		},
		{
			scalarHex: "cd7d1bd7c2bf26b7060d2407dbd88a5ba4add8044436bcbf6e00e690eae29699",
			valid:     false,
		},
		{
			scalarHex: "af5b7ede9808e45e80f427d27e660443851ee2063d53c23dacb487b6b375807f",
			valid:     false,
		},
		{
			scalarHex: "3a854a6125fbae08eeaf8f29177b83cec19f8f23f82e1c0213acfda03a19a057",
			valid:     false,
		},
		{
			scalarHex: "307e8bc876b71a1ad7da7e3f8687d3173a6605184b912d80b34cec7abd09fd05",
			valid:     true,
		},
		{
			scalarHex: "5936e762cfe201af05d2e7082a10e5f58cb71cee179406eb026b05437211f5a0",
			valid:     false,
		},
		{
			scalarHex: "c02d9bb9b3ce61da67d7eb839ad6135b28485be17b6749a5d07a3a442e33890a",
			valid:     true,
		},
		{
			scalarHex: "766459f999aa16992ec9639aed4e1ecdd7b78c6ab7cc63400f2f2dbad02796e7",
			valid:     false,
		},
		{
			scalarHex: "39be058fddfed6dee0840cd03f7c407c3a30b0a0e87cc8ada9772b4150fd9a07",
			valid:     true,
		},
		{
			scalarHex: "18c994dad23375e3860e21ce5d430ef2b876820bd0fe9e84a3b3a5ba45c790c5",
			valid:     false,
		},
		{
			scalarHex: "0583e64db042f0e22e7abd759ef1b51ab5155314caf362de34781e8da9b4a20b",
			valid:     true,
		},
		{
			scalarHex: "8d71c7122df856668387356a704a0bcd1865493b65ab254b24c9bd06c5414386",
			valid:     false,
		},
		{
			scalarHex: "ca94f5f35fe069a057e9a9431e7ea3b99c7efd9700c0f838d3b99ab82648d343",
			valid:     false,
		},
		{
			scalarHex: "850d3c5acb9103049e0ce06fcd64a543a38bf20fb31ce3a47e275d1e043c5506",
			valid:     true,
		},
		{
			scalarHex: "06580accc5c48aa72ce261547614b8629c10e2ee8d62971f8453b2007f91a60f",
			valid:     true,
		},
		{
			scalarHex: "bef7fb38fcb2d6a5c22f1db03c4c26a0f61eaaa1e7f846b2bb46ff06aa9ca407",
			valid:     true,
		},
		{
			scalarHex: "fc7956e14919bb2f531c6cba13e0bf3568db3927b24a2f9131e0c42aac0cbbb1",
			valid:     false,
		},
		{
			scalarHex: "bf2ed7b0fdd8cc1222db5ae9e699de5d54abee91a3f8bd988d248ab041738aa3",
			valid:     false,
		},
		{
			scalarHex: "126213a8e3de2ab1fc862ad8bb27d3b13de5ba55657bf6e04efcf4015dd70895",
			valid:     false,
		},
		{
			scalarHex: "d1729b969a9395788931713933c587ef3748c96976a488f6a1e2f1e1ed6973d0",
			valid:     false,
		},
		{
			scalarHex: "2448345de229bb4c2695f9653653301c4562ff4a60c568a3f483474b6800875c",
			valid:     false,
		},
		{
			scalarHex: "e72284802f2737c293838a6357251398101ec60ba6e8458304bbd25de570010b",
			valid:     true,
		},
		{
			scalarHex: "87c0898c4c283f24b7164fa8aaab9fc3b06ad5237b389d9735d58c54dd989186",
			valid:     false,
		},
		{
			scalarHex: "4cceb19c61263feaffeedc984280a6f5ecc1eeebeed05a407359661d61bdf8a6",
			valid:     false,
		},
		{
			scalarHex: "2e286296d5300eb4ae7200036acbd47a319d3f9a8eb66a986a36683b85b26002",
			valid:     true,
		},
		{
			scalarHex: "1b7140ff3a09a82554d4219425c2213abd16d9f342e30e2f7c05f2a5188432a4",
			valid:     false,
		},
		{
			scalarHex: "c0cdfe0ac8dd014bc1fb4c5d3f22d74f30623054f3008c0d1e34af4f2eafdec4",
			valid:     false,
		},
		{
			scalarHex: "0000000000000000000000000000000000000000000000000000000000000000",
			valid:     true,
		},
		{
			scalarHex: "0100000000000000000000000000000000000000000000000000000000000000",
			valid:     true,
		},
		{
			scalarHex: "0200000000000000000000000000000000000000000000000000000000000000",
			valid:     true,
		},
		{
			scalarHex: "0300000000000000000000000000000000000000000000000000000000000000",
			valid:     true,
		},
		{
			scalarHex: "0400000000000000000000000000000000000000000000000000000000000000",
			valid:     true,
		},
		{
			scalarHex: "0500000000000000000000000000000000000000000000000000000000000000",
			valid:     true,
		},
		{
			scalarHex: "0600000000000000000000000000000000000000000000000000000000000000",
			valid:     true,
		},
		{
			scalarHex: "0700000000000000000000000000000000000000000000000000000000000000",
			valid:     true,
		},
		{
			scalarHex: "0800000000000000000000000000000000000000000000000000000000000000",
			valid:     true,
		},
		{
			scalarHex: "0900000000000000000000000000000000000000000000000000000000000000",
			valid:     true,
		},
		{
			scalarHex: "0a00000000000000000000000000000000000000000000000000000000000000",
			valid:     true,
		},
		{
			scalarHex: "0b00000000000000000000000000000000000000000000000000000000000000",
			valid:     true,
		},
		{
			scalarHex: "0c00000000000000000000000000000000000000000000000000000000000000",
			valid:     true,
		},
		{
			scalarHex: "0d00000000000000000000000000000000000000000000000000000000000000",
			valid:     true,
		},
		{
			scalarHex: "0e00000000000000000000000000000000000000000000000000000000000000",
			valid:     true,
		},
		{
			scalarHex: "0f00000000000000000000000000000000000000000000000000000000000000",
			valid:     true,
		},
		{
			scalarHex: "1000000000000000000000000000000000000000000000000000000000000000",
			valid:     true,
		},
		{
			scalarHex: "1100000000000000000000000000000000000000000000000000000000000000",
			valid:     true,
		},
		{
			scalarHex: "1200000000000000000000000000000000000000000000000000000000000000",
			valid:     true,
		},
		{
			scalarHex: "1300000000000000000000000000000000000000000000000000000000000000",
			valid:     true,
		},
		{
			scalarHex: "d9d3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     true,
		},
		{
			scalarHex: "dad3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     true,
		},
		{
			scalarHex: "dbd3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     true,
		},
		{
			scalarHex: "dcd3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     true,
		},
		{
			scalarHex: "ddd3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     true,
		},
		{
			scalarHex: "ded3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     true,
		},
		{
			scalarHex: "dfd3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     true,
		},
		{
			scalarHex: "e0d3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     true,
		},
		{
			scalarHex: "e1d3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     true,
		},
		{
			scalarHex: "e2d3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     true,
		},
		{
			scalarHex: "e3d3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     true,
		},
		{
			scalarHex: "e4d3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     true,
		},
		{
			scalarHex: "e5d3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     true,
		},
		{
			scalarHex: "e6d3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     true,
		},
		{
			scalarHex: "e7d3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     true,
		},
		{
			scalarHex: "e8d3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     true,
		},
		{
			scalarHex: "e9d3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     true,
		},
		{
			scalarHex: "ead3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     true,
		},
		{
			scalarHex: "ebd3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     true,
		},
		{
			scalarHex: "ecd3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     true,
		},
		{
			scalarHex: "edd3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     false,
		},
		{
			scalarHex: "eed3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     false,
		},
		{
			scalarHex: "efd3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     false,
		},
		{
			scalarHex: "f0d3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     false,
		},
		{
			scalarHex: "f1d3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     false,
		},
		{
			scalarHex: "f2d3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     false,
		},
		{
			scalarHex: "f3d3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     false,
		},
		{
			scalarHex: "f4d3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     false,
		},
		{
			scalarHex: "f5d3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     false,
		},
		{
			scalarHex: "f6d3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     false,
		},
		{
			scalarHex: "f7d3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     false,
		},
		{
			scalarHex: "f8d3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     false,
		},
		{
			scalarHex: "f9d3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     false,
		},
		{
			scalarHex: "fad3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     false,
		},
		{
			scalarHex: "fbd3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     false,
		},
		{
			scalarHex: "fcd3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     false,
		},
		{
			scalarHex: "fdd3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     false,
		},
		{
			scalarHex: "fed3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     false,
		},
		{
			scalarHex: "ffd3f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     false,
		},
		{
			scalarHex: "00d4f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     false,
		},
		{
			scalarHex: "01d4f55c1a631258d69cf7a2def9de1400000000000000000000000000000010",
			valid:     false,
		},
		{
			scalarHex: "ecffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			valid:     false,
		},
		{
			scalarHex: "edffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			valid:     false,
		},
		{
			scalarHex: "eeffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			valid:     false,
		},
		{
			scalarHex: "efffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			valid:     false,
		},
		{
			scalarHex: "f0ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			valid:     false,
		},
		{
			scalarHex: "f1ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			valid:     false,
		},
		{
			scalarHex: "f2ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			valid:     false,
		},
		{
			scalarHex: "f3ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			valid:     false,
		},
		{
			scalarHex: "f4ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			valid:     false,
		},
		{
			scalarHex: "f5ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			valid:     false,
		},
		{
			scalarHex: "f6ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			valid:     false,
		},
		{
			scalarHex: "f7ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			valid:     false,
		},
		{
			scalarHex: "f8ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			valid:     false,
		},
		{
			scalarHex: "f9ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			valid:     false,
		},
		{
			scalarHex: "faffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			valid:     false,
		},
		{
			scalarHex: "fbffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			valid:     false,
		},
		{
			scalarHex: "fcffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			valid:     false,
		},
		{
			scalarHex: "fdffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			valid:     false,
		},
		{
			scalarHex: "feffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			valid:     false,
		},
		{
			scalarHex: "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			valid:     false,
		},
	}
	for _, test := range tests {
		scalar := HexToKey(test.scalarHex)
		got := ScValid(&scalar)
		if test.valid != got {
			t.Errorf("%x: want %t, got %t", scalar, test.valid, got)
		}
	}
}

func TestAddKeys3_3(t *testing.T) {

	for i := 0; i < 1000; i++ {
		s1 := *(RandomScalar())
		s2 := *(RandomScalar()) //*(identity()) // *(RandomScalar())

		p1 := s1.PublicKey()
		p2 := s2.PublicKey()

		first_part := ScalarMultKey(p1, &s1)
		second_part := ScalarMultKey(p2, &s2)

		// do it traditional way
		var sum_result Key
		AddKeys(&sum_result, first_part, second_part)

		// lets do it optimal way, pre compute
		var p1_extended, p2_extended ExtendedGroupElement
		var p1_precomputed, p2_precomputed [8]CachedGroupElement

		p1_extended.FromBytes(p1)
		p2_extended.FromBytes(p2)

		GePrecompute(&p1_precomputed, &p1_extended)
		GePrecompute(&p2_precomputed, &p2_extended)

		var fast_result Key
		AddKeys3_3(&fast_result, &s1, &p1_precomputed, &s2, &p2_precomputed)

		if sum_result != fast_result {
			t.Fatalf("AddKeys3_3 Failed Expected %s Actual %s", sum_result, fast_result)
		}
	}

}

func TestElements(t *testing.T) {

	for i := 0; i < 100; i++ {
		p1 := *(RandomScalar()).PublicKey()
		p2 := *(RandomScalar()).PublicKey()

		var exp1, exp2 ExtendedGroupElement
		var cached CachedGroupElement
		var precomputed PreComputedGroupElement
		var actual, naive, fast Key
		var tmpe ExtendedGroupElement

		exp1.FromBytes(&p1)
		exp2.FromBytes(&p2)

		//original
		AddKeys(&actual, &p1, &p2)

		// naive
		var r CompletedGroupElement
		exp2.ToCached(&cached)
		geAdd(&r, &exp1, &cached)
		r.ToExtended(&tmpe)
		tmpe.ToBytes(&naive)

		if actual != naive {
			t.Fatalf("Naive addition failed")
		}

		exp2.ToPreComputed(&precomputed)
		geMixedAdd(&r, &exp1, &precomputed)
		r.ToExtended(&tmpe)
		tmpe.ToBytes(&fast)

		if actual != fast {
			t.Fatalf("Fast addition failed")
		}
	}
}

func TestPrecompute(t *testing.T) {

	var table PRECOMPUTE_TABLE

	s1 := *(RandomScalar())
	//p1 := &GBASE //identity() // s1.PublicKey()
	p1 := s1.PublicKey()

	GenPrecompute(&table, *p1)

	//s1[1]=29

	expected := ScalarMultKey(p1, &s1)

	var actual Key

	var result_extended ExtendedGroupElement
	ScalarMultPrecompute(&result_extended, &s1, &table)

	result_extended.ToBytes(&actual)

	//t.Logf("Super compute Expected %s actual %s", expected,actual)

	if actual != *expected {
		t.Logf("Super compute failed Expected %s actual %s", expected, actual)
	}

}

func TestSuperPrecompute(t *testing.T) {
	var table PRECOMPUTE_TABLE
	var stable SUPER_PRECOMPUTE_TABLE

	s1 := *(RandomScalar())
	p1 := s1.PublicKey()

	GenPrecompute(&table, *p1)
	GenSuperPrecompute(&stable, &table)

	//s1[1]=29

	expected := ScalarMultKey(p1, &s1)

	var actual Key

	var result_extended ExtendedGroupElement
	ScalarMultSuperPrecompute(&result_extended, &s1, &stable)

	result_extended.ToBytes(&actual)

	//t.Logf("Super compute Expected %s actual %s", expected,actual)

	if actual != *expected {
		t.Logf("Super compute failed Expected %s actual %s", expected, actual)
	}

}

func Test_DoubleScalarDoubleBaseMulPrecomputed(t *testing.T) {

	var ex ExtendedGroupElement

	s1 := *(RandomScalar())
	s2 := *(RandomScalar()) //*(RandomScalar()) //*(identity()) // *(RandomScalar())

	p1 := s1.PublicKey()
	p2 := s2.PublicKey()

	first_part := ScalarMultKey(p1, &s1)
	second_part := ScalarMultKey(p2, &s2)

	// do it traditional way
	var sum_result Key
	var actual Key
	AddKeys(&sum_result, first_part, second_part)

	// lets do it using precomputed tables
	var table PRECOMPUTE_TABLE

	GenDoublePrecompute(&table, *p1, *p2)

	/*
	   multprecompscalar(&ex,&s1)


	   ex.ToBytes(&actual)

	   if *first_part != actual{

	   	//t.Logf("%+v table ", table)
	   	t.Fatalf("simple scalar precompyed failed  expected %s  actual %s", sum_result,actual)
	   }
	*/

	DoubleScalarDoubleBaseMulPrecomputed(&ex, &s1, &s2, &table)
	ex.ToBytes(&actual)

	//  t.Logf("first part %s", first_part)
	//  t.Logf("second_part part %s", second_part)
	//  t.Logf("actual %s expected %s", actual,sum_result)

	if sum_result != actual {

		//t.Logf("%+v table ", table)
		t.Fatalf("Double scalar precompyed failed %s %s", sum_result, actual)
	}

}

func BenchmarkDoublePrecompute(b *testing.B) {
	var table PRECOMPUTE_TABLE

	s1 := *(identity()) //  *(RandomScalar())
	s2 := *(identity()) //*(RandomScalar()) //*(identity()) // *(RandomScalar())

	p1 := identity() //s1.PublicKey()
	p2 := s2.PublicKey()

	GenDoublePrecompute(&table, *p1, *p2)

	cpufile, err := os.Create("/tmp/dprecompute_cpuprofile.prof")
	if err != nil {

	}
	if err := pprof.StartCPUProfile(cpufile); err != nil {
	}
	defer pprof.StopCPUProfile()

	var result_extended ExtendedGroupElement
	result_extended.Zero() // make it identity

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		DoubleScalarDoubleBaseMulPrecomputed(&result_extended, &s1, &s2, &table)

	}
}

// test 64 bit version used for bulletproofs
func Test_DoubleScalarDoubleBaseMulPrecomputed64(t *testing.T) {

	var s1, s2 [64]Key
	var p1, p2 [64]Key

	for i := 0; i < 64; i++ {
		s1[i] = *(RandomScalar())
		s2[i] = *(RandomScalar()) //*(RandomScalar()) //*(identity()) // *(RandomScalar())

		p1[i] = *(s1[i].PublicKey())
		p2[i] = *(s2[i].PublicKey())

	}

	// compute actual result using naive method
	naive_result := Identity
	for i := 0; i < 64; i++ {

		first_part := ScalarMultKey(&p1[i], &s1[i])
		second_part := ScalarMultKey(&p2[i], &s2[i])

		// do it traditional way
		var sum_result Key
		AddKeys(&sum_result, first_part, second_part)
		AddKeys(&naive_result, &naive_result, &sum_result)

	}

	// lets do it using precomputed tables
	var table [64]PRECOMPUTE_TABLE
	for i := 0; i < 64; i++ {
		GenDoublePrecompute(&table[i], p1[i], p2[i])
	}

	var ex ExtendedGroupElement
	var actual Key
	DoubleScalarDoubleBaseMulPrecomputed64(&ex, s1[:], s2[:], table[:])

	ex.ToBytes(&actual)

	//  t.Logf("first part %s", first_part)
	//  t.Logf("second_part part %s", second_part)
	//  t.Logf("actual %s expected %s", actual,sum_result)

	if naive_result != actual {

		//t.Logf("%+v table ", table)
		t.Fatalf("Double scalar precomputed 64 failed  expected %s actual %s", naive_result, actual)
	}

}

// test 64 bit version used for bulletproofs
func Benchmark_DoubleScalarDoubleBaseMulPrecomputed64(b *testing.B) {

	var s1, s2 [64]Key
	var p1, p2 [64]Key

	for i := 0; i < 64; i++ {
		s1[i] = *(RandomScalar())
		s2[i] = *(RandomScalar()) //*(RandomScalar()) //*(identity()) // *(RandomScalar())

		p1[i] = *(s1[i].PublicKey())
		p2[i] = *(s2[i].PublicKey())

	}

	// compute actual result using naive method
	naive_result := Identity
	for i := 0; i < 64; i++ {

		first_part := ScalarMultKey(&p1[i], &s1[i])
		second_part := ScalarMultKey(&p2[i], &s2[i])

		// do it traditional way
		var sum_result Key
		AddKeys(&sum_result, first_part, second_part)
		AddKeys(&naive_result, &naive_result, &sum_result)

	}

	// lets do it using precomputed tables
	var table [64]PRECOMPUTE_TABLE
	for i := 0; i < 64; i++ {
		GenDoublePrecompute(&table[i], p1[i], p2[i])
	}

	var ex ExtendedGroupElement
	var actual Key

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		DoubleScalarDoubleBaseMulPrecomputed64(&ex, s1[:], s2[:], table[:])

		ex.ToBytes(&actual)

		//  t.Logf("first part %s", first_part)
		//  t.Logf("second_part part %s", second_part)
		//  t.Logf("actual %s expected %s", actual,sum_result)

		if naive_result != actual {

			//t.Logf("%+v table ", table)
			b.Fatalf("Double scalar precomputed 64 failed  expected %s actual %s", naive_result, actual)
		}
	}

}

func BenchmarkPrecompute(b *testing.B) {
	var table PRECOMPUTE_TABLE

	GenPrecompute(&table, GBASE)

	s1 := *(RandomScalar())
	//s1[1]=29

	/*
	   	cpufile,err := os.Create("/tmp/precompute_cpuprofile.prof")
	   			if err != nil{

	   			}
	   			if err := pprof.StartCPUProfile(cpufile); err != nil {
	               }
	           	defer pprof.StopCPUProfile()

	*/

	var result_extended ExtendedGroupElement
	result_extended.Zero() // make it identity

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result_extended.Zero() // make it identity

		ScalarMultPrecompute(&result_extended, &s1, &table)

	}
}

func BenchmarkSuperPrecompute(b *testing.B) {
	var table PRECOMPUTE_TABLE
	var stable SUPER_PRECOMPUTE_TABLE

	GenPrecompute(&table, GBASE)
	GenSuperPrecompute(&stable, &table)

	s1 := *(RandomScalar())
	//s1[1]=29

	/*
	   	cpufile,err := os.Create("/tmp/superprecompute_cpuprofile.prof")
	   			if err != nil{

	   			}
	   			if err := pprof.StartCPUProfile(cpufile); err != nil {
	               }
	           	defer pprof.StopCPUProfile()

	*/

	var result_extended ExtendedGroupElement
	result_extended.Zero() // make it identity

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		result_extended.Zero() // make it identity

		ScalarMultSuperPrecompute(&result_extended, &s1, &stable)

	}
}

func BenchmarkGeScalarMultBase(b *testing.B) {
	var s Key
	rand.Reader.Read(s[:])
	var P ExtendedGroupElement

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GeScalarMultBase(&P, &s)
	}
}

func BenchmarkGeScalarMult(b *testing.B) {
	var s Key
	rand.Reader.Read(s[:])

	var P ExtendedGroupElement
	var E ProjectiveGroupElement
	s[31] &= 127
	GeScalarMultBase(&P, &s)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GeScalarMult(&E, &s, &P)
	}
}

func BenchmarkGeDoubleScalarMultVartime(b *testing.B) {
	var s Key
	rand.Reader.Read(s[:])

	var P, Pout ExtendedGroupElement
	s[31] &= 127
	GeScalarMultBase(&P, &s)
	_ = Pout

	var out ProjectiveGroupElement

	b.ResetTimer()
	var x Key
	for i := 0; i < b.N; i++ {
		GeDoubleScalarMultVartime(&out, &s, &P, &x)
		//out.ToExtended(&Pout)
	}
}

/*
func BenchmarkGeAdd(b *testing.B) {
	var s Key
	rand.Reader.Read(s[:])

	var R, P ExtendedGroupElement
	s[31] &= 127
	GeScalarMultBase(&P, &s)
	R = P

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GeAdd(&R, &R, &P)
	}
}
*/

/*
func BenchmarkGeDouble(b *testing.B) {
	var s [32]byte
	rand.Reader.Read(s[:])

	var R, P ExtendedGroupElement
	s[31] &= 127
	GeScalarMultBase(&P, &s)
	R = P

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GeDouble(&R, &P)
	}
}
*/

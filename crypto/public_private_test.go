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

//import "fmt"
import "testing"

// these test were orignally written for mnemonics, but they can be used here also

func Test_Public_Private_Key(t *testing.T) {

	tests := []struct {
		name             string
		seed             string
		spend_key_secret string
		spend_key_public string
		view_key_secret  string
		view_key_public  string
		Address          string
	}{
		{
			name:             "English",
			seed:             "sequence atlas unveil summon pebbles tuesday beer rudely snake rockets different fuselage woven tagged bested dented vegan hover rapid fawns obvious muppet randomly seasons randomly",
			spend_key_secret: "b0ef6bd527b9b23b9ceef70dc8b4cd1ee83ca14541964e764ad23f5151204f0f",
			spend_key_public: "7d996b0f2db6dbb5f2a086211f2399a4a7479b2c911af307fdc3f7f61a88cb0e",
			view_key_secret:  "42ba20adb337e5eca797565be11c9adb0a8bef8c830bccc2df712535d3b8f608",
			view_key_public:  "1c06bcac7082f73af10460b5f2849aded79374b2fbdaae5d9384b9b6514fddcb",
			Address:          "dETocsF4EuzXaxLNbDLLWi6xNEzzBJ2He5WSf7He8peuPt4nTyakAFyNuXqrHAGQt1PBSBonCRRj8daUtF7TPXFW42YQkxUQzg",
		},

		{
			name:             "Deutsch",
			seed:             "Dekade Spagat Bereich Radclub Yeti Dialekt Unimog Nomade Anlage Hirte Besitz Märzluft Krabbe Nabel Halsader Chefarzt Hering tauchen Neuerung Reifen Umgang Hürde Alchimie Amnesie Reifen",
			spend_key_secret: "a00b3c431e0037426f12b255aaca918863c8bbc690ff3765564bcc1de7fbb303",
			spend_key_public: "6f6d202f715d23ff4034a42298fd7b36431dbbfe8559df642911a9c3143ab802",
			view_key_secret:  "eb2e288ac34e8d4f51cf73b8fda5240c2f3a25f72f4e7b4ac33f33fdc6602300",
			view_key_public:  "ec9769a6fc9a10a474ba27bb62a58cb55e51aca4107b3daa613c054e4a104d75",
			Address:          "dETobGVx1rYGaagDmPbfGjjQ5zeVkHXDeG2ssGnyrWgbApeTgRHhPJBSmdAj76XJFKUh3FtNv7fjMMcbGsLnms8V1cqG9Euc5i",
		},

		{
			name:             "Español",
			seed:             "perfil lujo faja puma favor pedir detalle doble carbón neón paella cuarto ánimo cuento conga correr dental moneda león donar entero logro realidad acceso doble",
			spend_key_secret: "4f1101c4cc6adc6e6a63acde4e71fd76dc4471fa54769866d5e80a0a3d53d00c",
			spend_key_public: "7f28c0b9c5e59fb57a6c4171f6b19faa369bd50c53b792fcdf2402369ded26a7",
			view_key_secret:  "2de547b9fa3be9db08871d3964027a8f0d8d9c36755257841ba838cfff9fd60a",
			view_key_public:  "caf4787a2eb8aa94e11a5a7ac912daa820a8148ea175b03a13d4322adab0c33f",
			Address:          "dETod3T65GgfQe4xkLfu2mWiAzdR6ZRPgXhuAdJqKRUpgffP7dP3C5BXtVhPTwLJZE49uhjBDBkLxLgj3835HxiM7hWoS7trc1",
		},

		{
			name:             "Français",
			seed:             "lisser onctueux pierre trace flair riche machine ordre soir nougat talon balle biceps crier trame tenu gorge cuisine taverne presque laque argent roche secte ordre",
			spend_key_secret: "340ed50ef73ba172e2fef23dc9f60e314f609bb7692d22d59f3938c391570b0a",
			spend_key_public: "1b2e329fb7fbf1e3d93a30d66e72f13e19addc49ece84839bf2772d685b1bf20",
			view_key_secret:  "415068b05bfe34ee250b1b0fb0847a7cef7bc1e99ab407dedcf779348097210b",
			view_key_public:  "a7d037f5f73677c4d1ba32559238886c9011d682121946cf06f9bb59efe75862",
			Address:          "dEToRmE1GKxj9BVmA46whoLE5vKNnBH6BfrRoaoGLig4WjN9WHF3FCJA7QZwkkGP1KATSXC7cLB9s5EDT5Xfczdk9mV1pUkkUg",
		},

		{
			name:             "Italiano",
			seed:             "sospiro uomo sommario orecchio muscolo testa avido sponda mutande levare lamento frumento volpe zainetto ammirare stufa convegno patente salto venire pianeta marinaio minuto moneta moneta",
			spend_key_secret: "21cb90e631866952954cdcba042a3ae90407c40c052cc067226df0b454933502",
			spend_key_public: "540850fa07de29317467980349436266652e1a0d4ba9b758568428a5881a6e4a",
			view_key_secret:  "a106482cc83587c3b39f152ddfa0a447a7647b7977deb400e56166fa7fcc9c0a",
			view_key_public:  "e076d540a85fb63fdb417304effa0541de65b80883b7b773b5d736a7face10dc",
			Address:          "dEToYBF9F58eAEsRFgQjaYCGiM1ngwjiSVPTfH6c1A5D5RQuWLjg7cBH1XUaQnY9v6ipWigbhHQN6XjHK6yXdUws8ovPF3F19g",
		},

		{
			name:             "Nederlands",
			seed:             "veto hobo tolvrij cricket somber omaans lourdes kokhals ionisch lipman freon neptunus zimmerman rijbaan wisgerhof oudachtig nerd walraven ruis gevecht foolen onheilig rugnummer russchen cricket",
			spend_key_secret: "921cbd7640df5fd12effb8f4269c5a47bac0ef3f75a0c25aa9f174f589801102",
			spend_key_public: "f136d2467e0e1826218b83d148374cd215358d1fe6951eab0b6046e632170072",
			view_key_secret:  "ff10ba6ee19a563b60ac8ab5d94916993a1ff05c644c3f73056c91edd1423b06",
			view_key_public:  "7506289e53bcfaee88ce5ff5a83d2818d066db2b8d776a15808e5c3fd9a49cde",
			Address:          "dEToqunYRNu3MjViW649P5AFUquDmjZW5RwedfYZ6xbT4r9TKMQVk34YcLzXmSx4CPBEJARk23ZKzLyUDJMAzJdN7EovWv3GRd",
		},

		{
			name:             "Português",
			seed:             "guloso caatinga enunciar newtoniano aprumo ilogismo vazio gibi imovel mixuruca bauxita paludismo unanimidade zumbi vozes roer anzol leonardo roer ucraniano elmo paete susto taco imovel",
			spend_key_secret: "92da2722c7138e65559973131fd2c69b4a0ae4faf0e12b7abe096d5870d6a700",
			spend_key_public: "f2daf0540863ae830c6f994fc467900f77dfac153f36e84e9e0eaa074333a3e9",
			view_key_secret:  "2de03327764d8d8fdb210560ce3e78b6698023ea9993a104afccf34495a82506",
			view_key_public:  "2e0db07e7ea17fd49a2a92d7ca5071445033c0107f7d4f5adb9cb063b7aab7ea",
			Address:          "dERopX3LiCoHg3AFLgZB5dJKgvLESnnqLABfvPAM4u869dya9oocJGVU1kG8wQjC25ETPmRyMKw98MxfVwqxbmdY7UEGsR2czw",
		},

		{
			name:             "русский язык",
			seed:             "шорох рента увлекать пешеход гонка сеять пчела ваза апатия пишущий готовый вибрация юбка здоровье машина штука охрана доза рынок клоун рецепт отпуск шестерка эволюция вибрация",
			spend_key_secret: "3d0fb729be695b865c073eed68ee91f06d429a27f8eaaaa6a99f954edbef8406",
			spend_key_public: "c564b0e4d2992b534f796d70d38d283e3fb53cee717ecd1aad1156e2b79abc85",
			view_key_secret:  "7c048a8c4f9b1aaba092f5cbc07bded9bfe35fcad7ce46634639d042011b8b04",
			view_key_public:  "8d9f057765ab73b2ae890c5073737ba009539c8b4746a5e58e74a6cb7150a8c2",
			Address:          "dERojPZ7WjBScw9H8PS6oCQcQJBXJaQdNND8ZYbYBr2SSt8wzG8HFg8VgJKcsDonhYLKL5Q71UNWACpNiaESzsJx44HUmJzaWW",
		},

		{
			name:             "日本語",
			seed:             "かわく ねまき けもの せいげん ためる にんめい てあみ にりんしゃ さわやか えらい うちき けいかく あたる せっきゃく ずっしり かいよう おおや てらす くれる ばかり なこうど たいうん そまつ たいえき せいげん",
			spend_key_secret: "e12da07065554a32ac798396f74dbb35557164f9f39291a3f95705e62e0d7703",
			spend_key_public: "09cc0d2adecd40f1118dffce60d1f5d6876ed898ef9360e867bc1e52df36b6fb",
			view_key_secret:  "90346fc93565c41686eb23fe4fae22a89e59f2ea14d0ba8ad3b16a99ebd80e0b",
			view_key_public:  "982046c49edc082df0e44a739b831a7c0a325766f9c4fa09fbf8c984a8c6bc06",
			Address:          "dERoNDzRBZfbLDdeoKD8Zdc7svU36AKPtReksucYRpy4A9oWVcc5dFsdoaxnS7ddmtNvsL5w1eWkYZwvQqApAg8X8XomwZSLws",
		},

		{
			name:             "简体中文 (中国)",
			seed:             "启 写 输 苯 加 担 乳 代 集 预 懂 均 倒 革 熟 载 隆 台 幸 谋 轮 抚 扩 急 输",
			spend_key_secret: "3bfd99190f28bd7830b3631cfa514176fc24e88281fe056ce447a5a7fcdc9a02",
			spend_key_public: "f27221193c3ab0709009e2225578ff93d86efc178ebc0b482e8d9ec7e741df40",
			view_key_secret:  "1786f5656bc093a06d5064a80bb891e2e1873699da5e3b63bb16af2bc5563b0c",
			view_key_public:  "976d8c38f15220a3ee33724c3eb5315589b41e7b2f4b0ee2be04e92edae94418",
			Address:          "dERopUMxPjhApMpS4vt2B6MEqTpvQbZo3YTGFr7MoabYC23RE7DvTQCEjju9jXHfwBXJoBfHaCVupDZAAgZpLaY99qhsXcCFmg",
		},

		{
			name:             "Esperanto",
			seed:             "amrakonto facila vibri obtuza gondolo membro alkoholo oferti ciumi reinspekti azteka kupro gombo keglo dugongo diino hemisfero sume servilo bambuo sekretario alta diurno duloka hemisfero",
			spend_key_secret: "61abd2a5625a95371882117d8652e0735779b7b535008c73d65735b9477b1105",
			spend_key_public: "74484fccc824cecdb0a9e69a163938f5d075fcf4d649444e86187cde130b2f04",
			view_key_secret:  "ba6202a8b877b6c58eb7239d9a23166202ae454a9c98de74d29ba593782ce20c",
			view_key_public:  "22437d58b154e9b5f35526dbfd6d6e71769103536de248eea98df7209de9759b",
			Address:          "dERoaEnz1jm7A5kv9xFxgdAa8Y7SCqepmDFnEo4Rkwe22sVjoBThDweFCmBjkJMLSQKJd4soX6wBierFKbDTh1SL9r8XY829pb",
		},
	}

	for _, test := range tests {
		spend_key_secret := HexToKey(test.spend_key_secret)
		spend_key_public := HexToKey(test.spend_key_public)

		if spend_key_public != *(spend_key_secret.PublicKey()) {
			t.Errorf("%s key generation testing failed ", test.name)
		}

		view_key_secret := HexToKey(test.view_key_secret)
		view_key_public := HexToKey(test.view_key_public)

		if view_key_public != *(view_key_secret.PublicKey()) {
			t.Errorf("%s key generation testing failed for spend key ", test.name)
		}

	}

}

package astrobwt

import "crypto/rand"
import "strings"
import "testing"
import "encoding/hex"

// see https://www.geeksforgeeks.org/burrows-wheeler-data-transform-algorithm/
func TestBWTTransform(t *testing.T) {

	tests := []struct {
		input string
		bwt   string
	}{
		{"BANANA", "ANNB$AA"}, // from https://www.geeksforgeeks.org/burrows-wheeler-data-transform-algorithm/
		{"abracadabra", "ard$rcaaaabb"},
		{"appellee", "e$elplepa"},
		{"GATGCGAGAGATG", "GGGGGGTCAA$TAA"},
	}
	for _, test := range tests {

		input := "\x00" + test.input + "\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"

		var output = make([]byte, 64, 64)
		sort_indices(len(test.input)+1, []byte(input), output, &Data{})

		output = output[:len(test.input)+1]
		output_s := strings.Replace(string(output), "\x00", "$", -1)

		if output_s != test.bwt {
			t.Errorf("Test failed: Transform  %s  %s  %x", output_s, test.bwt, output)
		}
	}

}

func TestPOW_optimized_v1(t *testing.T) {
	p := POW([]byte{0, 0, 0, 0})
	p0 := POW_0alloc([]byte{0, 0, 0, 0})
	p_optimized,_ := POW_optimized_v1([]byte{0, 0, 0, 0}, MAX_LENGTH)
	if string(p[:]) != string(p0[:]) {
		t.Error("Test failed: POW and POW_0alloc returns different ")
	}
	if string(p[:]) != string(p_optimized[:]) {
		t.Error("Test failed: POW and POW_rewrite returns different ")
	}

	for i := 20; i < 200; i++ {
		buf := make([]byte, 20, 20)
		rand.Read(buf)

		p := POW(buf)
		p0 := POW_0alloc(buf)
		p_optimized,_ := POW_optimized_v1(buf, MAX_LENGTH)
		if string(p[:]) != string(p0[:]) {
			t.Errorf("Test failed: POW and POW_0alloc returns different for i=%d buf %x", i, buf)
		}
		if string(p[:]) != string(p_optimized[:]) {
			t.Errorf("Test failed: POW and POW_rewrite returns different for i=%d buf %x", i, buf)
		}

	}
}

func TestPOW_optimized_v1Tests(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"57b84420d2028aef8a05b42ea893a8c4f2219f73"},
		{"67d49bd1c53645ec96c50230083e55b120b5005ffbaf4b2f"},
		{"67d49bd1c53645ec96c50230183e55b120b5005ffbaf4b2f"},
		{"3fe7baa520edb5d0b43b7a6999c146262b2c7f26e030fd7c256611262db727833a40a79f9988"},
		{"ed32ea6f1c0ee2514eb73f6a0f9f00d1e2c2392a8896963eefddfd6600c105d52db6e93ba98f8454433894293eaa9f31973658c49f67f3361af70fac27bac6f8f5c69f52be9d7c86a5fea3e5b5d99d8f73888b4d4a7dbb28169035583632ef26604e472eb26a5da9da4e95f80460dfc7b788e8ee75194a7d2f1190a788f92e98cb83fd4c63d9976ec06c2df005d321baed360599af58ff45aa63b00261ea60b5adf623f256bfbc75da961c5960db68e8"},
		{"7cf76f0d4072574bae246c4f7184000af5ce818943605151a73a49d7b704c127891e6e7008c331fa41776540b0db3b2ea2c187e119191adde6b0f5438fb48cc242c02420f44d070ef4c87a00952560f2ffcc5ac5932c5a0f40df9029ddc10d29b23ff4150fbe0dda5b14a73eadd90a3b6eaf049075b89c1c16da33f049c3235f158c"},
		{"fbaf4f7ebec36c97f8994e67e74b281960846e6b5ce30e4fd95ce68d8875e19ab3ebf716e5887adb6eefbc3c5ca6096f936643f4bf22a9f61a1e35b019cfaabfe331ad2897a3b70bd6846c5003a999719d26246796a1d60b18bf89bdf4f5fea3b976ad7739e00089f7f11a5833351515e330d8580f918ea694a438f384946cdae0d9d3ccda33bc6de1a64d6c25c0b3f7d905172956"},
		{"ff7f99a16b3e2c0f9daa2a44c9a364b212d836ba57f8d9b0e050490d1e74"},
		{"f299d507d916e67f93345a42042e859170eb755262355826fcb7ed0d2e9c999bb21662275d1b99a53b397bf77f4e2af38a41358c41e9ecd750f3cc2859a3fef8a9ef9c189b7489fb0048903cfe78f5171f2476f86aae2346e5390740b09bb185268af16146ccab9876d8931f670f9ba93805f0277a3cba0fc9671cc78ac53ce60f538c7aa616660e3ca1e1eabf8938c095baeacb4ce11889c52ce63b9511d2f176d563a75a34418fddcb4e712a5936e4f72a2269b423954dadcf"},
	}

	for i, test := range tests {

		buf, err := hex.DecodeString(test.input)
		if err != nil {
			t.Error(err)
		}

		p := POW(buf)
		p0 := POW_0alloc(buf)
		p_optimized,_ := POW_optimized_v1(buf, MAX_LENGTH)
		if string(p[:]) != string(p0[:]) {
			t.Errorf("Test failed: POW and POW_0alloc returns different for i=%d buf %x", i, buf)
		}
		if string(p[:]) != string(p_optimized[:]) {
			t.Errorf("Test failed: POW and POW_optimized returns different for i=%d buf %x", i, buf)
		}
	}

}

func BenchmarkPOW_optimized_v1(t *testing.B) {
	for i := 0; i < t.N; i++ {
		rand.Read(cases[0][:])
		_,_ = POW_optimized_v1(cases[0][:], MAX_LENGTH)
	}
}

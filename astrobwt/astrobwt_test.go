package astrobwt

import "math/rand"
import "testing"

// see https://www.geeksforgeeks.org/burrows-wheeler-data-transform-algorithm/
func TestBWTAndInverseTransform(t *testing.T) {

	tests := []struct {
		input string
		bwt   string
	}{
		{"BANANA", "ANNB$AA"}, // from https://www.geeksforgeeks.org/burrows-wheeler-data-transform-algorithm/
		{"abracadabra", "ard$rcaaaabb"},
		{"appellee", "e$elplepa"},
		{"GATGCGAGAGATG", "GGGGGGTCAA$TAA"},
		{"abcdefg", "g$abcdef"},
	}

	for _, test := range tests {
		trans2, eos := BWT([]byte(test.input))
		trans2[eos] = '$'

		if string(trans2) != test.bwt {
			t.Errorf("Test failed: Transform %s", test.input)
		}
		if string(InverseTransform([]byte(trans2), '$')) != test.input {
			t.Errorf("Test failed: InverseTransform  expected '%s' actual '%s`", test.input, string(InverseTransform([]byte(trans2), '$')))
		}

		p := POW([]byte(test.input))
		p0 := POW_0alloc([]byte(test.input))

		if string(p[:]) != string(p0[:]) {
			t.Error("Test failed: difference between pow and pow_0alloc")

		}
	}

}

func TestFromSuffixArray(t *testing.T) {
	s := "GATGCGAGAGATG"
	trans := "GGGGGGTCAA$TAA"

	sa := SuffixArray([]byte(s))
	B, err := FromSuffixArray([]byte(s), sa, '$')
	if err != nil {
		t.Error("Test failed: FromSuffixArray error")
	}
	if string(B) != trans {
		t.Error("Test failed: FromSuffixArray returns wrong result")
	}
}

func TestPow_Powalloc(t *testing.T) {

	p := POW([]byte{0, 0, 0, 0})
	p0 := POW_0alloc([]byte{0, 0, 0, 0})
	if string(p[:]) != string(p0[:]) {
		t.Error("Test failed: POW and POW_0alloc returns different ")
	}
}

var cases [][]byte

func init() {
	rand.Seed(1)
	alphabet := "abcdefghjijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ01234567890"
	n := len(alphabet)
	_ = n
	scales := []int{100000}
	cases = make([][]byte, len(scales))
	for i, scale := range scales {
		l := scale
		buf := make([]byte, int(l))
		for j := 0; j < int(l); j++ {
			buf[j] = byte(rand.Uint32() & 0xff) //alphabet[rand.Intn(n)]
		}
		cases[i] = buf
	}
	POW([]byte{0x99})
}

var result []byte

func BenchmarkTransform(t *testing.B) {
	var r []byte
	var err error
	for i := 0; i < t.N; i++ {
		r, err = Transform(cases[0], '$')
		if err != nil {
			t.Error(err)
			return
		}
	}
	result = r
}

func BenchmarkTransform_quick(t *testing.B) {
	var r []byte
	for i := 0; i < t.N; i++ {
		//r, err = Transform(cases[0], '$')
		r, _ = BWT(cases[0])
	}
	result = r
}

func BenchmarkPOW(t *testing.B) {
	for i := 0; i < t.N; i++ {
		rand.Read(cases[0][:])
		_ = POW(cases[0][:])
	}
}

func BenchmarkPOW_0alloc(t *testing.B) {
	for i := 0; i < t.N; i++ {
		rand.Read(cases[0][:])
		_ = POW_0alloc(cases[0][:])
	}
}

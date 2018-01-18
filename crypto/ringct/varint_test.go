package ringct

import "bytes"
import "testing"

// this package needs to be verified for bug,
// just in case, the top bit is set, it is impossible to do varint 64 bit number into 8 bytes, if the number is too big
// in that case go needs 9 bytes, we should verify whether the number can ever reach there and thus place
// suitable checks to avoid falling into the trap later on
func TestVarInt(t *testing.T) {
	tests := []struct {
		name   string
		varInt []byte
		want   uint64
	}{
		{
			name:   "1 byte",
			varInt: []byte{0x01},
			want:   1,
		},
		{
			name:   "3 bytes",
			varInt: []byte{0x8f, 0xd6, 0x17},
			want:   387855,
		},
		{
			name:   "4 bytes",
			varInt: []byte{0x80, 0x92, 0xf4, 0x01},
			want:   4000000,
		},
		{
			name:   "7 bytes",
			varInt: []byte{0x80, 0xc0, 0xca, 0xf3, 0x84, 0xa3, 0x02},
			want:   10000000000000,
		},
	}
	var got uint64
	var err error
	var gotVarInt []byte
	buf := new(bytes.Buffer)
	for _, test := range tests {
		gotVarInt = Uint64ToBytes(test.want)
		if bytes.Compare(gotVarInt, test.varInt) != 0 {
			t.Errorf("%s: varint want %x, got %x", test.name, test.varInt, gotVarInt)
			continue
		}
		buf.Reset()
		buf.Write(test.varInt)
		got, err = ReadVarInt(buf)
		if err != nil {
			t.Errorf("%s: %s", test.name, err)
			continue
		}
		if test.want != got {
			t.Errorf("%s: want %d, got %d", test.name, test.want, got)
			continue
		}
	}
}

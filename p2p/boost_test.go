package p2p

import "testing"

/* this func decode boost variant
 *
 * 0   0a 60    == (60 >> 2 ) = 24  bytes
 * 1   0a c0    == (c0 >> 2 ) = 48  bytes
 * 2   0a 21 01 == (121 >> 2 ) = 72
 * 3   0a 81 01 == (181 >> 2 ) = 96  bytes
 * 4   0a e1 01
 * 5   0a 41 02
 * 7   0a 01 03
 * 8   0a 61 03
 * 9   0a c1 03
 *
 *
 *
 * */

func Test_Boost_Varint_Serdes(t *testing.T) {
	buf := make([]byte, 8, 8)

	bytes_required := Encode_Boost_Varint(buf, 24)
	if bytes_required != 1 || buf[0] != 0x60 {
		t.Errorf("Single bytes boost varint  encode test failed %d", buf[0])
	}

	// decode and test
	value, bytes_required := Decode_Boost_Varint(buf)

	if value != 24 || bytes_required != 1 {
		t.Errorf("Single bytes boost varint  decode test failed  value %d bytes %d", value, bytes_required)
	}

	// 2 bytes test
	bytes_required = Encode_Boost_Varint(buf, 72)
	if bytes_required != 2 || buf[0] != 0x21 || buf[1] != 1 {
		t.Errorf("2 bytes boost varint  encode test failed")
	}

	// decode and test
	value, bytes_required = Decode_Boost_Varint(buf)

	if value != 72 || bytes_required != 2 {
		t.Errorf("2 bytes boost varint  decode test failed")
	}

	bytes_required = Encode_Boost_Varint(buf, 6000)
	if bytes_required != 2 || buf[0] != 0xc1 || buf[1] != 0x5d {
		t.Errorf("2 bytes boost varint  encode test failed")
	}

	// decode and test
	value, bytes_required = Decode_Boost_Varint(buf)

	if value != 6000 || bytes_required != 2 {
		t.Errorf("2 bytes boost varint  decode test failed")
	}

	// 3 bytes test

	bytes_required = Encode_Boost_Varint(buf, 40096)
	if bytes_required != 4 || buf[0] != 0x82 || buf[1] == 0x72 && buf[2] != 0x2 && buf[3] == 0 {
		t.Errorf("3 bytes boost varint  encode test failed")
	}

	// decode and test
	value, bytes_required = Decode_Boost_Varint(buf)
	if value != 40096 || bytes_required != 4 {
		t.Errorf("3 bytes boost varint  decode test failed")
	}

}

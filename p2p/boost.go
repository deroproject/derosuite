package p2p

import "encoding/binary"

// boost varint can encode upto 1 gb size
func Decode_Boost_Varint(buf []byte) (uint64, int) {
	first_byte := buf[0]
	num_bytes := first_byte & 3 // lower 2 bits contain bytes count, which follow current byte

	// grab full 4 bytes

	value := binary.LittleEndian.Uint32(buf)

	switch num_bytes {
	case 0:
		value &= 0xff
	case 1:
		value &= 0xffff
	case 2:
		value &= 0xffffff
	case 3:
		value &= 0xffffffff
	}
	value = value >> 2 // discard lower 2 bits

	if num_bytes == 2 { // holy hell,  let god have mercy on this boost parser
		num_bytes = 3
	}

	return uint64(value), int(num_bytes + 1)
}

// this function encodes Value to bu in boost varint style
func Encode_Boost_Varint(buf []byte, Value uint64) int {

	bytes_required := byte(0)
	switch {
	case Value > 1073741823:
		panic("Exceeded boost varint capacity while encoding\n") // (2^30) -1
	case Value > 4194303:
		bytes_required = 4
	case Value > 16383:
		bytes_required = 3
	case Value > 63:
		bytes_required = 2
	default:
		bytes_required = 1
	}
	first_byte := (Value % 64) << 2
	Value = Value >> 6
	second_byte := Value % 256
	Value = Value >> 8
	third_byte := Value % 256
	Value = Value >> 8
	fourth_byte := Value % 256
	Value = Value >> 8

	// encode bytes length in lower 2 bits of first byte
	first_byte |= uint64(byte(bytes_required - 1))

	buf[0] = byte(first_byte)
	buf[1] = byte(second_byte)
	buf[2] = byte(third_byte)
	buf[3] = byte(fourth_byte)

	if bytes_required == 3 { // thank god we are soon going to migrate from boost hell, once and for all
		bytes_required = 4
	}

	return int(bytes_required)

}

package ringct

import "io"
import "fmt"

// these functions basically what golang varint does, (  however theere is a minor spec diff, so these are here for exact compatibility)

func ReadVarInt(buf io.Reader) (result uint64, err error) {
	b := make([]byte, 1)
	var r uint64
	var n int
	for i := 0; ; i++ {
		n, err = buf.Read(b)
		if err != nil {
			return
		}
		if n != 1 {
			err = fmt.Errorf("Buffer ended prematurely for varint")
			return
		}
		r += (uint64(b[0]) & 0x7f) << uint(i*7)
		if uint64(b[0])&0x80 == 0 {
			break
		}
	}
	result = r
	return
}

func Uint64ToBytes(num uint64) (result []byte) { 
	for ; num >= 0x80; num >>= 7 {
		result = append(result, byte((num&0x7f)|0x80))
	}
	result = append(result, byte(num))
	return
}

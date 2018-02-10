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

package cryptonight

type size_t uint32
type __locale_data struct {
}
type __locale_t struct {
	__locales       [][]__locale_data
	__ctype_b       []uint16
	__ctype_tolower []int
	__ctype_toupper []int
	__names         [][]byte
}
type locale_t struct {
	__locales       [][]__locale_data
	__ctype_b       []uint16
	__ctype_tolower []int
	__ctype_toupper []int
	__names         [][]byte
}
type ptrdiff_t int32
type wchar_t int
type max_align_t struct {
	__clang_max_align_nonce1 int64
	__clang_max_align_nonce2 float64
}

var sbox []uint8 = []uint8{99, 124, 119, 123, 242, 107, 111, 197, 48, 1, 103, 43, 254, 215, 171, 118, 202, 130, 201, 125, 250, 89, 71, 240, 173, 212, 162, 175, 156, 164, 114, 192, 183, 253, 147, 38, 54, 63, 247, 204, 52, 165, 229, 241, 113, 216, 49, 21, 4, 199, 35, 195, 24, 150, 5, 154, 7, 18, 128, 226, 235, 39, 178, 117, 9, 131, 44, 26, 27, 110, 90, 160, 82, 59, 214, 179, 41, 227, 47, 132, 83, 209, 0, 237, 32, 252, 177, 91, 106, 203, 190, 57, 74, 76, 88, 207, 208, 239, 170, 251, 67, 77, 51, 133, 69, 249, 2, 127, 80, 60, 159, 168, 81, 163, 64, 143, 146, 157, 56, 245, 188, 182, 218, 33, 16, 255, 243, 210, 205, 12, 19, 236, 95, 151, 68, 23, 196, 167, 126, 61, 100, 93, 25, 115, 96, 129, 79, 220, 34, 42, 144, 136, 70, 238, 184, 20, 222, 94, 11, 219, 224, 50, 58, 10, 73, 6, 36, 92, 194, 211, 172, 98, 145, 149, 228, 121, 231, 200, 55, 109, 141, 213, 78, 169, 108, 86, 244, 234, 101, 122, 174, 8, 186, 120, 37, 46, 28, 166, 180, 198, 232, 221, 116, 31, 75, 189, 139, 138, 112, 62, 181, 102, 72, 3, 246, 14, 97, 53, 87, 185, 134, 193, 29, 158, 225, 248, 152, 17, 105, 217, 142, 148, 155, 30, 135, 233, 206, 85, 40, 223, 140, 161, 137, 13, 191, 230, 66, 104, 65, 153, 45, 15, 176, 84, 187, 22}
var mul2 []uint8 = []uint8{0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32, 34, 36, 38, 40, 42, 44, 46, 48, 50, 52, 54, 56, 58, 60, 62, 64, 66, 68, 70, 72, 74, 76, 78, 80, 82, 84, 86, 88, 90, 92, 94, 96, 98, 100, 102, 104, 106, 108, 110, 112, 114, 116, 118, 120, 122, 124, 126, 128, 130, 132, 134, 136, 138, 140, 142, 144, 146, 148, 150, 152, 154, 156, 158, 160, 162, 164, 166, 168, 170, 172, 174, 176, 178, 180, 182, 184, 186, 188, 190, 192, 194, 196, 198, 200, 202, 204, 206, 208, 210, 212, 214, 216, 218, 220, 222, 224, 226, 228, 230, 232, 234, 236, 238, 240, 242, 244, 246, 248, 250, 252, 254, 27, 25, 31, 29, 19, 17, 23, 21, 11, 9, 15, 13, 3, 1, 7, 5, 59, 57, 63, 61, 51, 49, 55, 53, 43, 41, 47, 45, 35, 33, 39, 37, 91, 89, 95, 93, 83, 81, 87, 85, 75, 73, 79, 77, 67, 65, 71, 69, 123, 121, 127, 125, 115, 113, 119, 117, 107, 105, 111, 109, 99, 97, 103, 101, 155, 153, 159, 157, 147, 145, 151, 149, 139, 137, 143, 141, 131, 129, 135, 133, 187, 185, 191, 189, 179, 177, 183, 181, 171, 169, 175, 173, 163, 161, 167, 165, 219, 217, 223, 221, 211, 209, 215, 213, 203, 201, 207, 205, 195, 193, 199, 197, 251, 249, 255, 253, 243, 241, 247, 245, 235, 233, 239, 237, 227, 225, 231, 229}

func mix_bytes(i0 uint8, i1 uint8, i2 uint8, i3 uint8, i4 uint8, i5 uint8, i6 uint8, i7 uint8, output []uint8) {
	var t0 uint8
	var t1 uint8
	var t2 uint8
	var t3 uint8
	var t4 uint8
	var t5 uint8
	var t6 uint8
	var t7 uint8
	var x0 uint8
	var x1 uint8
	var x2 uint8
	var x3 uint8
	var x4 uint8
	var x5 uint8
	var x6 uint8
	var x7 uint8
	var y0 uint8
	var y1 uint8
	var y2 uint8
	var y3 uint8
	var y4 uint8
	var y5 uint8
	var y6 uint8
	var y7 uint8
	t0 = i0 ^ i1
	t1 = i1 ^ i2
	t2 = i2 ^ i3
	t3 = i3 ^ i4
	t4 = i4 ^ i5
	t5 = i5 ^ i6
	t6 = i6 ^ i7
	t7 = i7 ^ i0
	x0 = t0 ^ t3
	x1 = t1 ^ t4
	x2 = t2 ^ t5
	x3 = t3 ^ t6
	x4 = t4 ^ t7
	x5 = t5 ^ t0
	x6 = t6 ^ t1
	x7 = t7 ^ t2
	y0 = t0 ^ t2 ^ i6
	y1 = t1 ^ t3 ^ i7
	y2 = t2 ^ t4 ^ i0
	y3 = t3 ^ t5 ^ i1
	y4 = t4 ^ t6 ^ i2
	y5 = t5 ^ t7 ^ i3
	y6 = t6 ^ t0 ^ i4
	y7 = t7 ^ t1 ^ i5
	x3 = func() uint8 {
		if (x3 & 128) != 0 {
			return (x3 << uint64(1)) ^ 27
		} else {
			return (x3 << uint64(1))
		}
	}()
	x0 = func() uint8 {
		if (x0 & 128) != 0 {
			return (x0 << uint64(1)) ^ 27
		} else {
			return (x0 << uint64(1))
		}
	}()
	t0 = x3 ^ y7
	t0 = func() uint8 {
		if (t0 & 128) != 0 {
			return (t0 << uint64(1)) ^ 27
		} else {
			return (t0 << uint64(1))
		}
	}()
	t5 = x0 ^ y4
	t5 = func() uint8 {
		if (t5 & 128) != 0 {
			return (t5 << uint64(1)) ^ 27
		} else {
			return (t5 << uint64(1))
		}
	}()
	output[0] = t0 ^ y4
	output[5] = t5 ^ y1
	output[1] = mul2[mul2[x4]^y0] ^ y5
	output[2] = mul2[mul2[x5]^y1] ^ y6
	output[3] = mul2[mul2[x6]^y2] ^ y7
	output[4] = mul2[mul2[x7]^y3] ^ y0
	output[6] = mul2[mul2[x1]^y5] ^ y2
	output[7] = mul2[mul2[x2]^y6] ^ y3
}

func perm_P(input []uint8, output []uint8) {
	var r0 uint8
	var r1 uint8
	var r2 uint8
	var r3 uint8
	var r4 uint8
	var r5 uint8
	var r6 uint8
	var r7 uint8
	var round uint8
	var state []uint8 = make([]uint8, 64, 64)
	var write []uint8 = state
	var read []uint8 = input
	var p_tmp []uint8
	for {
		break
	}
	{
		for round = uint8(0); round < 10; func() uint8 {
			round += 1
			return round
		}() {
			r0 = sbox[read[0]^round]
			r1 = sbox[read[9]]
			r2 = sbox[read[18]]
			r3 = sbox[read[27]]
			r4 = sbox[read[36]]
			r5 = sbox[read[45]]
			r6 = sbox[read[54]]
			r7 = sbox[read[63]]
			mix_bytes(r0, r1, r2, r3, r4, r5, r6, r7, write)
			r0 = sbox[read[8]^round^16]
			r1 = sbox[read[17]]
			r2 = sbox[read[26]]
			r3 = sbox[read[35]]
			r4 = sbox[read[44]]
			r5 = sbox[read[53]]
			r6 = sbox[read[62]]
			r7 = sbox[read[7]]
			mix_bytes(r0, r1, r2, r3, r4, r5, r6, r7, write[8:])
			r0 = sbox[read[16]^round^32]
			r1 = sbox[read[25]]
			r2 = sbox[read[34]]
			r3 = sbox[read[43]]
			r4 = sbox[read[52]]
			r5 = sbox[read[61]]
			r6 = sbox[read[6]]
			r7 = sbox[read[15]]
			mix_bytes(r0, r1, r2, r3, r4, r5, r6, r7, write[16:])
			r0 = sbox[read[24]^round^48]
			r1 = sbox[read[33]]
			r2 = sbox[read[42]]
			r3 = sbox[read[51]]
			r4 = sbox[read[60]]
			r5 = sbox[read[5]]
			r6 = sbox[read[14]]
			r7 = sbox[read[23]]
			mix_bytes(r0, r1, r2, r3, r4, r5, r6, r7, write[24:])
			r0 = sbox[read[32]^round^64]
			r1 = sbox[read[41]]
			r2 = sbox[read[50]]
			r3 = sbox[read[59]]
			r4 = sbox[read[4]]
			r5 = sbox[read[13]]
			r6 = sbox[read[22]]
			r7 = sbox[read[31]]
			mix_bytes(r0, r1, r2, r3, r4, r5, r6, r7, write[32:])
			r0 = sbox[read[40]^round^80]
			r1 = sbox[read[49]]
			r2 = sbox[read[58]]
			r3 = sbox[read[3]]
			r4 = sbox[read[12]]
			r5 = sbox[read[21]]
			r6 = sbox[read[30]]
			r7 = sbox[read[39]]
			mix_bytes(r0, r1, r2, r3, r4, r5, r6, r7, write[40:])
			r0 = sbox[read[48]^round^96]
			r1 = sbox[read[57]]
			r2 = sbox[read[2]]
			r3 = sbox[read[11]]
			r4 = sbox[read[20]]
			r5 = sbox[read[29]]
			r6 = sbox[read[38]]
			r7 = sbox[read[47]]
			mix_bytes(r0, r1, r2, r3, r4, r5, r6, r7, write[48:])
			r0 = sbox[read[56]^round^112]
			r1 = sbox[read[1]]
			r2 = sbox[read[10]]
			r3 = sbox[read[19]]
			r4 = sbox[read[28]]
			r5 = sbox[read[37]]
			r6 = sbox[read[46]]
			r7 = sbox[read[55]]
			mix_bytes(r0, r1, r2, r3, r4, r5, r6, r7, write[56:])
			if round == uint8(0) {
				read = output
			}
			p_tmp = read
			read = write
			write = p_tmp
		}
	}
}

func perm_Q(input []uint8, output []uint8) {
	var r0 uint8
	var r1 uint8
	var r2 uint8
	var r3 uint8
	var r4 uint8
	var r5 uint8
	var r6 uint8
	var r7 uint8
	var round uint8
	var state []uint8 = make([]uint8, 64, 64)
	var write []uint8 = state
	var read []uint8 = input
	var p_tmp []uint8
	for {
		break
	}
	{
		for round = uint8(0); round < 10; func() uint8 {
			round += 1
			return round
		}() {
			r0 = sbox[read[8]^255]
			r1 = sbox[read[25]^255]
			r2 = sbox[read[42]^255]
			r3 = sbox[read[59]^255]
			r4 = sbox[read[4]^255]
			r5 = sbox[read[21]^255]
			r6 = sbox[read[38]^255]
			r7 = sbox[read[55]^159^round]
			mix_bytes(r0, r1, r2, r3, r4, r5, r6, r7, write)
			r0 = sbox[read[16]^255]
			r1 = sbox[read[33]^255]
			r2 = sbox[read[50]^255]
			r3 = sbox[read[3]^255]
			r4 = sbox[read[12]^255]
			r5 = sbox[read[29]^255]
			r6 = sbox[read[46]^255]
			r7 = sbox[read[63]^143^round]
			mix_bytes(r0, r1, r2, r3, r4, r5, r6, r7, write[8:])
			r0 = sbox[read[24]^255]
			r1 = sbox[read[41]^255]
			r2 = sbox[read[58]^255]
			r3 = sbox[read[11]^255]
			r4 = sbox[read[20]^255]
			r5 = sbox[read[37]^255]
			r6 = sbox[read[54]^255]
			r7 = sbox[read[7]^255^round]
			mix_bytes(r0, r1, r2, r3, r4, r5, r6, r7, write[16:])
			r0 = sbox[read[32]^255]
			r1 = sbox[read[49]^255]
			r2 = sbox[read[2]^255]
			r3 = sbox[read[19]^255]
			r4 = sbox[read[28]^255]
			r5 = sbox[read[45]^255]
			r6 = sbox[read[62]^255]
			r7 = sbox[read[15]^239^round]
			mix_bytes(r0, r1, r2, r3, r4, r5, r6, r7, write[24:])
			r0 = sbox[read[40]^255]
			r1 = sbox[read[57]^255]
			r2 = sbox[read[10]^255]
			r3 = sbox[read[27]^255]
			r4 = sbox[read[36]^255]
			r5 = sbox[read[53]^255]
			r6 = sbox[read[6]^255]
			r7 = sbox[read[23]^223^round]
			mix_bytes(r0, r1, r2, r3, r4, r5, r6, r7, write[32:])
			r0 = sbox[read[48]^255]
			r1 = sbox[read[1]^255]
			r2 = sbox[read[18]^255]
			r3 = sbox[read[35]^255]
			r4 = sbox[read[44]^255]
			r5 = sbox[read[61]^255]
			r6 = sbox[read[14]^255]
			r7 = sbox[read[31]^207^round]
			mix_bytes(r0, r1, r2, r3, r4, r5, r6, r7, write[40:])
			r0 = sbox[read[56]^255]
			r1 = sbox[read[9]^255]
			r2 = sbox[read[26]^255]
			r3 = sbox[read[43]^255]
			r4 = sbox[read[52]^255]
			r5 = sbox[read[5]^255]
			r6 = sbox[read[22]^255]
			r7 = sbox[read[39]^191^round]
			mix_bytes(r0, r1, r2, r3, r4, r5, r6, r7, write[48:])
			r0 = sbox[read[0]^255]
			r1 = sbox[read[17]^255]
			r2 = sbox[read[34]^255]
			r3 = sbox[read[51]^255]
			r4 = sbox[read[60]^255]
			r5 = sbox[read[13]^255]
			r6 = sbox[read[30]^255]
			r7 = sbox[read[47]^175^round]
			mix_bytes(r0, r1, r2, r3, r4, r5, r6, r7, write[56:])
			if round == uint8(0) {
				read = output
			}
			p_tmp = read
			read = write
			write = p_tmp
		}
	}
}

func crypto_hash(out []uint8, in []uint8, inlen uint64) int {
	if inlen >= (1 << uint64(16)) {
		return -1
	}
	var msg_len uint32 = uint32(uint32(inlen))
	var padded_len uint32 = ((msg_len+9-1)/64)*64 + 64
	var pad_block_len uint8 = uint8(padded_len - msg_len)
	var pad_block []uint8 = make([]uint8, pad_block_len, pad_block_len)

	pad_block[0] = uint8(128)
	var blocks uint32 = uint32((padded_len >> uint64(6)))
	pad_block[pad_block_len-1] = (uint8(blocks) & 255)
	var h_state []uint8 = make([]uint8, 64, 64)
	var p_state []uint8 = make([]uint8, 64, 64)
	var q_state []uint8 = make([]uint8, 64, 64)
	var x_state []uint8 = make([]uint8, 64, 64)
	var buf []uint8 = make([]uint8, 64, 64)

	h_state[62] = uint8(1)
	var i uint8
	var block uint32
	var remaining uint32 = uint32(0)
	var message_left_len uint32 = msg_len
	for {
		break
	}
	{
		for block = uint32(0); block < blocks; func() uint32 {
			block += 1
			return block
		}() {
			if block*64+64 < msg_len {
				//memcpy(buf, in+64*block, uint32(64))
				copy(buf, in[64*block:64*block+64]) //copy full block
				message_left_len -= 64
			} else {
				if message_left_len > 0 {
					remaining = uint32(64 - message_left_len)
					//memcpy(buf, in+64*block, message_left_len)
					copy(buf, in[64*block:64*block+message_left_len])
					//memcpy(buf+message_left_len, pad_block, uint32(remaining))
					copy(buf[message_left_len:], pad_block[:remaining])
					message_left_len = uint32(0)
				} else {
					//memcpy(buf, pad_block+remaining, uint32(64))
					copy(buf, pad_block[remaining:remaining+64])
				}
			}
			for {
				break
			}
			{
				for i = uint8(0); i < 64; func() uint8 {
					i += 1
					return i
				}() {
					x_state[i] = buf[i] ^ h_state[i]
				}
			}
			perm_P(x_state, p_state)
			perm_Q(buf, q_state)
			for {
				break
			}
			{
				for i = uint8(0); i < 64; func() uint8 {
					i += 1
					return i
				}() {
					h_state[i] ^= p_state[i] ^ q_state[i]
				}
			}
		}
	}
	perm_P(h_state, p_state)
	for {
		break
	}
	{
		for i = uint8(32); i < 64; func() uint8 {
			i += 1
			return i
		}() {
			out[i-32] = h_state[i] ^ p_state[i]
		}
	}
	return 0
}
func __init() {
}

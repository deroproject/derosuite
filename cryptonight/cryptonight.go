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

//import "fmt"
import "unsafe"

//import "encoding/hex"
import "encoding/binary"
import "github.com/aead/skein"
import "github.com/dchest/blake256"

var HardwareAES bool = false

const MAX_ARRAY_LIMIT = (4 * 1024 * 1024)

func cryptonight(input []byte) []byte {

	var dummy [256]byte
	var S [25]uint64

	var key1 [64]uint32
	var key2 [64]uint32

	var a [2]uint64
	var b [2]uint64
	var c [2]uint64

	a_uint32 := (*(*[4]uint32)(unsafe.Pointer(&a[0])))[:len(a)*2]
	//b_uint32 := (*(*[4]uint32)(unsafe.Pointer(&b[0])))[:len(b)*2]
	c_uint32 := (*(*[4]uint32)(unsafe.Pointer(&c[0])))[:len(c)*2]

	// same array is accessed as 3 different ways, buts it bettter than copying
	var ScratchPad = make([]uint64, 1<<18, 1<<18)
	ScratchPad_uint32 := (*(*[MAX_ARRAY_LIMIT]uint32)(unsafe.Pointer(&ScratchPad[0])))[:len(ScratchPad)*2]
	// ScratchPad_byte := (*(*[MAX_ARRAY_LIMIT]byte)(unsafe.Pointer(&ScratchPad[0])))[:len(ScratchPad)*8]

	copy(dummy[:], input)

	for i := 0; i < 16; i++ {
		S[i] = binary.LittleEndian.Uint64(dummy[i<<3:])
	}
	S[16] = 0x8000000000000000

	keccakf(&S)

	// lets convert everything back to bytes
	for i := 0; i < 25; i++ {
		binary.LittleEndian.PutUint64(dummy[i<<3:], S[i])
	}

	// extract keys
	/*for i := 0 ; i <8;i++{
	    key1[i] = binary.LittleEndian.Uint32(dummy[i<<2:])
	    key2[i] = binary.LittleEndian.Uint32(dummy[32+(i<<2):])
	}*/

	expandKeyGo(dummy[0:32], key1[:], nil)
	expandKeyGo(dummy[32:64], key2[:], nil)

	/* for i :=0; i< 60;i++{
	         fmt.Printf("%2d %X\n", i, key2[i])
	}*/

	// var text [128]byte
	var text_uint32 [32]uint32
	//copy(text[:],dummy[64:64+128]) // copy 128 bytes

	for i := 0; i < 32; i++ {
		text_uint32[i] = binary.LittleEndian.Uint32(dummy[64+(i<<2):])
	}

	/* for i :=0; i< 32;i++{
	         fmt.Printf("%2d %X  i  %08X %08X\n", i, text_uint32[i] , key1[i],key1[i<<1])
	}*/

	for i := 0; i < 0x4000; i++ {
		for j := 0; j < 8; j++ {
			CNAESTransform((text_uint32[(j << 2):]), key1[:])
		}

		//memcpy(CNCtx.Scratchpad + (i << 4), text, 128);
		copy(ScratchPad_uint32[i*32:], text_uint32[:])
	}

	a[0] = S[0] ^ S[4]
	a[1] = S[1] ^ S[5]
	b[0] = S[2] ^ S[6]
	b[1] = S[3] ^ S[7]

	var x [2]uint64
	_ = x

	// the big slow
	for i := 0; i < 0x80000; i++ {
		// {
		c[0] = ScratchPad[((a[0]&0x1FFFF0)>>3)+0]
		c[1] = ScratchPad[((a[0]&0x1FFFF0)>>3)+1]

		CNAESRnd(c_uint32, a_uint32)

		b[0] ^= c[0]
		b[1] ^= c[1]

		ScratchPad[((a[0]&0x1FFFF0)>>3)+0] = b[0]
		ScratchPad[((a[0]&0x1FFFF0)>>3)+1] = b[1]

		b[0] = ScratchPad[((c[0]&0x1FFFF0)>>3)+0]
		b[1] = ScratchPad[((c[0]&0x1FFFF0)>>3)+1]

		/*  // faster assembly implmentation for mul on amd64
		x[0]= c[0]
		x[1]= b[0]
		mul6464128(&x[0])
		a[1] += x[1]
		a[0] += x[0]
		*/

		// time to do 64 bit * 64 bit multiply
		// this is faster than the assembly
		var lower64, upper64 uint64
		{

			x := c[0]
			y := b[0]

			a := x >> 32
			b := x & 0xffffffff
			c := y >> 32
			d := y & 0xffffffff

			ac := a * c
			bc := b * c
			ad := a * d
			bd := b * d

			mid34 := (bd >> 32) + (bc & 0xffffffff) + (ad & 0xffffffff)

			upper64 = ac + (bc >> 32) + (ad >> 32) + (mid34 >> 32)
			lower64 = (mid34 << 32) | (bd & 0xffffffff)
			_ = lower64
			_ = upper64

		}
		a[1] += lower64
		a[0] += upper64

		ScratchPad[((c[0]&0x1FFFF0)>>3)+0] = a[0]
		ScratchPad[((c[0]&0x1FFFF0)>>3)+1] = a[1]

		a[0] ^= b[0]
		a[1] ^= b[1]

		b[0] = c[0]
		b[1] = c[1]

	}

	// fmt.Printf(" a %X %X\n", a[0],a[1]);
	// fmt.Printf(" b %X %X\n", b[0],b[1]);

	for i := 0; i < 32; i++ {
		text_uint32[i] = binary.LittleEndian.Uint32(dummy[64+(i<<2):])
	}

	for i := 0; i < 0x4000; i++ {

		for j := 0; j < 32; j++ {
			text_uint32[j] ^= ScratchPad_uint32[(i*32)+j]

		}
		for j := 0; j < 8; j++ {
			CNAESTransform((text_uint32[(j << 2):]), key2[:])
		}

	}

	/*for i :=0; i< 32;i++{
	         fmt.Printf("%2d %X\n", i, text_uint32[i])
	}*/

	for i := 0; i < 32; i++ {
		binary.LittleEndian.PutUint32(dummy[64+(i<<2):], text_uint32[i])
	}

	for i := 8; i < 25; i++ {
		S[i] = binary.LittleEndian.Uint64(dummy[i<<3:])
	}

	keccakf(&S) // do the keccak round
	/* for i :=0; i< 25;i++{
	         fmt.Printf("S %02d %X\n", i, S[i])
	}*/

	// lets convert everything back to bytes
	for i := 0; i < 25; i++ {
		binary.LittleEndian.PutUint64(dummy[i<<3:], S[i])
	}

	var resulthash []byte

	switch S[0] & 3 {

	case 0:
		// fmt.Printf("blake\n")

		// blake
		blakehash := blake256.New()
		blakehash.Write(dummy[:200])
		resulthash = blakehash.Sum(nil) //matching

	case 1: // groestl
		// fmt.Printf("groestl not implemented\n")
		var output [32]byte
		var input = dummy[:200]
		crypto_hash(output[:], input, uint64(len(input)))
		resulthash = output[:]
	case 2: // jh
		//fmt.Printf("jh not implemented\n")

		myjhash := NewJhash256()
		myjhash.Write(dummy[:200])
		resulthash = myjhash.Sum(nil)

	case 3: // skein
		//  fmt.Printf("skein\n")
		skeinhash := skein.New256(nil)
		skeinhash.Write(dummy[:200])
		resulthash = skeinhash.Sum(nil) //matchin

	}

	//fmt.Printf("result hash %x\n", resulthash)

	return resulthash

}

/* from original cryptonote source
0100000000000000000000000000000000000000000000000000000000000000000000102700005a18d9489bcd353aeaf4a19323d04e90353f98f0d7cc2a030cfd76e19495547d01
a73bd37aba3454776b40733854a8349fe6359eb2c91d93bc727c69431c1d1f95hash of blob

// from our implementation

Get long hash a73bd37aba3454776b40733854a8349fe6359eb2c91d93bc727c69431c1d1f95
*/

func SlowHash(msg []byte) []byte {
	return cryptonight(append(msg, byte(0x01)))
}

// Rotate

func ROTL32(x uint32, y uint32) uint32 { return (((x) << (y)) | ((x) >> (32 - (y)))) }
func BYTE(x, y uint32) uint32          { return (((x) >> ((y) << 3)) & 0xFF) }

func SubWord(inw uint32) uint32 {
	return ((CNAESSbox[BYTE(inw, 3)] << 24) | (CNAESSbox[BYTE(inw, 2)] << 16) | (CNAESSbox[BYTE(inw, 1)] << 8) | CNAESSbox[BYTE(inw, 0)])
}

/*
func AESExpandKey256(keybuf []uint32){
    i := uint32(1)
    t := uint32(0)
	for c := 8 ; c < 60; c++	{
		// For 256-bit keys, an sbox permutation is done every other 4th uint generated, AND every 8th
		// t := ((!(c & 3))) ? SubWord(keybuf[c - 1]) : keybuf[c - 1];
              if  (c & 3) == 0 {
                  t = SubWord(keybuf[c - 1])
              }else{
                  t = keybuf[c - 1];
              }

		// If the uint we're generating has an index that is a multiple of 8, rotate and XOR with the round constant,
		// then XOR this with previously generated uint. If it's 4 after a multiple of 8, only the sbox permutation
		// is done, followed by the XOR. If neither are true, only the XOR with the previously generated uint is done.
		//keybuf[c] = keybuf[c - 8] ^ ((!(c & 7)) ? ROTL32(t, 24U) ^ ((uint32_t)(CNAESRcon[i++])) : t);
		if (keybuf[c - 8] ^ ((!(c & 7))) > 0) {
                    keybuf[c] = ROTL32(t, 24) ^ ((uint32)(CNAESRcon[i]))
                    i++
                }else{
                    keybuf[c] = t;
                }
	}
}
*/

func CNAESTransform(X, Key []uint32) {

	//  fmt.Printf("X %08X %08X  \n", X[0],X[1]) ;

	/*
		for i := uint32(0); i < 10; i++ {
			CNAESRnd(X, Key[(i<<2):])
		}*/
	if HardwareAES {
		/*encryptAESRound(&Key[0<<2],&X[0],&X[0])
		encryptAESRound(&Key[1<<2],&X[0],&X[0])
		encryptAESRound(&Key[2<<2],&X[0],&X[0])
		encryptAESRound(&Key[3<<2],&X[0],&X[0])
		encryptAESRound(&Key[4<<2],&X[0],&X[0])
		encryptAESRound(&Key[5<<2],&X[0],&X[0])
		encryptAESRound(&Key[6<<2],&X[0],&X[0])
		encryptAESRound(&Key[7<<2],&X[0],&X[0])
		encryptAESRound(&Key[8<<2],&X[0],&X[0])
		encryptAESRound(&Key[9<<2],&X[0],&X[0])
		*/
		encrypt10AESRound(&Key[0<<2], &X[0], &X[0])
	} else {
		for i := uint32(0); i < 10; i++ {
			CNAESRnd(X, Key[(i<<2):])
		}

	}

}

func CNAESRnd(X, key []uint32) {

	if HardwareAES {
		encryptAESRound(&key[0], &X[0], &X[0])
		return
	}

	// use software implementation, of no AES detected or CPU not X86
	var Y [4]uint32

	Y[0] = CNAESTbl[BYTE(X[0], 0)] ^ ROTL32(CNAESTbl[BYTE(X[1], 1)], 8) ^ ROTL32(CNAESTbl[BYTE(X[2], 2)], 16) ^ ROTL32(CNAESTbl[BYTE(X[3], 3)], 24)
	Y[1] = CNAESTbl[BYTE(X[1], 0)] ^ ROTL32(CNAESTbl[BYTE(X[2], 1)], 8) ^ ROTL32(CNAESTbl[BYTE(X[3], 2)], 16) ^ ROTL32(CNAESTbl[BYTE(X[0], 3)], 24)
	Y[2] = CNAESTbl[BYTE(X[2], 0)] ^ ROTL32(CNAESTbl[BYTE(X[3], 1)], 8) ^ ROTL32(CNAESTbl[BYTE(X[0], 2)], 16) ^ ROTL32(CNAESTbl[BYTE(X[1], 3)], 24)
	Y[3] = CNAESTbl[BYTE(X[3], 0)] ^ ROTL32(CNAESTbl[BYTE(X[0], 1)], 8) ^ ROTL32(CNAESTbl[BYTE(X[1], 2)], 16) ^ ROTL32(CNAESTbl[BYTE(X[2], 3)], 24)

	for i := 0; i < 4; i++ {
		X[i] = Y[i] ^ key[i]
	}

}

// copied from https://golang.org/src/crypto/aes/block.go

// Apply sbox0 to each byte in w.

func subw(w uint32) uint32 {
	return uint32(sbox0[w>>24])<<24 |
		uint32(sbox0[w>>16&0xff])<<16 |
		uint32(sbox0[w>>8&0xff])<<8 |
		uint32(sbox0[w&0xff])
}

// Rotate
func rotw(w uint32) uint32 { return w<<8 | w>>24 }

func swap_uint32(val uint32) uint32 {
	val = ((val << 8) & 0xFF00FF00) | ((val >> 8) & 0xFF00FF)
	return (val << 16) | (val >> 16)
}

// Key expansion algorithm. See FIPS-197, Figure 11.
// Their rcon[i] is our powx[i-1] << 24.

func expandKeyGo(key []byte, enc, dec []uint32) {
	// Encryption key setup.
	var i int
	nk := len(key) / 4
	for i = 0; i < nk; i++ {
		enc[i] = uint32(key[4*i])<<24 | uint32(key[4*i+1])<<16 | uint32(key[4*i+2])<<8 | uint32(key[4*i+3])
	}

	for ; i < len(enc); i++ {
		t := enc[i-1]
		if i%nk == 0 {
			t = subw(rotw(t)) ^ (uint32(powx[i/nk-1]) << 24)
		} else if nk > 6 && i%nk == 4 {
			t = subw(t)
		}
		enc[i] = enc[i-nk] ^ t

		// fmt.Printf("%2d %X\n", i, enc[i])
	}

	// key generated by golang need to be swapped
	for i := 0; i < 60; i++ {
		enc[i] = swap_uint32(enc[i])
		//fmt.Printf("%2d %X\n", i, enc[i])
	}

	/*
	   	// Derive decryption key from encryption key.
	   	// Reverse the 4-word round key sets from enc to produce dec.
	   	// All sets but the first and last get the MixColumn transform applied.
	   	if dec == nil {
	   		return
	   	}
	   	n := len(enc)
	   	for i := 0; i < n; i += 4 {
	   		ei := n - i - 4
	               for j := 0; j < 4; j++ {
	   			x := enc[ei+j]
	   			if i > 0 && i+4 < n {
	   			x = td0[sbox0[x>>24]] ^ td1[sbox0[x>>16&0xff]] ^ td2[sbox0[x>>8&0xff]] ^ td3[sbox0[x&0xff]]
	   		}
	   		dec[i+j] = x
	   		}
	   	}

	*/
}

// copied from https://golang.org/src/crypto/aes/const.go
// Powers of x mod poly in GF(2).
var powx = [16]byte{
	0x01,
	0x02,
	0x04,
	0x08,
	0x10,
	0x20,
	0x40,
	0x80,
	0x1b,
	0x36,
	0x6c,
	0xd8,
	0xab,
	0x4d,
	0x9a,
	0x2f,
}

// FIPS-197 Figure 7. S-box substitution values in hexadecimal format.
var sbox0 = [256]byte{
	0x63, 0x7c, 0x77, 0x7b, 0xf2, 0x6b, 0x6f, 0xc5, 0x30, 0x01, 0x67, 0x2b, 0xfe, 0xd7, 0xab, 0x76,
	0xca, 0x82, 0xc9, 0x7d, 0xfa, 0x59, 0x47, 0xf0, 0xad, 0xd4, 0xa2, 0xaf, 0x9c, 0xa4, 0x72, 0xc0,
	0xb7, 0xfd, 0x93, 0x26, 0x36, 0x3f, 0xf7, 0xcc, 0x34, 0xa5, 0xe5, 0xf1, 0x71, 0xd8, 0x31, 0x15,
	0x04, 0xc7, 0x23, 0xc3, 0x18, 0x96, 0x05, 0x9a, 0x07, 0x12, 0x80, 0xe2, 0xeb, 0x27, 0xb2, 0x75,
	0x09, 0x83, 0x2c, 0x1a, 0x1b, 0x6e, 0x5a, 0xa0, 0x52, 0x3b, 0xd6, 0xb3, 0x29, 0xe3, 0x2f, 0x84,
	0x53, 0xd1, 0x00, 0xed, 0x20, 0xfc, 0xb1, 0x5b, 0x6a, 0xcb, 0xbe, 0x39, 0x4a, 0x4c, 0x58, 0xcf,
	0xd0, 0xef, 0xaa, 0xfb, 0x43, 0x4d, 0x33, 0x85, 0x45, 0xf9, 0x02, 0x7f, 0x50, 0x3c, 0x9f, 0xa8,
	0x51, 0xa3, 0x40, 0x8f, 0x92, 0x9d, 0x38, 0xf5, 0xbc, 0xb6, 0xda, 0x21, 0x10, 0xff, 0xf3, 0xd2,
	0xcd, 0x0c, 0x13, 0xec, 0x5f, 0x97, 0x44, 0x17, 0xc4, 0xa7, 0x7e, 0x3d, 0x64, 0x5d, 0x19, 0x73,
	0x60, 0x81, 0x4f, 0xdc, 0x22, 0x2a, 0x90, 0x88, 0x46, 0xee, 0xb8, 0x14, 0xde, 0x5e, 0x0b, 0xdb,
	0xe0, 0x32, 0x3a, 0x0a, 0x49, 0x06, 0x24, 0x5c, 0xc2, 0xd3, 0xac, 0x62, 0x91, 0x95, 0xe4, 0x79,
	0xe7, 0xc8, 0x37, 0x6d, 0x8d, 0xd5, 0x4e, 0xa9, 0x6c, 0x56, 0xf4, 0xea, 0x65, 0x7a, 0xae, 0x08,
	0xba, 0x78, 0x25, 0x2e, 0x1c, 0xa6, 0xb4, 0xc6, 0xe8, 0xdd, 0x74, 0x1f, 0x4b, 0xbd, 0x8b, 0x8a,
	0x70, 0x3e, 0xb5, 0x66, 0x48, 0x03, 0xf6, 0x0e, 0x61, 0x35, 0x57, 0xb9, 0x86, 0xc1, 0x1d, 0x9e,
	0xe1, 0xf8, 0x98, 0x11, 0x69, 0xd9, 0x8e, 0x94, 0x9b, 0x1e, 0x87, 0xe9, 0xce, 0x55, 0x28, 0xdf,
	0x8c, 0xa1, 0x89, 0x0d, 0xbf, 0xe6, 0x42, 0x68, 0x41, 0x99, 0x2d, 0x0f, 0xb0, 0x54, 0xbb, 0x16,
}

var CNAESSbox = [256]uint32{
	0x63, 0x7C, 0x77, 0x7B, 0xF2, 0x6B, 0x6F, 0xC5, 0x30, 0x01, 0x67, 0x2B, 0xFE, 0xD7, 0xAB, 0x76,
	0xCA, 0x82, 0xC9, 0x7D, 0xFA, 0x59, 0x47, 0xF0, 0xAD, 0xD4, 0xA2, 0xAF, 0x9C, 0xA4, 0x72, 0xC0,
	0xB7, 0xFD, 0x93, 0x26, 0x36, 0x3F, 0xF7, 0xCC, 0x34, 0xA5, 0xE5, 0xF1, 0x71, 0xD8, 0x31, 0x15,
	0x04, 0xC7, 0x23, 0xC3, 0x18, 0x96, 0x05, 0x9A, 0x07, 0x12, 0x80, 0xE2, 0xEB, 0x27, 0xB2, 0x75,
	0x09, 0x83, 0x2C, 0x1A, 0x1B, 0x6E, 0x5A, 0xA0, 0x52, 0x3B, 0xD6, 0xB3, 0x29, 0xE3, 0x2F, 0x84,
	0x53, 0xD1, 0x00, 0xED, 0x20, 0xFC, 0xB1, 0x5B, 0x6A, 0xCB, 0xBE, 0x39, 0x4A, 0x4C, 0x58, 0xCF,
	0xD0, 0xEF, 0xAA, 0xFB, 0x43, 0x4D, 0x33, 0x85, 0x45, 0xF9, 0x02, 0x7F, 0x50, 0x3C, 0x9F, 0xA8,
	0x51, 0xA3, 0x40, 0x8F, 0x92, 0x9D, 0x38, 0xF5, 0xBC, 0xB6, 0xDA, 0x21, 0x10, 0xFF, 0xF3, 0xD2,
	0xCD, 0x0C, 0x13, 0xEC, 0x5F, 0x97, 0x44, 0x17, 0xC4, 0xA7, 0x7E, 0x3D, 0x64, 0x5D, 0x19, 0x73,
	0x60, 0x81, 0x4F, 0xDC, 0x22, 0x2A, 0x90, 0x88, 0x46, 0xEE, 0xB8, 0x14, 0xDE, 0x5E, 0x0B, 0xDB,
	0xE0, 0x32, 0x3A, 0x0A, 0x49, 0x06, 0x24, 0x5C, 0xC2, 0xD3, 0xAC, 0x62, 0x91, 0x95, 0xE4, 0x79,
	0xE7, 0xC8, 0x37, 0x6D, 0x8D, 0xD5, 0x4E, 0xA9, 0x6C, 0x56, 0xF4, 0xEA, 0x65, 0x7A, 0xAE, 0x08,
	0xBA, 0x78, 0x25, 0x2E, 0x1C, 0xA6, 0xB4, 0xC6, 0xE8, 0xDD, 0x74, 0x1F, 0x4B, 0xBD, 0x8B, 0x8A,
	0x70, 0x3E, 0xB5, 0x66, 0x48, 0x03, 0xF6, 0x0E, 0x61, 0x35, 0x57, 0xB9, 0x86, 0xC1, 0x1D, 0x9E,
	0xE1, 0xF8, 0x98, 0x11, 0x69, 0xD9, 0x8E, 0x94, 0x9B, 0x1E, 0x87, 0xE9, 0xCE, 0x55, 0x28, 0xDF,
	0x8C, 0xA1, 0x89, 0x0D, 0xBF, 0xE6, 0x42, 0x68, 0x41, 0x99, 0x2D, 0x0F, 0xB0, 0x54, 0xBB, 0x16,
}
var CNAESRcon = [8]uint32{0x8d, 0x01, 0x02, 0x04, 0x08, 0x10, 0x20, 0x40}
var CNAESTbl = [256]uint32{0xA56363C6, 0x847C7CF8, 0x997777EE, 0x8D7B7BF6,
	0x0DF2F2FF, 0xBD6B6BD6, 0xB16F6FDE, 0x54C5C591,
	0x50303060, 0x03010102, 0xA96767CE, 0x7D2B2B56,
	0x19FEFEE7, 0x62D7D7B5, 0xE6ABAB4D, 0x9A7676EC,
	0x45CACA8F, 0x9D82821F, 0x40C9C989, 0x877D7DFA,
	0x15FAFAEF, 0xEB5959B2, 0xC947478E, 0x0BF0F0FB,
	0xECADAD41, 0x67D4D4B3, 0xFDA2A25F, 0xEAAFAF45,
	0xBF9C9C23, 0xF7A4A453, 0x967272E4, 0x5BC0C09B,
	0xC2B7B775, 0x1CFDFDE1, 0xAE93933D, 0x6A26264C,
	0x5A36366C, 0x413F3F7E, 0x02F7F7F5, 0x4FCCCC83,
	0x5C343468, 0xF4A5A551, 0x34E5E5D1, 0x08F1F1F9,
	0x937171E2, 0x73D8D8AB, 0x53313162, 0x3F15152A,
	0x0C040408, 0x52C7C795, 0x65232346, 0x5EC3C39D,
	0x28181830, 0xA1969637, 0x0F05050A, 0xB59A9A2F,
	0x0907070E, 0x36121224, 0x9B80801B, 0x3DE2E2DF,
	0x26EBEBCD, 0x6927274E, 0xCDB2B27F, 0x9F7575EA,
	0x1B090912, 0x9E83831D, 0x742C2C58, 0x2E1A1A34,
	0x2D1B1B36, 0xB26E6EDC, 0xEE5A5AB4, 0xFBA0A05B,
	0xF65252A4, 0x4D3B3B76, 0x61D6D6B7, 0xCEB3B37D,
	0x7B292952, 0x3EE3E3DD, 0x712F2F5E, 0x97848413,
	0xF55353A6, 0x68D1D1B9, 0x00000000, 0x2CEDEDC1,
	0x60202040, 0x1FFCFCE3, 0xC8B1B179, 0xED5B5BB6,
	0xBE6A6AD4, 0x46CBCB8D, 0xD9BEBE67, 0x4B393972,
	0xDE4A4A94, 0xD44C4C98, 0xE85858B0, 0x4ACFCF85,
	0x6BD0D0BB, 0x2AEFEFC5, 0xE5AAAA4F, 0x16FBFBED,
	0xC5434386, 0xD74D4D9A, 0x55333366, 0x94858511,
	0xCF45458A, 0x10F9F9E9, 0x06020204, 0x817F7FFE,
	0xF05050A0, 0x443C3C78, 0xBA9F9F25, 0xE3A8A84B,
	0xF35151A2, 0xFEA3A35D, 0xC0404080, 0x8A8F8F05,
	0xAD92923F, 0xBC9D9D21, 0x48383870, 0x04F5F5F1,
	0xDFBCBC63, 0xC1B6B677, 0x75DADAAF, 0x63212142,
	0x30101020, 0x1AFFFFE5, 0x0EF3F3FD, 0x6DD2D2BF,
	0x4CCDCD81, 0x140C0C18, 0x35131326, 0x2FECECC3,
	0xE15F5FBE, 0xA2979735, 0xCC444488, 0x3917172E,
	0x57C4C493, 0xF2A7A755, 0x827E7EFC, 0x473D3D7A,
	0xAC6464C8, 0xE75D5DBA, 0x2B191932, 0x957373E6,
	0xA06060C0, 0x98818119, 0xD14F4F9E, 0x7FDCDCA3,
	0x66222244, 0x7E2A2A54, 0xAB90903B, 0x8388880B,
	0xCA46468C, 0x29EEEEC7, 0xD3B8B86B, 0x3C141428,
	0x79DEDEA7, 0xE25E5EBC, 0x1D0B0B16, 0x76DBDBAD,
	0x3BE0E0DB, 0x56323264, 0x4E3A3A74, 0x1E0A0A14,
	0xDB494992, 0x0A06060C, 0x6C242448, 0xE45C5CB8,
	0x5DC2C29F, 0x6ED3D3BD, 0xEFACAC43, 0xA66262C4,
	0xA8919139, 0xA4959531, 0x37E4E4D3, 0x8B7979F2,
	0x32E7E7D5, 0x43C8C88B, 0x5937376E, 0xB76D6DDA,
	0x8C8D8D01, 0x64D5D5B1, 0xD24E4E9C, 0xE0A9A949,
	0xB46C6CD8, 0xFA5656AC, 0x07F4F4F3, 0x25EAEACF,
	0xAF6565CA, 0x8E7A7AF4, 0xE9AEAE47, 0x18080810,
	0xD5BABA6F, 0x887878F0, 0x6F25254A, 0x722E2E5C,
	0x241C1C38, 0xF1A6A657, 0xC7B4B473, 0x51C6C697,
	0x23E8E8CB, 0x7CDDDDA1, 0x9C7474E8, 0x211F1F3E,
	0xDD4B4B96, 0xDCBDBD61, 0x868B8B0D, 0x858A8A0F,
	0x907070E0, 0x423E3E7C, 0xC4B5B571, 0xAA6666CC,
	0xD8484890, 0x05030306, 0x01F6F6F7, 0x120E0E1C,
	0xA36161C2, 0x5F35356A, 0xF95757AE, 0xD0B9B969,
	0x91868617, 0x58C1C199, 0x271D1D3A, 0xB99E9E27,
	0x38E1E1D9, 0x13F8F8EB, 0xB398982B, 0x33111122,
	0xBB6969D2, 0x70D9D9A9, 0x898E8E07, 0xA7949433,
	0xB69B9B2D, 0x221E1E3C, 0x92878715, 0x20E9E9C9,
	0x49CECE87, 0xFF5555AA, 0x78282850, 0x7ADFDFA5,
	0x8F8C8C03, 0xF8A1A159, 0x80898909, 0x170D0D1A,
	0xDABFBF65, 0x31E6E6D7, 0xC6424284, 0xB86868D0,
	0xC3414182, 0xB0999929, 0x772D2D5A, 0x110F0F1E,
	0xCBB0B07B, 0xFC5454A8, 0xD6BBBB6D, 0x3A16162C,
}

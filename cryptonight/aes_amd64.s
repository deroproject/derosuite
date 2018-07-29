#include "textflag.h"

// single round AES encryption
// func encryptAESRound( xk *uint32, dst, src *uint32)
TEXT ·encryptAESRound(SB),NOSPLIT,$0
	MOVQ xk+0(FP), AX
	MOVQ dst+8(FP), DX
	MOVQ src+16(FP), BX
	MOVUPS 0(AX), X1
	MOVUPS 0(BX), X0
	AESENC X1, X0
	MOVUPS X0, 0(DX)
	RET

// func encrypt10AESRound( xk *uint32, dst, src *uint32)
TEXT ·encrypt10AESRound(SB),NOSPLIT,$0
	MOVQ xk+0(FP), AX
	MOVQ dst+8(FP), DX
	MOVQ src+16(FP), BX
	MOVUPS 0(AX), X1
	MOVUPS 0(BX), X0
	AESENC X1, X0
	MOVUPS 16(AX), X1
	AESENC X1, X0
	MOVUPS 32(AX), X1
	AESENC X1, X0
	MOVUPS 48(AX), X1
	AESENC X1, X0
	MOVUPS 64(AX), X1
	AESENC X1, X0
	MOVUPS 80(AX), X1
	AESENC X1, X0
	MOVUPS 96(AX), X1
	AESENC X1, X0
	MOVUPS 112(AX), X1
	AESENC X1, X0
	MOVUPS 128(AX), X1
	AESENC X1, X0
	MOVUPS 144(AX), X1
	AESENC X1, X0

	MOVUPS X0, 0(DX)
	RET

// 128 bit = 64 bit * 64 bit
// func mul6464128( input *uint64)
TEXT ·mul6464128(SB),NOSPLIT,$0
	MOVQ input+0(FP), BX
	MOVQ 0(BX), AX
	MOVQ 8(BX), DX

// MUL RDX   and place result,  RDX/RAX
	BYTE $0x48
	BYTE $0xF7
	BYTE $0xE2     

	MOVQ DX,0(BX)
	MOVQ AX,8(BX)
	RET



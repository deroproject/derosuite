package keccak

import (
	"bytes"
	"testing"
)

func TestSHA3224(t *testing.T) {
	h := NewSHA3224()
	for i := range tstShort {
		h.Reset()
		h.Write(sha3tests[i].msg)
		d := h.Sum(nil)
		if !bytes.Equal(d, sha3tests[i].output224) {
			t.Errorf("testcase SHA3224 %d: expected %x got %x", i, sha3tests[i].output224, d)
		}
	}
}

func TestSHA3256(t *testing.T) {
	h := NewSHA3256()
	for i := range sha3tests {
		h.Reset()
		h.Write(sha3tests[i].msg)
		d := h.Sum(nil)
		if !bytes.Equal(d, sha3tests[i].output256) {
			t.Errorf("testcase SHA3256 %d: expected %x got %x", i, sha3tests[i].output256, d)
		}
	}
}

func TestSHA3384(t *testing.T) {
	h := NewSHA3384()
	for i := range sha3tests {
		h.Reset()
		h.Write(sha3tests[i].msg)
		d := h.Sum(nil)
		if !bytes.Equal(d, sha3tests[i].output384) {
			t.Errorf("testcase SHA3384 %d: expected %x got %x", i, sha3tests[i].output384, d)
		}
	}
}

func TestSHA3512(t *testing.T) {
	h := NewSHA3512()
	for i := range sha3tests {
		h.Reset()
		h.Write(sha3tests[i].msg)
		d := h.Sum(nil)
		if !bytes.Equal(d, sha3tests[i].output512) {
			t.Errorf("testcase SHA3512 %d: expected %x got %x", i, sha3tests[i].output512, d)
		}
	}
}

package keccak

import (
	"bytes"
	"testing"
)

func TestSHAKE128(t *testing.T) {
	for i := range shaketests {
		h := NewSHAKE128(len(shaketests[i].output128))
		h.Write(shaketests[i].msg)
		d := h.Sum(nil)
		if !bytes.Equal(d, shaketests[i].output128) {
			t.Errorf("testcase SHAKE128 %d: expected %x got %x", i, shaketests[i].output128, d)
		}
	}
}

func TestSHAKE256(t *testing.T) {
	for i := range shaketests {
		h := NewSHAKE256(len(shaketests[i].output256))
		h.Write(shaketests[i].msg)
		d := h.Sum(nil)
		if !bytes.Equal(d, shaketests[i].output256) {
			t.Errorf("testcase SHAKE256 %d: expected %x got %x", i, shaketests[i].output256, d)
		}
	}
}

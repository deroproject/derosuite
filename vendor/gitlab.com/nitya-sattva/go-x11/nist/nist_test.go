// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package nist

import "testing"

////////////////

func TestGet(t *testing.T) {
	if ln := len(Get(0)); 0 != ln {
		t.Errorf("Get: expected length: %d, got %d", 0, ln)
	}
	if ln := len(Get(1)); 1 != ln {
		t.Errorf("Get: expected length: %d, got %d", 1, ln)
	}
}

func TestHashIsEqual(t *testing.T) {
	var a = []byte{0x00, 0x01}
	var b = []byte{0x00, 0x01}
	var c = []byte{0x01, 0x02}
	var d = []byte{0x01, 0x02, 0x03}

	if !IsEqual(a, a) {
		t.Errorf("HashIsEqual: expected true, got false")
	}
	if !IsEqual(a, b) {
		t.Errorf("HashIsEqual: expected true, got false")
	}
	if IsEqual(a, c) {
		t.Errorf("HashIsEqual: expected false, got true")
	}
	if IsEqual(c, d) {
		t.Errorf("HashIsEqual: expected false, got true")
	}
	if IsEqual(a, nil) {
		t.Errorf("HashIsEqual: expected false, got true")
	}
	if IsEqual(nil, b) {
		t.Errorf("HashIsEqual: expected false, got true")
	}
	if !IsEqual(nil, nil) {
		t.Errorf("HashIsEqual: expected true, got false")
	}
}

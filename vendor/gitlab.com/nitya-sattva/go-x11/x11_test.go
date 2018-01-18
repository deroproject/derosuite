// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package x11

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestHash(t *testing.T) {
	hs := New()
	out := [32]byte{}

	for i := range tsInfo {
		ln := len(tsInfo[i].out)
		dest := make([]byte, ln)

		hs.Hash(tsInfo[i].in[:], out[:])
		if ln != hex.Encode(dest, out[:]) {
			t.Errorf("%s: invalid length", tsInfo[i])
		}
		if !bytes.Equal(dest[:], tsInfo[i].out[:]) {
			t.Errorf("%s: invalid hash", tsInfo[i].id)
		}
	}
}

////////////////

var tsInfo = []struct {
	id  string
	in  []byte
	out []byte
}{
	{
		"Empty",
		[]byte(""),
		[]byte("51b572209083576ea221c27e62b4e22063257571ccb6cc3dc3cd17eb67584eba"),
	},
	{
		"Dash",
		[]byte("DASH"),
		[]byte("fe809ebca8753d907f6ad32cdcf8e5c4e090d7bece5df35b2147e10b88c12d26"),
	},
	{
		"Fox",
		[]byte("The quick brown fox jumps over the lazy dog"),
		[]byte("534536a4e4f16b32447f02f77200449dc2f23b532e3d9878fe111c9de666bc5c"),
	},
}

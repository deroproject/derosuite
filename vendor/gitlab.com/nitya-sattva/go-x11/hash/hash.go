// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package hash

import "io"

type Hash interface {
	// Write (via the embedded io.Writer interface) adds more
	// data to the running hash. It never returns an error.
	io.Writer

	// Reset resets the Hash to its initial state.
	Reset()

	// Sum appends the current hash to dst and returns the result
	// as a slice. It does not change the underlying hash state.
	Sum(dst []byte) []byte

	// Size returns the number of bytes Sum will return.
	Size() int

	// BlockSize returns the hash's underlying block size.
	// The Write method must be able to accept any amount
	// of data, but it may operate more efficiently if
	// all writes are a multiple of the block size.
	BlockSize() int
}

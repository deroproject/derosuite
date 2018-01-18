go-x11
======

Implements the x11 hash and required functions in go.


Usage
-----

```go
	package main

	import (
		"fmt"
		"gitlab.com/nitya-sattva/go-x11"
	)

	func main() {
		hs, out := x11.New(), [32]byte{}
		hs.Hash([]byte("DASH"), out[:])
		fmt.Printf("%x \n", out[:])
	}
```


Notes
-----

Echo, Simd and Shavite do not have 100% test coverage, a full test on these
requires the test to hash a blob of bytes that is several gigabytes large.


License
-------

go-x11 is licensed under the [copyfree](http://copyfree.org) ISC license.

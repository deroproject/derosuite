FastGoSkein
===========

A go port of the skein hash function originally written in C from https://github.com/cberner/xkcd_miner. Ive made a few tweaks an optimizations on the original code for added performance.
Note: Its still not as fast as the original, but I will continue to chip away at optimizations as much as I can!

#Use:

Install the package with `go get`

    go get github.com/whyrusleeping/FastGoSkein

Import it in your golang code

    import "github.com/whyrusleeping/FastGoSkein"

And use it!

    toHash := []byte("This is a super secret keyphrase to be hashed!")
    sk := new(skein.Skein1024)
    sk.Init(1024)
    sk.Update(toHash)
    outputBuffer := make([]byte, 128)
    sk.Final(outputBuffer)

    //Nice hex formatting from encoding/hex
    fmt.Println(hex.EncodeToString(outputBuffer))

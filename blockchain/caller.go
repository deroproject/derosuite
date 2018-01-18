package blockchain

// this package identifies the 2nd level caller of this function
// this is done to ensure  checks regarding locking etc and to be sure no spuros calls are possible

import "runtime"

func CallerName() string {
	pc, _, _, ok := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	if ok && details != nil {
		//fmt.Printf("called from %s\n", details.Name()) // we should only give last parse after .
		return details.Name()
	}

	return ""
}

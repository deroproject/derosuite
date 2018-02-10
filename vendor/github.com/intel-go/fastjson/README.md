# fastjson: optimized standard library JSON for Go

`fastjson` has the same API as json from standard library `encoding/json`. 
The `Unmarshal` and `Decode` functions are faster, but everything else is the same as `encoding/json`

## Getting Started
```
$go get github.com/intel-go/fastjson
```
##Perfomance
The performance depends on the content of your json structures, not the structure you parse to.
If `.json` has a lot of strings or numbers, fastjson is significantly faster than `encoding/json`


##Example
```Go
import (
    "github.com/intel-go/fastjson"
    "fmt"
)

func main() {
    var jsonBlob = []byte(`[
	{"Name": "Platypus", "Order": "Monotremata"},
	{"Name": "Quoll",    "Order": "Dasyuromorphia"}
    ]`)
    type Animal struct {
	Name  string
	Order string
    }
    var animals []Animal
    err := fastjson.Unmarshal(jsonBlob, &animals)
    if err != nil {
	fmt.Println("error:", err)
    }
    fmt.Printf("%+v", animals)
    // Output:
    // [{Name:Platypus Order:Monotremata} {Name:Quoll Order:Dasyuromorphia}]
}
```
##API
API is the same as encoding/json
[GoDoc](https://golang.org/pkg/encoding/json/#Unmarshal)

package ringct


import "testing"
import "math/rand"

func Test_Range_and_Borromean_Signature(t *testing.T){
    var c,mask Key
    
    for i := 0; i < 50;i++{  // test it 500 times
    var amount uint64 = rand.Uint64()
    sig := ProveRange(&c,&mask,amount)
    if VerifyRange(&c, *sig) == false {
     t.Errorf("Range Test failed") 
     return
        }
    }
    
    
    
    
    
}

package generators

import (
	"fmt"
	"testing"
)

func TestBigInt(t *testing.T)  {
	t.SkipNow()
	biSigned := newBigInt(true)

	for i := 0; i< 10; i++ {
		fmt.Println(biSigned.Gen())
	}

	biUnSigned := newBigInt(false)
	for i := 0; i< 10; i++ {
		fmt.Println(biUnSigned.Gen())
	}

}

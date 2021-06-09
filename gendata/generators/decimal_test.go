package generators

import (
	"fmt"
	"testing"
)

func TestDecimal(t *testing.T) {
	d := &Decimal{}
	for i := 0; i < 10; i++ {
		fmt.Println(d.Gen())
	}
}

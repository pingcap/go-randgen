package generators

import (
	"fmt"
	"testing"
)

func TestInt(t *testing.T) {
	in := newInt(0, -1)
	for i := 0; i < 10; i++ {
		fmt.Println(in.Gen())
	}
}

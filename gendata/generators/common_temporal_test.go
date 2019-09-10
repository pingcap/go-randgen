package generators

import (
	"fmt"
	"testing"
)

func TestTemporal(t *testing.T) {
	temporal := newTemporal(yyyy, MM)

	for i := 0; i < 10; i++ {
		fmt.Println(temporal.Gen())
	}
}

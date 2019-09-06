package generators

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLetter(t *testing.T) {
	l := &Letter{}


	for i := 0; i < 100; i++ {
		assert.Condition(t, func() (success bool) {
			res := l.Gen()
			return len(res) == 1 && res[0] >= 'A' && res[0] <= 'Z'
		}, "generate must in 'A'~'Z'")
	}
}

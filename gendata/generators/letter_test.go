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
			return len(res) == 3 && res[1] >= 'a' && res[1] <= 'z'
		}, "generate must in 'a'~'z'")
	}
}

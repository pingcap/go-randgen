package generators

import (
	"github.com/magiconair/properties/assert"
	assert2 "github.com/stretchr/testify/assert"
	"testing"
)

func TestEnglish(t *testing.T) {
	e := newEnglish()
	assert.Equal(t, 100, len(e.dict))

	for i := 0; i < 10; i++ {
		// it cannot be empty
		word := e.Gen()
		assert2.NotEqual(t, "", word)

		// `\r` should be removed
		assert2.NotContains(t, "\r", word)
	}
}

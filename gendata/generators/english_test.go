package generators

import (
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestEnglish(t *testing.T) {
	e := newEnglish()
	assert.Equal(t, 100, len(e.dict))

	for i := 0; i < 10; i++ {
		e.Gen()
	}
}

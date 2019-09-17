package sql_generator

import (
	"github.com/magiconair/properties/assert"
	"testing"
)

func TestQueue(t *testing.T) {
	q := newQueue()
	assert.Equal(t, true, q.isEmpty())
	q.enqueue("haha")
	q.enqueue("popo")
	assert.Equal(t, "haha", q.dequeue())
	assert.Equal(t, false, q.isEmpty())
}

package gendata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTraverse(t *testing.T) {
	o, err := newOptions(nil, nil, "", nil)
	assert.Equal(t, nil, err)
	o.addField("arr0", []string{"0", "1"})
	o.addField("arr1", []string{"2", "3", "9"})
	o.addField("arr2", []string{"5", "7", "4"})

	expected := [][]string{
		{"0", "2", "5"},
		{"0", "2", "7"},
		{"0", "2", "4"},
		{"0", "3", "5"},
		{"0", "3", "7"},
		{"0", "3", "4"},
		{"0", "9", "5"},
		{"0", "9", "7"},
		{"0", "9", "4"},
		{"1", "2", "5"},
		{"1", "2", "7"},
		{"1", "2", "4"},
		{"1", "3", "5"},
		{"1", "3", "7"},
		{"1", "3", "4"},
		{"1", "9", "5"},
		{"1", "9", "7"},
		{"1", "9", "4"},
	}

	assert.Equal(t, 18, o.numbers)
	pos := 0
	err = o.traverse(func(cur []string) error {
		assert.Equal(t, expected[pos], cur)
		pos++
		return nil
	})
	assert.Equal(t, nil, err)
}

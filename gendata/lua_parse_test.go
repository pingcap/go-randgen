package gendata

import (
	"github.com/stretchr/testify/assert"
	"github.com/yuin/gopher-lua"
	"testing"
)

var testLuaScript = `
tables = {
    rows = {10, 20},
    -- SHOW CHARACTER SET;
    charsets = {'utf8', 'utf8mb4', 'ascii', 'latin1', 'binary'},
    partitions = {4, 6, 8, 15},
}

fields = {
    types = {'bigint', 'float', 'double'},
    sign = {'signed', 'unsigned'}
}

data = {
    numbers = {'null', 'tinyint', 'smallint',
        '12.991',
    },
    strings = {'null', 'letter', 'english', 'string(15)'}
}
`

func TestExtractSlice(t *testing.T) {
	l := lua.NewState()
	defer l.Close()
	if err := l.DoString(testLuaScript); err != nil {
		panic(err)
	}

	result, err := extractSlice(l, "tables", "rows", []string{})

	assert.Equal(t, nil, err)
	assert.Equal(t, []string{"10", "20"}, result)

	defaul := []string{"mm", "poiu"}

	result, err = extractSlice(l, "tables", "aaaa", defaul)
	assert.Equal(t, defaul, result)
}

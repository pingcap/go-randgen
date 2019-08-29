package gendata

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFields(t *testing.T) {
	zzScript := `
fields = {
    types = {'bigint', 'float', 'double', 'varchar'},
    sign = {'signed', 'unsigned'},
    keys = {'undef', 'key'}
}
`
	l, err := runLua(zzScript)
	assert.Equal(t, nil, err)

	fields, err := newFields(l)
	assert.Equal(t, nil, err)

	stmts, err := fields.gen()
	assert.Equal(t, nil, err)

	assert.Equal(t, 14, len(stmts))

	for _, stmt := range stmts {
		fmt.Println(stmt)
	}
}

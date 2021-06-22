package gendata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTables(t *testing.T) {
	zzScript := `
tables = {
    rows = {10, 20, 30},
    charsets = {'utf8', 'utf8mb4', 'undef'},
    partitions = {4, 6, 'undef'},
}
`
	l, err := runLua(zzScript)
	assert.Equal(t, nil, err)

	tables, err := newTables(l)
	assert.Equal(t, nil, err)

	stmts, err := tables.gen()
	assert.Equal(t, nil, err)

	assert.Equal(t, tables.numbers, len(stmts))

	/*	for _, stmt := range stmts {
		fmt.Println("==========")
		fmt.Println(stmt.format)
		fmt.Println(stmt.rowNum)
	}*/
}

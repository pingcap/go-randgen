package gendata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestData(t *testing.T) {
	zzScript := `
data = {
    numbers = {'null', 'tinyint', 'smallint',
        '12.991', '1.009', 'decimal',
    },
    strings = {'null', 'letter', 'english'},
    temporals = { 'date', 'year', 'null', undef, '2019-08-23', '2018-09-10 10:29:30'},
}
`
	l, err := runLua(zzScript)
	assert.Equal(t, nil, err)

	t.Run("test one record gen", func(t *testing.T) {
		data, err := newData(l)
		assert.Equal(t, nil, err)

		recordGen := data.getRecordGen([]*fieldExec{
			{
				name: "",
				tp:   "enum",
			},
			{
				name: "",
				tp:   "enum",
			},
			{
				name: "",
				tp:   "enum",
			},
		})

		row := make([]string, 3)

		recordGen.oneRow(row)
	})

}

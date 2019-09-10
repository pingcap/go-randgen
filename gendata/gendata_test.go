package gendata

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenDdls(t *testing.T) {
	testScript := `
tables = {
    rows = {10, 20, 30},
    charsets = {'utf8', 'utf8mb4', 'undef'},
    partitions = {4, 6, 'undef'},
}

fields = {
    types = {'bigint', 'float', 'double', 'enum', 'decimal(10,4)'},
    sign = {'signed', 'unsigned'},
    keys = {'undef', 'key'}
}

data = {
    numbers = {'null', 'tinyint', 'smallint',
        '12.991', '1.009', '-9.1823',
        'decimal',
    },
    strings = {'null', 'letter', 'english'},
}
`

	l, err := runLua(testScript)
	assert.Equal(t, nil, err)

	config, err := newZzConfig(l)
	assert.Equal(t, nil, err)

	t.Run("test gen ddls", func(t *testing.T) {
		ddls, fieldExecs, err := config.genDdls()
		assert.Equal(t, nil, err)

		assert.Equal(t, config.Tables.numbers, len(ddls))
		assert.Equal(t, 18, len(fieldExecs))

/*		for _, sql := range ddls {
			fmt.Println(sql.ddl)
		}

		for _, exec := range fieldExecs {
			fmt.Println(exec)
		}*/

	})

	t.Run("gen sqls", func(t *testing.T) {
		sqls, kf, err := ByConfig(config)
		assert.Equal(t, nil, err)
		assert.Equal(t, config.Tables.numbers * 2, len(sqls))
		fmt.Println(kf["_digit"]())
/*		for _, sql := range sqls {
			fmt.Println(sql)
		}*/
	})

}

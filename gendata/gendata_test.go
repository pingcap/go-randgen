package gendata

import (
	"github.com/DATA-DOG/go-sqlmock"
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
		sqls, _, err := ByConfig(config)
		assert.Equal(t, nil, err)
		assert.Equal(t, config.Tables.numbers * 2, len(sqls))
	})

}

func TestByDb(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Equal(t, nil, err)

	tableOrders := []string{"table1", "table2", "table3"}

	tableSet := make(map[string]bool)
	for _, tbl := range tableOrders {
		tableSet[tbl] = true
	}

	rows := sqlmock.NewRows([]string{"Tables_in_test"})
	for _, tname := range tableOrders {
		rows.AddRow(tname)
	}

	mock.ExpectQuery("show tables").
		WillReturnRows(rows)

	type fieldInfo struct {
		tp   string
	}

	infoOrders := []string{"v1", "v2"}

	infos := map[string]*fieldInfo{
		"v1":{
			tp:"int(11)",
		},
		"v2":{
			tp: "varchar(255)",
		},
	}


	fRows := sqlmock.NewRows([]string{"Field", "Type", "Null",
		"Key", "Default", "Extra"})

	for _, infoName := range infoOrders {
		info := infos[infoName]
		fRows.AddRow(infoName, info.tp, "YES", "", nil, "")
	}

	mock.ExpectQuery("desc table1").
		WillReturnRows(fRows)

	kf, err := ByDb(db)
	assert.Equal(t, nil, err)

	for i := 0; i < 50; i++ {
		assert.Condition(t, func() (success bool) {
			res, err := kf["_table"]()
			assert.Equal(t, nil, err)
			_, ok := tableSet[res]
			return ok
		})

		assert.Condition(t, func() (success bool) {
			res, err := kf["_field"]()
			assert.Equal(t, nil, err)
			_, ok := infos[res[1:len(res)-1]]
			return ok
		})

		assertMustEqual(t, "`v1`", kf["_field_int"])

		assertMustEqual(t, "`v2`", kf["_field_char"])
	}

	assertMustEqual(t, "`v1`,`v2`", kf["_field_list"])

	assertMustEqual(t, "`v1`", kf["_field_int_list"])
	assertMustEqual(t, "`v2`", kf["_field_char_list"])
}

func assertMustEqual(t *testing.T, expected string, kf func() (string, error)) {
	res, err := kf()
	assert.Equal(t, nil, err)
	assert.Equal(t, expected, res)
}

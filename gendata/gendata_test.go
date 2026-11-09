package gendata

import (
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	defer l.Close()

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
		assert.Equal(t, config.Tables.numbers*2, len(sqls))
	})

}

func TestByDbQuotesCatalogIdentifiers(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	tableName := "safe`; DROP DATABASE production; --"
	fieldName := "value`; GRANT ALL ON *.* TO attacker; --"
	mock.ExpectQuery("show tables").WillReturnRows(
		sqlmock.NewRows([]string{"Tables_in_test"}).AddRow(tableName))
	mock.ExpectQuery(regexp.QuoteMeta("desc `safe``; DROP DATABASE production; --`")).
		WillReturnRows(sqlmock.NewRows([]string{"Field", "Type", "Null", "Key", "Default", "Extra"}).
			AddRow(fieldName, "int(11)", "YES", "", nil, ""))

	kf, err := ByDb(db, "mysql")
	require.NoError(t, err)
	assertMustEqual(t, "`safe``; DROP DATABASE production; --`", kf["_table"])
	assertMustEqual(t, "`value``; GRANT ALL ON *.* TO attacker; --`", kf["_field"])
	assertMustEqual(t, "`value``; GRANT ALL ON *.* TO attacker; --`", kf["_field_list"])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestByDbBindsCatalogLookup(t *testing.T) {
	for _, test := range []struct {
		name      string
		dbms      string
		tablesSQL string
		fieldsSQL string
		fieldRows *sqlmock.Rows
		tableName string
		wantTable string
	}{
		{
			name:      "postgres",
			dbms:      "postgres",
			tablesSQL: "SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname='public';",
			fieldsSQL: "SELECT table_name, column_name, data_type FROM information_schema.columns WHERE table_schema = 'public' AND table_name = $1;",
			fieldRows: sqlmock.NewRows([]string{"table_name", "column_name", "data_type"}).AddRow("odd\"name", "id", "integer"),
			tableName: "odd\"name",
			wantTable: `"odd""name"`,
		},
		{
			name:      "sqlite",
			dbms:      "sqlite3",
			tablesSQL: "SELECT name FROM sqlite_master WHERE type='table';",
			fieldsSQL: "SELECT name, type FROM pragma_table_info(?)",
			fieldRows: sqlmock.NewRows([]string{"name", "type"}).AddRow("id", "integer"),
			tableName: "odd\"name",
			wantTable: `"odd""name"`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer db.Close()

			mock.ExpectQuery(regexp.QuoteMeta(test.tablesSQL)).WillReturnRows(
				sqlmock.NewRows([]string{"name"}).AddRow(test.tableName))
			mock.ExpectQuery(regexp.QuoteMeta(test.fieldsSQL)).WithArgs(test.tableName).WillReturnRows(test.fieldRows)

			kf, err := ByDb(db, test.dbms)
			require.NoError(t, err)
			assertMustEqual(t, test.wantTable, kf["_table"])
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestByDbSQLiteIntegration(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE "odd""table" ("odd""field" INTEGER)`)
	require.NoError(t, err)
	kf, err := ByDb(db, "sqlite3")
	require.NoError(t, err)

	assertMustEqual(t, `"odd""table"`, kf["_table"])
	assertMustEqual(t, `"odd""field"`, kf["_field"])
	tableName, err := kf["_table"]()
	require.NoError(t, err)
	fieldName, err := kf["_field"]()
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO " + tableName + " (" + fieldName + ") VALUES (1)")
	require.NoError(t, err)
}

func TestByDb(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Equal(t, nil, err)

	tableOrders := []string{"table1", "table2", "table3"}

	tableSet := make(map[string]bool)
	for _, tbl := range tableOrders {
		tableSet["`"+tbl+"`"] = true
	}

	rows := sqlmock.NewRows([]string{"Tables_in_test"})
	for _, tname := range tableOrders {
		rows.AddRow(tname)
	}

	mock.ExpectQuery("show tables").
		WillReturnRows(rows)

	type fieldInfo struct {
		tp string
	}

	infoOrders := []string{"v1", "v2"}

	infos := map[string]*fieldInfo{
		"v1": {
			tp: "int(11)",
		},
		"v2": {
			tp: "varchar(255)",
		},
	}

	fRows := sqlmock.NewRows([]string{"Field", "Type", "Null",
		"Key", "Default", "Extra"})

	for _, infoName := range infoOrders {
		info := infos[infoName]
		fRows.AddRow(infoName, info.tp, "YES", "", nil, "")
	}

	mock.ExpectQuery(regexp.QuoteMeta("desc `table1`")).
		WillReturnRows(fRows)

	kf, err := ByDb(db, "mysql")
	require.NoError(t, err)

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

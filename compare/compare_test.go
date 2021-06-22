package compare

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

type mockQuery struct {
	sql         string
	exec        bool
	header      []string
	db1MockRes  [][]driver.Value
	db1err      error
	db1Affected int64
	db2MockRes  [][]driver.Value
	db2err      error
	db2Affected int64
}

func getRows(header []string, rows [][]driver.Value) *sqlmock.Rows {
	mockRows := sqlmock.NewRows(header)
	for _, row := range rows {
		mockRows.AddRow(row...)
	}

	return mockRows
}

func initMockDb(t *testing.T, mqs []*mockQuery) (*sql.DB, *sql.DB) {
	db1, mock1, err := sqlmock.New()
	assert.Equal(t, nil, err)

	db2, mock2, err := sqlmock.New()
	assert.Equal(t, nil, err)

	for _, mq := range mqs {

		if mq.exec {
			if mq.db1err != nil {
				mock1.ExpectExec(mq.sql).
					WillReturnError(mq.db1err)
			} else {
				mock1.ExpectExec(mq.sql).
					WillReturnResult(sqlmock.NewResult(0, mq.db1Affected))
			}

			if mq.db2err != nil {
				mock2.ExpectExec(mq.sql).
					WillReturnError(mq.db2err)
			} else {
				mock2.ExpectExec(mq.sql).
					WillReturnResult(sqlmock.NewResult(0, mq.db2Affected))
			}

		} else {
			if mq.db1err != nil {
				mock1.ExpectQuery(mq.sql).
					WillReturnError(mq.db1err)
			} else {
				mock1.ExpectQuery(mq.sql).
					WillReturnRows(getRows(mq.header, mq.db1MockRes))
			}

			if mq.db2err != nil {
				mock2.ExpectQuery(mq.sql).
					WillReturnError(mq.db2err)
			} else {
				mock2.ExpectQuery(mq.sql).
					WillReturnRows(getRows(mq.header, mq.db2MockRes))
			}
		}
	}

	return db1, db2
}

/*
Test follow situations:
 1. both query err
 2. one query err
 3. no err, rows is consistent
 4. no err, rows is inconsistent
 5. no err, record num is inconsistent
 6. no err, row is non order
*/
var mqs = []*mockQuery{
	// 1 both query err
	{
		sql:    "select a from test0",
		db1err: errors.New("test1 error1"),
		db2err: errors.New("test1 error2"),
	},
	// 2 one query err
	{
		sql:    "SELECT b FROM test1",
		db1err: errors.New("test2 error1"),
		header: []string{"name", "age", "sex"},
		db2MockRes: [][]driver.Value{
			{"Tom", 10, "male"},
		},
	},
	// 3 no err, rows is consistent
	{
		sql:    "SELECT c FROM test2",
		header: []string{"name", "age", "sex"},
		db1MockRes: [][]driver.Value{
			{"Tom", 11, "male"},
			{"Lily", 29, "female"},
		},
		db2MockRes: [][]driver.Value{
			{"Tom", 11, "male"},
			{"Lily", 29, "female"},
		},
	},
	// 4 no err, rows is inconsistent
	{
		sql:    "select d from test3",
		header: []string{"name", "age", "sex"},
		db1MockRes: [][]driver.Value{
			{"Tom", 10, "male"},
			{"Lily", 29, "female"},
		},
		db2MockRes: [][]driver.Value{
			{"Tom", 11, "male"},
			{"Lily", 29, "female"},
		},
	},
	// 5 no err, record num is inconsistent
	{
		sql:    "SELECT E FROM test4",
		header: []string{"name", "age", "sex"},
		db1MockRes: [][]driver.Value{
			{"Tom", 10, "male"},
			{"Lily", 29, "female"},
			{"Zhangsan", nil, "male"},
			{"Wang", nil, nil},
			{"Worker", 13, "male"},
		},
		db2MockRes: [][]driver.Value{
			{"Tom", 10, "male"},
			{"Lily", 29, "female"},
			{"Zhangsan", nil, "male"},
			{"Worker", 13, "male"},
		},
	},
	// 6 no err, row is non order
	{
		sql:    "SELECT mmm FROM test5",
		header: []string{"name", "age", "sex"},
		db1MockRes: [][]driver.Value{
			{"Tom", 10, "male"},
			{"Lily", 29, "female"},
			{"Zhangsan", nil, "male"},
			{"Wang", nil, nil},
		},
		db2MockRes: [][]driver.Value{
			{"Wang", nil, nil},
			{"Lily", 29, "female"},
			{"Tom", 10, "male"},
			{"Zhangsan", nil, "male"},
		},
	},
	// 7 8 test exec
	{
		sql:         "CREATE a SET m=10",
		exec:        true,
		db1Affected: 0,
		db2Affected: 0,
	},
	{
		sql:         "UPDATE a SET m=10",
		exec:        true,
		db1Affected: 0,
		db2Affected: 1,
	},
}

func getMqSqls() []string {
	sqls := make([]string, 0)
	for _, mq := range mqs {
		sqls = append(sqls, mq.sql)
	}

	return sqls
}

func TestByDb(t *testing.T) {

	t.Run("ordered test", func(t *testing.T) {
		db1, db2 := initMockDb(t, mqs)
		defer db1.Close()
		defer db2.Close()

		// expected visit order
		expected := []int{1, 3, 4, 5, 7}
		counter := 0

		err := ByDb(getMqSqls(), db1, db2, false,
			func(sql string, dsn1Res DsnRes, dsn2Res DsnRes) error {
				// corresponding mock query
				correMp := mqs[expected[counter]]

				assert.Equal(t, correMp.sql, sql)
				counter++
				return nil
			})
		assert.Equal(t, nil, err)
	})

	t.Run("non ordered test", func(t *testing.T) {
		db1, db2 := initMockDb(t, mqs)
		defer db1.Close()
		defer db2.Close()

		expected := []int{1, 3, 4, 7}
		counter := 0

		err := ByDb(getMqSqls(), db1, db2, true,
			func(sql string, dsn1Res DsnRes, dsn2Res DsnRes) error {

				correMp := mqs[expected[counter]]

				assert.Equal(t, correMp.sql, sql)

				counter++
				return nil
			})
		assert.Equal(t, nil, err)
	})
}

func TestSimplePrint(t *testing.T) {
	t.SkipNow()
	sql := `SELECT * FROM table_90_utf8_6`

	db1, err := OpenDBWithRetry("mysql", "root:@tcp(127.0.0.1:4000)/randgen")
	assert.Equal(t, nil, err)

	db2, err := OpenDBWithRetry("mysql", "root:123456@tcp(127.0.0.1:4406)/randgen")
	assert.Equal(t, nil, err)

	res1 := newQueryDsnRes(db1, sql)
	res2 := newQueryDsnRes(db2, sql)
	assert.Equal(t, nil, res1.err)
	assert.Equal(t, nil, res2.err)

	fmt.Println(res1.Res.NonOrderEqualTo(res2.Res))
}

package compare

import (
	"database/sql"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
)

func TestSqlResult_BytesEqualTo(t *testing.T) {
	var mockRes1 = &SqlResult{
		Header:[]string{"aaa", "bbbb"},
		Data:[][][]byte{
			{
				[]byte("haha"),
				[]byte("baba"),
			},
			{
				[]byte("mmmm"),
				[]byte("popo"),
			},
		},
	}

	var mockRes2 = &SqlResult{
		Header:[]string{"aaa", "bbbb"},
		Data:[][][]byte{
			{
				[]byte("haha"),
				[]byte("baba"),
			},
			{
				[]byte("mmmm"),
				[]byte("popo"),
			},
		},
	}

	var mockRes3 = &SqlResult{
		Header:[]string{"aaa", "bbbb"},
		Data:[][][]byte{
			{
				[]byte("mmmm"),
				[]byte("popo"),
			},
			{
				[]byte("haha"),
				[]byte("baba"),
			},
		},
	}

	assert.Equal(t, true, mockRes1.BytesEqualTo(mockRes2))
	assert.Equal(t, false, mockRes1.BytesEqualTo(mockRes3))
}

func TestQuery(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Equal(t, nil, err)
	defer db.Close()

	mock1Rows := sqlmock.NewRows([]string{"aaa", "bb", "ccccc"})
	mock1Rows.AddRow("", nil, "ojji")
	mock1Rows.AddRow("poi", 10, "drr")
	mock1Rows.AddRow("fd", 765, "dewrtty")

	q1 := "select test1"

	expected1 := `+-----------------------+
| aaa  | bb   | ccccc   |
+-----------------------+
|      | NULL | ojji    |
| poi  | 10   | drr     |
| fd   | 765  | dewrtty |
+-----------------------+`

	mock.ExpectQuery(q1).
		WillReturnRows(mock1Rows)

	mock2Rows := sqlmock.NewRows([]string{"aaa", "bb", "ccccc"})
	mock2Rows.AddRow("fd", 765, "dewrtty")
	mock2Rows.AddRow("", nil, "ojji")
	mock2Rows.AddRow("poi", 10, "drr")
	q2 := "select test1"

	mock.ExpectQuery(q2).
		WillReturnRows(mock2Rows)

	r1, err := query(db, q1)
	assert.Equal(t, nil, err)

	r2, err := query(db, q2)
	assert.Equal(t, nil, err)

	assert.Equal(t, expected1, r1.String())

	assert.Equal(t, true, r1.NonOrderEqualTo(r2))
	assert.Equal(t, false, r1.BytesEqualTo(r2))
}

func getMockDb(t *testing.T, expects []string) *sql.DB {
	db, mock, err := sqlmock.New()
	assert.Equal(t, nil, err)
	for i, expect := range expects {
		if i == 10 {
			mock.ExpectExec(expect).WillReturnError(errors.New("mock err"))
		} else {
			mock.ExpectExec(expect).WillReturnResult(sqlmock.NewResult(1, 1))
		}
	}

	return db
}

func getSql(num int) []string {
	sqls := make([]string, 0, num)
	for i := 0; i < num; i++ {
		sqls = append(sqls, fmt.Sprintf("exec %d", i))
	}
	return sqls
}

func TestExecSqlsInDbs(t *testing.T) {

	t.Run("success case", func(t *testing.T) {
		expectSqls := getSql(9)
		mockdb0 := getMockDb(t, expectSqls)
		mockdb1 := getMockDb(t, expectSqls)
		_, err := ExecSqlsInDbs(expectSqls, mockdb0, mockdb1)
		assert.Equal(t, nil, err)
	})

	t.Run("fail case", func(t *testing.T) {
		expectSqls := getSql(30)
		mockdb0 := getMockDb(t, expectSqls)
		mockdb1 := getMockDb(t, expectSqls)
		goroutineNum := runtime.NumGoroutine()
		sql, err := ExecSqlsInDbs(expectSqls, mockdb0, mockdb1)
		assert.Equal(t, "mock err", err.Error())
		assert.Equal(t, "exec 10", sql)
		// test  gorountine leak
		assert.Equal(t, goroutineNum, runtime.NumGoroutine())
	})

}




func TestQueryMysql(t *testing.T)  {
	t.SkipNow()
	db, err := OpenDBWithRetry("mysql", "root:123456@tcp(127.0.0.1:3306)/randgen")
	assert.Equal(t, nil, err)

	result, err := db.Exec("select * from test")

	//result, err := query(db, "update test set a=10 where a=9")
	fmt.Println(result.RowsAffected())
	fmt.Println(err)
}

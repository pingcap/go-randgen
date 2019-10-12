package compare

import (
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
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

func TestQueryMysql(t *testing.T)  {
	t.SkipNow()
	db, err := OpenDBWithRetry("mysql", "root:123456@tcp(127.0.0.1:3306)/randgen")
	assert.Equal(t, nil, err)

	result, err := db.Exec("select * from test")

	//result, err := query(db, "update test set a=10 where a=9")
	fmt.Println(result.RowsAffected())
	fmt.Println(err)
}

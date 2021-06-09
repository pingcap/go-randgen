package main

import (
	"fmt"
	"github.com/pingcap/go-randgen/compare"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExecErr(t *testing.T) {
	reInitCmd()
	_, err := executeCommand(rootCmd, "exec")
	assert.Equal(t, "yy are required", err.Error())

	_, err = executeCommand(rootCmd, "exec", "-Y", "yyy", "--dsn1", "ddd1")
	assert.Equal(t, "dsn must have a pair", err.Error())
}

func TestGetColorDiff(t *testing.T) {
	t.SkipNow()
	s1 := "mmm\nppp"
	s2 := "ddd\nuplo"

	c1, c2 := getColorDiff(s1, s2)
	fmt.Println(c1)
	fmt.Println("===")
	fmt.Println(c2)
}

var mockRes1 = &compare.SqlResult{
	Header: []string{"aaa", "bbbb"},
	Data: [][][]byte{
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

func TestDumpInfo(t *testing.T) {
	info := &dumpInfo{
		num:  1,
		sql:  "select * from test",
		dsn1: "xxx:password@protocol(address)/dbname",
		dsn2: "yyy:password@protocol(address)/dbname",
		dsn1Res: &compare.QueryDsnRes{
			Res: mockRes1,
		},
		dsn2Res: &compare.QueryDsnRes{
			Res: mockRes1,
		},
	}

	expected := `[sql]

select * from test

[err]

[[xxx:password@protocol(address)/dbname]]

[[yyy:password@protocol(address)/dbname]]

[compare]

[[xxx:password@protocol(address)/dbname]]

+-------------+
| aaa  | bbbb |
+-------------+
| haha | baba |
| mmmm | popo |
+-------------+

[[yyy:password@protocol(address)/dbname]]

+-------------+
| aaa  | bbbb |
+-------------+
| haha | baba |
| mmmm | popo |
+-------------+`

	assert.Equal(t, expected, info.String())
}

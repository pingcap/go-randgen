package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/pingcap/go-randgen/compare"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

[[xxx:******@protocol(address)/dbname]]

[[yyy:******@protocol(address)/dbname]]

[compare]

[[xxx:******@protocol(address)/dbname]]

+-------------+
| aaa  | bbbb |
+-------------+
| haha | baba |
| mmmm | popo |
+-------------+

[[yyy:******@protocol(address)/dbname]]

+-------------+
| aaa  | bbbb |
+-------------+
| haha | baba |
| mmmm | popo |
+-------------+`

	actual := info.String()
	assert.Equal(t, expected, actual)
	assert.NotContains(t, actual, "password")
}

func TestDumpVisitorCreatesOwnerOnlyLog(t *testing.T) {
	oldDumpDir := dumpDir
	defer func() { dumpDir = oldDumpDir }()

	dumpDir = filepath.Join(t.TempDir(), "dump")
	require.NoError(t, createDumpDir(dumpDir))

	result1 := &compare.QueryDsnRes{Res: mockRes1}
	result2 := &compare.QueryDsnRes{Res: mockRes1}
	err := dumpVisitor(
		"root:secret@tcp(localhost:4000)/one",
		"root:secret@tcp(localhost:4000)/two",
	)("select 1", result1, result2)
	require.NoError(t, err)

	logPath := filepath.Join(dumpDir, "0.log")
	data, err := ioutil.ReadFile(logPath)
	require.NoError(t, err)
	assert.NotContains(t, string(data), "secret")
	assert.Contains(t, string(data), "******")

	dirInfo, err := os.Stat(dumpDir)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0700), dirInfo.Mode().Perm())

	fileInfo, err := os.Stat(logPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), fileInfo.Mode().Perm())
}

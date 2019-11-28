package main

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestEmptyYyOption(t *testing.T) {
	reInitCmd()
	_, err := executeCommand(rootCmd, "gentest")
	assert.Equal(t, "yy are required", err.Error())
}

const testCreateUniqueTableExpect = `CREATE TABLE table1 (a int);
CREATE TABLE table2 (a int);
CREATE TABLE table3 (a int);
CREATE TABLE table4 (a int);
CREATE TABLE table5 (a int);`

func TestCreateUniqueTable(t *testing.T) {
	reInitCmd()
	_, err := executeCommand(rootCmd, "gentest", "-Y",
		"../../examples/toturial/create_unique_table.yy", "-B", "-Q", "5", "-O", "cute", "--skip-zz")
	assert.Equal(t, nil, err)

	randFilePath := "./cute.rand.sql"
	content, err := ioutil.ReadFile(randFilePath)
	assert.Equal(t, nil, err)
	assert.Equal(t, testCreateUniqueTableExpect, string(content))

	err = os.Remove(randFilePath)
	assert.Equal(t, nil, err)
}

func TestUpdateSql(t *testing.T) {
	reInitCmd()
	_, err := executeCommand(rootCmd, "gentest", "-Y",
		"../../examples/toturial/test_update.yy", "-B", "-Q", "4", "-O", "upda")
	assert.Equal(t, nil, err)

	randFilePath := "./upda.rand.sql"
	content, err := ioutil.ReadFile(randFilePath)
	assert.Equal(t, nil, err)

	sqls := strings.Split(string(content), "\n")
	assert.Equal(t, 4, len(sqls))
	var tableName string
	for i := 0; i < 4; i++ {
		if i == 0 {
			assert.Equal(t, "BEGIN;", sqls[i])
			continue
		}
		if i == 3 {
			assert.Equal(t, "END;", sqls[i])
			continue
		}

		if i == 1 {
			// UPDATE table_20_utf8_4 SET `col_bigint_undef_signed` = 10;
			tableName = strings.Split(sqls[i], " ")[1]
		}
		if i == 2 {
			// SELECT * FROM table_20_utf8_4;
			selectedTable := strings.Split(sqls[i], " ")[3]
			assert.Equal(t, tableName, selectedTable[:len(selectedTable)-1])
		}
	}

	err = os.Remove(randFilePath)
	assert.Equal(t, nil, err)
	err = os.Remove("./upda.data.sql")
}


func TestSeed(t *testing.T) {

	for i := 0; i < 10; i++ {
		reInitCmd()
		_, err := executeCommand(rootCmd, "gentest", "-Y",
			"../../examples/toturial/embed_lua.yy", "-B", "-Q", "5", "-O",
			"lua", "--skip-zz", "--seed", "0")
		assert.Equal(t, nil, err)

		randFilePath := "./lua.rand.sql"
		content, err := ioutil.ReadFile(randFilePath)
		assert.Equal(t, nil, err)

		expected := `0;
0;
3;
0;
3;`

		assert.Equal(t, expected, string(content))
		err = os.Remove(randFilePath)
		assert.Equal(t, nil, err)
	}
}

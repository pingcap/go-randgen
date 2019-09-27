package main

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
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

	content, err := ioutil.ReadFile("./cute.rand.sql")
	assert.Equal(t, nil, err)
	assert.Equal(t, testCreateUniqueTableExpect, string(content))
}

package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/spf13/cobra"
	"io/ioutil"
	"testing"
)

func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	_, output, err = executeCommandC(root, args...)
	return output, err
}
func executeCommandC(root *cobra.Command, args ...string) (c *cobra.Command,
	output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs(args)
	c, err = root.ExecuteC()
	return c, buf.String(), err
}

func TestEmptyYyOption(t *testing.T) {
	_, err := executeCommand(rootCmd)
	assert.Equal(t, "yy are required", err.Error())
}

const testCreateUniqueTableExpect = `CREATE TABLE table1 (a int);
CREATE TABLE table2 (a int);
CREATE TABLE table3 (a int);
CREATE TABLE table4 (a int);
CREATE TABLE table5 (a int);`

func TestCreateUniqueTable(t *testing.T) {
	_, err := executeCommand(rootCmd, "-Y",
		"../../examples/create_unique_table.yy", "-B", "-Q", "5", "-O", "cute", "--skip-zz")
	assert.Equal(t, nil, err)

	content, err := ioutil.ReadFile("./cute.rand.sql")
	assert.Equal(t, nil, err)
	assert.Equal(t, testCreateUniqueTableExpect, string(content))
}
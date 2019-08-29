package main

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/spf13/cobra"
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
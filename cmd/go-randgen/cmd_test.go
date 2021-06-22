package main

import (
	"bytes"

	"github.com/spf13/cobra"
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

// clear cmd after one test, to not affect next unit test
func reInitCmd() {
	rootCmd.ResetFlags()
	rootCmd.ResetCommands()
	initCmd()

}

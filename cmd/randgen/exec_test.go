package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExecErr(t *testing.T) {
	_, err := executeCommand(rootCmd, "exec")
	assert.Equal(t, "yy are required", err.Error())

	_, err = executeCommand(rootCmd, "exec", "-Y", "yyy", "--dsn1", "ddd1")
	assert.Equal(t, "dsn must have a pair", err.Error())
}

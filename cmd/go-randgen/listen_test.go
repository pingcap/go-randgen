package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListenRequiresGraphToken(t *testing.T) {
	original, hadToken := os.LookupEnv("GO_RANDGEN_GRAPH_TOKEN")
	defer func() {
		if hadToken {
			os.Setenv("GO_RANDGEN_GRAPH_TOKEN", original)
		} else {
			os.Unsetenv("GO_RANDGEN_GRAPH_TOKEN")
		}
		reInitCmd()
	}()
	require.NoError(t, os.Unsetenv("GO_RANDGEN_GRAPH_TOKEN"))

	reInitCmd()
	yyPath = "test.yy"
	cmd := newListenCmd()
	require.Error(t, cmd.PreRunE(cmd, nil))

	require.NoError(t, os.Setenv("GO_RANDGEN_GRAPH_TOKEN", "test-token"))
	require.NoError(t, cmd.PreRunE(cmd, nil))
	require.Equal(t, "test-token", graphToken)

	require.NoError(t, os.Unsetenv("GO_RANDGEN_GRAPH_TOKEN"))
	dir, err := ioutil.TempDir("", "randgen-graph-token")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	file := filepath.Join(dir, "token")
	require.NoError(t, ioutil.WriteFile(file, []byte("file-token\n"), 0600))
	require.NoError(t, cmd.Flags().Set("token-file", file))
	require.NoError(t, cmd.PreRunE(cmd, nil))
	require.Equal(t, "file-token", graphToken)
}

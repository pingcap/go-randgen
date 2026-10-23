package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGensqlCredentialSources(t *testing.T) {
	defer os.Unsetenv("GO_RANDGEN_DSN")
	require.NoError(t, os.Setenv("GO_RANDGEN_DSN", "user:env-secret@tcp(db:3306)/test"))

	cmd := newGensqlCmd()
	require.NoError(t, cmd.PreRunE(cmd, nil))
	assert.Equal(t, "user:env-secret@tcp(db:3306)/test", gensqlDsn)
}

func TestGensqlRejectsMultipleCredentialSources(t *testing.T) {
	defer os.Unsetenv("GO_RANDGEN_DSN")
	require.NoError(t, os.Setenv("GO_RANDGEN_DSN", "from-env"))

	cmd := newGensqlCmd()
	require.NoError(t, cmd.Flags().Set("dsn", "from-flag"))
	err := cmd.PreRunE(cmd, nil)

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "from-env")
	assert.NotContains(t, err.Error(), "from-flag")
}

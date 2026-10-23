package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveCredential(t *testing.T) {
	const envName = "GO_RANDGEN_TEST_DSN"
	defer os.Unsetenv(envName)

	t.Run("reads environment variable", func(t *testing.T) {
		require.NoError(t, os.Setenv(envName, "user:env-secret@tcp(db:3306)/test"))

		got, err := resolveCredential(credentialInput{envName: envName, label: "dsn"})

		require.NoError(t, err)
		assert.Equal(t, "user:env-secret@tcp(db:3306)/test", got)
	})

	t.Run("reads owner-only secret file and removes newline", func(t *testing.T) {
		require.NoError(t, os.Unsetenv(envName))
		path := writeCredentialFile(t, "user:file-secret@tcp(db:3306)/test\n", 0600)
		defer os.Remove(path)

		got, err := resolveCredential(credentialInput{filePath: path, envName: envName, label: "dsn"})

		require.NoError(t, err)
		assert.Equal(t, "user:file-secret@tcp(db:3306)/test", got)
	})

	t.Run("rejects multiple sources", func(t *testing.T) {
		require.NoError(t, os.Setenv(envName, "from-env"))

		_, err := resolveCredential(credentialInput{
			flagValue: "from-flag",
			flagSet:   true,
			envName:   envName,
			label:     "dsn",
		})

		require.Error(t, err)
		assert.NotContains(t, err.Error(), "from-env")
		assert.NotContains(t, err.Error(), "from-flag")
	})

	t.Run("rejects empty source", func(t *testing.T) {
		require.NoError(t, os.Setenv(envName, ""))

		_, err := resolveCredential(credentialInput{envName: envName, label: "dsn"})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "empty")
	})

	t.Run("rejects secret file readable by group or others", func(t *testing.T) {
		require.NoError(t, os.Unsetenv(envName))
		path := writeCredentialFile(t, "credential-value-leak-marker", 0644)
		defer os.Remove(path)

		_, err := resolveCredential(credentialInput{filePath: path, label: "dsn"})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "permissions")
		assert.NotContains(t, err.Error(), "credential-value-leak-marker")
	})
}

func TestResolveCredentialList(t *testing.T) {
	path := writeCredentialFile(t, "user:fake,marker@tcp(db:3306)/one\ndsn-two\n", 0600)
	defer os.Remove(path)

	got, err := resolveCredentialList(credentialInput{filePath: path, label: "dsns"})

	require.NoError(t, err)
	assert.Equal(t, []string{"user:fake,marker@tcp(db:3306)/one", "dsn-two"}, got)
}

func TestCommandsReadCredentialFiles(t *testing.T) {
	for _, name := range []string{"GO_RANDGEN_DSN", "GO_RANDGEN_DSNS", "GO_RANDGEN_DSN1", "GO_RANDGEN_DSN2"} {
		require.NoError(t, os.Unsetenv(name))
	}

	t.Run("gensql", func(t *testing.T) {
		path := writeCredentialFile(t, "gensql-dsn\n", 0600)
		defer os.Remove(path)
		cmd := newGensqlCmd()
		require.NoError(t, cmd.Flags().Set("dsn-file", path))
		require.NoError(t, cmd.PreRunE(cmd, nil))
		assert.Equal(t, "gensql-dsn", gensqlDsn)
	})

	t.Run("gendata", func(t *testing.T) {
		path := writeCredentialFile(t, "first-dsn\nsecond-dsn\n", 0600)
		defer os.Remove(path)
		cmd := newGenDataCmd()
		require.NoError(t, cmd.Flags().Set("dsns-file", path))
		require.NoError(t, cmd.PreRunE(cmd, nil))
		assert.Equal(t, []string{"first-dsn", "second-dsn"}, gendataDsns)
	})

	t.Run("exec", func(t *testing.T) {
		firstPath := writeCredentialFile(t, "first-dsn\n", 0600)
		secondPath := writeCredentialFile(t, "second-dsn\n", 0600)
		defer os.Remove(firstPath)
		defer os.Remove(secondPath)
		yyPath = "grammar.yy"
		cmd := newExecCmd()
		require.NoError(t, cmd.Flags().Set("dsn1-file", firstPath))
		require.NoError(t, cmd.Flags().Set("dsn2-file", secondPath))
		require.NoError(t, cmd.PreRunE(cmd, nil))
		assert.Equal(t, "first-dsn", dsn1)
		assert.Equal(t, "second-dsn", dsn2)
	})
}

func writeCredentialFile(t *testing.T, contents string, mode os.FileMode) string {
	t.Helper()
	file, err := ioutil.TempFile("", "go-randgen-credential-test-")
	require.NoError(t, err)
	path := file.Name()
	require.NoError(t, file.Close())
	require.NoError(t, ioutil.WriteFile(path, []byte(contents), mode))
	require.NoError(t, os.Chmod(path, mode))
	return path
}

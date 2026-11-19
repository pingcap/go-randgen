package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOfflineSQLCommandsRejectBlockedOutput(t *testing.T) {
	dir, err := ioutil.TempDir("", "randgen-sql-output")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	grammar := filepath.Join(dir, "test.yy")
	require.NoError(t, ioutil.WriteFile(grammar, []byte("query:\n    PRAGMA\n"), 0600))

	for _, tc := range []struct {
		name     string
		args     []string
		suffixes []string
	}{
		{"combined", []string{"gentest"}, []string{".sql"}},
		{"split", []string{"gentest", "-B"}, []string{".data.sql", ".rand.sql"}},
		{"metadata", []string{"gensql", "--dbms", "sqlite3", "--dsn", ":memory:"}, []string{".rand.sql"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reInitCmd()
			output := filepath.Join(dir, tc.name)
			args := append(tc.args, "-Y", grammar, "-Q", "1", "-O", output)
			_, err := executeCommand(rootCmd, args...)
			assert.EqualError(t, err, "SQL statement blocked by safety policy: PRAGMA operation")
			for _, suffix := range tc.suffixes {
				_, err := os.Stat(output + suffix)
				assert.True(t, os.IsNotExist(err), "unexpected SQL artifact: %s", suffix)
			}

			const existing = "SELECT 1;"
			for _, suffix := range tc.suffixes {
				require.NoError(t, ioutil.WriteFile(output+suffix, []byte(existing), 0600))
			}
			reInitCmd()
			_, err = executeCommand(rootCmd, args...)
			require.Error(t, err)
			for _, suffix := range tc.suffixes {
				content, err := ioutil.ReadFile(output + suffix)
				require.NoError(t, err)
				assert.Equal(t, existing, string(content))
			}
		})
	}
}

func TestOfflineSQLValidatesRenderedPostgresOutput(t *testing.T) {
	dir, err := ioutil.TempDir("", "randgen-sql-output")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	grammar := filepath.Join(dir, "test.yy")
	require.NoError(t, ioutil.WriteFile(grammar, []byte("query:\n    SELECT \"PRAGMA\"\n"), 0600))
	output := filepath.Join(dir, "allowed")

	reInitCmd()
	_, err = executeCommand(rootCmd, "gentest", "-B", "--skip-zz", "--dbms", "postgres",
		"-Y", grammar, "-Q", "1", "-O", output)
	require.NoError(t, err)
	sql, err := ioutil.ReadFile(output + ".rand.sql")
	require.NoError(t, err)
	assert.Equal(t, "SELECT 'PRAGMA';", string(sql))
}

func TestOfflineSQLRejectsPrivilegedEquivalents(t *testing.T) {
	dir, err := ioutil.TempDir("", "randgen-sql-policy")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	existing := filepath.Join(dir, "existing.sql")
	const original = "SELECT 1;"
	require.NoError(t, ioutil.WriteFile(existing, []byte(original), 0600))
	output := filepath.Join(dir, "new.sql")
	for _, statement := range []string{
		"SELECT * FROM mysql.user",
		"SELECT * FROM pg_catalog.pg_authid",
		"SELECT set_config",
		"CREATE TABLE sample (id INT) ENGINE='FEDERATED'",
	} {
		err := writeSQLArtifacts(
			sqlArtifact{path: existing, sql: "SELECT 2;"},
			sqlArtifact{path: output, sql: statement},
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "blocked by safety policy")
		content, err := ioutil.ReadFile(existing)
		require.NoError(t, err)
		assert.Equal(t, original, string(content))
		_, err = os.Stat(output)
		assert.True(t, os.IsNotExist(err))
	}
}

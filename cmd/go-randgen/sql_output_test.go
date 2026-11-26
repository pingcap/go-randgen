package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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
		"SET @first = 1, GLOBAL policy_test_option = DEFAULT",
		"SET @first = 1, @@global.policy_test_option = DEFAULT",
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

func TestSQLArtifactsRejectClientSyntax(t *testing.T) {
	dir, err := ioutil.TempDir("", "randgen-client-syntax")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	existing := filepath.Join(dir, "existing.sql")
	output := filepath.Join(dir, "new.sql")
	const original = "SELECT 1;"
	require.NoError(t, ioutil.WriteFile(existing, []byte(original), 0600))

	statements := []string{
		`\p`, `SELECT 1;\p`, `SELECT '\p'`, `SELECT "\p"`, "SELECT `\\p`",
		`SELECT $$\p$$`, "-- \\p\nSELECT 1", `/* \p */ SELECT 1`,
		`/*+ \p */ SELECT 1`, `SELECT 'escaped\nvalue'`, `SELECT 'C:\data'`,
		".policy_test", " \t.policy_test", "SELECT 1;\n.policy_test",
		"SELECT 1;\r\n \t.policy_test", "SELECT 1;\n\v\f.policy_test",
		"SELECT 'first\n.policy_test'", "/*\n.policy_test\n*/ SELECT 1",
	}
	for _, command := range []string{
		"system", "source", "pager", "tee", "connect", "edit", "delimiter", "charset", "ssl_session_data_print",
	} {
		for _, prefix := range []string{"", " \t", "SELECT 1;", "SELECT 1;\n", "SELECT 1;\r\n", "SELECT 1;\v\f"} {
			statements = append(statements, prefix+command+" inert_token")
		}
		statements = append(statements, strings.ToUpper(command)+";", "SELECT '"+command+"'", "/* "+command+" */ SELECT 1")
	}
	for _, statement := range statements {
		t.Run(statement, func(t *testing.T) {
			err := writeSQLArtifacts(sqlArtifact{path: existing, sql: "SELECT 2;"}, sqlArtifact{path: output, sql: statement})
			require.Error(t, err)
			assert.Contains(t, err.Error(), "SQL artifact blocked by safety policy")
			assert.NotContains(t, err.Error(), "inert_token")
			content, err := ioutil.ReadFile(existing)
			require.NoError(t, err)
			assert.Equal(t, original, string(content))
			_, err = os.Stat(output)
			assert.True(t, os.IsNotExist(err))
		})
	}
	require.NoError(t, writeSQLArtifacts(sqlArtifact{path: output, sql: "SELECT .5, system_value, source_id, 'teetime' FROM test_table;"}))
}

func TestOfflineSQLCommandsRejectClientSyntax(t *testing.T) {
	dir, err := ioutil.TempDir("", "randgen-client-output")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	grammar := filepath.Join(dir, "test.yy")
	for _, statement := range []string{`SELECT "\p"`, "source inert_token", ".policy_test"} {
		require.NoError(t, ioutil.WriteFile(grammar, []byte("query:\n    "+statement+"\n"), 0600))
		for _, tc := range []struct {
			name     string
			args     []string
			suffixes []string
		}{
			{"combined", []string{"gentest"}, []string{".sql"}},
			{"split", []string{"gentest", "-B"}, []string{".data.sql", ".rand.sql"}},
			{"skip-zz", []string{"gentest", "-B", "--skip-zz"}, []string{".rand.sql"}},
			{"postgres", []string{"gentest", "-B", "--skip-zz", "--dbms", "postgres"}, []string{".rand.sql"}},
			{"metadata", []string{"gensql", "--dbms", "sqlite3", "--dsn", ":memory:"}, []string{".rand.sql"}},
		} {
			t.Run(statement+"/"+tc.name, func(t *testing.T) {
				output := filepath.Join(dir, tc.name)
				for _, suffix := range tc.suffixes {
					require.NoError(t, ioutil.WriteFile(output+suffix, []byte("SELECT 1;"), 0600))
				}
				reInitCmd()
				_, err := executeCommand(rootCmd, append(tc.args, "-Y", grammar, "-Q", "1", "-O", output)...)
				require.Error(t, err)
				assert.Contains(t, err.Error(), "SQL artifact blocked by safety policy")
				for _, suffix := range tc.suffixes {
					content, err := ioutil.ReadFile(output + suffix)
					require.NoError(t, err)
					assert.Equal(t, "SELECT 1;", string(content))
				}
			})
		}
	}
}

package compare

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateSQLSafetyQuoteEscapes(t *testing.T) {
	for _, statement := range []string{
		`SELECT 'safe\'; ATTACH DATABASE ':memory:' AS injected; -- '`,
		`SELECT 'safe\'; COPY t TO PROGRAM 'command'; -- '`,
		`SELECT 'safe\'; GRANT ALL ON t TO attacker; -- '`,
		`SELECT 'safe\'; CREATE ROLE attacker; -- '`,
		`SELECT 'safe\'; CREATE SCHEMA injected; -- '`,
		`SELECT 'safe\'; DROP DATABASE production; -- '`,
		`SELECT 'safe\'; SELECT pg_read_file('secret'); -- '`,
		`SELECT 'safe\'; SELECT load_extension('extension'); -- '`,
		`SELECT 'safe\'; SELECT secret INTO OUTFILE 'leak' FROM t; -- '`,
		// SQLite permits a column named e followed by a single-quoted alias.
		`SELECT e'safe\' FROM t; ATTACH DATABASE ':memory:' AS injected; -- '`,
		// With backslash escapes, the second quote starts the hidden SQL tail.
		`SELECT 'safe\' AS value, '; GRANT ALL ON t TO attacker; -- '`,
		`SELECT E'safe\' AS value, '; SELECT pg_read_file('secret'); -- '`,
		// PostgreSQL can mix ordinary strings and E-strings in one statement.
		`SELECT 'prefix\', E'safe\' AS value, '; SELECT pg_read_file('secret'); -- '`,
	} {
		t.Run(statement, func(t *testing.T) {
			err := validateSQLSafety(statement)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "safety policy")
		})
	}
}

func TestSkipSQLQuoted(t *testing.T) {
	for _, quote := range []byte{'\'', '"', '`'} {
		q := string(quote)
		for _, tc := range []struct {
			statement        string
			backslashEscapes bool
			end              int
		}{
			{q + "a\\" + q + "b" + q + " tail", false, 4},
			{q + "a\\" + q + "b" + q + " tail", true, 6},
			{q + "a" + q + q + "b" + q + " tail", false, 6},
			{q + "a" + q + q + "b" + q + " tail", true, 6},
			{q + "a\\\\" + q + " tail", false, 5},
			{q + "a\\\\" + q + " tail", true, 5},
			{q + "unfinished", false, 11},
			{q + "unfinished", true, 11},
		} {
			assert.Equal(t, tc.end, skipSQLQuoted(tc.statement, 0, quote, tc.backslashEscapes),
				"statement=%q backslashEscapes=%t", tc.statement, tc.backslashEscapes)
		}
	}
}

func TestDatabaseExecutionRejectsQuoteBypass(t *testing.T) {
	for _, statement := range []string{
		`SELECT 'safe\'; ATTACH DATABASE ':memory:' AS injected; -- '`,
		`SELECT e'safe\' FROM t; ATTACH DATABASE ':memory:' AS injected; -- '`,
	} {
		// SQLite really executes the tail when the statement reaches its driver.
		direct, err := sql.Open("sqlite3", ":memory:")
		require.NoError(t, err)
		defer direct.Close()
		_, err = direct.Exec("CREATE TABLE t(e INTEGER)")
		require.NoError(t, err)
		direct.SetMaxOpenConns(1)
		_, err = direct.Exec(statement)
		require.NoError(t, err)
		var value string
		require.NoError(t, direct.QueryRow("SELECT name FROM pragma_database_list WHERE name = 'injected'").Scan(&value))
		assert.Equal(t, "injected", value)

		for _, entry := range []string{"query", "exec", "ExecSqlsInDbs"} {
			t.Run(entry, func(t *testing.T) {
				db, err := sql.Open("sqlite3", ":memory:")
				require.NoError(t, err)
				defer db.Close()
				_, err = db.Exec("CREATE TABLE t(e INTEGER)")
				require.NoError(t, err)
				db.SetMaxOpenConns(1)
				switch entry {
				case "query":
					_, err = query(db, statement)
				case "exec":
					_, err = exec(db, statement)
				case "ExecSqlsInDbs":
					_, err = ExecSqlsInDbs([]string{statement}, db)
				}
				require.Error(t, err)
				assert.Contains(t, err.Error(), "safety policy")
				var count int
				require.NoError(t, db.QueryRow("SELECT count(*) FROM pragma_database_list WHERE name = 'injected'").Scan(&count))
				assert.Zero(t, count)
			})
		}
	}
}

func TestValidateSQLSafetyAllowsTableWorkloads(t *testing.T) {
	for _, statement := range []string{
		"SELECT 'INTO OUTFILE is text'",
		`SELECT 'it''s ordinary text'`,
		`SELECT 'it\'s ordinary text'`,
		`SELECT E'it\'s ordinary text'`,
		`SELECT 'C:\temporary\data'`,
		`SELECT 'ends in two backslashes\\'`,
		"CREATE TABLE test.t (id BIGINT)",
		"CREATE TEMPORARY TABLE test.t (id BIGINT)",
		"CREATE UNIQUE INDEX idx_t_id ON test.t (id)",
		"CREATE OR REPLACE VIEW test.v AS SELECT id FROM test.t",
		"DROP TABLE test.t",
		"INSERT INTO test.t VALUES (1)",
		"UPDATE test.t SET id = 2",
		"DELETE FROM test.t",
		"SET @value = 1",
	} {
		assert.NoError(t, validateSQLSafety(statement), statement)
	}
}

func TestDatabaseExecutionRejectsUnsafeSQL(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	_, err = query(db, "DROP DATABASE production")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "safety policy")

	_, err = exec(db, "GRANT ALL ON *.* TO attacker")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "safety policy")

	_, err = ExecSqlsInDbs([]string{"SELECT LOAD_FILE('/run/secrets/value')"}, db)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "safety policy")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestValidateSQLSafetyBlocksPrivilegedOperations(t *testing.T) {
	for _, statement := range []string{
		"GRANT ALL ON *.* TO attacker",
		"REVOKE ALL ON *.* FROM attacker",
		"CREATE USER attacker IDENTIFIED BY 'x'",
		"DROP DATABASE production",
		"/* ordinary comment */ DROP DATABASE production",
		"/*!50000 DROP DATABASE production */",
		"/*M!100100 DROP DATABASE production */",
		"SET GLOBAL general_log = ON",
		"RENAME USER old TO new",
		"SET PASSWORD FOR user = 'x'",
		"SET ROLE admin",
		"SET SESSION AUTHORIZATION admin",
		"CREATE SERVER external FOREIGN DATA WRAPPER mysql OPTIONS (HOST 'remote')",
		"CREATE OR REPLACE FUNCTION dangerous() RETURNS void AS 'x' LANGUAGE c",
		"CREATE DEFINER = CURRENT_USER PROCEDURE dangerous() SELECT 1",
		"CREATE IF NOT EXISTS USER attacker IDENTIFIED BY 'x'",
		"ALTER SERVER external OPTIONS (SET HOST 'remote')",
		"DROP SERVER external",
		"SELECT secret INTO OUTFILE '/tmp/leak' FROM credentials",
		"SELECT LOAD_FILE('/run/secrets/randgen-dsn')",
		"LOAD DATA LOCAL INFILE '/tmp/input' INTO TABLE test.t",
		"COPY test.t TO PROGRAM 'curl attacker'",
		"ATTACH DATABASE '/run/secrets/value' AS stolen",
		"SELECT pg_read_file('/run/secrets/value')",
		`SELECT "pg_read_file"('/run/secrets/value')`,
		`SELECT U&"pg\005Fread\005Ffile"('/run/secrets/value')`,
		"SELECT `load_file`('/run/secrets/value')",
		"SELECT load_extension('/tmp/malicious.so')",
		"ALTER SYSTEM SET shared_preload_libraries = 'malicious'",
		"INSTALL PLUGIN malicious SONAME 'malicious.so'",
		"PREPARE stmt FROM 'DROP DATABASE production'",
		"EXECUTE stmt",
		"CALL privileged_procedure()",
		`SELECT 'safe\'; DROP DATABASE production`,
		`SELECT 'safe\'; SELECT pg_read_file('/run/secrets/value')`,
	} {
		err := validateSQLSafety(statement)
		if assert.Error(t, err, statement) {
			assert.NotContains(t, err.Error(), statement)
		}
	}
}

func TestValidateSQLSafetyCommentRules(t *testing.T) {
	for _, statement := range []string{
		`SELECT --LOAD_FILE('placeholder')`,
		`SELECT 1 # length(pg_read_file('placeholder'))`,
		`SELECT 1 /* unterminated`,
		`SELECT 1 /*!50000 DROP DATABASE placeholder`,
		"SELECT 1 # '\n; GRANT ALL ON t TO placeholder; -- '",
		"SELECT 1 --'\n; COPY t TO PROGRAM 'placeholder'; -- '",
		`/*!50000 SELECT --LOAD_FILE('placeholder') */`,
		`SELECT "#pg_read_file"('placeholder')`,
	} {
		t.Run(statement, func(t *testing.T) {
			err := validateSQLSafety(statement)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "safety policy")
		})
	}
	for _, statement := range []string{
		"SELECT 1 -- ordinary comment\n; SELECT 2",
		"SELECT 1 --\tordinary comment\n; SELECT 2",
		"SELECT 1 # ordinary comment\n; SELECT 2",
		`SELECT 1 /* DROP DATABASE is comment text */`,
		`SELECT '--LOAD_FILE is text', '#pg_read_file is text', '/* text'`,
	} {
		assert.NoError(t, validateSQLSafety(statement), statement)
	}
}

func TestDatabaseExecutionRejectsCommentBypass(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	for _, statement := range []string{
		`SELECT --LOAD_FILE('placeholder')`,
		`SELECT 1 # length(pg_read_file('placeholder'))`,
		`SELECT 1 /* unterminated`,
	} {
		_, err = query(db, statement)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "safety policy")
		_, err = exec(db, statement)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "safety policy")
		_, err = ExecSqlsInDbs([]string{statement}, db)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "safety policy")
	}
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSQLSecurityTokensLineComments(t *testing.T) {
	for _, tc := range []struct {
		statement string
		mysql     bool
		want      []string
	}{
		{"SELECT --LOAD_FILE('placeholder')", true, []string{"SELECT", "LOAD_FILE"}},
		{"SELECT --LOAD_FILE('placeholder')", false, []string{"SELECT"}},
		{"SELECT #pg_read_file('placeholder')", true, []string{"SELECT"}},
		{"SELECT #pg_read_file('placeholder')", false, []string{"SELECT", "PG_READ_FILE"}},
		{"SELECT --", true, []string{"SELECT"}},
		{"SELECT -- comment\rSELECT 2", true, []string{"SELECT", "SELECT", "2"}},
		{"SELECT --\tcomment\nSELECT 2", true, []string{"SELECT", "SELECT", "2"}},
		{"SELECT --\x7fcomment\nSELECT 2", true, []string{"SELECT", "SELECT", "2"}},
		{"SELECT #comment\nSELECT 2", true, []string{"SELECT", "SELECT", "2"}},
	} {
		assert.Equal(t, tc.want, sqlSecurityTokens(tc.statement, sqlQuoteStandard, tc.mysql),
			"statement=%q mysqlComments=%t", tc.statement, tc.mysql)
	}
}

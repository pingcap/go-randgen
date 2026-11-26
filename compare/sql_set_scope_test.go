package compare

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLSafetyChecksWholeSetStatement(t *testing.T) {
	for _, statement := range []string{
		"SET @first = 1, GLOBAL policy_test_option = DEFAULT",
		"SET @first = 1, PERSIST policy_test_option = DEFAULT",
		"SET @first = 1, PERSIST_ONLY policy_test_option = DEFAULT",
		"SET @first = 1, @@global.policy_test_option = DEFAULT",
		"SET @first = 1, SESSION policy_test_option = DEFAULT, GLOBAL other_option = DEFAULT",
		"SELECT 1; SET @first = 1, GLOBAL policy_test_option = DEFAULT",
		"SET LOCAL SESSION AUTHORIZATION",
		"SET LOCAL ROLE",
		"SET @first = 1, /* assignment */ GLOBAL policy_test_option = DEFAULT",
		"SET @first = 1, @@`global`.policy_test_option = DEFAULT",
		"SET @first = \"; SELECT 1\", GLOBAL policy_test_option = DEFAULT",
	} {
		t.Run(statement, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			assert.Error(t, ValidateSQLSafety(statement))
			_, err = query(db, statement)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "blocked by safety policy")
			_, err = exec(db, statement)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "blocked by safety policy")
			_, err = ExecSqlsInDbs([]string{statement}, db)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "blocked by safety policy")
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSQLSafetyAllowsOrdinarySetStatements(t *testing.T) {
	for _, statement := range []string{
		"SET @first = 1, @second = 2",
		"SET SESSION sql_mode = 'ANSI'",
		"SET @@session.sql_mode = 'ANSI'",
		"SET LOCAL search_path = public",
		"SET @value = 'GLOBAL PERSIST PERSIST_ONLY PASSWORD ROLE AUTHORIZATION'",
		"SET @first = 1 /* GLOBAL */; SELECT global FROM sample",
		"SELECT 1; SET @first = 1, @second = 2; SELECT role FROM sample",
		"UPDATE sample SET id = 1, global = 2, role = 3",
		"SET SESSION sql_mode = 'ANSI'; UPDATE sample SET id = 1, global = 2",
		"SET @first = \"; SELECT 1\", @second = 2",
	} {
		assert.NoError(t, ValidateSQLSafety(statement), statement)
	}
}

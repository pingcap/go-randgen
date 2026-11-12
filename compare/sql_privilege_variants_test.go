package compare

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLSafetyRejectsPrivilegeVariants(t *testing.T) {
	for _, statement := range []string{
		"SET SESSION ROLE", "SET LOCAL ROLE", "DROP GROUP",
		"INSTALL COMPONENT", "UNINSTALL COMPONENT",
		"SELECT dblink_connect", "SELECT dblink_connect_u", "SELECT dblink_send_query",
		"SELECT lo_get", "SELECT lo_put", "SELECT pg_ls_waldir",
		"SELECT pg_catalog.\"pg_ls_waldir\"", "SELECT [lo_get]",
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

func TestSQLSafetyAllowsSimilarWorkloads(t *testing.T) {
	for _, statement := range []string{
		"SELECT 'dblink_connect, lo_get, pg_ls_waldir, DROP GROUP'",
		"SELECT dblink_connect_status, lo_get_count FROM test_data",
		"SELECT 1 /* SET SESSION ROLE */",
		"SET SESSION sql_mode = 'ANSI'",
		"DROP TABLE test_group",
	} {
		assert.NoError(t, ValidateSQLSafety(statement), statement)
	}
}

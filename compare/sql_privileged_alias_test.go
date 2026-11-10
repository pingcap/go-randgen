package compare

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPrivilegedAliasesRejectedBeforeDriver(t *testing.T) {
	for _, statement := range []string{
		`ALTER ROLE placeholder SUPERUSER`,
		`SELECT 1; ALTER ROLE placeholder WITH LOGIN SUPERUSER CREATEDB CREATEROLE`,
		`ALTER ROLE placeholder PASSWORD 'placeholder'`,
		`DROP SCHEMA placeholder CASCADE`,
		`ALTER SCHEMA placeholder OWNER TO placeholder`,
		`ALTER DATABASE placeholder SET search_path = placeholder`,
		`ALTER GROUP placeholder ADD USER placeholder`,
		`FLUSH PRIVILEGES`, `FLUSH TABLES WITH READ LOCK`, `KILL CONNECTION 1`,
		`CHANGE MASTER TO MASTER_HOST = 'placeholder'`,
		`CHANGE REPLICATION SOURCE TO SOURCE_HOST = 'placeholder'`, `RESET MASTER`,
		`SELECT pg_ls_dir('placeholder')`, `SELECT pg_stat_file('placeholder')`,
		`SELECT pg_file_write('placeholder', 'placeholder', false)`,
	} {
		t.Run(statement, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()
			require.Error(t, validateSQLSafety(statement))
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
	for _, statement := range []string{
		`ALTER TABLE t ADD COLUMN c INTEGER`,
		`ALTER TABLE t CHANGE old_column new_column INTEGER`, `ALTER INDEX idx RENAME TO other_idx`,
		`DROP TABLE t`, `DROP INDEX idx`, `SELECT 'ALTER ROLE, DROP SCHEMA, FLUSH, KILL are text'`,
		`SELECT 1 /* ALTER ROLE placeholder SUPERUSER */`,
	} {
		assert.NoError(t, validateSQLSafety(statement), statement)
	}
}

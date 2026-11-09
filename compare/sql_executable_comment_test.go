package compare

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestExecutableCommentSafetyEntrypoints(t *testing.T) {
	for _, statement := range []string{
		"SET/*M! a*/GLOBAL general_log = ON",
		"SELECT 1 INTO/*!99999 a*/OUTFILE 'placeholder'",
		"LOAD/*M! a*/DATA LOCAL INFILE 'placeholder' INTO TABLE t",
		"ALTER/*! a*/SYSTEM SET archive_command = 'placeholder'",
		"SET SESSION/*!99999 a*/AUTHORIZATION placeholder",
		"/*!99999 SELECT 1 */",
		"/*M!100100 SELECT 1 */",
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
}

func TestOrdinaryCommentsRemainAllowed(t *testing.T) {
	for _, statement := range []string{`SELECT '/*! marker is string data */'`, `CREATE TABLE t (id INTEGER) /* ordinary comment */`} {
		assert.NoError(t, validateSQLSafety(statement), statement)
	}
}

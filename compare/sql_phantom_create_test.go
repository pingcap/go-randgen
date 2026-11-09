package compare

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPhantomCommentCannotDisableCreatePolicy(t *testing.T) {
	for _, statement := range []string{
		"/*M! x*/CREATE SCHEMA placeholder",
		"/*M! z*/CREATE/*M! a*/USER placeholder",
		"/*!99999 x*/CREATE SCHEMA placeholder",
		"/*!99999 z*/CREATE/*!99999 a*/USER placeholder",
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

func TestCreateNamedColumnsRemainAllowed(t *testing.T) {
	for _, statement := range []string{`CREATE TABLE t ("create" INTEGER)`, `SELECT "create" FROM t`} {
		assert.NoError(t, validateSQLSafety(statement), statement)
	}
}

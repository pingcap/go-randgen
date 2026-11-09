package compare

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDialectBlindLineCommentReportRegression(t *testing.T) {
	for _, statement := range []string{
		"SELECT --LOAD_FILE('placeholder')",
		"SELECT 1 # length(pg_read_file('placeholder')); COPY t TO PROGRAM 'placeholder'",
		"SELECT 1 # '\n; GRANT ALL ON t TO placeholder; -- '",
		"SELECT 1 --'\n; COPY t TO PROGRAM 'placeholder'; -- '",
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

func TestDialectLineCommentCompatibility(t *testing.T) {
	for _, statement := range []string{"SELECT 1 -- comment\n; SELECT 2", "SELECT 1 # comment\n; SELECT 2", `SELECT '#pg_read_file is text', '--LOAD_FILE is text'`} {
		assert.NoError(t, validateSQLSafety(statement), statement)
	}
}

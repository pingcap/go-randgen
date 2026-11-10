package compare

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNestedBlockCommentBypassRejected(t *testing.T) {
	for _, statement := range []string{
		`ALTER/*/* */j*/SYSTEM SET archive_command='placeholder'`,
		`ALTER/*/* */j*/USER placeholder WITH SUPERUSER`,
		`SET/*/* */j*/SESSION AUTHORIZATION placeholder`,
		`/*/* */j*/CREATE OR REPLACE FUNCTION f() RETURNS void AS 'placeholder' LANGUAGE plpgsql`,
		`/*/* */j*/CREATE SCHEMA placeholder`,
		`SELECT 1 /* outer /* inner */`,
		`SELECT 1 /* outer /* inner */ ' */; SELECT pg_read_file('placeholder')`,
		`SELECT 1 /* outer /* inner */; SELECT LOAD_FILE('placeholder'); /* */`,
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
		`SELECT 1 /* outer /* inner */ tail */`,
		`CREATE TABLE t (id INTEGER) /* outer /* inner */ tail */`,
		`SELECT '/* nested /* marker */ is string text'`,
	} {
		assert.NoError(t, validateSQLSafety(statement), statement)
	}
}

func TestBlockCommentDialectBoundaries(t *testing.T) {
	const statement = `SELECT 1 /* outer /* inner */ phantom */; SELECT 2`
	for _, mode := range []sqlQuoteMode{sqlQuotePostgres, sqlQuotePostgresBackslash} {
		assert.Equal(t, []string{"SELECT", "1", ";", "SELECT", "2"}, sqlSecurityTokens(statement, mode, false))
	}
	assert.Equal(t, []string{"SELECT", "1", "PHANTOM", ";", "SELECT", "2"}, sqlSecurityTokens(statement, sqlQuoteStandard, false))
	for _, tc := range []struct {
		statement string
		nested    bool
		end       int
	}{
		{`/**/tail`, true, 4}, {`/*`, true, -1}, {`/*/`, true, -1},
		{`/*/* */`, true, -1}, {`/*/* */`, false, 7},
		{`/* /* /* */ */ */tail`, true, 17},
	} {
		assert.Equal(t, tc.end, skipSQLBlockComment(tc.statement, 0, tc.nested), tc.statement)
	}
}

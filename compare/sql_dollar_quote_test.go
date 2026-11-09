package compare

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDollarQuoteSafetyEntrypoints(t *testing.T) {
	for _, statement := range []string{
		"SELECT $$'$$, pg_read_file('placeholder')",
		"SELECT $tag$'$tag$; COPY t TO PROGRAM 'placeholder'; -- '",
		"SELECT $$'$$; CREATE SCHEMA placeholder; -- '",
		"SELECT 'prefix\\', $$'$$, pg_read_file('placeholder')",
		"SELECT $Tag$unterminated$tag$",
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

func TestDollarQuoteWorkloads(t *testing.T) {
	for _, statement := range []string{
		`SELECT $$ordinary text$$`,
		`SELECT $tag$single ' and double " quotes$tag$`,
		`SELECT $tag$other $nested$ markers$tag$`,
		`SELECT $1, value$tag FROM t`,
	} {
		assert.NoError(t, validateSQLSafety(statement), statement)
	}
}

func TestSQLSecurityTokensDollarQuotes(t *testing.T) {
	for _, mode := range []sqlQuoteMode{sqlQuotePostgres, sqlQuotePostgresBackslash} {
		for _, tc := range []struct {
			statement string
			want      []string
		}{
			{`SELECT $$'; COPY ignored; --$$; SELECT 1`, []string{"SELECT", ";", "SELECT", "1"}},
			{`SELECT $Tag_1$other $tag$ ' " \\ $Tag_1$; SELECT 2`, []string{"SELECT", ";", "SELECT", "2"}},
			{`SELECT $中文$' $中文$; SELECT 3`, []string{"SELECT", ";", "SELECT", "3"}},
			{`SELECT $Tag$body$tag$`, []string{"SELECT", "UNTERMINATED_DOLLAR_QUOTE"}},
			{`SELECT $1, column$tag FROM t`, []string{"SELECT", "$1", "COLUMN$TAG", "FROM", "T"}},
		} {
			assert.Equal(t, tc.want, sqlSecurityTokens(tc.statement, mode, false), tc.statement)
		}
	}
}

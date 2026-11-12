package compare

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSQLSecurityTokensLineCommentNewlines(t *testing.T) {
	for _, tc := range []struct {
		name      string
		statement string
		mode      sqlQuoteMode
		mysql     bool
		want      []string
	}{
		{"mysql hash", "SELECT 1 # comment\r still comment\n; SELECT 2", sqlQuoteBackslash, true,
			[]string{"SELECT", "1", ";", "SELECT", "2"}},
		{"mysql dash", "SELECT 1 -- comment\r still comment\n; SELECT 2", sqlQuoteStandard, true,
			[]string{"SELECT", "1", ";", "SELECT", "2"}},
		{"mysql eof", "SELECT 1 # comment\r still comment", sqlQuoteBackslash, true,
			[]string{"SELECT", "1"}},
		{"sqlite dash", "SELECT 1 --comment\r still comment\n; SELECT 2", sqlQuoteSQLite, false,
			[]string{"SELECT", "1", ";", "SELECT", "2"}},
		{"postgres dash", "SELECT 1 --comment\r; SELECT 2", sqlQuotePostgres, false,
			[]string{"SELECT", "1", ";", "SELECT", "2"}},
		{"mysql CRLF", "SELECT 1 -- comment\r\n; SELECT 2", sqlQuoteStandard, true,
			[]string{"SELECT", "1", ";", "SELECT", "2"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, sqlSecurityTokens(tc.statement, tc.mode, tc.mysql))
		})
	}
}

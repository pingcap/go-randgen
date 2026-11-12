package compare

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSQLSafetySQLiteBracketIdentifiers(t *testing.T) {
	assert.NoError(t, validateSQLSafety("SELECT [reader's note] FROM [samples]"))
	assert.NoError(t, validateSQLSafety("SELECT ARRAY[1, 2]"))
	assert.NoError(t, validateSQLSafety("SELECT \"left[bracket\""))
	assert.NoError(t, validateSQLSafety("SELECT $$left[bracket$$"))
}

func TestSQLSecurityTokensSQLiteBracketIdentifiers(t *testing.T) {
	assert.Equal(t, []string{"SELECT", "[]", "FROM", "[]", ";", "SELECT", "1"},
		sqlSecurityTokens("SELECT [reader's ; note] FROM [samples]; SELECT 1", sqlQuoteSQLite, false))
	assert.Equal(t, []string{"SELECT", "[]", "AS", "LABEL"},
		sqlSecurityTokens("SELECT [value]] AS label", sqlQuoteSQLite, false))
	assert.Equal(t, []string{"SELECT", "LOAD_EXTENSION"},
		sqlSecurityTokens("SELECT [load_extension]", sqlQuoteSQLite, false))
	assert.Equal(t, []string{"SELECT", "[]", ";", "SELECT", "1"},
		sqlSecurityTokens("SELECT [value]; -- comment\r still a comment\nSELECT 1", sqlQuoteSQLite, false))
	assert.Equal(t, []string{"SELECT", "1", ";", "SELECT"},
		sqlSecurityTokens("SELECT 1; SELECT [unfinished", sqlQuoteSQLite, false))
}

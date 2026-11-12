package compare

import (
	"fmt"
	"strings"
)

var blockedSQLKeywords = map[string]bool{
	"ATTACH":                    true,
	"BACKUP":                    true,
	"CALL":                      true,
	"COPY":                      true,
	"FLUSH":                     true,
	"KILL":                      true,
	"RESET":                     true,
	"DETACH":                    true,
	"DO":                        true,
	"EXECUTE":                   true,
	"GRANT":                     true,
	"PREPARE":                   true,
	"PRAGMA":                    true,
	"RESTORE":                   true,
	"REVOKE":                    true,
	"SHUTDOWN":                  true,
	"USE":                       true,
	"VACUUM":                    true,
	"UNICODE_QUOTED_IDENTIFIER": true,
	"UNTERMINATED_COMMENT":      true,
	"UNTERMINATED_DOLLAR_QUOTE": true,
	"EXECUTABLE_COMMENT":        true,
}

var blockedSQLPairs = map[string]bool{
	"CHANGE MASTER":      true,
	"CHANGE REPLICATION": true,
	"ALTER EVENT":        true,
	"ALTER ROLE":         true,
	"ALTER SCHEMA":       true,
	"ALTER DATABASE":     true,
	"ALTER GROUP":        true,
	"ALTER FUNCTION":     true,
	"ALTER INSTANCE":     true,
	"ALTER PROCEDURE":    true,
	"ALTER SERVER":       true,
	"ALTER SYSTEM":       true,
	"ALTER USER":         true,
	"CREATE DATABASE":    true,
	"CREATE EVENT":       true,
	"CREATE EXTENSION":   true,
	"CREATE FUNCTION":    true,
	"CREATE LANGUAGE":    true,
	"CREATE PROCEDURE":   true,
	"CREATE ROLE":        true,
	"CREATE SERVER":      true,
	"CREATE TRIGGER":     true,
	"CREATE USER":        true,
	"DROP DATABASE":      true,
	"DROP SCHEMA":        true,
	"DROP EVENT":         true,
	"DROP EXTENSION":     true,
	"DROP FUNCTION":      true,
	"DROP LANGUAGE":      true,
	"DROP PROCEDURE":     true,
	"DROP ROLE":          true,
	"DROP SERVER":        true,
	"DROP TRIGGER":       true,
	"DROP USER":          true,
	"INSTALL PLUGIN":     true,
	"INTO DUMPFILE":      true,
	"INTO OUTFILE":       true,
	"LOAD DATA":          true,
	"LOAD XML":           true,
	"RENAME USER":        true,
	"SET GLOBAL":         true,
	"SET PASSWORD":       true,
	"SET PERSIST":        true,
	"SET PERSIST_ONLY":   true,
	"SET ROLE":           true,
	"UNINSTALL PLUGIN":   true,
}

var blockedSQLTriples = map[string]bool{
	"SET SESSION AUTHORIZATION": true,
}

var blockedSQLFunctions = map[string]bool{
	"DBLINK":              true,
	"DBLINK_EXEC":         true,
	"LOAD_EXTENSION":      true,
	"LOAD_FILE":           true,
	"LO_EXPORT":           true,
	"LO_IMPORT":           true,
	"PG_READ_BINARY_FILE": true,
	"PG_READ_FILE":        true,
	"PG_LS_DIR":           true,
	"PG_STAT_FILE":        true,
	"PG_FILE_WRITE":       true,
	"PG_WRITE_FILE":       true,
	"SYS_EVAL":            true,
	"SYS_EXEC":            true,
}

type sqlQuoteMode int

const (
	sqlQuoteStandard sqlQuoteMode = iota
	sqlQuoteBackslash
	sqlQuotePostgres
	sqlQuotePostgresBackslash
)

func validateSQLSafety(statement string) error {
	// Drivers disagree on quotes and line comments; every reading must be safe.
	for _, mode := range []sqlQuoteMode{sqlQuoteStandard, sqlQuoteBackslash, sqlQuotePostgres, sqlQuotePostgresBackslash} {
		for _, mysqlComments := range []bool{false, true} {
			tokens := sqlSecurityTokens(statement, mode, mysqlComments)
			for i, token := range tokens {
				if token == "CREATE" && (i == 0 || tokens[i-1] == ";") && !allowedCreateStatement(tokens[i:]) {
					return fmt.Errorf("SQL statement blocked by safety policy: CREATE operation")
				}
				if blockedSQLKeywords[token] || blockedSQLFunctions[token] {
					return fmt.Errorf("SQL statement blocked by safety policy: %s operation", token)
				}
				if i+1 < len(tokens) {
					pair := token + " " + tokens[i+1]
					if blockedSQLPairs[pair] {
						return fmt.Errorf("SQL statement blocked by safety policy: %s operation", pair)
					}
				}
				if i+2 < len(tokens) {
					triple := token + " " + tokens[i+1] + " " + tokens[i+2]
					if blockedSQLTriples[triple] {
						return fmt.Errorf("SQL statement blocked by safety policy: %s operation", triple)
					}
				}
			}
		}
	}
	return nil
}

func allowedCreateStatement(tokens []string) bool {
	allowedPrefixes := [][]string{
		{"CREATE", "TABLE"},
		{"CREATE", "TEMPORARY", "TABLE"},
		{"CREATE", "INDEX"},
		{"CREATE", "UNIQUE", "INDEX"},
		{"CREATE", "VIEW"},
		{"CREATE", "OR", "REPLACE", "VIEW"},
	}
	for _, prefix := range allowedPrefixes {
		if len(tokens) >= len(prefix) {
			matched := true
			for i := range prefix {
				if tokens[i] != prefix[i] {
					matched = false
					break
				}
			}
			if matched {
				return true
			}
		}
	}
	return false
}

// sqlSecurityTokens returns unquoted SQL words. Ordinary comments and quoted
// values are ignored; version-dependent executable comments are rejected.
func sqlSecurityTokens(statement string, mode sqlQuoteMode, mysqlComments bool) []string {
	var tokens []string
	for offset := 0; offset < len(statement); {
		if (statement[offset] == 'U' || statement[offset] == 'u') && offset+2 < len(statement) &&
			statement[offset+1] == '&' && statement[offset+2] == '"' {
			tokens = append(tokens, "UNICODE_QUOTED_IDENTIFIER")
			offset = skipSQLQuoted(statement, offset+2, '"', mode == sqlQuoteBackslash)
			continue
		}
		switch statement[offset] {
		case '\'':
			// PostgreSQL E-strings always use backslash escapes, even beside ordinary strings.
			escapeString := (mode == sqlQuotePostgres || mode == sqlQuotePostgresBackslash) && offset > 0 && (statement[offset-1] == 'E' || statement[offset-1] == 'e') &&
				(offset == 1 || !isSQLWordByte(statement[offset-2]))
			offset = skipSQLQuoted(statement, offset, statement[offset], mode == sqlQuoteBackslash || mode == sqlQuotePostgresBackslash || escapeString)
		case '"', '`':
			end := skipSQLQuoted(statement, offset, statement[offset], mode == sqlQuoteBackslash)
			if end > offset+1 {
				tokens = append(tokens, sqlSecurityTokens(statement[offset+1:end-1], mode, mysqlComments)...)
			}
			offset = end
		case ';':
			tokens = append(tokens, ";")
			offset++
		case '#':
			if mysqlComments {
				offset = skipSQLLine(statement, offset+1)
			} else {
				offset++
			}
		case '-':
			if offset+1 < len(statement) && statement[offset+1] == '-' &&
				(!mysqlComments || offset+2 < len(statement) &&
					(statement[offset+2] <= ' ' || statement[offset+2] == 0x7f)) {
				offset = skipSQLLine(statement, offset+2)
			} else {
				offset++
			}
		case '/':
			if offset+1 < len(statement) && statement[offset+1] == '*' {
				end := skipSQLBlockComment(statement, offset, mode == sqlQuotePostgres || mode == sqlQuotePostgresBackslash)
				if end < 0 {
					return append(tokens, "UNTERMINATED_COMMENT")
				}
				contentStart := offset + 2
				contentEnd := end - 2
				// Servers may execute or ignore the body depending on flavour and version.
				if contentStart < contentEnd && (statement[contentStart] == '!' ||
					contentStart+1 < contentEnd && (statement[contentStart] == 'M' || statement[contentStart] == 'm') && statement[contentStart+1] == '!') {
					return append(tokens, "EXECUTABLE_COMMENT")
				}
				offset = end
			} else {
				offset++
			}
		case '$':
			if mode == sqlQuotePostgres || mode == sqlQuotePostgresBackslash {
				if delimiter := sqlDollarQuoteDelimiter(statement[offset:]); delimiter != "" {
					bodyStart := offset + len(delimiter)
					end := strings.Index(statement[bodyStart:], delimiter)
					if end < 0 {
						return append(tokens, "UNTERMINATED_DOLLAR_QUOTE")
					}
					offset = bodyStart + end + len(delimiter)
					continue
				}
			}
			fallthrough
		default:
			if isSQLWordByte(statement[offset]) {
				start := offset
				for offset < len(statement) && isSQLWordByte(statement[offset]) {
					offset++
				}
				tokens = append(tokens, strings.ToUpper(statement[start:offset]))
			} else {
				offset++
			}
		}
	}
	return tokens
}

func skipSQLQuoted(statement string, offset int, quote byte, backslashEscapes bool) int {
	for offset++; offset < len(statement); offset++ {
		if backslashEscapes && statement[offset] == '\\' {
			offset++
			continue
		}
		if statement[offset] == quote {
			if offset+1 < len(statement) && statement[offset+1] == quote {
				offset++
				continue
			}
			return offset + 1
		}
	}
	return len(statement)
}

func skipSQLLine(statement string, offset int) int {
	for offset < len(statement) && statement[offset] != '\n' && statement[offset] != '\r' {
		offset++
	}
	return offset
}

func isSQLWordByte(char byte) bool {
	return char == '_' || char == '$' || char >= 'a' && char <= 'z' ||
		char >= 'A' && char <= 'Z' || char >= '0' && char <= '9'
}

// Dollar-quote tags follow identifier rules, excluding '$' within the tag.
func sqlDollarQuoteDelimiter(statement string) string {
	for i := 1; i < len(statement); i++ {
		c := statement[i]
		if c == '$' {
			return statement[:i+1]
		}
		if !(c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' ||
			c >= 0x80 || i > 1 && c >= '0' && c <= '9') {
			break
		}
	}
	return ""
}

// PostgreSQL nests block comments; other dialects stop at the first terminator.
func skipSQLBlockComment(statement string, offset int, nested bool) int {
	depth := 1
	for offset += 2; offset+1 < len(statement); {
		if statement[offset] == '/' && statement[offset+1] == '*' && nested {
			depth++
			offset += 2
		} else if statement[offset] == '*' && statement[offset+1] == '/' {
			depth--
			offset += 2
			if depth == 0 {
				return offset
			}
		} else {
			offset++
		}
	}
	return -1
}

// ValidateSQLSafety applies the SQL execution policy to generated SQL, including scripts.
func ValidateSQLSafety(statement string) error {
	return validateSQLSafety(statement)
}

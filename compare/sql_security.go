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
}

var blockedSQLPairs = map[string]bool{
	"ALTER EVENT":      true,
	"ALTER FUNCTION":   true,
	"ALTER INSTANCE":   true,
	"ALTER PROCEDURE":  true,
	"ALTER SERVER":     true,
	"ALTER SYSTEM":     true,
	"ALTER USER":       true,
	"CREATE DATABASE":  true,
	"CREATE EVENT":     true,
	"CREATE EXTENSION": true,
	"CREATE FUNCTION":  true,
	"CREATE LANGUAGE":  true,
	"CREATE PROCEDURE": true,
	"CREATE ROLE":      true,
	"CREATE SERVER":    true,
	"CREATE TRIGGER":   true,
	"CREATE USER":      true,
	"DROP DATABASE":    true,
	"DROP EVENT":       true,
	"DROP EXTENSION":   true,
	"DROP FUNCTION":    true,
	"DROP LANGUAGE":    true,
	"DROP PROCEDURE":   true,
	"DROP ROLE":        true,
	"DROP SERVER":      true,
	"DROP TRIGGER":     true,
	"DROP USER":        true,
	"INSTALL PLUGIN":   true,
	"INTO DUMPFILE":    true,
	"INTO OUTFILE":     true,
	"LOAD DATA":        true,
	"LOAD XML":         true,
	"RENAME USER":      true,
	"SET GLOBAL":       true,
	"SET PASSWORD":     true,
	"SET PERSIST":      true,
	"SET PERSIST_ONLY": true,
	"SET ROLE":         true,
	"UNINSTALL PLUGIN": true,
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
	"PG_WRITE_FILE":       true,
	"SYS_EVAL":            true,
	"SYS_EXEC":            true,
}

func validateSQLSafety(statement string) error {
	tokens := sqlSecurityTokens(statement)
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
// values are ignored; MySQL executable comments are scanned because the server
// executes their contents.
func sqlSecurityTokens(statement string) []string {
	var tokens []string
	for offset := 0; offset < len(statement); {
		if (statement[offset] == 'U' || statement[offset] == 'u') && offset+2 < len(statement) &&
			statement[offset+1] == '&' && statement[offset+2] == '"' {
			tokens = append(tokens, "UNICODE_QUOTED_IDENTIFIER")
			offset = skipSQLQuoted(statement, offset+2, '"')
			continue
		}
		switch statement[offset] {
		case '\'':
			offset = skipSQLQuoted(statement, offset, statement[offset])
		case '"', '`':
			end := skipSQLQuoted(statement, offset, statement[offset])
			if end > offset+1 {
				tokens = append(tokens, sqlSecurityTokens(statement[offset+1:end-1])...)
			}
			offset = end
		case ';':
			tokens = append(tokens, ";")
			offset++
		case '#':
			offset = skipSQLLine(statement, offset+1)
		case '-':
			if offset+1 < len(statement) && statement[offset+1] == '-' {
				offset = skipSQLLine(statement, offset+2)
			} else {
				offset++
			}
		case '/':
			if offset+1 < len(statement) && statement[offset+1] == '*' {
				end := strings.Index(statement[offset+2:], "*/")
				if end < 0 {
					return tokens
				}
				contentStart := offset + 2
				contentEnd := contentStart + end
				if contentStart < contentEnd && statement[contentStart] == '!' {
					contentStart++
				} else if contentStart+1 < contentEnd && (statement[contentStart] == 'M' || statement[contentStart] == 'm') && statement[contentStart+1] == '!' {
					contentStart += 2
				} else {
					offset = contentEnd + 2
					continue
				}
				{
					for contentStart < contentEnd && statement[contentStart] >= '0' && statement[contentStart] <= '9' {
						contentStart++
					}
					tokens = append(tokens, sqlSecurityTokens(statement[contentStart:contentEnd])...)
				}
				offset = contentEnd + 2
			} else {
				offset++
			}
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

func skipSQLQuoted(statement string, offset int, quote byte) int {
	for offset++; offset < len(statement); offset++ {
		if quote != '\'' && statement[offset] == '\\' {
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

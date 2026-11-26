package compare

import (
	"fmt"
	"strings"
)

var allowedSQLStatements = sqlNameSet(
	"SELECT WITH VALUES INSERT UPDATE DELETE REPLACE CREATE ALTER DROP TRUNCATE " +
		"SET BEGIN START END COMMIT ROLLBACK SAVEPOINT RELEASE EXPLAIN ANALYZE SHOW DESCRIBE DESC")

var allowedSQLSettings = sqlNameSet(
	"SQL_MODE AUTOCOMMIT TIME_ZONE NAMES CHARACTER_SET_CLIENT CHARACTER_SET_CONNECTION " +
		"CHARACTER_SET_RESULTS TRANSACTION TRANSACTION_ISOLATION TRANSACTION_READ_ONLY")

var allowedSQLCalls = sqlNameSet(
	"ABS ACOS ASIN ATAN ATAN2 AVG CEIL CEILING COALESCE CONCAT CONCAT_WS COS COT COUNT " +
		"EXP FLOOR GREATEST LEAST LN LOG LOG10 LOG2 MAX MIN MOD NULLIF IF IFNULL POW POWER " +
		"RAND RANDOM ROUND SIGN SIN SQRT SUM TAN TRUNCATE " +
		"ASCII CHAR_LENGTH CHARACTER_LENGTH LENGTH OCTET_LENGTH LOWER UPPER LCASE UCASE " +
		"LEFT RIGHT SUBSTR SUBSTRING TRIM LTRIM RTRIM REPLACE REVERSE LPAD RPAD " +
		"DATE TIME DATETIME TIMESTAMP NOW CURDATE CURTIME CURRENT_DATE CURRENT_TIME CURRENT_TIMESTAMP " +
		"YEAR MONTH DAY HOUR MINUTE SECOND DATE_ADD DATE_SUB DATEDIFF EXTRACT STRFTIME " +
		"CAST CONVERT HEX UNHEX JSON_EXTRACT JSON_ARRAY JSON_OBJECT " +
		"ROW_NUMBER RANK DENSE_RANK CUME_DIST PERCENT_RANK NTILE LEAD LAG FIRST_VALUE LAST_VALUE NTH_VALUE " +
		"XMLCOMMENT XMLCONCAT XMLELEMENT XMLFOREST XMLPARSE XMLPI XMLROOT XMLSERIALIZE " +
		"XML_IS_WELL_FORMED XML_IS_WELL_FORMED_DOCUMENT XML_IS_WELL_FORMED_CONTENT XPATH XPATH_EXISTS " +
		"CHAR BINARY")

var allowedSQLTypes = sqlNameSet(
	"INT INTEGER BIGINT SMALLINT TINYINT MEDIUMINT DECIMAL NUMERIC FLOAT DOUBLE REAL " +
		"CHAR VARCHAR BINARY VARBINARY TEXT BLOB ENUM SET BIT BOOL BOOLEAN")

func sqlNameSet(names string) map[string]bool {
	set := make(map[string]bool)
	for _, name := range strings.Fields(names) {
		set[name] = true
	}
	return set
}

func validateSQLAuthorization(tokens []string) error {
	statementStart := 0
	for i, token := range tokens {
		if i == 0 || tokens[i-1] == ";" {
			statementStart = i
			if token != ";" && !allowedSQLStatements[token] && !sqlNumericToken(token) {
				return fmt.Errorf("SQL statement blocked by safety policy: unapproved statement")
			}
			if token == "SET" {
				if err := validateSQLSettings(tokens[i+1:]); err != nil {
					return err
				}
			}
		}
		if i+1 < len(tokens) && tokens[i+1] == "(" &&
			(len(token) > 0 && (isSQLWordByte(token[0]) || token[0] >= 0x80 || token[0] == '"' || token[0] == 0)) &&
			!allowedSQLCall(tokens, i, statementStart) {
			return fmt.Errorf("SQL statement blocked by safety policy: unapproved function")
		}
	}
	return nil
}

func validateSQLSettings(tokens []string) error {
	start, depth := 0, 0
	for i := 0; i <= len(tokens); i++ {
		if i == len(tokens) || tokens[i] == ";" || tokens[i] == "," && depth == 0 {
			if !allowedSQLSetting(tokens[start:i]) {
				return fmt.Errorf("SQL statement blocked by safety policy: unapproved session setting")
			}
			if i == len(tokens) || tokens[i] == ";" {
				return nil
			}
			start = i + 1
			continue
		}
		switch tokens[i] {
		case "(":
			depth++
		case ")":
			depth--
		}
	}
	return nil
}

func allowedSQLSetting(tokens []string) bool {
	if len(tokens) == 0 {
		return false
	}
	if tokens[0] == "@" {
		if len(tokens) < 2 {
			return false
		}
		if tokens[1] != "@" {
			return len(tokens) > 2 && tokens[2] == "="
		}
		tokens = tokens[2:]
	}
	if len(tokens) > 0 && (tokens[0] == "SESSION" || tokens[0] == "LOCAL") {
		tokens = tokens[1:]
		if len(tokens) > 0 && tokens[0] == "." {
			tokens = tokens[1:]
		}
	}
	if len(tokens) >= 2 && tokens[0] == "TIME" && tokens[1] == "ZONE" {
		return true
	}
	return len(tokens) > 0 && allowedSQLSettings[tokens[0]]
}

func sqlNumericToken(token string) bool {
	if token == "" {
		return false
	}
	for _, char := range token {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func allowedSQLCall(tokens []string, index int, statementStart int) bool {
	name := tokens[index]
	previous := index - 1
	qualified := previous >= 0 && tokens[previous] == "."
	for previous >= 1 && tokens[previous] == "." {
		previous -= 2
	}
	// Table names and CTE column lists are syntax, not function calls.
	if previous >= 0 {
		switch tokens[previous] {
		case "TABLE", "INTO", "REFERENCES":
			return true
		case "EXISTS":
			if previous >= 3 && tokens[previous-3] == "TABLE" &&
				tokens[previous-2] == "IF" && tokens[previous-1] == "NOT" {
				return true
			}
		case "ON":
			if statementStart+2 < len(tokens) && tokens[statementStart] == "CREATE" &&
				(tokens[statementStart+1] == "INDEX" ||
					tokens[statementStart+1] == "UNIQUE" && tokens[statementStart+2] == "INDEX") {
				return true
			}
		}
	}
	if !qualified && sqlCTEColumnList(tokens, index, statementStart) {
		return true
	}
	if previous >= 0 && tokens[previous] == "KEY" &&
		sqlTableDefinitionPosition(tokens, index, statementStart, 1) &&
		(sqlConstraintStart(tokens, previous) ||
			previous > 0 && tokens[previous-1] == "UNIQUE" && sqlConstraintStart(tokens, previous-1)) {
		return true
	}
	if qualified || strings.HasPrefix(name, "\"") || strings.HasPrefix(name, "\x00") {
		return false
	}
	if allowedSQLCalls[name] {
		return true
	}
	switch name {
	case "SELECT", "WHERE", "HAVING", "AND", "OR", "WHEN", "THEN", "ELSE", "ON",
		"AS", "IN", "EXISTS", "NOT", "VALUES", "ROW":
		return true
	case "OVER":
		return previous >= 0 && tokens[previous] == ")"
	case "HASH", "RANGE", "LIST":
		return previous >= 1 && tokens[previous] == "BY" &&
			(tokens[previous-1] == "PARTITION" || tokens[previous-1] == "SUBPARTITION") &&
			sqlTableDefinitionPosition(tokens, index, statementStart, 0)
	}
	if allowedSQLTypes[name] && previous >= 0 && tokens[previous] == "AS" {
		return true
	}
	if !sqlTableDefinitionPosition(tokens, index, statementStart, 1) {
		return false
	}
	if allowedSQLTypes[name] {
		return previous >= 1 && (tokens[previous-1] == "(" || tokens[previous-1] == "," ||
			tokens[previous-1] == "ADD" || tokens[previous-1] == "MODIFY")
	}
	switch name {
	case "CHECK", "UNIQUE", "KEY":
		return sqlConstraintStart(tokens, index) ||
			previous >= 0 && (tokens[previous] == "PRIMARY" || tokens[previous] == "UNIQUE" ||
				tokens[previous] == "FOREIGN") && sqlConstraintStart(tokens, previous)
	}
	return false
}

func sqlCTEColumnList(tokens []string, index, statementStart int) bool {
	if tokens[statementStart] != "WITH" || index <= statementStart {
		return false
	}
	previous := tokens[index-1]
	if previous != "WITH" && previous != "RECURSIVE" && previous != "," {
		return false
	}
	depth := 0
	for _, token := range tokens[statementStart+1 : index] {
		if depth == 0 && allowedSQLStatements[token] {
			return false
		}
		switch token {
		case "(":
			depth++
		case ")":
			depth--
		}
	}
	return depth == 0
}

func sqlTableDefinitionPosition(tokens []string, index, start, createDepth int) bool {
	cursor := start + 1
	if cursor < len(tokens) && tokens[cursor] == "TEMPORARY" {
		cursor++
	}
	if cursor >= len(tokens) || tokens[cursor] != "TABLE" {
		return false
	}
	depth := 0
	for _, token := range tokens[start:index] {
		switch token {
		case "SELECT":
			return false
		case "(":
			depth++
		case ")":
			depth--
		}
	}
	return tokens[start] == "CREATE" && depth == createDepth || tokens[start] == "ALTER" && depth == 0
}

func sqlConstraintStart(tokens []string, index int) bool {
	previous := index - 1
	if previous >= 1 && tokens[previous-1] == "CONSTRAINT" {
		previous -= 2
	}
	return previous >= 0 && (tokens[previous] == "(" || tokens[previous] == "," ||
		tokens[previous] == "ADD")
}

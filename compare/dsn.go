package compare

import (
	"net/url"
	"strings"

	"github.com/go-sql-driver/mysql"
)

const (
	redactedPassword = "******"
	redactedDSN      = "<redacted-dsn>"
)

// RedactDSN removes passwords from supported database connection strings and
// fails closed for formats it cannot recognize safely.
func RedactDSN(dsn string) string {
	if dsn == "" {
		return ""
	}
	if safeDSN, ok := redactPostgresURL(dsn); ok {
		return safeDSN
	}
	if safeDSN, ok := redactMySQLDSN(dsn); ok {
		return safeDSN
	}
	if safeDSN, ok := redactPostgresKeywords(dsn); ok {
		return safeDSN
	}
	return redactedDSN
}

func redactPostgresURL(dsn string) (string, bool) {
	u, err := url.Parse(dsn)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") {
		return "", false
	}
	if u.User != nil {
		if _, hasPassword := u.User.Password(); hasPassword {
			u.User = url.UserPassword(u.User.Username(), redactedPassword)
		}
	}
	query := u.Query()
	for key := range query {
		if strings.EqualFold(key, "password") || strings.EqualFold(key, "passwd") {
			query.Set(key, redactedPassword)
		}
	}
	u.RawQuery = query.Encode()
	safeDSN := u.String()
	if separator := strings.LastIndex(safeDSN, "@"); separator >= 0 {
		userinfo := strings.ReplaceAll(safeDSN[:separator], url.QueryEscape(redactedPassword), redactedPassword)
		safeDSN = userinfo + safeDSN[separator:]
	}
	return safeDSN, true
}

func redactMySQLDSN(dsn string) (string, bool) {
	separator := strings.LastIndex(dsn, "@")
	if separator < 0 {
		return "", false
	}
	config, err := mysql.ParseDSN(dsn)
	if err != nil {
		return "", false
	}
	if strings.Contains(dsn[:separator], ":") {
		config.Passwd = redactedPassword
	}
	return config.FormatDSN(), true
}

type passwordRange struct {
	start  int
	end    int
	quoted bool
}

func redactPostgresKeywords(dsn string) (string, bool) {
	ranges, ok := postgresPasswordRanges(dsn)
	if !ok || len(ranges) == 0 {
		return "", false
	}
	var result strings.Builder
	last := 0
	for _, password := range ranges {
		result.WriteString(dsn[last:password.start])
		if password.quoted {
			result.WriteString("'")
		}
		result.WriteString(redactedPassword)
		if password.quoted {
			result.WriteString("'")
		}
		last = password.end
	}
	result.WriteString(dsn[last:])
	return result.String(), true
}

func postgresPasswordRanges(dsn string) ([]passwordRange, bool) {
	var ranges []passwordRange
	for offset := 0; ; {
		for offset < len(dsn) && isDSNSpace(dsn[offset]) {
			offset++
		}
		if offset == len(dsn) {
			return ranges, true
		}
		keyStart := offset
		for offset < len(dsn) && !isDSNSpace(dsn[offset]) && dsn[offset] != '=' {
			offset++
		}
		keyEnd := offset
		if keyStart == keyEnd {
			return nil, false
		}
		for offset < len(dsn) && isDSNSpace(dsn[offset]) {
			offset++
		}
		if offset == len(dsn) || dsn[offset] != '=' {
			return nil, false
		}
		offset++
		for offset < len(dsn) && isDSNSpace(dsn[offset]) {
			offset++
		}

		valueStart := offset
		quoted := offset < len(dsn) && dsn[offset] == '\''
		if quoted {
			offset++
			closed := false
			for offset < len(dsn) {
				switch dsn[offset] {
				case '\\':
					if offset+1 == len(dsn) {
						return nil, false
					}
					offset += 2
				case '\'':
					offset++
					closed = true
				default:
					offset++
				}
				if closed {
					break
				}
			}
			if !closed || (offset < len(dsn) && !isDSNSpace(dsn[offset])) {
				return nil, false
			}
		} else {
			for offset < len(dsn) && !isDSNSpace(dsn[offset]) {
				if dsn[offset] == '\\' {
					if offset+1 == len(dsn) {
						return nil, false
					}
					offset += 2
				} else {
					offset++
				}
			}
		}
		if strings.EqualFold(dsn[keyStart:keyEnd], "password") {
			ranges = append(ranges, passwordRange{start: valueStart, end: offset, quoted: quoted})
		}
	}
}

func isDSNSpace(char byte) bool {
	switch char {
	case ' ', '\t', '\n', '\r', '\v', '\f':
		return true
	default:
		return false
	}
}

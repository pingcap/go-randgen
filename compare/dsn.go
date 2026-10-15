package compare

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/go-sql-driver/mysql"
)

const (
	redactedPassword = "******"
	redactedDSN      = "<redacted-dsn>"
)

var postgresPasswordPattern = regexp.MustCompile(`(?i)(^|\s)(password\s*=\s*)(?:'([^']|'')*'|"([^"]|"")*"|[^\s]+)`)

// RedactDSN removes passwords from supported database connection strings. It
// fails closed when the input cannot be recognized safely.
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
	if postgresPasswordPattern.MatchString(dsn) {
		return postgresPasswordPattern.ReplaceAllString(dsn, `${1}${2}'`+redactedPassword+`'`)
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

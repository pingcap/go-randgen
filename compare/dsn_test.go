package compare

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactDSN(t *testing.T) {
	tests := []struct {
		name   string
		dsn    string
		want   string
		secret string
	}{
		{
			name: "empty",
			dsn:  "",
			want: "",
		},
		{
			name:   "mysql",
			dsn:    "root:secret@tcp(db.example.com:4000)/test",
			want:   "root:******@tcp(db.example.com:4000)/test",
			secret: "secret",
		},
		{
			name:   "postgres URL",
			dsn:    "postgres://alice:secret@db.example.com/app?sslmode=disable",
			want:   "postgres://alice:******@db.example.com/app?sslmode=disable",
			secret: "secret",
		},
		{
			name:   "postgres URL query password",
			dsn:    "postgres://alice@db.example.com/app?password=secret&sslmode=disable",
			want:   "postgres://alice@db.example.com/app?password=%2A%2A%2A%2A%2A%2A&sslmode=disable",
			secret: "secret",
		},
		{
			name:   "postgres keywords",
			dsn:    "user=alice password='top secret' host=db.example.com dbname=app",
			want:   "user=alice password='******' host=db.example.com dbname=app",
			secret: "top secret",
		},
		{
			name:   "postgres keywords with escaped space",
			dsn:    "user=alice password=top\\ secret host=db.example.com dbname=app",
			want:   "user=alice password=****** host=db.example.com dbname=app",
			secret: "secret",
		},
		{
			name:   "postgres keywords with escaped quote",
			dsn:    "user=alice password='foo\\'bar' host=db.example.com dbname=app",
			want:   "user=alice password='******' host=db.example.com dbname=app",
			secret: "bar",
		},
		{
			name:   "malformed postgres keywords",
			dsn:    "user=alice password='unterminated",
			want:   "<redacted-dsn>",
			secret: "unterminated",
		},
		{
			name: "unknown",
			dsn:  "opaque-secret-value",
			want: "<redacted-dsn>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RedactDSN(tt.dsn)
			assert.Equal(t, tt.want, got)
			if tt.secret != "" {
				assert.NotContains(t, got, tt.secret)
			}
		})
	}
}

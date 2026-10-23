package compare

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactDSN(t *testing.T) {
	tests := []struct {
		name       string
		dsn        string
		want       string
		leakMarker string
	}{
		{name: "empty", dsn: "", want: ""},
		{
			name:       "mysql",
			dsn:        "test-user:credential-leak-marker@tcp(db.example.com:4000)/test",
			want:       "test-user:******@tcp(db.example.com:4000)/test",
			leakMarker: "credential-leak-marker",
		},
		{
			name:       "postgres URL",
			dsn:        "postgres://test-user:credential-leak-marker@db.example.com/app?sslmode=disable",
			want:       "postgres://test-user:******@db.example.com/app?sslmode=disable",
			leakMarker: "credential-leak-marker",
		},
		{
			name:       "postgres keywords",
			dsn:        "user=test-user password='credential leak marker' host=db.example.com dbname=app",
			want:       "user=test-user password='******' host=db.example.com dbname=app",
			leakMarker: "credential leak marker",
		},
		{
			name:       "malformed value fails closed",
			dsn:        "opaque-credential-leak-marker",
			want:       "<redacted-dsn>",
			leakMarker: "credential-leak-marker",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RedactDSN(tt.dsn)
			assert.Equal(t, tt.want, got)
			if tt.leakMarker != "" {
				assert.NotContains(t, got, tt.leakMarker)
			}
		})
	}
}

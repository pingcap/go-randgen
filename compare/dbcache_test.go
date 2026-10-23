package compare

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDbCacheUsesNonReversibleKey(t *testing.T) {
	const dsn = "opaque-credential-leak-marker"
	const expectedKey = "2dc35229ba7c370dea16e05fc308875daef37aa1d84b0f352c2374fca76ba3dd"

	got := dbCacheKey(dsn)

	assert.Equal(t, expectedKey, got)
	assert.NotEqual(t, dsn, got)
}

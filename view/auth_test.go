package view

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGraphRequiresConfiguredToken(t *testing.T) {
	_, err := Graph("query: SELECT 1")
	require.Error(t, err)
}

func TestGraphChecksAuthorization(t *testing.T) {
	handler, err := Graph("query: SELECT 1", "test-token")
	require.NoError(t, err)
	for _, tc := range []struct {
		method string
		auth   string
		status int
	}{
		{"GET", "", 401},
		{"GET", "Bearer incorrect", 401},
		{"GET", "Bearer test-token", 200},
		{"POST", "Bearer test-token", 405},
	} {
		request := httptest.NewRequest(tc.method, "/graph", nil)
		request.Header.Set("Authorization", tc.auth)
		response := httptest.NewRecorder()
		handler(response, request)
		require.Equal(t, tc.status, response.Code)
		require.Empty(t, response.Header().Get("Access-Control-Allow-Origin"))
		if tc.status != 200 {
			require.NotContains(t, response.Body.String(), "SELECT")
		}
	}
}

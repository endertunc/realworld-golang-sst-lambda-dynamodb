package test

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"testing"
)

// SharedAuthenticationTestConfig holds the configuration for an endpoint's authentication tests
type SharedAuthenticationTestConfig struct {
	Method string
	Path   string
}

// RunAuthenticationTests runs a standard suite of authentication tests for an endpoint
// These basic two test cases should suffice to prove that endpoint is secured.
// More detailed test cases around token validation are implemented as unit tests
func RunAuthenticationTests(t *testing.T, config SharedAuthenticationTestConfig) {
	t.Run("without_token", func(t *testing.T) {
		respBody := ExecuteRequest[errutil.SimpleError](t, config.Method, config.Path, nil, http.StatusUnauthorized, nil)
		require.Equal(t, "authorization header is missing", respBody.Message)
	})

	t.Run("with_invalid_token", func(t *testing.T) {
		invalidToken := "invalid.token.here"
		respBody := ExecuteRequest[errutil.SimpleError](t, config.Method, config.Path, nil, http.StatusUnauthorized, &invalidToken)
		require.Equal(t, "invalid token", respBody.Message)
	})
}

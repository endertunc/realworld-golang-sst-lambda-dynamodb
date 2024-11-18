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

	// ToDo @ender this should make request with authorization header with malformed token such as "Basic $token"
	t.Run("with_malformed_token", func(t *testing.T) {
		t.Skip("this should be updated to make request with unsupported token type")
		malformedToken := "not-even-a-jwt-token"
		respBody := ExecuteRequest[errutil.SimpleError](t, config.Method, config.Path, nil, http.StatusUnauthorized, &malformedToken)
		//require.Equal(t, "invalid token", respBody.Message)
		require.Equal(t, "invalid token type", respBody.Message)
	})

	t.Run("with_empty_token", func(t *testing.T) {
		t.Skip("not sure if we should even care about this" +
			"at the end of the day, empty token is an invalid token...")
		emptyToken := ""
		respBody := ExecuteRequest[errutil.SimpleError](t, config.Method, config.Path, nil, http.StatusUnauthorized, &emptyToken)
		require.Equal(t, "bearer token is empty", respBody.Message)
	})
}

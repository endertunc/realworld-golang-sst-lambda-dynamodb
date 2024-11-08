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
		var respBody errutil.SimpleError
		MakeRequestAndParseResponse(t, nil, config.Method, config.Path, http.StatusUnauthorized, &respBody)
		require.Equal(t, "authorization header is missing", respBody.Message)
	})

	t.Run("with_invalid_token", func(t *testing.T) {
		invalidToken := "invalid.token.here"
		var respBody errutil.SimpleError
		MakeAuthenticatedRequestAndParseResponse(t, nil, config.Method, config.Path, http.StatusUnauthorized, &respBody, invalidToken)
		require.Equal(t, "invalid token", respBody.Message)
	})

	// ToDo @ender this should make request with authorization header with malformed token such as "Basic $token"
	t.Run("with_malformed_token", func(t *testing.T) {
		t.Skip("this should be updated to make request with unsupported token type")
		malformedToken := "not-even-a-jwt-token"
		var respBody errutil.SimpleError
		MakeAuthenticatedRequestAndParseResponse(t, nil, config.Method, config.Path, http.StatusUnauthorized, &respBody, malformedToken)
		//require.Equal(t, "invalid token", respBody.Message)
		require.Equal(t, "invalid token type", respBody.Message)
	})

	t.Run("with_empty_token", func(t *testing.T) {
		t.Skip("not sure if we should even care about this" +
			"at the end of the day, empty token is an invalid token...")
		emptyToken := ""
		var respBody errutil.SimpleError
		MakeAuthenticatedRequestAndParseResponse(t, nil, config.Method, config.Path, http.StatusUnauthorized, &respBody, emptyToken)
		require.Equal(t, "bearer token is empty", respBody.Message)
	})
}

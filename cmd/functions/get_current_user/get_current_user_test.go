package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/test"
	"testing"
)

func TestSuccessfulGetCurrentUser(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// First create a user
		user := test.DefaultNewUserRequestUserDto
		test.CreateUserEntity(t, user)

		// Login to get the token
		loginReqBody := dto.LoginRequestBodyDTO{
			User: dto.LoginRequestUserDto{
				Email:    user.Email,
				Password: user.Password,
			},
		}
		var loginRespBody dto.UserResponseBodyDTO
		test.MakeRequestAndParseResponse(t, loginReqBody, "POST", "/api/users/login", http.StatusOK, &loginRespBody)
		assert.NotEmpty(t, loginRespBody.User.Token)
		token := loginRespBody.User.Token

		// Now test get current user endpoint
		var respBody dto.UserResponseBodyDTO
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "GET", "/api/user", http.StatusOK, &respBody, token)

		// Verify the response
		t.Logf("response body: %+v", respBody)
		assert.Equal(t, user.Email, respBody.User.Email)
		assert.Equal(t, user.Username, respBody.User.Username)
		assert.NotEmpty(t, respBody.User.Token)
	})
}

// ToDo next two following tests are are shared among all the API endpoints that require authentication
//
//	I need to come up with a way to share these tests...
func TestGetCurrentUserWithInvalidToken(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		invalidToken := "invalid.token.here"

		var respBody errutil.GenericError
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "GET", "/api/user", http.StatusUnauthorized, &respBody, invalidToken)

		assert.Equal(t, "invalid token", respBody.Message)
	})
}

func TestGetCurrentUserWithoutToken(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		var respBody errutil.GenericError
		test.MakeRequestAndParseResponse(t, nil, "GET", "/api/user", http.StatusUnauthorized, &respBody)

		assert.Equal(t, "authorization header is missing", respBody.Message)
	})
}

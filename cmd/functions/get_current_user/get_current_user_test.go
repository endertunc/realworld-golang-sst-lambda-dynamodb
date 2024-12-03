package main

import (
	"github.com/stretchr/testify/assert"
	dtogen "realworld-aws-lambda-dynamodb-golang/internal/domain/dto/generator"
	"realworld-aws-lambda-dynamodb-golang/internal/test"
	"testing"
)

func TestAuthenticationScenarios(t *testing.T) {
	test.RunAuthenticationTests(t, test.SharedAuthenticationTestConfig{
		Method: "GET",
		Path:   "/api/user",
	})
}

func TestSuccessfulGetCurrentUser(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		userReq := dtogen.GenerateNewUserRequestUserDto()
		_, token := test.CreateAndLoginUser(t, userReq)

		// Now test get current user endpoint
		respBody := test.GetCurrentUser(t, token)

		// Verify the response
		assert.Equal(t, userReq.Email, respBody.Email)
		assert.Equal(t, userReq.Username, respBody.Username)
		assert.NotEmpty(t, respBody.Token)
	})
}

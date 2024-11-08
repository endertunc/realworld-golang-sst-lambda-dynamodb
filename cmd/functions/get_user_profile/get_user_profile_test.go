package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/test"
	"testing"
)

// ToDo @ender - we should add use cases where the user is following the profile user

func TestAnonymousUserFetchExistingProfile(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create a user whose profile we'll fetch
		user := test.DefaultNewUserRequestUserDto
		test.CreateUserEntity(t, user)

		// Fetch the profile without authentication
		var respBody dto.ProfileResponseBodyDTO
		test.MakeRequestAndParseResponse(t, nil, "GET", fmt.Sprintf("/api/profiles/%s", user.Username), http.StatusOK, &respBody)

		// Verify the response
		assert.Equal(t, user.Username, respBody.Profile.Username)
		assert.False(t, respBody.Profile.Following) // Anonymous user shouldn't be following anyone
	})
}

func TestAnonymousUserFetchNonExistingProfile(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		nonExistingUsername := "non-existing-user"

		var respBody errutil.GenericError
		test.MakeRequestAndParseResponse(t, nil, "GET", fmt.Sprintf("/api/profiles/%s", nonExistingUsername), http.StatusNotFound, &respBody)

		assert.Equal(t, "user not found", respBody.Message)
	})
}

func TestAuthenticatedUserFetchExistingProfile(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create two users: one who will do the fetching (logged in user) and one whose profile will be fetched
		loggedInUser := test.DefaultNewUserRequestUserDto
		test.CreateUserEntity(t, loggedInUser)

		targetUser := dto.NewUserRequestUserDto{
			Username: "target-user",
			Email:    "target@example.com",
			Password: "password123",
		}
		test.CreateUserEntity(t, targetUser)

		// Login to get the token
		loginReqBody := dto.LoginRequestBodyDTO{
			User: dto.LoginRequestUserDto{
				Email:    loggedInUser.Email,
				Password: loggedInUser.Password,
			},
		}
		var loginRespBody dto.UserResponseBodyDTO
		test.MakeRequestAndParseResponse(t, loginReqBody, "POST", "/api/users/login", http.StatusOK, &loginRespBody)
		token := loginRespBody.User.Token

		// Fetch the profile with authentication
		var respBody dto.ProfileResponseBodyDTO
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "GET", fmt.Sprintf("/api/profiles/%s", targetUser.Username), http.StatusOK, &respBody, token)

		// Verify the response
		assert.Equal(t, targetUser.Username, respBody.Profile.Username)
		assert.False(t, respBody.Profile.Following) // User shouldn't be following by default
	})
}

func TestAuthenticatedUserFetchNonExistingProfile(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create and login a user
		loggedInUser := test.DefaultNewUserRequestUserDto
		test.CreateUserEntity(t, loggedInUser)

		loginReqBody := dto.LoginRequestBodyDTO{
			User: dto.LoginRequestUserDto{
				Email:    loggedInUser.Email,
				Password: loggedInUser.Password,
			},
		}
		var loginRespBody dto.UserResponseBodyDTO
		test.MakeRequestAndParseResponse(t, loginReqBody, "POST", "/api/users/login", http.StatusOK, &loginRespBody)
		token := loginRespBody.User.Token

		// Try to fetch a non-existing profile
		nonExistingUsername := "non-existing-user"

		var respBody errutil.GenericError
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "GET", fmt.Sprintf("/api/profiles/%s", nonExistingUsername), http.StatusNotFound, &respBody, token)

		assert.Equal(t, "user not found", respBody.Message)
	})
}

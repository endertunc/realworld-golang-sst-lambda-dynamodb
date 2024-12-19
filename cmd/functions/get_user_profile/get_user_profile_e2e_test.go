package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	dtogen "realworld-aws-lambda-dynamodb-golang/internal/domain/dto/generator"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/test"
	"testing"
)

// ToDo @ender - we should add use cases where the user is following the profile user

func TestAnonymousUserFetchExistingProfile(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create a user whose profile we'll fetch
		user := test.CreateUserEntity(t, dtogen.GenerateNewUserRequestUserDto())

		// Fetch the profile without authentication
		respBody := test.GetUserProfile(t, user.Username, nil)

		// Verify the response
		assert.Equal(t, user.Username, respBody.Profile.Username)
		assert.False(t, respBody.Profile.Following) // Anonymous user shouldn't be following anyone
	})
}

func TestAnonymousUserFetchNonExistingProfile(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		nonExistingUsername := "non-existing-user"
		respBody := test.GetUserProfileWithResponse[errutil.SimpleError](t, nonExistingUsername, nil, http.StatusNotFound)
		assert.Equal(t, "user not found", respBody.Message)
	})
}

func TestAuthenticatedUserFetchExistingProfile(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create two users: one who will do the fetching (logged-in user) and one whose profile will be fetched
		_, token := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())

		targetUser := dtogen.GenerateNewUserRequestUserDto()
		test.CreateUserEntity(t, targetUser)

		// Fetch the profile with authentication
		respBody := test.GetUserProfile(t, targetUser.Username, &token)

		// Verify the response
		assert.Equal(t, targetUser.Username, respBody.Profile.Username)
		assert.False(t, respBody.Profile.Following) // User shouldn't be following by default
	})
}

func TestAuthenticatedUserFetchNonExistingProfile(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create and login a user
		_, token := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())

		// Try to fetch a non-existing profile
		nonExistingUsername := "non-existing-user"

		respBody := test.GetUserProfileWithResponse[errutil.SimpleError](t, nonExistingUsername, &token, http.StatusNotFound)
		assert.Equal(t, "user not found", respBody.Message)
	})
}

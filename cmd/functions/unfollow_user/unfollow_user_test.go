package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	dtogen "realworld-aws-lambda-dynamodb-golang/internal/domain/dto/generator"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/test"
	"testing"
)

func TestAuthenticationScenarios(t *testing.T) {
	test.RunAuthenticationTests(t, test.SharedAuthenticationTestConfig{
		Method: "DELETE",
		Path:   "/api/profiles/someuser/follow", // Use a dummy username as we're only testing auth
	})
}

func TestSuccessfulUnfollow(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create and login the follower user
		_, token := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())

		// Create the user to be followed/unfollowed
		userToUnfollow := dtogen.GenerateNewUserRequestUserDto()
		test.CreateUserEntity(t, userToUnfollow)

		// First follow the user
		followRespBody := test.FollowUser(t, userToUnfollow.Username, token)
		assert.True(t, followRespBody.Following)

		// Now unfollow the user
		unfollowRespBody := test.UnfollowUser(t, userToUnfollow.Username, token)

		// Verify the response
		assert.Equal(t, userToUnfollow.Username, unfollowRespBody.Profile.Username)
		assert.False(t, unfollowRespBody.Profile.Following)

		// Verify the following status by getting the profile
		profileRespBody := test.GetUserProfile(t, userToUnfollow.Username, &token)
		assert.False(t, profileRespBody.Profile.Following)
	})
}

func TestUnfollowNonExistentUser(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create and login a user
		_, token := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())

		// Try to unfollow non-existent user
		nonExistentUsername := "non-existent-user"
		respBody := test.UnfollowUserWithResponse[errutil.SimpleError](t, nonExistentUsername, token, http.StatusNotFound)
		assert.Equal(t, "user not found", respBody.Message)
	})
}

func TestUnfollowYourself(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create and login a user
		user, token := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())

		// Try to unfollow yourself
		respBody := test.UnfollowUserWithResponse[errutil.SimpleError](t, user.Username, token, http.StatusConflict)
		assert.Equal(t, "cannot unfollow yourself", respBody.Message)
	})
}

func TestUnfollowUserYouDontFollow(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create and login the follower user
		_, token := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())

		// Create another user
		userToUnfollow := dtogen.GenerateNewUserRequestUserDto()
		test.CreateUserEntity(t, userToUnfollow)

		// Try to unfollow without following first
		unfollowRespBody := test.UnfollowUser(t, userToUnfollow.Username, token)

		// Verify the response (should still return success, just with following=false)
		assert.Equal(t, userToUnfollow.Username, unfollowRespBody.Profile.Username)
		assert.False(t, unfollowRespBody.Profile.Following)
	})
}

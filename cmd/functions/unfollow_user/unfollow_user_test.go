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

func TestAuthenticationScenarios(t *testing.T) {
	test.RunAuthenticationTests(t, test.SharedAuthenticationTestConfig{
		Method: "DELETE",
		Path:   "/api/profiles/someuser/follow", // Use a dummy username as we're only testing auth
	})
}

func TestSuccessfulUnfollow(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create and login the follower user
		_, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)

		// Create the user to be followed/unfollowed
		userToUnfollow := dto.NewUserRequestUserDto{
			Username: "user-to-unfollow",
			Email:    "followed@example.com",
			Password: "password123",
		}
		test.CreateUserEntity(t, userToUnfollow)

		// First follow the user
		var followRespBody dto.ProfileResponseBodyDTO
		test.MakeAuthenticatedRequestAndParseResponse(
			t, nil, "POST",
			fmt.Sprintf("/api/profiles/%s/follow", userToUnfollow.Username),
			http.StatusOK, &followRespBody, token)
		assert.True(t, followRespBody.Profile.Following)

		// Now unfollow the user
		var unfollowRespBody dto.ProfileResponseBodyDTO
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "DELETE",
			fmt.Sprintf("/api/profiles/%s/follow", userToUnfollow.Username),
			http.StatusOK, &unfollowRespBody, token)

		// Verify the response
		assert.Equal(t, userToUnfollow.Username, unfollowRespBody.Profile.Username)
		assert.False(t, unfollowRespBody.Profile.Following)

		// Verify the following status by getting the profile
		var profileRespBody dto.ProfileResponseBodyDTO
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "GET",
			fmt.Sprintf("/api/profiles/%s", userToUnfollow.Username),
			http.StatusOK, &profileRespBody, token)
		assert.False(t, profileRespBody.Profile.Following)
	})
}

func TestUnfollowNonExistentUser(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create and login a user
		_, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)

		// Try to unfollow non-existent user
		nonExistentUsername := "non-existent-user"
		var respBody errutil.GenericError
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "DELETE",
			fmt.Sprintf("/api/profiles/%s/follow", nonExistentUsername),
			http.StatusNotFound, &respBody, token)
		assert.Equal(t, "user not found", respBody.Message)
	})
}

func TestUnfollowYourself(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create and login a user
		user, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)

		// Try to unfollow yourself
		var respBody errutil.GenericError
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "DELETE",
			fmt.Sprintf("/api/profiles/%s/follow", user.Username),
			http.StatusBadRequest, &respBody, token)
		assert.Equal(t, "cannot unfollow yourself", respBody.Message)
	})
}

func TestUnfollowUserYouDontFollow(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create and login the follower user
		_, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)

		// Create another user
		userToUnfollow := dto.NewUserRequestUserDto{
			Username: "user-to-unfollow",
			Email:    "followed@example.com",
			Password: "password123",
		}
		test.CreateUserEntity(t, userToUnfollow)

		// Try to unfollow without following first
		var unfollowRespBody dto.ProfileResponseBodyDTO
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "DELETE",
			fmt.Sprintf("/api/profiles/%s/follow", userToUnfollow.Username),
			http.StatusOK, &unfollowRespBody, token)

		// Verify the response (should still return success, just with following=false)
		assert.Equal(t, userToUnfollow.Username, unfollowRespBody.Profile.Username)
		assert.False(t, unfollowRespBody.Profile.Following)
	})
}

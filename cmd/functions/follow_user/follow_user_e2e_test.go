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
		Method: "POST",
		Path:   "/api/profiles/some-user/follow",
	})
}

func TestSuccessfulFollow(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create and login the follower user
		_, token := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())

		// Create the user to be followed
		userToFollow := dtogen.GenerateNewUserRequestUserDto()
		test.CreateUserEntity(t, userToFollow)

		// Follow the user
		followRespBody := test.FollowUser(t, userToFollow.Username, token)

		// Verify the response
		assert.Equal(t, userToFollow.Username, followRespBody.Username)
		assert.True(t, followRespBody.Following)

		// Verify the following status by getting the profile
		profileRespBody := test.GetUserProfile(t, userToFollow.Username, &token)
		assert.True(t, profileRespBody.Profile.Following)
	})
}

func TestFollowNonExistentUser(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create and login a user
		_, token := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())

		// Try to follow non-existent user
		nonExistentUsername := "non-existent-user"
		respBody := test.FollowUserWithResponse[errutil.SimpleError](t, nonExistentUsername, token, http.StatusNotFound)
		assert.Equal(t, "user not found", respBody.Message)
	})
}

func TestFollowYourself(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create and login a user
		user, token := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())

		// Try to follow yourself
		respBody := test.FollowUserWithResponse[errutil.SimpleError](t, user.Username, token, http.StatusBadRequest)
		assert.Equal(t, "cannot follow yourself", respBody.Message)
	})
}

func TestFollowAlreadyFollowedUser(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// create and login the follower user
		_, token := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())

		// create the user to be followed
		userToFollow := dtogen.GenerateNewUserRequestUserDto()
		test.CreateUserEntity(t, userToFollow)

		// follow the user
		followRespBody := test.FollowUser(t, userToFollow.Username, token)

		// verify the response
		assert.Equal(t, userToFollow.Username, followRespBody.Username)
		assert.True(t, followRespBody.Following)

		// verify the following status by getting the profile
		profileRespBodyOne := test.GetUserProfile(t, userToFollow.Username, &token)
		assert.True(t, profileRespBodyOne.Profile.Following)

		// follow the user again - should be successful
		_ = test.FollowUser(t, userToFollow.Username, token)

		// verify the following status by getting the profile again
		profileRespBodyTwo := test.GetUserProfile(t, userToFollow.Username, &token)
		assert.True(t, profileRespBodyTwo.Profile.Following)

	})
}

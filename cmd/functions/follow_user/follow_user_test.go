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

func TestSuccessfulFollow(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create and login the follower user
		_, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)

		// Create the user to be followed
		userToFollow := dto.NewUserRequestUserDto{
			Username: "user-to-follow",
			Email:    "followed@example.com",
			Password: "password123",
		}
		test.CreateUserEntity(t, userToFollow)

		// Follow the user
		var followRespBody dto.ProfileResponseBodyDTO
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "POST", fmt.Sprintf("/api/profiles/%s/follow", userToFollow.Username), http.StatusOK, &followRespBody, token)

		// Verify the response
		assert.Equal(t, userToFollow.Username, followRespBody.Profile.Username)
		assert.True(t, followRespBody.Profile.Following)

		// Verify the following status by getting the profile

		var profileRespBody dto.ProfileResponseBodyDTO
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "GET", fmt.Sprintf("/api/profiles/%s", userToFollow.Username), http.StatusOK, &profileRespBody, token)
		assert.True(t, profileRespBody.Profile.Following)
	})
}

func TestFollowNonExistentUser(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create and login a user
		_, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)

		// Try to follow non-existent user
		nonExistentUsername := "non-existent-user"
		var respBody errutil.SimpleError
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "POST", fmt.Sprintf("/api/profiles/%s/follow", nonExistentUsername), http.StatusNotFound, &respBody, token)
		assert.Equal(t, "user not found", respBody.Message)
	})
}

func TestFollowYourself(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create and login a user
		user, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)

		// Try to follow yourself
		var respBody errutil.SimpleError
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "POST", fmt.Sprintf("/api/profiles/%s/follow", user.Username), http.StatusBadRequest, &respBody, token)
		assert.Equal(t, "cannot follow yourself", respBody.Message)
	})
}

func TestFollowAlreadyFollowedUser(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// create and login the follower user
		_, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)

		// create the user to be followed
		userToFollow := dto.NewUserRequestUserDto{
			Username: "user-to-follow",
			Email:    "followed@example.com",
			Password: "password123",
		}
		test.CreateUserEntity(t, userToFollow)

		// follow the user
		var followRespBody dto.ProfileResponseBodyDTO
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "POST", fmt.Sprintf("/api/profiles/%s/follow", userToFollow.Username), http.StatusOK, &followRespBody, token)

		// verify the response
		assert.Equal(t, userToFollow.Username, followRespBody.Profile.Username)
		assert.True(t, followRespBody.Profile.Following)

		// verify the following status by getting the profile
		var profileRespBody dto.ProfileResponseBodyDTO
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "GET", fmt.Sprintf("/api/profiles/%s", userToFollow.Username), http.StatusOK, &profileRespBody, token)
		assert.True(t, profileRespBody.Profile.Following)

		// follow the user again
		followRespBody = dto.ProfileResponseBodyDTO{}
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "POST", fmt.Sprintf("/api/profiles/%s/follow", userToFollow.Username), http.StatusOK, &followRespBody, token)

		// verify the following status by getting the profile
		profileRespBody = dto.ProfileResponseBodyDTO{}
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "GET", fmt.Sprintf("/api/profiles/%s", userToFollow.Username), http.StatusOK, &profileRespBody, token)
		assert.True(t, profileRespBody.Profile.Following)

	})
}

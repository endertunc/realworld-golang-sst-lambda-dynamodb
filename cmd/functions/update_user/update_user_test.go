package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	dtogen "realworld-aws-lambda-dynamodb-golang/internal/domain/dto/generator"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/test"
	"testing"
)

func TestAuthenticationScenarios(t *testing.T) {
	test.RunAuthenticationTests(t, test.SharedAuthenticationTestConfig{
		Method: "PUT",
		Path:   "/api/user",
	})
}

func TestSuccessfulUpdateAllFields(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create and login a user first
		_, token := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())

		// Get original user for comparison
		originalUser := test.GetCurrentUser(t, token)

		// Generate update data
		updateReq := dtogen.GenerateUpdateUserRequestUserDTO()

		// Update the user
		updatedUser := test.UpdateUser(t, token, updateReq)

		// Verify the response
		assert.Equal(t, *updateReq.Email, updatedUser.Email)
		assert.Equal(t, *updateReq.Username, updatedUser.Username)
		assert.Equal(t, *updateReq.Bio, *updatedUser.Bio)
		assert.Equal(t, *updateReq.Image, *updatedUser.Image)
		assert.NotEmpty(t, updatedUser.Token)
		assert.NotEqual(t, originalUser.Token, updatedUser.Token)

		// Verify we can login with new credentials
		loginResp := test.LoginUser(t, dto.LoginRequestUserDto{
			Email:    *updateReq.Email,
			Password: *updateReq.Password,
		})
		assert.Equal(t, *updateReq.Email, loginResp.Email)
		assert.Equal(t, *updateReq.Username, loginResp.Username)
	})
}

func TestUpdateWithExistingEmail(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create first user
		firstUser := dtogen.GenerateNewUserRequestUserDto()
		_, _ = test.CreateAndLoginUser(t, firstUser)

		// Create and login second user
		secondUser := dtogen.GenerateNewUserRequestUserDto()
		_, token := test.CreateAndLoginUser(t, secondUser)

		// Get original user for comparison
		originalUser := test.GetCurrentUser(t, token)

		// Try to update second user with first user's email
		updateReq := dto.UpdateUserRequestUserDTO{
			Email: &firstUser.Email,
		}

		// Update should fail
		respErrorBody := test.UpdateUserWithResponse[errutil.SimpleError](t, token, updateReq, http.StatusConflict)
		assert.Equal(t, "email already exists", respErrorBody.Message)

		// Verify original user remains unchanged
		currentUser := test.GetCurrentUser(t, token)
		assert.Equal(t, originalUser.Email, currentUser.Email)
		assert.Equal(t, originalUser.Username, currentUser.Username)
		assert.Equal(t, originalUser.Bio, currentUser.Bio)
		assert.Equal(t, originalUser.Image, currentUser.Image)
	})
}

func TestUpdateWithExistingUsername(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create first user
		firstUser := dtogen.GenerateNewUserRequestUserDto()
		_, _ = test.CreateAndLoginUser(t, firstUser)

		// Create and login second user
		secondUser := dtogen.GenerateNewUserRequestUserDto()
		_, token := test.CreateAndLoginUser(t, secondUser)

		// Get original user for comparison
		originalUser := test.GetCurrentUser(t, token)

		// Try to update second user with first user's username
		updateReq := dto.UpdateUserRequestUserDTO{
			Username: &firstUser.Username,
		}

		// Update should fail
		respErrorBody := test.UpdateUserWithResponse[errutil.SimpleError](t, token, updateReq, http.StatusConflict)
		assert.Equal(t, "username already exists", respErrorBody.Message)

		// Verify original user remains unchanged
		currentUser := test.GetCurrentUser(t, token)
		assert.Equal(t, originalUser.Email, currentUser.Email)
		assert.Equal(t, originalUser.Username, currentUser.Username)
		assert.Equal(t, originalUser.Bio, currentUser.Bio)
		assert.Equal(t, originalUser.Image, currentUser.Image)
	})
}

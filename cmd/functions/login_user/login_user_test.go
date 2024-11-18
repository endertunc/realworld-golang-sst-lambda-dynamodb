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

func TestSuccessfulLogin(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create a user
		user := dtogen.GenerateNewUserRequestUserDto()
		test.CreateUserEntity(t, user)

		// Login the user
		loginRequest := dto.LoginRequestUserDto{
			Email:    user.Email,
			Password: user.Password,
		}
		respBody := test.LoginUser(t, loginRequest)

		// Verify the response
		assert.Equal(t, user.Email, respBody.Email)
		assert.Equal(t, user.Username, respBody.Username)
		assert.NotEmpty(t, respBody.Token)
	})
}

func TestInvalidPassword(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// make sure that at least one user exists
		user := dtogen.GenerateNewUserRequestUserDto()
		test.CreateUserEntity(t, user)

		// Login with invalid password
		reqWithInvalidPassword := dto.LoginRequestUserDto{
			Email:    user.Email,
			Password: "p@sswOrd",
		}
		respBody := test.LoginUserWithResponse[errutil.SimpleError](t, reqWithInvalidPassword, http.StatusUnauthorized)

		// Verify the response
		assert.Equal(t, "invalid credentials", respBody.Message)
	})
}

func TestInvalidEmail(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// make sure that at least one user exists
		user := dtogen.GenerateNewUserRequestUserDto()
		test.CreateUserEntity(t, user)

		// Login with invalid email
		reqWithInvalidEmail := dto.LoginRequestUserDto{
			Email:    "invalid@email.com",
			Password: user.Password,
		}
		respBody := test.LoginUserWithResponse[errutil.SimpleError](t, reqWithInvalidEmail, http.StatusUnauthorized)

		// Verify the response
		assert.Equal(t, "invalid credentials", respBody.Message)
	})
}

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

func TestRequestValidation(t *testing.T) {
	tests := []test.ApiRequestValidationTest[dto.LoginRequestUserDto]{
		{
			Name: "missing email",
			Input: dto.LoginRequestUserDto{
				Password: "password123",
			},
			ExpectedError: map[string]string{
				"User.Email": "Email is a required field",
			},
		},
		{
			Name: "invalid email format",
			Input: dto.LoginRequestUserDto{
				Email:    "invalid-email",
				Password: "password123",
			},
			ExpectedError: map[string]string{
				"User.Email": "Email must be a valid email address",
			},
		},
		{
			Name: "password too short",
			Input: dto.LoginRequestUserDto{
				Email:    "test@example.com",
				Password: "12345",
			},
			ExpectedError: map[string]string{
				"User.Password": "Password must be at least 6 characters in length",
			},
		},
		{
			Name: "password too long",
			Input: dto.LoginRequestUserDto{
				Email:    "test@example.com",
				Password: "123456789012345678901", // 21 characters
			},
			ExpectedError: map[string]string{
				"User.Password": "Password must be a maximum of 20 characters in length",
			},
		},
		{
			Name: "blank password",
			Input: dto.LoginRequestUserDto{
				Email:    "test@example.com",
				Password: "      ",
			},
			ExpectedError: map[string]string{
				"User.Password": "Password cannot be blank",
			},
		},
	}

	loginRequest := func(t *testing.T, input dto.LoginRequestUserDto) errutil.ValidationErrors {
		return test.LoginUserWithResponse[errutil.ValidationErrors](t, input, http.StatusBadRequest)
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			test.TestValidation(t, tt, loginRequest)
		})
	}
}

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

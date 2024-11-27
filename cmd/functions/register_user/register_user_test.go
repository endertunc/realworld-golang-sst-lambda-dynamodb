package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	dtogen "realworld-aws-lambda-dynamodb-golang/internal/domain/dto/generator"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/test"
	"strings"
	"testing"
)

//nolint:golint,exhaustruct
func TestRequestValidation(t *testing.T) {
	tests := []test.ApiRequestValidationTest[dto.NewUserRequestUserDto]{
		{
			Name: "missing email",
			Input: dto.NewUserRequestUserDto{
				Password: "password123",
				Username: "testuser",
			},
			ExpectedError: map[string]string{
				"User.Email": "Email is a required field",
			},
		},
		{
			Name: "invalid email format",
			Input: dto.NewUserRequestUserDto{
				Email:    "invalid-email",
				Password: "password123",
				Username: "testuser",
			},
			ExpectedError: map[string]string{
				"User.Email": "Email must be a valid email address",
			},
		},
		{
			Name: "password too short",
			Input: dto.NewUserRequestUserDto{
				Email:    "test@example.com",
				Password: "12345",
				Username: "testuser",
			},
			ExpectedError: map[string]string{
				"User.Password": "Password must be at least 6 characters in length",
			},
		},
		{
			Name: "password too long",
			Input: dto.NewUserRequestUserDto{
				Email:    "test@example.com",
				Password: "123456789012345678901", // 21 characters
				Username: "testuser",
			},
			ExpectedError: map[string]string{
				"User.Password": "Password must be a maximum of 20 characters in length",
			},
		},
		{
			Name: "username too short",
			Input: dto.NewUserRequestUserDto{
				Email:    "test@example.com",
				Password: "password123",
				Username: "ab",
			},
			ExpectedError: map[string]string{
				"User.Username": "Username must be at least 3 characters in length",
			},
		},
		{
			Name: "username too long",
			Input: dto.NewUserRequestUserDto{
				Email:    "test@example.com",
				Password: "password123",
				Username: strings.Repeat("a", 65),
			},
			ExpectedError: map[string]string{
				"User.Username": "Username must be a maximum of 64 characters in length",
			},
		},
		{
			Name: "blank username",
			Input: dto.NewUserRequestUserDto{
				Email:    "test@example.com",
				Password: "password123",
				Username: "     ",
			},
			ExpectedError: map[string]string{
				"User.Username": "Username cannot be blank",
			},
		},
		{
			Name: "multiple validation errors",
			Input: dto.NewUserRequestUserDto{
				Email:    "invalid-email",
				Password: "12345",
				Username: "ab",
			},
			ExpectedError: map[string]string{
				"User.Email":    "Email must be a valid email address",
				"User.Password": "Password must be at least 6 characters in length",
				"User.Username": "Username must be at least 3 characters in length",
			},
		},
	}

	loginRequest := func(t *testing.T, input dto.NewUserRequestUserDto) errutil.ValidationErrors {
		return test.RegisterUserWithResponse[errutil.ValidationErrors](t, input, http.StatusBadRequest)
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			test.TestValidation(t, tt, loginRequest)
		})
	}
}

func Test_SuccessfulRegister(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		user := dtogen.GenerateNewUserRequestUserDto()
		respBody := test.RegisterUser(t, user)

		assert.Equal(t, user.Email, respBody.Email)
		assert.NotEmpty(t, respBody.Token)
	})
}

func Test_RegisterAlreadyExistsEmail(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		user := dtogen.GenerateNewUserRequestUserDto()
		respBody := test.RegisterUser(t, user)
		assert.NotEmpty(t, respBody)

		userWithSameEmail := user
		// change username but keep the same email
		userWithSameEmail.Username = "test-user-two"

		respErrorBody := test.RegisterUserWithResponse[errutil.SimpleError](t, userWithSameEmail, http.StatusConflict)
		assert.Equal(t, "email already exists", respErrorBody.Message)
	})
}

func Test_RegisterAlreadyExistsUsername(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		user := dtogen.GenerateNewUserRequestUserDto()
		respBody := test.RegisterUser(t, user)
		assert.NotEmpty(t, respBody)

		userWithSameUsername := user
		// change email but keep the same username
		userWithSameUsername.Email = "test-two@example.com"

		respErrorBody := test.RegisterUserWithResponse[errutil.SimpleError](t, userWithSameUsername, http.StatusConflict)
		assert.Equal(t, "username already exists", respErrorBody.Message)
	})
}

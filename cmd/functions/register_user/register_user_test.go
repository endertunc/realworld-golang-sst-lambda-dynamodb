package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	dtogen "realworld-aws-lambda-dynamodb-golang/internal/domain/dto/generator"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/test"
	"testing"
)

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

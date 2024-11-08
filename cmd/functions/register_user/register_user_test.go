package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/test"
	"testing"
)

func Test_SuccessfulRegister(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		reqBody := dto.NewUserRequestBodyDTO{
			User: dto.NewUserRequestUserDto{
				Email:    "test@example.com",
				Username: "test-user",
				Password: "password123",
			},
		}
		var respBody dto.UserResponseBodyDTO
		test.MakeRequestAndParseResponse(t, reqBody, "POST", "/api/users", http.StatusOK, &respBody)
		assert.Equal(t, "test@example.com", respBody.User.Email)
		assert.NotEmpty(t, respBody.User.Token)
	})
}

func Test_RegisterAlreadyExistsEmail(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		reqBodyOne := test.DefaultNewUserRequestBodyDTO
		var respBodyOne dto.UserResponseBodyDTO
		test.MakeRequestAndParseResponse(t, reqBodyOne, "POST", "/api/users", http.StatusOK, &respBodyOne)
		assert.NotEmpty(t, respBodyOne)

		reqBodyTwo := test.DefaultNewUserRequestBodyDTO
		// change username but keep the same email
		reqBodyTwo.User.Username = "test-user-two"

		var respErrorBody errutil.SimpleError
		test.MakeRequestAndParseResponse(t, reqBodyTwo, "POST", "/api/users", http.StatusConflict, &respErrorBody)
		assert.Equal(t, "email already exists", respErrorBody.Message)
	})
}

func Test_RegisterAlreadyExistsUsername(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		reqBodyOne := test.DefaultNewUserRequestBodyDTO
		var respBodyOne dto.UserResponseBodyDTO
		test.MakeRequestAndParseResponse(t, reqBodyOne, "POST", "/api/users", http.StatusOK, &respBodyOne)
		assert.NotEmpty(t, respBodyOne)

		reqBodyTwo := test.DefaultNewUserRequestBodyDTO
		// change email but keep the same username
		reqBodyTwo.User.Email = "test-two@example.com"

		var respErrorBody errutil.SimpleError
		test.MakeRequestAndParseResponse(t, reqBodyTwo, "POST", "/api/users", http.StatusConflict, &respErrorBody)
		assert.Equal(t, "username already exists", respErrorBody.Message)
	})
}

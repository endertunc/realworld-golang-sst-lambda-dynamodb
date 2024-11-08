package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/test"
	"testing"
)

func TestSuccessfulLogin(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		user := test.DefaultNewUserRequestUserDto
		test.CreateUserEntity(t, user)
		reqBody := dto.LoginRequestBodyDTO{
			User: dto.LoginRequestUserDto{
				Email:    user.Email,
				Password: user.Password,
			},
		}

		var respBody dto.UserResponseBodyDTO
		test.MakeRequestAndParseResponse(t, reqBody, "POST", "/api/users/login", http.StatusOK, &respBody)
		assert.Equal(t, user.Email, respBody.User.Email)
		assert.Equal(t, user.Username, respBody.User.Username)
		assert.NotEmpty(t, respBody.User.Token)
	})

}

//func TestInvalidRequestBodyMissingPassword(t *testing.T) {
//	t.Skip()
//	test.WithSetupAndTeardown(t, func() {
//		reqBody := dto.LoginRequestBodyDTO{
//			User: dto.LoginRequestUserDto{
//				Email: "test@example.com",
//			},
//		}
//		test.MakeRequestAndCheckError(t, reqBody, "/api/users/login", http.StatusBadRequest, "error decoding request body")
//	})
//
//}
//
//func TestInvalidRequestBodyMissingEmail(t *testing.T) {
//	t.Skip()
//	test.WithSetupAndTeardown(t, func() {
//		reqBody := dto.LoginRequestBodyDTO{
//			User: dto.LoginRequestUserDto{
//				Password: "123456",
//			},
//		}
//		test.MakeRequestAndCheckError(t, reqBody, "/api/users/login", http.StatusBadRequest, "error decoding request body")
//	})
//
//}

func TestInvalidPassword(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// make sure that at least one user exists
		user := test.DefaultNewUserRequestUserDto
		test.CreateUserEntity(t, user)

		reqWithInvalidPassword := dto.LoginRequestBodyDTO{
			User: dto.LoginRequestUserDto{
				Email:    user.Email,
				Password: "p@sswOrd",
			},
		}
		var respBody errutil.SimpleError
		test.MakeRequestAndParseResponse(t, reqWithInvalidPassword, "POST", "/api/users/login", http.StatusUnauthorized, &respBody)
		assert.Equal(t, "invalid credentials", respBody.Message)
	})
}

func TestInvalidEmail(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// make sure that at least one user exists
		user := test.DefaultNewUserRequestUserDto
		test.CreateUserEntity(t, user)

		reqWithInvalidPassword := dto.LoginRequestBodyDTO{
			User: dto.LoginRequestUserDto{
				Email:    "invalid@email.com",
				Password: user.Password,
			},
		}
		var respBody errutil.SimpleError
		test.MakeRequestAndParseResponse(t, reqWithInvalidPassword, "POST", "/api/users/login", http.StatusUnauthorized, &respBody)
		assert.Equal(t, "invalid credentials", respBody.Message)
	})
}

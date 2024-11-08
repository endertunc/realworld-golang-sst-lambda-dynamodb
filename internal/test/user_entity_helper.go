package test

import (
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"testing"
)

// ToDo @ender I have different file name pattern... this file camelCase, auth_test_suite.go, other with kebab-case....

var DefaultNewUserRequestBodyDTO = dto.NewUserRequestBodyDTO{
	User: DefaultNewUserRequestUserDto,
}

var DefaultNewUserRequestUserDto = dto.NewUserRequestUserDto{
	Email:    "test@example.com",
	Username: "test-user",
	Password: "123456",
}

func CreateUserEntity(t *testing.T, user dto.NewUserRequestUserDto) dto.UserResponseUserDto {
	body := dto.NewUserRequestBodyDTO{User: user}
	var respBody dto.UserResponseBodyDTO
	MakeRequestAndParseResponse(t, body, "POST", "/api/users", http.StatusOK, &respBody)
	return respBody.User
}

package test

import (
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"testing"
)

// ToDo @ender I have different file name pattern... this file camelCase, auth_test_suite.go, other with kebab-case....
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

func GetCurrentUser(t *testing.T, token string) dto.UserResponseUserDto {
	return GetCurrentUserWithResponse[dto.UserResponseBodyDTO](t, token, http.StatusOK).User
}

func GetCurrentUserWithResponse[T interface{}](t *testing.T, token string, expectedStatusCode int) T {
	var respBody T
	MakeAuthenticatedRequestAndParseResponse(t, nil, "GET", "/api/user", expectedStatusCode, &respBody, token)
	return respBody
}

func FollowUser(t *testing.T, username, token string) dto.ProfileResponseDto {
	return FollowUserWithResponse[dto.ProfileResponseBodyDTO](t, username, token, http.StatusOK).Profile
}

func FollowUserWithResponse[T interface{}](t *testing.T, username, token string, expectedStatusCode int) T {
	var respBody T
	MakeAuthenticatedRequestAndParseResponse(t, nil, "POST", "/api/profiles/"+username+"/follow", expectedStatusCode, &respBody, token)
	return respBody
}

func UnfollowUser(t *testing.T, username string, token string) dto.ProfileResponseBodyDTO {
	return UnfollowUserWithResponse[dto.ProfileResponseBodyDTO](t, username, token, http.StatusOK)
}

func UnfollowUserWithResponse[T interface{}](t *testing.T, username string, token string, expectedStatusCode int) T {
	var respBody T
	MakeAuthenticatedRequestAndParseResponse(t, nil, "DELETE", "/api/profiles/"+username+"/follow", expectedStatusCode, &respBody, token)
	return respBody
}

func GetUserProfile(t *testing.T, username string, token *string) dto.ProfileResponseBodyDTO {
	return GetUserProfileWithResponse[dto.ProfileResponseBodyDTO](t, username, token, http.StatusOK)
}

func GetUserProfileWithResponse[T interface{}](t *testing.T, username string, token *string, expectedStatusCode int) T {
	var respBody T
	if token == nil {
		MakeRequestAndParseResponse(t, nil, "GET", "/api/profiles/"+username, expectedStatusCode, &respBody)
	} else {
		MakeAuthenticatedRequestAndParseResponse(t, nil, "GET", "/api/profiles/"+username, expectedStatusCode, &respBody, *token)
	}
	return respBody
}

func RegisterUser(t *testing.T, user dto.NewUserRequestUserDto) dto.UserResponseUserDto {
	return RegisterUserWithResponse[dto.UserResponseBodyDTO](t, user, http.StatusOK).User
}

func RegisterUserWithResponse[T interface{}](t *testing.T, user dto.NewUserRequestUserDto, expectedStatusCode int) T {
	reqBody := dto.NewUserRequestBodyDTO{User: user}
	var respBody T
	MakeRequestAndParseResponse(t, reqBody, "POST", "/api/users", expectedStatusCode, &respBody)
	return respBody
}

package test

import (
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"testing"
)

// CreateAndLoginUser creates a new user and logs them in, returning the user data and authentication token
func CreateAndLoginUser(t *testing.T, user dto.NewUserRequestUserDto) (dto.NewUserRequestUserDto, string) {
	RegisterUser(t, user)
	loginRespBody := LoginUser(t, dto.LoginRequestUserDto{
		Email:    user.Email,
		Password: user.Password,
	})

	return user, loginRespBody.Token
}

func CreateUserEntity(t *testing.T, user dto.NewUserRequestUserDto) dto.UserResponseUserDto {
	reqBody := dto.NewUserRequestBodyDTO{User: user}
	return ExecuteRequest[dto.UserResponseBodyDTO](t, "POST", "/api/users", reqBody, http.StatusOK, nil).User
}

func GetCurrentUser(t *testing.T, token string) dto.UserResponseUserDto {
	return GetCurrentUserWithResponse[dto.UserResponseBodyDTO](t, token, http.StatusOK).User
}

func GetCurrentUserWithResponse[T interface{}](t *testing.T, token string, expectedStatusCode int) T {
	return ExecuteRequest[T](t, "GET", "/api/user", nil, expectedStatusCode, &token)
}

func FollowUser(t *testing.T, username, token string) dto.ProfileResponseDto {
	return FollowUserWithResponse[dto.ProfileResponseBodyDTO](t, username, token, http.StatusOK).Profile
}

func FollowUserWithResponse[T interface{}](t *testing.T, username, token string, expectedStatusCode int) T {
	return ExecuteRequest[T](t, "POST", "/api/profiles/"+username+"/follow", nil, expectedStatusCode, &token)
}

func UnfollowUser(t *testing.T, username string, token string) dto.ProfileResponseBodyDTO {
	return UnfollowUserWithResponse[dto.ProfileResponseBodyDTO](t, username, token, http.StatusOK)
}

func UnfollowUserWithResponse[T interface{}](t *testing.T, username string, token string, expectedStatusCode int) T {
	return ExecuteRequest[T](t, "DELETE", "/api/profiles/"+username+"/follow", nil, expectedStatusCode, &token)
}

func GetUserProfile(t *testing.T, username string, token *string) dto.ProfileResponseBodyDTO {
	return GetUserProfileWithResponse[dto.ProfileResponseBodyDTO](t, username, token, http.StatusOK)
}

func GetUserProfileWithResponse[T interface{}](t *testing.T, username string, token *string, expectedStatusCode int) T {
	return ExecuteRequest[T](t, "GET", "/api/profiles/"+username, nil, expectedStatusCode, token)
}

func RegisterUser(t *testing.T, user dto.NewUserRequestUserDto) dto.UserResponseUserDto {
	return RegisterUserWithResponse[dto.UserResponseBodyDTO](t, user, http.StatusOK).User
}

func RegisterUserWithResponse[T interface{}](t *testing.T, user dto.NewUserRequestUserDto, expectedStatusCode int) T {
	reqBody := dto.NewUserRequestBodyDTO{User: user}
	return ExecuteRequest[T](t, "POST", "/api/users", reqBody, expectedStatusCode, nil)
}

func UpdateUser(t *testing.T, token string, updateReq dto.UpdateUserRequestUserDTO) dto.UserResponseUserDto {
	return UpdateUserWithResponse[dto.UserResponseBodyDTO](t, token, updateReq, http.StatusOK).User
}

func UpdateUserWithResponse[T interface{}](t *testing.T, token string, user dto.UpdateUserRequestUserDTO, expectedStatusCode int) T {
	reqBody := dto.UpdateUserRequestBodyDTO{User: user}
	return ExecuteRequest[T](t, "PUT", "/api/user", reqBody, expectedStatusCode, &token)
}

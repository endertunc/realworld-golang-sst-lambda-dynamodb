package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/steinfletcher/apitest"
	"github.com/stretchr/testify/assert"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
)

// Helper function to create a user
func createUser(t *testing.T, username, email, password string) dto.UserResponseBodyDTO {
	var user dto.UserResponseBodyDTO
	apitest.New().
		EnableNetworking().
		Post("/register").
		JSON(dto.NewUserRequestBodyDTO{
			User: dto.NewUserRequestUserDto{
				Username: username,
				Email:    email,
				Password: password,
			},
		}).
		Expect(t).
		Status(200).
		Assert(func(res *http.Response, req *http.Request) error {
			if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
				return err
			}
			assert.Equal(t, email, user.User.Email)
			assert.Equal(t, username, user.User.Username)
			assert.NotEmpty(t, user.User.Token)
			return nil
		}).
		End()
	return user
}

func TestFollowUserHandler(t *testing.T) {
	// Step 1: Create a user to be followed
	userToBeFollowed := createUser(t, "userToBeFollowed", "followed@example.com", "password123")

	// Step 2: Create another user who will follow the first user
	followerUser := createUser(t, "followerUser", "follower@example.com", "password123")

	// Step 3: Authenticate the second user and follow the first user
	apitest.New().
		EnableNetworking().
		Post("/follow/"+userToBeFollowed.User.Username).
		Header("Authorization", "Token "+followerUser.User.Token).
		Expect(t).
		Status(200).
		Assert(func(res *http.Response, req *http.Request) error {
			var responseBody dto.ProfileResponseBodyDTO
			if err := json.NewDecoder(res.Body).Decode(&responseBody); err != nil {
				return err
			}
			assert.Equal(t, "userToBeFollowed", responseBody.Profile.Username)
			assert.True(t, responseBody.Profile.Following)
			return nil
		}).
		End()
}

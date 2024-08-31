package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/steinfletcher/apitest"
	"github.com/stretchr/testify/assert"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
)

func TestRegisterUserHandler(t *testing.T) {
	// Test Case 1: Successful user registration
	apitest.New().
		EnableNetworking().
		Post("/register").
		JSON(dto.NewUserRequestBodyDTO{
			User: dto.NewUserRequestUserDto{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
		}).
		Expect(t).
		Status(200).
		Assert(func(res *http.Response, req *http.Request) error {
			var responseBody dto.UserResponseBodyDTO
			if err := json.NewDecoder(res.Body).Decode(&responseBody); err != nil {
				return err
			}
			assert.Equal(t, "test@example.com", responseBody.User.Email)
			assert.Equal(t, "testuser", responseBody.User.Username)
			assert.NotEmpty(t, responseBody.User.Token)
			return nil
		}).
		End()

	// Test Case 2: Invalid request body
	apitest.New().
		EnableNetworking().
		Post("/register").
		JSON(dto.NewUserRequestBodyDTO{
			User: dto.NewUserRequestUserDto{
				Email: "test@example.com",
			},
		}).
		Expect(t).
		Status(400).
		Body(`{"message": "error decoding request body"}`).
		End()
}

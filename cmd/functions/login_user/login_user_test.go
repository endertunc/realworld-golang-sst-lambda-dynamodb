package main

import (
	"encoding/json"
	"github.com/steinfletcher/apitest"
	"github.com/stretchr/testify/assert"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"testing"
)

func TestLoginUserHandler(t *testing.T) {
	// Test Case 1: Successful login
	apitest.New().
		EnableNetworking().
		Post("/login").
		JSON(dto.LoginRequestBodyDTO{
			User: dto.LoginRequestUserDto{
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
			assert.NotEmpty(t, responseBody.User.Token)
			return nil
		}).
		End()

	// Test Case 2: Invalid request body
	apitest.New().
		EnableNetworking().
		Post("/login").
		JSON(dto.LoginRequestBodyDTO{
			User: dto.LoginRequestUserDto{
				Email: "test@example.com",
			},
		}).
		Expect(t).
		Status(400).
		Body(`{"message": "error decoding request body"}`).
		End()

	// Test Case 2: Invalid request body
	apitest.New().
		EnableNetworking().
		Post("/login").
		JSON(dto.LoginRequestBodyDTO{
			User: dto.LoginRequestUserDto{
				Password: "123456",
			},
		}).
		Expect(t).
		Status(400).
		Body(`{"message": "error decoding request body"}`).
		End()
}

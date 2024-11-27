package dto

import (
	"strings"
	"testing"
)

//nolint:golint,exhaustruct
func TestLoginRequestBodyDTO_Validate(t *testing.T) {
	tests := []ValidationTestCase[LoginRequestBodyDTO]{
		{
			Name: "valid login request",
			Input: LoginRequestBodyDTO{
				User: LoginRequestUserDto{
					Email:    "test@example.com",
					Password: "password123",
				},
			},
			WantErrors: false,
		},
		{
			Name: "missing email",
			Input: LoginRequestBodyDTO{
				User: LoginRequestUserDto{
					Password: "password123",
				},
			},
			WantErrors: true,
			ExpectedError: map[string]string{
				"User.Email": "Email is a required field",
			},
		},
		{
			Name: "invalid email format",
			Input: LoginRequestBodyDTO{
				User: LoginRequestUserDto{
					Email:    "invalid-email",
					Password: "password123",
				},
			},
			WantErrors: true,
			ExpectedError: map[string]string{
				"User.Email": "Email must be a valid email address",
			},
		},
		{
			Name: "password too short",
			Input: LoginRequestBodyDTO{
				User: LoginRequestUserDto{
					Email:    "test@example.com",
					Password: "12345",
				},
			},
			WantErrors: true,
			ExpectedError: map[string]string{
				"User.Password": "Password must be at least 6 characters in length",
			},
		},
		{
			Name: "password too long",
			Input: LoginRequestBodyDTO{
				User: LoginRequestUserDto{
					Email:    "test@example.com",
					Password: "123456789012345678901", // 21 characters
				},
			},
			WantErrors: true,
			ExpectedError: map[string]string{
				"User.Password": "Password must be a maximum of 20 characters in length",
			},
		},
		{
			Name: "blank password",
			Input: LoginRequestBodyDTO{
				User: LoginRequestUserDto{
					Email:    "test@example.com",
					Password: "      ",
				},
			},
			WantErrors: true,
			ExpectedError: map[string]string{
				"User.Password": "Password cannot be blank",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			testValidation(t, tt)
		})
	}
}

//nolint:golint,exhaustruct
func TestNewUserRequestBodyDTO_Validate(t *testing.T) {
	tests := []ValidationTestCase[NewUserRequestBodyDTO]{
		{
			Name: "valid new user request",
			Input: NewUserRequestBodyDTO{
				User: NewUserRequestUserDto{
					Email:    "test@example.com",
					Password: "password123",
					Username: "testuser",
				},
			},
			WantErrors: false,
		},
		{
			Name: "missing email",
			Input: NewUserRequestBodyDTO{
				User: NewUserRequestUserDto{
					Password: "password123",
					Username: "testuser",
				},
			},
			WantErrors: true,
			ExpectedError: map[string]string{
				"User.Email": "Email is a required field",
			},
		},
		{
			Name: "invalid email format",
			Input: NewUserRequestBodyDTO{
				User: NewUserRequestUserDto{
					Email:    "invalid-email",
					Password: "password123",
					Username: "testuser",
				},
			},
			WantErrors: true,
			ExpectedError: map[string]string{
				"User.Email": "Email must be a valid email address",
			},
		},
		{
			Name: "password too short",
			Input: NewUserRequestBodyDTO{
				User: NewUserRequestUserDto{
					Email:    "test@example.com",
					Password: "12345",
					Username: "testuser",
				},
			},
			WantErrors: true,
			ExpectedError: map[string]string{
				"User.Password": "Password must be at least 6 characters in length",
			},
		},
		{
			Name: "password too long",
			Input: NewUserRequestBodyDTO{
				User: NewUserRequestUserDto{
					Email:    "test@example.com",
					Password: "123456789012345678901", // 21 characters
					Username: "testuser",
				},
			},
			WantErrors: true,
			ExpectedError: map[string]string{
				"User.Password": "Password must be a maximum of 20 characters in length",
			},
		},
		{
			Name: "username too short",
			Input: NewUserRequestBodyDTO{
				User: NewUserRequestUserDto{
					Email:    "test@example.com",
					Password: "password123",
					Username: "ab",
				},
			},
			WantErrors: true,
			ExpectedError: map[string]string{
				"User.Username": "Username must be at least 3 characters in length",
			},
		},
		{
			Name: "username too long",
			Input: NewUserRequestBodyDTO{
				User: NewUserRequestUserDto{
					Email:    "test@example.com",
					Password: "password123",
					Username: strings.Repeat("a", 65),
				},
			},
			WantErrors: true,
			ExpectedError: map[string]string{
				"User.Username": "Username must be a maximum of 64 characters in length",
			},
		},
		{
			Name: "blank username",
			Input: NewUserRequestBodyDTO{
				User: NewUserRequestUserDto{
					Email:    "test@example.com",
					Password: "password123",
					Username: "     ",
				},
			},
			WantErrors: true,
			ExpectedError: map[string]string{
				"User.Username": "Username cannot be blank",
			},
		},
		{
			Name: "multiple validation errors",
			Input: NewUserRequestBodyDTO{
				User: NewUserRequestUserDto{
					Email:    "invalid-email",
					Password: "12345",
					Username: "ab",
				},
			},
			WantErrors: true,
			ExpectedError: map[string]string{
				"User.Email":    "Email must be a valid email address",
				"User.Password": "Password must be at least 6 characters in length",
				"User.Username": "Username must be at least 3 characters in length",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			testValidation(t, tt)
		})
	}
}

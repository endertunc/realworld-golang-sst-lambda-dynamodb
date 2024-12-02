package security

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

//nolint:golint,exhaustruct
func TestGetLoggedInUser(t *testing.T) {
	ctx := context.Background()
	validUserId := uuid.New()
	validToken, err := GenerateToken(validUserId)
	assert.NoError(t, err)

	tests := []struct {
		name           string
		authHeader     string
		expectedUser   *uuid.UUID
		expectedToken  *domain.Token
		expectedStatus int
		expectedError  string
	}{
		{
			name:          "valid token",
			authHeader:    "Token " + string(*validToken),
			expectedUser:  &validUserId,
			expectedToken: validToken,
		},
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedStatus: 401,
			expectedError:  "authorization header is missing",
		},
		{
			name:           "empty authorization header",
			authHeader:     "   ",
			expectedStatus: 401,
			expectedError:  "authorization header is empty",
		},
		{
			name:           "invalid token type",
			authHeader:     "Bearer " + string(*validToken),
			expectedStatus: 401,
			expectedError:  "invalid token type",
		},
		{
			name:           "invalid token",
			authHeader:     "Token invalid-token",
			expectedStatus: 401,
			expectedError:  "invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				r.Header.Set("Authorization", tt.authHeader)
			}

			userId, token, ok := GetLoggedInUser(ctx, w, r)

			if tt.expectedToken != nil {
				assert.True(t, ok)
				assert.Equal(t, tt.expectedUser, &userId)
				assert.Equal(t, tt.expectedToken, &token)
			} else {
				assert.False(t, ok)
				assert.Equal(t, tt.expectedStatus, w.Code)

				var errorResponse errutil.SimpleError
				err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, errorResponse.Message)
			}
		})
	}
}

//nolint:golint,exhaustruct
func TestGetOptionalLoggedInUser(t *testing.T) {
	ctx := context.Background()
	validUserId := uuid.New()
	validToken, err := GenerateToken(validUserId)
	assert.NoError(t, err)

	tests := []struct {
		name           string
		authHeader     string
		expectedUser   *uuid.UUID
		expectedToken  *domain.Token
		expectedStatus int
		expectedError  string
	}{
		{
			name:          "valid token",
			authHeader:    "Token " + string(*validToken),
			expectedUser:  &validUserId,
			expectedToken: validToken,
		},
		{
			name:          "missing authorization header",
			authHeader:    "",
			expectedUser:  nil,
			expectedToken: nil,
		},
		{
			name:           "empty authorization header",
			authHeader:     "   ",
			expectedStatus: 401,
			expectedError:  "authorization header is empty",
		},
		{
			name:           "invalid token type",
			authHeader:     "Bearer " + string(*validToken),
			expectedStatus: 401,
			expectedError:  "invalid token type",
		},
		{
			name:           "invalid token",
			authHeader:     "Token invalid-token",
			expectedStatus: 401,
			expectedError:  "invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				r.Header.Set("Authorization", tt.authHeader)
			}

			userId, token, ok := GetOptionalLoggedInUser(ctx, w, r)

			if tt.expectedError == "" {
				assert.Equal(t, w.Body.Len(), 0, "response body should be empty")
				if tt.expectedUser == nil {
					assert.Nil(t, userId)
					assert.Nil(t, token)
				} else {
					assert.Equal(t, tt.expectedUser, userId)
					assert.Equal(t, tt.expectedToken, token)
				}
				assert.True(t, ok)
			} else {
				assert.Equal(t, tt.expectedStatus, w.Code)
				assert.False(t, ok)

				var errorResponse errutil.SimpleError
				err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, errorResponse.Message)
			}
		})
	}
}

package security

import (
	"context"
	"encoding/json"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetLoggedInUser(t *testing.T) {
	ctx := context.Background()
	validUserId := uuid.New()
	validToken, err := GenerateToken(validUserId)
	assert.NoError(t, err)

	tests := []struct {
		name           string
		request        events.APIGatewayProxyRequest
		expectedUser   *uuid.UUID
		expectedToken  *domain.Token
		expectedStatus int
		expectedError  string
	}{
		{
			name: "valid token",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"authorization": "Token " + string(*validToken),
				},
			},
			expectedUser:  &validUserId,
			expectedToken: validToken,
		},
		{
			name: "missing authorization header",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{},
			},
			expectedStatus: 401,
			expectedError:  "authorization header is missing",
		},
		{
			name: "empty authorization header",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"authorization": "   ",
				},
			},
			expectedStatus: 401,
			expectedError:  "authorization header is empty",
		},
		{
			name: "invalid token type",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"authorization": "Bearer " + string(*validToken),
				},
			},
			expectedStatus: 401,
			expectedError:  "invalid token type",
		},
		{
			name: "invalid token",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"authorization": "Token invalid-token",
				},
			},
			expectedStatus: 401,
			expectedError:  "invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userId, token, response := GetLoggedInUser(ctx, tt.request)

			if tt.expectedToken != nil {
				assert.Nil(t, response)
				assert.Equal(t, tt.expectedUser, &userId)
				assert.Equal(t, tt.expectedToken, &token)
			} else {
				assert.NotNil(t, response)
				assert.Equal(t, tt.expectedStatus, response.StatusCode)

				var errorResponse errutil.SimpleError
				err := json.Unmarshal([]byte(response.Body), &errorResponse)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, errorResponse.Message)

			}
		})
	}
}

func TestGetOptionalLoggedInUser(t *testing.T) {
	ctx := context.Background()
	validUserId := uuid.New()
	validToken, err := GenerateToken(validUserId)
	assert.NoError(t, err)

	tests := []struct {
		name           string
		request        events.APIGatewayProxyRequest
		expectedUser   *uuid.UUID
		expectedToken  *domain.Token
		expectedStatus int
		expectedError  string
	}{
		{
			name: "valid token",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"authorization": "Token " + string(*validToken),
				},
			},
			expectedUser:   &validUserId,
			expectedToken:  validToken,
			expectedStatus: 0, // expect no response
		},
		{
			name: "missing authorization header",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{},
			},
			expectedUser:   nil,
			expectedToken:  nil,
			expectedStatus: 0, // expect no response
		},
		{
			name: "empty authorization header",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"authorization": "   ",
				},
			},
			expectedStatus: 401,
			expectedError:  "authorization header is empty",
		},
		{
			name: "invalid token type",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"authorization": "Bearer " + string(*validToken),
				},
			},
			expectedStatus: 401,
			expectedError:  "invalid token type",
		},
		{
			name: "invalid token",
			request: events.APIGatewayProxyRequest{
				Headers: map[string]string{
					"authorization": "Token invalid-token",
				},
			},
			expectedStatus: 401,
			expectedError:  "invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userId, token, response := GetOptionalLoggedInUser(ctx, tt.request)

			if tt.expectedStatus == 0 {
				assert.Nil(t, response)
				if tt.expectedUser == nil {
					assert.Nil(t, userId)
					assert.Nil(t, token)
				} else {
					assert.Equal(t, tt.expectedUser, userId)
					assert.Equal(t, tt.expectedToken, token)
				}
			} else {
				assert.NotNil(t, response)
				assert.Equal(t, tt.expectedStatus, response.StatusCode)

				var errorResponse errutil.SimpleError
				err := json.Unmarshal([]byte(response.Body), &errorResponse)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedError, errorResponse.Message)
			}
		})
	}
}

package api

import (
	"context"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"
)

func TestGetOptionalIntQueryParam(t *testing.T) {
	ctx := context.Background()
	minLimit := 1
	maxLimit := 100

	tests := []struct {
		name          string
		request       events.APIGatewayProxyRequest
		paramName     string
		min           *int
		max           *int
		expectedValue *int
		expectError   bool
		errorMessage  string
	}{
		{
			name: "valid value",
			request: events.APIGatewayProxyRequest{
				QueryStringParameters: map[string]string{
					"limit": "10",
				},
			},
			paramName:     "limit",
			min:           &minLimit,
			max:           &maxLimit,
			expectedValue: aws.Int(10),
			expectError:   false,
		},
		{
			name: "missing parameter returns nil",
			request: events.APIGatewayProxyRequest{
				QueryStringParameters: map[string]string{},
			},
			paramName:     "limit",
			min:           &minLimit,
			max:           &maxLimit,
			expectedValue: nil,
			expectError:   false,
		},
		{
			name: "invalid integer",
			request: events.APIGatewayProxyRequest{
				QueryStringParameters: map[string]string{
					"limit": "not-a-number",
				},
			},
			paramName:    "limit",
			min:          &minLimit,
			max:          &maxLimit,
			expectError:  true,
			errorMessage: "query parameter limit must be a valid integer",
		},
		{
			name: "below minimum",
			request: events.APIGatewayProxyRequest{
				QueryStringParameters: map[string]string{
					"limit": "0",
				},
			},
			paramName:    "limit",
			min:          &minLimit,
			max:          &maxLimit,
			expectError:  true,
			errorMessage: "query parameter limit must be greater than or equal to 1",
		},
		{
			name: "at minimum",
			request: events.APIGatewayProxyRequest{
				QueryStringParameters: map[string]string{
					"limit": "1",
				},
			},
			paramName:     "limit",
			min:           &minLimit,
			max:           &maxLimit,
			expectedValue: aws.Int(1),
			expectError:   false,
		},
		{
			name: "above maximum",
			request: events.APIGatewayProxyRequest{
				QueryStringParameters: map[string]string{
					"limit": "101",
				},
			},
			paramName:    "limit",
			min:          &minLimit,
			max:          &maxLimit,
			expectError:  true,
			errorMessage: "query parameter limit must be less than or equal to 100",
		},
		{
			name: "at maximum",
			request: events.APIGatewayProxyRequest{
				QueryStringParameters: map[string]string{
					"limit": "100",
				},
			},
			paramName:     "limit",
			min:           &minLimit,
			max:           &maxLimit,
			expectedValue: aws.Int(100),
			expectError:   false,
		},
		{
			name: "no minLimit validation",
			request: events.APIGatewayProxyRequest{
				QueryStringParameters: map[string]string{
					"limit": "-10",
				},
			},
			paramName:     "limit",
			min:           nil,
			max:           &maxLimit,
			expectedValue: aws.Int(-10),
			expectError:   false,
		},
		{
			name: "no maxLimit validation",
			request: events.APIGatewayProxyRequest{
				QueryStringParameters: map[string]string{
					"limit": "1000",
				},
			},
			paramName:     "limit",
			min:           &minLimit,
			max:           nil,
			expectedValue: aws.Int(1000),
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, response := GetOptionalIntQueryParam(ctx, tt.request, tt.paramName, tt.min, tt.max)

			if tt.expectError {
				assert.NotNil(t, response)
				assert.Equal(t, 400, response.StatusCode)
				assert.Contains(t, response.Body, tt.errorMessage)
			} else {
				assert.Nil(t, response)
				if tt.expectedValue == nil {
					assert.Nil(t, value)
				} else {
					assert.Equal(t, *tt.expectedValue, *value)
				}
			}
		})
	}
}

func TestGetIntQueryParamOrDefault(t *testing.T) {
	ctx := context.Background()
	minLimit := 1
	maxLimit := 100

	tests := []struct {
		name          string
		request       events.APIGatewayProxyRequest
		paramName     string
		defaultValue  int
		min           *int
		max           *int
		expectedValue int
		expectError   bool
		errorMessage  string
	}{
		{
			name: "valid value",
			request: events.APIGatewayProxyRequest{
				QueryStringParameters: map[string]string{
					"limit": "10",
				},
			},
			paramName:     "limit",
			defaultValue:  20,
			min:           &minLimit,
			max:           &maxLimit,
			expectedValue: 10,
			expectError:   false,
		},
		{
			name: "missing parameter returns default",
			request: events.APIGatewayProxyRequest{
				QueryStringParameters: map[string]string{},
			},
			paramName:     "limit",
			defaultValue:  20,
			min:           &minLimit,
			max:           &maxLimit,
			expectedValue: 20,
			expectError:   false,
		},
		{
			name: "invalid integer",
			request: events.APIGatewayProxyRequest{
				QueryStringParameters: map[string]string{
					"limit": "not-a-number",
				},
			},
			paramName:    "limit",
			defaultValue: 20,
			min:          &minLimit,
			max:          &maxLimit,
			expectError:  true,
			errorMessage: "query parameter limit must be a valid integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, response := GetIntQueryParamOrDefault(ctx, tt.request, tt.paramName, tt.defaultValue, tt.min, tt.max)

			if tt.expectError {
				assert.NotNil(t, response)
				assert.Equal(t, 400, response.StatusCode)
				assert.Contains(t, response.Body, tt.errorMessage)
			} else {
				assert.Nil(t, response)
				assert.Equal(t, tt.expectedValue, value)
			}
		})
	}
}

func TestGetOptionalStringQueryParam(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		request       events.APIGatewayProxyRequest
		paramName     string
		expectedValue *string
		expectError   bool
		errorMessage  string
	}{
		{
			name: "valid value",
			request: events.APIGatewayProxyRequest{
				QueryStringParameters: map[string]string{
					"token": "abc123",
				},
			},
			paramName:     "token",
			expectedValue: aws.String("abc123"),
			expectError:   false,
		},
		{
			name: "missing parameter returns nil",
			request: events.APIGatewayProxyRequest{
				QueryStringParameters: map[string]string{},
			},
			paramName:     "token",
			expectedValue: nil,
			expectError:   false,
		},
		{
			name: "blank value",
			request: events.APIGatewayProxyRequest{
				QueryStringParameters: map[string]string{
					"token": "   ",
				},
			},
			paramName:    "token",
			expectError:  true,
			errorMessage: "query parameter token cannot be blank",
		},
		{
			name: "empty string",
			request: events.APIGatewayProxyRequest{
				QueryStringParameters: map[string]string{
					"token": "",
				},
			},
			paramName:    "token",
			expectError:  true,
			errorMessage: "query parameter token cannot be blank",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, response := GetOptionalStringQueryParam(ctx, tt.request, tt.paramName)

			if tt.expectError {
				assert.NotNil(t, response)
				assert.Equal(t, 400, response.StatusCode)
				assert.Contains(t, response.Body, tt.errorMessage)
			} else {
				assert.Nil(t, response)
				if tt.expectedValue == nil {
					assert.Nil(t, value)
				} else {
					assert.Equal(t, *tt.expectedValue, *value)
				}
			}
		})
	}
}

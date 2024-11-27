//nolint:golint,exhaustruct
package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"
)

func TestGetOptionalIntQueryParamHTTP(t *testing.T) {
	ctx := context.Background()
	minLimit := 1
	maxLimit := 100

	tests := []struct {
		name          string
		queryParams   string
		paramName     string
		min           *int
		max           *int
		expectedValue *int
		expectError   bool
		errorMessage  string
	}{
		{
			name:          "valid value",
			queryParams:   "limit=10",
			paramName:     "limit",
			min:           &minLimit,
			max:           &maxLimit,
			expectedValue: aws.Int(10),
			expectError:   false,
		},
		{
			name:          "missing parameter returns nil",
			queryParams:   "",
			paramName:     "limit",
			min:           &minLimit,
			max:           &maxLimit,
			expectedValue: nil,
			expectError:   false,
		},
		{
			name:         "invalid integer",
			queryParams:  "limit=not-a-number",
			paramName:    "limit",
			min:          &minLimit,
			max:          &maxLimit,
			expectError:  true,
			errorMessage: "query parameter limit must be a valid integer",
		},
		{
			name:         "below minimum",
			queryParams:  "limit=0",
			paramName:    "limit",
			min:          &minLimit,
			max:          &maxLimit,
			expectError:  true,
			errorMessage: "query parameter limit must be greater than or equal to 1",
		},
		{
			name:          "at minimum",
			queryParams:   "limit=1",
			paramName:     "limit",
			min:           &minLimit,
			max:           &maxLimit,
			expectedValue: aws.Int(1),
			expectError:   false,
		},
		{
			name:         "above maximum",
			queryParams:  "limit=101",
			paramName:    "limit",
			min:          &minLimit,
			max:          &maxLimit,
			expectError:  true,
			errorMessage: "query parameter limit must be less than or equal to 100",
		},
		{
			name:          "at maximum",
			queryParams:   "limit=100",
			paramName:     "limit",
			min:           &minLimit,
			max:           &maxLimit,
			expectedValue: aws.Int(100),
			expectError:   false,
		},
		{
			name:          "no minLimit validation",
			queryParams:   "limit=-10",
			paramName:     "limit",
			min:           nil,
			max:           &maxLimit,
			expectedValue: aws.Int(-10),
			expectError:   false,
		},
		{
			name:          "no maxLimit validation",
			queryParams:   "limit=1000",
			paramName:     "limit",
			min:           &minLimit,
			max:           nil,
			expectedValue: aws.Int(1000),
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/?"+tt.queryParams, nil)

			value, ok := GetOptionalIntQueryParam(ctx, w, r, tt.paramName, tt.min, tt.max)

			if tt.expectError {
				assert.False(t, ok)
				assert.Equal(t, http.StatusBadRequest, w.Code)
				assert.Contains(t, w.Body.String(), tt.errorMessage)
			} else {
				assert.True(t, ok)
				if tt.expectedValue == nil {
					assert.Nil(t, value)
				} else {
					assert.Equal(t, *tt.expectedValue, *value)
				}
			}
		})
	}
}

func TestGetIntQueryParamOrDefaultHTTP(t *testing.T) {
	ctx := context.Background()
	minLimit := 1
	maxLimit := 100

	tests := []struct {
		name          string
		queryParams   string
		paramName     string
		defaultValue  int
		min           *int
		max           *int
		expectedValue int
		expectError   bool
		errorMessage  string
	}{
		{
			name:          "valid value",
			queryParams:   "limit=10",
			paramName:     "limit",
			defaultValue:  20,
			min:           &minLimit,
			max:           &maxLimit,
			expectedValue: 10,
			expectError:   false,
		},
		{
			name:          "missing parameter returns default",
			queryParams:   "",
			paramName:     "limit",
			defaultValue:  20,
			min:           &minLimit,
			max:           &maxLimit,
			expectedValue: 20,
			expectError:   false,
		},
		{
			name:         "invalid integer",
			queryParams:  "limit=not-a-number",
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
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/?"+tt.queryParams, nil)

			value, ok := GetIntQueryParamOrDefault(ctx, w, r, tt.paramName, tt.defaultValue, tt.min, tt.max)

			if tt.expectError {
				assert.False(t, ok)
				assert.Equal(t, http.StatusBadRequest, w.Code)
				assert.Contains(t, w.Body.String(), tt.errorMessage)
			} else {
				assert.True(t, ok)
				assert.Equal(t, tt.expectedValue, value)
			}
		})
	}
}

func TestGetOptionalStringQueryParamHTTP(t *testing.T) {
	tests := []struct {
		name          string
		queryParams   map[string]string
		paramName     string
		expectedValue *string
		expectError   bool
		errorMessage  string
	}{
		{
			name: "valid value",
			queryParams: map[string]string{
				"token": "abc123",
			},
			paramName:     "token",
			expectedValue: aws.String("abc123"),
			expectError:   false,
		},
		{
			name:          "missing parameter returns nil",
			queryParams:   map[string]string{},
			paramName:     "token",
			expectedValue: nil,
			expectError:   false,
		},
		{
			name: "blank value",
			queryParams: map[string]string{
				"token": "     ",
			},
			paramName:    "token",
			expectError:  true,
			errorMessage: "query parameter token cannot be blank",
		},
		{
			name: "empty string returns nil",
			queryParams: map[string]string{
				"token": "",
			},
			paramName:     "token",
			expectError:   false,
			expectedValue: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)

			q := r.URL.Query()
			for key, value := range tt.queryParams {
				q.Add(key, value)
			}
			r.URL.RawQuery = q.Encode()

			value, ok := GetOptionalStringQueryParam(w, r, tt.paramName)

			if tt.expectError {
				assert.False(t, ok)
				assert.Equal(t, http.StatusBadRequest, w.Code)
				assert.Contains(t, w.Body.String(), tt.errorMessage)
			} else {
				assert.True(t, ok)
				if tt.expectedValue == nil {
					assert.Nil(t, value)
				} else {
					assert.Equal(t, *tt.expectedValue, *value)
				}
			}
		})
	}
}

func TestGetPathParamHTTP(t *testing.T) {
	ctx := context.Background()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		value, ok := GetPathParamHTTP(ctx, w, r, "id")
		if ok {
			_, _ = w.Write([]byte(value))
		}
	})

	tests := []struct {
		name          string
		path          string
		expectedValue string
		expectError   bool
		pathValues    map[string]string
		errorMessage  string
	}{
		{
			name:          "valid value",
			path:          "/users/123",
			expectedValue: "123",
			pathValues:    map[string]string{"id": "123"},
			expectError:   false,
		},
		{
			name:         "missing parameter",
			path:         "/users/",
			pathValues:   map[string]string{"slug": "hello-world"},
			expectError:  true,
			errorMessage: "path parameter id is missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, tt.path, nil)
			for key, value := range tt.pathValues {
				r.SetPathValue(key, value)
			}
			handler.ServeHTTP(w, r)
			if tt.expectError {
				assert.Equal(t, http.StatusBadRequest, w.Code)
				assert.Contains(t, w.Body.String(), tt.errorMessage)
			} else {
				assert.Equal(t, tt.expectedValue, w.Body.String())
			}
		})
	}
}

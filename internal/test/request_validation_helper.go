package test

import (
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"testing"
)

// ApiRequestValidationTest represents a test case for validating an API request
type ApiRequestValidationTest[T any] struct {
	Name          string
	Input         T
	ExpectedError map[string]string
}

// TestValidation is a helper function that handles the common aPI validation testing logic
func TestValidation[T any](t *testing.T, tt ApiRequestValidationTest[T], request func(t *testing.T, input T) errutil.ValidationErrors) {
	t.Helper()
	response := request(t, tt.Input)
	assert.Equal(t, len(tt.ExpectedError), len(response.Errors))
	for field, expectedMsg := range tt.ExpectedError {
		actualValidationError, found := lo.Find(response.Errors, func(err errutil.ValidationError) bool {
			return err.Field == field
		})
		if !found {
			t.Errorf("%s: Validate() missing error for field %s", tt.Name, field)
		} else if actualValidationError.Message != expectedMsg {
			t.Errorf("%s: Validate() error for field %s = %v, want %v", tt.Name, field, actualValidationError.Message, expectedMsg)
		}
	}
}

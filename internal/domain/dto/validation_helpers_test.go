package dto

import "testing"

// ValidationTestCase represents a test case for validation testing
type ValidationTestCase[T Validatable] struct {
	Name          string
	Input         T
	WantErrors    bool
	ExpectedError map[string]string
}

// testValidation is a helper function that handles the common validation testing logic
func testValidation[T Validatable](t *testing.T, tc ValidationTestCase[T]) {
	t.Helper()
	errors, hasErrors := tc.Input.Validate()
	if hasErrors != tc.WantErrors {
		t.Errorf("%s: Validate() hasErrors = %v, want %v", tc.Name, hasErrors, tc.WantErrors)
	}
	if tc.WantErrors {
		if len(errors) != len(tc.ExpectedError) {
			t.Errorf("%s: Validate() got %d errors, want %d errors. Got: %v", tc.Name, len(errors), len(tc.ExpectedError), errors)
		}
		for field, expectedMsg := range tc.ExpectedError {
			if actualMsg, ok := errors[field]; !ok {
				t.Errorf("%s: Validate() missing error for field %s", tc.Name, field)
			} else if actualMsg != expectedMsg {
				t.Errorf("%s: Validate() error for field %s = %v, want %v", tc.Name, field, actualMsg, expectedMsg)
			}
		}
	}
}

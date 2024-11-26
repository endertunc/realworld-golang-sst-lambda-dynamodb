package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"strconv"
	"strings"
)

func GetPathParamHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request, paramName string) (string, bool) {
	param := r.PathValue(paramName)
	if param == "" {
		ToSimpleHTTPError(w, http.StatusBadRequest, fmt.Sprintf("path parameter %s is missing", paramName))
		return "", false
	}
	return param, true
}

func ParseAndValidateBody[T dto.Validatable](ctx context.Context, w http.ResponseWriter, r *http.Request) (*T, bool) {
	var out T
	err := json.NewDecoder(r.Body).Decode(&out)
	if err != nil {
		cause := fmt.Errorf("error decoding request body: %w", err)
		slog.WarnContext(ctx, "error decoding request body", slog.Any("error", cause))
		// A note on err.Error():
		// json module returns a descriptive enough error message that we can use as is.
		// As far as I can tell, there is not much of risk of leaking sensitive information here.
		//
		// It's unnecessary in this project, but one could use *json.SyntaxError and *json.UnmarshalTypeError
		// to provide more structured error messages.
		// JsonParsingErrorStruct {
		//  Field string
		// 	Message string
		// 	Position int
		//  etc...
		// }
		ToSimpleHTTPError(w, http.StatusBadRequest, err.Error())
		return nil, false
	}

	validationErrorsMap := out.Validate()

	if len(validationErrorsMap) > 0 {
		ToFieldValidationHTTPError(w, http.StatusBadRequest, validationErrorsMap)
		return nil, false
	}

	return &out, true
}

func GetOptionalIntQueryParam(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	paramName string,
	min,
	max *int,
) (*int, bool) {
	param := r.URL.Query().Get(paramName)
	if param == "" {
		return nil, true
	}

	value, err := strconv.Atoi(param)
	if err != nil {
		ToSimpleHTTPError(w, http.StatusBadRequest, fmt.Sprintf("query parameter %s must be a valid integer", paramName))
		return nil, false
	}

	if min != nil && value < *min {
		ToSimpleHTTPError(w, http.StatusBadRequest, fmt.Sprintf("query parameter %s must be greater than or equal to %d", paramName, *min))
		return nil, false
	}

	if max != nil && value > *max {
		ToSimpleHTTPError(w, http.StatusBadRequest, fmt.Sprintf("query parameter %s must be less than or equal to %d", paramName, *max))
		return nil, false
	}

	return &value, true
}

func GetIntQueryParamOrDefault(
	ctx context.Context,
	w http.ResponseWriter,
	r *http.Request,
	paramName string,
	defaultValue int,
	min,
	max *int,
) (int, bool) {
	param, ok := GetOptionalIntQueryParam(ctx, w, r, paramName, min, max)
	if !ok {
		return 0, false
	} else if param == nil {
		return defaultValue, true
	} else {
		return *param, true
	}
}

func GetOptionalStringQueryParam(
	w http.ResponseWriter,
	r *http.Request,
	paramName string,
) (*string, bool) {
	param := r.URL.Query().Get(paramName)
	if param == "" {
		return nil, true
	}

	if strings.TrimSpace(param) == "" {
		ToSimpleHTTPError(w, http.StatusBadRequest, fmt.Sprintf("query parameter %s cannot be blank", paramName))
		return nil, false
	}

	return &param, true
}

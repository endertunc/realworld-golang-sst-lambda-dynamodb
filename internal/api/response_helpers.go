package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"

	"github.com/aws/aws-lambda-go/events"
)

var InternalServerError = events.APIGatewayProxyResponse{
	StatusCode: http.StatusInternalServerError,
	Body:       "internal server error",
}

func ToSimpleError(ctx context.Context, statusCode int, message string) events.APIGatewayProxyResponse {
	body, err := json.Marshal(errutil.SimpleError{Message: message})
	if err != nil {
		slog.ErrorContext(ctx, "error encoding response body", slog.Any("error", err))
		return InternalServerError
	}

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       string(body),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}
}

func ToSuccessHTTPResponse(w http.ResponseWriter, body interface{}) {
	if body == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	bodyJson, err := json.Marshal(body)
	if err != nil {
		cause := fmt.Errorf("%w: %w", errutil.ErrJsonEncode, err)
		ToInternalServerHTTPError(w, cause)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(bodyJson)
}

func ToSimpleHTTPError(w http.ResponseWriter, statusCode int, message string) {
	body, err := json.Marshal(errutil.SimpleError{Message: message})
	if err != nil {
		slog.Error("error encoding response body", slog.Any("error", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(body)
}

func ToFieldValidationHTTPError(w http.ResponseWriter, statusCode int, validationErrors dto.ValidationErrors) {
	body, err := json.Marshal(validationErrors.ToHttpValidationError())
	if err != nil {
		slog.Error("error encoding validation errors response body", slog.Any("error", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(body)
}

func ToInternalServerHTTPError(w http.ResponseWriter, err error) {
	slog.Error("unexpected error", slog.Any("error", err))
	http.Error(w, "internal server error", http.StatusInternalServerError)
}

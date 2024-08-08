package errutil

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"net/http"
)

type AppError struct {
	Code            int    `json:"code"`
	Message         string `json:"message"`
	Operation       string `json:"-"`
	InternalMessage string `json:"-"`
	Cause           error  `json:"-"`
}

func (e AppError) Error() string {
	return e.Message
}

func (e AppError) Unwrap() error {
	return e.Cause
}

func (e AppError) WithInternalMessage(message string) AppError {
	e.InternalMessage = message
	return e
}

func BadRequestError(operation, message string) AppError {
	return AppError{
		Code:      http.StatusBadRequest,
		Message:   message,
		Operation: operation,
	}
}

func UnauthorizedError(operation, message string) AppError {
	return AppError{
		Code:      http.StatusUnauthorized,
		Message:   message,
		Operation: operation,
	}
}

func UserNotFound(operation, message string, cause error) AppError {
	return AppError{
		Code:      http.StatusBadRequest,
		Message:   message,
		Operation: operation,
		Cause:     cause,
	}
}

func NotFoundError(operation, message string) AppError {
	return AppError{
		Code:      http.StatusNotFound,
		Message:   message,
		Operation: operation,
	}
}

func InternalError(operation string) AppError {
	return AppError{
		Code:      http.StatusInternalServerError,
		Message:   "internal error",
		Operation: operation,
	}
}

func (e AppError) Errorf(format string, args ...any) AppError {
	err := fmt.Errorf(format, args...)

	return AppError{
		Code:            e.Code,
		Message:         e.Message,
		Operation:       e.Operation,
		InternalMessage: err.Error(),
		Cause:           fmt.Errorf("%w: %w", e, err),
	}
}

//func (e AppError) Errorf(format string, args ...any) AppError {
//	err := fmt.Errorf(format, args...)
//
//	return AppError{
//		Code:            e.Code,
//		InternalMessage: err.Error(),
//		Message:         e.Message,
//		Operation:       e.Operation,
//		Cause:           errors.Unwrap(err),
//	}
//}

// ToAPIGatewayProxyResponse ToDo @ender this function should log the errors
func ToAPIGatewayProxyResponse(context context.Context, error error) events.APIGatewayProxyResponse {
	// ToDo log AppErr
	var appErr *AppError
	ok := errors.As(error, &appErr)
	if !ok {
		// ToDo @ender log "error" here...
		internalServerError, err := json.Marshal(AppError{
			Code:    http.StatusInternalServerError,
			Message: "internal error",
			Cause:   nil,
		})
		if err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       string(internalServerError),
				Headers:    map[string]string{"Content-Type": "application/json"},
			}
		}
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "internal error",
		}
	}

	responseBody, err := json.Marshal(appErr)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "internal error",
		}
	}

	return events.APIGatewayProxyResponse{
		StatusCode: appErr.Code,
		Body:       string(responseBody),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}
}

var (
	ErrUserNotFound  = NotFoundError("user.not_found", "user not found")
	ErrDynamoQuery   = InternalError("dynamodb.query")
	ErrDynamoMapping = InternalError("dynamodb.mapping")
	ErrPasswordHash  = BadRequestError("password.hash", "invalid credentials")
	ErrTokenGenerate = InternalError("token.generate")
	ErrJsonDecode    = BadRequestError("json.decoder", "error decoding request body")
	ErrJsonEncode    = InternalError("json.encode")
)

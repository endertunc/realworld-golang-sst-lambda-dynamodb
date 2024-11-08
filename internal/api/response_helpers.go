package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"log/slog"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
)

var InternalServerError = events.APIGatewayProxyResponse{
	StatusCode: http.StatusInternalServerError,
	Body:       "internal server error",
}

func ToSuccessAPIGatewayProxyResponse(context context.Context, body interface{}, handlerName string) events.APIGatewayProxyResponse {
	if body == nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
		}
	}

	bodyJson, err := json.Marshal(body)
	if err != nil {
		cause := fmt.Errorf("%w: %w", errutil.ErrJsonEncode, err)
		return ToInternalServerError(context, cause)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(bodyJson),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}
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

func ToInternalServerError(ctx context.Context, err error) events.APIGatewayProxyResponse {
	slog.ErrorContext(ctx, "unexpected error", slog.Any("error", err))
	return InternalServerError
}

package api

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"log/slog"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
)

//func ToApiGatewayProxyResponse(context context.Context, result interface{}, handlerName string) events.APIGatewayProxyResponse {
//	jsonResult, err := json.Marshal(result)
//	if err != nil {
//		cause := errutil.ErrJsonEncode.Errorf("%s - error encoding response body: %w", handlerName, err)
//		return errutil.ToAPIGatewayProxyResponse(context, handlerName, cause)
//	}
//
//	return events.APIGatewayProxyResponse{
//		StatusCode: 200,
//		Body:       string(jsonResult),
//		Headers:    map[string]string{"Content-Type": "application/json"},
//	}
//}

//func ToEmptyApiGatewayProxyResponse() events.APIGatewayProxyResponse {
//	return events.APIGatewayProxyResponse{
//		StatusCode: 200,
//	}
//}

var InternalServerError = events.APIGatewayProxyResponse{
	StatusCode: http.StatusInternalServerError,
	Body:       "internal server error",
}

func ToSimpleError(ctx context.Context, statusCode int, message string) events.APIGatewayProxyResponse {
	body, err := json.Marshal(errutil.GenericError{Message: message})
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

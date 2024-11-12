package main

import (
	"context"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/cmd/functions"
	"realworld-aws-lambda-dynamodb-golang/internal/api"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
)

const handlerName = "GetCurrentUserHandler"

func init() {
	http.Handle("GET /api/user", api.StartAuthenticatedHandlerHTTP(HandlerHTTP))
}

func HandlerHTTP(w http.ResponseWriter, r *http.Request, userId uuid.UUID, token domain.Token) {
	ctx := r.Context()

	result, err := functions.UserApi.GetCurrentUser(ctx, userId, token)

	if err != nil {
		if errors.Is(err, errutil.ErrUserNotFound) {
			slog.WarnContext(ctx, "user not found", slog.Any("error", err))
			api.ToSimpleHTTPError(w, http.StatusNotFound, "user not found")
			return
		}
		api.ToInternalServerHTTPError(w, err)
		return
	}

	api.ToSuccessHTTPResponse(w, result)
}

func Handler(ctx context.Context, _ events.APIGatewayProxyRequest, userId uuid.UUID, token domain.Token) events.APIGatewayProxyResponse {
	result, err := functions.UserApi.GetCurrentUser(ctx, userId, token)

	if err != nil {
		if errors.Is(err, errutil.ErrUserNotFound) {
			slog.WarnContext(ctx, "user not found", slog.Any("error", err))
			return api.ToSimpleError(ctx, http.StatusNotFound, "user not found")
		}
		return api.ToInternalServerError(ctx, err)
	} else {
		slog.DebugContext(ctx, "current user", slog.Any("user", result))
		return api.ToSuccessAPIGatewayProxyResponse(ctx, result, handlerName)
	}
}

func main() {
	lambda.Start(httpadapter.NewV2(http.DefaultServeMux).ProxyWithContext)
}

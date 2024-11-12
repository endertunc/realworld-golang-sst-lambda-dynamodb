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

const handlerName = "UnFollowUserHandler"

func init() {
	http.Handle("DELETE /api/profiles/{username}/follow", api.StartAuthenticatedHandlerHTTP(HandlerHTTP))
}

func HandlerHTTP(w http.ResponseWriter, r *http.Request, userId uuid.UUID, _ domain.Token) {
	ctx := r.Context()

	username, ok := api.GetPathParamHTTP(ctx, w, r, "username")
	if !ok {
		return
	}

	result, err := functions.ProfileApi.UnfollowUserByUsername(ctx, userId, username)

	if err != nil {
		if errors.Is(err, errutil.ErrUserNotFound) {
			slog.DebugContext(ctx, "user not found", slog.String("username", username))
			api.ToSimpleHTTPError(w, http.StatusNotFound, "user not found")
			return
		} else if errors.Is(err, errutil.ErrCantFollowYourself) {
			slog.DebugContext(ctx, "user is already unfollowed", slog.String("username", username), slog.String("userId", userId.String()))
			api.ToSimpleHTTPError(w, http.StatusConflict, "user is already unfollowed")
			return
		} else {
			api.ToInternalServerHTTPError(w, err)
			return
		}
	}

	api.ToSuccessHTTPResponse(w, result)
	return
}

func Handler(ctx context.Context, request events.APIGatewayProxyRequest, userId uuid.UUID, _ domain.Token) events.APIGatewayProxyResponse {
	// ToDo it's a bit annoying that this could fail even tho the path is required for this endpoint to match...
	username, response := api.GetPathParam(ctx, request, "username", handlerName)

	if response != nil {
		return *response
	}

	result, err := functions.ProfileApi.UnfollowUserByUsername(ctx, userId, username)

	if err != nil {
		if errors.Is(err, errutil.ErrUserNotFound) {
			slog.DebugContext(ctx, "user to unfollow not found", slog.Any("username", username), slog.Any("error", err))
			return api.ToSimpleError(ctx, http.StatusNotFound, "user not found")
		} else if errors.Is(err, errutil.ErrCantFollowYourself) {
			slog.DebugContext(ctx, "user tried to unfollow itself", slog.Any("username", username), slog.String("userId", userId.String()), slog.Any("error", err))
			return api.ToSimpleError(ctx, http.StatusBadRequest, "cannot unfollow yourself")
		} else {
			return api.ToInternalServerError(ctx, err)
		}
	} else {
		return api.ToSuccessAPIGatewayProxyResponse(ctx, result, handlerName)
	}
}

func main() {
	lambda.Start(httpadapter.NewV2(http.DefaultServeMux).ProxyWithContext)
}

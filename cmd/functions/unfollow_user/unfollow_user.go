package main

import (
	"context"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/cmd/functions"
	"realworld-aws-lambda-dynamodb-golang/internal/api"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
)

const handlerName = "UnFollowUserHandler"

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
	api.StartAuthenticatedHandler(Handler)
}

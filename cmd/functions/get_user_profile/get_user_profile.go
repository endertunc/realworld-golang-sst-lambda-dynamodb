package main

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/cmd/functions"
	"realworld-aws-lambda-dynamodb-golang/internal/api"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
)
import "github.com/aws/aws-lambda-go/events"

const handlerName = "GetUserProfileHandler"

func Handler(context context.Context, request events.APIGatewayProxyRequest, userId *uuid.UUID, _ *domain.Token) events.APIGatewayProxyResponse {
	// it's a bit annoying that this could fail even tho the path is required for this endpoint to match...
	username, response := api.GetPathParam(context, request, "username", handlerName)

	if response != nil {
		return *response
	}

	result, err := functions.ProfileApi.GetUserProfile(context, userId, username)

	if err != nil {
		if errors.Is(err, errutil.ErrUserNotFound) {
			slog.DebugContext(context, "user profile not found", slog.String("username", username), slog.Any("error", err))
			return api.ToSimpleError(context, http.StatusNotFound, "user not found")
		} else {
			return api.ToInternalServerError(context, err)
		}
	} else {
		return api.ToSuccessAPIGatewayProxyResponse(context, result, handlerName)
	}
}

func main() {
	api.StartOptionallyAuthenticatedHandler(Handler)
}

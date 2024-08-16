package main

import (
	"context"
	"github.com/google/uuid"
	"realworld-aws-lambda-dynamodb-golang/cmd/functions"
	"realworld-aws-lambda-dynamodb-golang/internal/api"
)
import "github.com/aws/aws-lambda-go/events"

func Handler(context context.Context, request events.APIGatewayProxyRequest, userId *uuid.UUID) events.APIGatewayProxyResponse {
	// it's a bit annoying that this could fail even tho the path is required for this endpoint to match...
	username, response := api.GetPathParam(context, request, "username")

	if response != nil {
		return *response
	}

	result, err := functions.UserApi.GetUserProfile(context, userId, username)

	if err != nil {
		return api.ToErrorAPIGatewayProxyResponse(context, err, "GetUserProfileHandler")
	}

	return api.ToSuccessAPIGatewayProxyResponse(context, result, "GetUserProfileHandler")
}

func main() {
	api.StartOptionallyAuthenticatedHandler(Handler)
}

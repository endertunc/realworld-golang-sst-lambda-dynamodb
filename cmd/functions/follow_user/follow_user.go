package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"realworld-aws-lambda-dynamodb-golang/cmd/functions"
	"realworld-aws-lambda-dynamodb-golang/internal/api"
)

func Handler(context context.Context, request events.APIGatewayProxyRequest, userId uuid.UUID) events.APIGatewayProxyResponse {
	// it's a bit annoying that this could fail even tho the path is required for this endpoint to match...
	username, response := api.GetPathParam(context, request, "username")

	if response != nil {
		return *response
	}

	result, err := functions.FollowerApi.FollowUserByUsername(context, userId, username)

	if err != nil {
		return api.ToErrorAPIGatewayProxyResponse(context, err, "FollowUserHandler")
	}
	return api.ToSuccessAPIGatewayProxyResponse(context, result, "FollowUserHandler")
}

func main() {
	api.StartAuthenticatedHandler(Handler)
}

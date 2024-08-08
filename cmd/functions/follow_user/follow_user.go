package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"realworld-aws-lambda-dynamodb-golang/cmd/functions"
	"realworld-aws-lambda-dynamodb-golang/internal/api"
	"realworld-aws-lambda-dynamodb-golang/internal/security"
)

func Handler(context context.Context, request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	// it's a bit annoying that this could fail even tho the path is required for this endpoint to match...
	username, ok := request.PathParameters["username"]
	// ToDo Ender how to handle such situations?
	if !ok {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "username path parameter is missing", // ToDo This is not a json tho...
			Headers:    map[string]string{"Content-Type": "application/json"},
		}
	}
	userId, response := security.GetLoggedInUser(request)
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
	lambda.Start(Handler)
}

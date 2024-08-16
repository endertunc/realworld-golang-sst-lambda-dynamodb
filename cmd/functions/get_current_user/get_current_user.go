package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"realworld-aws-lambda-dynamodb-golang/cmd/functions"
	"realworld-aws-lambda-dynamodb-golang/internal/api"
)

func Handler(context context.Context, request events.APIGatewayProxyRequest, userId uuid.UUID) events.APIGatewayProxyResponse {
	result, err := functions.UserApi.GetCurrentUser(context, userId)

	if err != nil {
		return api.ToErrorAPIGatewayProxyResponse(context, err, "GetCurrentUserHandler")
	}

	return api.ToSuccessAPIGatewayProxyResponse(context, result, "GetCurrentUserHandler")
}

func main() {
	api.StartAuthenticatedHandler(Handler)
}

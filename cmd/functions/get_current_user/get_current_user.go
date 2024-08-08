package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"realworld-aws-lambda-dynamodb-golang/cmd/functions"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/security"
)

func Handler(context context.Context, request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	userId, response := security.GetLoggedInUser(request)
	if response != nil {
		return *response
	}

	result, err := functions.UserApi.GetCurrentUser(context, userId)

	jsonResult, err := json.Marshal(result)
	if err != nil {
		cause := errutil.ErrJsonEncode.Errorf("GetCurrentUserHandler - error encoding response body: %w", err)
		return errutil.ToAPIGatewayProxyResponse(context, errutil.ErrJsonEncode.Errorf(
			"GetCurrentUserHandler - error encoding response body: %w", cause))
	}
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(jsonResult),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}
}

func main() {
	lambda.Start(Handler)
}

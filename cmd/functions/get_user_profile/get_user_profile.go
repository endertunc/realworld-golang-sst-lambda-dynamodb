package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/lambda"
	"realworld-aws-lambda-dynamodb-golang/cmd/functions"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/security"
)
import "github.com/aws/aws-lambda-go/events"

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

	userId, response := security.GetOptionalLoggedInUser(request)
	if response != nil {
		return *response
	}
	result, err := functions.UserApi.GetUserProfile(context, userId, username)

	jsonResult, err := json.Marshal(result)
	if err != nil {
		cause := errutil.ErrJsonEncode.Errorf("UserProfileHandler - error encoding response body: %w", err)
		return errutil.ToAPIGatewayProxyResponse(context, errutil.ErrJsonEncode.Errorf(
			"UserProfileHandler - error encoding response body: %w", cause))
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

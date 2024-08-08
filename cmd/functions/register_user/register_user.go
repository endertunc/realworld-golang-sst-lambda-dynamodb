package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/lambda"
	"realworld-aws-lambda-dynamodb-golang/cmd/functions"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
)
import "github.com/aws/aws-lambda-go/events"

func Handler(context context.Context, request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	newUserRequestBodyDTO := dto.NewUserRequestBodyDTO{}
	err := json.Unmarshal([]byte(request.Body), &newUserRequestBodyDTO)
	if err != nil {
		cause := errutil.ErrJsonDecode.Errorf("RegisterUserHandler - error decoding request body: %w", err)
		return errutil.ToAPIGatewayProxyResponse(context, cause)
	}

	result, err := functions.UserApi.RegisterUser(context, newUserRequestBodyDTO)

	jsonResult, err := json.Marshal(result)
	if err != nil {
		cause := errutil.ErrJsonEncode.Errorf("RegisterUserHandler - error encoding response body: %w", err)
		return errutil.ToAPIGatewayProxyResponse(context, errutil.ErrJsonEncode.Errorf(
			"RegisterUserHandler - error encoding response body: %w", cause))
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

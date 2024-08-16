package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	"realworld-aws-lambda-dynamodb-golang/cmd/functions"
	"realworld-aws-lambda-dynamodb-golang/internal/api"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
)
import "github.com/aws/aws-lambda-go/events"

const handlerName = "RegisterUserHandler"

func Handler(context context.Context, request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {

	newUserRequestBodyDTO, errResponse := api.ParseBodyAs[dto.NewUserRequestBodyDTO](context, request, handlerName)

	if errResponse != nil {
		return *errResponse
	}

	result, err := functions.UserApi.RegisterUser(context, *newUserRequestBodyDTO)

	if err != nil {
		return api.ToErrorAPIGatewayProxyResponse(context, err, handlerName)
	}

	return api.ToSuccessAPIGatewayProxyResponse(context, result, handlerName)
}

func main() {
	lambda.Start(Handler)
}

package api

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/uuid"
	"realworld-aws-lambda-dynamodb-golang/internal/security"
)

type AuthenticatedHandlerFn func(context.Context, events.APIGatewayProxyRequest, uuid.UUID) events.APIGatewayProxyResponse

func StartAuthenticatedHandler(handlerToWrap AuthenticatedHandlerFn) {
	lambda.Start(func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		userId, response := security.GetLoggedInUser(request)
		if response != nil {
			return *response, nil
		}
		return handlerToWrap(ctx, request, userId), nil
	})
}

type OptionallyAuthenticatedHandlerFn func(context.Context, events.APIGatewayProxyRequest, *uuid.UUID) events.APIGatewayProxyResponse

func StartOptionallyAuthenticatedHandler(handlerToWrap OptionallyAuthenticatedHandlerFn) {
	lambda.Start(func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		userId, response := security.GetOptionalLoggedInUser(request)
		if response != nil {
			return *response, nil
		}
		return handlerToWrap(ctx, request, userId), nil
	})
}

package api

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/uuid"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/security"
)

type AuthenticatedHandlerFn func(context.Context, events.APIGatewayProxyRequest, uuid.UUID, domain.Token) events.APIGatewayProxyResponse

func StartAuthenticatedHandler(handlerToWrap AuthenticatedHandlerFn) {
	lambda.Start(func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		userId, token, response := security.GetLoggedInUser(ctx, request)
		if response != nil {
			return *response, nil
		}
		return handlerToWrap(ctx, request, userId, token), nil
	})
}

type OptionallyAuthenticatedHandlerFn func(context.Context, events.APIGatewayProxyRequest, *uuid.UUID, *domain.Token) events.APIGatewayProxyResponse

func StartOptionallyAuthenticatedHandler(handlerToWrap OptionallyAuthenticatedHandlerFn) {
	lambda.Start(func(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		userId, token, response := security.GetOptionalLoggedInUser(ctx, request)
		if response != nil {
			return *response, nil
		}
		return handlerToWrap(ctx, request, userId, token), nil
	})
}

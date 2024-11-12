package api

import (
	"context"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/security"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/uuid"
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

// HTTP-specific middleware types and functions

type AuthenticatedHandlerHTTP func(w http.ResponseWriter, r *http.Request, userId uuid.UUID, token domain.Token)

func StartAuthenticatedHandlerHTTP(handlerToWrap AuthenticatedHandlerHTTP) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userId, token, ok := security.GetLoggedInUserHTTP(ctx, w, r)
		if !ok {
			return
		}
		handlerToWrap(w, r, userId, token)
	}
}

type OptionallyAuthenticatedHandlerHTTP func(w http.ResponseWriter, r *http.Request, userId *uuid.UUID, token *domain.Token)

func StartOptionallyAuthenticatedHandlerHTTP(handlerToWrap OptionallyAuthenticatedHandlerHTTP) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userId, token, ok := security.GetOptionalLoggedInUserHTTP(ctx, w, r)
		if !ok {
			return
		}
		handlerToWrap(w, r, userId, token)
	}
}

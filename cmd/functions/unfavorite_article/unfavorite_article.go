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
	slug, response := api.GetPathParam(context, request, "slug")

	if response != nil {
		return *response
	}

	result, err := functions.ArticleApi.UnfavoriteArticle(context, userId, slug)

	if err != nil {
		return api.ToErrorAPIGatewayProxyResponse(context, err, "UnFavoriteArticleHandler")
	}

	return api.ToSuccessAPIGatewayProxyResponse(context, result, "UnFavoriteArticleHandler")
}

func main() {
	api.StartAuthenticatedHandler(Handler)
}

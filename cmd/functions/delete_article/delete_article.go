package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"realworld-aws-lambda-dynamodb-golang/cmd/functions"
	"realworld-aws-lambda-dynamodb-golang/internal/api"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
)

const handlerName = "DeleteArticleHandler"

func Handler(context context.Context, request events.APIGatewayProxyRequest, userId uuid.UUID, _ domain.Token) events.APIGatewayProxyResponse {
	// it's a bit annoying that this could fail even tho the path is required for this endpoint to match...
	slug, response := api.GetPathParam(context, request, "slug", handlerName)

	if response != nil {
		return *response
	}

	err := functions.ArticleApi.DeleteArticle(context, userId, slug)

	if err != nil {
		// ToDo @ender handle article not found and forbidden
		return api.ToInternalServerError(context, err)
	}

	return api.ToSuccessAPIGatewayProxyResponse(context, nil, handlerName)
}

func main() {
	api.StartAuthenticatedHandler(Handler)

}

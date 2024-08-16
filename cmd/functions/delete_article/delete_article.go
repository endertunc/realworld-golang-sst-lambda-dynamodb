package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"realworld-aws-lambda-dynamodb-golang/cmd/functions"
	"realworld-aws-lambda-dynamodb-golang/internal/api"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
)

func Handler(context context.Context, request events.APIGatewayProxyRequest, userId uuid.UUID) events.APIGatewayProxyResponse {
	// it's a bit annoying that this could fail even tho the path is required for this endpoint to match...
	slug, response := api.GetPathParam(context, request, "slug")

	if response != nil {
		return *response
	}

	// ToDo @ender errors returned from service layer is ignored in all handlers
	err := functions.ArticleApi.DeleteArticle(context, userId, slug)

	if err != nil {
		return errutil.ToAPIGatewayProxyResponse(context, err)
	}

	return api.ToSuccessAPIGatewayProxyResponse(context, nil, "DeleteArticleHandler")
}

func main() {
	api.StartAuthenticatedHandler(Handler)

}

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
	slug, response := api.GetPathParam(context, request, "slug")

	if response != nil {
		return *response
	}
	commentIdAsString, response := api.GetPathParam(context, request, "id")

	if response != nil {
		return *response
	}

	commentId, err := uuid.Parse(commentIdAsString)
	if err != nil {
		message := "commentId path parameter must be a valid UUID"
		response := api.ToErrorAPIGatewayProxyResponse(context, errutil.BadRequestError("id", message), "")
		return response
	}

	// ToDo @ender errors returned from service layer is ignored in all handlers
	err = functions.ArticleApi.DeleteComment(context, userId, slug, commentId)
	if err != nil {
		return errutil.ToAPIGatewayProxyResponse(context, err)
	}

	return api.ToSuccessAPIGatewayProxyResponse(context, nil, "DeleteCommentHandler")
}

func main() {
	api.StartAuthenticatedHandler(Handler)

}

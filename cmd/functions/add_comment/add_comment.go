package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"realworld-aws-lambda-dynamodb-golang/cmd/functions"
	"realworld-aws-lambda-dynamodb-golang/internal/api"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
)

const handlerName = "AddCommentHandler"

func Handler(context context.Context, request events.APIGatewayProxyRequest, userId uuid.UUID) events.APIGatewayProxyResponse {
	// it's a bit annoying that this could fail even tho the path is required for this endpoint to match...
	slug, errResponse := api.GetPathParam(context, request, "slug")

	if errResponse != nil {
		return *errResponse
	}

	addCommentRequestBodyDTO, errResponse := api.ParseBodyAs[dto.AddCommentRequestBodyDTO](context, request, handlerName)

	if errResponse != nil {
		return *errResponse
	}

	// ToDo @ender - errors from api layer is ignored in all handlers...
	result, err := functions.ArticleApi.AddComment(context, userId, slug, *addCommentRequestBodyDTO)

	if err != nil {
		return errutil.ToAPIGatewayProxyResponse(context, err)
	}

	return api.ToSuccessAPIGatewayProxyResponse(context, result, handlerName)
}

func main() {
	api.StartAuthenticatedHandler(Handler)
}

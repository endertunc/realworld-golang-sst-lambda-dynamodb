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

const handlerName = "CreateArticleHandler"

func Handler(context context.Context, request events.APIGatewayProxyRequest, userId uuid.UUID) events.APIGatewayProxyResponse {
	createArticleRequestBodyDTO, errResponse := api.ParseBodyAs[dto.CreateArticleRequestBodyDTO](context, request, handlerName)

	if errResponse != nil {
		return *errResponse
	}

	// ToDo @ender - errors from api layer is ignored in all handlers...
	result, err := functions.ArticleApi.CreateArticle(context, userId, *createArticleRequestBodyDTO)

	if err != nil {
		return errutil.ToAPIGatewayProxyResponse(context, err)
	}

	return api.ToSuccessAPIGatewayProxyResponse(context, result, handlerName)
}

func main() {
	api.StartAuthenticatedHandler(Handler)
}

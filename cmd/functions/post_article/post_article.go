package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/google/uuid"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/cmd/functions"
	"realworld-aws-lambda-dynamodb-golang/internal/api"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
)

const handlerName = "CreateArticleHandler"

func init() {
	http.Handle("POST /api/articles", api.StartAuthenticatedHandlerHTTP(HandlerHTTP))
}

func HandlerHTTP(w http.ResponseWriter, r *http.Request, userId uuid.UUID, _ domain.Token) {
	ctx := r.Context()

	createArticleRequestBodyDTO, ok := api.ParseAndValidateBody[dto.CreateArticleRequestBodyDTO](ctx, w, r)

	if !ok {
		return
	}

	result, err := functions.ArticleApi.CreateArticle(ctx, userId, *createArticleRequestBodyDTO)

	if err != nil {
		api.ToInternalServerHTTPError(w, err)
		return
	}

	api.ToSuccessHTTPResponse(w, result)
}

func Handler(context context.Context, request events.APIGatewayProxyRequest, userId uuid.UUID, _ domain.Token) events.APIGatewayProxyResponse {
	createArticleRequestBodyDTO, errResponse := api.ParseBodyAs[dto.CreateArticleRequestBodyDTO](context, request)

	if errResponse != nil {
		return *errResponse
	}

	result, err := functions.ArticleApi.CreateArticle(context, userId, *createArticleRequestBodyDTO)

	if err != nil {
		return api.ToInternalServerError(context, err)
	} else {
		return api.ToSuccessAPIGatewayProxyResponse(context, result, handlerName)
	}
}

func main() {
	lambda.Start(httpadapter.NewV2(http.DefaultServeMux).ProxyWithContext)
}

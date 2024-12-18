package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/google/uuid"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/cmd/functions"
	"realworld-aws-lambda-dynamodb-golang/internal/api"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
)

func init() {
	h := api.WithMiddlewares(api.OptionallyAuthenticatedHandler(handler), api.DefaultMiddlewares)
	http.Handle("GET /api/articles/{slug}", h)
}

func handler(w http.ResponseWriter, r *http.Request, userId *uuid.UUID, _ *domain.Token) {
	functions.ArticleApi.GetArticle(w, r, userId)
}

func main() {
	lambda.Start(httpadapter.NewV2(http.DefaultServeMux).ProxyWithContext)
}

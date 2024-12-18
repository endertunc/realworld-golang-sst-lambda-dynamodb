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
	h := api.WithMiddlewares(api.AuthenticatedHandler(handler), api.DefaultMiddlewares)
	http.Handle("DELETE /api/articles/{slug}", h)
}

func handler(w http.ResponseWriter, r *http.Request, userId uuid.UUID, _ domain.Token) {
	functions.ArticleApi.DeleteArticle(w, r, userId)
}

func main() {
	lambda.Start(httpadapter.NewV2(http.DefaultServeMux).ProxyWithContext)
}

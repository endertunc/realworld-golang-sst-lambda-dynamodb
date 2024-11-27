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
	http.Handle("DELETE /api/articles/{slug}/favorite", api.StartAuthenticatedHandlerHTTP(handler))
}

func handler(w http.ResponseWriter, r *http.Request, userId uuid.UUID, _ domain.Token) {
	functions.ArticleApi.UnfavoriteArticle(r.Context(), w, r, userId)
}

func main() {
	lambda.Start(httpadapter.NewV2(http.DefaultServeMux).ProxyWithContext)
}

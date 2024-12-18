package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/cmd/functions"
	"realworld-aws-lambda-dynamodb-golang/internal/api"
)

func init() {
	h := api.WithMiddlewares(http.HandlerFunc(handler), api.DefaultMiddlewares)
	http.Handle("GET /api/tags", h)
}

func handler(w http.ResponseWriter, r *http.Request) {
	functions.ArticleApi.GetTags(w, r)
}

func main() {
	lambda.Start(httpadapter.NewV2(http.DefaultServeMux).ProxyWithContext)
}

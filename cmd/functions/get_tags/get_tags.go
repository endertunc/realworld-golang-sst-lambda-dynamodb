package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/cmd/functions"
	"realworld-aws-lambda-dynamodb-golang/internal/api"
)

func init() {
	http.Handle("GET /api/tags", api.RequestLoggerMiddleware(http.HandlerFunc(handler)))
}

func handler(w http.ResponseWriter, r *http.Request) {
	functions.ArticleApi.GetTags(r.Context(), w)
}

func main() {
	lambda.Start(httpadapter.NewV2(http.DefaultServeMux).ProxyWithContext)
}

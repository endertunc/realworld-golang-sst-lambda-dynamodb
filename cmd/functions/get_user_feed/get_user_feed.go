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
	http.Handle("GET /api/articles/feed", api.StartAuthenticatedHandlerHTTP(handler))
}

func handler(w http.ResponseWriter, r *http.Request, userId uuid.UUID, token domain.Token) {
	functions.UserFeedApi.FetchUserFeed(r.Context(), w, r, userId)
}

func main() {
	lambda.Start(httpadapter.NewV2(http.DefaultServeMux).ProxyWithContext)
}

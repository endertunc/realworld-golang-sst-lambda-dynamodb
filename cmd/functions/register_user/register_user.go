package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	sloghttp "github.com/samber/slog-http"
	"log/slog"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/cmd/functions"
)

func init() {
	http.Handle("POST /api/users", sloghttp.New(slog.Default())(http.HandlerFunc(handler)))
}

func handler(w http.ResponseWriter, r *http.Request) {
	functions.UserApi.RegisterUser(r.Context(), w, r)
	return
}

func main() {
	lambda.Start(httpadapter.NewV2(http.DefaultServeMux).ProxyWithContext)
}

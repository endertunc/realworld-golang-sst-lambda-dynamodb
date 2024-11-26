package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/cmd/functions"
)

func init() {
	http.Handle("POST /api/users/login", http.HandlerFunc(handler))
}

func handler(w http.ResponseWriter, r *http.Request) {
	functions.UserApi.LoginUser(r.Context(), w, r)
	return
}

func main() {
	lambda.Start(httpadapter.NewV2(http.DefaultServeMux).ProxyWithContext)
}

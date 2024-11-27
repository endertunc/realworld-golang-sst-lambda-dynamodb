package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"io"
	"net/http"
)

func main() {

	http.HandleFunc("GET /api/{username}", func(w http.ResponseWriter, r *http.Request) {
		r.PathValue("username")
		_, _ = io.WriteString(w, "Hello")
	})

	lambda.Start(httpadapter.NewV2(http.DefaultServeMux).ProxyWithContext)
}

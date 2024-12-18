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
	handler := api.WithMiddlewares(api.AuthenticatedHandler(handler), api.DefaultMiddlewares)
	http.Handle("POST /api/articles/{slug}/comments", handler)
}

// refactor: apply middlewares explicitly
// feat: propagate requestId to all log lines
// refactor: don't pass ctx as a separate parameter to api level. let api level use ctx from request.
// chore: clean up test/unused handlers
func handler(w http.ResponseWriter, r *http.Request, userId uuid.UUID, _ domain.Token) {
	functions.CommentApi.AddComment(w, r, userId)
}

func main() {
	lambda.Start(httpadapter.NewV2(http.DefaultServeMux).ProxyWithContext)
}

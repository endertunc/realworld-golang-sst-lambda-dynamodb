package main

import (
	"errors"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/google/uuid"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/cmd/functions"
	"realworld-aws-lambda-dynamodb-golang/internal/api"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
)

func init() {
	http.Handle("GET /api/articles", api.StartOptionallyAuthenticatedHandlerHTTP(HandlerHTTP))
}

func HandlerHTTP(w http.ResponseWriter, r *http.Request, userId *uuid.UUID, _ *domain.Token) {
	ctx := r.Context()

	// Get author from query
	author, _ := api.GetOptionalStringQueryParamHTTP(w, r, "author")

	if author != nil {
		result, err := functions.ArticleApi.ListArticles(ctx, userId, author, 10, nil)
		if err != nil {
			if errors.Is(err, errutil.ErrUserNotFound) {
				api.ToSimpleHTTPError(w, http.StatusNotFound, "author not found")
				return
			}
			api.ToInternalServerHTTPError(w, err)
			return
		}

		// Success response
		api.ToSuccessHTTPResponse(w, result)
		return
	}

	return
}

func main() {
	lambda.Start(httpadapter.NewV2(http.DefaultServeMux).ProxyWithContext)
}

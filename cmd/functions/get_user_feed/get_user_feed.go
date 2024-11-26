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

var paginationConfig = api.GetPaginationConfig()

func handler(w http.ResponseWriter, r *http.Request, userId uuid.UUID, token domain.Token) {
	ctx := r.Context()

	limit, ok := api.GetIntQueryParamOrDefaultHTTP(ctx, w, r, "limit", paginationConfig.DefaultLimit, &paginationConfig.MinLimit, &paginationConfig.MaxLimit)

	if !ok {
		return
	}

	nextPageToken, ok := api.GetOptionalStringQueryParamHTTP(w, r, "offset")

	if !ok {
		return
	}

	result, err := functions.UserFeedApi.FetchUserFeed(ctx, userId, limit, nextPageToken)
	if err != nil {
		api.ToInternalServerHTTPError(w, err)
		return
	}

	api.ToSuccessHTTPResponse(w, result)
	return
}

func main() {
	lambda.Start(httpadapter.NewV2(http.DefaultServeMux).ProxyWithContext)
}

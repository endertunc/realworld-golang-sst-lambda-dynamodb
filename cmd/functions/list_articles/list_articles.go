package main

import (
	"context"
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
	http.Handle("GET /api/articles", api.StartOptionallyAuthenticatedHandlerHTTP(handler))
}

var paginationConfig = api.GetPaginationConfig()

func handler(w http.ResponseWriter, r *http.Request, userId *uuid.UUID, _ *domain.Token) {
	ctx := r.Context()
	limit, offset, listArticlesQueryOptions, ok := extractArticleListRequestParameters(ctx, w, r)
	if !ok {
		return
	}
	result, err := functions.ArticleApi.ListArticles(ctx, userId, listArticlesQueryOptions, limit, offset)
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

func extractArticleListRequestParameters(ctx context.Context, w http.ResponseWriter, r *http.Request) (int, *string, api.ListArticlesQueryOptions, bool) {
	limit, ok := api.GetIntQueryParamOrDefaultHTTP(ctx, w, r, "limit", paginationConfig.DefaultLimit, &paginationConfig.MinLimit, &paginationConfig.MaxLimit)
	if !ok {
		return 0, nil, api.ListArticlesQueryOptions{}, ok
	}
	offset, ok := api.GetOptionalStringQueryParamHTTP(w, r, "offset")
	if !ok {
		return 0, nil, api.ListArticlesQueryOptions{}, ok
	}
	// ToDo @ender errors should NOT be ignored here
	author, ok := api.GetOptionalStringQueryParamHTTP(w, r, "author")
	if !ok {
		return 0, nil, api.ListArticlesQueryOptions{}, ok
	}
	favoritedBy, ok := api.GetOptionalStringQueryParamHTTP(w, r, "favorited")
	if !ok {
		return 0, nil, api.ListArticlesQueryOptions{}, ok
	}
	tag, ok := api.GetOptionalStringQueryParamHTTP(w, r, "tag")
	if !ok {
		return 0, nil, api.ListArticlesQueryOptions{}, ok
	}
	listArticlesQueryOptions := api.ListArticlesQueryOptions{
		Author:      author,
		FavoritedBy: favoritedBy,
		Tag:         tag,
	}

	return limit, offset, listArticlesQueryOptions, ok
}

func main() {
	lambda.Start(httpadapter.NewV2(http.DefaultServeMux).ProxyWithContext)
}

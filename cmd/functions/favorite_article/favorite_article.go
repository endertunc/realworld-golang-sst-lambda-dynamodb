package main

import (
	"context"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/cmd/functions"
	"realworld-aws-lambda-dynamodb-golang/internal/api"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
)

const handlerName = "FavoriteArticleHandler"

func init() {
	http.Handle("POST /api/articles/{slug}/favorite", api.StartAuthenticatedHandlerHTTP(HandlerHTTP))
}

func HandlerHTTP(w http.ResponseWriter, r *http.Request, userId uuid.UUID, token domain.Token) {
	ctx := r.Context()

	slug, ok := api.GetPathParamHTTP(ctx, w, r, "slug")
	if !ok {
		return
	}

	result, err := functions.ArticleApi.FavoriteArticle(ctx, userId, slug)

	if err != nil {
		if errors.Is(err, errutil.ErrArticleNotFound) {
			slog.DebugContext(ctx, "article not found", slog.String("slug", slug), slog.Any("error", err))
			api.ToSimpleHTTPError(w, http.StatusNotFound, "article not found")
			return
		} else if errors.Is(err, errutil.ErrAlreadyFavorited) {
			slog.DebugContext(ctx, "article already favorited", slog.String("slug", slug), slog.String("userId", userId.String()))
			api.ToSimpleHTTPError(w, http.StatusConflict, "article already favorited")
			return
		} else {
			api.ToInternalServerHTTPError(w, err)
			return
		}
	}

	api.ToSuccessHTTPResponse(w, result)
}

func Handler(context context.Context, request events.APIGatewayProxyRequest, userId uuid.UUID, _ domain.Token) events.APIGatewayProxyResponse {
	// it's a bit annoying that this could fail even tho the path is required for this endpoint to match...
	slug, response := api.GetPathParam(context, request, "slug", handlerName)

	if response != nil {
		return *response
	}

	result, err := functions.ArticleApi.FavoriteArticle(context, userId, slug)

	if err != nil {
		if errors.Is(err, errutil.ErrArticleNotFound) {
			slog.DebugContext(context, "article not found", slog.String("slug", slug))
			return api.ToSimpleError(context, http.StatusNotFound, "article not found")
		} else if errors.Is(err, errutil.ErrAlreadyFavorited) {
			slog.DebugContext(context, "article already favorited", slog.String("slug", slug), slog.String("userId", userId.String()))
			return api.ToSimpleError(context, http.StatusConflict, "article already favorited")
		} else {
			return api.ToInternalServerError(context, err)
		}
	} else {
		return api.ToSuccessAPIGatewayProxyResponse(context, result, handlerName)
	}

}

func main() {
	lambda.Start(httpadapter.NewV2(http.DefaultServeMux).ProxyWithContext)
}

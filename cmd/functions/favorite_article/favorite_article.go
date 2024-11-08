package main

import (
	"context"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/cmd/functions"
	"realworld-aws-lambda-dynamodb-golang/internal/api"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
)

const handlerName = "FavoriteArticleHandler"

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
	api.StartAuthenticatedHandler(Handler)
}

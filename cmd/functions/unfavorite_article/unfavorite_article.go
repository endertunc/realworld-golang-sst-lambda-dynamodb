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

const handlerName = "UnfavoriteArticleHandler"

func Handler(context context.Context, request events.APIGatewayProxyRequest, userId uuid.UUID, _ domain.Token) events.APIGatewayProxyResponse {
	// it's a bit annoying that this could fail even tho the path is required for this endpoint to match...
	slug, response := api.GetPathParam(context, request, "slug", handlerName)

	if response != nil {
		return *response
	}

	result, err := functions.ArticleApi.UnfavoriteArticle(context, userId, slug)

	if err != nil {
		if errors.Is(err, errutil.ErrArticleNotFound) {
			slog.DebugContext(context, "article not found", slog.String("slug", slug))
			return api.ToSimpleError(context, 404, "article not found")
		} else if errors.Is(err, errutil.ErrAlreadyUnfavorited) {
			slog.DebugContext(context, "article is already unfavorited", slog.String("slug", slug), slog.String("userId", userId.String()))
			return api.ToSimpleError(context, http.StatusConflict, "article is already unfavorited")
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

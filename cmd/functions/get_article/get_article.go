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

const handlerName = "GetArticleHandler"

func Handler(ctx context.Context, request events.APIGatewayProxyRequest, userId *uuid.UUID, _ *domain.Token) events.APIGatewayProxyResponse {
	// it's a bit annoying that this could fail even tho the path is required for this endpoint to match...
	slug, response := api.GetPathParam(ctx, request, "slug", handlerName)

	if response != nil {
		return *response
	}

	result, err := functions.ArticleApi.GetArticle(ctx, userId, slug)

	if err != nil {
		if errors.Is(err, errutil.ErrArticleNotFound) {
			slog.DebugContext(ctx, "article not found", slog.String("slug", slug))
			return api.ToSimpleError(ctx, http.StatusNotFound, "article not found")
		} else {
			return api.ToInternalServerError(ctx, err)
		}
	} else {
		return api.ToSuccessAPIGatewayProxyResponse(ctx, result, handlerName)
	}

}

func main() {
	api.StartOptionallyAuthenticatedHandler(Handler)
}

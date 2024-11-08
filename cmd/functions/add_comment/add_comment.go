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
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
)

const handlerName = "AddCommentHandler"

func Handler(ctx context.Context, request events.APIGatewayProxyRequest, userId uuid.UUID, _ domain.Token) events.APIGatewayProxyResponse {
	// it's a bit annoying that this could fail even tho the path is required for this endpoint to match...
	slug, errResponse := api.GetPathParam(ctx, request, "slug", handlerName)

	if errResponse != nil {
		return *errResponse
	}

	addCommentRequestBodyDTO, errResponse := api.ParseBodyAs[dto.AddCommentRequestBodyDTO](ctx, request)

	if errResponse != nil {
		return *errResponse
	}

	result, err := functions.ArticleApi.AddComment(ctx, userId, slug, *addCommentRequestBodyDTO)

	if err != nil {
		if errors.Is(err, errutil.ErrArticleNotFound) {
			slog.DebugContext(ctx, "article not found", slog.String("slug", slug), slog.Any("error", err))
			return api.ToSimpleError(ctx, http.StatusNotFound, "article not found")
		}

		return api.ToInternalServerError(ctx, err)
	} else {
		return api.ToSuccessAPIGatewayProxyResponse(ctx, result, handlerName)
	}
}

func main() {
	api.StartAuthenticatedHandler(Handler)
}

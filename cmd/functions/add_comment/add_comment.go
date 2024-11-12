package main

import (
	"context"
	"errors"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"log/slog"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/cmd/functions"
	"realworld-aws-lambda-dynamodb-golang/internal/api"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
)

const handlerName = "AddCommentHandler"

func init() {
	http.Handle("POST /api/articles/{slug}/comments", api.StartAuthenticatedHandlerHTTP(HandlerHTTP))
}

func HandlerHTTP(w http.ResponseWriter, r *http.Request, userId uuid.UUID, token domain.Token) {
	ctx := r.Context()

	// Get slug from path
	slug, ok := api.GetPathParamHTTP(ctx, w, r, "slug")
	if !ok {
		return
	}

	// Parse request body
	addCommentRequestBodyDTO, ok := api.ParseBodyAsHTTP[dto.AddCommentRequestBodyDTO](ctx, w, r)
	if !ok {
		return
	}

	// Add comment
	result, err := functions.ArticleApi.AddComment(ctx, userId, slug, *addCommentRequestBodyDTO)

	if err != nil {
		if errors.Is(err, errutil.ErrArticleNotFound) {
			slog.DebugContext(ctx, "article not found", slog.String("slug", slug), slog.Any("error", err))
			api.ToSimpleHTTPError(w, http.StatusNotFound, "article not found")
			return
		}

		api.ToInternalServerHTTPError(w, err)
		return
	}

	// Success response
	api.ToSuccessHTTPResponse(w, result)
}

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
	lambda.Start(httpadapter.NewV2(http.DefaultServeMux).ProxyWithContext)
}

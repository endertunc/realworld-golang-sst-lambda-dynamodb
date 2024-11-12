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

const handlerName = "DeleteCommentHandler"

func init() {
	http.Handle("DELETE /api/articles/{slug}/comments/{id}", api.StartAuthenticatedHandlerHTTP(HandlerHTTP))
}

func HandlerHTTP(w http.ResponseWriter, r *http.Request, userId uuid.UUID, token domain.Token) {
	ctx := r.Context()

	slug, ok := api.GetPathParamHTTP(ctx, w, r, "slug")
	if !ok {
		return
	}

	commentIdAsString, ok := api.GetPathParamHTTP(ctx, w, r, "id")
	if !ok {
		return
	}

	commentId, err := uuid.Parse(commentIdAsString)
	if err != nil {
		slog.DebugContext(ctx, "invalid commentId path param", slog.String("commentId", commentIdAsString), slog.Any("error", err))
		api.ToSimpleHTTPError(w, http.StatusBadRequest, "commentId path parameter must be a valid UUID")
		return
	}

	err = functions.ArticleApi.DeleteComment(ctx, userId, slug, commentId)
	if err != nil {
		if errors.Is(err, errutil.ErrCommentNotFound) {
			slog.DebugContext(ctx, "comment not found", slog.String("slug", slug), slog.String("commentId", commentId.String()), slog.Any("error", err))
			api.ToSimpleHTTPError(w, http.StatusNotFound, "comment not found")
			return
		} else if errors.Is(err, errutil.ErrArticleNotFound) {
			slog.DebugContext(ctx, "article not found", slog.String("slug", slug), slog.String("commentId", commentId.String()), slog.Any("error", err))
			api.ToSimpleHTTPError(w, http.StatusNotFound, "article not found")
			return
		} else if errors.Is(err, errutil.ErrCantDeleteOthersComment) {
			slog.DebugContext(ctx, "can't delete other's comment", slog.String("slug", slug), slog.String("commentId", commentId.String()), slog.Any("error", err))
			api.ToSimpleHTTPError(w, http.StatusForbidden, "forbidden")
			return
		} else {
			api.ToInternalServerHTTPError(w, err)
			return
		}
	}

	api.ToSuccessHTTPResponse(w, nil)
}

func Handler(ctx context.Context, request events.APIGatewayProxyRequest, userId uuid.UUID, _ domain.Token) events.APIGatewayProxyResponse {
	slug, response := api.GetPathParam(ctx, request, "slug", handlerName)

	if response != nil {
		return *response
	}
	commentIdAsString, response := api.GetPathParam(ctx, request, "id", handlerName)

	if response != nil {
		return *response
	}

	commentId, err := uuid.Parse(commentIdAsString)
	if err != nil {
		slog.DebugContext(ctx, "invalid commentId path param", slog.String("commentId", commentIdAsString), slog.Any("error", err))
		return api.ToSimpleError(ctx, http.StatusBadRequest, "commentId path parameter must be a valid UUID")
	}

	err = functions.ArticleApi.DeleteComment(ctx, userId, slug, commentId)
	if err != nil {
		// ToDo slog...
		if errors.Is(err, errutil.ErrCommentNotFound) {
			slog.DebugContext(ctx, "comment not found", slog.String("slug", slug), slog.String("commentId", commentId.String()), slog.Any("error", err))
			return api.ToSimpleError(ctx, http.StatusNotFound, "comment not found")
		} else if errors.Is(err, errutil.ErrArticleNotFound) {
			slog.DebugContext(ctx, "article not found", slog.String("slug", slug), slog.String("commentId", commentId.String()), slog.Any("error", err))
			return api.ToSimpleError(ctx, http.StatusNotFound, "article not found")
		} else if errors.Is(err, errutil.ErrCantDeleteOthersComment) {
			slog.DebugContext(ctx, "can't delete other's comment", slog.String("slug", slug), slog.String("commentId", commentId.String()), slog.Any("error", err))
			return api.ToSimpleError(ctx, http.StatusForbidden, "forbidden")
		} else {
			return api.ToInternalServerError(ctx, err)
		}
	} else {
		return api.ToSuccessAPIGatewayProxyResponse(ctx, nil, handlerName)
	}

}

func main() {
	lambda.Start(httpadapter.NewV2(http.DefaultServeMux).ProxyWithContext)
}

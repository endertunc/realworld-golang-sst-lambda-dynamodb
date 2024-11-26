package main

import (
	"errors"
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

func init() {
	http.Handle("DELETE /api/articles/{slug}/comments/{id}", api.StartAuthenticatedHandlerHTTP(handler))
}

func handler(w http.ResponseWriter, r *http.Request, userId uuid.UUID, token domain.Token) {
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

	err = functions.CommentApi.DeleteComment(ctx, userId, slug, commentId)
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
	return
}

func main() {
	lambda.Start(httpadapter.NewV2(http.DefaultServeMux).ProxyWithContext)
}

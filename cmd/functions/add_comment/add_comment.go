package main

import (
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

	"github.com/google/uuid"
)

func init() {
	http.Handle("POST /api/articles/{slug}/comments", api.StartAuthenticatedHandlerHTTP(HandlerHTTP))
}

func HandlerHTTP(w http.ResponseWriter, r *http.Request, userId uuid.UUID, token domain.Token) {
	ctx := r.Context()

	// Get slug from the request path
	slug, ok := api.GetPathParamHTTP(ctx, w, r, "slug")
	if !ok {
		return
	}

	// Parse request body
	addCommentRequestBodyDTO, ok := api.ParseAndValidateBody[dto.AddCommentRequestBodyDTO](ctx, w, r)
	if !ok {
		return
	}

	// Add comment
	result, err := functions.CommentApi.AddComment(ctx, userId, slug, *addCommentRequestBodyDTO)

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

func main() {
	lambda.Start(httpadapter.NewV2(http.DefaultServeMux).ProxyWithContext)
}

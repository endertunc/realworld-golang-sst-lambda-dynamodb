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
	http.Handle("POST /api/profiles/{username}/follow", api.StartAuthenticatedHandlerHTTP(handler))
}

func handler(w http.ResponseWriter, r *http.Request, userId uuid.UUID, token domain.Token) {
	ctx := r.Context()

	username, ok := api.GetPathParamHTTP(ctx, w, r, "username")
	if !ok {
		return
	}

	result, err := functions.ProfileApi.FollowUserByUsername(ctx, userId, username)

	if err != nil {
		if errors.Is(err, errutil.ErrUserNotFound) {
			slog.DebugContext(ctx, "user to follow not found", slog.Any("username", username), slog.Any("error", err))
			api.ToSimpleHTTPError(w, http.StatusNotFound, "user not found")
			return
		} else if errors.Is(err, errutil.ErrCantFollowYourself) {
			slog.DebugContext(ctx, "user tried to follow itself", slog.Any("username", username), slog.String("userId", userId.String()), slog.Any("error", err))
			api.ToSimpleHTTPError(w, http.StatusBadRequest, "cannot follow yourself")
			return
		} else {
			api.ToInternalServerHTTPError(w, err)
			return
		}
	}

	api.ToSuccessHTTPResponse(w, result)
	return
}

func main() {
	lambda.Start(httpadapter.NewV2(http.DefaultServeMux).ProxyWithContext)
}

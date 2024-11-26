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
	http.Handle("DELETE /api/profiles/{username}/follow", api.StartAuthenticatedHandlerHTTP(handler))
}

func handler(w http.ResponseWriter, r *http.Request, userId uuid.UUID, _ domain.Token) {
	ctx := r.Context()

	username, ok := api.GetPathParamHTTP(ctx, w, r, "username")
	if !ok {
		return
	}

	result, err := functions.ProfileApi.UnfollowUserByUsername(ctx, userId, username)

	if err != nil {
		if errors.Is(err, errutil.ErrUserNotFound) {
			slog.DebugContext(ctx, "user not found", slog.String("username", username))
			api.ToSimpleHTTPError(w, http.StatusNotFound, "user not found")
			return
		} else if errors.Is(err, errutil.ErrCantFollowYourself) {
			slog.DebugContext(ctx, "user tried to unfollow itself", slog.String("username", username), slog.String("userId", userId.String()))
			api.ToSimpleHTTPError(w, http.StatusConflict, "cannot unfollow yourself")
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

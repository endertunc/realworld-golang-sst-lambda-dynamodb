package main

import (
	"errors"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"log/slog"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/cmd/functions"
	"realworld-aws-lambda-dynamodb-golang/internal/api"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
)

func init() {
	http.Handle("POST /api/users/login", http.HandlerFunc(handler))
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	loginRequestBodyDTO, ok := api.ParseAndValidateBody[dto.LoginRequestBodyDTO](ctx, w, r)

	if !ok {
		return
	}

	result, err := functions.UserApi.LoginUser(ctx, *loginRequestBodyDTO)

	if err != nil {
		if errors.Is(err, errutil.ErrUserNotFound) || errors.Is(err, errutil.ErrInvalidPassword) {
			slog.WarnContext(ctx, "invalid credentials", slog.Any("error", err))
			api.ToSimpleHTTPError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		api.ToInternalServerHTTPError(w, err)
		return
	}

	api.ToSuccessHTTPResponse(w, result)
	return
}

func main() {
	lambda.Start(httpadapter.NewV2(http.DefaultServeMux).ProxyWithContext)
}

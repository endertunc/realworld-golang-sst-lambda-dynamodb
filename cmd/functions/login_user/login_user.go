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
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
)
import "github.com/aws/aws-lambda-go/events"

const handlerName = "LoginUserHandler"

func init() {
	http.Handle("POST /api/users/login", http.HandlerFunc(HandlerHTTP))
}

func HandlerHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	loginRequestBodyDTO, ok := api.ParseBodyAsHTTP[dto.LoginRequestBodyDTO](ctx, w, r)

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
}

func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	loginRequestBodyDTO, errResponse := api.ParseBodyAs[dto.LoginRequestBodyDTO](ctx, request)

	if errResponse != nil {
		return *errResponse, nil
	}

	result, err := functions.UserApi.LoginUser(ctx, *loginRequestBodyDTO)

	if err != nil {
		if errors.Is(err, errutil.ErrUserNotFound) || errors.Is(err, errutil.ErrInvalidPassword) {
			slog.WarnContext(ctx, "invalid credentials", slog.Any("error", err))
			return api.ToSimpleError(ctx, http.StatusUnauthorized, "invalid credentials"), nil
		}
		return api.ToInternalServerError(ctx, err), nil
	} else {
		return api.ToSuccessAPIGatewayProxyResponse(ctx, result, handlerName), nil
	}
}

func main() {
	lambda.Start(httpadapter.NewV2(http.DefaultServeMux).ProxyWithContext)
}

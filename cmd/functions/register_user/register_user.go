package main

import (
	"context"
	"errors"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	sloghttp "github.com/samber/slog-http"
	"log/slog"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/cmd/functions"
	"realworld-aws-lambda-dynamodb-golang/internal/api"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/test"

	"github.com/aws/aws-lambda-go/events"
)

const handlerName = "RegisterUserHandler"

func init() {
	http.Handle("POST /api/users", sloghttp.New(slog.Default())(http.HandlerFunc(HandlerHTTP)))
}

func HandlerHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	newUserRequestBodyDTO, ok := api.ParseBodyAsHTTP[dto.NewUserRequestBodyDTO](ctx, w, r)

	if !ok {
		return
	}

	result, err := functions.UserApi.RegisterUser(ctx, *newUserRequestBodyDTO)

	if err != nil {
		if errors.Is(err, errutil.ErrUsernameAlreadyExists) {
			username := newUserRequestBodyDTO.User.Username
			slog.WarnContext(ctx, "username already exists", slog.String("username", username), slog.Any("error", err))
			api.ToSimpleHTTPError(w, http.StatusConflict, "username already exists")
			return
		}

		if errors.Is(err, errutil.ErrEmailAlreadyExists) {
			email := newUserRequestBodyDTO.User.Email
			slog.WarnContext(ctx, "email already exists", slog.String("email", email), slog.Any("error", err))
			api.ToSimpleHTTPError(w, http.StatusConflict, "email already exists")
			return
		}
		api.ToInternalServerHTTPError(w, err)
		return
	}

	api.ToSuccessHTTPResponse(w, result)
}

func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	test.SlogAsJSON(request)

	slog.Info("", slog.Any("request", request))

	newUserRequestBodyDTO, errResponse := api.ParseBodyAs[dto.NewUserRequestBodyDTO](ctx, request)

	if errResponse != nil {
		return *errResponse, nil
	}

	result, err := functions.UserApi.RegisterUser(ctx, *newUserRequestBodyDTO)
	if err != nil {
		if errors.Is(err, errutil.ErrUsernameAlreadyExists) {
			username := newUserRequestBodyDTO.User.Username
			slog.WarnContext(ctx, "username already exists", slog.String("username", username), slog.Any("error", err))
			return api.ToSimpleError(ctx, http.StatusConflict, "username already exists"), nil
		}

		if errors.Is(err, errutil.ErrEmailAlreadyExists) {
			email := newUserRequestBodyDTO.User.Email
			slog.WarnContext(ctx, "email already exists", slog.String("email", email), slog.Any("error", err))
			return api.ToSimpleError(ctx, http.StatusConflict, "email already exists"), nil
		}
		return api.ToInternalServerError(ctx, err), nil
	} else {
		return api.ToSuccessAPIGatewayProxyResponse(ctx, result, handlerName), nil
	}
}

func main() {
	lambda.Start(httpadapter.NewV2(http.DefaultServeMux).ProxyWithContext)
}

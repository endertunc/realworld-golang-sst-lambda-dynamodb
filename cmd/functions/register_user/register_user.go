package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/cmd/functions"
	"realworld-aws-lambda-dynamodb-golang/internal/api"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

const handlerName = "RegisterUserHandler"

func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

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
	lambda.Start(Handler)
}

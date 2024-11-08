package main

import (
	"context"
	"errors"
	"github.com/aws/aws-lambda-go/lambda"
	"log/slog"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/cmd/functions"
	"realworld-aws-lambda-dynamodb-golang/internal/api"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
)
import "github.com/aws/aws-lambda-go/events"

const handlerName = "LoginUserHandler"

func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	loginRequestBodyDTO, errResponse := api.ParseBodyAs[dto.LoginRequestBodyDTO](ctx, request, handlerName)

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

//// create enum
//type ErrType int64
//
//const (
//	Internal ErrType = iota
//	UserNotFound
//	InvalidCredentials
//	DynamodbQuery
//	DynamodbMapping
//)
//
//type ExampleError struct {
//	Type            ErrType
//	Message         string
//	InternalMessage *string
//	Operation       string
//	Cause           error
//}
//
//var (
//	ErrInternal           = ExampleError{Type: Internal, Message: "Internal error"}
//	ErrDynamodbQuery      = ExampleError{Type: DynamodbQuery, Message: "Dynamodb error"}
//	ErrDynamodbMapping    = ExampleError{Type: DynamodbMapping, Message: "Dynamodb error"}
//	ErrUserNotFound       = ExampleError{Type: UserNotFound, Message: "Missing field"}
//	ErrInvalidCredentials = ExampleError{Type: InvalidCredentials, Message: "Missing field"}
//)
//
//func (e ExampleError) Error() string {
//	return e.Message
//}
//
//func (e ExampleError) Unwrap() error {
//	return e.Cause
//}
//
//func (e ExampleError) Is(target error) bool {
//	var t ExampleError
//	ok := errors.As(target, &t)
//	if !ok {
//		return false
//	}
//
//	return e.Type == t.Type
//}

func main() {
	//
	////var e1 error = ExampleError{
	////	Type:    UserNotFound,
	////	Message: "user not found",
	////	Cause:   MissingField,
	////}
	//
	//e2 := ExampleError{
	//	Type:    DynamodbMapping,
	//	Message: "invalid credentials",
	//	Cause:   ExampleError{Type: InvalidCredentials, Message: "Missing field"},
	//}
	//
	//e3 := ExampleError{
	//	Type:    UserNotFound,
	//	Message: "invalid something else",
	//	Cause:   e2,
	//}
	//
	//if errors.Is(e3, ErrInvalidCredentials) {
	//	fmt.Println("e2 is caused by given error")
	//} else {
	//	fmt.Println("e2 is NOT caused given error")
	//}
	//
	//exampleError := ExampleError{}
	//ok := errors.As(e3, &exampleError)
	//if ok && exampleError.Type == UserNotFound {
	//	fmt.Println("e2 is of type ExampleError")
	//} else {
	//	fmt.Println("e2 is NOT of type ExampleError")
	//}

	lambda.Start(Handler)
}

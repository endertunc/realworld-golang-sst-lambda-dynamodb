package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	"realworld-aws-lambda-dynamodb-golang/cmd/functions"
	"realworld-aws-lambda-dynamodb-golang/internal/api"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
)
import "github.com/aws/aws-lambda-go/events"

const handlerName = "LoginUserHandler"

func Handler(context context.Context, request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {

	loginRequestBodyDTO, errResponse := api.ParseBodyAs[dto.LoginRequestBodyDTO](context, request, handlerName)

	if errResponse != nil {
		return *errResponse
	}

	result, err := functions.UserApi.LoginUser(context, *loginRequestBodyDTO)

	if err != nil {
		return api.ToErrorAPIGatewayProxyResponse(context, err, handlerName)
	}

	return api.ToSuccessAPIGatewayProxyResponse(context, result, handlerName)
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

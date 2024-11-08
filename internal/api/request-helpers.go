package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"log/slog"
	"net/http"
	"strconv"
)

func GetPathParam(ctx context.Context, request events.APIGatewayProxyRequest, paramName, handlerName string) (string, *events.APIGatewayProxyResponse) {
	param, ok := request.PathParameters[paramName]
	if !ok {
		message := fmt.Sprintf("path parameter %s is missing", paramName)
		response := ToSimpleError(ctx, http.StatusBadRequest, message)
		//response := ToErrorAPIGatewayProxyResponse(ctx, handlerName, errutil.BadRequestError(paramName, message))
		return "", &response
	}
	return param, nil
}

func ParseBodyAs[T any](ctx context.Context, request events.APIGatewayProxyRequest, handlerName string) (*T, *events.APIGatewayProxyResponse) {
	var out T
	err := json.Unmarshal([]byte(request.Body), &out)
	if err != nil {
		cause := fmt.Errorf("error decoding request body: %w", err)
		slog.WarnContext(ctx, "error decoding request body", slog.Any("error", cause))
		// response:= errutil.ToAPIGatewayProxyResponse(ctx, handlerName, cause)

		// A note on err.Error():
		// json module returns a descriptive enough error message that we can use as is.
		// As far as I can tell, there is not much of risk of leaking sensitive information here.
		//
		// It's unnecessary in this project, but one could use *json.SyntaxError and *json.UnmarshalTypeError
		// to provide more structured error messages.
		// JsonParsingErrorStruct {
		//  Field string
		// 	Message string
		// 	Position int
		//  etc...
		// }
		response := ToSimpleError(ctx, http.StatusBadRequest, err.Error())
		return nil, &response
	}
	return &out, nil
}

// ToDo @ender let's see if we can make this generic using ozzo-validation
func GetQueryParamOrDefault(ctx context.Context, request events.APIGatewayProxyRequest, paramName, handlerName string, defaultValue int) (int, *events.APIGatewayProxyResponse) {
	param, ok := request.QueryStringParameters[paramName]
	if !ok {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(param)
	if err != nil {
		message := fmt.Sprintf("query parameter %s must be a valid integer", paramName)
		response := ToSimpleError(ctx, http.StatusBadRequest, message)
		//response := ToErrorAPIGatewayProxyResponse(ctx, handlerName, errutil.BadRequestError(paramName, message))
		return 0, &response
	}

	// ToDo @ender make sure it's not negative. simply use https://github.com/invopop/validation (fork of ozzo-validation)
	return value, nil
}

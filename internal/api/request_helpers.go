package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
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

func ParseBodyAs[T any](ctx context.Context, request events.APIGatewayProxyRequest) (*T, *events.APIGatewayProxyResponse) {
	var out T
	err := json.Unmarshal([]byte(request.Body), &out)
	if err != nil {
		cause := fmt.Errorf("error decoding request body: %w", err)
		slog.WarnContext(ctx, "error decoding request body", slog.Any("error", cause))

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

func GetOptionalIntQueryParam(
	ctx context.Context,
	request events.APIGatewayProxyRequest,
	paramName string,
	min,
	max *int,
) (*int, *events.APIGatewayProxyResponse) {
	param, ok := request.QueryStringParameters[paramName]
	if !ok {
		return nil, nil
	}

	value, err := strconv.Atoi(param)
	if err != nil {
		message := fmt.Sprintf("query parameter %s must be a valid integer", paramName)
		response := ToSimpleError(ctx, http.StatusBadRequest, message)
		return nil, &response
	}

	if min != nil && value < *min {
		message := fmt.Sprintf("query parameter %s must be greater than or equal to %d", paramName, *min)
		response := ToSimpleError(ctx, http.StatusBadRequest, message)
		return nil, &response
	}

	if max != nil && value > *max {
		message := fmt.Sprintf("query parameter %s must be less than or equal to %d", paramName, *max)
		response := ToSimpleError(ctx, http.StatusBadRequest, message)
		return nil, &response
	}

	return &value, nil
}

func GetIntQueryParamOrDefault(
	ctx context.Context,
	request events.APIGatewayProxyRequest,
	paramName string,
	defaultValue int,
	min,
	max *int,
) (int, *events.APIGatewayProxyResponse) {
	param, response := GetOptionalIntQueryParam(ctx, request, paramName, min, max)
	if response != nil {
		return 0, response
	} else if param == nil {
		return defaultValue, nil
	} else {
		return *param, nil
	}
}

func GetOptionalStringQueryParam(
	ctx context.Context,
	request events.APIGatewayProxyRequest,
	paramName string,
) (*string, *events.APIGatewayProxyResponse) {
	param, ok := request.QueryStringParameters[paramName]
	if !ok {
		return nil, nil
	}

	if strings.TrimSpace(param) == "" {
		message := fmt.Sprintf("query parameter %s cannot be blank", paramName)
		response := ToSimpleError(ctx, http.StatusBadRequest, message)
		return nil, &response
	}

	return &param, nil
}

//func GetOptionalStringQueryParamWithDefault(
//	ctx context.Context,
//	request events.APIGatewayProxyRequest,
//	paramName string,
//	defaultValue string,
//) (string, *events.APIGatewayProxyResponse) {
//	param, response := GetOptionalStringQueryParam(ctx, request, paramName)
//	if response != nil {
//		return "", response
//	} else if param == nil {
//		return defaultValue, nil
//	} else {
//		return *param, nil
//	}
//}
//
//func GetRequiredStringQueryParam(
//	ctx context.Context,
//	request events.APIGatewayProxyRequest,
//	paramName string,
//) (string, *events.APIGatewayProxyResponse) {
//	param, ok := request.QueryStringParameters[paramName]
//	if !ok {
//		message := fmt.Sprintf("query parameter %s is missing", paramName)
//		response := ToSimpleError(ctx, http.StatusBadRequest, message)
//		return "", &response
//	}
//
//	if strings.TrimSpace(param) == "" {
//		message := fmt.Sprintf("query parameter %s cannot be blank", paramName)
//		response := ToSimpleError(ctx, http.StatusBadRequest, message)
//		return "", &response
//	}
//
//	return param, nil
//}

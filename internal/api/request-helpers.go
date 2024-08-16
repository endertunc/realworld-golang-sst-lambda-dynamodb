package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
)

func GetPathParam(c context.Context, request events.APIGatewayProxyRequest, paramName string) (string, *events.APIGatewayProxyResponse) {
	param, ok := request.PathParameters[paramName]
	if !ok {
		message := fmt.Sprintf("path parameter %s is missing", paramName)
		response := ToErrorAPIGatewayProxyResponse(c, errutil.BadRequestError(paramName, message), "")
		return "", &response
	}
	return param, nil
}

func ParseBodyAs[T any](c context.Context, request events.APIGatewayProxyRequest, handlerName string) (*T, *events.APIGatewayProxyResponse) {
	var out T
	err := json.Unmarshal([]byte(request.Body), &out)
	if err != nil {
		cause := errutil.ErrJsonDecode.Errorf("%s - error decoding request body: %w", handlerName, err)
		response := errutil.ToAPIGatewayProxyResponse(c, cause)
		return nil, &response
	}
	return &out, nil
}

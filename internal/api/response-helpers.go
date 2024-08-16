package api

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
)

func ToApiGatewayProxyResponse(context context.Context, result interface{}, handlerName string) events.APIGatewayProxyResponse {
	jsonResult, err := json.Marshal(result)
	if err != nil {
		cause := errutil.ErrJsonEncode.Errorf("%s - error encoding response body: %w", handlerName, err)
		return errutil.ToAPIGatewayProxyResponse(context, cause)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(jsonResult),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}
}

func ToEmptyApiGatewayProxyResponse() events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
	}
}

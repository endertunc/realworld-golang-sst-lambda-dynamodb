package api

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
)

func ToSuccessAPIGatewayProxyResponse(context context.Context, result interface{}, opt string) events.APIGatewayProxyResponse {
	jsonResult, err := json.Marshal(result)
	if err != nil {
		cause := errutil.ErrJsonEncode.Errorf("%s - error encoding response body: %w", opt, err)
		return errutil.ToAPIGatewayProxyResponse(context, errutil.ErrJsonEncode.Errorf(
			"%s - error encoding response body: %w", opt, cause))
	}
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(jsonResult),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}
}

func ToErrorAPIGatewayProxyResponse(context context.Context, err error, opt string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "wake me up inside",
		Headers:    map[string]string{"Content-Type": "application/json"},
	}
}

package api

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
)

func ToSuccessAPIGatewayProxyResponse(context context.Context, body interface{}, handlerName string) events.APIGatewayProxyResponse {
	if body == nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
		}
	}

	bodyJson, err := json.Marshal(body)
	if err != nil {
		cause := errutil.ErrJsonEncode.Errorf("%s - error encoding response body: %w", handlerName, err)
		return errutil.ToAPIGatewayProxyResponse(context, cause)
	}
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(bodyJson),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}
}

func ToErrorAPIGatewayProxyResponse(context context.Context, err error, opt string) events.APIGatewayProxyResponse {
	// ToDo @ender - convert error to response
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "wake me up inside",
		Headers:    map[string]string{"Content-Type": "application/json"},
	}
}

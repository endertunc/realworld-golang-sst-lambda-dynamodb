package api

import (
	"context"
	"encoding/json"
	"fmt"
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
		cause := fmt.Errorf("%w: %w", errutil.ErrJsonEncode, err)
		return ToInternalServerError(context, cause)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(bodyJson),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}
}

//func ToErrorAPIGatewayProxyResponse(context context.Context, handlerName string, error error) events.APIGatewayProxyResponse {
//
//	return errutil.ToAPIGatewayProxyResponse(context, handlerName, error)
//
//	//return events.APIGatewayProxyResponse{
//	//	StatusCode: 200,
//	//	Body:       "wake me up inside",
//	//	Headers:    map[string]string{"Content-Type": "application/json"},
//	//}
//}

package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"realworld-aws-lambda-dynamodb-golang/cmd/functions"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/security"
)

func Handler(context context.Context, request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	// it's a bit annoying that this could fail even tho the path is required for this endpoint to match...
	slug, ok := request.PathParameters["slug"]
	// ToDo Ender how to handle such situations?
	if !ok {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "slug path parameter is missing", // ToDo This is not a json tho...
			Headers:    map[string]string{"Content-Type": "application/json"},
		}
	}

	userId, response := security.GetLoggedInUser(request)
	if response != nil {
		return *response
	}

	addCommentRequestBodyDTO := dto.AddCommentRequestBodyDTO{}
	err := json.Unmarshal([]byte(request.Body), &addCommentRequestBodyDTO)
	if err != nil {
		cause := errutil.ErrJsonDecode.Errorf("AddCommentHandler - error decoding request body: %w", err)
		return errutil.ToAPIGatewayProxyResponse(context, cause)
	}

	// ToDo @ender - errors from api layer is ignored in all handlers...
	result, err := functions.ArticleApi.AddComment(context, userId, slug, addCommentRequestBodyDTO)

	if err != nil {
		return errutil.ToAPIGatewayProxyResponse(context, err)
	}

	jsonResult, err := json.Marshal(result)
	if err != nil {
		cause := errutil.ErrJsonEncode.Errorf("AddCommentHandler - error encoding response body: %w", err)
		return errutil.ToAPIGatewayProxyResponse(context, errutil.ErrJsonEncode.Errorf(
			"AddCommentHandler - error encoding response body: %w", cause))
	}
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(jsonResult),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}
}

func main() {
	lambda.Start(Handler)
}

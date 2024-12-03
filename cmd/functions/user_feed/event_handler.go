package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"realworld-aws-lambda-dynamodb-golang/cmd/functions"
	"realworld-aws-lambda-dynamodb-golang/internal/eventhandler"
)

func handleRequest(ctx context.Context, event events.DynamoDBEvent) (eventhandler.BatchResult, error) {
	return functions.ArticleUserFeedHandler.HandleEvent(ctx, event)
}

func main() {
	lambda.Start(handleRequest)
}

package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"realworld-aws-lambda-dynamodb-golang/cmd/functions"
	"realworld-aws-lambda-dynamodb-golang/internal/api"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
)

const HandlerName = "GetUserFeedHandler"
const DefaultLimit = 2

func Handler(ctx context.Context, request events.APIGatewayProxyRequest, userId uuid.UUID, _ domain.Token) events.APIGatewayProxyResponse {
	api.GetQueryParamOrDefault(ctx, request, "limit", HandlerName, DefaultLimit)

	result, err := functions.UserFeedApi.FetchUserFeed(ctx, userId, DefaultLimit)
	if err != nil {
		return api.ToInternalServerError(ctx, err)
	}

	//slog.DebugContext(ctx, "result", slog.Any("result", result))

	return api.ToSuccessAPIGatewayProxyResponse(ctx, result, HandlerName)
}

func main() {
	api.StartAuthenticatedHandler(Handler)
}

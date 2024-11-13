package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/caarlos0/env/v11"
	"github.com/google/uuid"
	"log"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/cmd/functions"
	"realworld-aws-lambda-dynamodb-golang/internal/api"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
)

const HandlerName = "GetUserFeedHandler"

func init() {
	http.Handle("GET /api/articles/feed", api.StartAuthenticatedHandlerHTTP(HandlerHTTP))
}

type UserFeedConfig struct {
	DefaultLimit int `env:"DEFAULT_LIMIT,notEmpty" envDefault:"10"`
	MinLimit     int `env:"MIN_LIMIT,notEmpty" envDefault:"1"`
	MaxLimit     int `env:"MAX_LIMIT,notEmpty" envDefault:"20"`
}

// one could define an empty variable declaration and later initialize it inside init() or main()
// but I don't like to create an empty variable declaration and then later assign a value to it.
var config = func() UserFeedConfig {
	var cfg UserFeedConfig
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}
	return cfg
}()

func HandlerHTTP(w http.ResponseWriter, r *http.Request, userId uuid.UUID, token domain.Token) {
	ctx := r.Context()

	limit, ok := api.GetIntQueryParamOrDefaultHTTP(ctx, w, r, "limit", config.DefaultLimit, &config.MinLimit, &config.MaxLimit)

	if !ok {
		return
	}

	nextPageToken, ok := api.GetOptionalStringQueryParamHTTP(w, r, "offset")

	if !ok {
		return
	}

	result, err := functions.UserFeedApi.FetchUserFeed(ctx, userId, limit, nextPageToken)
	if err != nil {
		api.ToInternalServerHTTPError(w, err)
		return
	}

	api.ToSuccessHTTPResponse(w, result)
}

func Handler(ctx context.Context, request events.APIGatewayProxyRequest, userId uuid.UUID, _ domain.Token) events.APIGatewayProxyResponse {
	limit, response := api.GetIntQueryParamOrDefault(ctx, request, "limit", config.DefaultLimit, &config.MinLimit, &config.MaxLimit)

	if response != nil {
		return *response
	}

	nextPageToken, response := api.GetOptionalStringQueryParam(ctx, request, "offset")

	if response != nil {
		return *response
	}

	result, err := functions.UserFeedApi.FetchUserFeed(ctx, userId, limit, nextPageToken)
	if err != nil {
		return api.ToInternalServerError(ctx, err)
	}

	return api.ToSuccessAPIGatewayProxyResponse(ctx, result, HandlerName)
}

func main() {
	lambda.Start(httpadapter.NewV2(http.DefaultServeMux).ProxyWithContext)
}

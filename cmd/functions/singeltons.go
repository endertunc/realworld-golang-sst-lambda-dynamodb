package functions

import (
	slogctx "github.com/veqryn/slog-context"
	veqrynslog "github.com/veqryn/slog-context/http"
	"log/slog"
	"os"
	"realworld-aws-lambda-dynamodb-golang/internal/api"
	"realworld-aws-lambda-dynamodb-golang/internal/database"
	"realworld-aws-lambda-dynamodb-golang/internal/repository"
	"realworld-aws-lambda-dynamodb-golang/internal/service"
)

var (
	//logger = slog.New(devslog.NewHandler(os.Stdout, &devslog.Options{
	//	HandlerOptions: &slog.HandlerOptions{
	//		AddSource: true,
	//		Level:     slog.LevelInfo,
	//	},
	//}))
	//logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	//	//AddSource: true,
	//	Level: slog.LevelDebug,
	//}))
	dynamodbStore      = database.NewDynamoDBStore()
	UserRepository     = repository.NewDynamodbUserRepository(dynamodbStore)
	UserService        = service.UserService{UserRepository: UserRepository}
	FollowerRepository = repository.NewDynamodbFollowerRepository(dynamodbStore)
	ProfileService     = service.ProfileService{FollowerRepository: FollowerRepository, UserRepository: UserRepository}
	ProfileApi         = api.ProfileApi{ProfileService: ProfileService}
	UserApi            = api.UserApi{UserService: UserService}
	ArticleRepository  = repository.NewDynamodbArticleRepository(dynamodbStore)
	ArticleService     = service.NewArticleService(UserService, ProfileService, ArticleRepository)
	ArticleApi         = api.NewArticleApi(ArticleService, UserService, ProfileService)
	UserFeedRepository = repository.NewUserFeedRepository(dynamodbStore)
	UserFeedService    = service.NewUserFeedService(UserFeedRepository, ArticleService, ProfileService, UserService)
	UserFeedApi        = api.NewFeedApi(UserFeedService)
)

func init() {
	h := slogctx.NewHandler(
		slog.NewJSONHandler(os.Stdout, nil),
		&slogctx.HandlerOptions{
			Prependers: []slogctx.AttrExtractor{
				veqrynslog.ExtractAttrCollection,
			},
		},
	)
	slog.SetDefault(slog.New(h))
}

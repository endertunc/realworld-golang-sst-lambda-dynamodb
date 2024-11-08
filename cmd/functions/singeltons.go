package functions

import (
	"log/slog"
	"os"
	"realworld-aws-lambda-dynamodb-golang/internal/database"
	"realworld-aws-lambda-dynamodb-golang/internal/user"
)

var (
	//logger = slog.New(devslog.NewHandler(os.Stdout, &devslog.Options{
	//	HandlerOptions: &slog.HandlerOptions{
	//		AddSource: true,
	//		Level:     slog.LevelInfo,
	//	},
	//}))
	logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		//AddSource: true,
		Level: slog.LevelDebug,
	}))
	dynamodbStore      = database.NewDynamoDBStore()
	UserRepository     = user.NewDynamodbUserRepository(dynamodbStore)
	UserService        = user.UserService{UserRepository: UserRepository}
	FollowerRepository = user.NewDynamodbFollowerRepository(dynamodbStore)
	ProfileService     = user.ProfileService{FollowerRepository: FollowerRepository, UserRepository: UserRepository}
	ProfileApi         = user.ProfileApi{ProfileService: ProfileService}
	UserApi            = user.UserApi{UserService: UserService}
	ArticleRepository  = user.NewDynamodbArticleRepository(dynamodbStore)
	ArticleService     = user.NewArticleService(UserService, ArticleRepository)
	ArticleApi         = user.NewArticleApi(ArticleService, UserService, ProfileService)
	UserFeedRepository = user.NewUserFeedRepository(dynamodbStore)
	UserFeedService    = user.NewUserFeedService(UserFeedRepository, ArticleService, ProfileService, UserService)
	UserFeedApi        = user.NewFeedApi(UserFeedService)
)

func init() {
	slog.SetDefault(logger)
}

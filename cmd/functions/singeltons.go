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
	dynamodbStore   = database.NewDynamoDBStore()
	opensearchStore = database.NewOpenSearchStore()

	followerRepository = repository.NewDynamodbFollowerRepository(dynamodbStore)

	userRepository = repository.NewDynamodbUserRepository(dynamodbStore)
	userService    = service.NewUserService(userRepository)
	UserApi        = api.NewUserApi(userService)

	articleRepository           = repository.NewDynamodbArticleRepository(dynamodbStore)
	articleOpenSearchRepository = repository.NewArticleOpensearchRepository(opensearchStore)
	articleService              = service.NewArticleService(articleRepository, articleOpenSearchRepository, userService, profileService)
	ArticleApi                  = api.NewArticleApi(articleService, userService, profileService)

	profileService = service.NewProfileService(followerRepository, userRepository)
	ProfileApi     = api.NewProfileApi(profileService)

	commentRepository = repository.NewDynamodbCommentRepository(dynamodbStore)
	commentService    = service.NewCommentService(commentRepository, articleService)
	CommentApi        = api.NewCommentApi(commentService, userService, profileService)

	userFeedRepository = repository.NewUserFeedRepository(dynamodbStore)
	UserFeedService    = service.NewUserFeedService(userFeedRepository, articleService, profileService, userService)
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

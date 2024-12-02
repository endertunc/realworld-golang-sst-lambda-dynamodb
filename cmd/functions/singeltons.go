package functions

import (
	"github.com/caarlos0/env/v11"
	"log"
	"log/slog"
	"os"
	"realworld-aws-lambda-dynamodb-golang/internal/api"
	"realworld-aws-lambda-dynamodb-golang/internal/database"
	"realworld-aws-lambda-dynamodb-golang/internal/repository"
	"realworld-aws-lambda-dynamodb-golang/internal/security"
	"realworld-aws-lambda-dynamodb-golang/internal/service"

	slogctx "github.com/veqryn/slog-context"
	veqrynslog "github.com/veqryn/slog-context/http"
)

var (
	dynamodbStore   = database.NewDynamoDBStore()
	opensearchStore = database.NewOpensearchStore()

	paginationConfig = api.GetPaginationConfig()

	followerRepository = repository.NewDynamodbFollowerRepository(dynamodbStore)

	userRepository = repository.NewDynamodbUserRepository(dynamodbStore)
	userService    = service.NewUserService(userRepository)
	UserApi        = api.NewUserApi(userService)

	articleRepository           = repository.NewDynamodbArticleRepository(dynamodbStore)
	articleOpenSearchRepository = repository.NewArticleOpensearchRepository(opensearchStore)
	articleService              = service.NewArticleService(articleRepository, articleOpenSearchRepository, userService, profileService)
	articleListService          = service.NewArticleListService(articleRepository, articleOpenSearchRepository, userService, profileService)
	ArticleApi                  = api.NewArticleApi(articleService, articleListService, userService, profileService, paginationConfig)

	profileService = service.NewProfileService(followerRepository, userRepository)
	ProfileApi     = api.NewProfileApi(profileService)

	commentRepository = repository.NewDynamodbCommentRepository(dynamodbStore)
	commentService    = service.NewCommentService(commentRepository, articleService)
	CommentApi        = api.NewCommentApi(commentService, userService, profileService)

	userFeedRepository = repository.NewUserFeedRepository(dynamodbStore)
	UserFeedService    = service.NewUserFeedService(userFeedRepository, articleService, profileService, userService)
	UserFeedApi        = api.NewUserFeedApi(UserFeedService, paginationConfig)
)

type AppConfig struct {
	JWTKeyPairSecretName string `env:"JWT_KEY_PAIR_SECRET_NAME"`
}

func init() {
	var appConfig AppConfig
	err := env.Parse(&appConfig)

	if err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}

	h := slogctx.NewHandler(
		slog.NewJSONHandler(os.Stdout, nil),
		&slogctx.HandlerOptions{
			Prependers: []slogctx.AttrExtractor{
				veqrynslog.ExtractAttrCollection,
			},
		},
	)
	slog.SetDefault(slog.New(h))

	security.SetKeyProvider(security.NewAwsKeyProvider(appConfig.JWTKeyPairSecretName))
}

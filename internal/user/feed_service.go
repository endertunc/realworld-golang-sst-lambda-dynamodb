package user

import (
	"context"
	"github.com/google/uuid"
	"log/slog"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/utils"
	"time"
)

type UserFeedService struct {
	UserFeedRepository UserFeedRepositoryInterface
	ArticleService     ArticleServiceInterface
	ProfileService     ProfileServiceInterface
	UserService        UserServiceInterface
}

var _ FeedServiceInterface = UserFeedService{}

func NewUserFeedService(
	userFeedRepository UserFeedRepositoryInterface,
	articleService ArticleServiceInterface,
	profileService ProfileServiceInterface,
	userService UserServiceInterface) UserFeedService {
	return UserFeedService{
		UserFeedRepository: userFeedRepository,
		ArticleService:     articleService,
		ProfileService:     profileService,
		UserService:        userService}
}

func (uf UserFeedService) FanoutArticle(ctx context.Context, articleId, authorId uuid.UUID, createdAt time.Time) error {
	return uf.UserFeedRepository.FanoutArticle(ctx, articleId, authorId, createdAt)
}

/*
 * FetchUserFeed fetches the user feed for a given user
 * this function could be optimized by combining the following steps into a single BatchGetItems query
 * - FindArticlesByIds and IsFavoritedBulk: both needs articleIds can be fetched in a single query
 * - GetUserListByUserIDs and IsFollowingBulk: both needs uniqueAuthorIdsList and can be fetched in a single query
 */
func (uf UserFeedService) FetchArticlesFromFeed(ctx context.Context, userId uuid.UUID, limit int) ([]domain.FeedItem, error) {
	articleIds, err := uf.UserFeedRepository.FindArticleIdsInUserFeed(ctx, userId, limit)
	if err != nil {
		return nil, err
	}

	articles, err := uf.ArticleService.FindArticlesByIds(ctx, articleIds)
	if err != nil {
		return nil, err
	}
	authorIdsList := make([]uuid.UUID, 0)
	for _, article := range articles {
		authorIdsList = append(authorIdsList, article.AuthorId)
	}

	// @ToDo @ender - this code is duplicate
	uniqueAuthorIdsList := utils.RemoveDuplicatesFromList(authorIdsList)
	slog.DebugContext(ctx, "uniqueAuthorIdsList", slog.Any("uniqueAuthorIdsList", uniqueAuthorIdsList))
	// fetch authors (users) in bulk and create a map for lookup by authorId
	authors, err := uf.UserService.GetUserListByUserIDs(ctx, uniqueAuthorIdsList)
	if err != nil {
		return nil, err
	}

	// fetch isFollowing in bulk
	isFollowingMap, err := uf.ProfileService.IsFollowingBulk(ctx, userId, uniqueAuthorIdsList)
	if err != nil {
		return nil, err
	}

	// fetch isFollowing in bulk
	isFavoritedMap, err := uf.ArticleService.IsFavoritedBulk(ctx, userId, articleIds)
	if err != nil {
		return nil, err
	}

	authorsMap := make(map[uuid.UUID]domain.User)
	for _, author := range authors {
		authorsMap[author.Id] = author
	}

	feedItems := make([]domain.FeedItem, 0)
	for _, article := range articles {
		_, isFollowing := isFollowingMap[article.AuthorId]
		author, authorFound := authorsMap[article.AuthorId]
		// 1- we don't show articles from users that the current user is not following
		// 2- we should have the author in the map, otherwise let it skip
		if isFollowing && authorFound {
			_, isFavorited := isFavoritedMap[article.Id]
			feedItem := domain.FeedItem{
				Article:     article,
				Author:      author,
				IsFavorited: isFavorited,
				IsFollowing: isFollowing,
			}
			feedItems = append(feedItems, feedItem)
		}
	}
	return feedItems, err
}

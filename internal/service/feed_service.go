package service

import (
	"context"
	"github.com/google/uuid"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/repository"
	"realworld-aws-lambda-dynamodb-golang/internal/utils"
	"time"
)

type UserFeedService struct {
	UserFeedRepository repository.UserFeedRepositoryInterface
	ArticleService     ArticleServiceInterface
	ProfileService     ProfileServiceInterface
	UserService        UserServiceInterface
}

type FeedServiceInterface interface {
	FanoutArticle(ctx context.Context, articleId, authorId uuid.UUID, createdAt time.Time) error
	FetchArticlesFromFeed(ctx context.Context, userId uuid.UUID, limit int, nextPageToken *string) ([]domain.FeedItem, *string, error)
}

var _ FeedServiceInterface = UserFeedService{}

func NewUserFeedService(
	userFeedRepository repository.UserFeedRepositoryInterface,
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
func (uf UserFeedService) FetchArticlesFromFeed(ctx context.Context, userId uuid.UUID, limit int, nextPageToken *string) ([]domain.FeedItem, *string, error) {
	articleIds, nextToken, err := uf.UserFeedRepository.FindArticleIdsInUserFeed(ctx, userId, limit, nextPageToken)
	if err != nil {
		return nil, nil, err
	}

	articles, err := uf.ArticleService.FindArticlesByIds(ctx, articleIds)
	if err != nil {
		return nil, nil, err
	}

	authorIdToArticleMap := make(map[uuid.UUID]domain.Article)
	for _, article := range articles {
		authorIdToArticleMap[article.Id] = article
	}

	authorIdsList := make([]uuid.UUID, 0)
	for _, article := range articles {
		authorIdsList = append(authorIdsList, article.AuthorId)
	}

	// @ToDo @ender - this code is duplicate
	uniqueAuthorIdsList := utils.RemoveDuplicatesFromList(authorIdsList)

	// fetch authors (users) in bulk and create a map for lookup by authorId
	authors, err := uf.UserService.GetUserListByUserIDs(ctx, uniqueAuthorIdsList)
	if err != nil {
		return nil, nil, err
	}

	authorsMap := make(map[uuid.UUID]domain.User)
	for _, author := range authors {
		authorsMap[author.Id] = author
	}

	// fetch isFollowing in bulk
	followedAuthorsSet, err := uf.ProfileService.IsFollowingBulk(ctx, userId, uniqueAuthorIdsList)
	if err != nil {
		return nil, nil, err
	}

	// fetch isFollowing in bulk
	favoritedArticlesSet, err := uf.ArticleService.IsFavoritedBulk(ctx, userId, articleIds)
	if err != nil {
		return nil, nil, err
	}

	feedItems := make([]domain.FeedItem, 0)
	// we need to return article in the order of articleIds
	for _, articleId := range articleIds {
		article, articleFound := authorIdToArticleMap[articleId]
		isFollowing := followedAuthorsSet.ContainsOne(article.AuthorId)
		author, authorFound := authorsMap[article.AuthorId]

		// 1- we should have the article in the authorIdToArticleMap, otherwise let it skip
		// 2- we don't show articles from users that the current user is not following
		// 3- we should have the author in the authorsMap, otherwise let it skip
		if articleFound && isFollowing && authorFound {
			isFavorited := favoritedArticlesSet.ContainsOne(article.Id)
			feedItem := domain.FeedItem{
				Article:     article,
				Author:      author,
				IsFavorited: isFavorited,
				IsFollowing: isFollowing,
			}
			feedItems = append(feedItems, feedItem)
		}
	}

	return feedItems, nextToken, nil
}

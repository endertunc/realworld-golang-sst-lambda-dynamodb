package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/repository"
	"time"
)

type userFeedService struct {
	userFeedRepository repository.UserFeedRepositoryInterface
	articleService     ArticleServiceInterface
	profileService     ProfileServiceInterface
	userService        UserServiceInterface
}

type FeedServiceInterface interface {
	FanoutArticle(ctx context.Context, articleId, authorId uuid.UUID, createdAt time.Time) error
	FetchArticlesFromFeed(ctx context.Context, userId uuid.UUID, limit int, nextPageToken *string) ([]domain.ArticleAggregateView, *string, error)
}

var _ FeedServiceInterface = userFeedService{} //nolint:golint,exhaustruct

func NewUserFeedService(
	userFeedRepository repository.UserFeedRepositoryInterface,
	articleService ArticleServiceInterface,
	profileService ProfileServiceInterface,
	userService UserServiceInterface) FeedServiceInterface {
	return userFeedService{
		userFeedRepository: userFeedRepository,
		articleService:     articleService,
		profileService:     profileService,
		userService:        userService}
}

func (uf userFeedService) FanoutArticle(ctx context.Context, articleId, authorId uuid.UUID, createdAt time.Time) error {
	return uf.userFeedRepository.FanoutArticle(ctx, articleId, authorId, createdAt)
}

/*
 * FetchUserFeed fetches the user feed for a given user
 * this function could be optimized by combining the following steps into a single BatchGetItems query
 * - GetArticlesByIds and IsFavoritedBulk: both needs articleIds can be fetched in a single query
 * - GetUserListByUserIDs and IsFollowingBulk: both needs uniqueAuthorIdsList and can be fetched in a single query
 */
func (uf userFeedService) FetchArticlesFromFeed(ctx context.Context, userId uuid.UUID, limit int, nextPageToken *string) ([]domain.ArticleAggregateView, *string, error) {
	articleIds, nextToken, err := uf.userFeedRepository.FindArticleIdsInUserFeed(ctx, userId, limit, nextPageToken)
	if err != nil {
		return nil, nil, err
	}

	articles, err := uf.articleService.GetArticlesByIds(ctx, articleIds)
	if err != nil {
		return nil, nil, err
	}

	authorIdToArticleMap := make(map[uuid.UUID]domain.Article)
	for _, article := range articles {
		authorIdToArticleMap[article.Id] = article
	}

	authorIdsList := lo.Map(articles, func(article domain.Article, _ int) uuid.UUID {
		return article.AuthorId
	})
	uniqueAuthorIdsList := lo.Uniq(authorIdsList)

	// fetch authors (users) in bulk and create a map for lookup by authorId
	authors, err := uf.userService.GetUserListByUserIDs(ctx, uniqueAuthorIdsList)
	if err != nil {
		return nil, nil, err
	}

	authorsMap := make(map[uuid.UUID]domain.User)
	for _, author := range authors {
		authorsMap[author.Id] = author
	}

	// fetch isFollowing in bulk
	followedAuthorsSet, err := uf.profileService.IsFollowingBulk(ctx, userId, uniqueAuthorIdsList)
	if err != nil {
		return nil, nil, err
	}

	// fetch isFollowing in bulk
	favoritedArticlesSet, err := uf.articleService.IsFavoritedBulk(ctx, userId, articleIds)
	if err != nil {
		return nil, nil, err
	}

	feedItems := make([]domain.ArticleAggregateView, 0)
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
			feedItem := domain.ArticleAggregateView{
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

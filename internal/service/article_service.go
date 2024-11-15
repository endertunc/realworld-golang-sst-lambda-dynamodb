package service

import (
	"context"
	"errors"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/repository"
)

type articleService struct {
	articleRepository repository.ArticleRepositoryInterface
	userService       UserServiceInterface
	profileService    ProfileServiceInterface
	commentRepository repository.CommentRepositoryInterface
}

type ArticleServiceInterface interface {
	GetArticle(ctx context.Context, slug string) (domain.Article, error)
	GetArticlesByIds(ctx context.Context, articleIds []uuid.UUID) ([]domain.Article, error)
	GetArticleBySlug(ctx context.Context, slug string) (domain.Article, error)

	CreateArticle(ctx context.Context, author uuid.UUID, title, description, body string, tagList []string) (domain.Article, error)
	DeleteArticle(ctx context.Context, author uuid.UUID, slug string) error

	FavoriteArticle(ctx context.Context, userId uuid.UUID, slug string) (domain.Article, error)
	UnfavoriteArticle(ctx context.Context, userId uuid.UUID, slug string) (domain.Article, error)
	IsFavorited(ctx context.Context, articleId, userId uuid.UUID) (bool, error)
	IsFavoritedBulk(ctx context.Context, userId uuid.UUID, articleIds []uuid.UUID) (mapset.Set[uuid.UUID], error)

	GetMostRecentArticlesByAuthor(ctx context.Context, userId *uuid.UUID, author string, limit int, nextPageToken *string) ([]domain.FeedItem, *string, error)
	GetMostRecentArticlesFavoritedByUser(ctx context.Context, loggedInUser *uuid.UUID, favoritedByUsername string, limit int, nextPageToken *string) ([]domain.FeedItem, *string, error)
	GetMostRecentArticlesFavoritedByTag(ctx context.Context, loggedInUser *uuid.UUID, tag string, limit int, nextPageToken *string) ([]domain.FeedItem, *string, error)
	GetMostRecentArticlesGlobally(ctx context.Context, loggedInUser *uuid.UUID, limit int, nextPageToken *string) ([]domain.FeedItem, *string, error)
}

var _ ArticleServiceInterface = articleService{}

func NewArticleService(articleRepository repository.ArticleRepositoryInterface, userService UserServiceInterface, profileService ProfileServiceInterface) ArticleServiceInterface {
	return articleService{
		userService:       userService,
		profileService:    profileService,
		articleRepository: articleRepository,
	}
}

func (as articleService) GetArticle(ctx context.Context, slug string) (domain.Article, error) {
	article, err := as.articleRepository.FindArticleBySlug(ctx, slug)
	if err != nil {
		return domain.Article{}, err
	}
	return article, nil
}

func (as articleService) CreateArticle(ctx context.Context, author uuid.UUID, title, description, body string, tagList []string) (domain.Article, error) {
	article := domain.NewArticle(title, description, body, tagList, author)
	// ToDo @ender do we have any business validation we should apply in service level for an article?
	article, err := as.articleRepository.CreateArticle(ctx, article)
	if err != nil {
		return domain.Article{}, err
	}
	return article, nil
}

func (as articleService) UnfavoriteArticle(ctx context.Context, loggedInUserId uuid.UUID, slug string) (domain.Article, error) {
	article, err := as.articleRepository.FindArticleBySlug(ctx, slug)
	if err != nil {
		return domain.Article{}, err
	}
	// ToDo @ender check if the user already unfavorited the article
	err = as.articleRepository.UnfavoriteArticle(ctx, loggedInUserId, article.Id)
	if err != nil {
		return domain.Article{}, err
	}
	article.FavoritesCount--
	return article, nil
}

func (as articleService) FavoriteArticle(ctx context.Context, loggedInUserId uuid.UUID, slug string) (domain.Article, error) {
	article, err := as.articleRepository.FindArticleBySlug(ctx, slug)
	if err != nil {
		return domain.Article{}, err
	}
	err = as.articleRepository.FavoriteArticle(ctx, loggedInUserId, article.Id)
	if err != nil {
		return domain.Article{}, err
	}
	// we increment the count here to avoid another query
	// given the fact that favoritesCount is not critical data,
	article.FavoritesCount++
	return article, nil
}

func (as articleService) DeleteArticle(ctx context.Context, authorId uuid.UUID, slug string) error {
	article, err := as.articleRepository.FindArticleBySlug(ctx, slug)
	if err != nil {
		return err
	}

	if article.AuthorId != authorId {
		// ToDo @ender return a proper error
		return errors.New("you can't touch this")
	}

	err = as.articleRepository.DeleteArticleById(ctx, article.Id)
	if err != nil {
		return err
	}
	return nil
}

func (as articleService) IsFavorited(ctx context.Context, articleId, userId uuid.UUID) (bool, error) {
	return as.articleRepository.IsFavorited(ctx, articleId, userId)
}

func (as articleService) GetArticlesByIds(ctx context.Context, articleIds []uuid.UUID) ([]domain.Article, error) {
	return as.articleRepository.FindArticlesByIds(ctx, articleIds)
}

func (as articleService) GetArticleBySlug(ctx context.Context, slug string) (domain.Article, error) {
	return as.articleRepository.FindArticleBySlug(ctx, slug)
}

func (as articleService) IsFavoritedBulk(ctx context.Context, userId uuid.UUID, articleIds []uuid.UUID) (mapset.Set[uuid.UUID], error) {
	return as.articleRepository.IsFavoritedBulk(ctx, userId, articleIds)
}

// ToDo @ender to keep things short - we better pass limit and nextPageToken struct Pagination or something like that.
// ToDo @ender this is first iteration, GetMostRecentArticles* share a lot of common code
func (as articleService) GetMostRecentArticlesByAuthor(ctx context.Context, loggedInUser *uuid.UUID, author string, limit int, nextPageToken *string) ([]domain.FeedItem, *string, error) {
	// find the author user by username
	authorUser, err := as.userService.GetUserByUsername(ctx, author)
	if err != nil {
		// ToDo @ender if author is not found, should we return an error or an empty list?
		//  at the moment ErrUserNotFound is mapped to StatusNotFound
		return nil, nil, err
	}

	articles, nextToken, err := as.articleRepository.FindArticlesByAuthor(ctx, authorUser.Id, limit, nextPageToken)
	if err != nil {
		return nil, nil, err
	}

	isFollowing := false
	favoritedArticlesSet := mapset.NewSet[uuid.UUID]()
	if loggedInUser != nil {
		// check if the user is following the author
		isFollowing, err = as.profileService.IsFollowing(ctx, *loggedInUser, authorUser.Id)
		if err != nil {
			return nil, nil, err
		}

		// fetch isFollowing in bulk
		articleIds := lo.Map(articles, func(article domain.Article, _ int) uuid.UUID {
			return article.Id
		})
		favoritedArticlesSet, err = as.IsFavoritedBulk(ctx, *loggedInUser, articleIds)
		if err != nil {
			return nil, nil, err
		}
	}

	feedItems := make([]domain.FeedItem, 0, len(articles))

	// we need to return articles in the order they are returned from the database
	for _, article := range articles {
		isFavorited := favoritedArticlesSet.ContainsOne(article.Id)
		feedItem := domain.FeedItem{
			Article:     article,
			Author:      authorUser,
			IsFavorited: isFavorited,
			IsFollowing: isFollowing,
		}
		feedItems = append(feedItems, feedItem)
	}

	return feedItems, nextToken, nil

}

// ToDo @ender this is first iteration, GetMostRecentArticles* share a lot of common code
func (as articleService) GetMostRecentArticlesFavoritedByUser(ctx context.Context, loggedInUser *uuid.UUID, favoritedByUsername string, limit int, nextPageToken *string) ([]domain.FeedItem, *string, error) {
	// find the author user by username
	favoritedByUser, err := as.userService.GetUserByUsername(ctx, favoritedByUsername)
	if err != nil {
		// ToDo @ender if author is not found, should we return an error or an empty list?
		//  at the moment ErrUserNotFound is mapped to StatusNotFound
		return nil, nil, err
	}

	articleIds, nextToken, err := as.articleRepository.FindArticlesFavoritedByUser(ctx, favoritedByUser.Id, limit, nextPageToken)
	if err != nil {
		return nil, nil, err
	}

	articles, err := as.articleRepository.FindArticlesByIds(ctx, articleIds)
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
	authors, err := as.userService.GetUserListByUserIDs(ctx, uniqueAuthorIdsList)
	if err != nil {
		return nil, nil, err
	}

	authorsMap := make(map[uuid.UUID]domain.User)
	for _, author := range authors {
		authorsMap[author.Id] = author
	}

	followedAuthorsSet := mapset.NewThreadUnsafeSet[uuid.UUID]()
	favoritedArticlesSet := mapset.NewThreadUnsafeSet[uuid.UUID]()
	if loggedInUser != nil {
		// fetch isFollowing in bulk
		followedAuthorsSet, err = as.profileService.IsFollowingBulk(ctx, *loggedInUser, uniqueAuthorIdsList)
		if err != nil {
			return nil, nil, err
		}
		// fetch isFollowing in bulk
		favoritedArticlesSet, err = as.articleRepository.IsFavoritedBulk(ctx, *loggedInUser, articleIds)
		if err != nil {
			return nil, nil, err
		}
	}

	feedItems := make([]domain.FeedItem, 0)
	// we need to return article in the order of articleIds
	for _, articleId := range articleIds {
		article, articleFound := authorIdToArticleMap[articleId]
		author, authorFound := authorsMap[article.AuthorId]

		// 1- we should have the article in the authorIdToArticleMap, otherwise let it skip
		// 2- we should have the author in the authorsMap, otherwise let it skip
		if articleFound && authorFound {
			isFavorited := favoritedArticlesSet.ContainsOne(article.Id)
			isFollowing := followedAuthorsSet.ContainsOne(article.AuthorId)
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

// ToDo @ender this is first iteration, GetMostRecentArticles* share a lot of common code
func (as articleService) GetMostRecentArticlesFavoritedByTag(ctx context.Context, loggedInUser *uuid.UUID, tag string, limit int, nextPageToken *string) ([]domain.FeedItem, *string, error) {
	// this time we should fetch this information from elasticsearch
	var (
		articles  []domain.Article = nil
		nextToken *string          = nil
		err       error            = nil
	)

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
	authors, err := as.userService.GetUserListByUserIDs(ctx, uniqueAuthorIdsList)
	if err != nil {
		return nil, nil, err
	}

	authorsMap := make(map[uuid.UUID]domain.User)
	for _, author := range authors {
		authorsMap[author.Id] = author
	}

	followedAuthorsSet := mapset.NewThreadUnsafeSet[uuid.UUID]()
	favoritedArticlesSet := mapset.NewThreadUnsafeSet[uuid.UUID]()
	if loggedInUser != nil {
		// fetch isFollowing in bulk
		followedAuthorsSet, err = as.profileService.IsFollowingBulk(ctx, *loggedInUser, uniqueAuthorIdsList)
		if err != nil {
			return nil, nil, err
		}
		// fetch isFollowing in bulk
		articleIds := lo.Map(articles, func(article domain.Article, _ int) uuid.UUID { return article.Id })
		favoritedArticlesSet, err = as.articleRepository.IsFavoritedBulk(ctx, *loggedInUser, articleIds)
		if err != nil {
			return nil, nil, err
		}
	}

	feedItems := make([]domain.FeedItem, 0)
	// we need to return article in the order of articleIds
	for _, article := range articles {
		author, authorFound := authorsMap[article.AuthorId]
		// 1- we should have the author in the authorsMap, otherwise let it skip
		if authorFound {
			isFavorited := favoritedArticlesSet.ContainsOne(article.Id)
			isFollowing := followedAuthorsSet.ContainsOne(article.AuthorId)
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

// ToDo @ender this is first iteration, GetMostRecentArticles* share a lot of common code
func (as articleService) GetMostRecentArticlesGlobally(ctx context.Context, loggedInUser *uuid.UUID, limit int, nextPageToken *string) ([]domain.FeedItem, *string, error) {
	// this time we should fetch this information from elasticsearch
	var (
		articles  []domain.Article = nil
		nextToken *string          = nil
		err       error            = nil
	)
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
	authors, err := as.userService.GetUserListByUserIDs(ctx, uniqueAuthorIdsList)
	if err != nil {
		return nil, nil, err
	}

	authorsMap := make(map[uuid.UUID]domain.User)
	for _, author := range authors {
		authorsMap[author.Id] = author
	}

	followedAuthorsSet := mapset.NewThreadUnsafeSet[uuid.UUID]()
	favoritedArticlesSet := mapset.NewThreadUnsafeSet[uuid.UUID]()
	if loggedInUser != nil {
		// fetch isFollowing in bulk
		followedAuthorsSet, err = as.profileService.IsFollowingBulk(ctx, *loggedInUser, uniqueAuthorIdsList)
		if err != nil {
			return nil, nil, err
		}

		// fetch isFollowing in bulk
		articleIds := lo.Map(articles, func(article domain.Article, _ int) uuid.UUID { return article.Id })
		favoritedArticlesSet, err = as.articleRepository.IsFavoritedBulk(ctx, *loggedInUser, articleIds)
		if err != nil {
			return nil, nil, err
		}
	}

	feedItems := make([]domain.FeedItem, 0)
	// we need to return articles in the order they are returned from the database
	for _, article := range articles {
		author, authorFound := authorsMap[article.AuthorId]
		// 1- we should have the author in the authorsMap, otherwise let it skip
		if authorFound {
			isFavorited := favoritedArticlesSet.ContainsOne(article.Id)
			isFollowing := followedAuthorsSet.ContainsOne(article.AuthorId)
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

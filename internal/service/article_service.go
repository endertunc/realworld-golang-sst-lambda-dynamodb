package service

import (
	"context"
	"errors"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/repository"
)

type ArticleService struct {
	UserService       UserServiceInterface
	ProfileService    ProfileServiceInterface
	ArticleRepository repository.ArticleRepositoryInterface
}

type ArticleServiceInterface interface {
	GetArticle(ctx context.Context, slug string) (domain.Article, error)
	CreateArticle(ctx context.Context, author uuid.UUID, title, description, body string, tagList []string) (domain.Article, error)
	AddComment(ctx context.Context, loggedInUserId uuid.UUID, articleSlug string, body string) (domain.Comment, error)
	GetArticleComments(ctx context.Context, slug string) ([]domain.Comment, error)
	DeleteComment(ctx context.Context, author uuid.UUID, slug string, commentId uuid.UUID) error
	DeleteArticle(ctx context.Context, author uuid.UUID, slug string) error
	FavoriteArticle(ctx context.Context, userId uuid.UUID, slug string) (domain.Article, error)
	UnfavoriteArticle(ctx context.Context, userId uuid.UUID, slug string) (domain.Article, error)
	IsFavorited(ctx context.Context, articleId, userId uuid.UUID) (bool, error)
	FindArticlesByIds(ctx context.Context, articleIds []uuid.UUID) ([]domain.Article, error)
	IsFavoritedBulk(ctx context.Context, userId uuid.UUID, articleIds []uuid.UUID) (mapset.Set[uuid.UUID], error)
	GetMostRecentArticlesByAuthor(ctx context.Context, userId *uuid.UUID, author string, limit int, nextPageToken *string) ([]domain.FeedItem, *string, error)
	GetMostRecentArticlesFavoritedByUser(ctx context.Context, loggedInUser *uuid.UUID, favoritedByUsername string, limit int, nextPageToken *string) ([]domain.FeedItem, *string, error)
	GetMostRecentArticlesFavoritedByTag(ctx context.Context, loggedInUser *uuid.UUID, tag string, limit int, nextPageToken *string) ([]domain.FeedItem, *string, error)
	GetMostRecentArticlesGlobally(ctx context.Context, loggedInUser *uuid.UUID, limit int, nextPageToken *string) ([]domain.FeedItem, *string, error)
}

var _ ArticleServiceInterface = ArticleService{}

func NewArticleService(userService UserServiceInterface, profileService ProfileServiceInterface, articleRepository repository.ArticleRepositoryInterface) ArticleService {
	return ArticleService{
		UserService:       userService,
		ProfileService:    profileService,
		ArticleRepository: articleRepository,
	}
}

func (as ArticleService) GetArticle(ctx context.Context, slug string) (domain.Article, error) {
	article, err := as.ArticleRepository.FindArticleBySlug(ctx, slug)
	if err != nil {
		return domain.Article{}, err
	}
	return article, nil
}

func (as ArticleService) CreateArticle(ctx context.Context, author uuid.UUID, title, description, body string, tagList []string) (domain.Article, error) {

	article := domain.NewArticle(title, description, body, tagList, author)
	// ToDo @ender do we have any business validation we should apply in service level for an article?
	article, err := as.ArticleRepository.CreateArticle(ctx, article)
	if err != nil {
		return domain.Article{}, err
	}
	return article, nil
}

func (as ArticleService) AddComment(ctx context.Context, author uuid.UUID, articleSlug string, body string) (domain.Comment, error) {
	article, err := as.ArticleRepository.FindArticleBySlug(ctx, articleSlug)
	if err != nil {
		return domain.Comment{}, err
	}
	comment := domain.NewComment(article.Id, author, body)

	err = as.ArticleRepository.CreateComment(ctx, comment)
	if err != nil {
		return domain.Comment{}, err
	}
	return comment, nil
}

func (as ArticleService) UnfavoriteArticle(ctx context.Context, loggedInUserId uuid.UUID, slug string) (domain.Article, error) {
	article, err := as.ArticleRepository.FindArticleBySlug(ctx, slug)
	if err != nil {
		return domain.Article{}, err
	}
	// ToDo @ender check if the user already unfavorited the article
	err = as.ArticleRepository.UnfavoriteArticle(ctx, loggedInUserId, article.Id)
	if err != nil {
		return domain.Article{}, err
	}
	article.FavoritesCount--
	return article, nil
}

func (as ArticleService) FavoriteArticle(ctx context.Context, loggedInUserId uuid.UUID, slug string) (domain.Article, error) {
	article, err := as.ArticleRepository.FindArticleBySlug(ctx, slug)
	if err != nil {
		return domain.Article{}, err
	}
	err = as.ArticleRepository.FavoriteArticle(ctx, loggedInUserId, article.Id)
	// ToDo @ender check if the user already favorited the article
	if err != nil {
		return domain.Article{}, err
	}
	// we increment the count here to avoid another query
	// given the fact that favoritesCount is not a critical date
	article.FavoritesCount++
	return article, nil
}

// DeleteComment
/**
 * ToDo @ender We can actually delete a comment only if the comment belongs to the user with a single query
 */
func (as ArticleService) DeleteComment(ctx context.Context, loggedInUserId uuid.UUID, slug string, commentId uuid.UUID) error {
	article, err := as.ArticleRepository.FindArticleBySlug(ctx, slug)
	if err != nil {
		return err
	}
	// check if the comment belongs to the article / comment exists
	comment, err := as.ArticleRepository.FindCommentByCommentIdAndArticleId(ctx, commentId, article.Id)
	if err != nil {
		return err
	}

	// check if the comment belongs to the user
	if comment.AuthorId != loggedInUserId {
		return errutil.ErrCantDeleteOthersComment
	}

	err = as.ArticleRepository.DeleteCommentByArticleIdAndCommentId(ctx, loggedInUserId, article.Id, commentId)
	if err != nil {
		return err
	}
	return nil
}

func (as ArticleService) GetArticleComments(ctx context.Context, slug string) ([]domain.Comment, error) {
	article, err := as.ArticleRepository.FindArticleBySlug(ctx, slug)
	if err != nil {
		return []domain.Comment{}, err
	}
	comments, err := as.ArticleRepository.FindCommentsByArticleId(ctx, article.Id)
	if err != nil {
		return []domain.Comment{}, err
	}
	return comments, nil
}

func (as ArticleService) DeleteArticle(ctx context.Context, authorId uuid.UUID, slug string) error {
	article, err := as.ArticleRepository.FindArticleBySlug(ctx, slug)
	if err != nil {
		return err
	}

	if article.AuthorId != authorId {
		// ToDo @ender return a proper error
		return errors.New("you can't touch this")
	}

	err = as.ArticleRepository.DeleteArticleById(ctx, article.Id)
	if err != nil {
		return err
	}
	return nil
}

func (as ArticleService) IsFavorited(ctx context.Context, articleId, userId uuid.UUID) (bool, error) {
	return as.ArticleRepository.IsFavorited(ctx, articleId, userId)
}

func (as ArticleService) FindArticlesByIds(ctx context.Context, articleIds []uuid.UUID) ([]domain.Article, error) {
	return as.ArticleRepository.FindArticlesByIds(ctx, articleIds)
}

func (as ArticleService) IsFavoritedBulk(ctx context.Context, userId uuid.UUID, articleIds []uuid.UUID) (mapset.Set[uuid.UUID], error) {
	return as.ArticleRepository.IsFavoritedBulk(ctx, userId, articleIds)
}

// ToDo @ender to keep things short - we better pass limit and nextPageToken struct Pagination or something like that.
// ToDo @ender this is first iteration, GetMostRecentArticles* share a lot of common code
func (as ArticleService) GetMostRecentArticlesByAuthor(ctx context.Context, loggedInUser *uuid.UUID, author string, limit int, nextPageToken *string) ([]domain.FeedItem, *string, error) {
	// find the author user by username
	authorUser, err := as.UserService.GetUserByUsername(ctx, author)
	if err != nil {
		// ToDo @ender if author is not found, should we return an error or an empty list?
		//  at the moment ErrUserNotFound is mapped to StatusNotFound
		return nil, nil, err
	}

	articles, nextToken, err := as.ArticleRepository.FindArticlesByAuthor(ctx, authorUser.Id, limit, nextPageToken)
	if err != nil {
		return nil, nil, err
	}

	isFollowing := false
	favoritedArticlesSet := mapset.NewSet[uuid.UUID]()
	if loggedInUser != nil {
		// check if the user is following the author
		isFollowing, err = as.ProfileService.IsFollowing(ctx, *loggedInUser, authorUser.Id)
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
func (as ArticleService) GetMostRecentArticlesFavoritedByUser(ctx context.Context, loggedInUser *uuid.UUID, favoritedByUsername string, limit int, nextPageToken *string) ([]domain.FeedItem, *string, error) {
	// find the author user by username
	favoritedByUser, err := as.UserService.GetUserByUsername(ctx, favoritedByUsername)
	if err != nil {
		// ToDo @ender if author is not found, should we return an error or an empty list?
		//  at the moment ErrUserNotFound is mapped to StatusNotFound
		return nil, nil, err
	}

	articleIds, nextToken, err := as.ArticleRepository.FindArticlesFavoritedByUser(ctx, favoritedByUser.Id, limit, nextPageToken)
	if err != nil {
		return nil, nil, err
	}

	articles, err := as.ArticleRepository.FindArticlesByIds(ctx, articleIds)
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
	authors, err := as.UserService.GetUserListByUserIDs(ctx, uniqueAuthorIdsList)
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
		followedAuthorsSet, err = as.ProfileService.IsFollowingBulk(ctx, *loggedInUser, uniqueAuthorIdsList)
		if err != nil {
			return nil, nil, err
		}
		// fetch isFollowing in bulk
		favoritedArticlesSet, err = as.ArticleRepository.IsFavoritedBulk(ctx, *loggedInUser, articleIds)
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
func (as ArticleService) GetMostRecentArticlesFavoritedByTag(ctx context.Context, loggedInUser *uuid.UUID, tag string, limit int, nextPageToken *string) ([]domain.FeedItem, *string, error) {
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
	authors, err := as.UserService.GetUserListByUserIDs(ctx, uniqueAuthorIdsList)
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
		followedAuthorsSet, err = as.ProfileService.IsFollowingBulk(ctx, *loggedInUser, uniqueAuthorIdsList)
		if err != nil {
			return nil, nil, err
		}
		// fetch isFollowing in bulk
		articleIds := lo.Map(articles, func(article domain.Article, _ int) uuid.UUID { return article.Id })
		favoritedArticlesSet, err = as.ArticleRepository.IsFavoritedBulk(ctx, *loggedInUser, articleIds)
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
func (as ArticleService) GetMostRecentArticlesGlobally(ctx context.Context, loggedInUser *uuid.UUID, limit int, nextPageToken *string) ([]domain.FeedItem, *string, error) {
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
	authors, err := as.UserService.GetUserListByUserIDs(ctx, uniqueAuthorIdsList)
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
		followedAuthorsSet, err = as.ProfileService.IsFollowingBulk(ctx, *loggedInUser, uniqueAuthorIdsList)
		if err != nil {
			return nil, nil, err
		}

		// fetch isFollowing in bulk
		articleIds := lo.Map(articles, func(article domain.Article, _ int) uuid.UUID { return article.Id })
		favoritedArticlesSet, err = as.ArticleRepository.IsFavoritedBulk(ctx, *loggedInUser, articleIds)
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

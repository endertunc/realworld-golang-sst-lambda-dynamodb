package service

import (
	"context"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/repository"
)

type ArticleListServiceInterface interface {
	GetMostRecentArticlesByAuthor(ctx context.Context, userId *uuid.UUID, author string, limit int, nextPageToken *string) ([]domain.ArticleAggregateView, *string, error)
	GetMostRecentArticlesFavoritedByUser(ctx context.Context, loggedInUser *uuid.UUID, favoritedByUsername string, limit int, nextPageToken *string) ([]domain.ArticleAggregateView, *string, error)
	GetMostRecentArticlesFavoritedByTag(ctx context.Context, loggedInUser *uuid.UUID, tag string, limit int, nextPageToken *string) ([]domain.ArticleAggregateView, *string, error)
	GetMostRecentArticlesGlobally(ctx context.Context, loggedInUser *uuid.UUID, limit int, nextPageToken *int) ([]domain.ArticleAggregateView, *int, error)
}

type articleListService struct {
	articleRepository           repository.ArticleRepositoryInterface
	articleOpensearchRepository repository.ArticleOpensearchRepositoryInterface
	userService                 UserServiceInterface
	profileService              ProfileServiceInterface
}

var _ ArticleListServiceInterface = articleListService{}

func NewArticleListService(
	articleRepository repository.ArticleRepositoryInterface,
	articleOpensearchRepository repository.ArticleOpensearchRepositoryInterface,
	userService UserServiceInterface,
	profileService ProfileServiceInterface) ArticleListServiceInterface {
	return articleListService{
		articleOpensearchRepository: articleOpensearchRepository,
		articleRepository:           articleRepository,
		userService:                 userService,
		profileService:              profileService,
	}
}

// articleRetrievalStrategy is used to provide different article retrival strategies (e.g. by author, by tag, globally) to collectArticlesWithMetadata function
type articleRetrievalStrategy[T any] func(ctx context.Context, limit int, offset *T) ([]domain.Article, *T, error)

// - - - - - - - - - - - - - - - - ArticlesWithMetadataResult - - - - - - - - - - - - - - - -
// ArticlesWithMetadataResult is an intermediate struct that holds the result of the common article aggregate query
type ArticlesWithMetadataResult struct {
	Articles             []domain.Article
	FollowedAuthorsSet   mapset.Set[uuid.UUID]
	FavoritedArticlesSet mapset.Set[uuid.UUID]
	AuthorsMap           map[uuid.UUID]domain.User
}

func (r ArticlesWithMetadataResult) toArticleAggregateView() []domain.ArticleAggregateView {
	articleAggregateViews := make([]domain.ArticleAggregateView, 0, len(r.Articles))
	for _, article := range r.Articles {
		author, authorFound := r.AuthorsMap[article.AuthorId]
		if authorFound {
			isFavorited := r.FavoritedArticlesSet.ContainsOne(article.Id)
			isFollowing := r.FollowedAuthorsSet.ContainsOne(article.AuthorId)

			articleAggregateViews = append(articleAggregateViews, domain.ArticleAggregateView{
				Article:     article,
				Author:      author,
				IsFavorited: isFavorited,
				IsFollowing: isFollowing,
			})
		}
	}
	return articleAggregateViews
}

// ToDo @ender to keep things short - we better pass limit and nextPageToken struct Pagination or something like that.
// ToDo @ender iteration #2 - finalize the code
func (al articleListService) GetMostRecentArticlesGlobally(ctx context.Context, loggedInUser *uuid.UUID, limit int, nextPageToken *int) ([]domain.ArticleAggregateView, *int, error) {
	result, nextToken, err := collectArticlesWithMetadata(ctx, al, loggedInUser, limit, nextPageToken, al.articleOpensearchRepository.FindAllArticles)

	if err != nil {
		return nil, nil, err
	}

	return result.toArticleAggregateView(), nextToken, nil
}

func (al articleListService) GetMostRecentArticlesByAuthor(ctx context.Context, loggedInUser *uuid.UUID, author string, limit int, nextPageToken *string) ([]domain.ArticleAggregateView, *string, error) {
	authorUser, err := al.userService.GetUserByUsername(ctx, author)
	if err != nil {
		return nil, nil, err
	}

	var articlesByAuthorProvider articleRetrievalStrategy[string] = func(ctx context.Context, limit int, nextPageToken *string) ([]domain.Article, *string, error) {
		return al.articleRepository.FindArticlesByAuthor(ctx, authorUser.Id, limit, nextPageToken)
	}

	result, nextToken, err := collectArticlesWithMetadata(ctx, al, loggedInUser, limit, nextPageToken, articlesByAuthorProvider)

	if err != nil {
		return nil, nil, err
	}

	return result.toArticleAggregateView(), nextToken, nil
}

func (al articleListService) GetMostRecentArticlesFavoritedByTag(ctx context.Context, loggedInUser *uuid.UUID, tag string, limit int, nextPageToken *string) ([]domain.ArticleAggregateView, *string, error) {
	var articlesByTagProvider articleRetrievalStrategy[string] = func(ctx context.Context, limit int, offset *string) ([]domain.Article, *string, error) {
		return al.articleOpensearchRepository.FindArticlesByTag(ctx, tag, limit, offset)
	}
	result, nextToken, err := collectArticlesWithMetadata[string](ctx, al, loggedInUser, limit, nextPageToken, articlesByTagProvider)
	if err != nil {
		return nil, nil, err
	}
	return result.toArticleAggregateView(), nextToken, nil
}

func (al articleListService) GetMostRecentArticlesFavoritedByUser(ctx context.Context, loggedInUser *uuid.UUID, favoritedByUsername string, limit int, nextPageToken *string) ([]domain.ArticleAggregateView, *string, error) {
	favoritedByUser, err := al.userService.GetUserByUsername(ctx, favoritedByUsername)
	if err != nil {
		return nil, nil, err
	}

	articleIds, nextToken, err := al.articleRepository.FindArticlesFavoritedByUser(ctx, favoritedByUser.Id, limit, nextPageToken)
	if err != nil {
		return nil, nil, err
	}

	var articlesFavoritedByUserProvider articleRetrievalStrategy[string] = func(ctx context.Context, limit int, nextPageToken *string) ([]domain.Article, *string, error) {
		articles, err := al.articleRepository.FindArticlesByIds(ctx, articleIds)
		if err != nil {
			return nil, nil, err
		}
		return articles, nextToken, nil
	}

	result, nextToken, err := collectArticlesWithMetadata(ctx, al, loggedInUser, limit, nextPageToken, articlesFavoritedByUserProvider)

	if err != nil {
		return nil, nil, err
	}

	authorIdToArticleMap := make(map[uuid.UUID]domain.Article)
	for _, article := range result.Articles {
		authorIdToArticleMap[article.Id] = article
	}
	articleAggregateViews := make([]domain.ArticleAggregateView, 0)
	// we need to return article in the order of articleIds
	for _, articleId := range articleIds {
		article, articleFound := authorIdToArticleMap[articleId]
		author, authorFound := result.AuthorsMap[article.AuthorId]

		// 1- we should have the article in the authorIdToArticleMap, otherwise let it skip
		// 2- we should have the author in the authorsMap, otherwise let it skip
		if articleFound && authorFound {
			isFavorited := result.FavoritedArticlesSet.ContainsOne(article.Id)
			isFollowing := result.FollowedAuthorsSet.ContainsOne(article.AuthorId)
			feedItem := domain.ArticleAggregateView{
				Article:     article,
				Author:      author,
				IsFavorited: isFavorited,
				IsFollowing: isFollowing,
			}
			articleAggregateViews = append(articleAggregateViews, feedItem)
		}
	}

	return articleAggregateViews, nextToken, nil
}

func collectArticlesWithMetadata[T any](ctx context.Context, al articleListService, loggedInUser *uuid.UUID, limit int, nextPageToken *T, articleProviderFunc articleRetrievalStrategy[T]) (ArticlesWithMetadataResult, *T, error) {
	// Fetch articles using the provided function
	articles, nextToken, err := articleProviderFunc(ctx, limit, nextPageToken)
	if err != nil {
		return ArticlesWithMetadataResult{}, nil, err
	}

	// Extract unique author IDs
	authorIdsList := lo.Map(articles, func(article domain.Article, _ int) uuid.UUID {
		return article.AuthorId
	})
	uniqueAuthorIdsList := lo.Uniq(authorIdsList)

	// Fetch authors in bulk
	authors, err := al.userService.GetUserListByUserIDs(ctx, uniqueAuthorIdsList)
	if err != nil {
		return ArticlesWithMetadataResult{}, nil, err
	}

	// Create authors map for efficient lookup
	authorsMap := make(map[uuid.UUID]domain.User)
	for _, author := range authors {
		authorsMap[author.Id] = author
	}

	// Initialize sets for tracking favorites and following
	followedAuthorsSet := mapset.NewThreadUnsafeSet[uuid.UUID]()
	favoritedArticlesSet := mapset.NewThreadUnsafeSet[uuid.UUID]()

	if loggedInUser != nil {
		// Bulk fetch following status
		followedAuthorsSet, err = al.profileService.IsFollowingBulk(ctx, *loggedInUser, uniqueAuthorIdsList)
		if err != nil {
			return ArticlesWithMetadataResult{}, nil, err
		}

		// Bulk fetch favorited status
		articleIds := lo.Map(articles, func(article domain.Article, _ int) uuid.UUID { return article.Id })
		favoritedArticlesSet, err = al.articleRepository.IsFavoritedBulk(ctx, *loggedInUser, articleIds)
		if err != nil {
			return ArticlesWithMetadataResult{}, nil, err
		}
	}

	return ArticlesWithMetadataResult{articles, followedAuthorsSet, favoritedArticlesSet, authorsMap}, nextToken, nil
}

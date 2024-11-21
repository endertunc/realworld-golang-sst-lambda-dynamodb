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
	GetMostRecentArticlesFavoritedByTag(ctx context.Context, loggedInUser *uuid.UUID, tag string, limit int, nextPageToken *int) ([]domain.ArticleAggregateView, *int, error)
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
func (al articleListService) GetMostRecentArticlesGloballyShortV2(ctx context.Context, loggedInUser *uuid.UUID, limit int, nextPageToken *int) ([]domain.ArticleAggregateView, *int, error) {
	result, nextToken, err := collectArticlesWithMetadata(ctx, al, loggedInUser, limit, nextPageToken, al.articleOpensearchRepository.FindAllArticles)

	if err != nil {
		return nil, nil, err
	}

	return result.toArticleAggregateView(), nextToken, nil
}

func (al articleListService) GetMostRecentArticlesByAuthorShortV2(ctx context.Context, loggedInUser *uuid.UUID, author string, limit int, nextPageToken *string) ([]domain.ArticleAggregateView, *string, error) {
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

func (al articleListService) GetMostRecentArticlesByTagShortV2(ctx context.Context, loggedInUser *uuid.UUID, tag string, limit int, nextPageToken *int) ([]domain.ArticleAggregateView, *int, error) {
	var articlesByTagProvider articleRetrievalStrategy[int] = func(ctx context.Context, limit int, offset *int) ([]domain.Article, *int, error) {
		return al.articleOpensearchRepository.FindArticlesByTag(ctx, tag, limit, offset)
	}
	result, nextToken, err := collectArticlesWithMetadata(ctx, al, loggedInUser, limit, nextPageToken, articlesByTagProvider)
	if err != nil {
		return nil, nil, err
	}
	return result.toArticleAggregateView(), nextToken, nil
}

func (al articleListService) GetMostRecentArticlesFavoritedByUserShortV2(ctx context.Context, loggedInUser *uuid.UUID, favoritedByUsername string, limit int, nextPageToken *string) ([]domain.ArticleAggregateView, *string, error) {
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
	feedItems := make([]domain.ArticleAggregateView, 0)
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
			feedItems = append(feedItems, feedItem)
		}
	}

	return nil, nextToken, nil
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

// ToDo @ender to be removed iteration #0
func (al articleListService) GetMostRecentArticlesByAuthor(ctx context.Context, loggedInUser *uuid.UUID, author string, limit int, nextPageToken *string) ([]domain.ArticleAggregateView, *string, error) {
	// find the author user by username
	authorUser, err := al.userService.GetUserByUsername(ctx, author)
	if err != nil {
		return nil, nil, err
	}

	articles, nextToken, err := al.articleRepository.FindArticlesByAuthor(ctx, authorUser.Id, limit, nextPageToken)
	if err != nil {
		return nil, nil, err
	}

	isFollowing := false
	favoritedArticlesSet := mapset.NewSet[uuid.UUID]()
	if loggedInUser != nil {
		// check if the user is following the author
		isFollowing, err = al.profileService.IsFollowing(ctx, *loggedInUser, authorUser.Id)
		if err != nil {
			return nil, nil, err
		}

		// fetch isFollowing in bulk
		articleIds := lo.Map(articles, func(article domain.Article, _ int) uuid.UUID {
			return article.Id
		})
		favoritedArticlesSet, err = al.articleRepository.IsFavoritedBulk(ctx, *loggedInUser, articleIds)
		if err != nil {
			return nil, nil, err
		}
	}

	feedItems := make([]domain.ArticleAggregateView, 0, len(articles))

	// we need to return articles in the order they are returned from the database
	for _, article := range articles {
		isFavorited := favoritedArticlesSet.ContainsOne(article.Id)
		feedItem := domain.ArticleAggregateView{
			Article:     article,
			Author:      authorUser,
			IsFavorited: isFavorited,
			IsFollowing: isFollowing,
		}
		feedItems = append(feedItems, feedItem)
	}

	return feedItems, nextToken, nil

}
func (al articleListService) GetMostRecentArticlesFavoritedByUser(ctx context.Context, loggedInUser *uuid.UUID, favoritedByUsername string, limit int, nextPageToken *string) ([]domain.ArticleAggregateView, *string, error) {
	// find the author user by username
	favoritedByUser, err := al.userService.GetUserByUsername(ctx, favoritedByUsername)
	if err != nil {

		return nil, nil, err
	}

	articleIds, nextToken, err := al.articleRepository.FindArticlesFavoritedByUser(ctx, favoritedByUser.Id, limit, nextPageToken)
	if err != nil {
		return nil, nil, err
	}

	articles, err := al.articleRepository.FindArticlesByIds(ctx, articleIds)
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
	authors, err := al.userService.GetUserListByUserIDs(ctx, uniqueAuthorIdsList)
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
		followedAuthorsSet, err = al.profileService.IsFollowingBulk(ctx, *loggedInUser, uniqueAuthorIdsList)
		if err != nil {
			return nil, nil, err
		}
		// fetch isFollowing in bulk
		favoritedArticlesSet, err = al.articleRepository.IsFavoritedBulk(ctx, *loggedInUser, articleIds)
		if err != nil {
			return nil, nil, err
		}
	}

	feedItems := make([]domain.ArticleAggregateView, 0)
	// we need to return article in the order of articleIds
	for _, articleId := range articleIds {
		article, articleFound := authorIdToArticleMap[articleId]
		author, authorFound := authorsMap[article.AuthorId]

		// 1- we should have the article in the authorIdToArticleMap, otherwise let it skip
		// 2- we should have the author in the authorsMap, otherwise let it skip
		if articleFound && authorFound {
			isFavorited := favoritedArticlesSet.ContainsOne(article.Id)
			isFollowing := followedAuthorsSet.ContainsOne(article.AuthorId)
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
func (al articleListService) GetMostRecentArticlesFavoritedByTag(ctx context.Context, loggedInUser *uuid.UUID, tag string, limit int, offset *int) ([]domain.ArticleAggregateView, *int, error) {
	articles, nextToken, err := al.articleOpensearchRepository.FindArticlesByTag(ctx, tag, limit, offset)

	if err != nil {
		return nil, nil, err
	}

	authorIdsList := lo.Map(articles, func(article domain.Article, _ int) uuid.UUID { return article.AuthorId })
	uniqueAuthorIdsList := lo.Uniq(authorIdsList)

	// fetch authors (users) in bulk and create a map for lookup by authorId
	authors, err := al.userService.GetUserListByUserIDs(ctx, uniqueAuthorIdsList)
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
		followedAuthorsSet, err = al.profileService.IsFollowingBulk(ctx, *loggedInUser, uniqueAuthorIdsList)
		if err != nil {
			return nil, nil, err
		}
		// fetch isFollowing in bulk
		articleIds := lo.Map(articles, func(article domain.Article, _ int) uuid.UUID { return article.Id })
		favoritedArticlesSet, err = al.articleRepository.IsFavoritedBulk(ctx, *loggedInUser, articleIds)
		if err != nil {
			return nil, nil, err
		}
	}

	feedItems := make([]domain.ArticleAggregateView, 0)
	// we need to return article in the order of articleIds
	for _, article := range articles {
		author, authorFound := authorsMap[article.AuthorId]
		// 1- we should have the author in the authorsMap, otherwise let it skip
		if authorFound {
			isFavorited := favoritedArticlesSet.ContainsOne(article.Id)
			isFollowing := followedAuthorsSet.ContainsOne(article.AuthorId)
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
func (al articleListService) GetMostRecentArticlesGlobally(ctx context.Context, loggedInUser *uuid.UUID, limit int, offset *int) ([]domain.ArticleAggregateView, *int, error) {
	// this time we should fetch this information from elasticsearch
	articles, nextToken, err := al.articleOpensearchRepository.FindAllArticles(ctx, limit, offset)
	if err != nil {
		return nil, nil, err
	}

	// ToDo let's convert to this regular for loop?
	authorIdsList := lo.Map(articles, func(article domain.Article, _ int) uuid.UUID { return article.AuthorId })
	uniqueAuthorIdsList := lo.Uniq(authorIdsList)

	// fetch authors (users) in bulk and create a map for lookup by authorId
	authors, err := al.userService.GetUserListByUserIDs(ctx, uniqueAuthorIdsList)
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
		followedAuthorsSet, err = al.profileService.IsFollowingBulk(ctx, *loggedInUser, uniqueAuthorIdsList)
		if err != nil {
			return nil, nil, err
		}

		// fetch isFollowing in bulk
		articleIds := lo.Map(articles, func(article domain.Article, _ int) uuid.UUID { return article.Id })
		favoritedArticlesSet, err = al.articleRepository.IsFavoritedBulk(ctx, *loggedInUser, articleIds)
		if err != nil {
			return nil, nil, err
		}
	}

	feedItems := make([]domain.ArticleAggregateView, 0)
	// we need to return articles in the order they are returned from the database
	for _, article := range articles {
		author, authorFound := authorsMap[article.AuthorId]
		// 1- we should have the author in the authorsMap, otherwise let it skip
		if authorFound {
			isFavorited := favoritedArticlesSet.ContainsOne(article.Id)
			isFollowing := followedAuthorsSet.ContainsOne(article.AuthorId)
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

// ToDo @ender to be removed iteration #1
type deps struct {
	userService       UserServiceInterface
	profileService    ProfileServiceInterface
	articleRepository repository.ArticleRepositoryInterface
}

func getMostRecentArticles[T any](ctx context.Context, deps deps, loggedInUser *uuid.UUID, limit int, nextPageToken *T, articleProviderFunc articleRetrievalStrategy[T]) ([]domain.ArticleAggregateView, *T, error) {
	// Fetch articles using the provided function
	articles, nextToken, err := articleProviderFunc(ctx, limit, nextPageToken)
	if err != nil {
		return nil, nil, err
	}

	// Extract unique author IDs
	authorIdsList := lo.Map(articles, func(article domain.Article, _ int) uuid.UUID {
		return article.AuthorId
	})
	uniqueAuthorIdsList := lo.Uniq(authorIdsList)

	// Fetch authors in bulk
	authors, err := deps.userService.GetUserListByUserIDs(ctx, uniqueAuthorIdsList)
	if err != nil {
		return nil, nil, err
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
		followedAuthorsSet, err = deps.profileService.IsFollowingBulk(ctx, *loggedInUser, uniqueAuthorIdsList)
		if err != nil {
			return nil, nil, err
		}

		// Bulk fetch favorited status
		articleIds := lo.Map(articles, func(article domain.Article, _ int) uuid.UUID { return article.Id })
		favoritedArticlesSet, err = deps.articleRepository.IsFavoritedBulk(ctx, *loggedInUser, articleIds)
		if err != nil {
			return nil, nil, err
		}
	}

	// Construct article aggregate view items
	articleAggregateViews := make([]domain.ArticleAggregateView, 0, len(articles))
	for _, article := range articles {
		author, authorFound := authorsMap[article.AuthorId]
		if authorFound {
			isFavorited := favoritedArticlesSet.ContainsOne(article.Id)
			isFollowing := followedAuthorsSet.ContainsOne(article.AuthorId)

			articleAggregateViews = append(articleAggregateViews, domain.ArticleAggregateView{
				Article:     article,
				Author:      author,
				IsFavorited: isFavorited,
				IsFollowing: isFollowing,
			})
		}
	}

	return articleAggregateViews, nextToken, nil
}
func (al articleListService) GetMostRecentArticlesByAuthorShort(ctx context.Context, loggedInUser *uuid.UUID, author string, limit int, nextPageToken *string) ([]domain.ArticleAggregateView, *string, error) {
	var articlesByAuthorProvider articleRetrievalStrategy[string] = func(ctx context.Context, limit int, nextPageToken *string) ([]domain.Article, *string, error) {
		authorUser, err := al.userService.GetUserByUsername(ctx, author)
		if err != nil {
			return nil, nil, err
		}
		return al.articleRepository.FindArticlesByAuthor(ctx, authorUser.Id, limit, nextPageToken)
	}
	return getMostRecentArticles(ctx, deps{
		userService:       al.userService,
		profileService:    al.profileService,
		articleRepository: al.articleRepository,
	}, loggedInUser, limit, nextPageToken, articlesByAuthorProvider)

}
func (al articleListService) GetMostRecentArticlesFavoritedByUserShort(ctx context.Context, loggedInUser *uuid.UUID, favoritedByUsername string, limit int, nextPageToken *string) ([]domain.ArticleAggregateView, *string, error) {
	var articlesFavoritedByUserProvider articleRetrievalStrategy[string] = func(ctx context.Context, limit int, nextPageToken *string) ([]domain.Article, *string, error) {
		favoritedByUser, err := al.userService.GetUserByUsername(ctx, favoritedByUsername)
		if err != nil {
			return nil, nil, err
		}

		articleIds, nextToken, err := al.articleRepository.FindArticlesFavoritedByUser(ctx, favoritedByUser.Id, limit, nextPageToken)
		if err != nil {
			return nil, nil, err
		}

		articles, err := al.articleRepository.FindArticlesByIds(ctx, articleIds)
		if err != nil {
			return nil, nil, err
		}
		return articles, nextToken, nil
	}
	return getMostRecentArticles(ctx, deps{
		userService:       al.userService,
		profileService:    al.profileService,
		articleRepository: al.articleRepository,
	}, loggedInUser, limit, nextPageToken, articlesFavoritedByUserProvider)
}
func (al articleListService) GetMostRecentArticlesFavoritedByTagShort(ctx context.Context, loggedInUser *uuid.UUID, tag string, limit int, nextPageToken *int) ([]domain.ArticleAggregateView, *int, error) {
	var articlesByTagProvider articleRetrievalStrategy[int] = func(ctx context.Context, limit int, offset *int) ([]domain.Article, *int, error) {
		return al.articleOpensearchRepository.FindArticlesByTag(ctx, tag, limit, offset)
	}
	return getMostRecentArticles(ctx, deps{
		userService:       al.userService,
		profileService:    al.profileService,
		articleRepository: al.articleRepository,
	}, loggedInUser, limit, nextPageToken, articlesByTagProvider)
}
func (al articleListService) GetMostRecentArticlesGloballyShort(ctx context.Context, loggedInUser *uuid.UUID, limit int, nextPageToken *int) ([]domain.ArticleAggregateView, *int, error) {
	return getMostRecentArticles(ctx, deps{
		userService:       al.userService,
		profileService:    al.profileService,
		articleRepository: al.articleRepository,
	}, loggedInUser, limit, nextPageToken, al.articleOpensearchRepository.FindAllArticles)
}

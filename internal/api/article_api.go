package api

import (
	"context"
	"github.com/google/uuid"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"realworld-aws-lambda-dynamodb-golang/internal/service"
)

type ArticleApi struct {
	articleService     service.ArticleServiceInterface
	articleListService service.ArticleListServiceInterface
	userService        service.UserServiceInterface
	profileService     service.ProfileServiceInterface
}

func NewArticleApi(
	articleService service.ArticleServiceInterface,
	articleListService service.ArticleListServiceInterface,
	userService service.UserServiceInterface,
	profileService service.ProfileServiceInterface) ArticleApi {
	return ArticleApi{
		articleService:     articleService,
		articleListService: articleListService,
		userService:        userService,
		profileService:     profileService,
	}
}

func (aa ArticleApi) GetArticle(ctx context.Context, loggedInUserId *uuid.UUID, slug string) (dto.ArticleResponseBodyDTO, error) {
	article, err := aa.articleService.GetArticle(ctx, slug)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}
	author, err := aa.userService.GetUserByUserId(ctx, article.AuthorId)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}
	if loggedInUserId == nil {
		return dto.ToArticleResponseBodyDTO(article, author, false, false), nil
	} else {
		loggedInUser, err := aa.userService.GetUserByUserId(ctx, *loggedInUserId)
		if err != nil {
			return dto.ArticleResponseBodyDTO{}, err
		}

		// ToDo @ender we make multiple request. We could optimize this by using BatchGetItem - isFollowing and isFavorited
		isFollowing, err := aa.profileService.IsFollowing(ctx, loggedInUser.Id, article.AuthorId)
		if err != nil {
			return dto.ArticleResponseBodyDTO{}, err
		}

		isFavorited, err := aa.articleService.IsFavorited(ctx, article.Id, loggedInUser.Id)

		if err != nil {
			return dto.ArticleResponseBodyDTO{}, err
		}

		return dto.ToArticleResponseBodyDTO(article, author, isFavorited, isFollowing), nil
	}

}

func (aa ArticleApi) CreateArticle(ctx context.Context, loggedInUserId uuid.UUID, createArticleRequestBodyDTO dto.CreateArticleRequestBodyDTO) (dto.ArticleResponseBodyDTO, error) {
	articleDTO := createArticleRequestBodyDTO.Article
	// ToDo @ender do we have any business validation we should apply in service level for an article?
	// ToDo @ender [GENERAL] - in this project we don't seem to have much complex data types to pass to services
	//  thus I skipped creating a struct that "service accepts" and simply passed the params needed to create and article
	//  Once this list of parameters that needs to be passed to service gets crowded,
	//  one could introduce intermediate "CreateArticleRequest" that articleService accepts
	article, err := aa.articleService.CreateArticle(
		ctx,
		loggedInUserId,
		articleDTO.Title,
		articleDTO.Description,
		articleDTO.Body,
		articleDTO.TagList)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}

	user, err := aa.userService.GetUserByUserId(ctx, loggedInUserId)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}

	// the current user is the author, and the user can't follow itself thus we simply pass isFollowing as false
	// the article has just been created thus we simply pass isFavorited as false
	return dto.ToArticleResponseBodyDTO(article, user, false, false), nil
}

func (aa ArticleApi) UnfavoriteArticle(ctx context.Context, loggedInUserId uuid.UUID, slug string) (dto.ArticleResponseBodyDTO, error) {
	article, err := aa.articleService.UnfavoriteArticle(ctx, loggedInUserId, slug)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}

	author, err := aa.userService.GetUserByUserId(ctx, loggedInUserId)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}

	// ToDo @ender test if the parameters passed to isFollowing are correct
	isFollowing, err := aa.profileService.IsFollowing(ctx, loggedInUserId, article.AuthorId)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}

	return dto.ToArticleResponseBodyDTO(article, author, false, isFollowing), nil
}

func (aa ArticleApi) FavoriteArticle(ctx context.Context, loggedInUserId uuid.UUID, slug string) (dto.ArticleResponseBodyDTO, error) {
	article, err := aa.articleService.FavoriteArticle(ctx, loggedInUserId, slug)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}

	author, err := aa.userService.GetUserByUserId(ctx, article.AuthorId)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}

	// ToDo @ender test if the parameters passed to isFollowing are correct
	isFollowing, err := aa.profileService.IsFollowing(ctx, loggedInUserId, article.AuthorId)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}

	return dto.ToArticleResponseBodyDTO(article, author, true, isFollowing), nil
}

func (aa ArticleApi) DeleteArticle(ctx context.Context, loggedInUserId uuid.UUID, slug string) error {
	err := aa.articleService.DeleteArticle(ctx, loggedInUserId, slug)
	if err != nil {
		return err
	}
	return nil
}

type ListArticlesQueryOptions struct {
	Author      *string
	FavoritedBy *string
	Tag         *string
}

func (aa ArticleApi) ListArticles(ctx context.Context, loggedInUserId *uuid.UUID, queryOptions ListArticlesQueryOptions, limit int, nextPageToken *string) (dto.MultipleArticlesResponseBodyDTO, error) {
	feedItems, newNextPageToken, err := func() ([]domain.ArticleAggregateView, *string, error) {
		if queryOptions.Author != nil {
			return aa.articleListService.GetMostRecentArticlesByAuthor(ctx, loggedInUserId, *queryOptions.Author, limit, nextPageToken)
		} else if queryOptions.FavoritedBy != nil {
			return aa.articleListService.GetMostRecentArticlesFavoritedByUser(ctx, loggedInUserId, *queryOptions.FavoritedBy, limit, nextPageToken)
		} else if queryOptions.Tag != nil {
			return aa.articleListService.GetMostRecentArticlesFavoritedByTag(ctx, loggedInUserId, *queryOptions.Tag, limit, nextPageToken)
		} else {
			return aa.articleListService.GetMostRecentArticlesGlobally(ctx, loggedInUserId, limit, nextPageToken)
		}
	}()

	if err != nil {
		return dto.MultipleArticlesResponseBodyDTO{}, err
	}
	return dto.ToMultipleArticlesResponseBodyDTO(feedItems, newNextPageToken), nil
}

func (aa ArticleApi) GetTags(ctx context.Context) (dto.TagsResponseDTO, error) {
	tags, err := aa.articleService.GetTags(ctx)
	if err != nil {
		return dto.TagsResponseDTO{}, err
	}
	return dto.TagsResponseDTO{Tags: tags}, nil
}

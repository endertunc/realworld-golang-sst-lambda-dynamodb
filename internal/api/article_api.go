package api

import (
	"context"
	"github.com/google/uuid"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"realworld-aws-lambda-dynamodb-golang/internal/service"
)

type ArticleApi struct {
	ArticleService service.ArticleServiceInterface
	UserService    service.UserServiceInterface
	ProfileService service.ProfileServiceInterface
}

func NewArticleApi(articleService service.ArticleServiceInterface, userService service.UserServiceInterface, profileService service.ProfileServiceInterface) ArticleApi {
	return ArticleApi{
		ArticleService: articleService,
		UserService:    userService,
		ProfileService: profileService,
	}
}

func (aa ArticleApi) GetArticle(ctx context.Context, loggedInUserId *uuid.UUID, slug string) (dto.ArticleResponseBodyDTO, error) {
	// ToDo @ender [maybe not] --- we could get article and author in BatchGetItem
	article, err := aa.ArticleService.GetArticle(ctx, slug)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}
	author, err := aa.UserService.GetUserByUserId(ctx, article.AuthorId)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}
	if loggedInUserId == nil {
		return dto.ToArticleResponseBodyDTO(article, author, false, false), nil
	} else {
		loggedInUser, err := aa.UserService.GetUserByUserId(ctx, *loggedInUserId)
		if err != nil {
			return dto.ArticleResponseBodyDTO{}, err
		}

		// ToDo @ender we make multiple request. We could optimize this by using BatchGetItem - isFollowing and isFavorited
		isFollowing, err := aa.ProfileService.IsFollowing(ctx, loggedInUser.Id, article.AuthorId)
		if err != nil {
			return dto.ArticleResponseBodyDTO{}, err
		}

		isFavorited, err := aa.ArticleService.IsFavorited(ctx, article.Id, loggedInUser.Id)

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
	article, err := aa.ArticleService.CreateArticle(
		ctx,
		loggedInUserId,
		articleDTO.Title,
		articleDTO.Description,
		articleDTO.Body,
		articleDTO.TagList)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}

	user, err := aa.UserService.GetUserByUserId(ctx, loggedInUserId)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}

	// the current user is the author, and the user can't follow itself thus we simply pass isFollowing as false
	// the article has just been created thus we simply pass isFavorited as false
	return dto.ToArticleResponseBodyDTO(article, user, false, false), nil
}

func (aa ArticleApi) UnfavoriteArticle(ctx context.Context, loggedInUserId uuid.UUID, slug string) (dto.ArticleResponseBodyDTO, error) {
	article, err := aa.ArticleService.UnfavoriteArticle(ctx, loggedInUserId, slug)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}

	author, err := aa.UserService.GetUserByUserId(ctx, loggedInUserId)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}

	// ToDo @ender test if the parameters passed to isFollowing are correct
	isFollowing, err := aa.ProfileService.IsFollowing(ctx, loggedInUserId, article.AuthorId)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}

	return dto.ToArticleResponseBodyDTO(article, author, false, isFollowing), nil
}

func (aa ArticleApi) FavoriteArticle(ctx context.Context, loggedInUserId uuid.UUID, slug string) (dto.ArticleResponseBodyDTO, error) {
	article, err := aa.ArticleService.FavoriteArticle(ctx, loggedInUserId, slug)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}

	author, err := aa.UserService.GetUserByUserId(ctx, article.AuthorId)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}

	// ToDo @ender test if the parameters passed to isFollowing are correct
	isFollowing, err := aa.ProfileService.IsFollowing(ctx, loggedInUserId, article.AuthorId)
	if err != nil {
		return dto.ArticleResponseBodyDTO{}, err
	}

	return dto.ToArticleResponseBodyDTO(article, author, true, isFollowing), nil
}

func (aa ArticleApi) DeleteArticle(ctx context.Context, loggedInUserId uuid.UUID, slug string) error {
	err := aa.ArticleService.DeleteArticle(ctx, loggedInUserId, slug)
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
	feedItems, newNextPageToken, err := func() ([]domain.FeedItem, *string, error) {
		if queryOptions.Author != nil {
			return aa.ArticleService.GetMostRecentArticlesByAuthor(ctx, loggedInUserId, *queryOptions.Author, limit, nextPageToken)
		} else if queryOptions.FavoritedBy != nil {
			return aa.ArticleService.GetMostRecentArticlesFavoritedByUser(ctx, loggedInUserId, *queryOptions.FavoritedBy, limit, nextPageToken)
		} else if queryOptions.Tag != nil {
			return aa.ArticleService.GetMostRecentArticlesFavoritedByTag(ctx, loggedInUserId, *queryOptions.Tag, limit, nextPageToken)
		} else {
			return aa.ArticleService.GetMostRecentArticlesGlobally(ctx, loggedInUserId, limit, nextPageToken)
		}
	}()

	if err != nil {
		return dto.MultipleArticlesResponseBodyDTO{}, err
	}
	return dto.ToMultipleArticlesResponseBodyDTO(feedItems, newNextPageToken), nil
}

func (aa ArticleApi) GetTags(ctx context.Context) (dto.TagsResponseDTO, error) {
	tags, err := aa.ArticleService.GetTags(ctx)
	if err != nil {
		return dto.TagsResponseDTO{}, err
	}
	return dto.TagsResponseDTO{Tags: tags}, nil
}

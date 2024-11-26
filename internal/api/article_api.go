package api

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/service"
)

type ArticleApi struct {
	articleService     service.ArticleServiceInterface
	articleListService service.ArticleListServiceInterface
	userService        service.UserServiceInterface
	profileService     service.ProfileServiceInterface
	paginationConfig   PaginationConfig
}

func NewArticleApi(
	articleService service.ArticleServiceInterface,
	articleListService service.ArticleListServiceInterface,
	userService service.UserServiceInterface,
	profileService service.ProfileServiceInterface,
	paginationConfig PaginationConfig,
) ArticleApi {
	return ArticleApi{
		articleService:     articleService,
		articleListService: articleListService,
		userService:        userService,
		profileService:     profileService,
		paginationConfig:   paginationConfig,
	}
}

func (aa ArticleApi) GetArticle(ctx context.Context, w http.ResponseWriter, r *http.Request, loggedInUserId *uuid.UUID) {
	slug, ok := GetPathParamHTTP(ctx, w, r, "slug")
	if !ok {
		return
	}

	handleError := func(err error) {
		if errors.Is(err, errutil.ErrArticleNotFound) {
			slog.DebugContext(ctx, "article not found", slog.String("slug", slug), slog.Any("error", err))
			ToSimpleHTTPError(w, http.StatusNotFound, "article not found")
			return
		}

		ToInternalServerHTTPError(w, err)
		return
	}

	article, err := aa.articleService.GetArticle(ctx, slug)
	if err != nil {
		handleError(err)
		return
	}
	author, err := aa.userService.GetUserByUserId(ctx, article.AuthorId)
	if err != nil {
		handleError(err)
		return
	}
	if loggedInUserId == nil {
		resp := dto.ToArticleResponseBodyDTO(article, author, false, false)
		ToSuccessHTTPResponse(w, resp)
		return
	} else {
		loggedInUser, err := aa.userService.GetUserByUserId(ctx, *loggedInUserId)
		if err != nil {
			handleError(err)
			return
		}

		// ToDo @ender we make multiple request. We could optimize this by using BatchGetItem - isFollowing and isFavorited
		isFollowing, err := aa.profileService.IsFollowing(ctx, loggedInUser.Id, article.AuthorId)
		if err != nil {
			handleError(err)
			return
		}

		isFavorited, err := aa.articleService.IsFavorited(ctx, article.Id, loggedInUser.Id)

		if err != nil {
			handleError(err)
			return
		}

		resp := dto.ToArticleResponseBodyDTO(article, author, isFavorited, isFollowing)
		ToSuccessHTTPResponse(w, resp)
		return
	}

}

func (aa ArticleApi) CreateArticle(ctx context.Context, w http.ResponseWriter, r *http.Request, loggedInUserId uuid.UUID) {
	createArticleRequestBodyDTO, ok := ParseAndValidateBody[dto.CreateArticleRequestBodyDTO](ctx, w, r)

	if !ok {
		return
	}

	articleBody := createArticleRequestBodyDTO.Article
	// ToDo @ender do we have any business validation we should apply in service level for an article?
	// ToDo @ender [GENERAL] - in this project we don't seem to have much complex data types to pass to services
	//  thus I skipped creating a struct that "service accepts" and simply passed the params needed to create and article
	//  Once this list of parameters that needs to be passed to service gets crowded,
	//  one could introduce intermediate "CreateArticleRequest" that articleService accepts
	article, err := aa.articleService.CreateArticle(
		ctx,
		loggedInUserId,
		articleBody.Title,
		articleBody.Description,
		articleBody.Body,
		articleBody.TagList)
	if err != nil {
		ToInternalServerHTTPError(w, err)
		return
	}

	user, err := aa.userService.GetUserByUserId(ctx, loggedInUserId)
	if err != nil {
		ToInternalServerHTTPError(w, err)
		return
	}

	// the current user is the author, and the user can't follow itself thus we simply pass isFollowing as false
	// the article has just been created thus we simply pass isFavorited as false
	resp := dto.ToArticleResponseBodyDTO(article, user, false, false)
	ToSuccessHTTPResponse(w, resp)
	return
}

func (aa ArticleApi) UnfavoriteArticle(ctx context.Context, w http.ResponseWriter, r *http.Request, loggedInUserId uuid.UUID) {
	slug, ok := GetPathParamHTTP(ctx, w, r, "slug")
	if !ok {
		return
	}

	handleError := func(err error) {
		if errors.Is(err, errutil.ErrArticleNotFound) {
			slog.DebugContext(ctx, "article not found", slog.String("slug", slug))
			ToSimpleHTTPError(w, http.StatusNotFound, "article not found")
			return
		} else if errors.Is(err, errutil.ErrAlreadyUnfavorited) {
			slog.DebugContext(ctx, "article is already unfavorited", slog.String("slug", slug), slog.String("loggedInUserId", loggedInUserId.String()))
			ToSimpleHTTPError(w, http.StatusConflict, "article is already unfavorited")
			return
		} else {
			ToInternalServerHTTPError(w, err)
			return
		}
	}

	article, err := aa.articleService.UnfavoriteArticle(ctx, loggedInUserId, slug)
	if err != nil {
		handleError(err)
		return
	}

	author, err := aa.userService.GetUserByUserId(ctx, loggedInUserId)
	if err != nil {
		handleError(err)
		return
	}

	// ToDo @ender test if the parameters passed to isFollowing are correct
	isFollowing, err := aa.profileService.IsFollowing(ctx, loggedInUserId, article.AuthorId)
	if err != nil {
		handleError(err)
		return
	}

	resp := dto.ToArticleResponseBodyDTO(article, author, false, isFollowing)
	ToSuccessHTTPResponse(w, resp)
	return
}

func (aa ArticleApi) FavoriteArticle(ctx context.Context, w http.ResponseWriter, r *http.Request, loggedInUserId uuid.UUID) {
	slug, ok := GetPathParamHTTP(ctx, w, r, "slug")
	if !ok {
		return
	}

	handleError := func(err error) {
		if errors.Is(err, errutil.ErrArticleNotFound) {
			slog.DebugContext(ctx, "article not found", slog.String("slug", slug), slog.Any("error", err))
			ToSimpleHTTPError(w, http.StatusNotFound, "article not found")
			return
		}
		if errors.Is(err, errutil.ErrAlreadyFavorited) {
			slog.DebugContext(ctx, "article already favorited", slog.String("slug", slug), slog.String("userId", loggedInUserId.String()))
			ToSimpleHTTPError(w, http.StatusConflict, "article already favorited")
			return
		}
		ToInternalServerHTTPError(w, err)
		return
	}

	article, err := aa.articleService.FavoriteArticle(ctx, loggedInUserId, slug)
	if err != nil {
		handleError(err)
		return
	}

	author, err := aa.userService.GetUserByUserId(ctx, article.AuthorId)
	if err != nil {
		handleError(err)
		return
	}

	isFollowing, err := aa.profileService.IsFollowing(ctx, loggedInUserId, article.AuthorId)
	if err != nil {
		handleError(err)
		return
	}

	resp := dto.ToArticleResponseBodyDTO(article, author, true, isFollowing)
	ToSuccessHTTPResponse(w, resp)
	return
}

func (aa ArticleApi) DeleteArticle(ctx context.Context, w http.ResponseWriter, r *http.Request, loggedInUserId uuid.UUID) {
	slug, ok := GetPathParamHTTP(ctx, w, r, "slug")
	if !ok {
		return
	}

	err := aa.articleService.DeleteArticle(ctx, loggedInUserId, slug)
	if err != nil {
		if errors.Is(err, errutil.ErrArticleNotFound) {
			slog.DebugContext(ctx, "article not found", slog.String("slug", slug))
			ToSimpleHTTPError(w, http.StatusNotFound, "article not found")
			return
		}
		if errors.Is(err, errutil.ErrCantDeleteOthersArticle) {
			slog.DebugContext(ctx, "user can't delete others article", slog.String("slug", slug), slog.String("userId", loggedInUserId.String()))
			ToSimpleHTTPError(w, http.StatusForbidden, "forbidden")
			return
		}
		ToInternalServerHTTPError(w, err)
		return
	}
	ToSuccessHTTPResponse(w, nil)
	return
}

type ListArticlesQueryOptions struct {
	Author      *string
	FavoritedBy *string
	Tag         *string
}

func (aa ArticleApi) ListArticles(ctx context.Context, w http.ResponseWriter, r *http.Request, loggedInUserId *uuid.UUID) {
	queryOptions, limit, nextPageToken, ok := extractArticleListRequestParameters(ctx, w, r, aa.paginationConfig)
	if !ok {
		return
	}
	articleAggregateViews, newNextPageToken, err := func() ([]domain.ArticleAggregateView, *string, error) {
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
		if errors.Is(err, errutil.ErrUserNotFound) {
			ToSimpleHTTPError(w, http.StatusNotFound, "author not found")
			return
		}
		ToInternalServerHTTPError(w, err)
		return
	}
	// Success response
	ToSuccessHTTPResponse(w, dto.ToMultipleArticlesResponseBodyDTO(articleAggregateViews, newNextPageToken))
	return
}

func (aa ArticleApi) GetTags(ctx context.Context, w http.ResponseWriter) {
	tags, err := aa.articleService.GetTags(ctx)
	if err != nil {
		ToInternalServerHTTPError(w, err)
		return
	}
	resp := dto.TagsResponseDTO{Tags: tags}
	ToSuccessHTTPResponse(w, resp)
	return
}

func extractArticleListRequestParameters(ctx context.Context, w http.ResponseWriter, r *http.Request, config PaginationConfig) (ListArticlesQueryOptions, int, *string, bool) {
	limit, ok := GetIntQueryParamOrDefault(ctx, w, r, "limit", config.DefaultLimit, &config.MinLimit, &config.MaxLimit)
	if !ok {
		return ListArticlesQueryOptions{}, 0, nil, ok
	}
	offset, ok := GetOptionalStringQueryParam(w, r, "offset")
	if !ok {
		return ListArticlesQueryOptions{}, 0, nil, ok
	}

	author, ok := GetOptionalStringQueryParam(w, r, "author")
	if !ok {
		return ListArticlesQueryOptions{}, 0, nil, ok
	}
	favoritedBy, ok := GetOptionalStringQueryParam(w, r, "favorited")
	if !ok {
		return ListArticlesQueryOptions{}, 0, nil, ok
	}
	tag, ok := GetOptionalStringQueryParam(w, r, "tag")
	if !ok {
		return ListArticlesQueryOptions{}, 0, nil, ok
	}

	listArticlesQueryOptions := ListArticlesQueryOptions{
		Author:      author,
		FavoritedBy: favoritedBy,
		Tag:         tag,
	}

	return listArticlesQueryOptions, limit, offset, ok
}

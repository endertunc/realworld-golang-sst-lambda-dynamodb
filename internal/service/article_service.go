package service

import (
	"context"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/repository"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
)

type articleService struct {
	articleRepository           repository.ArticleRepositoryInterface
	articleOpensearchRepository repository.ArticleOpensearchRepositoryInterface
	userService                 UserServiceInterface
	profileService              ProfileServiceInterface
}

type ArticleServiceInterface interface {
	GetArticle(ctx context.Context, slug string) (domain.Article, error)
	GetArticlesByIds(ctx context.Context, articleIds []uuid.UUID) ([]domain.Article, error)
	GetArticleBySlug(ctx context.Context, slug string) (domain.Article, error)

	CreateArticle(ctx context.Context, author uuid.UUID, title, description, body string, tagList []string) (domain.Article, error)
	UpdateArticle(ctx context.Context, authorId uuid.UUID, slug string, title, description, body *string) (domain.Article, error)
	DeleteArticle(ctx context.Context, author uuid.UUID, slug string) error

	FavoriteArticle(ctx context.Context, userId uuid.UUID, slug string) (domain.Article, error)
	UnfavoriteArticle(ctx context.Context, userId uuid.UUID, slug string) (domain.Article, error)
	IsFavorited(ctx context.Context, articleId, userId uuid.UUID) (bool, error)
	IsFavoritedBulk(ctx context.Context, userId uuid.UUID, articleIds []uuid.UUID) (mapset.Set[uuid.UUID], error)

	GetTags(ctx context.Context) ([]string, error)
}

var _ ArticleServiceInterface = articleService{} //nolint:golint,exhaustruct

func NewArticleService(
	articleRepository repository.ArticleRepositoryInterface,
	articleOpensearchRepository repository.ArticleOpensearchRepositoryInterface,
	userService UserServiceInterface,
	profileService ProfileServiceInterface) ArticleServiceInterface {
	return articleService{
		articleRepository:           articleRepository,
		articleOpensearchRepository: articleOpensearchRepository,
		userService:                 userService,
		profileService:              profileService,
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
	// Note we don't seem to have any business validation in this example application,
	// but we could add it here if needed.
	article := domain.NewArticle(title, description, body, tagList, author)
	article, err := as.articleRepository.CreateArticle(ctx, article)
	if err != nil {
		return domain.Article{}, err
	}
	return article, nil
}

func (as articleService) UpdateArticle(ctx context.Context, authorId uuid.UUID, slug string, title, description, body *string) (domain.Article, error) {
	article, err := as.articleRepository.FindArticleBySlug(ctx, slug)
	if err != nil {
		return domain.Article{}, err
	}

	if article.AuthorId != authorId {
		return domain.Article{}, errutil.ErrCantUpdateOthersArticle
	}

	// Update fields if provided
	if title != nil {
		article.Title = *title
		article.Slug = domain.GenerateSlug(*title)
	}
	if description != nil {
		article.Description = *description
	}
	if body != nil {
		article.Body = *body
	}
	article.UpdatedAt = time.Now().Truncate(time.Millisecond)

	updatedArticle, err := as.articleRepository.UpdateArticle(ctx, article, slug)
	if err != nil {
		return domain.Article{}, err
	}

	return updatedArticle, nil
}

func (as articleService) UnfavoriteArticle(ctx context.Context, loggedInUserId uuid.UUID, slug string) (domain.Article, error) {
	article, err := as.articleRepository.FindArticleBySlug(ctx, slug)
	if err != nil {
		return domain.Article{}, err
	}

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
		return errutil.ErrCantDeleteOthersArticle
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

func (as articleService) GetTags(ctx context.Context) ([]string, error) {
	return as.articleOpensearchRepository.FindAllTags(ctx)
}

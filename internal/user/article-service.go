package user

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"time"
)

type ArticleRepositoryInterface interface {
	FindArticleBySlug(c context.Context, email string) (domain.Article, error)
	FindArticleById(c context.Context, articleId uuid.UUID) (domain.Article, error)
	CreateArticle(c context.Context, article domain.Article) (domain.Article, error)
	UnfavoriteArticle(c context.Context, loggedInUserId uuid.UUID, articleId uuid.UUID) error
	FavoriteArticle(c context.Context, loggedInUserId uuid.UUID, articleId uuid.UUID) error
	DeleteCommentByArticleIdAndCommentId(c context.Context, loggedInUserId uuid.UUID, articleId uuid.UUID, commentId uuid.UUID) error
	GetCommentsByArticleId(c context.Context, articleId uuid.UUID) ([]domain.Comment, error)
	CreateComment(c context.Context, comment domain.Comment) error
	DeleteArticleById(c context.Context, articleId uuid.UUID) error
}

func (as ArticleService) GetArticle(c context.Context, slug string) (domain.Article, error) {
	article, err := as.ArticleRepository.FindArticleBySlug(c, slug)
	if err != nil {
		return domain.Article{}, err
	}
	return article, nil
}

func (aa ArticleService) CreateArticle(c context.Context, author uuid.UUID, title, description, body string, tagList []string) (domain.Article, error) {
	now := time.Now()
	article := domain.Article{
		Id:             uuid.New(),
		Title:          title,
		Slug:           slug.Make(title), // ToDo generate slug
		Description:    description,
		Body:           body,
		TagList:        tagList,
		FavoritesCount: 0,
		AuthorId:       author,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	// ToDo @ender do we have any business validation we should apply in service level for an article?
	article, err := aa.ArticleRepository.CreateArticle(c, article)
	if err != nil {
		return domain.Article{}, err
	}
	return article, nil
}

func (aa ArticleService) AddComment(c context.Context, author uuid.UUID, articleSlug string, body string) (domain.Comment, error) {
	article, err := aa.ArticleRepository.FindArticleBySlug(c, articleSlug)
	if err != nil {
		return domain.Comment{}, err
	}
	now := time.Now()
	comment := domain.Comment{
		Id:        uuid.New(),
		ArticleId: article.Id,
		AuthorId:  author,
		Body:      body,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = aa.ArticleRepository.CreateComment(c, comment)
	if err != nil {
		return domain.Comment{}, err
	}
	return comment, nil
}

func (ap ArticleService) UnfavoriteArticle(c context.Context, loggedInUserId uuid.UUID, slug string) (domain.Article, error) {
	article, err := ap.ArticleRepository.FindArticleBySlug(c, slug)
	if err != nil {
		return domain.Article{}, err
	}
	err = ap.ArticleRepository.UnfavoriteArticle(c, loggedInUserId, article.Id)
	if err != nil {
		return domain.Article{}, err
	}
	// ToDo @ender favorited should be false
	// ToDo @ender favoritesCount should be decreased by 1
	return article, nil
}

func (ap ArticleService) FavoriteArticle(c context.Context, loggedInUserId uuid.UUID, slug string) (domain.Article, error) {
	article, err := ap.ArticleRepository.FindArticleBySlug(c, slug)
	if err != nil {
		return domain.Article{}, err
	}
	err = ap.ArticleRepository.FavoriteArticle(c, loggedInUserId, article.Id)
	if err != nil {
		return domain.Article{}, err
	}
	// ToDo @ender favorited should be true
	// ToDo @ender favoritesCount should be increased by 1
	return article, nil
}

func (ap ArticleService) DeleteComment(c context.Context, loggedInUserId uuid.UUID, slug string, commentId uuid.UUID) error {
	article, err := ap.ArticleRepository.FindArticleBySlug(c, slug)
	if err != nil {
		return err
	}
	// ToDo @ender check if the comment belongs to the article
	// ToDo @ender check if the comment belongs to the user
	err = ap.ArticleRepository.DeleteCommentByArticleIdAndCommentId(c, loggedInUserId, article.Id, commentId)
	if err != nil {
		return err
	}
	return nil
}

func (ap ArticleService) GetArticleComments(c context.Context, slug string) ([]domain.Comment, error) {
	article, err := ap.ArticleRepository.FindArticleBySlug(c, slug)
	if err != nil {
		return []domain.Comment{}, err
	}
	comments, err := ap.ArticleRepository.GetCommentsByArticleId(c, article.Id)
	if err != nil {
		return []domain.Comment{}, err
	}
	return comments, nil
}

func (ap ArticleService) DeleteArticle(c context.Context, authorId uuid.UUID, slug string) error {
	article, err := ap.ArticleRepository.FindArticleBySlug(c, slug)
	if err != nil {
		return err
	}

	if article.AuthorId != authorId {
		// ToDo @ender return a proper error
		return errors.New("you can't touch this")
	}

	err = ap.ArticleRepository.DeleteArticleById(c, article.Id)
	if err != nil {
		return err
	}
	return nil
}

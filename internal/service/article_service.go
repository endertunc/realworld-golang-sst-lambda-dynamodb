package service

import (
	"context"
	"errors"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/repository"
)

type ArticleService struct {
	UserService       UserServiceInterface
	ArticleRepository repository.ArticleRepositoryInterface
}

type ArticleServiceInterface interface {
	GetArticle(c context.Context, slug string) (domain.Article, error)
	CreateArticle(c context.Context, author uuid.UUID, title, description, body string, tagList []string) (domain.Article, error)
	AddComment(c context.Context, loggedInUserId uuid.UUID, articleSlug string, body string) (domain.Comment, error)
	GetArticleComments(c context.Context, slug string) ([]domain.Comment, error)
	DeleteComment(c context.Context, author uuid.UUID, slug string, commentId uuid.UUID) error
	DeleteArticle(c context.Context, author uuid.UUID, slug string) error
	FavoriteArticle(c context.Context, userId uuid.UUID, slug string) (domain.Article, error)
	UnfavoriteArticle(c context.Context, userId uuid.UUID, slug string) (domain.Article, error)
	IsFavorited(c context.Context, articleId, userId uuid.UUID) (bool, error)
	FindArticlesByIds(c context.Context, articleIds []uuid.UUID) ([]domain.Article, error)
	IsFavoritedBulk(c context.Context, userId uuid.UUID, articleIds []uuid.UUID) (mapset.Set[uuid.UUID], error)
	//UpdateArticle(c context.Context, loggedInUserId uuid.UUID) (domain.Token, domain.User, error)
}

var _ ArticleServiceInterface = ArticleService{}

func NewArticleService(userService UserServiceInterface, articleRepository repository.ArticleRepositoryInterface) ArticleService {
	return ArticleService{
		UserService:       userService,
		ArticleRepository: articleRepository,
	}
}

func (as ArticleService) GetArticle(c context.Context, slug string) (domain.Article, error) {
	article, err := as.ArticleRepository.FindArticleBySlug(c, slug)
	if err != nil {
		return domain.Article{}, err
	}
	return article, nil
}

func (as ArticleService) CreateArticle(c context.Context, author uuid.UUID, title, description, body string, tagList []string) (domain.Article, error) {

	article := domain.NewArticle(title, description, body, tagList, author)
	// ToDo @ender do we have any business validation we should apply in service level for an article?
	article, err := as.ArticleRepository.CreateArticle(c, article)
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

func (as ArticleService) UnfavoriteArticle(c context.Context, loggedInUserId uuid.UUID, slug string) (domain.Article, error) {
	article, err := as.ArticleRepository.FindArticleBySlug(c, slug)
	if err != nil {
		return domain.Article{}, err
	}
	// ToDo @ender check if the user already unfavorited the article
	err = as.ArticleRepository.UnfavoriteArticle(c, loggedInUserId, article.Id)
	if err != nil {
		return domain.Article{}, err
	}
	article.FavoritesCount--
	return article, nil
}

func (as ArticleService) FavoriteArticle(c context.Context, loggedInUserId uuid.UUID, slug string) (domain.Article, error) {
	article, err := as.ArticleRepository.FindArticleBySlug(c, slug)
	if err != nil {
		return domain.Article{}, err
	}
	err = as.ArticleRepository.FavoriteArticle(c, loggedInUserId, article.Id)
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

func (as ArticleService) GetArticleComments(c context.Context, slug string) ([]domain.Comment, error) {
	article, err := as.ArticleRepository.FindArticleBySlug(c, slug)
	if err != nil {
		return []domain.Comment{}, err
	}
	comments, err := as.ArticleRepository.FindCommentsByArticleId(c, article.Id)
	if err != nil {
		return []domain.Comment{}, err
	}
	return comments, nil
}

func (as ArticleService) DeleteArticle(c context.Context, authorId uuid.UUID, slug string) error {
	article, err := as.ArticleRepository.FindArticleBySlug(c, slug)
	if err != nil {
		return err
	}

	if article.AuthorId != authorId {
		// ToDo @ender return a proper error
		return errors.New("you can't touch this")
	}

	err = as.ArticleRepository.DeleteArticleById(c, article.Id)
	if err != nil {
		return err
	}
	return nil
}

func (as ArticleService) IsFavorited(c context.Context, articleId, userId uuid.UUID) (bool, error) {
	return as.ArticleRepository.IsFavorited(c, articleId, userId)
}

func (as ArticleService) FindArticlesByIds(c context.Context, articleIds []uuid.UUID) ([]domain.Article, error) {
	return as.ArticleRepository.FindArticlesByIds(c, articleIds)
}

func (as ArticleService) IsFavoritedBulk(c context.Context, userId uuid.UUID, articleIds []uuid.UUID) (mapset.Set[uuid.UUID], error) {
	return as.ArticleRepository.IsFavoritedBulk(c, userId, articleIds)
}

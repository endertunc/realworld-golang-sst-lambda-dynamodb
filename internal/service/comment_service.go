package service

import (
	"context"
	"github.com/google/uuid"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/repository"
)

type commentService struct {
	commentRepository repository.CommentRepositoryInterface
	articleService    ArticleServiceInterface
}

type CommentServiceInterface interface {
	AddComment(ctx context.Context, loggedInUserId uuid.UUID, articleSlug string, body string) (domain.Comment, error)
	GetArticleComments(ctx context.Context, slug string) ([]domain.Comment, error)
	DeleteComment(ctx context.Context, author uuid.UUID, slug string, commentId uuid.UUID) error
}

var _ CommentServiceInterface = commentService{} //nolint:golint,exhaustruct

func NewCommentService(commentRepository repository.CommentRepositoryInterface, articleService ArticleServiceInterface) CommentServiceInterface {
	return commentService{
		commentRepository: commentRepository,
		articleService:    articleService,
	}
}

func (as commentService) AddComment(ctx context.Context, author uuid.UUID, articleSlug string, body string) (domain.Comment, error) {
	article, err := as.articleService.GetArticleBySlug(ctx, articleSlug)
	if err != nil {
		return domain.Comment{}, err
	}
	comment := domain.NewComment(article.Id, author, body)

	err = as.commentRepository.CreateComment(ctx, comment)
	if err != nil {
		return domain.Comment{}, err
	}
	return comment, nil
}

// DeleteComment
/**
 * ToDo @ender We can actually delete a comment only if the comment belongs to the user with a single query
 */
func (as commentService) DeleteComment(ctx context.Context, loggedInUserId uuid.UUID, slug string, commentId uuid.UUID) error {
	article, err := as.articleService.GetArticleBySlug(ctx, slug)
	if err != nil {
		return err
	}
	// check if the comment belongs to the article / comment exists
	comment, err := as.commentRepository.FindCommentByCommentIdAndArticleId(ctx, commentId, article.Id)
	if err != nil {
		return err
	}

	// check if the comment belongs to the user
	if comment.AuthorId != loggedInUserId {
		return errutil.ErrCantDeleteOthersComment
	}

	err = as.commentRepository.DeleteCommentByArticleIdAndCommentId(ctx, loggedInUserId, article.Id, commentId)
	if err != nil {
		return err
	}
	return nil
}

func (as commentService) GetArticleComments(ctx context.Context, slug string) ([]domain.Comment, error) {
	article, err := as.articleService.GetArticleBySlug(ctx, slug)
	if err != nil {
		return []domain.Comment{}, err
	}
	comments, err := as.commentRepository.FindCommentsByArticleId(ctx, article.Id)
	if err != nil {
		return []domain.Comment{}, err
	}
	return comments, nil
}

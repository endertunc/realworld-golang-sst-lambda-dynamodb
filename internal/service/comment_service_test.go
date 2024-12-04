package service

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/generator"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	repoMocks "realworld-aws-lambda-dynamodb-golang/internal/repository/mocks"
	serviceMocks "realworld-aws-lambda-dynamodb-golang/internal/service/mocks"
)

func TestCommentService_AddComment(t *testing.T) {
	ctx := context.Background()

	t.Run("successful comment creation", func(t *testing.T) {
		withCommentTestContext(t, func(tc commentTestContext) {
			// Setup test data
			article := generator.GenerateArticle()
			body := gofakeit.LoremIpsumSentence(gofakeit.Number(10, 50))
			author := uuid.New()

			// Setup expectations
			tc.mockArticleService.EXPECT().
				GetArticleBySlug(ctx, article.Slug).
				Return(article, nil)

			var capturedComment domain.Comment
			tc.mockCommentRepo.EXPECT().
				CreateComment(ctx, mock.MatchedBy(func(comment domain.Comment) bool {
					capturedComment = comment
					return comment.ArticleId == article.Id && comment.AuthorId == author && comment.Body == body
				})).
				Return(nil)

			// Execute
			comment, err := tc.commentService.AddComment(ctx, author, article.Slug, body)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, capturedComment.Id, comment.Id)
			assert.Equal(t, article.Id, comment.ArticleId)
			assert.Equal(t, author, comment.AuthorId)
			assert.Equal(t, body, comment.Body)
		})
	})

	t.Run("article not found", func(t *testing.T) {
		withCommentTestContext(t, func(tc commentTestContext) {
			// Setup test data
			article := generator.GenerateArticle()
			nonExistentSlug := "non-existent-slug"

			// Setup expectations
			tc.mockArticleService.EXPECT().
				GetArticleBySlug(ctx, nonExistentSlug).
				Return(domain.Article{}, errutil.ErrArticleNotFound)

			// Execute
			comment, err := tc.commentService.AddComment(ctx, article.AuthorId, nonExistentSlug, gofakeit.LoremIpsumSentence(20))

			// Assert
			assert.ErrorIs(t, err, errutil.ErrArticleNotFound)
			assert.Empty(t, comment)
		})
	})
}

func TestCommentService_GetArticleComments(t *testing.T) {
	ctx := context.Background()

	t.Run("successful get comments", func(t *testing.T) {
		withCommentTestContext(t, func(tc commentTestContext) {
			// Setup test data
			article := generator.GenerateArticle()
			expectedComments := []domain.Comment{
				generator.GenerateCommentWithArticleId(article.Id),
				generator.GenerateCommentWithArticleId(article.Id),
			}

			// Setup expectations
			tc.mockArticleService.EXPECT().
				GetArticleBySlug(ctx, article.Slug).
				Return(article, nil)

			tc.mockCommentRepo.EXPECT().
				FindCommentsByArticleId(ctx, article.Id).
				Return(expectedComments, nil)

			// Execute
			comments, err := tc.commentService.GetArticleComments(ctx, article.Slug)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, expectedComments, comments)
		})
	})

	t.Run("article not found", func(t *testing.T) {
		withCommentTestContext(t, func(tc commentTestContext) {
			// Setup test data
			nonExistentSlug := "non-existent-slug"

			// Setup expectations
			tc.mockArticleService.EXPECT().
				GetArticleBySlug(ctx, nonExistentSlug).
				Return(domain.Article{}, errutil.ErrArticleNotFound)

			// Execute
			comments, err := tc.commentService.GetArticleComments(ctx, nonExistentSlug)

			// Assert
			assert.ErrorIs(t, err, errutil.ErrArticleNotFound)
			assert.Empty(t, comments)
		})
	})
}

func TestCommentService_DeleteComment(t *testing.T) {
	ctx := context.Background()

	t.Run("successful comment deletion", func(t *testing.T) {
		withCommentTestContext(t, func(tc commentTestContext) {
			// Setup test data
			article := generator.GenerateArticle()
			comment := generator.GenerateCommentWithArticleId(article.Id)
			author := comment.AuthorId

			// Setup expectations
			tc.mockArticleService.EXPECT().
				GetArticleBySlug(ctx, article.Slug).
				Return(article, nil)

			tc.mockCommentRepo.EXPECT().
				FindCommentByCommentIdAndArticleId(ctx, comment.Id, article.Id).
				Return(comment, nil)

			tc.mockCommentRepo.EXPECT().
				DeleteCommentByArticleIdAndCommentId(ctx, author, article.Id, comment.Id).
				Return(nil)

			// Execute
			err := tc.commentService.DeleteComment(ctx, author, article.Slug, comment.Id)

			// Assert
			assert.NoError(t, err)
		})
	})

	t.Run("article not found", func(t *testing.T) {
		withCommentTestContext(t, func(tc commentTestContext) {
			// Setup test data
			nonExistentSlug := "non-existent-slug"
			article := generator.GenerateArticle()

			// Setup expectations
			tc.mockArticleService.EXPECT().
				GetArticleBySlug(ctx, nonExistentSlug).
				Return(domain.Article{}, errutil.ErrArticleNotFound)

			// Execute
			err := tc.commentService.DeleteComment(ctx, article.AuthorId, nonExistentSlug, uuid.New())

			// Assert
			assert.ErrorIs(t, err, errutil.ErrArticleNotFound)
		})
	})

	t.Run("comment not found", func(t *testing.T) {
		withCommentTestContext(t, func(tc commentTestContext) {
			// Setup test data
			article := generator.GenerateArticle()
			nonExistentCommentId := uuid.New()

			// Setup expectations
			tc.mockArticleService.EXPECT().
				GetArticleBySlug(ctx, article.Slug).
				Return(article, nil)

			tc.mockCommentRepo.EXPECT().
				FindCommentByCommentIdAndArticleId(ctx, nonExistentCommentId, article.Id).
				Return(domain.Comment{}, errutil.ErrCommentNotFound)

			// Execute
			err := tc.commentService.DeleteComment(ctx, article.AuthorId, article.Slug, nonExistentCommentId)

			// Assert
			assert.ErrorIs(t, err, errutil.ErrCommentNotFound)
		})
	})

	t.Run("unauthorized deletion", func(t *testing.T) {
		withCommentTestContext(t, func(tc commentTestContext) {
			// Setup test data
			article := generator.GenerateArticle()
			comment := generator.GenerateCommentWithArticleId(article.Id)
			differentUser := uuid.New()

			// Setup expectations
			tc.mockArticleService.EXPECT().
				GetArticleBySlug(ctx, article.Slug).
				Return(article, nil)

			tc.mockCommentRepo.EXPECT().
				FindCommentByCommentIdAndArticleId(ctx, comment.Id, article.Id).
				Return(comment, nil)

			// Execute
			err := tc.commentService.DeleteComment(ctx, differentUser, article.Slug, comment.Id)

			// Assert
			assert.ErrorIs(t, err, errutil.ErrCantDeleteOthersComment)
		})
	})
}

// - - - - - - - - - - - - - - - - Test Context - - - - - - - - - - - - - - - -

type commentTestContext struct {
	commentService     CommentServiceInterface
	mockCommentRepo    *repoMocks.MockCommentRepositoryInterface
	mockArticleService *serviceMocks.MockArticleServiceInterface
}

func createCommentTestContext(t *testing.T) commentTestContext {
	mockCommentRepo := repoMocks.NewMockCommentRepositoryInterface(t)
	mockArticleService := serviceMocks.NewMockArticleServiceInterface(t)
	commentService := NewCommentService(mockCommentRepo, mockArticleService)

	return commentTestContext{
		commentService:     commentService,
		mockCommentRepo:    mockCommentRepo,
		mockArticleService: mockArticleService,
	}
}

func withCommentTestContext(t *testing.T, testFunc func(tc commentTestContext)) {
	testFunc(createCommentTestContext(t))
}

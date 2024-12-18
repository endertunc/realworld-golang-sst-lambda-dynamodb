package repository

import (
	"context"
	"realworld-aws-lambda-dynamodb-golang/internal/database"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/generator"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/test"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var commentRepo = NewDynamodbCommentRepository(database.NewDynamoDBStore())

func TestCreateComment(t *testing.T) {
	ctx := context.Background()
	test.WithSetupAndTeardown(t, func() {
		t.Run("success", func(t *testing.T) {
			comment := generator.GenerateComment()
			err := commentRepo.CreateComment(ctx, comment)
			require.NoError(t, err)

			foundComment, err := commentRepo.FindCommentByCommentIdAndArticleId(context.Background(), comment.Id, comment.ArticleId)
			require.NoError(t, err)
			assert.Equal(t, comment.Id, foundComment.Id)
			assert.Equal(t, comment.ArticleId, foundComment.ArticleId)
			assert.Equal(t, comment.AuthorId, foundComment.AuthorId)
			assert.Equal(t, comment.Body, foundComment.Body)
		})
	})
}

func TestFindCommentsByArticleId(t *testing.T) {
	ctx := context.Background()
	test.WithSetupAndTeardown(t, func() {
		t.Run("success", func(t *testing.T) {
			articleId := uuid.New()
			comment1 := generator.GenerateCommentWithArticleId(articleId)
			comment2 := generator.GenerateCommentWithArticleId(articleId)
			comment3 := generator.GenerateComment() // different article

			require.NoError(t, commentRepo.CreateComment(ctx, comment1))
			require.NoError(t, commentRepo.CreateComment(ctx, comment2))
			require.NoError(t, commentRepo.CreateComment(ctx, comment3))

			comments, err := commentRepo.FindCommentsByArticleId(ctx, articleId)
			require.NoError(t, err)
			assert.Len(t, comments, 2)

			// Verify both comments are found
			commentIds := map[uuid.UUID]bool{
				comments[0].Id: true,
				comments[1].Id: true,
			}
			assert.True(t, commentIds[comment1.Id])
			assert.True(t, commentIds[comment2.Id])
		})

		t.Run("no comments for article", func(t *testing.T) {
			comments, err := commentRepo.FindCommentsByArticleId(ctx, uuid.New())
			require.NoError(t, err)
			assert.Empty(t, comments)
		})
	})
}

func TestFindCommentByCommentIdAndArticleId(t *testing.T) {
	ctx := context.Background()
	test.WithSetupAndTeardown(t, func() {
		t.Run("existing comment", func(t *testing.T) {
			comment := generator.GenerateComment()
			require.NoError(t, commentRepo.CreateComment(ctx, comment))

			foundComment, err := commentRepo.FindCommentByCommentIdAndArticleId(ctx, comment.Id, comment.ArticleId)
			require.NoError(t, err)
			assert.Equal(t, comment.Id, foundComment.Id)
			assert.Equal(t, comment.ArticleId, foundComment.ArticleId)
			assert.Equal(t, comment.AuthorId, foundComment.AuthorId)
			assert.Equal(t, comment.Body, foundComment.Body)
		})

		t.Run("non-existent comment", func(t *testing.T) {
			_, err := commentRepo.FindCommentByCommentIdAndArticleId(ctx, uuid.New(), uuid.New())
			assert.ErrorIs(t, err, errutil.ErrCommentNotFound)
		})
	})
}

func TestDeleteCommentByArticleIdAndCommentId(t *testing.T) {
	ctx := context.Background()
	test.WithSetupAndTeardown(t, func() {
		t.Run("success", func(t *testing.T) {
			comment := generator.GenerateComment()
			require.NoError(t, commentRepo.CreateComment(ctx, comment))

			err := commentRepo.DeleteCommentByArticleIdAndCommentId(ctx, comment.ArticleId, comment.Id)
			require.NoError(t, err)

			// Verify comment is deleted
			_, err = commentRepo.FindCommentByCommentIdAndArticleId(ctx, comment.Id, comment.ArticleId)
			assert.ErrorIs(t, err, errutil.ErrCommentNotFound)
		})

		t.Run("non-existent comment", func(t *testing.T) {
			err := commentRepo.DeleteCommentByArticleIdAndCommentId(ctx, uuid.New(), uuid.New())
			require.NoError(t, err) // Delete is idempotent
		})
	})
}

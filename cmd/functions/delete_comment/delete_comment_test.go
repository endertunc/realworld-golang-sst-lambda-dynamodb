package main

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	dtogen "realworld-aws-lambda-dynamodb-golang/internal/domain/dto/generator"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/test"
	"testing"
)

func TestAuthenticationScenarios(t *testing.T) {
	test.RunAuthenticationTests(t, test.SharedAuthenticationTestConfig{
		Method: "DELETE",
		Path:   "/api/articles/some-article/comments/some-comment-id",
	})
}

func TestSuccessfulCommentDeletion(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create a user and an article
		_, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)
		article := test.CreateArticle(t, dtogen.GenerateCreateArticleRequestDTO(), token)

		// Create a comment
		comment := test.CreateComment(t, article.Slug, dtogen.GenerateAddCommentRequestDTO(), token)

		// Verify comment exists before deletion
		test.VerifyCommentExists(t, article.Slug, comment.Id, token)

		// Delete the comment
		test.DeleteComment(t, article.Slug, comment.Id, token)

		// Verify the comment no longer exists
		test.VerifyCommentNotExists(t, article.Slug, comment.Id, token)
	})
}

func TestDeleteNonExistingComment(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create a user and an article
		_, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)
		article := test.CreateArticle(t, dtogen.GenerateCreateArticleRequestDTO(), token)

		// Try to delete a non-existing comment
		nonExistingCommentId := uuid.New().String()
		respBody := test.DeleteCommentWithResponse[errutil.SimpleError](t, article.Slug, nonExistingCommentId, token, http.StatusNotFound)
		assert.Equal(t, "comment not found", respBody.Message)
	})
}

func TestDeleteCommentWithUnrelatedArticle(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create a user and two articles
		_, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)
		article1 := test.CreateArticle(t, dtogen.GenerateCreateArticleRequestDTO(), token)
		article2Request := dto.CreateArticleRequestDTO{
			Title:       "This is a new kind of article",
			Description: "Much better than the old one",
			Body:        "Are you ready for this?",
			TagList:     []string{},
		}
		article2 := test.CreateArticle(t, article2Request, token)

		// Create a comment on article1
		comment := test.CreateComment(t, article1.Slug, dtogen.GenerateAddCommentRequestDTO(), token)

		// Try to delete the comment using article2's slug
		respBody := test.DeleteCommentWithResponse[errutil.SimpleError](t, article2.Slug, comment.Id, token, http.StatusNotFound)
		assert.Equal(t, "comment not found", respBody.Message)

		// Verify the comment still exists in the original article
		test.VerifyCommentExists(t, article1.Slug, comment.Id, token)
	})
}

func TestDeleteCommentWithNonExistingArticle(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create a user and an article
		_, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)
		article := test.CreateArticle(t, dtogen.GenerateCreateArticleRequestDTO(), token)

		// Create a comment
		comment := test.CreateComment(t, article.Slug, dtogen.GenerateAddCommentRequestDTO(), token)

		// Try to delete the comment using a non-existing article slug
		respBody := test.DeleteCommentWithResponse[errutil.SimpleError](t, "non-existing-article", comment.Id, token, http.StatusNotFound)
		assert.Equal(t, "article not found", respBody.Message)

		// Verify the comment still exists in the original article
		test.VerifyCommentExists(t, article.Slug, comment.Id, token)
	})
}

func TestDeleteCommentWithInvalidCommentId(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create a user and an article
		_, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)
		article := test.CreateArticle(t, dtogen.GenerateCreateArticleRequestDTO(), token)

		// Try to delete a comment with invalid UUID format
		respBody := test.DeleteCommentWithResponse[errutil.SimpleError](t, article.Slug, "not-a-uuid", token, http.StatusBadRequest)
		assert.Equal(t, "commentId path parameter must be a valid UUID", respBody.Message)
	})
}

func TestDeleteCommentAsNonOwner(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create two users
		_, ownerToken := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)
		nonOwner := dtogen.GenerateNewUserRequestUserDto()
		_, nonOwnerToken := test.CreateAndLoginUser(t, nonOwner)

		// Create an article and comment as the owner
		article := test.CreateArticle(t, dtogen.GenerateCreateArticleRequestDTO(), ownerToken)
		comment := test.CreateComment(t, article.Slug, dtogen.GenerateAddCommentRequestDTO(), ownerToken)

		// Try to delete the comment as non-owner
		respBody := test.DeleteCommentWithResponse[errutil.SimpleError](t, article.Slug, comment.Id, nonOwnerToken, http.StatusForbidden)
		assert.Equal(t, "forbidden", respBody.Message)

		// Verify the comment still exists
		test.VerifyCommentExists(t, article.Slug, comment.Id, ownerToken)
	})
}

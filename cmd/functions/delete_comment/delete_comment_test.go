package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
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
		article := test.CreateDefaultArticle(t, token)

		// Create a comment
		comment := test.CreateDefaultComment(t, article.Slug, token)

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
		article := test.CreateDefaultArticle(t, token)

		// Try to delete a non-existing comment
		nonExistingCommentId := uuid.New().String()
		var respBody errutil.SimpleError
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "DELETE",
			fmt.Sprintf("/api/articles/%s/comments/%s", article.Slug, nonExistingCommentId),
			http.StatusNotFound, &respBody, token)
		assert.Equal(t, "comment not found", respBody.Message)
	})
}

func TestDeleteCommentWithUnrelatedArticle(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create a user and two articles
		_, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)
		article1 := test.CreateArticleEntity(t, test.DefaultCreateArticleRequestDTO, token)
		article2Request := dto.CreateArticleRequestDTO{
			Title:       "This is a new kind of article",
			Description: "Much better than the old one",
			Body:        "Are you ready for this?",
			TagList:     []string{},
		}
		article2 := test.CreateArticleEntity(t, article2Request, token)

		// Create a comment on article1
		comment := test.CreateDefaultComment(t, article1.Slug, token)

		// Try to delete the comment using article2's slug
		var respBody errutil.SimpleError
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "DELETE",
			fmt.Sprintf("/api/articles/%s/comments/%s", article2.Slug, comment.Id),
			http.StatusNotFound, &respBody, token)
		assert.Equal(t, "comment not found", respBody.Message)

		// Verify the comment still exists in the original article
		test.VerifyCommentExists(t, article1.Slug, comment.Id, token)
	})
}

func TestDeleteCommentWithNonExistingArticle(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create a user and an article
		_, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)
		article := test.CreateDefaultArticle(t, token)

		// Create a comment
		comment := test.CreateDefaultComment(t, article.Slug, token)

		// Try to delete the comment using a non-existing article slug
		var respBody errutil.SimpleError
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "DELETE",
			fmt.Sprintf("/api/articles/%s/comments/%s", "non-existing-article", comment.Id),
			http.StatusNotFound, &respBody, token)
		assert.Equal(t, "article not found", respBody.Message)

		// Verify the comment still exists in the original article
		test.VerifyCommentExists(t, article.Slug, comment.Id, token)
	})
}

func TestDeleteCommentWithInvalidCommentId(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create a user and an article
		_, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)
		article := test.CreateDefaultArticle(t, token)

		// Try to delete a comment with invalid UUID format
		var respBody errutil.SimpleError
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "DELETE",
			fmt.Sprintf("/api/articles/%s/comments/%s", article.Slug, "not-a-uuid"),
			http.StatusBadRequest, &respBody, token)
		assert.Equal(t, "commentId path parameter must be a valid UUID", respBody.Message)
	})
}

func TestDeleteCommentAsNonOwner(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create two users
		_, ownerToken := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)
		nonOwner := dto.NewUserRequestUserDto{
			Username: "non-owner",
			Email:    "nonowner@example.com",
			Password: "password123",
		}
		_, nonOwnerToken := test.CreateAndLoginUser(t, nonOwner)

		// Create an article and comment as the owner
		article := test.CreateDefaultArticle(t, ownerToken)
		comment := test.CreateDefaultComment(t, article.Slug, ownerToken)

		// Try to delete the comment as non-owner
		var respBody errutil.SimpleError
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "DELETE",
			fmt.Sprintf("/api/articles/%s/comments/%s", article.Slug, comment.Id),
			http.StatusForbidden, &respBody, nonOwnerToken)
		assert.Equal(t, "forbidden", respBody.Message)

		// Verify the comment still exists
		test.VerifyCommentExists(t, article.Slug, comment.Id, ownerToken)
	})
}

package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/test"
	"testing"
)

func TestAuthenticationScenarios(t *testing.T) {
	test.RunAuthenticationTests(t, test.SharedAuthenticationTestConfig{
		Method: "POST",
		Path:   "/api/articles/some-article/comments",
	})
}

func TestSuccessfulCommentCreation(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create a user and an article
		user, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)
		article := test.CreateDefaultArticle(t, token)

		// Create a comment
		comment := test.CreateDefaultComment(t, article.Slug, token)

		// Create expected comment response
		expectedComment := dto.CommentResponseDTO{
			Body: test.DefaultAddCommentRequestDTO.Body,
			Author: dto.AuthorDTO{
				Username:  user.Username,
				Bio:       nil,
				Image:     nil,
				Following: false,
			},
			// dynamic fields
			Id:        comment.Id,
			CreatedAt: comment.CreatedAt,
			UpdatedAt: comment.UpdatedAt,
		}

		// Compare the entire struct
		assert.Equal(t, expectedComment, comment)

		// Verify dynamic fields separately
		assert.NotEmpty(t, comment.Id)
		assert.NotZero(t, comment.CreatedAt)
		assert.NotZero(t, comment.UpdatedAt)

		// Verify the comment appears in the article's comments
		test.VerifyCommentExists(t, article.Slug, comment.Id, token)
	})
}

func TestCommentOnNonExistentArticle(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create a user
		_, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)

		// Try to comment on a non-existent article
		comment := dto.AddCommentRequestDTO{
			Body: "This is a great article!",
		}

		reqBody := dto.AddCommentRequestBodyDTO{
			Comment: comment,
		}

		respBody := errutil.GenericError{}
		test.MakeAuthenticatedRequestAndParseResponse(t, reqBody, "POST",
			"/api/articles/non-existent-article/comments",
			http.StatusNotFound, &respBody, token)

		assert.Equal(t, "article not found", respBody.Message)
	})
}

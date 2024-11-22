package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	dtogen "realworld-aws-lambda-dynamodb-golang/internal/domain/dto/generator"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/test"
	"strings"
	"testing"
)

func TestAuthenticationScenarios(t *testing.T) {
	test.RunAuthenticationTests(t, test.SharedAuthenticationTestConfig{
		Method: "POST",
		Path:   "/api/articles/some-article/comments",
	})
}

func TestRequestValidation(t *testing.T) {
	// create a user
	_, token := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())

	tests := []test.ApiRequestValidationTest[dto.AddCommentRequestDTO]{
		{
			Name:  "missing comment",
			Input: dto.AddCommentRequestDTO{},
			ExpectedError: map[string]string{
				"Comment": "Comment is a required field",
			},
		},
		{
			Name: "empty comment body",
			Input: dto.AddCommentRequestDTO{
				Body: "",
			},
			ExpectedError: map[string]string{
				"Comment": "Comment is a required field",
			},
		},
		{
			Name: "blank comment body",
			Input: dto.AddCommentRequestDTO{
				Body: "     ",
			},
			ExpectedError: map[string]string{
				"Comment.Body": "Body cannot be blank",
			},
		},
		{
			Name: "comment body too long",
			Input: dto.AddCommentRequestDTO{
				Body: strings.Repeat("a", 4097),
			},
			ExpectedError: map[string]string{
				"Comment.Body": "Body must be a maximum of 4,096 characters in length",
			},
		},
	}

	createCommentRequest := func(t *testing.T, input dto.AddCommentRequestDTO) errutil.ValidationErrors {
		return test.CreateCommentWithResponse[errutil.ValidationErrors](t, "does-not-matter", input, token, http.StatusBadRequest)
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			test.TestValidation(t, tt, createCommentRequest)
		})
	}
}

func TestSuccessfulCommentCreation(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create a user and an article
		user, token := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())
		article := test.CreateArticle(t, dtogen.GenerateCreateArticleRequestDTO(), token)

		// Create a commentResp
		commentReq := dtogen.GenerateAddCommentRequestDTO()
		commentResp := test.CreateComment(t, article.Slug, commentReq, token)

		// Create expected commentResp response
		expectedComment := dto.CommentResponseDTO{
			Body: commentReq.Body,
			Author: dto.AuthorDTO{
				Username:  user.Username,
				Bio:       nil,
				Image:     nil,
				Following: false,
			},
			// dynamic fields
			Id:        commentResp.Id,
			CreatedAt: commentResp.CreatedAt,
			UpdatedAt: commentResp.UpdatedAt,
		}

		// Compare the entire struct
		assert.Equal(t, expectedComment, commentResp)

		// Verify dynamic fields separately
		assert.NotEmpty(t, commentResp.Id)
		assert.NotZero(t, commentResp.CreatedAt)
		assert.NotZero(t, commentResp.UpdatedAt)

		// Verify the commentResp appears in the article's comments
		test.VerifyCommentExists(t, article.Slug, commentResp.Id, token)
	})
}

func TestCommentOnNonExistentArticle(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create a user
		_, token := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())

		// Try to comment on a non-existent article
		reqBody := dtogen.GenerateAddCommentRequestDTO()

		respBody := test.CreateCommentWithResponse[errutil.SimpleError](t, "non-existent-article", reqBody, token, http.StatusNotFound)
		assert.Equal(t, "article not found", respBody.Message)
	})
}

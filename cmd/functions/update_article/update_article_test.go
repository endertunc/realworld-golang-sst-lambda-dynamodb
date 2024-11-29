package main

import (
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	dtogen "realworld-aws-lambda-dynamodb-golang/internal/domain/dto/generator"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/test"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthenticationScenarios(t *testing.T) {
	test.RunAuthenticationTests(t, test.SharedAuthenticationTestConfig{
		Method: "PUT",
		Path:   "/api/articles/some-article",
	})
}

func TestSuccessfulArticleUpdate(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create a user and an article
		_, token := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())
		article := test.CreateArticle(t, dtogen.GenerateCreateArticleRequestDTO(), token)

		// Update the article
		updateReq := dtogen.GenerateUpdateArticleRequestDTO()

		updatedArticle := test.UpdateArticle(t, article.Slug, updateReq, token)

		// Verify the update
		assert.Equal(t, *updateReq.Title, updatedArticle.Title)
		assert.Equal(t, *updateReq.Description, updatedArticle.Description)
		assert.Equal(t, *updateReq.Body, updatedArticle.Body)
		assert.NotEqual(t, article.Slug, updatedArticle.Slug) // Slug should change with title
	})
}

func TestUpdateNonExistingArticle(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create a user
		_, token := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())

		// Try to update a non-existing article
		newTitle := "Updated Title"
		updateReq := dto.UpdateArticleRequestDTO{
			Title: &newTitle,
		}

		respBody := test.UpdateArticleWithResponse[errutil.SimpleError](t, "non-existing-article", updateReq, token, http.StatusNotFound)
		assert.Equal(t, "article not found", respBody.Message)
	})
}

func TestUpdateArticleAsNonOwner(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create two users
		_, ownerToken := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())
		nonOwner := dtogen.GenerateNewUserRequestUserDto()
		_, nonOwnerToken := test.CreateAndLoginUser(t, nonOwner)

		// Create an article as owner
		article := test.CreateArticle(t, dtogen.GenerateCreateArticleRequestDTO(), ownerToken)

		// Try to update the article as non-owner
		newTitle := "Updated Title"
		updateReq := dto.UpdateArticleRequestDTO{
			Title: &newTitle,
		}

		respBody := test.UpdateArticleWithResponse[errutil.SimpleError](t, article.Slug, updateReq, nonOwnerToken, http.StatusForbidden)
		assert.Equal(t, "forbidden", respBody.Message)

		// Verify the article remains unchanged
		existingArticle := test.GetArticle(t, article.Slug, &ownerToken)
		assert.Equal(t, article.Title, existingArticle.Title)
		assert.Equal(t, article.Description, existingArticle.Description)
		assert.Equal(t, article.Body, existingArticle.Body)
	})
}

func TestPartialArticleUpdate(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create a user and an article
		_, token := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())
		article := test.CreateArticle(t, dtogen.GenerateCreateArticleRequestDTO(), token)

		// Update only the title
		newTitle := "Updated Title"
		updateReq := dto.UpdateArticleRequestDTO{
			Title: &newTitle,
		}

		updatedArticle := test.UpdateArticle(t, article.Slug, updateReq, token)

		// Verify only title changed
		assert.Equal(t, newTitle, updatedArticle.Title)
		assert.Equal(t, article.Description, updatedArticle.Description)
		assert.Equal(t, article.Body, updatedArticle.Body)
		assert.NotEqual(t, article.Slug, updatedArticle.Slug) // Slug should change with title
	})
}

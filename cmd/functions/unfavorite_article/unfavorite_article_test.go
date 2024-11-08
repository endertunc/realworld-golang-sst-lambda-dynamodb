package main

import (
	"fmt"
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
		Path:   "/api/articles/test-article/favorite",
	})
}

func TestSuccessfulUnfavorite(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create and login a user
		_, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)

		// Create an article
		createdArticle := test.CreateArticleEntity(t, test.DefaultCreateArticleRequestDTO, token)

		// First favorite the article
		var favoriteRespBody dto.ArticleResponseBodyDTO
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "POST", fmt.Sprintf("/api/articles/%s/favorite", createdArticle.Slug), http.StatusOK, &favoriteRespBody, token)

		// Verify article is favorited
		assert.True(t, favoriteRespBody.Article.Favorited)
		assert.Equal(t, 1, favoriteRespBody.Article.FavoritesCount)

		// Now unfavorite the article
		var unfavoriteRespBody dto.ArticleResponseBodyDTO
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "DELETE", fmt.Sprintf("/api/articles/%s/favorite", createdArticle.Slug), http.StatusOK, &unfavoriteRespBody, token)

		// Verify the response
		assert.Equal(t, createdArticle.Slug, unfavoriteRespBody.Article.Slug)
		assert.False(t, unfavoriteRespBody.Article.Favorited)
		assert.Equal(t, 0, unfavoriteRespBody.Article.FavoritesCount)

		// Verify the unfavorite status by getting the article
		var articleRespBody dto.ArticleResponseBodyDTO
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "GET", fmt.Sprintf("/api/articles/%s", createdArticle.Slug), http.StatusOK, &articleRespBody, token)
		assert.False(t, articleRespBody.Article.Favorited)
		assert.Equal(t, 0, articleRespBody.Article.FavoritesCount)
	})
}

func TestUnfavoriteNonExistentArticle(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create and login a user
		_, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)

		// Try to unfavorite non-existent article
		nonExistentSlug := "non-existent-article"
		var respBody errutil.GenericError
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "DELETE", fmt.Sprintf("/api/articles/%s/favorite", nonExistentSlug), http.StatusNotFound, &respBody, token)
		assert.Equal(t, "article not found", respBody.Message)
	})
}

func TestUnfavoriteAlreadyUnfavoritedArticle(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create and login a user
		_, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)

		// Create an article
		createdArticle := test.CreateArticleEntity(t, test.DefaultCreateArticleRequestDTO, token)

		// Try to unfavorite an article that was never favorited
		var respBody errutil.GenericError
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "DELETE", fmt.Sprintf("/api/articles/%s/favorite", createdArticle.Slug), http.StatusConflict, &respBody, token)
		assert.Equal(t, "article is already unfavorited", respBody.Message)

		// Verify article status hasn't changed
		var articleRespBody dto.ArticleResponseBodyDTO
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "GET", fmt.Sprintf("/api/articles/%s", createdArticle.Slug), http.StatusOK, &articleRespBody, token)
		assert.False(t, articleRespBody.Article.Favorited)
		assert.Equal(t, 0, articleRespBody.Article.FavoritesCount)
	})
}

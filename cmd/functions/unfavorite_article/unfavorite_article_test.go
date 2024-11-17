package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
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
		createdArticle := test.CreateArticle(t, dtogen.GenerateCreateArticleRequestDTO(), token)

		// First favorite the article
		favoriteRespBody := test.FavoriteArticle(t, createdArticle.Slug, token)

		// Verify article is favorited
		assert.True(t, favoriteRespBody.Favorited)
		assert.Equal(t, 1, favoriteRespBody.FavoritesCount)

		// Now unfavorite the article
		unfavoriteRespBody := test.UnfavoriteArticle(t, createdArticle.Slug, token)

		// Verify the response
		assert.Equal(t, createdArticle.Slug, unfavoriteRespBody.Slug)
		assert.False(t, unfavoriteRespBody.Favorited)
		assert.Equal(t, 0, unfavoriteRespBody.FavoritesCount)

		// Verify the unfavorite status by getting the article
		articleRespBody := test.GetArticle(t, createdArticle.Slug, &token)
		assert.False(t, articleRespBody.Favorited)
		assert.Equal(t, 0, articleRespBody.FavoritesCount)
	})
}

func TestUnfavoriteNonExistentArticle(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create and login a user
		_, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)

		// Try to unfavorite non-existent article
		nonExistentSlug := "non-existent-article"
		respBody := test.UnfavoriteArticleWithResponse[errutil.SimpleError](t, nonExistentSlug, token, http.StatusNotFound)
		assert.Equal(t, "article not found", respBody.Message)
	})
}

func TestUnfavoriteAlreadyUnfavoritedArticle(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create and login a user
		_, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)

		// Create an article
		createdArticle := test.CreateArticle(t, dtogen.GenerateCreateArticleRequestDTO(), token)

		// Try to unfavorite an article that was never favorited
		respBody := test.UnfavoriteArticleWithResponse[errutil.SimpleError](t, createdArticle.Slug, token, http.StatusConflict)
		assert.Equal(t, "article is already unfavorited", respBody.Message)

		// Verify article status hasn't changed
		articleRespBody := test.GetArticle(t, createdArticle.Slug, &token)
		assert.False(t, articleRespBody.Favorited)
		assert.Equal(t, 0, articleRespBody.FavoritesCount)
	})
}

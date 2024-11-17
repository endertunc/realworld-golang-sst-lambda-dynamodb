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
		Method: "POST",
		Path:   "/api/articles/test-article/favorite",
	})
}

func TestSuccessfulFavorite(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create and login a user
		_, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)

		// Create an article
		article := dtogen.GenerateCreateArticleRequestDTO()

		createdArticle := test.CreateArticle(t, article, token)

		// Favorite the article
		favoriteRespBody := test.FavoriteArticle(t, createdArticle.Slug, token)

		// Verify the response
		assert.Equal(t, createdArticle.Slug, favoriteRespBody.Slug)
		assert.True(t, favoriteRespBody.Favorited)
		assert.Equal(t, 1, favoriteRespBody.FavoritesCount)

		// Verify the favorite status by getting the article
		articleRespBody := test.GetArticle(t, createdArticle.Slug, &token)
		assert.True(t, articleRespBody.Favorited)
		assert.Equal(t, createdArticle.Slug, articleRespBody.Slug)
		assert.Equal(t, 1, articleRespBody.FavoritesCount)
	})
}

func TestFavoriteNonExistentArticle(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create and login a user
		_, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)

		// Try to favorite non-existent article
		nonExistentSlug := "non-existent-article"
		respBody := test.GetArticleWithResponse[errutil.SimpleError](t, nonExistentSlug, &token, http.StatusNotFound)
		assert.Equal(t, "article not found", respBody.Message)
	})
}

func TestFavoriteAlreadyFavoritedArticle(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create and login a user
		_, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)

		// Create an article
		article := dtogen.GenerateCreateArticleRequestDTO()

		createdArticle := test.CreateArticle(t, article, token)

		// Favorite the article first time
		favoriteRespBody := test.FavoriteArticle(t, createdArticle.Slug, token)

		// Verify initial favorite
		assert.True(t, favoriteRespBody.Favorited)
		assert.Equal(t, 1, favoriteRespBody.FavoritesCount)

		// Favorite the article second time
		respBody := test.FavoriteArticleWithResponse[errutil.SimpleError](t, createdArticle.Slug, token, http.StatusConflict)
		assert.Equal(t, "article already favorited", respBody.Message)

		// Verify the favorite status remains the same by getting the article
		articleRespBody := test.GetArticle(t, createdArticle.Slug, &token)
		assert.True(t, articleRespBody.Favorited)
		assert.Equal(t, createdArticle.Slug, articleRespBody.Slug)
		assert.Equal(t, 1, articleRespBody.FavoritesCount)
	})
}

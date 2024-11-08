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
		Method: "POST",
		Path:   "/api/articles/test-article/favorite",
	})
}

func TestSuccessfulFavorite(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create and login a user
		_, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)

		// Create an article
		article := test.DefaultCreateArticleRequestDTO

		createdArticle := test.CreateArticleEntity(t, article, token)

		// Favorite the article
		var favoriteRespBody dto.ArticleResponseBodyDTO
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "POST", fmt.Sprintf("/api/articles/%s/favorite", createdArticle.Slug), http.StatusOK, &favoriteRespBody, token)

		// Verify the response
		assert.Equal(t, createdArticle.Slug, favoriteRespBody.Article.Slug)
		assert.True(t, favoriteRespBody.Article.Favorited)
		assert.Equal(t, 1, favoriteRespBody.Article.FavoritesCount)

		// Verify the favorite status by getting the article
		var articleRespBody dto.ArticleResponseBodyDTO
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "GET", fmt.Sprintf("/api/articles/%s", createdArticle.Slug), http.StatusOK, &articleRespBody, token)
		assert.True(t, articleRespBody.Article.Favorited)
		assert.Equal(t, 1, articleRespBody.Article.FavoritesCount)
	})
}

func TestFavoriteNonExistentArticle(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create and login a user
		_, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)

		// Try to favorite non-existent article
		nonExistentSlug := "non-existent-article"
		var respBody errutil.SimpleError
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "POST", fmt.Sprintf("/api/articles/%s/favorite", nonExistentSlug), http.StatusNotFound, &respBody, token)
		assert.Equal(t, "article not found", respBody.Message)
	})
}

func TestFavoriteAlreadyFavoritedArticle(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create and login a user
		_, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)

		// Create an article
		article := test.DefaultCreateArticleRequestDTO

		createdArticle := test.CreateArticleEntity(t, article, token)

		// Favorite the article first time
		var favoriteRespBody dto.ArticleResponseBodyDTO
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "POST", fmt.Sprintf("/api/articles/%s/favorite", createdArticle.Slug), http.StatusOK, &favoriteRespBody, token)

		// Verify initial favorite
		assert.True(t, favoriteRespBody.Article.Favorited)
		assert.Equal(t, 1, favoriteRespBody.Article.FavoritesCount)

		// Favorite the article second time
		errorRespBody := errutil.SimpleError{}
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "POST", fmt.Sprintf("/api/articles/%s/favorite", createdArticle.Slug), http.StatusConflict, &errorRespBody, token)
		assert.Equal(t, "article already favorited", errorRespBody.Message)

		// Verify the favorite status remains the same by getting the article
		var articleRespBody dto.ArticleResponseBodyDTO
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "GET", fmt.Sprintf("/api/articles/%s", createdArticle.Slug), http.StatusOK, &articleRespBody, token)
		assert.True(t, articleRespBody.Article.Favorited)
		assert.Equal(t, 1, articleRespBody.Article.FavoritesCount)
	})
}

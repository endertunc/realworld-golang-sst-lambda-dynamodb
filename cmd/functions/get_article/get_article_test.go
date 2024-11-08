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

func TestGetNonExistentArticle(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		nonExistentSlug := "non-existent-article"
		var respBody errutil.GenericError
		test.MakeRequestAndParseResponse(t, nil, "GET", fmt.Sprintf("/api/articles/%s", nonExistentSlug), http.StatusNotFound, &respBody)
		assert.Equal(t, "article not found", respBody.Message)
	})
}

func TestGetArticleUnauthenticated(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create a user and article
		_, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)
		createdArticle := test.CreateArticleEntity(t, test.DefaultCreateArticleRequestDTO, token)

		// Get article without authentication
		var respBody dto.ArticleResponseBodyDTO
		test.MakeRequestAndParseResponse(t, nil, "GET", fmt.Sprintf("/api/articles/%s", createdArticle.Slug), http.StatusOK, &respBody)

		// Verify response
		assert.Equal(t, createdArticle.Slug, respBody.Article.Slug)
		assert.False(t, respBody.Article.Favorited)
		assert.False(t, respBody.Article.Author.Following)
		assert.Equal(t, 0, respBody.Article.FavoritesCount)
	})
}

func TestGetArticleNotFollowingNotFavorited(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create author and article
		_, authorToken := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)
		createdArticle := test.CreateArticleEntity(t, test.DefaultCreateArticleRequestDTO, authorToken)

		// Create reader
		_, readerToken := test.CreateAndLoginUser(t, dto.NewUserRequestUserDto{
			Username: "reader",
			Email:    "reader@test.com",
			Password: "password123",
		})

		// Get article as reader
		var respBody dto.ArticleResponseBodyDTO
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "GET", fmt.Sprintf("/api/articles/%s", createdArticle.Slug), http.StatusOK, &respBody, readerToken)

		// Verify response
		assert.Equal(t, createdArticle.Slug, respBody.Article.Slug)
		assert.False(t, respBody.Article.Favorited)
		assert.False(t, respBody.Article.Author.Following)
		assert.Equal(t, 0, respBody.Article.FavoritesCount)
	})
}

func TestGetArticleFollowingAndFavorited(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create author and article
		author, authorToken := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)
		createdArticle := test.CreateArticleEntity(t, test.DefaultCreateArticleRequestDTO, authorToken)

		// Create reader
		_, readerToken := test.CreateAndLoginUser(t, dto.NewUserRequestUserDto{
			Username: "reader",
			Email:    "reader@test.com",
			Password: "password123",
		})

		// Follow the author
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "POST", fmt.Sprintf("/api/profiles/%s/follow", author.Username), http.StatusOK, nil, readerToken)

		// Favorite the article
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "POST", fmt.Sprintf("/api/articles/%s/favorite", createdArticle.Slug), http.StatusOK, nil, readerToken)

		// Get article as reader
		var respBody dto.ArticleResponseBodyDTO
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "GET", fmt.Sprintf("/api/articles/%s", createdArticle.Slug), http.StatusOK, &respBody, readerToken)

		// Verify response
		assert.Equal(t, createdArticle.Slug, respBody.Article.Slug)
		assert.True(t, respBody.Article.Favorited)
		assert.True(t, respBody.Article.Author.Following)
		assert.Equal(t, 1, respBody.Article.FavoritesCount)
	})
}

func TestGetOwnArticleNotFavorited(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create author and article
		_, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)
		createdArticle := test.CreateArticleEntity(t, test.DefaultCreateArticleRequestDTO, token)

		// Get own article
		var respBody dto.ArticleResponseBodyDTO
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "GET", fmt.Sprintf("/api/articles/%s", createdArticle.Slug), http.StatusOK, &respBody, token)

		// Verify response
		assert.Equal(t, createdArticle.Slug, respBody.Article.Slug)
		assert.False(t, respBody.Article.Favorited)
		assert.False(t, respBody.Article.Author.Following) // You don't follow yourself
		assert.Equal(t, 0, respBody.Article.FavoritesCount)
	})
}

func TestGetOwnArticleFavorited(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create author and article
		_, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)
		createdArticle := test.CreateArticleEntity(t, test.DefaultCreateArticleRequestDTO, token)

		// Favorite own article
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "POST", fmt.Sprintf("/api/articles/%s/favorite", createdArticle.Slug), http.StatusOK, nil, token)

		// Get own article
		var respBody dto.ArticleResponseBodyDTO
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "GET", fmt.Sprintf("/api/articles/%s", createdArticle.Slug), http.StatusOK, &respBody, token)

		// Verify response
		assert.Equal(t, createdArticle.Slug, respBody.Article.Slug)
		assert.True(t, respBody.Article.Favorited)
		assert.False(t, respBody.Article.Author.Following) // You don't follow yourself
		assert.Equal(t, 1, respBody.Article.FavoritesCount)
	})
}

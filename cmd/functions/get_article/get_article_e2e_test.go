package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	dtogen "realworld-aws-lambda-dynamodb-golang/internal/domain/dto/generator"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/test"
	"testing"
)

func TestGetNonExistentArticle(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		nonExistentSlug := "non-existent-article"
		respBody := test.GetArticleWithResponse[errutil.SimpleError](t, nonExistentSlug, nil, http.StatusNotFound)
		assert.Equal(t, "article not found", respBody.Message)
	})
}

func TestGetArticleUnauthenticated(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create a user and article
		_, token := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())
		createdArticle := test.CreateArticle(t, dtogen.GenerateCreateArticleRequestDTO(), token)

		// Get article without authentication
		respBody := test.GetArticle(t, createdArticle.Slug, nil)

		// Verify response
		assert.Equal(t, createdArticle.Slug, respBody.Slug)
		assert.False(t, respBody.Favorited)
		assert.False(t, respBody.Author.Following)
		assert.Equal(t, 0, respBody.FavoritesCount)
	})
}

func TestGetArticleNotFollowingNotFavorited(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create author and article
		_, authorToken := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())
		createdArticle := test.CreateArticle(t, dtogen.GenerateCreateArticleRequestDTO(), authorToken)

		// Create reader
		_, readerToken := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())

		// Get article as reader
		respBody := test.GetArticle(t, createdArticle.Slug, &readerToken)

		// Verify response
		assert.Equal(t, createdArticle.Slug, respBody.Slug)
		assert.False(t, respBody.Favorited)
		assert.False(t, respBody.Author.Following)
		assert.Equal(t, 0, respBody.FavoritesCount)
	})
}

func TestGetArticleFollowingAndFavorited(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create author and article
		author, authorToken := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())
		createdArticle := test.CreateArticle(t, dtogen.GenerateCreateArticleRequestDTO(), authorToken)

		// Create reader
		_, readerToken := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())

		// Follow the author
		test.FollowUser(t, author.Username, readerToken)
		// Favorite the article
		test.FavoriteArticle(t, createdArticle.Slug, readerToken)

		// Get article as reader
		respBody := test.GetArticle(t, createdArticle.Slug, &readerToken)

		// Verify response
		assert.Equal(t, createdArticle.Slug, respBody.Slug)
		assert.True(t, respBody.Favorited)
		assert.True(t, respBody.Author.Following)
		assert.Equal(t, 1, respBody.FavoritesCount)
	})
}

func TestGetOwnArticleNotFavorited(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create author and article
		_, token := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())
		createdArticle := test.CreateArticle(t, dtogen.GenerateCreateArticleRequestDTO(), token)

		// Get own article
		respBody := test.GetArticle(t, createdArticle.Slug, &token)

		// Verify response
		assert.Equal(t, createdArticle.Slug, respBody.Slug)
		assert.False(t, respBody.Favorited)
		assert.False(t, respBody.Author.Following) // You don't follow yourself
		assert.Equal(t, 0, respBody.FavoritesCount)
	})
}

func TestGetOwnArticleFavorited(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create author and article
		_, token := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())
		createdArticle := test.CreateArticle(t, dtogen.GenerateCreateArticleRequestDTO(), token)

		// Favorite own article
		test.FavoriteArticle(t, createdArticle.Slug, token)
		// Get own article
		respBody := test.GetArticle(t, createdArticle.Slug, &token)

		// Verify response
		assert.Equal(t, createdArticle.Slug, respBody.Slug)
		assert.True(t, respBody.Favorited)
		assert.False(t, respBody.Author.Following) // You don't follow yourself
		assert.Equal(t, 1, respBody.FavoritesCount)
	})
}

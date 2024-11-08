package main

import (
	"fmt"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"realworld-aws-lambda-dynamodb-golang/internal/test"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetUserFeed(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create viewer user
		_, viewerToken := test.CreateAndLoginUser(t, dto.NewUserRequestUserDto{
			Username: "viewer",
			Email:    "viewer@test.com",
			Password: "password123",
		})

		// Create first followed user
		firstAuthor, firstAuthorToken := test.CreateAndLoginUser(t, dto.NewUserRequestUserDto{
			Username: "first-author",
			Email:    "first.author@test.com",
			Password: "password123",
		})

		// Create second followed user
		secondAuthor, secondAuthorToken := test.CreateAndLoginUser(t, dto.NewUserRequestUserDto{
			Username: "second-author",
			Email:    "second.author@test.com",
			Password: "password123",
		})

		// Viewer follows both authors
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "POST", fmt.Sprintf("/api/profiles/%s/follow", firstAuthor.Username), http.StatusOK, nil, viewerToken)
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "POST", fmt.Sprintf("/api/profiles/%s/follow", secondAuthor.Username), http.StatusOK, nil, viewerToken)

		// Create first article (older)
		firstArticle := test.CreateArticleEntity(t, dto.CreateArticleRequestDTO{
			Title:       "First Article",
			Description: "First Description",
			Body:        "First Body",
			TagList:     []string{"first"},
		}, firstAuthorToken)

		// Wait a bit to ensure different creation times
		time.Sleep(1 * time.Second)

		// Create second article (newer)
		secondArticle := test.CreateArticleEntity(t, dto.CreateArticleRequestDTO{
			Title:       "Second Article",
			Description: "Second Description",
			Body:        "Second Body",
			TagList:     []string{"second"},
		}, secondAuthorToken)

		// Favorite the first article
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "POST", fmt.Sprintf("/api/articles/%s/favorite", firstArticle.Slug), http.StatusOK, nil, viewerToken)

		// Check feed with retries
		assert.EventuallyWithT(t, func(testingT *assert.CollectT) {

			var feedResp dto.MultipleArticlesResponseBodyDTO
			test.MakeAuthenticatedRequestAndParseResponse(t, nil, "GET", "/api/articles/feed", http.StatusOK, &feedResp, viewerToken)

			// Check if we have both articles
			assert.Equal(testingT, 2, len(feedResp.Articles))

			// Check ordering (newest first)
			firstArticleFromFeed := feedResp.Articles[0]
			assert.Equal(testingT, firstArticle.Slug, firstArticleFromFeed.Slug)
			assert.True(testingT, firstArticleFromFeed.Favorited)
			assert.Equal(testingT, 1, firstArticleFromFeed.FavoritesCount)

			secondArticleFromFeed := feedResp.Articles[1]
			assert.Equal(testingT, secondArticle.Slug, secondArticleFromFeed.Slug)
			assert.False(testingT, secondArticleFromFeed.Favorited)
			assert.Equal(testingT, 0, secondArticleFromFeed.FavoritesCount)

		}, 10*time.Second, 1*time.Second, "feed should contain expected articles in correct order")
		return
	})
}

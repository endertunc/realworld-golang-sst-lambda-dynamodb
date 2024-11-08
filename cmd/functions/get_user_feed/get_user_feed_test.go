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
		time.Sleep(100 * time.Millisecond)

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

		}, 5*time.Second, 1*time.Second, "feed should contain expected articles in correct order")
	})
}

func TestGetUserFeedPagination(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create viewer user
		_, viewerToken := test.CreateAndLoginUser(t, dto.NewUserRequestUserDto{
			Username: "viewer",
			Email:    "viewer@test.com",
			Password: "password123",
		})

		// Create first author
		firstAuthor, firstAuthorToken := test.CreateAndLoginUser(t, dto.NewUserRequestUserDto{
			Username: "first-author",
			Email:    "first.author@test.com",
			Password: "password123",
		})

		// Create second author
		secondAuthor, secondAuthorToken := test.CreateAndLoginUser(t, dto.NewUserRequestUserDto{
			Username: "second-author",
			Email:    "second.author@test.com",
			Password: "password123",
		})

		// Viewer follows both authors
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "POST", fmt.Sprintf("/api/profiles/%s/follow", firstAuthor.Username), http.StatusOK, nil, viewerToken)
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "POST", fmt.Sprintf("/api/profiles/%s/follow", secondAuthor.Username), http.StatusOK, nil, viewerToken)

		// First author creates two articles
		firstAuthorArticle1 := test.CreateArticleEntity(t, dto.CreateArticleRequestDTO{
			Title:       "First Author Article 1",
			Description: "First Author Description 1",
			Body:        "First Author Body 1",
			TagList:     []string{"first-1"},
		}, firstAuthorToken)
		firstAuthorArticle2 := test.CreateArticleEntity(t, dto.CreateArticleRequestDTO{
			Title:       "First Author Article 2",
			Description: "First Author Description 2",
			Body:        "First Author Body 2",
			TagList:     []string{"first-2"},
		}, firstAuthorToken)
		// Second author creates two articles
		secondAuthorArticle1 := test.CreateArticleEntity(t, dto.CreateArticleRequestDTO{
			Title:       "Second Author Article 1",
			Description: "Second Author Description 1",
			Body:        "Second Author Body 1",
			TagList:     []string{"second-1"},
		}, secondAuthorToken)
		secondAuthorArticle2 := test.CreateArticleEntity(t, dto.CreateArticleRequestDTO{
			Title:       "Second Author Article 2",
			Description: "Second Author Description 2",
			Body:        "Second Author Body 2",
			TagList:     []string{"second-2"},
		}, secondAuthorToken)

		// Check first page with retries (limit=3)
		assert.EventuallyWithT(t, func(testingT *assert.CollectT) {
			var firstPageResp dto.MultipleArticlesResponseBodyDTO
			test.MakeAuthenticatedRequestAndParseResponse(t, nil, "GET", "/api/articles/feed?limit=3", http.StatusOK, &firstPageResp, viewerToken)

			//assert.Equal(testingT, 3, len(firstPageResp.Articles))
			assert.Len(testingT, firstPageResp.Articles, 3)
			assert.NotNil(testingT, firstPageResp.NextPageToken)

			// Check ordering (newest first)
			assert.Equal(testingT, secondAuthorArticle2.Slug, firstPageResp.Articles[0].Slug)
			assert.Equal(testingT, secondAuthorArticle1.Slug, firstPageResp.Articles[1].Slug)
			assert.Equal(testingT, firstAuthorArticle2.Slug, firstPageResp.Articles[2].Slug)

			// Get second page using nextPageToken
			var secondPageResp dto.MultipleArticlesResponseBodyDTO
			test.MakeAuthenticatedRequestAndParseResponse(t, nil, "GET", fmt.Sprintf("/api/articles/feed?limit=3&offset=%s", *firstPageResp.NextPageToken), http.StatusOK, &secondPageResp, viewerToken)

			// Check the second page
			//require.Equal(t, 1, len(secondPageResp.Articles))
			assert.Len(testingT, secondPageResp.Articles, 1)

			assert.Nil(t, secondPageResp.NextPageToken)
			assert.Equal(t, firstAuthorArticle1.Slug, secondPageResp.Articles[0].Slug)

		}, 5*time.Second, 1*time.Second, "first page should contain 3 newest articles")

	})
}

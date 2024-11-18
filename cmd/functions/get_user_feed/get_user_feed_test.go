package main

import (
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	dtogen "realworld-aws-lambda-dynamodb-golang/internal/domain/dto/generator"
	"realworld-aws-lambda-dynamodb-golang/internal/test"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetUserFeed(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create viewer user
		_, viewerToken := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())

		// Create first followed user
		firstAuthor, firstAuthorToken := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())

		// Create second followed user
		secondAuthor, secondAuthorToken := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())

		// Viewer follows both authors
		test.FollowUser(t, firstAuthor.Username, viewerToken)
		test.FollowUser(t, secondAuthor.Username, viewerToken)

		// Create first article (older)
		firstArticle := test.CreateArticle(t, dtogen.GenerateCreateArticleRequestDTO(), firstAuthorToken)

		// Wait a bit to ensure different creation times
		time.Sleep(100 * time.Millisecond)

		// Create second article (newer)
		secondArticle := test.CreateArticle(t, dto.CreateArticleRequestDTO{
			Title:       "Second Article",
			Description: "Second Description",
			Body:        "Second Body",
			TagList:     []string{"second"},
		}, secondAuthorToken)

		// Favorite the first article
		test.FavoriteArticle(t, firstArticle.Slug, viewerToken)

		// Check feed with retries
		assert.EventuallyWithT(t, func(testingT *assert.CollectT) {
			feedResp := test.GetUserFeedWithPagination(t, viewerToken, 20, nil)

			// Check if we have both articles
			assert.Equal(testingT, 2, len(feedResp.Articles))

			// Check ordering (newest first)
			firstArticleFromFeed := feedResp.Articles[0]
			assert.Equal(testingT, secondArticle.Slug, firstArticleFromFeed.Slug)
			assert.False(testingT, firstArticleFromFeed.Favorited)
			assert.Equal(testingT, 0, firstArticleFromFeed.FavoritesCount)

			secondArticleFromFeed := feedResp.Articles[1]
			assert.Equal(testingT, firstArticle.Slug, secondArticleFromFeed.Slug)
			assert.True(testingT, secondArticleFromFeed.Favorited)
			assert.Equal(testingT, 1, secondArticleFromFeed.FavoritesCount)

		}, 5*time.Second, 1*time.Second, "feed should contain expected articles in correct order")
	})
}

func TestGetUserFeedPagination(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create viewer user
		_, viewerToken := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())

		// Create first author
		firstAuthor, firstAuthorToken := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())

		// Create second author
		secondAuthor, secondAuthorToken := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())

		// Viewer follows both authors
		test.FollowUser(t, firstAuthor.Username, viewerToken)
		test.FollowUser(t, secondAuthor.Username, viewerToken)

		// First author creates two articles
		firstAuthorArticle1 := test.CreateArticle(t, dtogen.GenerateCreateArticleRequestDTO(), firstAuthorToken)
		firstAuthorArticle2 := test.CreateArticle(t, dtogen.GenerateCreateArticleRequestDTO(), firstAuthorToken)
		// Second author creates two articles
		secondAuthorArticle1 := test.CreateArticle(t, dtogen.GenerateCreateArticleRequestDTO(), secondAuthorToken)
		secondAuthorArticle2 := test.CreateArticle(t, dtogen.GenerateCreateArticleRequestDTO(), secondAuthorToken)

		// Check first page with retries (limit=3)
		assert.EventuallyWithT(t, func(testingT *assert.CollectT) {
			firstPageResp := test.GetUserFeedWithPagination(t, viewerToken, 3, nil)

			//assert.Equal(testingT, 3, len(firstPageResp.Articles))
			assert.Len(testingT, firstPageResp.Articles, 3)
			assert.NotNil(testingT, firstPageResp.NextPageToken)

			// Check ordering (newest first)
			assert.Equal(testingT, secondAuthorArticle2.Slug, firstPageResp.Articles[0].Slug)
			assert.Equal(testingT, secondAuthorArticle1.Slug, firstPageResp.Articles[1].Slug)
			assert.Equal(testingT, firstAuthorArticle2.Slug, firstPageResp.Articles[2].Slug)

			// Get second page using nextPageToken
			secondPageResp := test.GetUserFeedWithPagination(t, viewerToken, 3, firstPageResp.NextPageToken)

			// Check the second page
			//require.Equal(t, 1, len(secondPageResp.Articles))
			assert.Len(testingT, secondPageResp.Articles, 1)

			assert.Nil(t, secondPageResp.NextPageToken)
			assert.Equal(t, firstAuthorArticle1.Slug, secondPageResp.Articles[0].Slug)

		}, 5*time.Second, 2*time.Second, "first page should contain 3 newest articles")

	})
}

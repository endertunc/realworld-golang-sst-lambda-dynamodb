package main

import (
	"github.com/stretchr/testify/assert"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto/generator"
	"realworld-aws-lambda-dynamodb-golang/internal/test"
	"testing"
	"time"
)

// Listing articles globally by most recent involves following steps:
//   - articles are created in dynamodb
//   - articles are ingested to opensearch via opensearch-ingest-pipeline automatically
//   - invoke list articles endpoint which fetches articles from opensearch
//
// As a result, each test case takes a long time to run because of the time it takes to ingest articles into opensearch.
// Therefore, we implement minimal test cases to cover the main functionality of the endpoint.
// The rest of the test cases (such as checking the correct favorite and following flags) are implemented as unit tests in the service layer.
func TestListArticlesByMostRecentWithoutAuth(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// create viewer user
		viewerUser := generator.GenerateNewUserRequestUserDto()
		test.CreateAndLoginUser(t, viewerUser)

		// create first author user
		author1User := generator.GenerateNewUserRequestUserDto()
		_, author1Token := test.CreateAndLoginUser(t, author1User)

		// create second author user
		author2User := generator.GenerateNewUserRequestUserDto()
		_, author2Token := test.CreateAndLoginUser(t, author2User)

		// create first articles for both authors
		article1 := generator.GenerateCreateArticleRequestDTO()
		createdArticle1 := test.CreateArticle(t, article1, author1Token)

		article2 := generator.GenerateCreateArticleRequestDTO()
		createdArticle2 := test.CreateArticle(t, article2, author2Token)

		// create second articles for both authors
		article3 := generator.GenerateCreateArticleRequestDTO()
		createdArticle3 := test.CreateArticle(t, article3, author1Token)

		article4 := generator.GenerateCreateArticleRequestDTO()
		createdArticle4 := test.CreateArticle(t, article4, author2Token)

		expectedArticlesInOrder := []dto.ArticleResponseDTO{createdArticle4, createdArticle3, createdArticle2, createdArticle1}

		assert.EventuallyWithT(t, func(testingT *assert.CollectT) {
			// test listing articles by author1 without auth
			listResponseNoAuth := test.ListArticles(t, nil, test.ArticleQueryParams{})
			assert.Equal(testingT, 4, listResponseNoAuth.ArticlesCount)
			assert.Equal(testingT, 4, len(listResponseNoAuth.Articles))
			assert.Equal(testingT, expectedArticlesInOrder, listResponseNoAuth.Articles)
			//assert.Condition(testingT, func() bool {
			//	if len(listResponseNoAuth.Articles) > 1 {
			//		article := listResponseNoAuth.Articles[1]
			//		return assert.Equal(testingT, createdArticle3, article)
			//	}
			//	return false
			//})
			//assert.Condition(testingT, func() bool {
			//	if len(listResponseNoAuth.Articles) > 2 {
			//		article := listResponseNoAuth.Articles[2]
			//		return assert.Equal(testingT, createdArticle2, article)
			//	}
			//	return false
			//})
		}, 60*time.Second, 5*time.Second) // see note above about long test duration

	})
}

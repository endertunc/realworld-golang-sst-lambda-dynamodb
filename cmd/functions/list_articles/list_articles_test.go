package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"realworld-aws-lambda-dynamodb-golang/internal/test"
	"testing"
)

func TestListArticlesByAuthor(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// create viewer user
		viewerUser := dto.NewUserRequestUserDto{
			Email:    "viewer@test.com",
			Username: "viewer-user",
			Password: "password123",
		}
		_, viewerToken := test.CreateAndLoginUser(t, viewerUser)

		// create first author user
		firstAuthor := dto.NewUserRequestUserDto{
			Email:    "first-author@test.com",
			Username: "first.author",
			Password: "password123",
		}
		_, author1Token := test.CreateAndLoginUser(t, firstAuthor)

		// create second author user
		secondAuthor := dto.NewUserRequestUserDto{
			Email:    "second-author@test.com",
			Username: "second.author",
			Password: "password123",
		}
		_, author2Token := test.CreateAndLoginUser(t, secondAuthor)

		// create article for firstAuthor
		article1 := dto.CreateArticleRequestDTO{
			Title:       "Test Article 1",
			Description: "This is test article 1",
			Body:        "Body of test article 1",
			TagList:     []string{"test"},
		}
		test.CreateArticleEntity(t, article1, author1Token)

		// create 2 articles for secondAuthor
		article2 := dto.CreateArticleRequestDTO{
			Title:       "Test Article 2",
			Description: "This is test article 2",
			Body:        "Body of test article 2",
			TagList:     []string{"test"},
		}
		createdArticle2 := test.CreateArticleEntity(t, article2, author2Token)

		article3 := dto.CreateArticleRequestDTO{
			Title:       "Test Article 3",
			Description: "This is test article 3",
			Body:        "Body of test article 3",
			TagList:     []string{"test"},
		}
		createdArticle3 := test.CreateArticleEntity(t, article3, author2Token)

		// viewer follows secondAuthor
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "POST",
			fmt.Sprintf("/api/profiles/%s/follow", secondAuthor.Username),
			http.StatusOK, nil, viewerToken)

		// viewer favorites one of secondAuthor's articles
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "POST",
			fmt.Sprintf("/api/articles/%s/favorite", createdArticle2.Slug),
			http.StatusOK, nil, viewerToken)

		// list articles by secondAuthor
		var listResponse dto.MultipleArticlesResponseBodyDTO
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "GET",
			fmt.Sprintf("/api/articles?author=%s", secondAuthor.Username),
			http.StatusOK, &listResponse, viewerToken)

		// verify response
		assert.Equal(t, 2, len(listResponse.Articles))
		assert.Equal(t, createdArticle3.Slug, listResponse.Articles[0].Slug) // most recent first
		assert.Equal(t, createdArticle2.Slug, listResponse.Articles[1].Slug)
		assert.True(t, listResponse.Articles[1].Favorited)  // article2 is favorited
		assert.False(t, listResponse.Articles[0].Favorited) // article3 is not favorited
		assert.Equal(t, 1, listResponse.Articles[1].FavoritesCount)
		assert.Equal(t, 0, listResponse.Articles[0].FavoritesCount)
	})
}

func TestAuthenticationScenarios(t *testing.T) {
	test.RunAuthenticationTests(t, test.SharedAuthenticationTestConfig{
		Method: "GET",
		Path:   "/api/articles",
	})
}

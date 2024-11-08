package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"realworld-aws-lambda-dynamodb-golang/internal/test"
	"testing"
)

func TestAuthenticationScenarios(t *testing.T) {
	test.RunAuthenticationTests(t, test.SharedAuthenticationTestConfig{
		Method: "POST",
		Path:   "/api/articles",
	})
}

func TestSuccessfulArticleCreation(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create and login a user
		user, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)

		// Create an article
		article := dto.CreateArticleRequestDTO{
			Title:       "How to train your dragon",
			Description: "Ever wonder how?",
			Body:        "You have to believe",
			TagList:     []string{"dragons", "training"},
		}

		reqBody := dto.CreateArticleRequestBodyDTO{
			Article: article,
		}

		var respBody dto.ArticleResponseBodyDTO
		test.MakeAuthenticatedRequestAndParseResponse(t, reqBody, "POST", "/api/articles", http.StatusOK, &respBody, token)

		// Create expected article response
		expectedArticle := dto.ArticleResponseDTO{
			Title:          article.Title,
			Description:    article.Description,
			Body:           article.Body,
			TagList:        article.TagList,
			Favorited:      false,
			FavoritesCount: 0,
			Author: dto.AuthorDTO{
				Username:  user.Username,
				Bio:       nil,
				Image:     nil,
				Following: false,
			},
			// dynamic fields
			Slug:      respBody.Article.Slug,
			CreatedAt: respBody.Article.CreatedAt,
			UpdatedAt: respBody.Article.UpdatedAt,
		}

		assert.Equal(t, expectedArticle, respBody.Article)
		assert.NotEmpty(t, respBody.Article.Slug)
		assert.NotZero(t, respBody.Article.CreatedAt)
		assert.NotZero(t, respBody.Article.UpdatedAt)
	})
}

func TestCreateArticlesWithSameTitle(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create and login a user
		user, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)

		// Create first article
		article := dto.CreateArticleRequestDTO{
			Title:       "How to train your dragon",
			Description: "Ever wonder how?",
			Body:        "You have to believe",
			TagList:     []string{"dragons", "training"},
		}

		reqBody := dto.CreateArticleRequestBodyDTO{
			Article: article,
		}

		var firstRespBody dto.ArticleResponseBodyDTO
		test.MakeAuthenticatedRequestAndParseResponse(t, reqBody, "POST", "/api/articles", http.StatusCreated, &firstRespBody, token)

		// Create second article with the same title
		var secondRespBody dto.ArticleResponseBodyDTO
		test.MakeAuthenticatedRequestAndParseResponse(t, reqBody, "POST", "/api/articles", http.StatusCreated, &secondRespBody, token)

		// Verify both articles were created successfully
		assert.Equal(t, article.Title, firstRespBody.Article.Title)
		assert.Equal(t, article.Title, secondRespBody.Article.Title)

		// Verify they have different slugs
		assert.NotEqual(t, firstRespBody.Article.Slug, secondRespBody.Article.Slug)
		assert.NotEmpty(t, firstRespBody.Article.Slug)
		assert.NotEmpty(t, secondRespBody.Article.Slug)

		// Verify other fields are set correctly for both articles
		baseExpectedArticle := dto.ArticleResponseDTO{
			Title:          article.Title,
			Description:    article.Description,
			Body:           article.Body,
			TagList:        article.TagList,
			Favorited:      false,
			FavoritesCount: 0,
			Author: dto.AuthorDTO{
				Username:  user.Username,
				Bio:       nil,
				Image:     nil,
				Following: false,
			},
		}

		// Compare non-dynamic fields for first article
		firstExpected := baseExpectedArticle
		firstExpected.Slug = firstRespBody.Article.Slug
		firstExpected.CreatedAt = firstRespBody.Article.CreatedAt
		firstExpected.UpdatedAt = firstRespBody.Article.UpdatedAt
		assert.Equal(t, firstExpected, firstRespBody.Article)

		// Compare non-dynamic fields for second article
		secondExpected := baseExpectedArticle
		secondExpected.Slug = secondRespBody.Article.Slug
		secondExpected.CreatedAt = secondRespBody.Article.CreatedAt
		secondExpected.UpdatedAt = secondRespBody.Article.UpdatedAt
		assert.Equal(t, secondExpected, secondRespBody.Article)
	})
}

func TestCreateArticleWithoutTags(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create and login a user
		user, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)

		// Create an article without tags
		article := dto.CreateArticleRequestDTO{
			Title:       "How to train your dragon",
			Description: "Ever wonder how?",
			Body:        "You have to believe",
		}

		reqBody := dto.CreateArticleRequestBodyDTO{
			Article: article,
		}

		var respBody dto.ArticleResponseBodyDTO
		test.MakeAuthenticatedRequestAndParseResponse(t, reqBody, "POST", "/api/articles", http.StatusOK, &respBody, token)

		// Create expected article response
		expectedArticle := dto.ArticleResponseDTO{
			Title:          article.Title,
			Description:    article.Description,
			Body:           article.Body,
			TagList:        nil,
			Favorited:      false,
			FavoritesCount: 0,
			Author: dto.AuthorDTO{
				Username:  user.Username,
				Bio:       nil,
				Image:     nil,
				Following: false,
			},
			Slug:      respBody.Article.Slug,
			CreatedAt: respBody.Article.CreatedAt,
			UpdatedAt: respBody.Article.UpdatedAt,
		}

		assert.Equal(t, expectedArticle, respBody.Article)
	})
}

func TestCreateArticleWithEmptyTags(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create and login a user
		user, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)

		// Create an article with empty tags array
		article := dto.CreateArticleRequestDTO{
			Title:       "How to train your dragon",
			Description: "Ever wonder how?",
			Body:        "You have to believe",
			TagList:     []string{},
		}

		reqBody := dto.CreateArticleRequestBodyDTO{
			Article: article,
		}

		var respBody dto.ArticleResponseBodyDTO
		test.MakeAuthenticatedRequestAndParseResponse(t, reqBody, "POST", "/api/articles", http.StatusCreated, &respBody, token)

		// Create expected article response
		expectedArticle := dto.ArticleResponseDTO{
			Title:          article.Title,
			Description:    article.Description,
			Body:           article.Body,
			TagList:        article.TagList,
			Favorited:      false,
			FavoritesCount: 0,
			Author: dto.AuthorDTO{
				Username:  user.Username,
				Bio:       nil,
				Image:     nil,
				Following: false,
			},
			Slug:      respBody.Article.Slug,
			CreatedAt: respBody.Article.CreatedAt,
			UpdatedAt: respBody.Article.UpdatedAt,
		}

		assert.Equal(t, expectedArticle, respBody.Article)
	})
}

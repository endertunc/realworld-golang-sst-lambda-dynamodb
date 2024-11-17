package main

import (
	"github.com/stretchr/testify/assert"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	dtogen "realworld-aws-lambda-dynamodb-golang/internal/domain/dto/generator"
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
		article := dtogen.GenerateCreateArticleRequestDTO()
		respBody := test.CreateArticle(t, article, token)

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
			Slug:      respBody.Slug,
			CreatedAt: respBody.CreatedAt,
			UpdatedAt: respBody.UpdatedAt,
		}

		assert.Equal(t, expectedArticle, respBody)
		assert.NotEmpty(t, respBody.Slug)
		assert.NotZero(t, respBody.CreatedAt)
		assert.NotZero(t, respBody.UpdatedAt)
	})
}

func TestCreateArticlesWithSameTitle(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create and login a user
		user, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)

		// Create first article
		article := dtogen.GenerateCreateArticleRequestDTO()

		firstRespBody := test.CreateArticle(t, article, token)

		// Create a second article with the same title
		secondRespBody := test.CreateArticle(t, article, token)

		// Verify both articles were created successfully
		assert.Equal(t, article.Title, firstRespBody.Title)
		assert.Equal(t, article.Title, secondRespBody.Title)

		// Verify they have different slugs
		assert.NotEqual(t, firstRespBody.Slug, secondRespBody.Slug)
		assert.NotEmpty(t, firstRespBody.Slug)
		assert.NotEmpty(t, secondRespBody.Slug)

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
		firstExpected.Slug = firstRespBody.Slug
		firstExpected.CreatedAt = firstRespBody.CreatedAt
		firstExpected.UpdatedAt = firstRespBody.UpdatedAt
		assert.Equal(t, firstExpected, firstRespBody)

		// Compare non-dynamic fields for second article
		secondExpected := baseExpectedArticle
		secondExpected.Slug = secondRespBody.Slug
		secondExpected.CreatedAt = secondRespBody.CreatedAt
		secondExpected.UpdatedAt = secondRespBody.UpdatedAt
		assert.Equal(t, secondExpected, secondRespBody)
	})
}

func TestCreateArticleWithoutTags(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create and login a user
		user, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)

		// Create an article without tags
		article := dtogen.GenerateCreateArticleRequestDTO()
		article.TagList = nil

		respBody := test.CreateArticle(t, article, token)

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
			Slug:      respBody.Slug,
			CreatedAt: respBody.CreatedAt,
			UpdatedAt: respBody.UpdatedAt,
		}

		assert.Equal(t, expectedArticle, respBody)
	})
}

func TestCreateArticleWithEmptyTags(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create and login a user
		user, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)

		// Create an article with empty tags array
		article := dtogen.GenerateCreateArticleRequestDTO()
		article.TagList = []string{}

		respBody := test.CreateArticle(t, article, token)

		// Create expected article response
		expectedArticle := dto.ArticleResponseDTO{
			Title:          article.Title,
			Description:    article.Description,
			Body:           article.Body,
			TagList:        []string{},
			Favorited:      false,
			FavoritesCount: 0,
			Author: dto.AuthorDTO{
				Username:  user.Username,
				Bio:       nil,
				Image:     nil,
				Following: false,
			},
			Slug:      respBody.Slug,
			CreatedAt: respBody.CreatedAt,
			UpdatedAt: respBody.UpdatedAt,
		}

		assert.Equal(t, expectedArticle, respBody)
	})
}

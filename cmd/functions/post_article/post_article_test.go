package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	dtogen "realworld-aws-lambda-dynamodb-golang/internal/domain/dto/generator"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/test"
	"strings"
	"testing"
)

func TestAuthenticationScenarios(t *testing.T) {
	test.RunAuthenticationTests(t, test.SharedAuthenticationTestConfig{
		Method: "POST",
		Path:   "/api/articles",
	})
}

//nolint:golint,exhaustruct
func TestRequestValidation(t *testing.T) {
	// create a user
	_, token := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())

	tests := []test.ApiRequestValidationTest[dto.CreateArticleRequestDTO]{
		{
			Name: "missing title",
			Input: dto.CreateArticleRequestDTO{
				Description: "This is a test article",
				Body:        "Article body content",
				TagList:     []string{"test", "article"},
			},
			ExpectedError: map[string]string{
				"Article.Title": "Title is a required field",
			},
		},
		{
			Name: "blank title",
			Input: dto.CreateArticleRequestDTO{
				Title:       "    ",
				Description: "This is a test article",
				Body:        "Article body content",
				TagList:     []string{"test", "article"},
			},
			ExpectedError: map[string]string{
				"Article.Title": "Title cannot be blank",
			},
		},
		{
			Name: "title too long",
			Input: dto.CreateArticleRequestDTO{
				Title:       strings.Repeat("a", 256),
				Description: "This is a test article",
				Body:        "Article body content",
				TagList:     []string{"test", "article"},
			},
			ExpectedError: map[string]string{
				"Article.Title": "Title must be a maximum of 255 characters in length",
			},
		},
		{
			Name: "missing description",
			Input: dto.CreateArticleRequestDTO{
				Title:   "Test Article",
				Body:    "Article body content",
				TagList: []string{"test", "article"},
			},
			ExpectedError: map[string]string{
				"Article.Description": "Description is a required field",
			},
		},
		{
			Name: "blank description",
			Input: dto.CreateArticleRequestDTO{
				Title:       "Test Article",
				Description: "     ",
				Body:        "Article body content",
				TagList:     []string{"test", "article"},
			},
			ExpectedError: map[string]string{
				"Article.Description": "Description cannot be blank",
			},
		},
		{
			Name: "description too long",
			Input: dto.CreateArticleRequestDTO{
				Title:       "Test Article",
				Description: strings.Repeat("a", 1025),
				Body:        "Article body content",
				TagList:     []string{"test", "article"},
			},
			ExpectedError: map[string]string{
				"Article.Description": "Description must be a maximum of 1,024 characters in length",
			},
		},
		{
			Name: "missing body",
			Input: dto.CreateArticleRequestDTO{
				Title:       "Test Article",
				Description: "This is a test article",
				TagList:     []string{"test", "article"},
			},
			ExpectedError: map[string]string{
				"Article.Body": "Body is a required field",
			},
		},
		{
			Name: "blank body",
			Input: dto.CreateArticleRequestDTO{
				Title:       "Test Article",
				Description: "This is a test article",
				Body:        "     ",
				TagList:     []string{"test", "article"},
			},
			ExpectedError: map[string]string{
				"Article.Body": "Body cannot be blank",
			},
		},
		{
			Name: "empty tag list",
			Input: dto.CreateArticleRequestDTO{
				Title:       "Test Article",
				Description: "This is a test article",
				Body:        "Article body content",
				TagList:     []string{},
			},
			ExpectedError: map[string]string{
				"Article.TagList": "TagList must contain more than 0 items",
			},
		},
		{
			Name: "duplicate tags",
			Input: dto.CreateArticleRequestDTO{
				Title:       "Test Article",
				Description: "This is a test article",
				Body:        "Article body content",
				TagList:     []string{"test", "test"},
			},
			ExpectedError: map[string]string{
				"Article.TagList": "TagList must contain unique values",
			},
		},
		{
			Name: "blank tag",
			Input: dto.CreateArticleRequestDTO{
				Title:       "Test Article",
				Description: "This is a test article",
				Body:        "Article body content",
				TagList:     []string{"test", "   "},
			},
			ExpectedError: map[string]string{
				"Article.TagList[1]": "TagList[1] cannot be blank",
			},
		},
		{
			Name: "tag too long",
			Input: dto.CreateArticleRequestDTO{
				Title:       "Test Article",
				Description: "This is a test article",
				Body:        "Article body content",
				TagList:     []string{"test", strings.Repeat("a", 65)},
			},
			ExpectedError: map[string]string{
				"Article.TagList[1]": "TagList[1] must be a maximum of 64 characters in length",
			},
		},
	}

	createArticleRequest := func(t *testing.T, input dto.CreateArticleRequestDTO) errutil.ValidationErrors {
		return test.CreateArticleWithResponse[errutil.ValidationErrors](t, input, token, http.StatusBadRequest)
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			test.TestValidation(t, tt, createArticleRequest)
		})
	}
}

func TestSuccessfulArticleCreation(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create and login a user
		user, token := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())

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
		user, token := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())

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
		baseExpectedArticle := dto.ArticleResponseDTO{ //nolint:golint,exhaustruct
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
		user, token := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())

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
		user, token := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())

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

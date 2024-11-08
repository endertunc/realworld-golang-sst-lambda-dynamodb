package test

import (
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"testing"
)

var DefaultCreateArticleRequestDTO = dto.CreateArticleRequestDTO{
	Title:       "Test Article",
	Description: "This is a test article",
	Body:        "This is the body of the test article",
	TagList:     []string{"tag-one", "tag-two"},
}

// ToDo @ender I have different file name pattern... this file camelCase, auth_test_suite.go, other with kebab-case....

// CreateArticleEntity creates an article for testing purposes
func CreateArticleEntity(t *testing.T, article dto.CreateArticleRequestDTO, token string) dto.ArticleResponseDTO {
	reqBody := dto.CreateArticleRequestBodyDTO{
		Article: article,
	}
	var respBody dto.ArticleResponseBodyDTO
	MakeAuthenticatedRequestAndParseResponse(t, reqBody, "POST", "/api/articles", http.StatusOK, &respBody, token)
	return respBody.Article
}

// CreateDefaultArticle creates an article with default test data
func CreateDefaultArticle(t *testing.T, token string) dto.ArticleResponseDTO {
	return CreateArticleEntity(t, DefaultCreateArticleRequestDTO, token)
}

package test

import (
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"testing"
)

// CreateArticle creates an article for testing purposes
//func CreateArticle(t *testing.T, article dto.CreateArticleRequestDTO, token string) dto.ArticleResponseDTO {
//	reqBody := dto.CreateArticleRequestBodyDTO{
//		Article: article,
//	}
//	var respBody dto.ArticleResponseBodyDTO
//	MakeAuthenticatedRequestAndParseResponse(t, reqBody, "POST", "/api/articles", http.StatusOK, &respBody, token)
//	return respBody.Article
//}

// CreateDefaultArticle creates an article with default test data
//func CreateDefaultArticle(t *testing.T, token string) dto.ArticleResponseDTO {
//	return CreateArticle(t, DefaultCreateArticleRequestDTO, token)
//}

func CreateArticle(t *testing.T, article dto.CreateArticleRequestDTO, token string) dto.ArticleResponseDTO {
	return CreateArticleWithResponse[dto.ArticleResponseBodyDTO](t, article, token, http.StatusOK).Article
}

func CreateArticleWithResponse[T interface{}](t *testing.T, article dto.CreateArticleRequestDTO, token string, expectedStatusCode int) T {
	var respBody T
	reqBody := dto.CreateArticleRequestBodyDTO{Article: article}
	MakeAuthenticatedRequestAndParseResponse(t, reqBody, "POST", "/api/articles", expectedStatusCode, &respBody, token)
	return respBody
}

func FavoriteArticle(t *testing.T, slug string, token string) dto.ArticleResponseDTO {
	return FavoriteArticleWithResponse[dto.ArticleResponseBodyDTO](t, slug, token, http.StatusOK).Article
}

func FavoriteArticleWithResponse[T interface{}](t *testing.T, slug string, token string, expectedStatusCode int) T {
	var respBody T
	MakeAuthenticatedRequestAndParseResponse(t, nil, "POST", "/api/articles/"+slug+"/favorite", expectedStatusCode, &respBody, token)
	return respBody
}

func UnfavoriteArticle(t *testing.T, slug string, token string) dto.ArticleResponseDTO {
	return UnfavoriteArticleWithResponse[dto.ArticleResponseBodyDTO](t, slug, token, http.StatusOK).Article
}

func UnfavoriteArticleWithResponse[T interface{}](t *testing.T, slug string, token string, expectedStatusCode int) T {
	var respBody T
	MakeAuthenticatedRequestAndParseResponse(t, nil, "DELETE", "/api/articles/"+slug+"/favorite", expectedStatusCode, &respBody, token)
	return respBody
}

func GetArticle(t *testing.T, slug string, token *string) dto.ArticleResponseDTO {
	return GetArticleWithResponse[dto.ArticleResponseBodyDTO](t, slug, token, http.StatusOK).Article
}

func GetArticleWithResponse[T interface{}](t *testing.T, slug string, token *string, expectedStatusCode int) T {
	var respBody T
	if token == nil {
		MakeRequestAndParseResponse(t, nil, "GET", "/api/articles/"+slug, expectedStatusCode, &respBody)
	} else {
		MakeAuthenticatedRequestAndParseResponse(t, nil, "GET", "/api/articles/"+slug, expectedStatusCode, &respBody, *token)
	}
	return respBody
}

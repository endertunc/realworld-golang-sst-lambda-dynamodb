package test

import (
	"fmt"
	"net/http"
	"net/url"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"strconv"
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

func CreateArticleWithResponse[T any](t *testing.T, article dto.CreateArticleRequestDTO, token string, expectedStatusCode int) T {
	reqBody := dto.CreateArticleRequestBodyDTO{Article: article}
	return ExecuteRequest[T](t, "POST", "/api/articles", reqBody, expectedStatusCode, &token)
}

func FavoriteArticle(t *testing.T, slug string, token string) dto.ArticleResponseDTO {
	return FavoriteArticleWithResponse[dto.ArticleResponseBodyDTO](t, slug, token, http.StatusOK).Article
}

func FavoriteArticleWithResponse[T interface{}](t *testing.T, slug string, token string, expectedStatusCode int) T {
	return ExecuteRequest[T](t, "POST", "/api/articles/"+slug+"/favorite", nil, expectedStatusCode, &token)
}

func UnfavoriteArticle(t *testing.T, slug string, token string) dto.ArticleResponseDTO {
	return UnfavoriteArticleWithResponse[dto.ArticleResponseBodyDTO](t, slug, token, http.StatusOK).Article
}

func UnfavoriteArticleWithResponse[T interface{}](t *testing.T, slug string, token string, expectedStatusCode int) T {
	return ExecuteRequest[T](t, "DELETE", "/api/articles/"+slug+"/favorite", nil, expectedStatusCode, &token)
}

func GetArticle(t *testing.T, slug string, token *string) dto.ArticleResponseDTO {
	return GetArticleWithResponse[dto.ArticleResponseBodyDTO](t, slug, token, http.StatusOK).Article
}

func GetArticleWithResponse[T interface{}](t *testing.T, slug string, token *string, expectedStatusCode int) T {
	return ExecuteRequest[T](t, "GET", "/api/articles/"+slug, nil, expectedStatusCode, token)
}

//func GetArticlesWithPagination(t *testing.T, token *string, limit int, offset *string) dto.MultipleArticlesResponseBodyDTO {
//	var respBody dto.MultipleArticlesResponseBodyDTO
//	path := fmt.Sprintf("/api/articles?limit=%d", limit)
//	if offset != nil {
//		path += fmt.Sprintf("%s&offset=%s", path, *offset)
//	}
//	if token == nil {
//		MakeRequestAndParseResponse(t, nil, "GET", path, http.StatusOK, &respBody)
//	} else {
//		MakeAuthenticatedRequestAndParseResponse(t, nil, "GET", "/api/articles/feed?limit=3", http.StatusOK, &respBody, *token)
//	}
//	return respBody
//}

type ArticleQueryParams struct {
	Limit     *int
	Offset    *string
	Author    *string
	Favorited *string
	Tag       *string
}

func (p ArticleQueryParams) ToQueryParams() string {
	query := url.Values{}
	if p.Limit != nil {
		query.Add("limit", strconv.Itoa(*p.Limit))
	}
	if p.Offset != nil {
		query.Add("offset", *p.Offset)
	}
	if p.Author != nil {
		query.Add("author", *p.Author)
	}
	if p.Favorited != nil {
		query.Add("favorited", *p.Favorited)
	}
	if p.Tag != nil {
		query.Add("tag", *p.Tag)
	}
	return query.Encode()
}

func ListArticles(t *testing.T, token *string, params ArticleQueryParams) dto.MultipleArticlesResponseBodyDTO {
	return ExecuteRequest[dto.MultipleArticlesResponseBodyDTO](t, "GET", "/api/articles?"+params.ToQueryParams(), nil, http.StatusOK, token)
}

func GetUserFeedWithPagination(t *testing.T, token string, limit int, offset *string) dto.MultipleArticlesResponseBodyDTO {
	path := fmt.Sprintf("/api/articles/feed?limit=%d", limit)
	if offset != nil {
		path = fmt.Sprintf("%s&offset=%s", path, *offset)
	}
	return ExecuteRequest[dto.MultipleArticlesResponseBodyDTO](t, "GET", path, nil, http.StatusOK, &token)
}

func GetTags(t *testing.T) dto.TagsResponseDTO {
	return ExecuteRequest[dto.TagsResponseDTO](t, "GET", "/api/tags", nil, http.StatusOK, nil)
}

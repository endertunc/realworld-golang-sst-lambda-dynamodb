package dto

import (
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"time"
)

// article request dtos
type CreateArticleRequestBodyDTO struct {
	Article CreateArticleRequestDTO `json:"article" validate:"required"`
}

type CreateArticleRequestDTO struct {
	Title       string   `json:"title" validate:"required,notblank,max=255"`
	Description string   `json:"description" validate:"required,notblank,max=1024"`
	Body        string   `json:"body" validate:"required,notblank"`
	TagList     []string `json:"tagList" validate:"gt=0,unique,dive,notblank,max=64"`
}

func (s CreateArticleRequestBodyDTO) Validate() (map[string]string, bool) {
	return validateStruct(s)
}

// article response dtos
type AuthorDTO struct {
	Username  string  `json:"username"`
	Bio       *string `json:"bio"`
	Image     *string `json:"image"`
	Following bool    `json:"following"`
}

type ArticleResponseBodyDTO struct {
	Article ArticleResponseDTO `json:"article"`
}

type ArticleResponseDTO struct {
	Slug           string    `json:"slug"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Body           string    `json:"body"`
	TagList        []string  `json:"tagList"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	Favorited      bool      `json:"favorited"`
	FavoritesCount int       `json:"favoritesCount"`
	Author         AuthorDTO `json:"author"`
}

type MultipleArticlesResponseBodyDTO struct {
	Articles      []ArticleResponseDTO `json:"article"`
	ArticlesCount int                  `json:"articlesCount"`
	NextPageToken *string              `json:"nextPageToken,omitempty"`
}

// factory methods
func ToArticleResponseDTO(article domain.Article, author domain.User, isFavorited, isFollowing bool) ArticleResponseDTO {
	return ArticleResponseDTO{
		Slug:           article.Slug,
		Title:          article.Title,
		Description:    article.Description,
		Body:           article.Body,
		TagList:        article.TagList,
		CreatedAt:      article.CreatedAt,
		UpdatedAt:      article.UpdatedAt,
		Favorited:      isFavorited,
		FavoritesCount: article.FavoritesCount,
		Author: AuthorDTO{
			Username:  author.Username,
			Bio:       author.Bio,
			Image:     author.Image,
			Following: isFollowing,
		},
	}
}

func ToArticleResponseBodyDTO(article domain.Article, author domain.User, isFavorited, isFollowing bool) ArticleResponseBodyDTO {
	return ArticleResponseBodyDTO{Article: ToArticleResponseDTO(article, author, isFavorited, isFollowing)}
}

func ToMultipleArticlesResponseBodyDTO(feedItems []domain.ArticleAggregateView, nextPageToken *string) MultipleArticlesResponseBodyDTO {
	articles := make([]ArticleResponseDTO, 0, len(feedItems))
	for _, feedItem := range feedItems {
		articleResponseDTO := ToArticleResponseDTO(feedItem.Article, feedItem.Author, feedItem.IsFavorited, feedItem.IsFollowing)
		articles = append(articles, articleResponseDTO)
	}
	return MultipleArticlesResponseBodyDTO{
		Articles:      articles,
		ArticlesCount: len(articles),
		NextPageToken: nextPageToken,
	}
}

type TagsResponseDTO struct {
	Tags []string `json:"tags"`
}

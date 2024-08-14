package dto

import (
	"github.com/google/uuid"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"time"
)

type CreateArticleRequestBodyDTO struct {
	Article CreateArticleRequestDTO `json:"comment"`
}

type CreateArticleRequestDTO struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Body        string `json:"body"`
	// ToDo @ender should this be *[]string
	// ToDo @ender do we want to differentiate null vs empty array?
	TagList []string `json:"tagList"`
}

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

func ToArticleResponseBodyDTO(article domain.Article, author domain.User, isFavorited, isFollowing bool) ArticleResponseBodyDTO {
	articleResponseDTO := ArticleResponseDTO{
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
	return ArticleResponseBodyDTO{Article: articleResponseDTO}
}

type AddCommentRequestBodyDTO struct {
	Comment AddCommentRequestDTO `json:"comment"`
}

type AddCommentRequestDTO struct {
	Body string `json:"body"`
}

type SingleCommentResponseBodyDTO struct {
	Comment CommentResponseDTO `json:"comment"`
}

type MultiCommentsResponseBodyDTO struct {
	Comment []CommentResponseDTO `json:"comment"`
}

type CommentResponseDTO struct {
	Id        string    `json:"id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Author    AuthorDTO `json:"author"`
}

func ToMultiCommentsResponseBodyDTO(comments []domain.Comment, authorIdToAuthorMap map[uuid.UUID]domain.User, isFollowingMap map[uuid.UUID]bool) MultiCommentsResponseBodyDTO {
	commentResponseDTOs := make([]CommentResponseDTO, len(comments))
	for _, comment := range comments {
		author := authorIdToAuthorMap[comment.AuthorId]
		commentResponseDTO := CommentResponseDTO{
			Id:        comment.Id.String(),
			Body:      comment.Body,
			CreatedAt: comment.CreatedAt,
			UpdatedAt: comment.UpdatedAt,
			Author: AuthorDTO{
				Username:  author.Username,
				Bio:       author.Bio,
				Image:     author.Image,
				Following: isFollowingMap[comment.AuthorId], // ToDo @ender we also need to get this from the database
			},
		}
		commentResponseDTOs = append(commentResponseDTOs, commentResponseDTO)
	}
	return MultiCommentsResponseBodyDTO{Comment: commentResponseDTOs}
}

func ToSingleCommentResponseBodyDTO(comment domain.Comment, author domain.User, isFollowing bool) SingleCommentResponseBodyDTO {
	commentResponseDTO := CommentResponseDTO{
		Id:        comment.Id.String(),
		Body:      comment.Body,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
		Author: AuthorDTO{
			Username:  author.Username,
			Bio:       author.Bio,
			Image:     author.Image,
			Following: isFollowing,
		},
	}
	return SingleCommentResponseBodyDTO{Comment: commentResponseDTO}
}

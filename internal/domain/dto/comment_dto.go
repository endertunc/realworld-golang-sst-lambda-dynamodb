package dto

import (
	"github.com/google/uuid"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"time"
)

// comment request dtos
type AddCommentRequestBodyDTO struct {
	Comment AddCommentRequestDTO `json:"comment" validate:"required"`
}

type AddCommentRequestDTO struct {
	Body string `json:"body" validate:"required,notblank,max=4096"`
}

func (s AddCommentRequestBodyDTO) Validate() (map[string]string, bool) {
	return validateStruct(s)
}

// comment response dtos
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

// factory methods
func ToMultiCommentsResponseBodyDTO(comments []domain.Comment, authorIdToAuthorMap map[uuid.UUID]domain.User, isFollowingMap map[uuid.UUID]bool) MultiCommentsResponseBodyDTO {
	commentResponseDTOs := make([]CommentResponseDTO, 0, len(comments))

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
				Following: isFollowingMap[comment.AuthorId],
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

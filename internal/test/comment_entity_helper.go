package test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"testing"
)

var DefaultAddCommentRequestDTO = dto.AddCommentRequestDTO{
	Body: "This is a test comment",
}

// CreateCommentEntity creates a comment for testing purposes
func CreateCommentEntity(t *testing.T, articleSlug string, comment dto.AddCommentRequestDTO, token string) dto.CommentResponseDTO {
	reqBody := dto.AddCommentRequestBodyDTO{
		Comment: comment,
	}
	var respBody dto.SingleCommentResponseBodyDTO
	MakeAuthenticatedRequestAndParseResponse(t, reqBody, "POST", "/api/articles/"+articleSlug+"/comments", http.StatusOK, &respBody, token)
	return respBody.Comment
}

// CreateDefaultComment creates a comment with default test data
func CreateDefaultComment(t *testing.T, articleSlug string, token string) dto.CommentResponseDTO {
	return CreateCommentEntity(t, articleSlug, DefaultAddCommentRequestDTO, token)
}

// CreateCommentWithBody creates a comment with a specific body
func CreateCommentWithBody(t *testing.T, articleSlug string, body string, token string) dto.CommentResponseDTO {
	comment := dto.AddCommentRequestDTO{Body: body}
	return CreateCommentEntity(t, articleSlug, comment, token)
}

// GetArticleComments retrieves all comments for an article
func GetArticleComments(t *testing.T, articleSlug string, token string) []dto.CommentResponseDTO {
	var respBody dto.MultiCommentsResponseBodyDTO
	MakeAuthenticatedRequestAndParseResponse(t, nil, "GET", fmt.Sprintf("/api/articles/%s/comments", articleSlug), http.StatusOK, &respBody, token)
	return respBody.Comment
}

// DeleteComment deletes a specific comment
func DeleteComment(t *testing.T, articleSlug string, commentId string, token string) {
	MakeAuthenticatedRequestAndParseResponse(t, nil, "DELETE", "/api/articles/"+articleSlug+"/comments/"+commentId, http.StatusOK, nil, token)
}

// VerifyCommentExists verifies that a specific comment exists in an article's comments
// ToDo @ender we are testing using API calls. Maybe we should use database in this case? It's a tricky one to choose
func VerifyCommentExists(t *testing.T, articleSlug string, commentId string, token string) {
	comments := GetArticleComments(t, articleSlug, token)
	found := false
	for _, comment := range comments {
		if comment.Id == commentId {
			found = true
			break
		}
	}
	require.True(t, found, "Comment with ID %s not found", commentId)
}

// VerifyCommentNotExists verifies that a specific comment does not exist in an article's comments
func VerifyCommentNotExists(t *testing.T, articleSlug string, commentId string, token string) {
	comments := GetArticleComments(t, articleSlug, token)
	for _, comment := range comments {
		require.NotEqual(t, commentId, comment.Id, "Comment with ID %s should not exist", commentId)
	}
}

package test

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"testing"
)

// CreateCommentEntity creates a comment for testing purposes
//func CreateCommentEntity(t *testing.T, articleSlug string, comment dto.AddCommentRequestDTO, token string) dto.CommentResponseDTO {
//	reqBody := dto.AddCommentRequestBodyDTO{
//		Comment: comment,
//	}
//	var respBody dto.SingleCommentResponseBodyDTO
//	MakeAuthenticatedRequestAndParseResponse(t, reqBody, "POST", "/api/articles/"+articleSlug+"/comments", http.StatusOK, &respBody, token)
//	return respBody.Comment
//}

//// CreateDefaultComment creates a comment with default test data
//func CreateDefaultComment(t *testing.T, articleSlug string, token string) dto.CommentResponseDTO {
//	return CreateComment(t, articleSlug, dtogen.GenerateAddCommentRequestDTO(), token)
//}

//// CreateCommentWithBody creates a comment with a specific body
//func CreateCommentWithBody(t *testing.T, articleSlug string, body string, token string) dto.CommentResponseDTO {
//	comment := dto.AddCommentRequestDTO{Body: body}
//	return CreateCommentEntity(t, articleSlug, comment, token)
//}

func CreateComment(t *testing.T, articleSlug string, reqBody dto.AddCommentRequestDTO, token string) dto.CommentResponseDTO {
	return CreateCommentWithResponse[dto.SingleCommentResponseBodyDTO](t, articleSlug, reqBody, token, http.StatusOK).Comment
}

func CreateCommentWithResponse[T interface{}](t *testing.T, articleSlug string, comment dto.AddCommentRequestDTO, token string, expectedStatusCode int) T {
	reqBody := dto.AddCommentRequestBodyDTO{Comment: comment}
	return ExecuteRequest[T](t, "POST", "/api/articles/"+articleSlug+"/comments", reqBody, expectedStatusCode, &token)
}

// GetArticleComments retrieves all comments for an article
func GetArticleComments(t *testing.T, articleSlug string, token *string) []dto.CommentResponseDTO {
	return GetArticleCommentsWithResponse[dto.MultiCommentsResponseBodyDTO](t, articleSlug, token, http.StatusOK).Comment
}

func GetArticleCommentsWithResponse[T interface{}](t *testing.T, articleSlug string, token *string, expectedStatusCode int) T {
	return ExecuteRequest[T](t, "GET", "/api/articles/"+articleSlug+"/comments", nil, expectedStatusCode, token)
}

// DeleteComment deletes a specific comment
func DeleteComment(t *testing.T, articleSlug string, commentId string, token string) {
	ExecuteRequest[Nothing](t, "DELETE", "/api/articles/"+articleSlug+"/comments/"+commentId, nil, http.StatusOK, &token)
}

func DeleteCommentWithResponse[T interface{}](t *testing.T, articleSlug string, commentId string, token string, expectedStatusCode int) T {
	return ExecuteRequest[T](t, "DELETE", "/api/articles/"+articleSlug+"/comments/"+commentId, nil, expectedStatusCode, &token)
}

// VerifyCommentExists verifies that a specific comment exists in an article's comments
func VerifyCommentExists(t *testing.T, articleSlug string, commentId string, token string) {
	comments := GetArticleComments(t, articleSlug, &token)
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
	comments := GetArticleComments(t, articleSlug, &token)
	for _, comment := range comments {
		require.NotEqual(t, commentId, comment.Id, "Comment with ID %s should not exist", commentId)
	}
}

func LoginUser(t *testing.T, user dto.LoginRequestUserDto) dto.UserResponseUserDto {
	return LoginUserWithResponse[dto.UserResponseBodyDTO](t, user, http.StatusOK).User
}

func LoginUserWithResponse[T interface{}](t *testing.T, user dto.LoginRequestUserDto, expectedStatusCode int) T {
	reqBody := dto.LoginRequestBodyDTO{User: user}
	return ExecuteRequest[T](t, "POST", "/api/users/login", reqBody, expectedStatusCode, nil)
}

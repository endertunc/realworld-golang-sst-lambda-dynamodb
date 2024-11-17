package test

import (
	"fmt"
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

func CreateCommentWithResponse[T interface{}](t *testing.T, articleSlug string, reqBody dto.AddCommentRequestDTO, token string, expectedStatusCode int) T {
	var respBody T
	MakeAuthenticatedRequestAndParseResponse(t, dto.AddCommentRequestBodyDTO{Comment: reqBody}, "POST", "/api/articles/"+articleSlug+"/comments", expectedStatusCode, &respBody, token)
	return respBody
}

// GetArticleComments retrieves all comments for an article
func GetArticleComments(t *testing.T, articleSlug string, token *string) []dto.CommentResponseDTO {
	return GetArticleCommentsWithResponse[dto.MultiCommentsResponseBodyDTO](t, articleSlug, token, http.StatusOK).Comment
}

func GetArticleCommentsWithResponse[T interface{}](t *testing.T, articleSlug string, token *string, expectedStatusCode int) T {
	var respBody T
	path := fmt.Sprintf("/api/articles/%s/comments", articleSlug)
	if token == nil {
		MakeRequestAndParseResponse(t, nil, "GET", path, http.StatusOK, &respBody)
	} else {
		MakeAuthenticatedRequestAndParseResponse(t, nil, "GET", path, expectedStatusCode, &respBody, *token)
	}
	return respBody
}

// DeleteComment deletes a specific comment
func DeleteComment(t *testing.T, articleSlug string, commentId string, token string) {
	//type emptyStruct struct {}
	//_ = DeleteCommentWithResponse[emptyStruct](t, articleSlug, commentId, token, http.StatusOK)
	MakeAuthenticatedRequestAndParseResponse(t, nil, "DELETE", "/api/articles/"+articleSlug+"/comments/"+commentId, http.StatusOK, nil, token)
}

func DeleteCommentWithResponse[T interface{}](t *testing.T, articleSlug string, commentId string, token string, expectedStatusCode int) T {
	var respBody T
	MakeAuthenticatedRequestAndParseResponse(t, nil, "DELETE", "/api/articles/"+articleSlug+"/comments/"+commentId, expectedStatusCode, &respBody, token)
	return respBody
}

// VerifyCommentExists verifies that a specific comment exists in an article's comments
// ToDo @ender we are testing using API calls. Maybe we should use database in this case? It's a tricky one to choose
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
	loginReqBody := dto.LoginRequestBodyDTO{User: user}
	var loginRespBody T
	MakeRequestAndParseResponse(t, loginReqBody, "POST", "/api/users/login", expectedStatusCode, &loginRespBody)
	return loginRespBody
}

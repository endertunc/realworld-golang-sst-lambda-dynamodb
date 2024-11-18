package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	dtogen "realworld-aws-lambda-dynamodb-golang/internal/domain/dto/generator"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/test"
	"testing"
)

func TestGetCommentsForArticleWithNoComments(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create a user and article
		_, token := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())
		article := test.CreateArticle(t, dtogen.GenerateCreateArticleRequestDTO(), token)

		// Get comments
		comments := test.GetArticleComments(t, article.Slug, &token) // with or without token doesn't matter

		// Verify an empty array is returned
		assert.Empty(t, comments)
	})
}

func TestGetCommentsForArticleWithMultipleComments(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create a user and article
		user, token := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())
		article := test.CreateArticle(t, dtogen.GenerateCreateArticleRequestDTO(), token)

		// Add multiple comments
		commentReq1 := dtogen.GenerateAddCommentRequestDTO()
		commentResp1 := test.CreateComment(t, article.Slug, commentReq1, token)
		commentReq2 := dtogen.GenerateAddCommentRequestDTO()
		commentResp2 := test.CreateComment(t, article.Slug, commentReq2, token)

		// Get comments
		comments := test.GetArticleComments(t, article.Slug, nil)

		// Verify comments are returned
		assert.Len(t, comments, 2)

		// Verify each comment matches the expected structure
		commentMap := make(map[string]dto.CommentResponseDTO)
		for _, c := range comments {
			commentMap[c.Id] = c
		}

		expectedComment1 := commentMap[commentResp1.Id]
		assert.Equal(t, commentReq1.Body, expectedComment1.Body)
		assert.Equal(t, user.Username, expectedComment1.Author.Username)
		assert.False(t, expectedComment1.Author.Following)

		expectedComment2 := commentMap[commentResp2.Id]
		assert.Equal(t, commentReq2.Body, expectedComment2.Body)
		assert.Equal(t, user.Username, expectedComment2.Author.Username)
		assert.False(t, expectedComment2.Author.Following)

	})
}

func TestGetCommentsAsAuthenticatedUser(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create article author and article
		_, authorToken := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())
		article := test.CreateArticle(t, dtogen.GenerateCreateArticleRequestDTO(), authorToken)

		// Create a followed user
		followedUser := dtogen.GenerateNewUserRequestUserDto()
		_, followedToken := test.CreateAndLoginUser(t, followedUser)

		// Create an unfollowed user
		unfollowedUser := dtogen.GenerateNewUserRequestUserDto()
		_, unfollowedToken := test.CreateAndLoginUser(t, unfollowedUser)

		// Create a viewerUser who will follow one of the users
		viewerUser := dtogen.GenerateNewUserRequestUserDto()
		_, viewerToken := test.CreateAndLoginUser(t, viewerUser)

		// Add comments from both users
		followedCommentReq := dtogen.GenerateAddCommentRequestDTO()
		followedComment := test.CreateComment(t, article.Slug, followedCommentReq, followedToken)
		unfollowedCommentReq := dtogen.GenerateAddCommentRequestDTO()
		unfollowedComment := test.CreateComment(t, article.Slug, unfollowedCommentReq, unfollowedToken)

		// Viewer follows one user
		test.FollowUser(t, followedUser.Username, viewerToken)

		// Get comments as viewerUser
		comments := test.GetArticleComments(t, article.Slug, &viewerToken)
		assert.Len(t, comments, 2)

		// Verify comments
		commentMap := make(map[string]dto.CommentResponseDTO)
		for _, c := range comments {
			commentMap[c.Id] = c
		}
		expectedFollowedComment := commentMap[followedComment.Id]
		assert.Equal(t, followedCommentReq.Body, expectedFollowedComment.Body)
		assert.Equal(t, followedUser.Username, expectedFollowedComment.Author.Username)
		assert.True(t, expectedFollowedComment.Author.Following)

		expectedUnfollowedComment := commentMap[unfollowedComment.Id]
		assert.Equal(t, unfollowedComment.Body, expectedUnfollowedComment.Body)
		assert.Equal(t, unfollowedUser.Username, expectedUnfollowedComment.Author.Username)
		assert.False(t, expectedUnfollowedComment.Author.Following)

	})
}

func TestGetCommentsAsAnonymousUser(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create user and article with comments
		user, token := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())
		article := test.CreateArticle(t, dtogen.GenerateCreateArticleRequestDTO(), token)
		commentReq := dtogen.GenerateAddCommentRequestDTO()
		test.CreateComment(t, article.Slug, commentReq, token)

		// Get comments without authentication
		comments := test.GetArticleComments(t, article.Slug, nil)

		assert.Len(t, comments, 1)

		commentResp := comments[0]
		assert.Equal(t, commentReq.Body, commentResp.Body)
		assert.Equal(t, user.Username, commentResp.Author.Username)
		assert.False(t, commentResp.Author.Following)
	})
}

// ToDo @ender check with auth as well
func TestGetCommentsForNonExistentArticle(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		respBody := test.GetArticleCommentsWithResponse[errutil.SimpleError](t, "non-existent-article", nil, http.StatusNotFound)
		assert.Equal(t, "article not found", respBody.Message)
	})
}

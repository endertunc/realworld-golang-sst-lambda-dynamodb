package main

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/test"
	"testing"
)

func TestGetCommentsForArticleWithNoComments(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create a user and article
		_, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)
		article := test.CreateDefaultArticle(t, token)

		// Get comments
		var respBody dto.MultiCommentsResponseBodyDTO
		test.MakeRequestAndParseResponse(t, nil, "GET",
			fmt.Sprintf("/api/articles/%s/comments", article.Slug),
			http.StatusOK, &respBody)

		// Verify an empty array is returned
		assert.Empty(t, respBody.Comment)
	})
}

func TestGetCommentsForArticleWithMultipleComments(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create a user and article
		user, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)
		article := test.CreateDefaultArticle(t, token)

		// Add multiple comments
		comment1 := test.CreateCommentWithBody(t, article.Slug, "First comment", token)
		comment2 := test.CreateCommentWithBody(t, article.Slug, "Second comment", token)

		// Get comments
		var respBody dto.MultiCommentsResponseBodyDTO
		test.MakeRequestAndParseResponse(t, nil, "GET",
			fmt.Sprintf("/api/articles/%s/comments", article.Slug),
			http.StatusOK, &respBody)

		// Verify comments are returned
		assert.Len(t, respBody.Comment, 2)

		// Create expected comments
		expectedComment1 := dto.CommentResponseDTO{
			Id:   comment1.Id,
			Body: "First comment",
			Author: dto.AuthorDTO{
				Username:  user.Username,
				Bio:       nil,
				Image:     nil,
				Following: false,
			},
			CreatedAt: comment1.CreatedAt,
			UpdatedAt: comment1.UpdatedAt,
		}

		expectedComment2 := dto.CommentResponseDTO{
			Id:   comment2.Id,
			Body: "Second comment",
			Author: dto.AuthorDTO{
				Username:  user.Username,
				Bio:       nil,
				Image:     nil,
				Following: false,
			},
			CreatedAt: comment2.CreatedAt,
			UpdatedAt: comment2.UpdatedAt,
		}

		// Verify each comment matches expected structure
		commentMap := make(map[string]dto.CommentResponseDTO)
		for _, c := range respBody.Comment {
			commentMap[c.Id] = c
		}

		assert.Equal(t, expectedComment1, commentMap[comment1.Id])
		assert.Equal(t, expectedComment2, commentMap[comment2.Id])
	})
}

func TestGetCommentsAsAuthenticatedUser(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create article author and article
		_, authorToken := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)
		article := test.CreateDefaultArticle(t, authorToken)

		// Create a followed user
		followedUser := dto.NewUserRequestUserDto{
			Username: "followed-user",
			Email:    "followed@example.com",
			Password: "password123",
		}
		_, followedToken := test.CreateAndLoginUser(t, followedUser)

		// Create an unfollowed user
		unfollowedUser := dto.NewUserRequestUserDto{
			Username: "unfollowed-user",
			Email:    "unfollowed@example.com",
			Password: "password123",
		}
		_, unfollowedToken := test.CreateAndLoginUser(t, unfollowedUser)

		// Create a viewer who will follow one of the users
		viewer := dto.NewUserRequestUserDto{
			Username: "viewer",
			Email:    "viewer@example.com",
			Password: "password123",
		}
		_, viewerToken := test.CreateAndLoginUser(t, viewer)

		// Add comments from both users
		followedCommentBody := "Comment from followed user"
		followedComment := test.CreateCommentWithBody(t, article.Slug, followedCommentBody, followedToken)
		unfollowedCommentBody := "Comment from unfollowed user"
		unfollowedComment := test.CreateCommentWithBody(t, article.Slug, unfollowedCommentBody, unfollowedToken)

		// Viewer follows one user
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "POST",
			fmt.Sprintf("/api/profiles/%s/follow", followedUser.Username),
			http.StatusOK, nil, viewerToken)

		// Get comments as viewer
		respBody := dto.MultiCommentsResponseBodyDTO{}
		test.MakeAuthenticatedRequestAndParseResponse(t, nil, "GET",
			fmt.Sprintf("/api/articles/%s/comments", article.Slug),
			http.StatusOK, &respBody, viewerToken)

		assert.Len(t, respBody.Comment, 2)

		// Create expected comments
		expectedFollowedComment := dto.CommentResponseDTO{
			Id:   followedComment.Id,
			Body: followedCommentBody,
			Author: dto.AuthorDTO{
				Username:  followedUser.Username,
				Bio:       nil,
				Image:     nil,
				Following: true,
			},
			CreatedAt: followedComment.CreatedAt,
			UpdatedAt: followedComment.UpdatedAt,
		}

		expectedUnfollowedComment := dto.CommentResponseDTO{
			Id:   unfollowedComment.Id,
			Body: unfollowedCommentBody,
			Author: dto.AuthorDTO{
				Username:  unfollowedUser.Username,
				Bio:       nil,
				Image:     nil,
				Following: false,
			},
			CreatedAt: unfollowedComment.CreatedAt,
			UpdatedAt: unfollowedComment.UpdatedAt,
		}

		slog.DebugContext(context.Background(), "expectedFollowedComment", slog.Any("response", respBody.Comment))

		// Verify comments
		commentMap := make(map[string]dto.CommentResponseDTO)
		for _, c := range respBody.Comment {
			commentMap[c.Id] = c
		}

		assert.Equal(t, expectedFollowedComment, commentMap[followedComment.Id])
		assert.Equal(t, expectedUnfollowedComment, commentMap[unfollowedComment.Id])
	})
}

func TestGetCommentsAsAnonymousUser(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// Create user and article with comments
		user, token := test.CreateAndLoginUser(t, test.DefaultNewUserRequestUserDto)
		article := test.CreateDefaultArticle(t, token)
		comment := test.CreateDefaultComment(t, article.Slug, token)

		// Get comments without authentication
		var respBody dto.MultiCommentsResponseBodyDTO
		test.MakeRequestAndParseResponse(t, nil, "GET",
			fmt.Sprintf("/api/articles/%s/comments", article.Slug),
			http.StatusOK, &respBody)

		assert.Len(t, respBody.Comment, 1)

		expectedComment := dto.CommentResponseDTO{
			Id:   comment.Id,
			Body: test.DefaultAddCommentRequestDTO.Body,
			Author: dto.AuthorDTO{
				Username:  user.Username,
				Bio:       nil,
				Image:     nil,
				Following: false,
			},
			CreatedAt: comment.CreatedAt,
			UpdatedAt: comment.UpdatedAt,
		}

		assert.Equal(t, expectedComment, respBody.Comment[0])
	})
}

func TestGetCommentsForNonExistentArticle(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		var respBody errutil.GenericError
		test.MakeRequestAndParseResponse(t, nil, "GET",
			"/api/articles/non-existent-article/comments",
			http.StatusNotFound, &respBody)

		assert.Equal(t, "article not found", respBody.Message)
	})
}

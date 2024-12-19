package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	dtogen "realworld-aws-lambda-dynamodb-golang/internal/domain/dto/generator"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/test"
	"testing"
)

func TestAuthenticationScenarios(t *testing.T) {
	test.RunAuthenticationTests(t, test.SharedAuthenticationTestConfig{
		Method: "DELETE",
		Path:   "/api/articles/some-article",
	})
}

func TestSuccessfulArticleDeletion(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// create a user and an article
		_, token := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())
		article := test.CreateArticle(t, dtogen.GenerateCreateArticleRequestDTO(), token)

		// verify article exists before deletion
		existingArticle := test.GetArticle(t, article.Slug, &token)
		assert.Equal(t, article.Slug, existingArticle.Slug)

		// delete the article
		test.DeleteArticle(t, article.Slug, token)

		// verify the article no longer exists
		resp := test.GetArticleWithResponse[errutil.SimpleError](t, article.Slug, &token, http.StatusNotFound)
		assert.Equal(t, "article not found", resp.Message)
	})
}

func TestDeleteNonExistingArticle(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// create a user
		_, token := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())

		// try to delete a non-existing article
		resp := test.DeleteArticleWithResponse[errutil.SimpleError](t, "non-existing-article", token, http.StatusNotFound)
		assert.Equal(t, "article not found", resp.Message)
	})
}

func TestDeleteArticleAsNonOwner(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// create two users
		_, ownerToken := test.CreateAndLoginUser(t, dtogen.GenerateNewUserRequestUserDto())
		nonOwner := dtogen.GenerateNewUserRequestUserDto()
		_, nonOwnerToken := test.CreateAndLoginUser(t, nonOwner)

		// create an article as owner
		article := test.CreateArticle(t, dtogen.GenerateCreateArticleRequestDTO(), ownerToken)

		// try to delete the article as non-owner
		respBody := test.DeleteArticleWithResponse[errutil.SimpleError](t, article.Slug, nonOwnerToken, http.StatusForbidden)
		assert.Equal(t, "forbidden", respBody.Message)

		// verify the article still exists
		existingArticle := test.GetArticle(t, article.Slug, &ownerToken)
		assert.Equal(t, article.Slug, existingArticle.Slug)
	})
}

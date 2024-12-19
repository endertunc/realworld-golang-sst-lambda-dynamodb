package main

import (
	"github.com/stretchr/testify/assert"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto/generator"
	"realworld-aws-lambda-dynamodb-golang/internal/test"
	"testing"
	"time"
)

func TestGetTags(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// create first author user
		author1User := generator.GenerateNewUserRequestUserDto()
		_, author1Token := test.CreateAndLoginUser(t, author1User)

		// create second author user
		author2User := generator.GenerateNewUserRequestUserDto()
		_, author2Token := test.CreateAndLoginUser(t, author2User)

		// create 3 articles in total
		createdArticle1 := test.CreateArticle(t, generator.GenerateCreateArticleRequestDTO(), author1Token)
		createdArticle2 := test.CreateArticle(t, generator.GenerateCreateArticleRequestDTO(), author2Token)
		createdArticle3 := test.CreateArticle(t, generator.GenerateCreateArticleRequestDTO(), author1Token)

		expectedTags := make([]string, 0)
		expectedTags = append(expectedTags, createdArticle1.TagList...)
		expectedTags = append(expectedTags, createdArticle2.TagList...)
		expectedTags = append(expectedTags, createdArticle3.TagList...)

		assert.EventuallyWithT(t, func(testingT *assert.CollectT) {
			tags := test.GetTags(t)
			// by default, each article has 2 tags and we have 3 articles
			assert.Equal(testingT, 6, len(tags.Tags))
			assert.ElementsMatch(testingT, expectedTags, tags.Tags)
		}, 60*time.Second, 5*time.Second)

	})
}

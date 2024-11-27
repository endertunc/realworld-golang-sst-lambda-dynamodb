//nolint:golint,exhaustruct
package main

import (
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto/generator"
	"realworld-aws-lambda-dynamodb-golang/internal/test"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ToDo @ender viewerUser should be renamed to participantUser I think.
// ToDo @ender visitorUser should be renamed to viewerUser I think.
// ToDo @ender we almost always have 2 author and 2 articles. We could move this to a setup function.
func TestListArticlesFavoritedByUserWithoutAuth(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// create users
		viewerUser := generator.GenerateNewUserRequestUserDto()
		_, viewerToken := test.CreateAndLoginUser(t, viewerUser)

		author1User := generator.GenerateNewUserRequestUserDto()
		_, author1Token := test.CreateAndLoginUser(t, author1User)

		author2User := generator.GenerateNewUserRequestUserDto()
		_, author2Token := test.CreateAndLoginUser(t, author2User)

		// create first articles for both authors
		createdArticle1 := test.CreateArticle(t, generator.GenerateCreateArticleRequestDTO(), author1Token)
		_ = test.CreateArticle(t, generator.GenerateCreateArticleRequestDTO(), author2Token)

		// create second articles for both authors
		_ = test.CreateArticle(t, generator.GenerateCreateArticleRequestDTO(), author1Token)
		createdArticle4 := test.CreateArticle(t, generator.GenerateCreateArticleRequestDTO(), author2Token)

		// viewer follows author1
		test.FollowUser(t, author1User.Username, viewerToken)

		// viewer favorites one article from each author
		test.FavoriteArticle(t, createdArticle1.Slug, viewerToken)
		test.FavoriteArticle(t, createdArticle4.Slug, viewerToken)

		// list articles favorited by viewer without auth
		listResponse := test.ListArticles(t, nil, test.ArticleQueryParams{Favorited: &viewerUser.Username})

		// verify response
		assert.Equal(t, 2, len(listResponse.Articles))
		assert.Equal(t, createdArticle4.Slug, listResponse.Articles[0].Slug) // most recent first
		assert.Equal(t, createdArticle1.Slug, listResponse.Articles[1].Slug)
		assert.False(t, listResponse.Articles[0].Favorited) // not favorited since no auth
		assert.False(t, listResponse.Articles[1].Favorited)
		assert.False(t, listResponse.Articles[0].Author.Following) // not following since no auth
		assert.False(t, listResponse.Articles[1].Author.Following)
		assert.Equal(t, 1, listResponse.Articles[0].FavoritesCount) // counts still show total favorites
		assert.Equal(t, 1, listResponse.Articles[1].FavoritesCount)
	})
}

func TestListArticlesFavoritedByUserWithVisitorFavorites(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// create users
		viewerUser := generator.GenerateNewUserRequestUserDto()
		_, viewerToken := test.CreateAndLoginUser(t, viewerUser)

		visitorUser := generator.GenerateNewUserRequestUserDto()
		_, visitorToken := test.CreateAndLoginUser(t, visitorUser)

		author1User := generator.GenerateNewUserRequestUserDto()
		_, author1Token := test.CreateAndLoginUser(t, author1User)

		author2User := generator.GenerateNewUserRequestUserDto()
		_, author2Token := test.CreateAndLoginUser(t, author2User)

		// create first articles for both authors
		createdArticle1 := test.CreateArticle(t, generator.GenerateCreateArticleRequestDTO(), author1Token)
		_ = test.CreateArticle(t, generator.GenerateCreateArticleRequestDTO(), author2Token)

		// create second articles for both authors
		_ = test.CreateArticle(t, generator.GenerateCreateArticleRequestDTO(), author1Token)
		createdArticle4 := test.CreateArticle(t, generator.GenerateCreateArticleRequestDTO(), author2Token)

		// viewer favorites article1 and article4
		test.FavoriteArticle(t, createdArticle1.Slug, viewerToken)
		test.FavoriteArticle(t, createdArticle4.Slug, viewerToken)

		// visitor favorites article1 (same as viewer)
		test.FavoriteArticle(t, createdArticle1.Slug, visitorToken)

		// visitor lists articles favorited by viewer
		listResponse := test.ListArticles(t, &visitorToken, test.ArticleQueryParams{Favorited: &viewerUser.Username})

		// verify response
		assert.Equal(t, 2, len(listResponse.Articles))
		assert.Equal(t, createdArticle4.Slug, listResponse.Articles[0].Slug) // most recent first
		assert.Equal(t, createdArticle1.Slug, listResponse.Articles[1].Slug)
		assert.False(t, listResponse.Articles[0].Favorited) // visitor didn't favorite article4
		assert.True(t, listResponse.Articles[1].Favorited)  // visitor favorited article1
		assert.Equal(t, 1, listResponse.Articles[0].FavoritesCount)
		assert.Equal(t, 2, listResponse.Articles[1].FavoritesCount)
	})
}

func TestListArticlesFavoritedByUserWithVisitorFollowing(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// create users
		viewerUser := generator.GenerateNewUserRequestUserDto()
		_, viewerToken := test.CreateAndLoginUser(t, viewerUser)

		visitorUser := generator.GenerateNewUserRequestUserDto()
		_, visitorToken := test.CreateAndLoginUser(t, visitorUser)

		author1User := generator.GenerateNewUserRequestUserDto()
		_, author1Token := test.CreateAndLoginUser(t, author1User)

		author2User := generator.GenerateNewUserRequestUserDto()
		_, author2Token := test.CreateAndLoginUser(t, author2User)

		// create first articles for both authors
		createdArticle1 := test.CreateArticle(t, generator.GenerateCreateArticleRequestDTO(), author1Token)
		_ = test.CreateArticle(t, generator.GenerateCreateArticleRequestDTO(), author2Token)

		// create second articles for both authors
		_ = test.CreateArticle(t, generator.GenerateCreateArticleRequestDTO(), author1Token)
		createdArticle4 := test.CreateArticle(t, generator.GenerateCreateArticleRequestDTO(), author2Token)

		// viewer favorites article1 and article4
		test.FavoriteArticle(t, createdArticle1.Slug, viewerToken)
		test.FavoriteArticle(t, createdArticle4.Slug, viewerToken)

		// visitor follows author1
		test.FollowUser(t, author1User.Username, visitorToken)

		// visitor lists articles favorited by viewer
		listResponse := test.ListArticles(t, &visitorToken, test.ArticleQueryParams{Favorited: &viewerUser.Username})

		// verify response
		assert.Equal(t, 2, len(listResponse.Articles))
		assert.Equal(t, createdArticle4.Slug, listResponse.Articles[0].Slug) // most recent first
		assert.Equal(t, createdArticle1.Slug, listResponse.Articles[1].Slug)
		assert.False(t, listResponse.Articles[0].Author.Following) // visitor doesn't follow author2
		assert.True(t, listResponse.Articles[1].Author.Following)  // visitor follows author1
		assert.Equal(t, 1, listResponse.Articles[0].FavoritesCount)
		assert.Equal(t, 1, listResponse.Articles[1].FavoritesCount)
	})
}

func TestListArticlesFavoritedByUserWithBothFavoritesAndFollowing(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// create users
		viewerUser := generator.GenerateNewUserRequestUserDto()
		_, viewerToken := test.CreateAndLoginUser(t, viewerUser)

		visitorUser := generator.GenerateNewUserRequestUserDto()
		_, visitorToken := test.CreateAndLoginUser(t, visitorUser)

		author1User := generator.GenerateNewUserRequestUserDto()
		_, author1Token := test.CreateAndLoginUser(t, author1User)

		author2User := generator.GenerateNewUserRequestUserDto()
		_, author2Token := test.CreateAndLoginUser(t, author2User)

		// create first articles for both authors
		article1 := generator.GenerateCreateArticleRequestDTO()
		createdArticle1 := test.CreateArticle(t, article1, author1Token)

		_ = test.CreateArticle(t, generator.GenerateCreateArticleRequestDTO(), author2Token)

		// create second articles for both authors
		_ = test.CreateArticle(t, generator.GenerateCreateArticleRequestDTO(), author1Token)

		createdArticle4 := test.CreateArticle(t, generator.GenerateCreateArticleRequestDTO(), author2Token)

		// viewer favorites article1 and article4
		test.FavoriteArticle(t, createdArticle1.Slug, viewerToken)
		test.FavoriteArticle(t, createdArticle4.Slug, viewerToken)

		// visitor favorites article1 and follows author1
		test.FavoriteArticle(t, createdArticle1.Slug, visitorToken)
		test.FollowUser(t, author1User.Username, visitorToken)

		// visitor lists articles favorited by viewer
		listResponse := test.ListArticles(t, &visitorToken, test.ArticleQueryParams{Favorited: &viewerUser.Username})

		// verify response
		assert.Equal(t, 2, len(listResponse.Articles))
		assert.Equal(t, createdArticle4.Slug, listResponse.Articles[0].Slug) // most recent first
		assert.Equal(t, createdArticle1.Slug, listResponse.Articles[1].Slug)
		assert.False(t, listResponse.Articles[0].Favorited)        // visitor didn't favorite article4
		assert.True(t, listResponse.Articles[1].Favorited)         // visitor favorited article1
		assert.False(t, listResponse.Articles[0].Author.Following) // visitor doesn't follow author2
		assert.True(t, listResponse.Articles[1].Author.Following)  // visitor follows author1
		assert.Equal(t, 1, listResponse.Articles[0].FavoritesCount)
		assert.Equal(t, 2, listResponse.Articles[1].FavoritesCount)
	})
}

func TestListArticlesFavoritedByUserWithNoOverlap(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// create users
		viewerUser := generator.GenerateNewUserRequestUserDto()
		_, viewerToken := test.CreateAndLoginUser(t, viewerUser)

		visitorUser := generator.GenerateNewUserRequestUserDto()
		_, visitorToken := test.CreateAndLoginUser(t, visitorUser)

		author1User := generator.GenerateNewUserRequestUserDto()
		_, author1Token := test.CreateAndLoginUser(t, author1User)

		author2User := generator.GenerateNewUserRequestUserDto()
		_, author2Token := test.CreateAndLoginUser(t, author2User)

		// create first articles for both authors
		createdArticle1 := test.CreateArticle(t, generator.GenerateCreateArticleRequestDTO(), author1Token)
		createdArticle2 := test.CreateArticle(t, generator.GenerateCreateArticleRequestDTO(), author2Token)

		// create second articles for both authors
		_ = test.CreateArticle(t, generator.GenerateCreateArticleRequestDTO(), author1Token)
		createdArticle4 := test.CreateArticle(t, generator.GenerateCreateArticleRequestDTO(), author2Token)

		// viewer favorites article1 and article4
		test.FavoriteArticle(t, createdArticle1.Slug, viewerToken)
		test.FavoriteArticle(t, createdArticle4.Slug, viewerToken)

		// visitor favorites article2 and follows author2
		test.FavoriteArticle(t, createdArticle2.Slug, visitorToken)
		test.FollowUser(t, author2User.Username, visitorToken)

		// visitor lists articles favorited by viewer
		listResponse := test.ListArticles(t, &visitorToken, test.ArticleQueryParams{Favorited: &viewerUser.Username})

		// verify response
		assert.Equal(t, 2, len(listResponse.Articles))
		assert.Equal(t, createdArticle4.Slug, listResponse.Articles[0].Slug) // most recent first
		assert.Equal(t, createdArticle1.Slug, listResponse.Articles[1].Slug)
		assert.False(t, listResponse.Articles[0].Favorited) // visitor favorited different articles
		assert.False(t, listResponse.Articles[1].Favorited)
		assert.True(t, listResponse.Articles[0].Author.Following)  // visitor follows author2
		assert.False(t, listResponse.Articles[1].Author.Following) // visitor doesn't follow author1
		assert.Equal(t, 1, listResponse.Articles[0].FavoritesCount)
		assert.Equal(t, 1, listResponse.Articles[1].FavoritesCount)
	})
}

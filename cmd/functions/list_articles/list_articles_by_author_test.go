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
func TestListArticlesByAuthorWithoutAuth(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// create viewer user
		viewerUser := generator.GenerateNewUserRequestUserDto()
		_, viewerToken := test.CreateAndLoginUser(t, viewerUser)

		// create first author user
		author1User := generator.GenerateNewUserRequestUserDto()
		_, author1Token := test.CreateAndLoginUser(t, author1User)

		// create second author user
		author2User := generator.GenerateNewUserRequestUserDto()
		_, author2Token := test.CreateAndLoginUser(t, author2User)

		// create first articles for both authors
		article1 := generator.GenerateCreateArticleRequestDTO()
		createdArticle1 := test.CreateArticle(t, article1, author1Token)

		article2 := generator.GenerateCreateArticleRequestDTO()
		createdArticle2 := test.CreateArticle(t, article2, author2Token)

		// create second articles for both authors
		article3 := generator.GenerateCreateArticleRequestDTO()
		createdArticle3 := test.CreateArticle(t, article3, author1Token)

		article4 := generator.GenerateCreateArticleRequestDTO()
		createdArticle4 := test.CreateArticle(t, article4, author2Token)

		// viewer follows author1 and favorites articles
		test.FollowUser(t, author1User.Username, viewerToken)
		test.FavoriteArticle(t, createdArticle1.Slug, viewerToken)
		test.FavoriteArticle(t, createdArticle4.Slug, viewerToken)

		// test listing articles by author1 without auth
		listResponseNoAuth := test.ListArticles(t, nil, test.ArticleQueryParams{Author: &author1User.Username})

		// verify response without auth for author1
		assert.Equal(t, 2, len(listResponseNoAuth.Articles))
		assert.Equal(t, createdArticle3.Slug, listResponseNoAuth.Articles[0].Slug) // most recent first
		assert.Equal(t, createdArticle1.Slug, listResponseNoAuth.Articles[1].Slug)
		assert.False(t, listResponseNoAuth.Articles[0].Favorited) // not favorited since no auth
		assert.False(t, listResponseNoAuth.Articles[1].Favorited)
		assert.Equal(t, 0, listResponseNoAuth.Articles[0].FavoritesCount)
		assert.Equal(t, 1, listResponseNoAuth.Articles[1].FavoritesCount) // viewer favorited
		assert.False(t, listResponseNoAuth.Articles[0].Author.Following)  // not following since no auth
		assert.False(t, listResponseNoAuth.Articles[1].Author.Following)

		// test listing articles by author2 without auth
		listResponseNoAuth2 := test.ListArticles(t, nil, test.ArticleQueryParams{Author: &author2User.Username})

		// verify response without auth for author2
		assert.Equal(t, 2, len(listResponseNoAuth2.Articles))
		assert.Equal(t, createdArticle4.Slug, listResponseNoAuth2.Articles[0].Slug) // most recent first
		assert.Equal(t, createdArticle2.Slug, listResponseNoAuth2.Articles[1].Slug)
		assert.False(t, listResponseNoAuth2.Articles[0].Favorited) // not favorited since no auth
		assert.False(t, listResponseNoAuth2.Articles[1].Favorited)
		assert.Equal(t, 1, listResponseNoAuth2.Articles[0].FavoritesCount) // viewer favorited
		assert.Equal(t, 0, listResponseNoAuth2.Articles[1].FavoritesCount)
		assert.False(t, listResponseNoAuth2.Articles[0].Author.Following) // not following since no auth
		assert.False(t, listResponseNoAuth2.Articles[1].Author.Following)
	})
}

func TestListArticlesByAuthorWithoutFollowingAndFavorite(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// create viewer user
		viewerUser := generator.GenerateNewUserRequestUserDto()
		_, viewerToken := test.CreateAndLoginUser(t, viewerUser)

		// create visitor user
		visitorUser := generator.GenerateNewUserRequestUserDto()
		_, visitorToken := test.CreateAndLoginUser(t, visitorUser)

		// create first author user
		author1User := generator.GenerateNewUserRequestUserDto()
		_, author1Token := test.CreateAndLoginUser(t, author1User)

		// create second author user
		author2User := generator.GenerateNewUserRequestUserDto()
		_, author2Token := test.CreateAndLoginUser(t, author2User)

		// create first articles for both authors
		article1 := generator.GenerateCreateArticleRequestDTO()
		createdArticle1 := test.CreateArticle(t, article1, author1Token)

		article2 := generator.GenerateCreateArticleRequestDTO()
		createdArticle2 := test.CreateArticle(t, article2, author2Token)

		// create second articles for both authors
		article3 := generator.GenerateCreateArticleRequestDTO()
		createdArticle3 := test.CreateArticle(t, article3, author1Token)

		article4 := generator.GenerateCreateArticleRequestDTO()
		createdArticle4 := test.CreateArticle(t, article4, author2Token)

		// viewer follows author1 and favorites articles
		test.FollowUser(t, author1User.Username, viewerToken)
		test.FavoriteArticle(t, createdArticle1.Slug, viewerToken)
		test.FavoriteArticle(t, createdArticle4.Slug, viewerToken)

		// test listing articles by author1 with visitor auth
		listResponseWithAuth := test.ListArticles(t, &visitorToken, test.ArticleQueryParams{Author: &author1User.Username})

		// verify response with visitor auth for author1
		assert.Equal(t, 2, len(listResponseWithAuth.Articles))
		assert.Equal(t, createdArticle3.Slug, listResponseWithAuth.Articles[0].Slug) // most recent first
		assert.Equal(t, createdArticle1.Slug, listResponseWithAuth.Articles[1].Slug)
		assert.False(t, listResponseWithAuth.Articles[0].Favorited) // visitor didn't favorite
		assert.False(t, listResponseWithAuth.Articles[1].Favorited)
		assert.Equal(t, 0, listResponseWithAuth.Articles[0].FavoritesCount)
		assert.Equal(t, 1, listResponseWithAuth.Articles[1].FavoritesCount) // viewer favorited
		assert.False(t, listResponseWithAuth.Articles[0].Author.Following)  // visitor doesn't follow
		assert.False(t, listResponseWithAuth.Articles[1].Author.Following)

		// test listing articles by author2 with visitor auth
		listResponseWithAuth2 := test.ListArticles(t, &visitorToken, test.ArticleQueryParams{Author: &author2User.Username})

		// verify response with visitor auth for author2
		assert.Equal(t, 2, len(listResponseWithAuth2.Articles))
		assert.Equal(t, createdArticle4.Slug, listResponseWithAuth2.Articles[0].Slug) // most recent first
		assert.Equal(t, createdArticle2.Slug, listResponseWithAuth2.Articles[1].Slug)
		assert.False(t, listResponseWithAuth2.Articles[0].Favorited) // visitor didn't favorite
		assert.False(t, listResponseWithAuth2.Articles[1].Favorited)
		assert.Equal(t, 1, listResponseWithAuth2.Articles[0].FavoritesCount) // viewer favorited
		assert.Equal(t, 0, listResponseWithAuth2.Articles[1].FavoritesCount)
		assert.False(t, listResponseWithAuth2.Articles[0].Author.Following) // visitor doesn't follow
		assert.False(t, listResponseWithAuth2.Articles[1].Author.Following)
	})
}

func TestListArticlesByAuthorWithFollowing(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// create viewer user
		viewerUser := generator.GenerateNewUserRequestUserDto()
		_, viewerToken := test.CreateAndLoginUser(t, viewerUser)

		// create visitor user
		visitorUser := generator.GenerateNewUserRequestUserDto()
		_, visitorToken := test.CreateAndLoginUser(t, visitorUser)

		// create first author user
		author1User := generator.GenerateNewUserRequestUserDto()
		_, author1Token := test.CreateAndLoginUser(t, author1User)

		// create second author user
		author2User := generator.GenerateNewUserRequestUserDto()
		_, author2Token := test.CreateAndLoginUser(t, author2User)

		// create first articles for both authors
		article1 := generator.GenerateCreateArticleRequestDTO()
		createdArticle1 := test.CreateArticle(t, article1, author1Token)

		article2 := generator.GenerateCreateArticleRequestDTO()
		createdArticle2 := test.CreateArticle(t, article2, author2Token)

		// create second articles for both authors
		article3 := generator.GenerateCreateArticleRequestDTO()
		createdArticle3 := test.CreateArticle(t, article3, author1Token)

		article4 := generator.GenerateCreateArticleRequestDTO()
		createdArticle4 := test.CreateArticle(t, article4, author2Token)

		// viewer follows author1 and favorites articles
		test.FollowUser(t, author1User.Username, viewerToken)
		test.FavoriteArticle(t, createdArticle1.Slug, viewerToken)
		test.FavoriteArticle(t, createdArticle4.Slug, viewerToken)

		// visitor follows author1
		test.FollowUser(t, author1User.Username, visitorToken)

		// test listing articles by author1 with visitor auth
		listResponse := test.ListArticles(t, &visitorToken, test.ArticleQueryParams{Author: &author1User.Username})

		// verify response for followed author
		assert.Equal(t, 2, len(listResponse.Articles))
		assert.Equal(t, createdArticle3.Slug, listResponse.Articles[0].Slug) // most recent first
		assert.Equal(t, createdArticle1.Slug, listResponse.Articles[1].Slug)
		assert.False(t, listResponse.Articles[0].Favorited) // visitor didn't favorite
		assert.False(t, listResponse.Articles[1].Favorited)
		assert.Equal(t, 0, listResponse.Articles[0].FavoritesCount)
		assert.Equal(t, 1, listResponse.Articles[1].FavoritesCount) // viewer favorited
		assert.True(t, listResponse.Articles[0].Author.Following)   // visitor follows author1
		assert.True(t, listResponse.Articles[1].Author.Following)

		// test listing articles by author2 with visitor auth
		listResponse2 := test.ListArticles(t, &visitorToken, test.ArticleQueryParams{Author: &author2User.Username})

		// verify response for non-followed author
		assert.Equal(t, 2, len(listResponse2.Articles))
		assert.Equal(t, createdArticle4.Slug, listResponse2.Articles[0].Slug) // most recent first
		assert.Equal(t, createdArticle2.Slug, listResponse2.Articles[1].Slug)
		assert.False(t, listResponse2.Articles[0].Favorited) // visitor didn't favorite
		assert.False(t, listResponse2.Articles[1].Favorited)
		assert.Equal(t, 1, listResponse2.Articles[0].FavoritesCount) // viewer favorited
		assert.Equal(t, 0, listResponse2.Articles[1].FavoritesCount)
		assert.False(t, listResponse2.Articles[0].Author.Following) // visitor doesn't follow author2
		assert.False(t, listResponse2.Articles[1].Author.Following)
	})
}

func TestListArticlesByAuthorWithFavorite(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// create viewer user
		viewerUser := generator.GenerateNewUserRequestUserDto()
		_, viewerToken := test.CreateAndLoginUser(t, viewerUser)

		// create visitor user
		visitorUser := generator.GenerateNewUserRequestUserDto()
		_, visitorToken := test.CreateAndLoginUser(t, visitorUser)

		// create first author user
		author1User := generator.GenerateNewUserRequestUserDto()
		_, author1Token := test.CreateAndLoginUser(t, author1User)

		// create second author user
		author2User := generator.GenerateNewUserRequestUserDto()
		_, author2Token := test.CreateAndLoginUser(t, author2User)

		// create first articles for both authors
		article1 := generator.GenerateCreateArticleRequestDTO()
		createdArticle1 := test.CreateArticle(t, article1, author1Token)

		article2 := generator.GenerateCreateArticleRequestDTO()
		createdArticle2 := test.CreateArticle(t, article2, author2Token)

		// create second articles for both authors
		article3 := generator.GenerateCreateArticleRequestDTO()
		createdArticle3 := test.CreateArticle(t, article3, author1Token)

		article4 := generator.GenerateCreateArticleRequestDTO()
		createdArticle4 := test.CreateArticle(t, article4, author2Token)

		// viewer follows author1 and favorites articles
		test.FollowUser(t, author1User.Username, viewerToken)
		test.FavoriteArticle(t, createdArticle1.Slug, viewerToken)
		test.FavoriteArticle(t, createdArticle4.Slug, viewerToken)

		// visitor favorites author1's first article
		test.FavoriteArticle(t, createdArticle1.Slug, visitorToken)

		// test listing articles by author1 with visitor auth
		listResponse := test.ListArticles(t, &visitorToken, test.ArticleQueryParams{Author: &author1User.Username})

		// verify response for author with favorited article
		assert.Equal(t, 2, len(listResponse.Articles))
		assert.Equal(t, createdArticle3.Slug, listResponse.Articles[0].Slug) // most recent first
		assert.Equal(t, createdArticle1.Slug, listResponse.Articles[1].Slug)
		assert.False(t, listResponse.Articles[0].Favorited) // visitor didn't favorite
		assert.True(t, listResponse.Articles[1].Favorited)  // visitor favorited
		assert.Equal(t, 0, listResponse.Articles[0].FavoritesCount)
		assert.Equal(t, 2, listResponse.Articles[1].FavoritesCount) // both viewer and visitor favorited

		// test listing articles by author2 with visitor auth
		listResponse2 := test.ListArticles(t, &visitorToken, test.ArticleQueryParams{Author: &author2User.Username})

		// verify response for author without favorited articles
		assert.Equal(t, 2, len(listResponse2.Articles))
		assert.Equal(t, createdArticle4.Slug, listResponse2.Articles[0].Slug) // most recent first
		assert.Equal(t, createdArticle2.Slug, listResponse2.Articles[1].Slug)
		assert.False(t, listResponse2.Articles[0].Favorited) // visitor didn't favorite
		assert.False(t, listResponse2.Articles[1].Favorited)
		assert.Equal(t, 1, listResponse2.Articles[0].FavoritesCount) // viewer favorited
		assert.Equal(t, 0, listResponse2.Articles[1].FavoritesCount)
	})
}

func TestListArticlesByAuthorWithNoOverlap(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		// create viewer user
		viewerUser := generator.GenerateNewUserRequestUserDto()
		_, viewerToken := test.CreateAndLoginUser(t, viewerUser)

		// create visitor user
		visitorUser := generator.GenerateNewUserRequestUserDto()
		_, visitorToken := test.CreateAndLoginUser(t, visitorUser)

		// create first author user
		author1User := generator.GenerateNewUserRequestUserDto()
		_, author1Token := test.CreateAndLoginUser(t, author1User)

		// create second author user
		author2User := generator.GenerateNewUserRequestUserDto()
		_, author2Token := test.CreateAndLoginUser(t, author2User)

		// create first articles for both authors
		article1 := generator.GenerateCreateArticleRequestDTO()
		createdArticle1 := test.CreateArticle(t, article1, author1Token)

		article2 := generator.GenerateCreateArticleRequestDTO()
		createdArticle2 := test.CreateArticle(t, article2, author2Token)

		// create second articles for both authors
		article3 := generator.GenerateCreateArticleRequestDTO()
		createdArticle3 := test.CreateArticle(t, article3, author1Token)

		article4 := generator.GenerateCreateArticleRequestDTO()
		createdArticle4 := test.CreateArticle(t, article4, author2Token)

		// viewer follows author1 and favorites articles from both authors
		test.FollowUser(t, author1User.Username, viewerToken)
		test.FavoriteArticle(t, createdArticle1.Slug, viewerToken)
		test.FavoriteArticle(t, createdArticle4.Slug, viewerToken)

		// visitor follows author2 and favorites articles from both authors
		test.FollowUser(t, author2User.Username, visitorToken)
		test.FavoriteArticle(t, createdArticle2.Slug, visitorToken)
		test.FavoriteArticle(t, createdArticle3.Slug, visitorToken)

		// test listing articles by author1 with visitor auth
		listResponse := test.ListArticles(t, &visitorToken, test.ArticleQueryParams{Author: &author1User.Username})

		// verify response for author with favorited article
		assert.Equal(t, 2, len(listResponse.Articles))
		assert.Equal(t, createdArticle3.Slug, listResponse.Articles[0].Slug) // most recent first
		assert.Equal(t, createdArticle1.Slug, listResponse.Articles[1].Slug)
		assert.True(t, listResponse.Articles[0].Favorited)          // visitor didn't favorite
		assert.False(t, listResponse.Articles[1].Favorited)         // visitor favorited
		assert.Equal(t, 1, listResponse.Articles[0].FavoritesCount) // favorited by visitor
		assert.Equal(t, 1, listResponse.Articles[1].FavoritesCount) // favorited by viewer
		assert.False(t, listResponse.Articles[0].Author.Following)  // visitor follows author1
		assert.False(t, listResponse.Articles[1].Author.Following)

		// test listing articles by author2 with visitor auth
		listResponse2 := test.ListArticles(t, &visitorToken, test.ArticleQueryParams{Author: &author2User.Username})

		// verify response for author without favorited articles
		assert.Equal(t, 2, len(listResponse2.Articles))
		assert.Equal(t, createdArticle4.Slug, listResponse2.Articles[0].Slug) // most recent first
		assert.Equal(t, createdArticle2.Slug, listResponse2.Articles[1].Slug)
		assert.False(t, listResponse2.Articles[0].Favorited)         // visitor didn't favorite
		assert.True(t, listResponse2.Articles[1].Favorited)          // visitor favorited
		assert.Equal(t, 1, listResponse2.Articles[0].FavoritesCount) // viewer favorited
		assert.Equal(t, 1, listResponse2.Articles[1].FavoritesCount)
		assert.True(t, listResponse2.Articles[0].Author.Following) // visitor doesn't follow author2
		assert.True(t, listResponse2.Articles[1].Author.Following)
	})
}

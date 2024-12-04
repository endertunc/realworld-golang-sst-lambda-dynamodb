//nolint:golint,exhaustruct
package service

import (
	"context"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/generator"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/service/mocks"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/stretchr/testify/mock"

	"realworld-aws-lambda-dynamodb-golang/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	rmocks "realworld-aws-lambda-dynamodb-golang/internal/repository/mocks"
)

var (
	ctx   = context.Background()
	limit = 10
)

func TestListArticleByAuthor(t *testing.T) {
	var (
		nextPageTokenRequest  *string = nil
		nextPageTokenResponse *string = nil
	)

	t.Run("user not found", func(t *testing.T) {
		WithTestContext(t, func(tc articleTestContext) {
			// Setup test data
			author := generator.GenerateUser()

			tc.mockUserService.EXPECT().
				GetUserByUsername(mock.Anything, author.Username).
				Return(domain.User{}, errutil.ErrUserNotFound)

			// Execute
			result, _, err := tc.articleListService.GetMostRecentArticlesByAuthor(ctx, nil, author.Username, limit, nextPageTokenRequest)

			// Assert
			assert.ErrorIs(t, err, errutil.ErrUserNotFound)
			assert.Empty(t, result)
		})

	})

	t.Run("no auth", func(t *testing.T) {
		WithTestContext(t, func(tc articleTestContext) {
			// Setup test data
			author := generator.GenerateUser()

			article1 := generator.GenerateArticle()
			article1.AuthorId = author.Id

			article2 := generator.GenerateArticle()
			article2.AuthorId = author.Id

			// Setup expectations
			tc.mockUserService.EXPECT().
				GetUserByUsername(mock.Anything, author.Username).
				Return(author, nil)

			tc.mockUserService.EXPECT().
				GetUserListByUserIDs(mock.Anything, []uuid.UUID{author.Id}).
				Return([]domain.User{author}, nil)

			tc.mockArticleRepo.EXPECT().
				FindArticlesByAuthor(mock.Anything, author.Id, limit, nextPageTokenRequest).
				Return([]domain.Article{article1, article2}, nextPageTokenResponse, nil)

			// Execute
			result, nextToken, err := tc.articleListService.GetMostRecentArticlesByAuthor(ctx, nil, author.Username, limit, nextPageTokenRequest)

			// Assert
			assert.NoError(t, err)
			assert.Len(t, result, 2)
			assert.Equal(t, nextPageTokenResponse, nextToken)

			assert.Equal(t, article1.Id, result[0].Article.Id)
			assert.Equal(t, author.Username, result[0].Author.Username)
			assert.False(t, result[0].IsFavorited)
			assert.False(t, result[0].IsFollowing)

			assert.Equal(t, article2.Id, result[1].Article.Id)
			assert.Equal(t, author.Username, result[1].Author.Username)
			assert.False(t, result[1].IsFavorited)
			assert.False(t, result[1].IsFollowing)
		})

	})

	t.Run("viewer without following the author and favorited article", func(t *testing.T) {
		WithTestContext(t, func(tc articleTestContext) {
			// Setup test data
			author := generator.GenerateUser()
			viewer := generator.GenerateUser()

			article1 := generator.GenerateArticle()
			article1.AuthorId = author.Id

			article2 := generator.GenerateArticle()
			article2.AuthorId = author.Id

			// Setup expectations
			tc.mockUserService.EXPECT().
				GetUserByUsername(mock.Anything, author.Username).
				Return(author, nil)

			tc.mockUserService.EXPECT().
				GetUserListByUserIDs(mock.Anything, []uuid.UUID{author.Id}).
				Return([]domain.User{author}, nil)

			tc.mockArticleRepo.EXPECT().
				FindArticlesByAuthor(mock.Anything, author.Id, limit, nextPageTokenRequest).
				Return([]domain.Article{article1, article2}, nextPageTokenResponse, nil)

			tc.mockProfileService.EXPECT().
				IsFollowingBulk(mock.Anything, viewer.Id, []uuid.UUID{author.Id}).
				Return(mapset.NewSetWithSize[uuid.UUID](0), nil)

			tc.mockArticleRepo.EXPECT().
				IsFavoritedBulk(mock.Anything, viewer.Id, []uuid.UUID{article1.Id, article2.Id}).
				Return(mapset.NewSetWithSize[uuid.UUID](0), nil)

			// Execute
			result, _, err := tc.articleListService.GetMostRecentArticlesByAuthor(ctx, &viewer.Id, author.Username, limit, nextPageTokenRequest)

			// Assert
			assert.NoError(t, err)
			assert.Len(t, result, 2)

			// verify specific details
			assert.Equal(t, article1.Id, result[0].Article.Id)
			assert.Equal(t, author.Username, result[0].Author.Username)
			assert.False(t, result[0].IsFavorited)
			assert.False(t, result[0].IsFollowing)

			assert.Equal(t, article2.Id, result[1].Article.Id)
			assert.Equal(t, author.Username, result[1].Author.Username)
			assert.False(t, result[1].IsFavorited)
			assert.False(t, result[1].IsFollowing)
		})

	})

	t.Run("viewer with following the author and favorited article", func(t *testing.T) {
		WithTestContext(t, func(tc articleTestContext) {
			// Setup test data
			author := generator.GenerateUser()
			viewer := generator.GenerateUser()

			article1 := generator.GenerateArticle()
			article1.AuthorId = author.Id

			article2 := generator.GenerateArticle()
			article2.AuthorId = author.Id

			tc.mockUserService.EXPECT().
				GetUserByUsername(mock.Anything, author.Username).
				Return(author, nil)

			tc.mockUserService.EXPECT().
				GetUserListByUserIDs(mock.Anything, []uuid.UUID{author.Id}).
				Return([]domain.User{author}, nil)

			tc.mockArticleRepo.EXPECT().
				FindArticlesByAuthor(mock.Anything, author.Id, limit, nextPageTokenRequest).
				Return([]domain.Article{article1, article2}, nextPageTokenResponse, nil)

			tc.mockProfileService.EXPECT().
				IsFollowingBulk(mock.Anything, viewer.Id, []uuid.UUID{author.Id}).
				Return(mapset.NewSet[uuid.UUID](author.Id), nil)

			tc.mockArticleRepo.EXPECT().
				IsFavoritedBulk(mock.Anything, viewer.Id, []uuid.UUID{article1.Id, article2.Id}).
				Return(mapset.NewSet[uuid.UUID](article1.Id), nil)

			// Execute
			result, _, err := tc.articleListService.GetMostRecentArticlesByAuthor(ctx, &viewer.Id, author.Username, limit, nextPageTokenRequest)

			// Assert
			assert.NoError(t, err)
			assert.Len(t, result, 2)

			// verify specific details
			assert.Equal(t, article1.Id, result[0].Article.Id)
			assert.Equal(t, author.Username, result[0].Author.Username)
			assert.True(t, result[0].IsFavorited) // viewer has favorited this article
			assert.True(t, result[0].IsFollowing)

			assert.Equal(t, article2.Id, result[1].Article.Id)
			assert.Equal(t, author.Username, result[1].Author.Username)
			assert.False(t, result[1].IsFavorited)
			assert.True(t, result[1].IsFollowing)
		})

	})
}

func TestListArticleByFavorited(t *testing.T) {
	var (
		nextPageTokenRequest  *string = nil
		nextPageTokenResponse *string = nil
	)

	t.Run("user not found", func(t *testing.T) {
		WithTestContext(t, func(tc articleTestContext) {
			// Setup test data
			favoritedByUser := generator.GenerateUser()

			// Setup expectations
			tc.mockUserService.EXPECT().
				GetUserByUsername(mock.Anything, favoritedByUser.Username).
				Return(domain.User{}, errutil.ErrUserNotFound)

			// Execute
			result, _, err := tc.articleListService.GetMostRecentArticlesFavoritedByUser(ctx, nil, favoritedByUser.Username, limit, nextPageTokenRequest)

			// Assert
			assert.ErrorIs(t, err, errutil.ErrUserNotFound)
			assert.Empty(t, result)
		})
	})

	t.Run("no auth", func(t *testing.T) {
		WithTestContext(t, func(tc articleTestContext) {
			// Setup test data
			author1 := generator.GenerateUser()
			author2 := generator.GenerateUser()

			author1Article1 := generator.GenerateArticle()
			author1Article1.AuthorId = author1.Id

			author2Article1 := generator.GenerateArticle()
			author2Article1.AuthorId = author2.Id

			favoritedByUser := generator.GenerateUser()

			// Setup expectations
			tc.mockUserService.EXPECT().
				GetUserByUsername(mock.Anything, favoritedByUser.Username).
				Return(favoritedByUser, nil)

			tc.mockUserService.EXPECT().
				GetUserListByUserIDs(mock.Anything, []uuid.UUID{author1.Id, author2.Id}).
				Return([]domain.User{author1, author2}, nil)

			tc.mockArticleRepo.EXPECT().
				FindArticlesFavoritedByUser(mock.Anything, favoritedByUser.Id, limit, nextPageTokenRequest).
				Return([]uuid.UUID{author1Article1.Id, author2Article1.Id}, nextPageTokenResponse, nil)

			tc.mockArticleRepo.EXPECT().
				FindArticlesByIds(mock.Anything, []uuid.UUID{author1Article1.Id, author2Article1.Id}).
				Return([]domain.Article{author1Article1, author2Article1}, nil)

			// Execute
			result, nextToken, err := tc.articleListService.GetMostRecentArticlesFavoritedByUser(ctx, nil, favoritedByUser.Username, limit, nextPageTokenRequest)

			// Assert
			assert.NoError(t, err)
			assert.Len(t, result, 2)
			assert.Equal(t, nextPageTokenResponse, nextToken)

			// Verify specific item details
			assert.Equal(t, author1Article1.Id, result[0].Article.Id)
			assert.Equal(t, author1.Username, result[0].Author.Username)
			assert.False(t, result[0].IsFavorited)
			assert.False(t, result[0].IsFollowing)

			assert.Equal(t, author2Article1.Id, result[1].Article.Id)
			assert.Equal(t, author2.Username, result[1].Author.Username)
			assert.False(t, result[1].IsFavorited)
			assert.False(t, result[1].IsFollowing)
		})

	})

	t.Run("viewer without following author and favorited article", func(t *testing.T) {
		WithTestContext(t, func(tc articleTestContext) {
			// Setup test data
			author1 := generator.GenerateUser()
			author2 := generator.GenerateUser()

			author1Article1 := generator.GenerateArticle()
			author1Article1.AuthorId = author1.Id

			author2Article1 := generator.GenerateArticle()
			author2Article1.AuthorId = author2.Id

			favoritedByUser := generator.GenerateUser()
			viewer := generator.GenerateUser()

			// Setup expectations
			tc.mockUserService.EXPECT().
				GetUserByUsername(mock.Anything, favoritedByUser.Username).
				Return(favoritedByUser, nil)

			tc.mockUserService.EXPECT().
				GetUserListByUserIDs(mock.Anything, []uuid.UUID{author1.Id, author2.Id}).
				Return([]domain.User{author1, author2}, nil)

			tc.mockArticleRepo.EXPECT().
				FindArticlesFavoritedByUser(mock.Anything, favoritedByUser.Id, limit, nextPageTokenRequest).
				Return([]uuid.UUID{author1Article1.Id, author2Article1.Id}, nextPageTokenResponse, nil)

			tc.mockArticleRepo.EXPECT().
				FindArticlesByIds(mock.Anything, []uuid.UUID{author1Article1.Id, author2Article1.Id}).
				Return([]domain.Article{author1Article1, author2Article1}, nil)

			tc.mockProfileService.EXPECT().
				IsFollowingBulk(mock.Anything, viewer.Id, []uuid.UUID{author1.Id, author2.Id}).
				Return(mapset.NewSetWithSize[uuid.UUID](0), nil)

			tc.mockArticleRepo.EXPECT().
				IsFavoritedBulk(mock.Anything, viewer.Id, []uuid.UUID{author1Article1.Id, author2Article1.Id}).
				Return(mapset.NewSetWithSize[uuid.UUID](0), nil)

			// Execute
			result, _, err := tc.articleListService.GetMostRecentArticlesFavoritedByUser(ctx, &viewer.Id, favoritedByUser.Username, limit, nextPageTokenRequest)

			// Assert
			assert.NoError(t, err)
			assert.Len(t, result, 2)

			// Verify specific item details
			assert.Equal(t, author1Article1.Id, result[0].Article.Id)
			assert.Equal(t, author1.Username, result[0].Author.Username)
			assert.False(t, result[0].IsFavorited)
			assert.False(t, result[0].IsFollowing)

			assert.Equal(t, author2Article1.Id, result[1].Article.Id)
			assert.Equal(t, author2.Username, result[1].Author.Username)
			assert.False(t, result[1].IsFavorited)
		})

	})

	t.Run("viewer with following author and favorited article", func(t *testing.T) {
		WithTestContext(t, func(tc articleTestContext) {
			// Setup test data
			author1 := generator.GenerateUser()
			author2 := generator.GenerateUser()

			author1Article1 := generator.GenerateArticle()
			author1Article1.AuthorId = author1.Id

			author2Article1 := generator.GenerateArticle()
			author2Article1.AuthorId = author2.Id

			favoritedByUser := generator.GenerateUser()
			viewer := generator.GenerateUser()

			// Setup expectations
			tc.mockUserService.EXPECT().
				GetUserByUsername(mock.Anything, favoritedByUser.Username).
				Return(favoritedByUser, nil)

			tc.mockUserService.EXPECT().
				GetUserListByUserIDs(mock.Anything, []uuid.UUID{author1.Id, author2.Id}).
				Return([]domain.User{author1, author2}, nil)

			tc.mockArticleRepo.EXPECT().
				FindArticlesFavoritedByUser(mock.Anything, favoritedByUser.Id, limit, nextPageTokenRequest).
				Return([]uuid.UUID{author1Article1.Id, author2Article1.Id}, nextPageTokenResponse, nil)

			tc.mockArticleRepo.EXPECT().
				FindArticlesByIds(mock.Anything, []uuid.UUID{author1Article1.Id, author2Article1.Id}).
				Return([]domain.Article{author1Article1, author2Article1}, nil)

			tc.mockProfileService.EXPECT().
				IsFollowingBulk(mock.Anything, viewer.Id, []uuid.UUID{author1.Id, author2.Id}).
				Return(mapset.NewSet[uuid.UUID](author1.Id, author2.Id), nil)

			tc.mockArticleRepo.EXPECT().
				IsFavoritedBulk(mock.Anything, viewer.Id, []uuid.UUID{author1Article1.Id, author2Article1.Id}).
				Return(mapset.NewSet[uuid.UUID](author1Article1.Id), nil)

			// Execute
			result, _, err := tc.articleListService.GetMostRecentArticlesFavoritedByUser(ctx, &viewer.Id, favoritedByUser.Username, limit, nextPageTokenRequest)

			// Assert
			assert.NoError(t, err)
			assert.Len(t, result, 2)

			// Verify specific item details
			assert.Equal(t, author1Article1.Id, result[0].Article.Id)
			assert.Equal(t, author1.Username, result[0].Author.Username)
			assert.True(t, result[0].IsFavorited) // viewer has favorited this article
			assert.True(t, result[0].IsFollowing)

			assert.Equal(t, author2Article1.Id, result[1].Article.Id)
			assert.Equal(t, author2.Username, result[1].Author.Username)
			assert.False(t, result[1].IsFavorited)
			assert.True(t, result[1].IsFollowing)
		})
	})

}

func TestListAaticleByTag(t *testing.T) {
	var (
		nextPageTokenRequest  *string = nil
		nextPageTokenResponse *string = nil
	)

	t.Run("no auth", func(t *testing.T) {
		WithTestContext(t, func(tc articleTestContext) {
			tag := gofakeit.Word()

			author1 := generator.GenerateUser()
			author2 := generator.GenerateUser()

			author1Article1 := generator.GenerateArticle()
			author1Article1.AuthorId = author1.Id

			author2Article1 := generator.GenerateArticle()
			author2Article1.AuthorId = author2.Id

			// Setup expectations
			tc.mockArticleOpensearchRepo.EXPECT().
				FindArticlesByTag(mock.Anything, tag, limit, nextPageTokenRequest).
				Return([]domain.Article{author1Article1, author2Article1}, nextPageTokenResponse, nil)

			tc.mockUserService.EXPECT().
				GetUserListByUserIDs(mock.Anything, []uuid.UUID{author1.Id, author2.Id}).
				Return([]domain.User{author1, author2}, nil)

			// Execute
			result, nextToken, err := tc.articleListService.GetMostRecentArticlesFavoritedByTag(ctx, nil, tag, limit, nextPageTokenRequest)

			// Assert
			assert.NoError(t, err)
			assert.Len(t, result, 2)
			assert.Equal(t, nextPageTokenResponse, nextToken)

			assert.Equal(t, author1Article1.Id, result[0].Article.Id)
			assert.Equal(t, author1.Username, result[0].Author.Username)
			assert.False(t, result[0].IsFavorited)
			assert.False(t, result[0].IsFollowing)

			assert.Equal(t, author2Article1.Id, result[1].Article.Id)
			assert.Equal(t, author2.Username, result[1].Author.Username)
			assert.False(t, result[1].IsFavorited)
			assert.False(t, result[1].IsFollowing)
		})
	})

	t.Run("viewer without following author and favorited article", func(t *testing.T) {
		WithTestContext(t, func(tc articleTestContext) {
			tag := gofakeit.Word()

			author1 := generator.GenerateUser()
			author2 := generator.GenerateUser()

			author1Article1 := generator.GenerateArticle()
			author1Article1.AuthorId = author1.Id

			author2Article1 := generator.GenerateArticle()
			author2Article1.AuthorId = author2.Id

			viewer := generator.GenerateUser()

			// Setup expectations
			tc.mockArticleOpensearchRepo.EXPECT().
				FindArticlesByTag(mock.Anything, tag, limit, nextPageTokenRequest).
				Return([]domain.Article{author1Article1, author2Article1}, nextPageTokenResponse, nil)

			tc.mockUserService.EXPECT().
				GetUserListByUserIDs(mock.Anything, []uuid.UUID{author1.Id, author2.Id}).
				Return([]domain.User{author1, author2}, nil)

			tc.mockProfileService.EXPECT().
				IsFollowingBulk(mock.Anything, viewer.Id, []uuid.UUID{author1.Id, author2.Id}).
				Return(mapset.NewSetWithSize[uuid.UUID](0), nil)

			tc.mockArticleRepo.EXPECT().
				IsFavoritedBulk(mock.Anything, viewer.Id, []uuid.UUID{author1Article1.Id, author2Article1.Id}).
				Return(mapset.NewSetWithSize[uuid.UUID](0), nil)

			// Execute
			result, _, err := tc.articleListService.GetMostRecentArticlesFavoritedByTag(ctx, &viewer.Id, tag, limit, nextPageTokenRequest)

			// Assert
			assert.NoError(t, err)
			assert.Len(t, result, 2)

			assert.Equal(t, author1Article1.Id, result[0].Article.Id)
			assert.Equal(t, author1.Username, result[0].Author.Username)
			assert.False(t, result[0].IsFavorited)
			assert.False(t, result[0].IsFollowing)

			assert.Equal(t, author2Article1.Id, result[1].Article.Id)
			assert.Equal(t, author2.Username, result[1].Author.Username)
			assert.False(t, result[1].IsFavorited)
			assert.False(t, result[1].IsFollowing)
		})
	})

	t.Run("viewer with following author and favorited article", func(t *testing.T) {
		WithTestContext(t, func(tc articleTestContext) {
			tag := gofakeit.Word()

			author1 := generator.GenerateUser()
			author2 := generator.GenerateUser()

			author1Article1 := generator.GenerateArticle()
			author1Article1.AuthorId = author1.Id

			author2Article1 := generator.GenerateArticle()
			author2Article1.AuthorId = author2.Id

			viewer := generator.GenerateUser()

			// Setup expectations
			tc.mockArticleOpensearchRepo.EXPECT().
				FindArticlesByTag(mock.Anything, tag, limit, nextPageTokenRequest).
				Return([]domain.Article{author1Article1, author2Article1}, nextPageTokenResponse, nil)

			tc.mockUserService.EXPECT().
				GetUserListByUserIDs(mock.Anything, []uuid.UUID{author1.Id, author2.Id}).
				Return([]domain.User{author1, author2}, nil)

			tc.mockProfileService.EXPECT().
				IsFollowingBulk(mock.Anything, viewer.Id, []uuid.UUID{author1.Id, author2.Id}).
				Return(mapset.NewSet[uuid.UUID](author1.Id, author2.Id), nil)

			tc.mockArticleRepo.EXPECT().
				IsFavoritedBulk(mock.Anything, viewer.Id, []uuid.UUID{author1Article1.Id, author2Article1.Id}).
				Return(mapset.NewSet[uuid.UUID](author1Article1.Id), nil)

			// Execute
			result, _, err := tc.articleListService.GetMostRecentArticlesFavoritedByTag(ctx, &viewer.Id, tag, limit, nextPageTokenRequest)

			// Assert
			assert.NoError(t, err)
			assert.Len(t, result, 2)

			assert.Equal(t, author1Article1.Id, result[0].Article.Id)
			assert.Equal(t, author1.Username, result[0].Author.Username)
			assert.True(t, result[0].IsFavorited)
			assert.True(t, result[0].IsFollowing)

			assert.Equal(t, author2Article1.Id, result[1].Article.Id)
			assert.Equal(t, author2.Username, result[1].Author.Username)
			assert.False(t, result[1].IsFavorited)
			assert.True(t, result[1].IsFollowing)
		})

	})
}

// - - - - - - - - - - - - - - - - Test Context - - - - - - - - - - - - - - - -

type articleTestContext struct {
	articleListService        ArticleListServiceInterface
	mockArticleRepo           *rmocks.MockArticleRepositoryInterface
	mockArticleOpensearchRepo *rmocks.MockArticleOpensearchRepositoryInterface
	mockProfileService        *mocks.MockProfileServiceInterface
	mockUserService           *mocks.MockUserServiceInterface
}

func createTestContext(t *testing.T) articleTestContext {
	mockArticleRepo := rmocks.NewMockArticleRepositoryInterface(t)
	mockArticleOpensearchRepo := rmocks.NewMockArticleOpensearchRepositoryInterface(t)
	mockProfileService := mocks.NewMockProfileServiceInterface(t)
	mockUserService := mocks.NewMockUserServiceInterface(t)
	articleListService := articleListService{
		articleRepository:           mockArticleRepo,
		articleOpensearchRepository: mockArticleOpensearchRepo,
		profileService:              mockProfileService,
		userService:                 mockUserService,
	}
	return articleTestContext{
		articleListService:        articleListService,
		mockArticleRepo:           mockArticleRepo,
		mockArticleOpensearchRepo: mockArticleOpensearchRepo,
		mockProfileService:        mockProfileService,
		mockUserService:           mockUserService,
	}
}

func WithTestContext(t *testing.T, testFunc func(tc articleTestContext)) {
	testFunc(createTestContext(t))
}

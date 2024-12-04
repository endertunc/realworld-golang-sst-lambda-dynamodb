package service

import (
	"context"
	"errors"
	"testing"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/generator"
	repoMocks "realworld-aws-lambda-dynamodb-golang/internal/repository/mocks"
	serviceMocks "realworld-aws-lambda-dynamodb-golang/internal/service/mocks"
)

var errInternal = errors.New("internal error")
const defaultLimit = 10

func TestFeedService_FetchArticlesFromFeed(t *testing.T) {
	ctx := context.Background()

	t.Run("followed author's article should appear in feed with following flag true", func(t *testing.T) {
		withFeedTestContext(t, func(tc feedTestContext) {
			// Setup test data
			feedUser := generator.GenerateUser()
			author := generator.GenerateUser()
			article := generator.GenerateArticle()
			article.AuthorId = author.Id
			var nextPageToken *string

			// Setup expectations
			tc.mockUserFeedRepo.EXPECT().
				FindArticleIdsInUserFeed(ctx, feedUser.Id, defaultLimit, nextPageToken).
				Return([]uuid.UUID{article.Id}, nextPageToken, nil)

			tc.mockArticleService.EXPECT().
				GetArticlesByIds(ctx, []uuid.UUID{article.Id}).
				Return([]domain.Article{article}, nil)

			tc.mockUserService.EXPECT().
				GetUserListByUserIDs(ctx, []uuid.UUID{author.Id}).
				Return([]domain.User{author}, nil)

			tc.mockProfileService.EXPECT().
				IsFollowingBulk(ctx, feedUser.Id, []uuid.UUID{author.Id}).
				Return(mapset.NewSet[uuid.UUID](author.Id), nil)

			tc.mockArticleService.EXPECT().
				IsFavoritedBulk(ctx, feedUser.Id, []uuid.UUID{article.Id}).
				Return(mapset.NewSet[uuid.UUID](), nil)

			// Execute
			feedItems, nextToken, err := tc.feedService.FetchArticlesFromFeed(ctx, feedUser.Id, defaultLimit, nextPageToken)

			// Assert
			assert.NoError(t, err)
			assert.Len(t, feedItems, 1)
			assert.Equal(t, nextPageToken, nextToken)
			assert.True(t, feedItems[0].IsFollowing)
			assert.False(t, feedItems[0].IsFavorited)
		})
	})

	t.Run("unfollowed author's article should not appear in feed", func(t *testing.T) {
		withFeedTestContext(t, func(tc feedTestContext) {
			// Setup test data
			feedUser := generator.GenerateUser()
			author := generator.GenerateUser()
			article := generator.GenerateArticle()
			article.AuthorId = author.Id
			var nextPageToken *string

			// Setup expectations
			tc.mockUserFeedRepo.EXPECT().
				FindArticleIdsInUserFeed(ctx, feedUser.Id, defaultLimit, nextPageToken).
				Return([]uuid.UUID{article.Id}, nextPageToken, nil)

			tc.mockArticleService.EXPECT().
				GetArticlesByIds(ctx, []uuid.UUID{article.Id}).
				Return([]domain.Article{article}, nil)

			tc.mockUserService.EXPECT().
				GetUserListByUserIDs(ctx, []uuid.UUID{author.Id}).
				Return([]domain.User{author}, nil)

			tc.mockProfileService.EXPECT().
				IsFollowingBulk(ctx, feedUser.Id, []uuid.UUID{author.Id}).
				Return(mapset.NewSet[uuid.UUID](), nil)

			tc.mockArticleService.EXPECT().
				IsFavoritedBulk(ctx, feedUser.Id, []uuid.UUID{article.Id}).
				Return(mapset.NewSet[uuid.UUID](), nil)

			// Execute
			feedItems, nextToken, err := tc.feedService.FetchArticlesFromFeed(ctx, feedUser.Id, defaultLimit, nextPageToken)

			// Assert
			assert.NoError(t, err)
			assert.Empty(t, feedItems)
			assert.Equal(t, nextPageToken, nextToken)
		})
	})

	t.Run("followed author's article should have favorited flag true when favorited", func(t *testing.T) {
		withFeedTestContext(t, func(tc feedTestContext) {
			// Setup test data
			feedUser := generator.GenerateUser()
			author := generator.GenerateUser()
			article := generator.GenerateArticle()
			article.AuthorId = author.Id
			var nextPageToken *string

			// Setup expectations
			tc.mockUserFeedRepo.EXPECT().
				FindArticleIdsInUserFeed(ctx, feedUser.Id, defaultLimit, nextPageToken).
				Return([]uuid.UUID{article.Id}, nextPageToken, nil)

			tc.mockArticleService.EXPECT().
				GetArticlesByIds(ctx, []uuid.UUID{article.Id}).
				Return([]domain.Article{article}, nil)

			tc.mockUserService.EXPECT().
				GetUserListByUserIDs(ctx, []uuid.UUID{author.Id}).
				Return([]domain.User{author}, nil)

			tc.mockProfileService.EXPECT().
				IsFollowingBulk(ctx, feedUser.Id, []uuid.UUID{author.Id}).
				Return(mapset.NewSet[uuid.UUID](author.Id), nil)

			tc.mockArticleService.EXPECT().
				IsFavoritedBulk(ctx, feedUser.Id, []uuid.UUID{article.Id}).
				Return(mapset.NewSet[uuid.UUID](article.Id), nil)

			// Execute
			feedItems, nextToken, err := tc.feedService.FetchArticlesFromFeed(ctx, feedUser.Id, defaultLimit, nextPageToken)

			// Assert
			assert.NoError(t, err)
			assert.Len(t, feedItems, 1)
			assert.Equal(t, nextPageToken, nextToken)
			assert.True(t, feedItems[0].IsFollowing)
			assert.True(t, feedItems[0].IsFavorited)
		})
	})

	t.Run("followed author's article should have favorited flag false when not favorited", func(t *testing.T) {
		withFeedTestContext(t, func(tc feedTestContext) {
			// Setup test data
			feedUser := generator.GenerateUser()
			author := generator.GenerateUser()
			article := generator.GenerateArticle()
			article.AuthorId = author.Id
			var nextPageToken *string

			// Setup expectations
			tc.mockUserFeedRepo.EXPECT().
				FindArticleIdsInUserFeed(ctx, feedUser.Id, defaultLimit, nextPageToken).
				Return([]uuid.UUID{article.Id}, nextPageToken, nil)

			tc.mockArticleService.EXPECT().
				GetArticlesByIds(ctx, []uuid.UUID{article.Id}).
				Return([]domain.Article{article}, nil)

			tc.mockUserService.EXPECT().
				GetUserListByUserIDs(ctx, []uuid.UUID{author.Id}).
				Return([]domain.User{author}, nil)

			tc.mockProfileService.EXPECT().
				IsFollowingBulk(ctx, feedUser.Id, []uuid.UUID{author.Id}).
				Return(mapset.NewSet[uuid.UUID](author.Id), nil)

			tc.mockArticleService.EXPECT().
				IsFavoritedBulk(ctx, feedUser.Id, []uuid.UUID{article.Id}).
				Return(mapset.NewSet[uuid.UUID](), nil)

			// Execute
			feedItems, nextToken, err := tc.feedService.FetchArticlesFromFeed(ctx, feedUser.Id, defaultLimit, nextPageToken)

			// Assert
			assert.NoError(t, err)
			assert.Len(t, feedItems, 1)
			assert.Equal(t, nextPageToken, nextToken)
			assert.True(t, feedItems[0].IsFollowing)
			assert.False(t, feedItems[0].IsFavorited)
		})
	})

	t.Run("empty feed", func(t *testing.T) {
		withFeedTestContext(t, func(tc feedTestContext) {
			// Setup test data
			feedUser := generator.GenerateUser()
			var nextPageToken *string

			// Setup expectations
			tc.mockUserFeedRepo.EXPECT().
				FindArticleIdsInUserFeed(ctx, feedUser.Id, defaultLimit, nextPageToken).
				Return([]uuid.UUID{}, nextPageToken, nil)

			// Execute
			feedItems, nextToken, err := tc.feedService.FetchArticlesFromFeed(ctx, feedUser.Id, defaultLimit, nextPageToken)

			// Assert
			assert.NoError(t, err)
			assert.Empty(t, feedItems)
			assert.Equal(t, nextPageToken, nextToken)
		})
	})

	// Error cases
	t.Run("error from FindArticleIdsInUserFeed", func(t *testing.T) {
		withFeedTestContext(t, func(tc feedTestContext) {
			// Setup test data
			feedUser := generator.GenerateUser()
			var nextPageToken *string

			// Setup expectations
			tc.mockUserFeedRepo.EXPECT().
				FindArticleIdsInUserFeed(ctx, feedUser.Id, defaultLimit, nextPageToken).
				Return(nil, nil, errInternal)

			// Execute
			feedItems, nextToken, err := tc.feedService.FetchArticlesFromFeed(ctx, feedUser.Id, defaultLimit, nextPageToken)

			// Assert
			assert.ErrorIs(t, err, errInternal)
			assert.Empty(t, feedItems)
			assert.Nil(t, nextToken)
		})
	})

	t.Run("error from GetArticlesByIds", func(t *testing.T) {
		withFeedTestContext(t, func(tc feedTestContext) {
			// Setup test data
			feedUser := generator.GenerateUser()
			article := generator.GenerateArticle()
			var nextPageToken *string

			// Setup expectations
			tc.mockUserFeedRepo.EXPECT().
				FindArticleIdsInUserFeed(ctx, feedUser.Id, defaultLimit, nextPageToken).
				Return([]uuid.UUID{article.Id}, nextPageToken, nil)

			tc.mockArticleService.EXPECT().
				GetArticlesByIds(ctx, []uuid.UUID{article.Id}).
				Return(nil, errInternal)

			// Execute
			feedItems, nextToken, err := tc.feedService.FetchArticlesFromFeed(ctx, feedUser.Id, defaultLimit, nextPageToken)

			// Assert
			assert.ErrorIs(t, err, errInternal)
			assert.Empty(t, feedItems)
			assert.Nil(t, nextToken)
		})
	})

	t.Run("error from GetUserListByUserIDs", func(t *testing.T) {
		withFeedTestContext(t, func(tc feedTestContext) {
			// Setup test data
			feedUser := generator.GenerateUser()
			article := generator.GenerateArticle()
			var nextPageToken *string

			// Setup expectations
			tc.mockUserFeedRepo.EXPECT().
				FindArticleIdsInUserFeed(ctx, feedUser.Id, defaultLimit, nextPageToken).
				Return([]uuid.UUID{article.Id}, nextPageToken, nil)

			tc.mockArticleService.EXPECT().
				GetArticlesByIds(ctx, []uuid.UUID{article.Id}).
				Return([]domain.Article{article}, nil)

			tc.mockUserService.EXPECT().
				GetUserListByUserIDs(ctx, []uuid.UUID{article.AuthorId}).
				Return(nil, errInternal)

			// Execute
			feedItems, nextToken, err := tc.feedService.FetchArticlesFromFeed(ctx, feedUser.Id, defaultLimit, nextPageToken)

			// Assert
			assert.ErrorIs(t, err, errInternal)
			assert.Empty(t, feedItems)
			assert.Nil(t, nextToken)
		})
	})

	t.Run("error from IsFollowingBulk", func(t *testing.T) {
		withFeedTestContext(t, func(tc feedTestContext) {
			// Setup test data
			feedUser := generator.GenerateUser()
			author := generator.GenerateUser()
			article := generator.GenerateArticle()
			article.AuthorId = author.Id
			var nextPageToken *string

			// Setup expectations
			tc.mockUserFeedRepo.EXPECT().
				FindArticleIdsInUserFeed(ctx, feedUser.Id, defaultLimit, nextPageToken).
				Return([]uuid.UUID{article.Id}, nextPageToken, nil)

			tc.mockArticleService.EXPECT().
				GetArticlesByIds(ctx, []uuid.UUID{article.Id}).
				Return([]domain.Article{article}, nil)

			tc.mockUserService.EXPECT().
				GetUserListByUserIDs(ctx, []uuid.UUID{author.Id}).
				Return([]domain.User{author}, nil)

			tc.mockProfileService.EXPECT().
				IsFollowingBulk(ctx, feedUser.Id, []uuid.UUID{author.Id}).
				Return(nil, errInternal)

			// Execute
			feedItems, nextToken, err := tc.feedService.FetchArticlesFromFeed(ctx, feedUser.Id, defaultLimit, nextPageToken)

			// Assert
			assert.ErrorIs(t, err, errInternal)
			assert.Empty(t, feedItems)
			assert.Nil(t, nextToken)
		})
	})

	t.Run("error from IsFavoritedBulk", func(t *testing.T) {
		withFeedTestContext(t, func(tc feedTestContext) {
			// Setup test data
			feedUser := generator.GenerateUser()
			author := generator.GenerateUser()
			article := generator.GenerateArticle()
			article.AuthorId = author.Id
			var nextPageToken *string

			// Setup expectations
			tc.mockUserFeedRepo.EXPECT().
				FindArticleIdsInUserFeed(ctx, feedUser.Id, defaultLimit, nextPageToken).
				Return([]uuid.UUID{article.Id}, nextPageToken, nil)

			tc.mockArticleService.EXPECT().
				GetArticlesByIds(ctx, []uuid.UUID{article.Id}).
				Return([]domain.Article{article}, nil)

			tc.mockUserService.EXPECT().
				GetUserListByUserIDs(ctx, []uuid.UUID{author.Id}).
				Return([]domain.User{author}, nil)

			tc.mockProfileService.EXPECT().
				IsFollowingBulk(ctx, feedUser.Id, []uuid.UUID{author.Id}).
				Return(mapset.NewSet(author.Id), nil)

			tc.mockArticleService.EXPECT().
				IsFavoritedBulk(ctx, feedUser.Id, []uuid.UUID{article.Id}).
				Return(nil, errInternal)

			// Execute
			feedItems, nextToken, err := tc.feedService.FetchArticlesFromFeed(ctx, feedUser.Id, defaultLimit, nextPageToken)

			// Assert
			assert.ErrorIs(t, err, errInternal)
			assert.Empty(t, feedItems)
			assert.Nil(t, nextToken)
		})
	})

	t.Run("with valid nextPageToken", func(t *testing.T) {
		withFeedTestContext(t, func(tc feedTestContext) {
			// Setup test data
			feedUser := generator.GenerateUser()
			author := generator.GenerateUser()
			article := generator.GenerateArticle()
			article.AuthorId = author.Id
			nextPageToken := uuid.New().String()
			newNextPageToken := uuid.New().String()

			// Setup expectations
			tc.mockUserFeedRepo.EXPECT().
				FindArticleIdsInUserFeed(ctx, feedUser.Id, defaultLimit, &nextPageToken).
				Return([]uuid.UUID{article.Id}, &newNextPageToken, nil)

			tc.mockArticleService.EXPECT().
				GetArticlesByIds(ctx, []uuid.UUID{article.Id}).
				Return([]domain.Article{article}, nil)

			tc.mockUserService.EXPECT().
				GetUserListByUserIDs(ctx, []uuid.UUID{author.Id}).
				Return([]domain.User{author}, nil)

			tc.mockProfileService.EXPECT().
				IsFollowingBulk(ctx, feedUser.Id, []uuid.UUID{author.Id}).
				Return(mapset.NewSet[uuid.UUID](author.Id), nil)

			tc.mockArticleService.EXPECT().
				IsFavoritedBulk(ctx, feedUser.Id, []uuid.UUID{article.Id}).
				Return(mapset.NewSet[uuid.UUID](article.Id), nil)

			// Execute
			feedItems, nextToken, err := tc.feedService.FetchArticlesFromFeed(ctx, feedUser.Id, defaultLimit, &nextPageToken)

			// Assert
			assert.NoError(t, err)
			assert.Len(t, feedItems, 1)
			assert.Equal(t, &newNextPageToken, nextToken)
			assert.Equal(t, article.Id, feedItems[0].Article.Id)
			assert.Equal(t, author.Username, feedItems[0].Author.Username)
			assert.True(t, feedItems[0].IsFavorited)
			assert.True(t, feedItems[0].IsFollowing)
		})
	})
}

func TestFeedService_FanoutArticle(t *testing.T) {
	ctx := context.Background()

	t.Run("successful fanout", func(t *testing.T) {
		withFeedTestContext(t, func(tc feedTestContext) {
			// Setup test data
			article := generator.GenerateArticle()

			// Setup expectations
			tc.mockUserFeedRepo.EXPECT().
				FanoutArticle(ctx, article.Id, article.AuthorId, article.CreatedAt).
				Return(nil)

			// Execute
			err := tc.feedService.FanoutArticle(ctx, article.Id, article.AuthorId, article.CreatedAt)

			// Assert
			assert.NoError(t, err)
		})
	})

	t.Run("error from repository", func(t *testing.T) {
		withFeedTestContext(t, func(tc feedTestContext) {
			// Setup test data
			article := generator.GenerateArticle()

			// Setup expectations
			tc.mockUserFeedRepo.EXPECT().
				FanoutArticle(ctx, article.Id, article.AuthorId, article.CreatedAt).
				Return(errInternal)

			// Execute
			err := tc.feedService.FanoutArticle(ctx, article.Id, article.AuthorId, article.CreatedAt)

			// Assert
			assert.ErrorIs(t, err, errInternal)
		})
	})
}

type feedTestContext struct {
	feedService        FeedServiceInterface
	mockUserFeedRepo   *repoMocks.MockUserFeedRepositoryInterface
	mockArticleService *serviceMocks.MockArticleServiceInterface
	mockProfileService *serviceMocks.MockProfileServiceInterface
	mockUserService    *serviceMocks.MockUserServiceInterface
}

func createFeedTestContext(t *testing.T) feedTestContext {
	mockUserFeedRepo := repoMocks.NewMockUserFeedRepositoryInterface(t)
	mockArticleService := serviceMocks.NewMockArticleServiceInterface(t)
	mockProfileService := serviceMocks.NewMockProfileServiceInterface(t)
	mockUserService := serviceMocks.NewMockUserServiceInterface(t)

	feedService := NewUserFeedService(
		mockUserFeedRepo,
		mockArticleService,
		mockProfileService,
		mockUserService,
	)

	return feedTestContext{
		feedService:        feedService,
		mockUserFeedRepo:   mockUserFeedRepo,
		mockArticleService: mockArticleService,
		mockProfileService: mockProfileService,
		mockUserService:    mockUserService,
	}
}

func withFeedTestContext(t *testing.T, testFunc func(tc feedTestContext)) {
	testFunc(createFeedTestContext(t))
}
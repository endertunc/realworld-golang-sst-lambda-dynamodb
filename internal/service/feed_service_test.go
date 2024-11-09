package service

import (
	"context"
	mapset "github.com/deckarep/golang-set/v2"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/generator"
	rmocks "realworld-aws-lambda-dynamodb-golang/internal/repository/mocks"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/service/mocks"
)

func TestFetchArticlesFromFeed_SuccessfulMultipleArticles(t *testing.T) {
	// Create a mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock dependencies
	mockUserFeedRepo := rmocks.NewMockUserFeedRepositoryInterface(ctrl)
	mockArticleService := mocks.NewMockArticleServiceInterface(ctrl)
	mockProfileService := mocks.NewMockProfileServiceInterface(ctrl)
	mockUserService := mocks.NewMockUserServiceInterface(ctrl)

	// Create a service instance with mock dependencies
	feedService := UserFeedService{
		UserFeedRepository: mockUserFeedRepo,
		ArticleService:     mockArticleService,
		ProfileService:     mockProfileService,
		UserService:        mockUserService,
	}

	// Prepare test data
	author1 := generator.GenerateUser()
	author2 := generator.GenerateUser()

	article1 := generator.GenerateArticle()
	article1.AuthorId = author1.Id

	article2 := generator.GenerateArticle()
	article2.AuthorId = author2.Id

	feedUser := generator.GenerateUser()

	limit := 10
	var nextPageToken *string

	// Set up expectations for UserFeedRepository
	mockUserFeedRepo.EXPECT().
		FindArticleIdsInUserFeed(gomock.Any(), feedUser.Id, limit, nextPageToken).
		Return([]uuid.UUID{article1.Id, article2.Id}, nextPageToken, nil)

	// Set up expectations for ArticleService
	mockArticleService.EXPECT().
		FindArticlesByIds(gomock.Any(), []uuid.UUID{article1.Id, article2.Id}).
		Return([]domain.Article{article1, article2}, nil)

	// Set up expectations for UserService
	mockUserService.EXPECT().
		GetUserListByUserIDs(gomock.Any(), []uuid.UUID{author1.Id, author2.Id}).
		Return([]domain.User{author1, author2}, nil)

	// Set up expectations for ProfileService
	mockProfileService.EXPECT().
		IsFollowingBulk(gomock.Any(), feedUser.Id, []uuid.UUID{author1.Id, author2.Id}).
		Return(mapset.NewSet(author1.Id, author2.Id), nil)

	// Set up expectations for ArticleService
	mockArticleService.EXPECT().
		IsFavoritedBulk(gomock.Any(), feedUser.Id, []uuid.UUID{article1.Id, article2.Id}).
		Return(mapset.NewSet(article1.Id), nil)

	// Call the method under test
	feedItems, nextToken, err := feedService.FetchArticlesFromFeed(context.Background(), feedUser.Id, limit, nextPageToken)

	// Assertions
	assert.NoError(t, err)
	assert.Len(t, feedItems, 2)
	assert.Equal(t, nextPageToken, nextToken)

	// Verify specific feed item details
	assert.Equal(t, article1.Id, feedItems[0].Article.Id)
	assert.Equal(t, author1.Username, feedItems[0].Author.Username)
	assert.True(t, feedItems[0].IsFavorited)
	assert.True(t, feedItems[0].IsFollowing)

	assert.Equal(t, article2.Id, feedItems[1].Article.Id)
	assert.Equal(t, author2.Username, feedItems[1].Author.Username)
	assert.False(t, feedItems[1].IsFavorited)
	assert.True(t, feedItems[1].IsFollowing)
}

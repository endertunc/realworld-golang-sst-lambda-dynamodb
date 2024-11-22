package service

//import (
//	"context"
//	"github.com/stretchr/testify/mock"
//	"realworld-aws-lambda-dynamodb-golang/internal/domain/generator"
//	"realworld-aws-lambda-dynamodb-golang/internal/service/mocks"
//	"testing"
//
//	mapset "github.com/deckarep/golang-set/v2"
//
//	"github.com/google/uuid"
//	"github.com/stretchr/testify/assert"
//	"realworld-aws-lambda-dynamodb-golang/internal/domain"
//
//	rmocks "realworld-aws-lambda-dynamodb-golang/internal/repository/mocks"
//)
//
//func TestFetchArticlesFromFeed_SuccessfulMultipleArticles(t *testing.T) {
//	ctx := context.Background()
//	mockUserFeedRepo := rmocks.NewMockUserFeedRepositoryInterface(t)
//	mockArticleService := mocks.NewMockArticleServiceInterface(t)
//	mockProfileService := mocks.NewMockProfileServiceInterface(t)
//	mockUserService := mocks.NewMockUserServiceInterface(t)
//
//	// Create a service instance with mock dependencies
//	feedService := userFeedService{
//		userFeedRepository: mockUserFeedRepo,
//		articleService:     mockArticleService,
//		profileService:     mockProfileService,
//		userService:        mockUserService,
//	}
//
//	// Prepare test data
//	author1 := generator.GenerateUser()
//	author2 := generator.GenerateUser()
//
//	article1 := generator.GenerateArticle()
//	article1.AuthorId = author1.Id
//
//	article2 := generator.GenerateArticle()
//	article2.AuthorId = author2.Id
//
//	feedUser := generator.GenerateUser()
//
//	limit := 10
//	var nextPageToken *string
//
//	// Set up expectations for userFeedRepository
//	mockUserFeedRepo.EXPECT().
//		FindArticleIdsInUserFeed(mock.Anything, feedUser.Id, limit, nextPageToken).
//		Return([]uuid.UUID{article1.Id, article2.Id}, nextPageToken, nil)
//
//	// Set up expectations for articleService
//	mockArticleService.EXPECT().
//		FindArticlesByIds(mock.Anything, []uuid.UUID{article1.Id, article2.Id}).
//		Return([]domain.Article{article1, article2}, nil)
//
//	// Set up expectations for userService
//	mockUserService.EXPECT().
//		GetUserListByUserIDs(mock.Anything, []uuid.UUID{author1.Id, author2.Id}).
//		Return([]domain.User{author1, author2}, nil)
//
//	// Set up expectations for profileService
//	mockProfileService.EXPECT().
//		IsFollowingBulk(mock.Anything, feedUser.Id, []uuid.UUID{author1.Id, author2.Id}).
//		Return(mapset.NewSet(author1.Id, author2.Id), nil)
//
//	// Set up expectations for articleService
//	mockArticleService.EXPECT().
//		IsFavoritedBulk(mock.Anything, feedUser.Id, []uuid.UUID{article1.Id, article2.Id}).
//		Return(mapset.NewSet(article1.Id), nil)
//
//	// Call the method under test
//
//	feedItems, nextToken, err := feedService.FetchArticlesFromFeed(ctx, feedUser.Id, limit, nextPageToken)
//
//	// Assertions
//	assert.NoError(t, err)
//	assert.Len(t, feedItems, 2)
//	assert.Equal(t, nextPageToken, nextToken)
//
//	// Verify specific feed item details
//	assert.Equal(t, article1.Id, feedItems[0].Article.Id)
//	assert.Equal(t, author1.Username, feedItems[0].Author.Username)
//	assert.True(t, feedItems[0].IsFavorited)
//	assert.True(t, feedItems[0].IsFollowing)
//
//	assert.Equal(t, article2.Id, feedItems[1].Article.Id)
//	assert.Equal(t, author2.Username, feedItems[1].Author.Username)
//	assert.False(t, feedItems[1].IsFavorited)
//	assert.True(t, feedItems[1].IsFollowing)
//}

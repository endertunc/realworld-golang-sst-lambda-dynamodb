package service

import (
	"context"
	"testing"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"realworld-aws-lambda-dynamodb-golang/internal/domain/generator"
	"realworld-aws-lambda-dynamodb-golang/internal/repository/mocks"
)

func ptr[T any](v T) *T {
	return &v
}

func TestProfileService_GetUserProfile(t *testing.T) {
	ctx := context.Background()

	t.Run("get profile when logged in and following", func(t *testing.T) {
		withProfileTestContext(t, func(tc profileTestContext) {
			// Setup test data
			loggedInUserId := uuid.New()
			targetUser := generator.GenerateUser()

			// Setup expectations
			tc.mockUserRepo.EXPECT().
				FindUserByUsername(ctx, targetUser.Username).
				Return(targetUser, nil)
			tc.mockFollowerRepo.EXPECT().
				FindFollowees(ctx, loggedInUserId, []uuid.UUID{targetUser.Id}).
				Return(mapset.NewSet(targetUser.Id), nil)

			// Execute
			profile, following, err := tc.profileService.GetUserProfile(ctx, &loggedInUserId, targetUser.Username)

			// Assert
			assert.NoError(t, err)
			assert.True(t, following)
			assert.Equal(t, targetUser, profile)
		})
	})

	t.Run("get profile when logged in and not following", func(t *testing.T) {
		withProfileTestContext(t, func(tc profileTestContext) {
			// Setup test data
			loggedInUserId := uuid.New()
			targetUser := generator.GenerateUser()

			// Setup expectations
			tc.mockUserRepo.EXPECT().
				FindUserByUsername(ctx, targetUser.Username).
				Return(targetUser, nil)
			tc.mockFollowerRepo.EXPECT().
				FindFollowees(ctx, loggedInUserId, []uuid.UUID{targetUser.Id}).
				Return(mapset.NewSet[uuid.UUID](), nil)

			// Execute
			profile, following, err := tc.profileService.GetUserProfile(ctx, &loggedInUserId, targetUser.Username)

			// Assert
			assert.NoError(t, err)
			assert.False(t, following)
			assert.Equal(t, targetUser, profile)
		})
	})

	t.Run("get profile when not logged in", func(t *testing.T) {
		withProfileTestContext(t, func(tc profileTestContext) {
			// Setup test data
			targetUser := generator.GenerateUser()

			// Setup expectations
			tc.mockUserRepo.EXPECT().
				FindUserByUsername(ctx, targetUser.Username).
				Return(targetUser, nil)

			// Execute
			profile, following, err := tc.profileService.GetUserProfile(ctx, nil, targetUser.Username)

			// Assert
			assert.NoError(t, err)
			assert.False(t, following)
			assert.Equal(t, targetUser, profile)
		})
	})
}

func TestProfileService_Follow(t *testing.T) {
	ctx := context.Background()

	t.Run("follow user successfully", func(t *testing.T) {
		withProfileTestContext(t, func(tc profileTestContext) {
			// Setup test data
			followerUserId := uuid.New()
			targetUser := generator.GenerateUser()

			// Setup expectations
			tc.mockUserRepo.EXPECT().
				FindUserByUsername(ctx, targetUser.Username).
				Return(targetUser, nil)
			tc.mockFollowerRepo.EXPECT().
				Follow(ctx, followerUserId, targetUser.Id).
				Return(nil)

			// Execute
			profile, err := tc.profileService.Follow(ctx, followerUserId, targetUser.Username)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, targetUser, profile)
		})
	})

	t.Run("cannot follow self", func(t *testing.T) {
		withProfileTestContext(t, func(tc profileTestContext) {
			// Setup test data
			userId := uuid.New()
			user := generator.GenerateUser()
			user.Id = userId // Set same ID to test self-follow

			// Setup expectations
			tc.mockUserRepo.EXPECT().
				FindUserByUsername(ctx, user.Username).
				Return(user, nil)

			// Execute
			_, err := tc.profileService.Follow(ctx, userId, user.Username)

			// Assert
			assert.Error(t, err)
		})
	})
}

func TestProfileService_UnFollow(t *testing.T) {
	ctx := context.Background()

	t.Run("unfollow user successfully", func(t *testing.T) {
		withProfileTestContext(t, func(tc profileTestContext) {
			// Setup test data
			followerUserId := uuid.New()
			targetUser := generator.GenerateUser()

			// Setup expectations
			tc.mockUserRepo.EXPECT().
				FindUserByUsername(ctx, targetUser.Username).
				Return(targetUser, nil)
			tc.mockFollowerRepo.EXPECT().
				UnFollow(ctx, followerUserId, targetUser.Id).
				Return(nil)

			// Execute
			profile, err := tc.profileService.UnFollow(ctx, followerUserId, targetUser.Username)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, targetUser, profile)
		})
	})

	t.Run("cannot unfollow self", func(t *testing.T) {
		withProfileTestContext(t, func(tc profileTestContext) {
			// Setup test data
			userId := uuid.New()
			user := generator.GenerateUser()
			user.Id = userId // Set same ID to test self-unfollow

			// Setup expectations
			tc.mockUserRepo.EXPECT().
				FindUserByUsername(ctx, user.Username).
				Return(user, nil)

			// Execute
			_, err := tc.profileService.UnFollow(ctx, userId, user.Username)

			// Assert
			assert.Error(t, err)
		})
	})
}

func TestProfileService_IsFollowing(t *testing.T) {
	ctx := context.Background()

	t.Run("check if following - true", func(t *testing.T) {
		withProfileTestContext(t, func(tc profileTestContext) {
			// Setup test data
			followerUserId := uuid.New()
			followeeUserId := generator.GenerateUser().Id

			// Setup expectations
			tc.mockFollowerRepo.EXPECT().
				FindFollowees(ctx, followerUserId, []uuid.UUID{followeeUserId}).
				Return(mapset.NewSet(followeeUserId), nil)

			// Execute
			following, err := tc.profileService.IsFollowing(ctx, followerUserId, followeeUserId)

			// Assert
			assert.NoError(t, err)
			assert.True(t, following)
		})
	})

	t.Run("check if following - false", func(t *testing.T) {
		withProfileTestContext(t, func(tc profileTestContext) {
			// Setup test data
			followerUserId := uuid.New()
			followeeUserId := generator.GenerateUser().Id

			// Setup expectations
			tc.mockFollowerRepo.EXPECT().
				FindFollowees(ctx, followerUserId, []uuid.UUID{followeeUserId}).
				Return(mapset.NewSet[uuid.UUID](), nil)

			// Execute
			following, err := tc.profileService.IsFollowing(ctx, followerUserId, followeeUserId)

			// Assert
			assert.NoError(t, err)
			assert.False(t, following)
		})
	})
}

func TestProfileService_IsFollowingBulk(t *testing.T) {
	ctx := context.Background()

	t.Run("check multiple followees", func(t *testing.T) {
		withProfileTestContext(t, func(tc profileTestContext) {
			// Setup test data
			followerUserId := uuid.New()
			followeeIds := []uuid.UUID{generator.GenerateUser().Id, generator.GenerateUser().Id, generator.GenerateUser().Id}
			expectedFollowing := mapset.NewSet(followeeIds[0], followeeIds[1])

			// Setup expectations
			tc.mockFollowerRepo.EXPECT().
				FindFollowees(ctx, followerUserId, followeeIds).
				Return(expectedFollowing, nil)

			// Execute
			following, err := tc.profileService.IsFollowingBulk(ctx, followerUserId, followeeIds)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, expectedFollowing, following)
		})
	})
}

// - - - - - - - - - - - - - - - - Test Context - - - - - - - - - - - - - - - -

type profileTestContext struct {
	profileService   ProfileServiceInterface
	mockFollowerRepo *mocks.MockFollowerRepositoryInterface
	mockUserRepo     *mocks.MockUserRepositoryInterface
}

func createProfileTestContext(t *testing.T) profileTestContext {
	mockFollowerRepo := mocks.NewMockFollowerRepositoryInterface(t)
	mockUserRepo := mocks.NewMockUserRepositoryInterface(t)
	profileService := NewProfileService(mockFollowerRepo, mockUserRepo)

	return profileTestContext{
		profileService:   profileService,
		mockFollowerRepo: mockFollowerRepo,
		mockUserRepo:     mockUserRepo,
	}
}

func withProfileTestContext(t *testing.T, testFunc func(tc profileTestContext)) {
	testFunc(createProfileTestContext(t))
}

package repository

import (
	"context"
	"realworld-aws-lambda-dynamodb-golang/internal/database"
	"realworld-aws-lambda-dynamodb-golang/internal/test"
	"testing"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var followerRepo = NewDynamodbFollowerRepository(database.NewDynamoDBStore())

func TestFollow(t *testing.T) {
	ctx := context.Background()
	test.WithSetupAndTeardown(t, func() {
		t.Run("success", func(t *testing.T) {
			follower := uuid.New()
			followee := uuid.New()

			err := followerRepo.Follow(ctx, follower, followee)
			require.NoError(t, err)

			// Verify follow relationship exists
			followees, err := followerRepo.FindFollowees(ctx, follower, []uuid.UUID{followee})
			require.NoError(t, err)
			assert.True(t, followees.Contains(followee))
		})

		// ToDo @ender - at the moment this is fine but if we were to add createdAt,
		// then we should be careful with allowing the same follower to follow the same followee
		t.Run("follow same user twice", func(t *testing.T) {
			follower := uuid.New()
			followee := uuid.New()

			err := followerRepo.Follow(ctx, follower, followee)
			require.NoError(t, err)

			// Following same user again should not error
			err = followerRepo.Follow(ctx, follower, followee)
			require.NoError(t, err)

			// Verify follow relationship still exists
			followees, err := followerRepo.FindFollowees(ctx, follower, []uuid.UUID{followee})
			require.NoError(t, err)
			assert.True(t, followees.Contains(followee))
		})
	})
}

func TestUnFollow(t *testing.T) {
	ctx := context.Background()
	test.WithSetupAndTeardown(t, func() {
		t.Run("success", func(t *testing.T) {
			follower := uuid.New()
			followee := uuid.New()

			// First follow
			err := followerRepo.Follow(ctx, follower, followee)
			require.NoError(t, err)

			// Verify follow relationship exists
			followees, err := followerRepo.FindFollowees(ctx, follower, []uuid.UUID{followee})
			require.NoError(t, err)
			assert.True(t, followees.Contains(followee))

			// Then unfollow
			err = followerRepo.UnFollow(ctx, follower, followee)
			require.NoError(t, err)

			// Verify follow relationship no longer exists
			followees, err = followerRepo.FindFollowees(ctx, follower, []uuid.UUID{followee})
			require.NoError(t, err)
			assert.False(t, followees.Contains(followee))
		})

		t.Run("unfollow non-existent relationship", func(t *testing.T) {
			follower := uuid.New()
			followee := uuid.New()

			// Unfollowing a non-existent relationship should not error
			err := followerRepo.UnFollow(ctx, follower, followee)
			require.NoError(t, err)
		})
	})
}

func TestFindFollowees(t *testing.T) {
	ctx := context.Background()
	test.WithSetupAndTeardown(t, func() {
		t.Run("success", func(t *testing.T) {
			follower := uuid.New()
			followee1 := uuid.New()
			followee2 := uuid.New()
			followee3 := uuid.New()

			// Follow two users
			require.NoError(t, followerRepo.Follow(ctx, follower, followee1))
			require.NoError(t, followerRepo.Follow(ctx, follower, followee2))

			// Check all three users
			followees := []uuid.UUID{followee1, followee2, followee3}
			followingSet, err := followerRepo.FindFollowees(ctx, follower, followees)
			require.NoError(t, err)

			// Should only be following two users
			assert.Equal(t, 2, followingSet.Cardinality())
			assert.True(t, followingSet.Contains(followee1))
			assert.True(t, followingSet.Contains(followee2))
			assert.False(t, followingSet.Contains(followee3))
		})

		t.Run("empty followees list", func(t *testing.T) {
			follower := uuid.New()

			followingSet, err := followerRepo.FindFollowees(ctx, follower, []uuid.UUID{})
			require.NoError(t, err)
			assert.Equal(t, mapset.NewThreadUnsafeSet[uuid.UUID](), followingSet)
		})

		t.Run("not following any users", func(t *testing.T) {
			follower := uuid.New()
			followee1 := uuid.New()
			followee2 := uuid.New()

			followees := []uuid.UUID{followee1, followee2}
			followingSet, err := followerRepo.FindFollowees(ctx, follower, followees)
			require.NoError(t, err)
			assert.Equal(t, 0, followingSet.Cardinality())
		})
	})
}

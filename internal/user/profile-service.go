package user

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"log/slog"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
)

type FollowerRepositoryInterface interface {
	IsFollowing(ctx context.Context, follower, followee uuid.UUID) (bool, error)
	BatchIsFollowing(ctx context.Context, follower uuid.UUID, followee []uuid.UUID) (map[uuid.UUID]bool, error)
	Follow(ctx context.Context, follower, followee uuid.UUID) error
	UnFollow(ctx context.Context, follower, followee uuid.UUID) error
}

func (p ProfileService) IsFollowing(ctx context.Context, follower, followee uuid.UUID) (bool, error) {
	return p.FollowerRepository.IsFollowing(ctx, follower, followee)
}

func (p ProfileService) IsFollowingBulk(ctx context.Context, follower uuid.UUID, followee []uuid.UUID) (map[uuid.UUID]bool, error) {
	return p.FollowerRepository.BatchIsFollowing(ctx, follower, followee)
}

func (p ProfileService) Follow(ctx context.Context, follower uuid.UUID, followeeUsername string) (domain.User, error) {
	followedUser, err := p.UserRepository.FindUserByUsername(ctx, followeeUsername)
	if err != nil {
		return domain.User{}, err
	}

	if followedUser.Id == follower {
		return domain.User{}, errutil.ErrCantFollowYourself
	}

	err = p.FollowerRepository.Follow(ctx, follower, followedUser.Id)
	if err != nil {
		return domain.User{}, err
	}
	return followedUser, nil
}

func (p ProfileService) UnFollow(ctx context.Context, follower uuid.UUID, followeeUsername string) (domain.User, error) {
	followedUser, err := p.UserRepository.FindUserByUsername(ctx, followeeUsername)
	if err != nil {
		return domain.User{}, err
	}

	if followedUser.Id == follower {
		return domain.User{}, errutil.ErrCantFollowYourself
	}

	err = p.FollowerRepository.UnFollow(ctx, follower, followedUser.Id)
	if err != nil {
		return domain.User{}, err
	}
	return followedUser, nil
}

func (p ProfileService) GetUserProfile(ctx context.Context, loggedInUserId *uuid.UUID, username string) (domain.User, bool, error) {
	followedUser, err := p.UserRepository.FindUserByUsername(ctx, username)
	if err != nil {
		return domain.User{}, false, err
	}

	if loggedInUserId == nil {
		slog.DebugContext(ctx, "no logged in user. skipping isFollowing check")
		return followedUser, false, nil
	} else {
		isFollowing, err := p.IsFollowing(ctx, *loggedInUserId, followedUser.Id)
		if err != nil {
			return domain.User{}, false, err
		}
		slog.DebugContext(ctx, fmt.Sprintf("is user %s following %s: %v", loggedInUserId, followedUser.Id, isFollowing))
		return followedUser, isFollowing, nil
	}
}

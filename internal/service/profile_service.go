package service

import (
	"context"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"log/slog"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/repository"
)

type ProfileService struct {
	FollowerRepository repository.FollowerRepositoryInterface
	UserRepository     repository.UserRepositoryInterface
}

type ProfileServiceInterface interface {
	GetUserProfile(c context.Context, loggedInUserId *uuid.UUID, username string) (domain.User, bool, error)
	Follow(c context.Context, follower uuid.UUID, followeeUsername string) (domain.User, error)
	UnFollow(c context.Context, follower uuid.UUID, followeeUsername string) (domain.User, error)
	IsFollowing(c context.Context, follower, followee uuid.UUID) (bool, error)
	IsFollowingBulk(ctx context.Context, follower uuid.UUID, followee []uuid.UUID) (mapset.Set[uuid.UUID], error)
}

var _ ProfileServiceInterface = ProfileService{}

func (p ProfileService) IsFollowing(ctx context.Context, follower, followee uuid.UUID) (bool, error) {
	return p.FollowerRepository.IsFollowing(ctx, follower, followee)
}

func (p ProfileService) IsFollowingBulk(ctx context.Context, follower uuid.UUID, followee []uuid.UUID) (mapset.Set[uuid.UUID], error) {
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

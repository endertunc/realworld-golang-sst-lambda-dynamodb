package user

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
)

type FollowerRepositoryInterface interface {
	IsFollowing(c context.Context, follower, followee uuid.UUID) (bool, error)
	Follow(c context.Context, follower, followee uuid.UUID) error
	UnFollow(c context.Context, follower, followee uuid.UUID) error
}

func (s FollowerService) IsFollowing(c context.Context, follower, followee uuid.UUID) (bool, error) {
	return s.FollowerRepository.IsFollowing(c, follower, followee)
}

func (s FollowerService) Follow(c context.Context, follower uuid.UUID, followeeUsername string) (domain.User, bool, error) {
	followeeUser, err := s.UserService.GetUserByEmail(c, followeeUsername)
	if err != nil {
		return domain.User{}, false, err
	}

	if followeeUser.Id == follower {
		return domain.User{}, false, errors.New("cannot follow yourself")
	}

	err = s.FollowerRepository.Follow(c, follower, followeeUser.Id)
	if err != nil {
		return domain.User{}, false, err
	}
	return followeeUser, false, nil
}

func (s FollowerService) UnFollow(c context.Context, follower uuid.UUID, followeeUsername string) (domain.User, bool, error) {
	followeeUser, err := s.UserService.GetUserByEmail(c, followeeUsername)
	if err != nil {
		return domain.User{}, false, err
	}

	if followeeUser.Id == follower {
		return domain.User{}, false, errors.New("cannot unfollow yourself")
	}

	err = s.FollowerRepository.UnFollow(c, followeeUser.Id, follower)
	if err != nil {
		return domain.User{}, false, err
	}
	return followeeUser, false, nil
}

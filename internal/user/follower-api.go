package user

import (
	"context"
	"github.com/google/uuid"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
)

type FollowerApi struct {
	FollowerService FollowerServiceInterface
}

type FollowerService struct {
	FollowerRepository FollowerRepositoryInterface
	UserService        UserServiceInterface
}

type FollowerServiceInterface interface {
	IsFollowing(c context.Context, loggedInUser uuid.UUID, followerUser uuid.UUID) (bool, error)
	Follow(c context.Context, loggedInUser uuid.UUID, followerUsername string) (domain.User, bool, error)
	UnFollow(c context.Context, loggedInUser uuid.UUID, followerUsername string) (domain.User, bool, error)
}

func (fa FollowerApi) UnfollowUserByUsername(ctx context.Context, loggedInUser uuid.UUID, followerUsername string) (dto.ProfileResponseBodyDTO, error) {
	user, isFollowing, err := fa.FollowerService.UnFollow(ctx, loggedInUser, followerUsername)
	if err != nil {
		return dto.ProfileResponseBodyDTO{}, err
	}
	return dto.ToProfileResponseBodyDTO(user, isFollowing), nil
}

func (fa FollowerApi) FollowUserByUsername(ctx context.Context, loggedInUser uuid.UUID, followerUsername string) (dto.ProfileResponseBodyDTO, error) {
	user, isFollowing, err := fa.FollowerService.Follow(ctx, loggedInUser, followerUsername)
	if err != nil {
		return dto.ProfileResponseBodyDTO{}, err
	}
	return dto.ToProfileResponseBodyDTO(user, isFollowing), nil
}

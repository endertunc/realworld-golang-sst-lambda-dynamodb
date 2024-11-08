package user

import (
	"context"
	"github.com/google/uuid"
	"log/slog"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
)

type ProfileApi struct {
	ProfileService ProfileServiceInterface
}

type ProfileService struct {
	FollowerRepository FollowerRepositoryInterface
	UserRepository     UserRepositoryInterface
}

type ProfileServiceInterface interface {
	GetUserProfile(c context.Context, loggedInUserId *uuid.UUID, username string) (domain.User, bool, error)
	Follow(c context.Context, follower uuid.UUID, followeeUsername string) (domain.User, error)
	UnFollow(c context.Context, follower uuid.UUID, followeeUsername string) (domain.User, error)
	IsFollowing(c context.Context, follower, followee uuid.UUID) (bool, error)
	// ToDo @ender - map[uuid.UUID]bool is a bit weird return type. That being said, it fits to purpose without boilerplate
	IsFollowingBulk(ctx context.Context, follower uuid.UUID, followee []uuid.UUID) (map[uuid.UUID]bool, error)
}

var _ ProfileServiceInterface = ProfileService{}

func (pa ProfileApi) UnfollowUserByUsername(ctx context.Context, loggedInUser uuid.UUID, followerUsername string) (dto.ProfileResponseBodyDTO, error) {
	user, err := pa.ProfileService.UnFollow(ctx, loggedInUser, followerUsername)
	if err != nil {
		return dto.ProfileResponseBodyDTO{}, err
	}
	return dto.ToProfileResponseBodyDTO(user, false), nil
}

func (pa ProfileApi) FollowUserByUsername(ctx context.Context, loggedInUser uuid.UUID, followerUsername string) (dto.ProfileResponseBodyDTO, error) {
	user, err := pa.ProfileService.Follow(ctx, loggedInUser, followerUsername)
	if err != nil {
		return dto.ProfileResponseBodyDTO{}, err
	}
	return dto.ToProfileResponseBodyDTO(user, true), nil
}

func (pa ProfileApi) GetUserProfile(context context.Context, loggedInUserId *uuid.UUID, profileUsername string) (dto.ProfileResponseBodyDTO, error) {
	slog.DebugContext(context, "getting user profile", slog.Any("loggedInUserId", loggedInUserId), slog.String("profileUsername", profileUsername))
	user, isFollowing, err := pa.ProfileService.GetUserProfile(context, loggedInUserId, profileUsername)
	if err != nil {
		return dto.ProfileResponseBodyDTO{}, err
	}
	return dto.ToProfileResponseBodyDTO(user, isFollowing), nil
}

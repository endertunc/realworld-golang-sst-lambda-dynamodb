package api

import (
	"context"
	"github.com/google/uuid"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"realworld-aws-lambda-dynamodb-golang/internal/service"
)

type ProfileApi struct {
	ProfileService service.ProfileServiceInterface
}

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
	user, isFollowing, err := pa.ProfileService.GetUserProfile(context, loggedInUserId, profileUsername)
	if err != nil {
		return dto.ProfileResponseBodyDTO{}, err
	}
	return dto.ToProfileResponseBodyDTO(user, isFollowing), nil
}

package api

import (
	"errors"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/service"
)

type ProfileApi struct {
	ProfileService service.ProfileServiceInterface
}

func NewProfileApi(profileService service.ProfileServiceInterface) ProfileApi {
	return ProfileApi{ProfileService: profileService}
}

func (pa ProfileApi) UnfollowUserByUsername(w http.ResponseWriter, r *http.Request, loggedInUser uuid.UUID) {
	ctx := r.Context()

	followeeUsername, ok := GetPathParamHTTP(ctx, w, r, "username")
	if !ok {
		return
	}

	user, err := pa.ProfileService.UnFollow(ctx, loggedInUser, followeeUsername)
	if err != nil {
		if errors.Is(err, errutil.ErrUserNotFound) {
			slog.DebugContext(ctx, "user not found", slog.String("username", followeeUsername))
			ToSimpleHTTPError(w, http.StatusNotFound, "user not found")
			return
		}
		if errors.Is(err, errutil.ErrCantFollowYourself) {
			slog.DebugContext(ctx, "user tried to unfollow itself", slog.String("username", followeeUsername), slog.String("userId", loggedInUser.String()))
			ToSimpleHTTPError(w, http.StatusConflict, "cannot unfollow yourself")
			return
		}
		ToInternalServerHTTPError(w, err)
		return
	}
	resp := dto.ToProfileResponseBodyDTO(user, false)
	ToSuccessHTTPResponse(w, resp)
}

func (pa ProfileApi) FollowUserByUsername(w http.ResponseWriter, r *http.Request, loggedInUser uuid.UUID) {
	ctx := r.Context()

	followeeUsername, ok := GetPathParamHTTP(ctx, w, r, "username")
	if !ok {
		return
	}

	user, err := pa.ProfileService.Follow(ctx, loggedInUser, followeeUsername)
	if err != nil {
		if errors.Is(err, errutil.ErrUserNotFound) {
			slog.DebugContext(ctx, "user to follow not found", slog.Any("username", followeeUsername), slog.Any("error", err))
			ToSimpleHTTPError(w, http.StatusNotFound, "user not found")
			return
		}
		if errors.Is(err, errutil.ErrCantFollowYourself) {
			slog.DebugContext(ctx, "user tried to follow itself", slog.Any("username", followeeUsername), slog.String("userId", loggedInUser.String()), slog.Any("error", err))
			ToSimpleHTTPError(w, http.StatusBadRequest, "cannot follow yourself")
			return
		}
		ToInternalServerHTTPError(w, err)
		return
	}
	resp := dto.ToProfileResponseBodyDTO(user, true)
	ToSuccessHTTPResponse(w, resp)
}

func (pa ProfileApi) GetUserProfile(w http.ResponseWriter, r *http.Request, loggedInUserId *uuid.UUID) {
	ctx := r.Context()

	profileUsername, ok := GetPathParamHTTP(ctx, w, r, "username")
	if !ok {
		return
	}

	user, isFollowing, err := pa.ProfileService.GetUserProfile(ctx, loggedInUserId, profileUsername)
	if err != nil {
		if errors.Is(err, errutil.ErrUserNotFound) {
			slog.DebugContext(ctx, "user profile not found", slog.String("username", profileUsername), slog.Any("error", err))
			ToSimpleHTTPError(w, http.StatusNotFound, "user not found")
			return
		}
		ToInternalServerHTTPError(w, err)
		return
	}
	resp := dto.ToProfileResponseBodyDTO(user, isFollowing)
	ToSuccessHTTPResponse(w, resp)

}

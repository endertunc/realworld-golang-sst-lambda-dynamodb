package api

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/service"
)

type UserApi struct {
	UserService service.UserServiceInterface
}

func NewUserApi(userService service.UserServiceInterface) UserApi {
	return UserApi{UserService: userService}
}

func (ua UserApi) LoginUser(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	loginRequestBodyDTO, ok := ParseAndValidateBody[dto.LoginRequestBodyDTO](ctx, w, r)

	if !ok {
		return
	}
	loginUser := loginRequestBodyDTO.User
	token, user, err := ua.UserService.LoginUser(ctx, loginUser.Email, loginUser.Password)
	if err != nil {
		if errors.Is(err, errutil.ErrUserNotFound) || errors.Is(err, errutil.ErrInvalidPassword) {
			slog.WarnContext(ctx, "invalid credentials", slog.Any("error", err))
			ToSimpleHTTPError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		ToInternalServerHTTPError(w, err)
		return
	}
	resp := dto.ToUserResponseBodyDTO(*user, *token)
	ToSuccessHTTPResponse(w, resp)
}

func (ua UserApi) RegisterUser(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	newUserRequestBodyDTO, ok := ParseAndValidateBody[dto.NewUserRequestBodyDTO](ctx, w, r)
	if !ok {
		return
	}

	newUser := newUserRequestBodyDTO.User
	token, user, err := ua.UserService.RegisterUser(ctx, newUser.Email, newUser.Username, newUser.Password)
	if err != nil {
		if errors.Is(err, errutil.ErrUsernameAlreadyExists) {
			username := newUserRequestBodyDTO.User.Username
			slog.WarnContext(ctx, "username already exists", slog.String("username", username), slog.Any("error", err))
			ToSimpleHTTPError(w, http.StatusConflict, "username already exists")
			return
		}

		if errors.Is(err, errutil.ErrEmailAlreadyExists) {
			email := newUserRequestBodyDTO.User.Email
			slog.WarnContext(ctx, "email already exists", slog.String("email", email), slog.Any("error", err))
			ToSimpleHTTPError(w, http.StatusConflict, "email already exists")
			return
		}
		ToInternalServerHTTPError(w, err)
		return
	}
	resp := dto.ToUserResponseBodyDTO(*user, *token)
	ToSuccessHTTPResponse(w, resp)
}

func (ua UserApi) GetCurrentUser(ctx context.Context, w http.ResponseWriter, r *http.Request, userID uuid.UUID, token domain.Token) {
	user, err := ua.UserService.GetUserByUserId(ctx, userID)
	if err != nil {
		// this should not happen since user has a valid token, therefore it should exist
		if errors.Is(err, errutil.ErrUserNotFound) {
			slog.WarnContext(ctx, "user not found", slog.Any("error", err))
			ToSimpleHTTPError(w, http.StatusNotFound, "user not found")
			return
		}
		ToInternalServerHTTPError(w, err)
		return
	}
	resp := dto.ToUserResponseBodyDTO(user, token)
	ToSuccessHTTPResponse(w, resp)
}

func (ua UserApi) UpdateCurrentUser(ctx context.Context, w http.ResponseWriter, r *http.Request, userID uuid.UUID, token domain.Token) {
	updateUserRequestBodyDTO, ok := ParseAndValidateBody[dto.UpdateUserRequestBodyDTO](ctx, w, r)
	if !ok {
		return
	}

	updateUser := updateUserRequestBodyDTO.User
	newToken, user, err := ua.UserService.UpdateUser(ctx, userID, updateUser.Email, updateUser.Username, updateUser.Password, updateUser.Bio, updateUser.Image)
	if err != nil {
		if errors.Is(err, errutil.ErrUsernameAlreadyExists) {
			slog.WarnContext(ctx, "username already exists", slog.Any("error", err))
			ToSimpleHTTPError(w, http.StatusConflict, "username already exists")
			return
		}

		if errors.Is(err, errutil.ErrEmailAlreadyExists) {
			slog.WarnContext(ctx, "email already exists", slog.Any("error", err))
			ToSimpleHTTPError(w, http.StatusConflict, "email already exists")
			return
		}

		ToInternalServerHTTPError(w, err)
		return
	}

	resp := dto.ToUserResponseBodyDTO(*user, *newToken)
	ToSuccessHTTPResponse(w, resp)
}

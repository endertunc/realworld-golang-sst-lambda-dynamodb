package api

import (
	"context"
	"github.com/google/uuid"
	"log/slog"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
	"realworld-aws-lambda-dynamodb-golang/internal/service"
)

type UserApi struct {
	UserService service.UserServiceInterface
}

func (ua UserApi) LoginUser(ctx context.Context, loginRequestBodyDTO dto.LoginRequestBodyDTO) (dto.UserResponseBodyDTO, error) {
	loginUser := loginRequestBodyDTO.User
	token, user, err := ua.UserService.LoginUser(ctx, loginUser.Email, loginUser.Password)
	if err != nil {
		return dto.UserResponseBodyDTO{}, err
	}
	return dto.ToUserResponseBodyDTO(*user, *token), nil
}

func (ua UserApi) RegisterUser(ctx context.Context, newUserRequestBodyDTO dto.NewUserRequestBodyDTO) (dto.UserResponseBodyDTO, error) {
	newUser := newUserRequestBodyDTO.User
	token, user, err := ua.UserService.RegisterUser(ctx, newUser.Email, newUser.Username, newUser.Password)
	if err != nil {
		return dto.UserResponseBodyDTO{}, err
	}
	return dto.ToUserResponseBodyDTO(*user, *token), nil
}

func (ua UserApi) GetCurrentUser(ctx context.Context, userID uuid.UUID, token domain.Token) (dto.UserResponseBodyDTO, error) {
	user, err := ua.UserService.GetCurrentUser(ctx, userID)
	if err != nil {
		return dto.UserResponseBodyDTO{}, err
	}
	slog.DebugContext(ctx, "token", slog.Any("token", token))
	return dto.ToUserResponseBodyDTO(user, token), nil
}

func (ua UserApi) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	return ua.UserService.GetUserByUsername(ctx, email)
}

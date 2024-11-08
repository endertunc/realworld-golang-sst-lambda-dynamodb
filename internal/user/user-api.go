package user

import (
	"context"
	"github.com/google/uuid"
	"log/slog"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
)

type UserApi struct {
	UserService UserServiceInterface
}

type UserService struct {
	UserRepository UserRepositoryInterface
}

type UserServiceInterface interface {
	LoginUser(ctx context.Context, email, plainTextPassword string) (*domain.Token, *domain.User, error)
	RegisterUser(ctx context.Context, email, username, plainTextPassword string) (*domain.Token, *domain.User, error)
	GetCurrentUser(ctx context.Context, userID uuid.UUID) (domain.User, error)
	GetUserByUserId(ctx context.Context, userID uuid.UUID) (domain.User, error)
	//GetUserProfile(ctx context.Context, loggedInUserId *uuid.UUID, profileUsername string) (domain.User, bool, error)
	GetUserByUsername(ctx context.Context, email string) (domain.User, error)
	GetUserListByUserIDs(ctx context.Context, userIds []uuid.UUID) ([]domain.User, error)
}

var _ UserServiceInterface = UserService{}

func (s UserService) GetUserByUsername(ctx context.Context, username string) (domain.User, error) {
	return s.UserRepository.FindUserByUsername(ctx, username)
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

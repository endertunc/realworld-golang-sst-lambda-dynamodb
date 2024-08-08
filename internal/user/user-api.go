package user

import (
	"context"
	"github.com/google/uuid"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"
)

type UserApi struct {
	UserService UserServiceInterface
}

type UserService struct {
	UserRepository  UserRepositoryInterface
	FollowerService FollowerServiceInterface
}

func (s UserService) GetUserByEmail(c context.Context, email string) (domain.User, error) {
	//TODO implement me
	panic("implement me")
}

type UserServiceInterface interface {
	LoginUser(c context.Context, email, plainTextPassword string) (*domain.Token, *domain.User, error)
	RegisterUser(c context.Context, email, username, plainTextPassword string) (*domain.Token, *domain.User, error)
	GetCurrentUser(c context.Context, userID uuid.UUID) (*domain.Token, *domain.User, error)
	GetUserProfile(c context.Context, loggedInUserId *uuid.UUID, profileUsername string) (domain.User, bool, error)
	GetUserByEmail(c context.Context, email string) (domain.User, error)
}

func (ua UserApi) LoginUser(context context.Context, loginRequestBodyDTO dto.LoginRequestBodyDTO) (dto.UserResponseBodyDTO, error) {
	loginUser := loginRequestBodyDTO.User
	token, user, err := ua.UserService.LoginUser(context, loginUser.Email, loginUser.Password)
	if err != nil {
		return dto.UserResponseBodyDTO{}, err
	}
	return dto.ToUserResponseBodyDTO(*user, *token), nil
}

func (ua UserApi) RegisterUser(context context.Context, newUserRequestBodyDTO dto.NewUserRequestBodyDTO) (dto.UserResponseBodyDTO, error) {
	newUser := newUserRequestBodyDTO.User
	token, user, err := ua.UserService.RegisterUser(context, newUser.Email, newUser.Username, newUser.Password)
	if err != nil {
		return dto.UserResponseBodyDTO{}, err
	}
	return dto.ToUserResponseBodyDTO(*user, *token), nil
}

func (ua UserApi) GetCurrentUser(context context.Context, userID uuid.UUID) (dto.UserResponseBodyDTO, error) {
	token, user, err := ua.UserService.GetCurrentUser(context, userID)
	if err != nil {
		return dto.UserResponseBodyDTO{}, err
	}
	return dto.ToUserResponseBodyDTO(*user, *token), nil
}

func (ua UserApi) GetUserByEmail(c context.Context, email string) (domain.User, error) {
	return ua.UserService.GetUserByEmail(c, email)
}

func (ua UserApi) GetUserProfile(context context.Context, loggedInUserId *uuid.UUID, profileUsername string) (dto.ProfileResponseBodyDTO, error) {
	user, isFollowing, err := ua.UserService.GetUserProfile(context, loggedInUserId, profileUsername)
	if err != nil {
		return dto.ProfileResponseBodyDTO{}, err
	}
	return dto.ToProfileResponseBodyDTO(user, isFollowing), nil
}

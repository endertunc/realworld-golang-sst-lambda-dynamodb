package user

import (
	"context"
	"fmt"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/security"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserRepositoryInterface interface {
	FindUserByEmail(c context.Context, email string) (domain.User, error)
	FindUserByUsername(c context.Context, username string) (domain.User, error)
	FindUserById(c context.Context, userId uuid.UUID) (domain.User, error)
	InsertNewUser(c context.Context, newUser domain.User) (domain.User, error)
	FindUserListByUserIDs(c context.Context, userIds []uuid.UUID) ([]domain.User, error)
}

func (s UserService) LoginUser(c context.Context, email, plainTextPassword string) (*domain.Token, *domain.User, error) {
	user, err := s.UserRepository.FindUserByEmail(c, email)
	if err != nil {
		return nil, nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(plainTextPassword))
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", errutil.ErrInvalidPassword, err)
	}

	token, err := security.GenerateToken(user.Id)
	if err != nil {
		return nil, nil, err
	}

	return token, &user, nil
}

func (s UserService) RegisterUser(ctx context.Context, email, username, plainTextPassword string) (*domain.Token, *domain.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainTextPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", errutil.ErrHashPassword, err)
	}

	// ToDo @ender - we should make sure that regardless of the casing, username and email should be unique
	// 	dynamoDB does not support case-insensitive queries out of the box tho...
	newUser := domain.NewUser(email, username, string(hashedPassword)) // ToDo Ender string(hashedPassword)

	user, err := s.UserRepository.InsertNewUser(ctx, newUser)
	if err != nil {
		return nil, nil, err
	}

	token, err := security.GenerateToken(user.Id)
	if err != nil {
		return nil, nil, err
	}

	return token, &user, nil
}

func (s UserService) GetCurrentUser(c context.Context, userId uuid.UUID) (domain.User, error) {
	user, err := s.UserRepository.FindUserById(c, userId)
	if err != nil {
		return domain.User{}, err
	}
	return user, nil
}

func (s UserService) GetUserByUserId(c context.Context, userId uuid.UUID) (domain.User, error) {
	user, err := s.UserRepository.FindUserById(c, userId)
	if err != nil {
		return domain.User{}, err
	}
	return user, nil
}

func (s UserService) GetUserListByUserIDs(c context.Context, userIds []uuid.UUID) ([]domain.User, error) {
	users, err := s.UserRepository.FindUserListByUserIDs(c, userIds)
	if err != nil {
		return nil, err
	}
	return users, nil
}

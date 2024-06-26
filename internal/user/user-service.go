package user

import (
	"context"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/security"
)

type UserService struct {
	UserRepository UserRepositoryInterface
}

type Token string

type UserServiceInterface interface {
	LoginUser(c context.Context, email string, plainTextPassword string) (*Token, *domain.User, error)
}

func (s UserService) LoginUser(c context.Context, email string, plainTextPassword string) (*Token, *domain.User, error) {
	user, err := s.UserRepository.FindUserByEmail(c, email)
	if err != nil {
		return nil, nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(plainTextPassword))
	if err != nil {
		return nil, nil, errutil.InternalError(fmt.Errorf("LoginUser - password hash compare error: %w", err))
	}

	token, err := security.GenerateToken(email)
	if err != nil {
		return nil, nil, err
	}

	return token, &user, nil
}

package service

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/repository"
	"realworld-aws-lambda-dynamodb-golang/internal/security"
	"time"
)

type userService struct {
	userRepository repository.UserRepositoryInterface
}

type UserServiceInterface interface {
	LoginUser(ctx context.Context, email, plainTextPassword string) (*domain.Token, *domain.User, error)
	RegisterUser(ctx context.Context, email, username, plainTextPassword string) (*domain.Token, *domain.User, error)
	GetUserByUserId(ctx context.Context, userID uuid.UUID) (domain.User, error)
	GetUserByUsername(ctx context.Context, username string) (domain.User, error)
	GetUserListByUserIDs(ctx context.Context, userIds []uuid.UUID) ([]domain.User, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, email, username, plainTextPassword *string, bio, image *string) (*domain.Token, *domain.User, error)
}

var _ UserServiceInterface = userService{} //nolint:golint,exhaustruct

func NewUserService(userRepository repository.UserRepositoryInterface) UserServiceInterface {
	return userService{userRepository: userRepository}
}

func (s userService) LoginUser(c context.Context, email, plainTextPassword string) (*domain.Token, *domain.User, error) {
	user, err := s.userRepository.FindUserByEmail(c, email)
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

func (s userService) RegisterUser(ctx context.Context, email, username, plainTextPassword string) (*domain.Token, *domain.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainTextPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", errutil.ErrHashPassword, err)
	}

	// ToDo @ender - we should make sure that regardless of the casing, username and email should be unique
	// 	dynamoDB does not support case-insensitive queries out of the box tho...
	newUser := domain.NewUser(email, username, string(hashedPassword)) // ToDo Ender string(hashedPassword)

	user, err := s.userRepository.InsertNewUser(ctx, newUser)
	if err != nil {
		return nil, nil, err
	}

	token, err := security.GenerateToken(user.Id)
	if err != nil {
		return nil, nil, err
	}

	return token, &user, nil
}

func (s userService) GetUserByUserId(c context.Context, userId uuid.UUID) (domain.User, error) {
	user, err := s.userRepository.FindUserById(c, userId)
	if err != nil {
		return domain.User{}, err
	}
	return user, nil
}

func (s userService) GetUserListByUserIDs(c context.Context, userIds []uuid.UUID) ([]domain.User, error) {
	users, err := s.userRepository.FindUserListByUserIDs(c, userIds)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s userService) GetUserByUsername(ctx context.Context, username string) (domain.User, error) {
	return s.userRepository.FindUserByUsername(ctx, username)
}

func (s userService) UpdateUser(ctx context.Context, userID uuid.UUID, email, username, plainTextPassword *string, bio, image *string) (*domain.Token, *domain.User, error) {
	user, err := s.userRepository.FindUserById(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	oldEmail := user.Email
	oldUsername := user.Username

	// Update fields if provided
	if email != nil {
		user.Email = *email
	}
	if username != nil {
		user.Username = *username
	}
	if plainTextPassword != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*plainTextPassword), bcrypt.DefaultCost)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: %w", errutil.ErrHashPassword, err)
		}
		user.HashedPassword = string(hashedPassword)
	}
	if bio != nil {
		user.Bio = bio
	}
	if image != nil {
		user.Image = image
	}

	user.UpdatedAt = time.Now()

	updatedUser, err := s.userRepository.UpdateUser(ctx, user, oldEmail, oldUsername)
	if err != nil {
		return nil, nil, err
	}

	token, err := security.GenerateToken(updatedUser.Id)
	if err != nil {
		return nil, nil, err
	}

	return token, &updatedUser, nil
}

package user

import (
	"context"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/security"
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
		return nil, nil, errutil.ErrPasswordHash.Errorf("LoginUser - password hash compare error: %w", err)
	}

	token, err := security.GenerateToken(user.Id)
	if err != nil {
		return nil, nil, err
	}

	return token, &user, nil
}

func (s UserService) RegisterUser(c context.Context, email, username, plainTextPassword string) (*domain.Token, *domain.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainTextPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, errutil.ErrPasswordHash.Errorf("RegisterUser - password hash error: %w", err)
	}
	newUser := domain.NewUser(email, username, string(hashedPassword)) // ToDo Ender string(hashedPassword)
	user, err := s.UserRepository.InsertNewUser(c, newUser)

	token, err := security.GenerateToken(user.Id)
	if err != nil {
		return nil, nil, err
	}

	return token, &user, nil
}

func (s UserService) GetCurrentUser(c context.Context, userId uuid.UUID) (domain.Token, domain.User, error) {
	user, err := s.UserRepository.FindUserById(c, userId)
	if err != nil {
		return "", domain.User{}, err
	}

	token, err := security.GenerateToken(userId)
	if err != nil {
		return "", domain.User{}, err
	}

	return *token, user, nil
}

// ToDo @ender - this should move to a profile service so we can use both userservice and profile service...
// or API layer should just call the userservice and profile service separately since this is purely needed for the API...
func (s UserService) GetUserProfile(c context.Context, loggedInUserId *uuid.UUID, username string) (domain.User, bool, error) {
	user, err := s.UserRepository.FindUserByUsername(c, username)
	if err != nil {
		return domain.User{}, false, err
	}

	if loggedInUserId == nil {
		return user, false, nil
	} else {
		// ToDo @ender we need to fix this...
		//isFollowing, err := s.FollowerService.IsFollowing(c, *loggedInUserId, user.Id)
		isFollowing, err := fakeIsFollowing(c, *loggedInUserId, user.Id)
		if err != nil {
			return domain.User{}, false, err
		}
		return user, isFollowing, nil
	}
}

func fakeIsFollowing(c context.Context, follower, followee uuid.UUID) (bool, error) {
	return false, nil
}

func (s UserService) GetUserListByUserIDs(c context.Context, userIds []uuid.UUID) ([]domain.User, error) {
	users, err := s.UserRepository.FindUserListByUserIDs(c, userIds)
	if err != nil {
		return nil, err
	}
	return users, nil
}

package service

import (
	"context"
	"errors"
	"realworld-aws-lambda-dynamodb-golang/internal/test"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/generator"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/repository/mocks"
)

func TestUserService_LoginUser(t *testing.T) {
	ctx := context.Background()

	t.Run("successful login", func(t *testing.T) {
		withUserTestContext(t, func(tc userTestContext) {
			// Setup test data
			expectedUser := generator.GenerateUser()
			password := gofakeit.Password(true, true, true, true, false, 10)
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			expectedUser.HashedPassword = string(hashedPassword)

			// Setup expectations
			tc.mockRepo.EXPECT().
				FindUserByEmail(mock.Anything, expectedUser.Email).
				Return(expectedUser, nil)

			// Execute
			token, user, err := tc.userService.LoginUser(ctx, expectedUser.Email, password)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, token)
			assert.NotNil(t, user)
			assert.Equal(t, expectedUser.Email, user.Email)
		})
	})

	t.Run("user not found", func(t *testing.T) {
		withUserTestContext(t, func(tc userTestContext) {
			// Setup test data
			email := gofakeit.Email()
			password := gofakeit.Password(true, true, true, true, false, 10)

			// Setup expectations
			tc.mockRepo.EXPECT().
				FindUserByEmail(mock.Anything, email).
				Return(domain.User{}, errutil.ErrUserNotFound)

			// Execute
			token, user, err := tc.userService.LoginUser(ctx, email, password)

			// Assert
			assert.True(t, errors.Is(err, errutil.ErrUserNotFound))
			assert.Nil(t, token)
			assert.Nil(t, user)
		})
	})

	t.Run("invalid password", func(t *testing.T) {
		withUserTestContext(t, func(tc userTestContext) {
			// Setup test data
			expectedUser := generator.GenerateUser()
			correctPassword := gofakeit.Password(true, true, true, true, false, 10)
			wrongPassword := gofakeit.Password(true, true, true, true, false, 10)
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)
			expectedUser.HashedPassword = string(hashedPassword)

			// Setup expectations
			tc.mockRepo.EXPECT().
				FindUserByEmail(mock.Anything, expectedUser.Email).
				Return(expectedUser, nil)

			// Execute
			token, user, err := tc.userService.LoginUser(ctx, expectedUser.Email, wrongPassword)

			// Assert
			assert.True(t, errors.Is(err, errutil.ErrInvalidPassword))
			assert.Nil(t, token)
			assert.Nil(t, user)
		})
	})
}

func TestUserService_RegisterUser(t *testing.T) {
	ctx := context.Background()

	t.Run("successful registration", func(t *testing.T) {
		withUserTestContext(t, func(tc userTestContext) {
			// Setup test data
			expectedUser := generator.GenerateUser()
			password := gofakeit.Password(true, true, true, true, false, 10)

			// Setup expectations
			tc.mockRepo.EXPECT().
				InsertNewUser(mock.Anything, mock.MatchedBy(func(user domain.User) bool {
					return user.Email == expectedUser.Email && user.Username == expectedUser.Username
				})).
				Return(expectedUser, nil)

			// Execute
			token, user, err := tc.userService.RegisterUser(ctx, expectedUser.Email, expectedUser.Username, password)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, token)
			assert.NotNil(t, user)
			assert.Equal(t, expectedUser.Email, user.Email)
			assert.Equal(t, expectedUser.Username, user.Username)
		})
	})

	t.Run("duplicate email", func(t *testing.T) {
		withUserTestContext(t, func(tc userTestContext) {
			// Setup test data
			email := gofakeit.Email()
			username := gofakeit.Username()
			password := gofakeit.Password(true, true, true, true, false, 10)

			// Setup expectations
			tc.mockRepo.EXPECT().
				InsertNewUser(mock.Anything, mock.Anything).
				Return(domain.User{}, errutil.ErrEmailAlreadyExists)

			// Execute
			token, user, err := tc.userService.RegisterUser(ctx, email, username, password)

			// Assert
			assert.True(t, errors.Is(err, errutil.ErrEmailAlreadyExists))
			assert.Nil(t, token)
			assert.Nil(t, user)
		})
	})

	t.Run("duplicate username", func(t *testing.T) {
		withUserTestContext(t, func(tc userTestContext) {
			// Setup test data
			email := gofakeit.Email()
			username := gofakeit.Username()
			password := gofakeit.Password(true, true, true, true, false, 10)

			// Setup expectations
			tc.mockRepo.EXPECT().
				InsertNewUser(mock.Anything, mock.Anything).
				Return(domain.User{}, errutil.ErrUsernameAlreadyExists)

			// Execute
			token, user, err := tc.userService.RegisterUser(ctx, email, username, password)

			// Assert
			assert.True(t, errors.Is(err, errutil.ErrUsernameAlreadyExists))
			assert.Nil(t, token)
			assert.Nil(t, user)
		})
	})

}

func TestUserService_GetUserByUserId(t *testing.T) {
	ctx := context.Background()

	t.Run("successful get user", func(t *testing.T) {
		withUserTestContext(t, func(tc userTestContext) {
			// Setup test data
			expectedUser := generator.GenerateUser()

			// Setup expectations
			tc.mockRepo.EXPECT().
				FindUserById(mock.Anything, expectedUser.Id).
				Return(expectedUser, nil)

			// Execute
			user, err := tc.userService.GetUserByUserId(ctx, expectedUser.Id)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, expectedUser, user)
		})
	})

	t.Run("user not found", func(t *testing.T) {
		withUserTestContext(t, func(tc userTestContext) {
			// Setup test data
			userId := uuid.New()

			// Setup expectations
			tc.mockRepo.EXPECT().
				FindUserById(mock.Anything, userId).
				Return(domain.User{}, errutil.ErrUserNotFound)

			// Execute
			user, err := tc.userService.GetUserByUserId(ctx, userId)

			// Assert
			assert.True(t, errors.Is(err, errutil.ErrUserNotFound))
			assert.Empty(t, user)
		})
	})
}

func TestUserService_UpdateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		withUserTestContext(t, func(tc userTestContext) {
			// Setup test data
			oldUser := generator.GenerateUser()
			updatedUser := generator.GenerateUser()
			updatedUser.Id = oldUser.Id
			bio := gofakeit.Quote()
			updatedUser.Bio = &bio

			newPassword := gofakeit.Password(true, true, true, true, false, 10)

			// Setup expectations
			tc.mockRepo.EXPECT().
				FindUserById(mock.Anything, oldUser.Id).
				Return(oldUser, nil)

			tc.mockRepo.EXPECT().
				UpdateUser(mock.Anything, mock.MatchedBy(func(user domain.User) bool {
					return user.Email == updatedUser.Email &&
						user.Username == updatedUser.Username &&
						*user.Bio == bio &&
						user.HashedPassword != oldUser.HashedPassword // Password should be updated
				}), oldUser.Email, oldUser.Username).
				Return(updatedUser, nil)

			// Execute
			token, user, err := tc.userService.UpdateUser(ctx, oldUser.Id, &updatedUser.Email, &updatedUser.Username, &newPassword, &bio, nil)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, token)
			assert.NotNil(t, user)
			assert.Equal(t, updatedUser.Email, user.Email)
			assert.Equal(t, updatedUser.Username, user.Username)
			assert.Equal(t, bio, *user.Bio)
		})
	})

	t.Run("user not found", func(t *testing.T) {
		withUserTestContext(t, func(tc userTestContext) {
			// Setup test data
			userId := uuid.New()
			email := gofakeit.Email()

			// Setup expectations
			tc.mockRepo.EXPECT().
				FindUserById(mock.Anything, userId).
				Return(domain.User{}, errutil.ErrUserNotFound)

			// Execute
			token, user, err := tc.userService.UpdateUser(ctx, userId, &email, nil, nil, nil, nil)

			// Assert
			assert.True(t, errors.Is(err, errutil.ErrUserNotFound))
			assert.Nil(t, token)
			assert.Nil(t, user)
		})
	})
}

// - - - - - - - - - - - - - - - - Test Context - - - - - - - - - - - - - - - -

type userTestContext struct {
	userService UserServiceInterface
	mockRepo    *mocks.MockUserRepositoryInterface
}

func createUserTestContext(t *testing.T) userTestContext {
	mockRepo := mocks.NewMockUserRepositoryInterface(t)
	userService := NewUserService(mockRepo)
	test.SetupMockKeyProvider(t)

	return userTestContext{
		userService: userService,
		mockRepo:    mockRepo,
	}
}

func withUserTestContext(t *testing.T, testFunc func(tc userTestContext)) {
	testFunc(createUserTestContext(t))
}

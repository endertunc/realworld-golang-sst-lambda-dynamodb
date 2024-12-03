package repository

import (
	"context"
	"realworld-aws-lambda-dynamodb-golang/internal/database"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/generator"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/test"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var userRepo = NewDynamodbUserRepository(database.NewDynamoDBStore())

func TestInsertNewUser(t *testing.T) {
	ctx := context.Background()
	test.WithSetupAndTeardown(t, func() {
		t.Run("success", func(t *testing.T) {
			newUser := generator.GenerateUser()

			insertedUser, err := userRepo.InsertNewUser(ctx, newUser)
			require.NoError(t, err)
			assert.Equal(t, newUser.Email, insertedUser.Email)
			assert.Equal(t, newUser.Username, insertedUser.Username)
			assert.Equal(t, newUser.HashedPassword, insertedUser.HashedPassword)
		})

		t.Run("user with duplicate email", func(t *testing.T) {
			// First user
			user1 := generator.GenerateUser()
			_, err := userRepo.InsertNewUser(ctx, user1)
			require.NoError(t, err)

			// Second user with same email
			user2 := generator.GenerateUser()
			user2.Email = user1.Email
			_, err = userRepo.InsertNewUser(ctx, user2)
			assert.ErrorIs(t, err, errutil.ErrEmailAlreadyExists)
		})

		t.Run("user with duplicate username", func(t *testing.T) {
			// First user
			user1 := generator.GenerateUser()
			_, err := userRepo.InsertNewUser(ctx, user1)
			require.NoError(t, err)

			// Second user with same username
			user2 := generator.GenerateUser()
			user2.Username = user1.Username
			_, err = userRepo.InsertNewUser(ctx, user2)
			assert.ErrorIs(t, err, errutil.ErrUsernameAlreadyExists)
		})
	})
}

func TestFindUserByEmail(t *testing.T) {
	ctx := context.Background()
	test.WithSetupAndTeardown(t, func() {
		t.Run("existing user", func(t *testing.T) {
			user := generator.GenerateUser()
			_, err := userRepo.InsertNewUser(ctx, user)
			require.NoError(t, err)

			foundUser, err := userRepo.FindUserByEmail(ctx, user.Email)
			require.NoError(t, err)
			assert.Equal(t, user.Email, foundUser.Email)
			assert.Equal(t, user.Username, foundUser.Username)
		})

		t.Run("non-existent user", func(t *testing.T) {
			_, err := userRepo.FindUserByEmail(ctx, "nonexistent@example.com")
			assert.ErrorIs(t, err, errutil.ErrUserNotFound)
		})
	})
}

func TestFindUserById(t *testing.T) {
	ctx := context.Background()
	test.WithSetupAndTeardown(t, func() {
		t.Run("existing user", func(t *testing.T) {
			user := generator.GenerateUser()
			insertedUser, err := userRepo.InsertNewUser(ctx, user)
			require.NoError(t, err)

			foundUser, err := userRepo.FindUserById(ctx, insertedUser.Id)
			require.NoError(t, err)
			assert.Equal(t, insertedUser.Id, foundUser.Id)
			assert.Equal(t, insertedUser.Email, foundUser.Email)
			assert.Equal(t, insertedUser.Username, foundUser.Username)
		})

		t.Run("non-existent user", func(t *testing.T) {
			_, err := userRepo.FindUserById(ctx, uuid.New())
			assert.ErrorIs(t, err, errutil.ErrUserNotFound)
		})
	})
}

func TestFindUserByUsername(t *testing.T) {
	ctx := context.Background()
	test.WithSetupAndTeardown(t, func() {
		t.Run("existing user", func(t *testing.T) {
			user := generator.GenerateUser()
			_, err := userRepo.InsertNewUser(ctx, user)
			require.NoError(t, err)

			foundUser, err := userRepo.FindUserByUsername(ctx, user.Username)
			require.NoError(t, err)
			assert.Equal(t, user.Email, foundUser.Email)
			assert.Equal(t, user.Username, foundUser.Username)
		})

		t.Run("non-existent user", func(t *testing.T) {
			_, err := userRepo.FindUserByUsername(ctx, "nonexistentuser")
			assert.ErrorIs(t, err, errutil.ErrUserNotFound)
		})
	})
}

func TestFindUserListByUserIDs(t *testing.T) {
	ctx := context.Background()
	test.WithSetupAndTeardown(t, func() {
		t.Run("multiple existing users", func(t *testing.T) {
			// Create test users
			users := make([]domain.User, 3)
			userIDs := make([]uuid.UUID, 3)

			for i := 0; i < 3; i++ {
				user := generator.GenerateUser()
				insertedUser, err := userRepo.InsertNewUser(ctx, user)
				require.NoError(t, err)
				users[i] = insertedUser
				userIDs[i] = insertedUser.Id
			}

			foundUsers, err := userRepo.FindUsersByIds(ctx, userIDs)
			require.NoError(t, err)
			assert.Equal(t, len(users), len(foundUsers))

			// Verify each user is found
			for _, user := range users {
				found := false
				for _, foundUser := range foundUsers {
					if foundUser.Id == user.Id {
						found = true
						assert.Equal(t, user.Email, foundUser.Email)
						assert.Equal(t, user.Username, foundUser.Username)
						break
					}
				}
				assert.True(t, found, "User with ID %s not found", user.Id)
			}
		})

		t.Run("empty list", func(t *testing.T) {
			foundUsers, err := userRepo.FindUsersByIds(ctx, []uuid.UUID{})
			require.NoError(t, err)
			assert.Empty(t, foundUsers)
		})

		t.Run("mixed existing and non-existing users", func(t *testing.T) {
			user := generator.GenerateUser()
			insertedUser, err := userRepo.InsertNewUser(ctx, user)
			require.NoError(t, err)

			userIDs := []uuid.UUID{insertedUser.Id, uuid.New()}
			foundUsers, err := userRepo.FindUsersByIds(ctx, userIDs)
			require.NoError(t, err)
			assert.Equal(t, 1, len(foundUsers))
			assert.Equal(t, insertedUser.Id, foundUsers[0].Id)
		})
	})
}

func TestUpdateUser(t *testing.T) {
	ctx := context.Background()
	test.WithSetupAndTeardown(t, func() {
		t.Run("success", func(t *testing.T) {
			// Create user
			user := generator.GenerateUser()
			insertedUser, err := userRepo.InsertNewUser(ctx, user)
			require.NoError(t, err)

			// Create updated user
			updatedUser := generator.GenerateUser()
			// Id and CreatedAt should not change
			updatedUser.Id = insertedUser.Id
			updatedUser.CreatedAt = insertedUser.CreatedAt

			userAfterUpdate, err := userRepo.UpdateUser(ctx, updatedUser, insertedUser.Email, insertedUser.Username)
			require.NoError(t, err)
			assert.Equal(t, userAfterUpdate, updatedUser)

			// Verify user is updated
			userFromDatabase, err := userRepo.FindUserByEmail(ctx, updatedUser.Email)
			require.NoError(t, err)
			assert.Equal(t, updatedUser, userFromDatabase)
		})

		t.Run("update to existing email", func(t *testing.T) {
			user1 := generator.GenerateUser()
			insertedUser1, err := userRepo.InsertNewUser(ctx, user1)
			require.NoError(t, err)

			user2 := generator.GenerateUser()
			insertedUser2, err := userRepo.InsertNewUser(ctx, user2)
			require.NoError(t, err)

			// Try to update user2's email to user1's email
			updatedUser := insertedUser2
			updatedUser.Email = insertedUser1.Email
			_, err = userRepo.UpdateUser(ctx, updatedUser, insertedUser2.Email, insertedUser2.Username)
			assert.ErrorIs(t, err, errutil.ErrEmailAlreadyExists)

			// Verify user2 is not changed
			user2FromDatabase, err := userRepo.FindUserByEmail(ctx, insertedUser2.Email)
			require.NoError(t, err)
			assert.Equal(t, insertedUser2, user2FromDatabase)

		})

		t.Run("update to existing username", func(t *testing.T) {
			user1 := generator.GenerateUser()
			insertedUser1, err := userRepo.InsertNewUser(ctx, user1)
			require.NoError(t, err)

			user2 := generator.GenerateUser()
			insertedUser2, err := userRepo.InsertNewUser(ctx, user2)
			require.NoError(t, err)

			// Try to update user2's username to user1's username
			updatedUser := insertedUser2
			updatedUser.Username = insertedUser1.Username
			_, err = userRepo.UpdateUser(ctx, updatedUser, insertedUser2.Email, insertedUser2.Username)
			assert.ErrorIs(t, err, errutil.ErrUsernameAlreadyExists)

			// Verify user2 is not changed
			user2FromDatabase, err := userRepo.FindUserByEmail(ctx, insertedUser2.Email)
			require.NoError(t, err)
			assert.Equal(t, insertedUser2, user2FromDatabase)
		})
	})
}

// ToDo @ender this really needs to be moved to helpers/commons
func stringPtr(s string) *string {
	return &s
}

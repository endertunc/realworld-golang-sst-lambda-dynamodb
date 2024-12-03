package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"realworld-aws-lambda-dynamodb-golang/internal/database"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"time"
)

const (
	userTable              = "user"
	userEmailGSI           = "user_email_gsi"
	userUsernameGSI        = "user_username_gsi"
	conditionalCheckFailed = "ConditionalCheckFailed"
)

type dynamodbUserRepository struct {
	db *database.DynamoDBStore
}

type UserRepositoryInterface interface {
	FindUserByEmail(c context.Context, email string) (domain.User, error)
	FindUserByUsername(c context.Context, username string) (domain.User, error)
	FindUserById(c context.Context, userId uuid.UUID) (domain.User, error)
	InsertNewUser(c context.Context, newUser domain.User) (domain.User, error)
	FindUsersByIds(c context.Context, userIds []uuid.UUID) ([]domain.User, error)
	UpdateUser(c context.Context, user domain.User, oldEmail string, oldUsername string) (domain.User, error)
}

var _ UserRepositoryInterface = dynamodbUserRepository{} //nolint:golint,exhaustruct

func NewDynamodbUserRepository(db *database.DynamoDBStore) UserRepositoryInterface {
	return dynamodbUserRepository{db: db}
}

type DynamodbUserItem struct {
	Id             DynamodbUUID `dynamodbav:"pk"`
	Email          string       `dynamodbav:"email"`
	HashedPassword string       `dynamodbav:"hashedPassword"`
	Username       string       `dynamodbav:"username"`
	Bio            *string      `dynamodbav:"bio,omitempty"`
	Image          *string      `dynamodbav:"image,omitempty"`
	CreatedAt      int64        `dynamodbav:"createdAt"`
	UpdatedAt      int64        `dynamodbav:"updatedAt"`
}

var _ UserRepositoryInterface = (*dynamodbUserRepository)(nil)

func (s dynamodbUserRepository) FindUserByEmail(ctx context.Context, email string) (domain.User, error) {
	input := dynamodb.QueryInput{
		TableName:              aws.String(userTable),
		IndexName:              aws.String(userEmailGSI),
		KeyConditionExpression: aws.String("email = :email"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":email": &ddbtypes.AttributeValueMemberS{Value: email},
		},
		Select: ddbtypes.SelectAllAttributes,
	}

	user, err := QueryOne(ctx, s.db.Client, &input, toDomainUser)
	if err != nil {
		if errors.Is(err, ErrDynamodbItemNotFound) {
			return domain.User{}, errutil.ErrUserNotFound
		}
		return domain.User{}, err
	}
	return user, nil
}

func (s dynamodbUserRepository) InsertNewUser(ctx context.Context, newUser domain.User) (domain.User, error) {
	dynamodbUserItem := toDynamoDbUser(newUser)
	userAttributes, err := attributevalue.MarshalMap(dynamodbUserItem)

	if err != nil {
		return domain.User{}, fmt.Errorf("%w: %w", errutil.ErrDynamoMapping, err)
	}

	transactWriteItems := dynamodb.TransactWriteItemsInput{
		TransactItems: []ddbtypes.TransactWriteItem{
			{
				Put: &ddbtypes.Put{
					TableName:           aws.String(userTable),
					Item:                userAttributes,
					ConditionExpression: aws.String("attribute_not_exists(pk)"),
				},
			},
			{
				Put: &ddbtypes.Put{
					TableName: aws.String(userTable),
					Item: map[string]ddbtypes.AttributeValue{
						"pk": &ddbtypes.AttributeValueMemberS{Value: "username#" + newUser.Username},
					},
					ConditionExpression: aws.String("attribute_not_exists(pk)"),
				},
			},
			{
				Put: &ddbtypes.Put{
					TableName: aws.String(userTable),
					Item: map[string]ddbtypes.AttributeValue{
						"pk": &ddbtypes.AttributeValueMemberS{Value: "email#" + newUser.Email},
					},
					ConditionExpression: aws.String("attribute_not_exists(pk)"),
				},
			},
		},
	}

	_, err = s.db.Client.TransactWriteItems(ctx, &transactWriteItems)

	if err != nil {
		var canceledException *ddbtypes.TransactionCanceledException
		if errors.As(err, &canceledException) {
			for index, reason := range canceledException.CancellationReasons {
				if reason.Code != nil && *reason.Code == conditionalCheckFailed && index == 1 {
					return domain.User{}, fmt.Errorf("%w: %w", errutil.ErrUsernameAlreadyExists, err)
				}
				if reason.Code != nil && *reason.Code == conditionalCheckFailed && index == 2 {
					return domain.User{}, fmt.Errorf("%w: %w", errutil.ErrEmailAlreadyExists, err)
				}
			}
		}
		return domain.User{}, fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}

	return newUser, nil
}

func (s dynamodbUserRepository) FindUserById(ctx context.Context, userId uuid.UUID) (domain.User, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(userTable),
		Key: map[string]ddbtypes.AttributeValue{
			"pk": &ddbtypes.AttributeValueMemberS{Value: userId.String()},
		},
	}

	user, err := GetItem(ctx, s.db.Client, input, toDomainUser)
	if err != nil {
		if errors.Is(err, ErrDynamodbItemNotFound) {
			return domain.User{}, errutil.ErrUserNotFound
		}
		return domain.User{}, err
	}
	return user, nil
}

func (s dynamodbUserRepository) FindUserByUsername(ctx context.Context, username string) (domain.User, error) {
	input := dynamodb.QueryInput{
		TableName:              aws.String(userTable),
		IndexName:              aws.String(userUsernameGSI),
		KeyConditionExpression: aws.String("username = :username"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":username": &ddbtypes.AttributeValueMemberS{Value: username},
		},
	}

	user, err := QueryOne(ctx, s.db.Client, &input, toDomainUser)
	if err != nil {
		if errors.Is(err, ErrDynamodbItemNotFound) {
			return domain.User{}, errutil.ErrUserNotFound
		}
		return domain.User{}, err
	}
	return user, nil
}

func (s dynamodbUserRepository) UpdateUser(ctx context.Context, user domain.User, oldEmail string, oldUsername string) (domain.User, error) {
	dynamodbUserItem := toDynamoDbUser(user)
	userAttributes, err := attributevalue.MarshalMap(dynamodbUserItem)
	if err != nil {
		return domain.User{}, fmt.Errorf("%w: %w", errutil.ErrDynamoMapping, err)
	}

	transactItems := []ddbtypes.TransactWriteItem{
		{
			Put: &ddbtypes.Put{
				TableName: aws.String(userTable),
				Item:      userAttributes,
			},
		},
	}

	// If email changed, update email index
	if user.Email != oldEmail {
		transactItems = append(transactItems,
			ddbtypes.TransactWriteItem{
				Delete: &ddbtypes.Delete{
					TableName: aws.String(userTable),
					Key: map[string]ddbtypes.AttributeValue{
						"pk": &ddbtypes.AttributeValueMemberS{Value: "email#" + oldEmail},
					},
				},
			},
			ddbtypes.TransactWriteItem{
				Put: &ddbtypes.Put{
					TableName: aws.String(userTable),
					Item: map[string]ddbtypes.AttributeValue{
						"pk": &ddbtypes.AttributeValueMemberS{Value: "email#" + user.Email},
					},
					ConditionExpression: aws.String("attribute_not_exists(pk)"),
				},
			},
		)
	}

	// If username changed, update username index
	if user.Username != oldUsername {
		transactItems = append(transactItems,
			ddbtypes.TransactWriteItem{
				Delete: &ddbtypes.Delete{
					TableName: aws.String(userTable),
					Key: map[string]ddbtypes.AttributeValue{
						"pk": &ddbtypes.AttributeValueMemberS{Value: "username#" + oldUsername},
					},
				},
			},
			ddbtypes.TransactWriteItem{
				Put: &ddbtypes.Put{
					TableName: aws.String(userTable),
					Item: map[string]ddbtypes.AttributeValue{
						"pk": &ddbtypes.AttributeValueMemberS{Value: "username#" + user.Username},
					},
					ConditionExpression: aws.String("attribute_not_exists(pk)"),
				},
			},
		)
	}

	_, err = s.db.Client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: transactItems,
	})

	if err != nil {
		var canceledException *ddbtypes.TransactionCanceledException
		if errors.As(err, &canceledException) {
			// Index 0: Main user record update
			// Index 1-2: Email update (if changed)
			// Last two indices: Username update (if changed, regardless of email change)
			emailChanged := user.Email != oldEmail
			usernameChanged := user.Username != oldUsername

			for i, reason := range canceledException.CancellationReasons {
				if reason.Code != nil && *reason.Code == conditionalCheckFailed {
					// Email uniqueness check fails at index 2
					if emailChanged && i == 2 {
						return domain.User{}, fmt.Errorf("%w: %w", errutil.ErrEmailAlreadyExists, err)
					}
					// Username uniqueness check fails at last index
					if usernameChanged && i == len(canceledException.CancellationReasons)-1 {
						return domain.User{}, fmt.Errorf("%w: %w", errutil.ErrUsernameAlreadyExists, err)
					}
				}
			}
		}
		return domain.User{}, fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}

	return user, nil
}

func (s dynamodbUserRepository) FindUsersByIds(ctx context.Context, userIds []uuid.UUID) ([]domain.User, error) {
	// short circuit if userIds is empty, no need to query
	// also, dynamodb will throw a validation error if we try to query with empty keys
	if len(userIds) == 0 {
		return []domain.User{}, nil
	}

	keys := make([]map[string]ddbtypes.AttributeValue, 0, len(userIds))
	for _, userId := range userIds {
		keys = append(keys, map[string]ddbtypes.AttributeValue{
			"pk": &ddbtypes.AttributeValueMemberS{Value: userId.String()},
		})
	}

	return BatchGetItems(ctx, s.db.Client, userTable, keys, toDomainUser)
}

func toDynamoDbUser(user domain.User) DynamodbUserItem {
	return DynamodbUserItem{
		Id:             DynamodbUUID(user.Id),
		Email:          user.Email,
		HashedPassword: user.HashedPassword,
		Username:       user.Username,
		Bio:            user.Bio,
		Image:          user.Image,
		CreatedAt:      user.CreatedAt.UnixMilli(),
		UpdatedAt:      user.UpdatedAt.UnixMilli(),
	}
}

func toDomainUser(user DynamodbUserItem) domain.User {
	return domain.User{
		Id:             uuid.UUID(user.Id),
		Email:          user.Email,
		HashedPassword: user.HashedPassword,
		Username:       user.Username,
		Bio:            user.Bio,
		Image:          user.Image,
		CreatedAt:      time.UnixMilli(user.CreatedAt),
		UpdatedAt:      time.UnixMilli(user.UpdatedAt),
	}
}

package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"realworld-aws-lambda-dynamodb-golang/internal/database"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

const (
	userTable              = "user"
	userEmailGSI           = "user_email_gsi"
	userUsernameGSI        = "user_username_gsi"
	conditionalCheckFailed = "ConditionalCheckFailed"
)

type DynamodbUserRepository struct {
	db *database.DynamoDBStore
}

func NewDynamodbUserRepository(db *database.DynamoDBStore) DynamodbUserRepository {
	return DynamodbUserRepository{db: db}
}

type UserRepositoryInterface interface {
	FindUserByEmail(c context.Context, email string) (domain.User, error)
	FindUserByUsername(c context.Context, username string) (domain.User, error)
	FindUserById(c context.Context, userId uuid.UUID) (domain.User, error)
	InsertNewUser(c context.Context, newUser domain.User) (domain.User, error)
	FindUserListByUserIDs(c context.Context, userIds []uuid.UUID) ([]domain.User, error)
}

var _ UserRepositoryInterface = DynamodbUserRepository{}

type DynamodbUserItem struct {
	Id             string    `dynamodbav:"pk"`
	Email          string    `dynamodbav:"email"`
	HashedPassword string    `dynamodbav:"hashedPassword"`
	Username       string    `dynamodbav:"username"`
	Bio            *string   `dynamodbav:"bio,omitempty"`
	Image          *string   `dynamodbav:"image,omitempty"`
	CreatedAt      time.Time `dynamodbav:"createdAt,unixtime"`
	UpdatedAt      time.Time `dynamodbav:"updatedAt,unixtime"`
}

var _ UserRepositoryInterface = (*DynamodbUserRepository)(nil)

// ToDo @ender there are a lot of duplicate code between FindByXXX methods duplication is not always a bad thing but just to be aware
func (s DynamodbUserRepository) FindUserByEmail(c context.Context, email string) (domain.User, error) {
	response, err := s.db.Client.Query(c, &dynamodb.QueryInput{
		TableName:              aws.String(userTable),
		IndexName:              aws.String(userEmailGSI),
		KeyConditionExpression: aws.String("email = :email"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":email": &ddbtypes.AttributeValueMemberS{Value: email},
		},
		Select: ddbtypes.SelectAllAttributes,
	})

	if err != nil {
		slog.ErrorContext(c, "query error", slog.Any("error", err))
		return domain.User{}, fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}

	if len(response.Items) == 0 {
		return domain.User{}, fmt.Errorf("%w: %s", errutil.ErrUserNotFound, email)
	}

	dynamodbUser := DynamodbUserItem{}
	err = attributevalue.UnmarshalMap(response.Items[0], &dynamodbUser)

	if err != nil {
		slog.ErrorContext(c, "mapping error", slog.Any("error", err))
		return domain.User{}, fmt.Errorf("%w: %w", errutil.ErrDynamoMapping, err)
	}
	domainUser := toDomainUser(dynamodbUser)
	return domainUser, nil
}

func (s DynamodbUserRepository) InsertNewUser(ctx context.Context, newUser domain.User) (domain.User, error) {
	opt := "InsertNewUser"
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
		//slog.WarnContext(ctx, "failed to insert new user",
		//	slog.String("func", opt),
		//	slog.Any("error", err),
		//)
		var canceledException *ddbtypes.TransactionCanceledException
		if errors.As(err, &canceledException) {
			for index, reason := range canceledException.CancellationReasons {
				//jsonData, _ := json.MarshalIndent(reason, "", "  ")
				if reason.Code != nil && *reason.Code == conditionalCheckFailed && index == 1 {
					slog.WarnContext(ctx, "username already exists",
						slog.String("func", opt),
						slog.Any("reason", reason),
						slog.Any("error", err),
						slog.Group("context", slog.String("username", newUser.Username)),
					)
					return domain.User{}, fmt.Errorf("%w: %w", errutil.ErrUsernameAlreadyExists, err)
				}
				if reason.Code != nil && *reason.Code == conditionalCheckFailed && index == 2 {
					slog.WarnContext(ctx, "email already exists",
						slog.String("func", opt),
						slog.Any("reason", reason),
						slog.Any("error", err),
						slog.Group("context", slog.String("email", newUser.Email)),
					)
					return domain.User{}, fmt.Errorf("%w: %w", errutil.ErrEmailAlreadyExists, err)
				}
			}
		}

		return domain.User{}, fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}

	return newUser, nil
}

func (s DynamodbUserRepository) FindUserById(c context.Context, userId uuid.UUID) (domain.User, error) {
	response, err := s.db.Client.GetItem(c, &dynamodb.GetItemInput{
		TableName: aws.String(userTable),
		Key: map[string]ddbtypes.AttributeValue{
			"pk": &ddbtypes.AttributeValueMemberS{Value: userId.String()},
		},
	})

	if err != nil {
		return domain.User{}, fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}

	if len(response.Item) == 0 {
		return domain.User{}, errutil.ErrUserNotFound
	}

	dynamodbUser := DynamodbUserItem{}
	err = attributevalue.UnmarshalMap(response.Item, &dynamodbUser)

	if err != nil {
		return domain.User{}, fmt.Errorf("%w: %w", errutil.ErrDynamoMapping, err)
	}
	domainUser := toDomainUser(dynamodbUser)
	return domainUser, nil
}

func (s DynamodbUserRepository) FindUserByUsername(c context.Context, username string) (domain.User, error) {
	response, err := s.db.Client.Query(c, &dynamodb.QueryInput{
		TableName:              aws.String(userTable),
		IndexName:              aws.String(userUsernameGSI),
		KeyConditionExpression: aws.String("username = :username"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":username": &ddbtypes.AttributeValueMemberS{Value: username},
		},
		Select: ddbtypes.SelectAllAttributes,
	})

	if err != nil {
		return domain.User{}, fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}

	if len(response.Items) == 0 {
		return domain.User{}, errutil.ErrUserNotFound
	}

	dynamodbUser := DynamodbUserItem{}
	err = attributevalue.UnmarshalMap(response.Items[0], &dynamodbUser)

	if err != nil {
		return domain.User{}, fmt.Errorf("%w: %w", errutil.ErrDynamoMapping, err)
	}
	user := toDomainUser(dynamodbUser)
	return user, nil
}

// FindUserListByUserIDs
/*
* ToDo @ender there are fundamental issues with how "realworld application" design and you can feel it here.
*  there is no partial fetch (like with "scrolling" with token) thus we are loading all comments.
*  dynamodb has batch request count (max 100 item) and size (16mb) limits to be aware.
 */
func (s DynamodbUserRepository) FindUserListByUserIDs(ctx context.Context, userIds []uuid.UUID) ([]domain.User, error) {
	// this check is necessary otherwise we will get "ValidationException" from dynamodb
	// because you must provide a non-empty list of keys to BatchGetItem
	if len(userIds) == 0 {
		return []domain.User{}, nil
	}

	userKeys := make([]map[string]ddbtypes.AttributeValue, 0, len(userIds))
	for _, userId := range userIds {
		userKeys = append(userKeys, map[string]ddbtypes.AttributeValue{
			"pk": &ddbtypes.AttributeValueMemberS{Value: userId.String()},
		})
	}

	batchGetItemInput := dynamodb.BatchGetItemInput{
		RequestItems: map[string]ddbtypes.KeysAndAttributes{
			userTable: {
				Keys: userKeys,
			},
		},
	}

	response, err := s.db.Client.BatchGetItem(ctx, &batchGetItemInput)

	if err != nil {
		return nil, fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}

	// ToDo @ender this logic seem to be wrong.
	//  We try to get list of users by their ids. Why is it an "error" if we can't find any user?
	//  It's not a concern for this method at least.
	//if len(response.Responses[userTable]) == 0 {
	//	return nil, errutil.ErrUserNotFound
	//}

	users := make([]domain.User, 0, len(response.Responses[userTable]))
	for _, item := range response.Responses[userTable] {
		dynamodbUser := DynamodbUserItem{}
		err = attributevalue.UnmarshalMap(item, &dynamodbUser)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", errutil.ErrDynamoMapping, err)
		}
		user := toDomainUser(dynamodbUser)
		users = append(users, user)
	}

	return users, nil
}

func toDynamoDbUser(user domain.User) DynamodbUserItem {
	return DynamodbUserItem{
		Id:             user.Id.String(),
		Email:          user.Email,
		HashedPassword: user.HashedPassword,
		Username:       user.Username,
		Bio:            user.Bio,
		Image:          user.Image,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
	}
}

func toDomainUser(user DynamodbUserItem) domain.User {
	return domain.User{
		Id:             uuid.MustParse(user.Id),
		Email:          user.Email,
		HashedPassword: user.HashedPassword,
		Username:       user.Username,
		Bio:            user.Bio,
		Image:          user.Image,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
	}
}

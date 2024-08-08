package user

import (
	"context"
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

const userTable = "user"
const userEmailGSIName = "user_email_gsi"

type DynamodbUserRepository struct {
	db database.DynamoDBStore
}

type DynamodbUserItem struct {
	Id             uuid.UUID `dynamodbav:"PK"`
	Email          string    `dynamodbav:"email"`
	HashedPassword string    `dynamodbav:"hashed_password"`
	Username       string    `dynamodbav:"username"`
	Bio            *string   `dynamodbav:"bio,omitempty"`
	Image          *string   `dynamodbav:"image,omitempty"`
	CreatedAt      time.Time `dynamodbav:"created_at,unixtime"`
	UpdatedAt      time.Time `dynamodbav:"updated_at,unixtime"`
}

func (s DynamodbUserRepository) FindUserByEmail(c context.Context, email string) (domain.User, error) {
	response, err := s.db.Client.Query(c, &dynamodb.QueryInput{
		TableName:              aws.String(userTable),
		IndexName:              aws.String(userEmailGSIName),
		KeyConditionExpression: aws.String("PK = :email"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":email": &ddbtypes.AttributeValueMemberS{Value: email},
		},
		Select: ddbtypes.SelectAllAttributes,
		Limit:  aws.Int32(1),
	})
	user := domain.User{}
	if err != nil {
		return user, errutil.ErrDynamoQuery.Errorf("FindUserByEmail - query error: %w", err)
	}

	if len(response.Items) == 0 {
		return user, errutil.ErrUserNotFound.Errorf("FindUserByEmail - user with email [%s] not found", email)
	}

	err = attributevalue.UnmarshalMap(response.Items[0], &user)

	if err != nil {
		return user, errutil.ErrDynamoMapping.Errorf("FindUserByEmail - mapping error: %v", err)
	}

	return user, nil
}

func (s DynamodbUserRepository) InsertNewUser(c context.Context, newUser domain.User) (domain.User, error) {
	dynamodbUserItem := toDynamoDbUser(newUser)
	userAttributes, err := attributevalue.MarshalMap(dynamodbUserItem)

	if err != nil {
		return domain.User{}, errutil.ErrDynamoMapping.Errorf("InsertNewUser - mapping error: %w", err)
	}

	transactWriteItems := dynamodb.TransactWriteItemsInput{
		TransactItems: []ddbtypes.TransactWriteItem{
			{
				Put: &ddbtypes.Put{
					TableName:           aws.String(userTable),
					Item:                userAttributes,
					ConditionExpression: aws.String("attribute_not_exists(PK)"),
				},
			},
			{
				Put: &ddbtypes.Put{
					TableName: aws.String(userTable),
					Item: map[string]ddbtypes.AttributeValue{
						"PK": &ddbtypes.AttributeValueMemberS{Value: "username#" + newUser.Username},
					},
					ConditionExpression: aws.String("attribute_not_exists(PK)"),
				},
			},
			{
				Put: &ddbtypes.Put{
					TableName: aws.String(userTable),
					Item: map[string]ddbtypes.AttributeValue{
						"PK": &ddbtypes.AttributeValueMemberS{Value: "email#" + newUser.Email},
					},
					ConditionExpression: aws.String("attribute_not_exists(PK)"),
				},
			},
		},
	}

	_, err = s.db.Client.TransactWriteItems(c, &transactWriteItems)

	if err != nil {
		return domain.User{}, errutil.ErrDynamoQuery.Errorf("InsertNewUser - query error: %w", err)
	}

	return newUser, nil
}

func (s DynamodbUserRepository) FindUserById(c context.Context, userId uuid.UUID) (domain.User, error) {
	operation := "FindUserById"
	response, err := s.db.Client.GetItem(c, &dynamodb.GetItemInput{
		TableName: aws.String(userTable),
		Key: map[string]ddbtypes.AttributeValue{
			"PK": &ddbtypes.AttributeValueMemberS{Value: userId.String()},
		},
	})
	user := domain.User{}
	if err != nil {
		return user, errutil.ErrDynamoQuery.Errorf("%s - query error: %w", operation, err)
	}

	if len(response.Item) == 0 {
		return user, errutil.ErrUserNotFound.Errorf("%s - user with id [%s] not found", operation, userId.String())
	}

	err = attributevalue.UnmarshalMap(response.Item, &user)

	if err != nil {
		return user, errutil.ErrDynamoMapping.Errorf("%s - mapping error: %v", operation, err)
	}

	return user, nil
}

func (s DynamodbUserRepository) FindUserByUsername(c context.Context, username string) (domain.User, error) {
	//TODO implement me
	panic("implement me")
}

func toDynamoDbUser(user domain.User) DynamodbUserItem {
	return DynamodbUserItem{
		Id:             user.Id,
		Email:          user.Email,
		HashedPassword: user.HashedPassword,
		Username:       user.Username,
		Bio:            user.Bio,
		Image:          user.Image,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
	}
}

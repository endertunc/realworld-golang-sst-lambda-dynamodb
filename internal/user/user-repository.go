package user

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"realworld-aws-lambda-dynamodb-golang/internal/database"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
)

const TableName = "users"

type UserRepository struct {
	db database.DynamoDBStore
}

type UserRepositoryInterface interface {
	FindUserByEmail(c context.Context, email string) (domain.User, error)
}

func (s UserRepository) FindUserByEmail(c context.Context, email string) (*domain.User, error) {
	response, err := s.db.Client.GetItem(c, &dynamodb.GetItemInput{
		TableName: aws.String(TableName), // ToDo find an alternative to this wrapping maybe...
		Key: map[string]ddbtypes.AttributeValue{
			"email": &ddbtypes.AttributeValueMemberS{Value: email},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("FindUserByEmail - query error: %w", err)
	}
	user := domain.User{}
	if len(response.Item) == 0 {
		return nil, ErrUserNotFound
	}

	err = attributevalue.UnmarshalMap(response.Item, &user)

	if err != nil {
		return nil, errutil.InternalError(fmt.Errorf("FindUserByEmail - mapping error: %w", err))
	}

	return &user, nil
}

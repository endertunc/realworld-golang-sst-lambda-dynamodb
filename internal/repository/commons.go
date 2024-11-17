package repository

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
)

func QueryOne[DomainType interface{}, DynamodbType interface{}](ctx context.Context, client *dynamodb.Client, input *dynamodb.QueryInput, mapper func(input DynamodbType) DomainType) (DomainType, error) {
	var result DomainType
	response, err := client.Query(ctx, input)

	if err != nil {
		return result, fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}

	if len(response.Items) == 0 {
		return result, errutil.ErrUserNotFound // ToDo @ender this should be item not found or let the caller decide
	}

	var dynamodbItem DynamodbType
	err = attributevalue.UnmarshalMap(response.Items[0], &dynamodbItem)

	if err != nil {
		return result, fmt.Errorf("%w: %w", errutil.ErrDynamoMapping, err)
	}
	result = mapper(dynamodbItem)
	return result, nil
}

type DynamodbUUID uuid.UUID

// UnmarshalDynamoDBAttributeValue converts a DynamoDB string to UUID
func (u *DynamodbUUID) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	avS, ok := av.(*types.AttributeValueMemberS)
	if !ok {
		return fmt.Errorf("unexpected dynamodb attribute value type: %T", av)
	}

	parsed, err := uuid.Parse(avS.Value)
	if err != nil {
		return fmt.Errorf("failed to parse UUID: %w", err)
	}

	*u = DynamodbUUID(parsed)
	return nil
}

// String returns the string representation of the Status
func (u *DynamodbUUID) String() string {
	return u.String()
}

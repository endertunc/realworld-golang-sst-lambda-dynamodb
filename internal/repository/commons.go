package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"reflect"
)

var (
	ErrDynamodbItemNotFound = errors.New("dynamodb item not found")
)

func QueryOne[DomainType any, DynamodbType any](ctx context.Context, client *dynamodb.Client, input *dynamodb.QueryInput, mapper func(input DynamodbType) DomainType) (DomainType, error) {
	var result DomainType
	response, err := client.Query(ctx, input)

	if err != nil {
		return result, fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}

	if len(response.Items) == 0 {
		return result, ErrDynamodbItemNotFound
	}

	var dynamodbItem DynamodbType
	err = attributevalue.UnmarshalMap(response.Items[0], &dynamodbItem)

	if err != nil {
		return result, fmt.Errorf("%w: %w", errutil.ErrDynamoMapping, err)
	}
	result = mapper(dynamodbItem)
	return result, nil
}

func QueryMany[DomainType any, DynamodbType any](ctx context.Context, client *dynamodb.Client, input *dynamodb.QueryInput, mapper func(input DynamodbType) DomainType) ([]DomainType, error) {
	response, err := client.Query(ctx, input)

	if err != nil {
		return nil, fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}

	dynamodbItems := make([]DynamodbType, 0, len(response.Items))
	err = attributevalue.UnmarshalListOfMaps(response.Items, &dynamodbItems)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errutil.ErrDynamoMapping, err)
	}

	return lo.Map(dynamodbItems, func(dynamodbItem DynamodbType, _ int) DomainType {
		return mapper(dynamodbItem)
	}), nil
}

type DynamodbUUID uuid.UUID

var uuidType = reflect.TypeOf(uuid.UUID{})

// UnmarshalDynamoDBAttributeValue converts a DynamoDB string to UUID
// adopted from UnixTime implementation in aws-sdk-go-v2
func (u *DynamodbUUID) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	avS, ok := av.(*types.AttributeValueMemberS)
	if !ok {
		return &attributevalue.UnmarshalTypeError{
			Value: fmt.Sprintf("%T", av),
			Type:  reflect.TypeOf((*uuid.UUID)(nil)),
		}
	}

	parsedUuid, err := uuid.Parse(avS.Value)
	if err != nil {
		return &attributevalue.UnmarshalError{
			Err: err, Value: avS.Value, Type: uuidType,
		}
	}

	*u = DynamodbUUID(parsedUuid)
	return nil
}

// MarshalDynamoDBAttributeValue converts UUID to DynamoDB String
// adopted from UnixTime implementation in aws-sdk-go-v2
func (u DynamodbUUID) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return &types.AttributeValueMemberS{
		Value: uuid.UUID(u).String(),
	}, nil
}

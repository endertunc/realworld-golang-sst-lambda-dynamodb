package repository

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"reflect"
)

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

func Identity[T any](item T) T { return item }

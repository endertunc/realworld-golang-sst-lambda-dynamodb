package repository

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
)

var (
	ErrDynamodbItemNotFound = errors.New("dynamodb item not found")
)

// - - - - - - - - - - - - - - - - Query Helpers - - - - - - - - - - - - - - - -
func QueryOne[DomainType any, DynamodbType any](ctx context.Context, client *dynamodb.Client, input *dynamodb.QueryInput, mapper func(input DynamodbType) DomainType) (DomainType, error) {
	var result DomainType
	response, err := client.Query(ctx, input)

	if err != nil {
		return result, fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}

	if response.Count == 0 {
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

// QueryMany is a helper function to query multiple items from dynamodb.
// it will internally handle pagination and keep querying until the desired limit is reached
func QueryMany[DomainType any, DynamodbType any](ctx context.Context, client *dynamodb.Client, input *dynamodb.QueryInput, desiredLimit int, exclusiveStartKey map[string]types.AttributeValue, mapper func(input DynamodbType) DomainType) ([]DomainType, map[string]types.AttributeValue, error) {

	var domainItems []DomainType
	lastEvaluatedKey := exclusiveStartKey

	for len(domainItems) < desiredLimit {
		remaining := desiredLimit - len(domainItems)
		input.Limit = aws.Int32(int32(remaining))
		input.ExclusiveStartKey = lastEvaluatedKey

		response, err := client.Query(ctx, input)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
		}

		dynamodbItems := make([]DynamodbType, 0, len(response.Items))
		err = attributevalue.UnmarshalListOfMaps(response.Items, &dynamodbItems)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: %w", errutil.ErrDynamoMapping, err)
		}
		for _, dynamodbItem := range dynamodbItems {
			domainItems = append(domainItems, mapper(dynamodbItem))
		}
		// If no more items, break
		if response.LastEvaluatedKey == nil {
			lastEvaluatedKey = nil
			break
		}
		lastEvaluatedKey = response.LastEvaluatedKey
	}

	return domainItems, lastEvaluatedKey, nil
}

// BatchGetItems is a helper function to batch get multiple items from dynamodb.
// it will internally handle unprocessed keys and keep querying until all items are fetched
func BatchGetItems[DomainType any, DynamodbType any](
	ctx context.Context,
	client *dynamodb.Client,
	table string,
	keys []map[string]types.AttributeValue,
	mapper func(input DynamodbType) DomainType) ([]DomainType, error) {

	domainItems := make([]DomainType, 0, len(keys))
	unprocessedKeys := keys
	for len(unprocessedKeys) > 0 {
		response, err := client.BatchGetItem(ctx, &dynamodb.BatchGetItemInput{
			RequestItems: map[string]types.KeysAndAttributes{
				table: {
					Keys: unprocessedKeys,
				},
			},
		})

		if err != nil {
			return nil, fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
		}

		dynamodbItems := make([]DynamodbType, 0, len(response.Responses[table]))
		err = attributevalue.UnmarshalListOfMaps(response.Responses[table], &dynamodbItems)

		if err != nil {
			return nil, fmt.Errorf("%w: %w", errutil.ErrDynamoMapping, err)
		}

		for _, dynamodbItem := range dynamodbItems {
			domainItems = append(domainItems, mapper(dynamodbItem))
		}

		if len(response.UnprocessedKeys) == 0 {
			break
		}
		unprocessedKeys = response.UnprocessedKeys[table].Keys
	}

	return domainItems, nil
}

func GetItem[DomainType any, DynamodbType any](ctx context.Context, client *dynamodb.Client, input *dynamodb.GetItemInput, mapper func(input DynamodbType) DomainType) (DomainType, error) {
	var domainItem DomainType
	response, err := client.GetItem(ctx, input)
	if err != nil {
		return domainItem, fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}

	if len(response.Item) == 0 {
		return domainItem, ErrDynamodbItemNotFound
	}

	var dynamodbItem DynamodbType
	err = attributevalue.UnmarshalMap(response.Item, &dynamodbItem)
	if err != nil {
		return domainItem, fmt.Errorf("%w: %w", errutil.ErrDynamoMapping, err)
	}
	domainItem = mapper(dynamodbItem)
	return domainItem, nil
}

// - - - - - - - - - - - - - - - - LastEvaluatedKey Encoder/Decoder - - - - - - - - - - - - - - - -
func encodeLastEvaluatedKey(input map[string]types.AttributeValue) (*string, error) {
	var inputMap map[string]interface{}
	err := attributevalue.UnmarshalMap(input, &inputMap)
	if err != nil {
		return nil, err
	}
	bytesJSON, err := json.Marshal(inputMap)
	if err != nil {
		return nil, err
	}
	output := base64.StdEncoding.EncodeToString(bytesJSON)
	return &output, nil
}

func decodeLastEvaluatedKey(input string) (map[string]types.AttributeValue, error) {
	bytesJSON, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return nil, err
	}
	var outputJSON map[string]interface{}
	err = json.Unmarshal(bytesJSON, &outputJSON)
	if err != nil {
		return nil, err
	}

	return attributevalue.MarshalMap(outputJSON)
}

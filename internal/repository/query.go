package repository

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
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

// ToDo @ender here is the tricky thing with dynamodb query and pagination:
// query operation could return a result without reaching to the limit e.g., if 1 MB of data read before reaching the limit,
// therefore, we need to paginate INTERNALLY to request remaining items until we reach the limit or LastEvaluatedKey is empty
func QueryMany[DomainType any, DynamodbType any](ctx context.Context, client *dynamodb.Client, input *dynamodb.QueryInput, mapper func(input DynamodbType) DomainType) ([]DomainType, *string, error) {
	response, err := client.Query(ctx, input)

	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}

	dynamodbItems := make([]DynamodbType, 0, len(response.Items))
	err = attributevalue.UnmarshalListOfMaps(response.Items, &dynamodbItems)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", errutil.ErrDynamoMapping, err)
	}

	var nextPageToken *string
	if len(response.LastEvaluatedKey) > 0 {
		encodedToken, err := encodeLastEvaluatedKey(response.LastEvaluatedKey)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: %w", errutil.ErrDynamoTokenEncoding, err)
		}
		nextPageToken = encodedToken
	}

	domainItems := make([]DomainType, 0, len(dynamodbItems))
	for _, dynamodbItem := range dynamodbItems {
		domainItems = append(domainItems, mapper(dynamodbItem))
	}
	return domainItems, nextPageToken, nil
}

// ToDo @ender handle UnprocessedKeys
func BatchGetItems[DomainType any, DynamodbType any](ctx context.Context, client *dynamodb.Client,
	table string, keys []map[string]ddbtypes.AttributeValue, mapper func(input DynamodbType) DomainType) ([]DomainType, error) {

	response, err := client.BatchGetItem(ctx, &dynamodb.BatchGetItemInput{
		RequestItems: map[string]ddbtypes.KeysAndAttributes{
			favoriteTable: {
				Keys: keys,
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

	domainItems := make([]DomainType, 0, len(dynamodbItems))
	for _, dynamodbItem := range dynamodbItems {
		domainItems = append(domainItems, mapper(dynamodbItem))
	}
	return domainItems, nil
}

// - - - - - - - - - - - - - - - - LastEvaluatedKey Encoder/Decoder - - - - - - - - - - - - - - - -
func encodeLastEvaluatedKey(input map[string]ddbtypes.AttributeValue) (*string, error) {
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

func decodeLastEvaluatedKey(input string) (map[string]ddbtypes.AttributeValue, error) {
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

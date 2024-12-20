package repository

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"realworld-aws-lambda-dynamodb-golang/internal/database"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
)

var followerTable = "follower"

type dynamodbFollowerRepository struct {
	db *database.DynamoDBStore
}

type FollowerRepositoryInterface interface {
	FindFollowees(ctx context.Context, follower uuid.UUID, followee []uuid.UUID) (mapset.Set[uuid.UUID], error)
	Follow(ctx context.Context, follower, followee uuid.UUID) error
	UnFollow(ctx context.Context, follower, followee uuid.UUID) error
}

var _ FollowerRepositoryInterface = (*dynamodbFollowerRepository)(nil)

func NewDynamodbFollowerRepository(db *database.DynamoDBStore) FollowerRepositoryInterface {
	return dynamodbFollowerRepository{db: db}
}

// Note: there is no "use case" at the moment, but we should add createdAt to this item
type DynamodbFollowerItem struct {
	Follower DynamodbUUID `dynamodbav:"follower"`
	Followee DynamodbUUID `dynamodbav:"followee"`
}

func (s dynamodbFollowerRepository) IsFollowing(ctx context.Context, follower, followee uuid.UUID) (bool, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(followerTable),
		KeyConditionExpression: aws.String("followee = :followee AND follower = :follower"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":followee": &ddbtypes.AttributeValueMemberS{Value: followee.String()},
			":follower": &ddbtypes.AttributeValueMemberS{Value: follower.String()},
		},
		Select: ddbtypes.SelectCount,
	}

	result, err := s.db.Client.Query(ctx, input)
	if err != nil {
		return false, fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}
	return result.Count > 0, nil
}

func (s dynamodbFollowerRepository) Follow(ctx context.Context, follower, followee uuid.UUID) error {
	dynamodbFollowerItem := toDynamodbFollowerItem(follower, followee)
	followerAttributes, err := attributevalue.MarshalMap(dynamodbFollowerItem)

	if err != nil {
		return fmt.Errorf("%w: %w", errutil.ErrDynamoMapping, err)
	}

	input := &dynamodb.PutItemInput{Item: followerAttributes, TableName: aws.String(followerTable)}
	_, err = s.db.Client.PutItem(ctx, input)

	if err != nil {
		return fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}

	return nil
}

func (s dynamodbFollowerRepository) UnFollow(ctx context.Context, follower, followee uuid.UUID) error {
	dynamodbFollowerItem := toDynamodbFollowerItem(follower, followee)
	followerAttributes, err := attributevalue.MarshalMap(dynamodbFollowerItem)

	if err != nil {
		return fmt.Errorf("%w: %w", errutil.ErrDynamoMapping, err)
	}

	input := &dynamodb.DeleteItemInput{Key: followerAttributes, TableName: aws.String(followerTable)}
	_, err = s.db.Client.DeleteItem(ctx, input)

	if err != nil {
		return fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}

	return nil
}

func (s dynamodbFollowerRepository) FindFollowees(ctx context.Context, follower uuid.UUID, followees []uuid.UUID) (mapset.Set[uuid.UUID], error) {
	resultSet := mapset.NewThreadUnsafeSet[uuid.UUID]()
	// short circuit if followees is empty, no need to query
	// also, dynamodb will throw a validation error if we try to query with empty keys
	if len(followees) == 0 {
		return resultSet, nil
	}

	keys := make([]map[string]ddbtypes.AttributeValue, 0, len(follower))
	for _, followee := range followees {
		keys = append(keys, map[string]ddbtypes.AttributeValue{
			"follower": &ddbtypes.AttributeValueMemberS{Value: follower.String()},
			"followee": &ddbtypes.AttributeValueMemberS{Value: followee.String()},
		})
	}

	input := dynamodb.BatchGetItemInput{
		RequestItems: map[string]ddbtypes.KeysAndAttributes{
			followerTable: {
				Keys: keys,
			},
		},
	}

	response, err := s.db.Client.BatchGetItem(ctx, &input)
	if err != nil {
		return resultSet, fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}

	followersItems := response.Responses[followerTable]

	dynamodbFollowerItems := make([]DynamodbFollowerItem, 0, len(followersItems))
	err = attributevalue.UnmarshalListOfMaps(followersItems, &dynamodbFollowerItems)

	if err != nil {
		return nil, fmt.Errorf("%w: %w", errutil.ErrDynamoMapping, err)
	}

	for _, item := range dynamodbFollowerItems {
		resultSet.Add(uuid.UUID(item.Followee))
	}

	return resultSet, nil
}

func toDynamodbFollowerItem(follower, followee uuid.UUID) DynamodbFollowerItem {
	return DynamodbFollowerItem{
		Follower: DynamodbUUID(follower),
		Followee: DynamodbUUID(followee),
	}
}

package repository

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"log/slog"
	"realworld-aws-lambda-dynamodb-golang/internal/database"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
)

var followerTable = "follower"

type DynamodbFollowerRepository struct {
	db *database.DynamoDBStore // ToDo @ender should this be pointer or not??? Investigate to understand what is more proper
}

func NewDynamodbFollowerRepository(db *database.DynamoDBStore) DynamodbFollowerRepository {
	return DynamodbFollowerRepository{db: db}
}

type FollowerRepositoryInterface interface {
	IsFollowing(ctx context.Context, follower, followee uuid.UUID) (bool, error)
	BatchIsFollowing(ctx context.Context, follower uuid.UUID, followee []uuid.UUID) (map[uuid.UUID]bool, error)
	Follow(ctx context.Context, follower, followee uuid.UUID) error
	UnFollow(ctx context.Context, follower, followee uuid.UUID) error
}

var _ FollowerRepositoryInterface = (*DynamodbFollowerRepository)(nil)

type DynamodbFollowerItem struct {
	Follower string `dynamodbav:"follower"`
	Followee string `dynamodbav:"followee"`
}

// ToDo @ender - should we use GetItem or QueryInput in this case?
func (s DynamodbFollowerRepository) IsFollowing(ctx context.Context, follower, followee uuid.UUID) (bool, error) {
	isFollowingQueryInput := &dynamodb.QueryInput{
		TableName:              aws.String(followerTable),
		KeyConditionExpression: aws.String("followee = :followee AND follower = :follower"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":followee": &ddbtypes.AttributeValueMemberS{Value: followee.String()},
			":follower": &ddbtypes.AttributeValueMemberS{Value: follower.String()},
		},
		Select: ddbtypes.SelectCount,
	}

	result, err := s.db.Client.Query(ctx, isFollowingQueryInput)

	if err != nil {
		return false, fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}
	return result.Count > 0, nil
}

func (s DynamodbFollowerRepository) Follow(ctx context.Context, follower, followee uuid.UUID) error {
	dynamodbFollowerItem := DynamodbFollowerItem{
		Followee: followee.String(),
		Follower: follower.String(),
	}
	followerAttributes, err := attributevalue.MarshalMap(dynamodbFollowerItem)

	if err != nil {
		return fmt.Errorf("%w: %w", errutil.ErrDynamoMapping, err)
	}

	_, err = s.db.Client.PutItem(ctx, &dynamodb.PutItemInput{Item: followerAttributes, TableName: aws.String(followerTable)})

	if err != nil {
		return fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}

	return nil
}

func (s DynamodbFollowerRepository) UnFollow(ctx context.Context, follower, followee uuid.UUID) error {
	dynamodbFollowerItem := DynamodbFollowerItem{
		Follower: follower.String(),
		Followee: followee.String(),
	}
	followerAttributes, err := attributevalue.MarshalMap(dynamodbFollowerItem)

	if err != nil {
		return fmt.Errorf("%w: %w", errutil.ErrDynamoMapping, err)
	}

	// ToDo @ender we can't tell whether something was actually deleted or not.
	// 	It's doable, however, it doesn't seem to be relevant in our case
	_, err = s.db.Client.DeleteItem(ctx, &dynamodb.DeleteItemInput{Key: followerAttributes, TableName: aws.String(followerTable)})

	if err != nil {
		return fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}

	return nil
}

func (s DynamodbFollowerRepository) BatchIsFollowing(ctx context.Context, follower uuid.UUID, followees []uuid.UUID) (map[uuid.UUID]bool, error) {
	// short circuit if followees is empty, no need to query
	// also, dynamodb will throw a validation error if we try to query with empty keys
	if len(followees) == 0 {
		return map[uuid.UUID]bool{}, nil
	}

	keys := make([]map[string]ddbtypes.AttributeValue, 0, len(follower))
	for _, followee := range followees {
		keys = append(keys, map[string]ddbtypes.AttributeValue{
			"follower": &ddbtypes.AttributeValueMemberS{Value: follower.String()},
			"followee": &ddbtypes.AttributeValueMemberS{Value: followee.String()},
		})
	}

	slog.DebugContext(ctx, "IsFollowingBulk keys", slog.Any("keys", keys))

	response, err := s.db.Client.BatchGetItem(ctx, &dynamodb.BatchGetItemInput{
		RequestItems: map[string]ddbtypes.KeysAndAttributes{
			followerTable: {
				Keys: keys,
			},
		},
	})
	if err != nil {
		return map[uuid.UUID]bool{}, fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}

	followersItems := response.Responses[followerTable]
	result := make(map[uuid.UUID]bool, len(followees))
	for _, item := range followersItems {
		dynamodbFollowerItem := DynamodbFollowerItem{}
		err = attributevalue.UnmarshalMap(item, &dynamodbFollowerItem)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", errutil.ErrDynamoMapping, err)
		}
		// ToDo @ender do not use MustParse - it panics if the string is not a valid UUID
		result[uuid.MustParse(dynamodbFollowerItem.Followee)] = true
	}

	return result, nil
}

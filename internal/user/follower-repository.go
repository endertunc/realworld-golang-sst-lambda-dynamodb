package user

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"realworld-aws-lambda-dynamodb-golang/internal/database"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
)

var followerTable = "followers"

type DynamodbFollowerRepository struct {
	db database.DynamoDBStore
}

var _ FollowerRepositoryInterface = (*DynamodbFollowerRepository)(nil)

type DynamodbFollowerItem struct {
	Follower uuid.UUID `dynamodbav:"follower"`
	Followee uuid.UUID `dynamodbav:"followee"`
}

func (s DynamodbFollowerRepository) IsFollowing(c context.Context, follower, followee uuid.UUID) (bool, error) {
	isFollowingQueryInput := &dynamodb.QueryInput{TableName: aws.String(followerTable),
		KeyConditionExpression: aws.String("followee = :followee AND follower = :follower"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":followee": &ddbtypes.AttributeValueMemberS{Value: followee.String()},
			":follower": &ddbtypes.AttributeValueMemberS{Value: follower.String()},
		},
		Select: ddbtypes.SelectCount,
	}

	result, err := s.db.Client.Query(c, isFollowingQueryInput)

	if err != nil {
		return false, errutil.ErrDynamoQuery.Errorf("IsFollowing - query error: %w", err)
	}

	return result.Count > 0, nil
}

func (s DynamodbFollowerRepository) Follow(c context.Context, follower, followee uuid.UUID) error {
	dynamodbFollowerItem := DynamodbFollowerItem{
		Followee: followee,
		Follower: follower,
	}
	followerAttributes, err := attributevalue.MarshalMap(dynamodbFollowerItem)

	if err != nil {
		return errutil.ErrDynamoMapping.Errorf("Follow - mapping error: %w", err)
	}

	_, err = s.db.Client.PutItem(c, &dynamodb.PutItemInput{Item: followerAttributes, TableName: aws.String(followerTable)})

	if err != nil {
		return errutil.ErrDynamoQuery.Errorf("Follow - query error: %w", err)
	}

	return nil
}

func (s DynamodbFollowerRepository) UnFollow(c context.Context, follower, followee uuid.UUID) error {
	dynamodbFollowerItem := DynamodbFollowerItem{
		Follower: follower,
		Followee: followee,
	}
	followerAttributes, err := attributevalue.MarshalMap(dynamodbFollowerItem)

	if err != nil {
		return errutil.ErrDynamoMapping.Errorf("UnFollow - mapping error: %w", err)
	}

	_, err = s.db.Client.DeleteItem(c, &dynamodb.DeleteItemInput{Key: followerAttributes, TableName: aws.String(followerTable)})

	if err != nil {
		return errutil.ErrDynamoQuery.Errorf("UnFollow - query error: %w", err)
	}

	return nil
}

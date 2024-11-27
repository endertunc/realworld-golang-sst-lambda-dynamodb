package repository

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"realworld-aws-lambda-dynamodb-golang/internal/database"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"time"
)

var (
	feedTable           = "feed"
	followerFolloweeGSI = "follower_followee_gsi"
)

type userFeedRepository struct {
	db *database.DynamoDBStore
}

type UserFeedRepositoryInterface interface {
	FanoutArticle(ctx context.Context, articleId, authorId uuid.UUID, createdAt time.Time) error
	FindArticleIdsInUserFeed(ctx context.Context, userId uuid.UUID, limit int, nextPageToken *string) ([]uuid.UUID, *string, error)
}

var _ UserFeedRepositoryInterface = userFeedRepository{} //nolint:golint,exhaustruct

func NewUserFeedRepository(db *database.DynamoDBStore) UserFeedRepositoryInterface {
	return userFeedRepository{db: db}
}

type DynamodbFeedItem struct {
	UserId    DynamodbUUID `dynamodbav:"userId"`    // pk
	CreatedAt int64        `dynamodbav:"createdAt"` // sk
	ArticleId DynamodbUUID `dynamodbav:"articleId"`
	AuthorId  DynamodbUUID `dynamodbav:"authorId"`
}

func (uf userFeedRepository) FanoutArticle(ctx context.Context, articleId, authorId uuid.UUID, createdAt time.Time) error {
	paginator := dynamodb.NewQueryPaginator(uf.db.Client, &dynamodb.QueryInput{
		TableName:              aws.String(followerTable),
		IndexName:              aws.String(followerFolloweeGSI),
		KeyConditionExpression: aws.String("followee = :followee"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":followee": &types.AttributeValueMemberS{Value: authorId.String()},
		},
	})

	// ToDo @ender we should log where the process stops in case of an error
	for paginator.HasMorePages() {
		result, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
		}
		var writeRequests []types.WriteRequest
		for _, item := range result.Items {
			var dynamodbFollowerItem DynamodbFollowerItem
			err = attributevalue.UnmarshalMap(item, &dynamodbFollowerItem)
			if err != nil {
				return fmt.Errorf("%w: %w", errutil.ErrDynamoMapping, err)
			}
			dynamodbFeedItem := DynamodbFeedItem{
				UserId:    dynamodbFollowerItem.Follower,
				CreatedAt: createdAt.UnixMilli(),
				ArticleId: DynamodbUUID(articleId),
				AuthorId:  DynamodbUUID(authorId),
			}
			feedItemAttributes, err := attributevalue.MarshalMap(dynamodbFeedItem)
			if err != nil {
				return fmt.Errorf("%w: %w", errutil.ErrDynamoMarshalling, err)
			}
			writeRequests = append(writeRequests, types.WriteRequest{
				PutRequest: &types.PutRequest{
					Item: feedItemAttributes,
				},
			})
		}
		if len(writeRequests) > 0 {
			_, err = uf.db.Client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
				RequestItems: map[string][]types.WriteRequest{
					feedTable: writeRequests,
				},
			})
			if err != nil {
				return fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
			}
		}
	}
	return nil
}

func (uf userFeedRepository) FindArticleIdsInUserFeed(ctx context.Context, userId uuid.UUID, limit int, nextPageToken *string) ([]uuid.UUID, *string, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(feedTable),
		KeyConditionExpression: aws.String("userId = :userId"),
		Limit:                  aws.Int32(int32(limit)),
		ScanIndexForward:       aws.Bool(false),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":userId": &types.AttributeValueMemberS{Value: userId.String()},
		},
	}

	// decode and set LastEvaluatedKey if nextPageToken is provided
	// ToDo @ender shoul we pass this to QueryMany and handle it there in a single place?
	//  it would be weird to pass QueryInput and nextPageToken to QueryMany separately tho...
	var exclusiveStartKey map[string]types.AttributeValue
	if nextPageToken != nil {
		decodedLastEvaluatedKey, err := decodeLastEvaluatedKey(*nextPageToken)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: %w", errutil.ErrDynamoTokenDecoding, err)
		}
		exclusiveStartKey = decodedLastEvaluatedKey
	}

	// identity mapper to return original DynamodbType without any conversion
	identityMapper := func(item DynamodbFeedItem) DynamodbFeedItem { return item }
	feedItems, lastEvaluatedKey, err := QueryMany(ctx, uf.db.Client, input, limit, exclusiveStartKey, identityMapper)
	if err != nil {
		return nil, nil, err
	}

	// convert to article ids
	articleIds := make([]uuid.UUID, 0, len(feedItems))
	for _, item := range feedItems {
		articleIds = append(articleIds, uuid.UUID(item.ArticleId))
	}

	var newNextPageToken *string
	if len(lastEvaluatedKey) > 0 {
		encodedToken, err := encodeLastEvaluatedKey(lastEvaluatedKey)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: %w", errutil.ErrDynamoTokenEncoding, err)
		}
		newNextPageToken = encodedToken
	}

	return articleIds, newNextPageToken, nil
}

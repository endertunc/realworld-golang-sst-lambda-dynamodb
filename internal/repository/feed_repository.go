package repository

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
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

var _ UserFeedRepositoryInterface = userFeedRepository{}

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
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":followee": &ddbtypes.AttributeValueMemberS{Value: authorId.String()},
		},
	})

	// ToDo @ender we should log where the process stops in case of an error
	for paginator.HasMorePages() {
		result, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
		}
		// ToDo @ender delete me
		//slog.DebugContext(ctx, "fan out article", slog.Int("follower_count", len(result.Items)), slog.Any("result", result))
		var writeRequests []ddbtypes.WriteRequest
		for _, item := range result.Items {
			dynamodbFollowerItem := DynamodbFollowerItem{}
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
			writeRequests = append(writeRequests, ddbtypes.WriteRequest{
				PutRequest: &ddbtypes.PutRequest{
					Item: feedItemAttributes,
				},
			})
		}
		if len(writeRequests) > 0 {
			_, err = uf.db.Client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
				RequestItems: map[string][]ddbtypes.WriteRequest{
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
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":userId": &ddbtypes.AttributeValueMemberS{Value: userId.String()},
		},
	}

	// decode and set LastEvaluatedKey if nextPageToken is provided
	// ToDo @ender shoul we pass this to QueryMany and handle it there in a single place?
	//  it would be weird to pass QueryInput and nextPageToken to QueryMany separately tho...
	if nextPageToken != nil {
		decodedLastEvaluatedKey, err := decodeLastEvaluatedKey(*nextPageToken)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: %w", errutil.ErrDynamoTokenDecoding, err)
		}
		input.ExclusiveStartKey = decodedLastEvaluatedKey
	}

	// identity mapper to return original DynamodbType without any conversion
	identityMapper := func(item DynamodbFeedItem) DynamodbFeedItem { return item }
	feedItems, nextPageToken, err := QueryMany(ctx, uf.db.Client, input, identityMapper)
	if err != nil {
		return nil, nil, err
	}

	// convert to article ids
	articleIds := make([]uuid.UUID, 0, len(feedItems))
	for _, item := range feedItems {
		articleIds = append(articleIds, uuid.UUID(item.ArticleId))
	}

	return articleIds, nextPageToken, nil
}

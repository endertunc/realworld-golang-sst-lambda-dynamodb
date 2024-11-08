package user

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
	"time"
)

var (
	feedTable           = "feed"
	followerFolloweeGSI = "follower_followee_gsi"
)

type UserFeedRepository struct {
	db *database.DynamoDBStore
}

type UserFeedRepositoryInterface interface {
	FanoutArticle(ctx context.Context, articleId, authorId uuid.UUID, createdAt time.Time) error
	FindArticleIdsInUserFeed(ctx context.Context, userId uuid.UUID, limit int) ([]uuid.UUID, error)
}

var _ UserFeedRepositoryInterface = UserFeedRepository{}

func NewUserFeedRepository(db *database.DynamoDBStore) UserFeedRepository {
	return UserFeedRepository{db: db}
}

type DynamodbFeedItem struct {
	UserId    string    `dynamodbav:"userId"`             // pk
	CreatedAt time.Time `dynamodbav:"createdAt,unixtime"` // sk
	ArticleId string    `dynamodbav:"articleId"`
	AuthorId  string    `dynamodbav:"authorId"`
}

func (uf UserFeedRepository) FanoutArticle(ctx context.Context, articleId, authorId uuid.UUID, createdAt time.Time) error {
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
				CreatedAt: createdAt,
				ArticleId: articleId.String(),
				AuthorId:  authorId.String(),
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

func (uf UserFeedRepository) FindArticleIdsInUserFeed(ctx context.Context, userId uuid.UUID, limit int) ([]uuid.UUID, error) {
	paginator := dynamodb.NewQueryPaginator(uf.db.Client, &dynamodb.QueryInput{
		TableName:              aws.String(feedTable),
		KeyConditionExpression: aws.String("userId = :userId"),
		Limit:                  aws.Int32(int32(limit)),
		ScanIndexForward:       aws.Bool(false),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":userId": &ddbtypes.AttributeValueMemberS{Value: userId.String()},
		},
	})

	// ToDo @ender we should log where the process stops in case of an error
	articleIds := make([]uuid.UUID, 0, limit)
	for paginator.HasMorePages() {
		result, err := paginator.NextPage(ctx)

		slog.DebugContext(ctx, "find article ids in user feed",
			slog.Any("LastEvaluatedKey", result.LastEvaluatedKey),
			slog.Any("Count", result.Count),
			slog.Any("ResultMetadata", result.ResultMetadata))

		if err != nil {
			return []uuid.UUID{}, fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
		}
		// ToDo @ender we only need articleId not the whole item.
		dynamodbFeedItems := make([]DynamodbFeedItem, 0, len(result.Items))
		err = attributevalue.UnmarshalListOfMaps(result.Items, &dynamodbFeedItems)
		if err != nil {
			return []uuid.UUID{}, fmt.Errorf("%w: %w", errutil.ErrDynamoMapping, err)
		}
		for _, item := range dynamodbFeedItems {
			articleId, err := uuid.Parse(item.ArticleId)
			if err != nil {
				return []uuid.UUID{}, fmt.Errorf("%w: %w", errutil.ErrDynamoMapping, err)
			}
			articleIds = append(articleIds, articleId)
		}
	}

	return articleIds, nil
}

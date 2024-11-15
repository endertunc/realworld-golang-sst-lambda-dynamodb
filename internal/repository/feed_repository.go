package repository

import (
	"context"
	"encoding/base64"
	"encoding/json"
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
	UserId    string `dynamodbav:"userId"`    // pk
	CreatedAt int64  `dynamodbav:"createdAt"` // sk
	ArticleId string `dynamodbav:"articleId"`
	AuthorId  string `dynamodbav:"authorId"`
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
	if nextPageToken != nil {
		decodedLastEvaluatedKey, err := decodeLastEvaluatedKey(*nextPageToken)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: %w", errutil.ErrDynamoTokenDecoding, err)
		}
		input.ExclusiveStartKey = decodedLastEvaluatedKey
	}

	result, err := uf.db.Client.Query(ctx, input)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}

	slog.DebugContext(ctx, "find article ids in user feed",
		slog.Any("LastEvaluatedKey", result.LastEvaluatedKey),
		slog.Any("Count", result.Count),
		slog.Any("ResultMetadata", result.ResultMetadata))

	// parse feed items
	dynamodbFeedItems := make([]DynamodbFeedItem, 0, len(result.Items))
	err = attributevalue.UnmarshalListOfMaps(result.Items, &dynamodbFeedItems)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", errutil.ErrDynamoMapping, err)
	}

	// convert to article ids
	articleIds := make([]uuid.UUID, 0, len(dynamodbFeedItems))
	for _, item := range dynamodbFeedItems {
		articleId, err := uuid.Parse(item.ArticleId)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: %w", errutil.ErrDynamoMapping, err)
		}
		articleIds = append(articleIds, articleId)
	}

	// prepare next page token if there are more results
	var nextToken *string
	if len(result.LastEvaluatedKey) > 0 {
		encodedToken, err := encodeLastEvaluatedKey(result.LastEvaluatedKey)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: %w", errutil.ErrDynamoTokenEncoding, err)
		}
		nextToken = encodedToken
	}

	return articleIds, nextToken, nil
}

// ToDo move this to commons in the repositories package.
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

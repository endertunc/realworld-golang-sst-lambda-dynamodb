package eventhandler

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"log/slog"
	"realworld-aws-lambda-dynamodb-golang/internal/service"
	"strconv"
	"time"
)

type UserFeedHandler struct {
	UserFeedService service.FeedServiceInterface
}

func NewArticleUserFeedHandler(UserFeedService service.FeedServiceInterface) UserFeedHandler {
	return UserFeedHandler{
		UserFeedService: UserFeedService,
	}
}

func (a UserFeedHandler) HandleEvent(ctx context.Context, event events.DynamoDBEvent) (BatchResult, error) {
	var batchItemFailures []BatchItemFailure
	for _, record := range event.Records {
		if record.EventName == "INSERT" {
			articleId, authorId, createdAt, err := parseDynamoDBEventRecord(ctx, record)
			if err != nil {
				return BatchResult{}, err
			}
			err = a.UserFeedService.FanoutArticle(ctx, articleId, authorId, createdAt)
			if err != nil {
				slog.DebugContext(ctx, "error while fanning out article", slog.Any("error", err))
				batchItemFailures = append(batchItemFailures, BatchItemFailure{
					ItemIdentifier: record.Change.SequenceNumber,
				})
				continue
			}
		}
	}
	return BatchResult{
		BatchItemFailures: batchItemFailures,
	}, nil

}

func parseDynamoDBEventRecord(ctx context.Context, record events.DynamoDBEventRecord) (uuid.UUID, uuid.UUID, time.Time, error) {
	slog.DebugContext(ctx, "Processing DynamoDB event record", slog.Any("record", record))
	articleId, err := uuid.Parse(record.Change.NewImage["pk"].String())
	if err != nil {
		return uuid.Nil, uuid.Nil, time.Time{}, err
	}
	authorId, err := uuid.Parse(record.Change.NewImage["authorId"].String())

	if err != nil {
		return uuid.Nil, uuid.Nil, time.Time{}, err
	}

	createdAt, err := decodeUnixTime(record.Change.NewImage["createdAt"].Number())
	if err != nil {
		return uuid.Nil, uuid.Nil, time.Time{}, err
	}
	return articleId, authorId, createdAt, nil
}

func decodeUnixTime(n string) (time.Time, error) {
	v, err := strconv.ParseInt(n, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.UnixMilli(v), nil
}

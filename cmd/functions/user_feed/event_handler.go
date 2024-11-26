package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/uuid"
	"log/slog"
	"realworld-aws-lambda-dynamodb-golang/cmd/functions"
	"strconv"
	"time"
)

// https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/example_serverless_DynamoDB_Lambda_batch_item_failures_section.html
type BatchItemFailure struct {
	ItemIdentifier string `json:"ItemIdentifier"`
}

type BatchResult struct {
	BatchItemFailures []BatchItemFailure `json:"BatchItemFailures"`
}

func handleRequest(ctx context.Context, event events.DynamoDBEvent) (BatchResult, error) {
	var batchItemFailures []BatchItemFailure
	for _, record := range event.Records {
		if record.EventName == "INSERT" {

			articleId, authorId, createdAt, err := parseDynamoDBEventRecord(record)
			if err != nil {
				return BatchResult{}, err
			}
			err = functions.UserFeedService.FanoutArticle(ctx, articleId, authorId, createdAt)
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

// ToDo @ender move to somewhere else...
func parseDynamoDBEventRecord(record events.DynamoDBEventRecord) (uuid.UUID, uuid.UUID, time.Time, error) {
	slog.DebugContext(context.Background(), "Processing DynamoDB event record", slog.Any("record", record))
	// ToDo @ender why somewhere we use pk and other places regular field name...
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

func main() {
	lambda.Start(handleRequest)
}

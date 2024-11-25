package repository

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"realworld-aws-lambda-dynamodb-golang/internal/database"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/domain/generator"
	"realworld-aws-lambda-dynamodb-golang/internal/test"
	"testing"
	"time"
)

var db = database.DynamoDBStore{Client: test.DynamodbClient()}

// ToDo @ender think if we can prove that internal pagination happens

// it uses comment table as a test table
func TestBatchGetItemsWithUnprocessedKeys(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		ctx := context.Background()
		comments := populateTableWithComments(t, ctx, 20, 350, nil)

		// prepare keys for batch get items
		keys := make([]map[string]types.AttributeValue, 0, len(comments))
		for _, comment := range comments {
			keys = append(keys, map[string]types.AttributeValue{
				"commentId": &types.AttributeValueMemberS{Value: comment.Id.String()},
				"articleId": &types.AttributeValueMemberS{Value: comment.ArticleId.String()},
			})
		}
		assert.EventuallyWithT(t, func(testingT *assert.CollectT) {
			// batch get items
			dynamodbCommentItems, err := BatchGetItems(ctx, db.Client, commentTable, keys, func(item DynamodbCommentItem) DynamodbCommentItem {
				return item
			})
			if err != nil {
				t.Fatalf("failed to batch get items: %v", err)
			}

			// assert that all items are fetched
			assert.Equal(t, len(comments), len(dynamodbCommentItems))
		}, 5*time.Second, 1*time.Second)

	})
}

// it uses comment table as a test table
func TestQueryManyWithInternalPagination(t *testing.T) {
	test.WithSetupAndTeardown(t, func() {
		ctx := context.Background()
		articleId := uuid.New()
		populateTableWithComments(t, ctx, 20, 350, &articleId)

		// prepare a query to fetch comments by articleId
		input := &dynamodb.QueryInput{
			TableName:              &commentTable,
			IndexName:              &commentArticleGSI,
			KeyConditionExpression: aws.String("articleId = :articleId"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":articleId": &types.AttributeValueMemberS{Value: articleId.String()},
			},
		}

		// wait for eventual consistency since we are querying by GSI
		assert.EventuallyWithT(t, func(testingT *assert.CollectT) {
			// query many with internal pagination
			// 380 kb item size and 10 items per page proven to be enough to trigger pagination several times
			dynamodbCommentItemsPageOne, lastEvaluatedKeyPageOne, err := QueryMany(ctx, db.Client, input, 10, nil, func(item DynamodbCommentItem) DynamodbCommentItem {
				return item
			})
			if err != nil {
				t.Fatalf("failed to query many: %v", err)
			}

			// assert that the first page is full fetched and lastEvaluatedKey is NOT empty
			assert.Equal(testingT, 10, len(dynamodbCommentItemsPageOne))
			assert.NotEmpty(testingT, lastEvaluatedKeyPageOne)

			dynamodbCommentItemsPageTwo, lastEvaluatedKeyPageTwo, err := QueryMany(ctx, db.Client, input, 20, lastEvaluatedKeyPageOne, func(item DynamodbCommentItem) DynamodbCommentItem {
				return item
			})
			// assert that the second page is full fetched and lastEvaluatedKey is empty
			assert.Equal(testingT, 10, len(dynamodbCommentItemsPageTwo))
			assert.Empty(testingT, lastEvaluatedKeyPageTwo)
		}, 5*time.Second, 1*time.Second)
	})
}

func populateTableWithComments(t *testing.T, ctx context.Context, count int, itemSizeInKb int, articleIdOverride *uuid.UUID) []domain.Comment {
	comments := make([]domain.Comment, 0, count)
	for i := 0; i < count; i++ {
		comment := generator.GenerateComment()
		comment.Body = gofakeit.LetterN(uint(itemSizeInKb * 1024))
		if articleIdOverride != nil {
			comment.ArticleId = *articleIdOverride
		}
		comments = append(comments, comment)
	}

	var writeRequests []types.WriteRequest
	for _, comment := range comments {
		dynamodbCommentItem := toDynamodbCommentItem(comment)
		commentItemAttributes, err := attributevalue.MarshalMap(dynamodbCommentItem)
		if err != nil {
			t.Fatalf("failed to marshal map: %v", err)
		}
		writeRequests = append(writeRequests, types.WriteRequest{
			PutRequest: &types.PutRequest{
				Item: commentItemAttributes,
			},
		})
	}

	_, err := db.Client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			commentTable: writeRequests,
		},
	})

	if err != nil {
		t.Fatalf("failed to batch write item: %v", err)
	}

	return comments
}
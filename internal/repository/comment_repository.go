package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"realworld-aws-lambda-dynamodb-golang/internal/database"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"time"
)

type dynamodbCommentRepository struct {
	db *database.DynamoDBStore
}

type CommentRepositoryInterface interface {
	DeleteCommentByArticleIdAndCommentId(ctx context.Context, articleId uuid.UUID, commentId uuid.UUID) error
	FindCommentsByArticleId(ctx context.Context, articleId uuid.UUID) ([]domain.Comment, error)
	CreateComment(ctx context.Context, comment domain.Comment) error
	FindCommentByCommentIdAndArticleId(ctx context.Context, commentId, articleId uuid.UUID) (domain.Comment, error)
}

var _ CommentRepositoryInterface = dynamodbCommentRepository{} //nolint:golint,exhaustruct

func NewDynamodbCommentRepository(db *database.DynamoDBStore) CommentRepositoryInterface {
	return dynamodbCommentRepository{db: db}
}

var (
	commentTable      = "comment"
	commentArticleGSI = "comment_article_gsi"
)

type DynamodbCommentItem struct {
	Id        DynamodbUUID `dynamodbav:"commentId"`
	ArticleId DynamodbUUID `dynamodbav:"articleId"`
	AuthorId  DynamodbUUID `dynamodbav:"authorId"`
	Body      string       `dynamodbav:"body"`
	CreatedAt int64        `dynamodbav:"createdAt"`
	UpdatedAt int64        `dynamodbav:"updatedAt"`
}

func (c dynamodbCommentRepository) DeleteCommentByArticleIdAndCommentId(ctx context.Context, articleId uuid.UUID, commentId uuid.UUID) error {
	input := &dynamodb.DeleteItemInput{
		TableName: &commentTable,
		Key: map[string]types.AttributeValue{
			"commentId": &types.AttributeValueMemberS{Value: commentId.String()},
			"articleId": &types.AttributeValueMemberS{Value: articleId.String()},
		},
	}

	_, err := c.db.Client.DeleteItem(ctx, input)
	if err != nil {
		return fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}

	return nil
}

// the API specs that this project is based on using a bad design IMHO
// therefore, I will add pagination and sort result by creation date like we do with other entities
func (c dynamodbCommentRepository) FindCommentsByArticleId(ctx context.Context, articleId uuid.UUID) ([]domain.Comment, error) {
	input := &dynamodb.QueryInput{
		TableName:              &commentTable,
		IndexName:              &commentArticleGSI,
		KeyConditionExpression: aws.String("articleId = :articleId"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":articleId": &types.AttributeValueMemberS{Value: articleId.String()},
		},
	}

	comments, _, err := QueryMany(ctx, c.db.Client, input, 10, nil, toDomainComment)
	return comments, err
}

func (c dynamodbCommentRepository) CreateComment(ctx context.Context, comment domain.Comment) error {
	dynamodbCommentItem := toDynamodbCommentItem(comment)
	commentAttributes, err := attributevalue.MarshalMap(dynamodbCommentItem)
	if err != nil {
		return fmt.Errorf("%w: %w", errutil.ErrDynamoMarshalling, err)
	}

	input := &dynamodb.PutItemInput{
		TableName: &commentTable,
		Item:      commentAttributes,
	}

	_, err = c.db.Client.PutItem(ctx, input)
	if err != nil {
		return fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}

	return nil
}

func (c dynamodbCommentRepository) FindCommentByCommentIdAndArticleId(ctx context.Context, commentId, articleId uuid.UUID) (domain.Comment, error) {
	input := &dynamodb.GetItemInput{
		TableName: &commentTable,
		Key: map[string]types.AttributeValue{
			"commentId": &types.AttributeValueMemberS{Value: commentId.String()},
			"articleId": &types.AttributeValueMemberS{Value: articleId.String()},
		},
	}
	comment, err := GetItem(ctx, c.db.Client, input, toDomainComment)
	if err != nil {
		if errors.Is(err, ErrDynamodbItemNotFound) {
			return domain.Comment{}, errutil.ErrCommentNotFound
		}
		return domain.Comment{}, err
	}
	return comment, nil
}

func toDynamodbCommentItem(article domain.Comment) DynamodbCommentItem {
	return DynamodbCommentItem{
		Id:        DynamodbUUID(article.Id),
		ArticleId: DynamodbUUID(article.ArticleId),
		AuthorId:  DynamodbUUID(article.AuthorId),
		Body:      article.Body,
		CreatedAt: article.CreatedAt.UnixMilli(),
		UpdatedAt: article.UpdatedAt.UnixMilli(),
	}
}

func toDomainComment(comment DynamodbCommentItem) domain.Comment {
	return domain.Comment{
		Id:        uuid.UUID(comment.Id),
		ArticleId: uuid.UUID(comment.ArticleId),
		AuthorId:  uuid.UUID(comment.AuthorId),
		Body:      comment.Body,
		CreatedAt: time.UnixMilli(comment.CreatedAt),
		UpdatedAt: time.UnixMilli(comment.UpdatedAt),
	}
}

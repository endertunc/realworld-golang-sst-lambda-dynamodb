package repository

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/samber/oops"
	veqrynslog "github.com/veqryn/slog-context/http"
	"log/slog"
	"realworld-aws-lambda-dynamodb-golang/internal/database"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"time"
)

type dynamodbCommentRepository struct {
	db *database.DynamoDBStore
}

type CommentRepositoryInterface interface {
	DeleteCommentByArticleIdAndCommentId(ctx context.Context, loggedInUserId uuid.UUID, articleId uuid.UUID, commentId uuid.UUID) error
	FindCommentsByArticleId(ctx context.Context, articleId uuid.UUID) ([]domain.Comment, error)
	CreateComment(ctx context.Context, comment domain.Comment) error
	FindCommentByCommentIdAndArticleId(ctx context.Context, commentId, articleId uuid.UUID) (domain.Comment, error)
}

var _ CommentRepositoryInterface = dynamodbCommentRepository{}

func NewDynamodbCommentRepository(db *database.DynamoDBStore) CommentRepositoryInterface {
	return dynamodbCommentRepository{db: db}
}

var (
	commentTable      = "comment"
	commentArticleGSI = "comment_article_gsi"
)

type DynamodbCommentItem struct {
	Id        string `dynamodbav:"commentId"`
	ArticleId string `dynamodbav:"articleId"`
	AuthorId  string `dynamodbav:"authorId"`
	Body      string `dynamodbav:"body"`
	CreatedAt int64  `dynamodbav:"createdAt"`
	UpdatedAt int64  `dynamodbav:"updatedAt"`
}

// ToDo @ender delete only existing comments???
// ToDo @ender loggedin user id is not used
func (c dynamodbCommentRepository) DeleteCommentByArticleIdAndCommentId(ctx context.Context, loggedInUserId uuid.UUID, articleId uuid.UUID, commentId uuid.UUID) error {
	input := &dynamodb.DeleteItemInput{
		TableName: &commentTable,
		Key: map[string]ddbtypes.AttributeValue{
			"commentId": &ddbtypes.AttributeValueMemberS{Value: commentId.String()},
			"articleId": &ddbtypes.AttributeValueMemberS{Value: articleId.String()},
		},
	}

	_, err := c.db.Client.DeleteItem(ctx, input)
	if err != nil {
		return fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}

	return nil
}

/*
 * ToDo Ender by design, dynamodb returns item in any order,
 *  it's not necessary in our case but we could sort the comments by createdAt field.
 */

func (c dynamodbCommentRepository) FindCommentsByArticleId(ctx context.Context, articleId uuid.UUID) ([]domain.Comment, error) {
	input := &dynamodb.QueryInput{
		TableName:              &commentTable,
		IndexName:              &commentArticleGSI,
		KeyConditionExpression: aws.String("articleId = :articleId"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":articleId": &ddbtypes.AttributeValueMemberS{Value: articleId.String()},
		},
	}

	result, err := c.db.Client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}

	dynamodbCommentItems := make([]DynamodbCommentItem, 0, len(result.Items))
	err = attributevalue.UnmarshalListOfMaps(result.Items, &dynamodbCommentItems)
	if err != nil {
		// ToDo @ender experiment with this oops library
		fancyError := oops.Wrap(fmt.Errorf("%w: %w", errutil.ErrDynamoMarshalling, err))
		veqrynslog.With(ctx, "regularError", fmt.Errorf("%w: %w", errutil.ErrDynamoMarshalling, err))
		veqrynslog.With(ctx, "fancyError", fancyError)
		veqrynslog.With(ctx, slog.Group("errorContext", slog.String("articleId", articleId.String())))
		return nil, fmt.Errorf("%w: %w", errutil.ErrDynamoMarshalling, err)
	}

	return lo.Map(dynamodbCommentItems, func(comment DynamodbCommentItem, _ int) domain.Comment {
		return toDomainComment(comment)
	}), nil
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
		Key: map[string]ddbtypes.AttributeValue{
			"commentId": &ddbtypes.AttributeValueMemberS{Value: commentId.String()},
			"articleId": &ddbtypes.AttributeValueMemberS{Value: articleId.String()},
		},
	}

	result, err := c.db.Client.GetItem(ctx, input)
	if err != nil {
		return domain.Comment{}, fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}

	if result.Item == nil {
		return domain.Comment{}, errutil.ErrCommentNotFound
	}

	dynamodbCommentItem := DynamodbCommentItem{}
	err = attributevalue.UnmarshalMap(result.Item, &dynamodbCommentItem)
	if err != nil {
		return domain.Comment{}, fmt.Errorf("%w: %w", errutil.ErrDynamoMapping, err)
	}

	return toDomainComment(dynamodbCommentItem), nil
}

func toDynamodbCommentItem(article domain.Comment) DynamodbCommentItem {
	return DynamodbCommentItem{
		Id:        article.Id.String(),
		ArticleId: article.ArticleId.String(),
		AuthorId:  article.AuthorId.String(),
		Body:      article.Body,
		CreatedAt: article.CreatedAt.UnixMilli(),
		UpdatedAt: article.UpdatedAt.UnixMilli(),
	}
}

func toDomainComment(comment DynamodbCommentItem) domain.Comment {
	return domain.Comment{
		Id:        uuid.MustParse(comment.Id),
		ArticleId: uuid.MustParse(comment.ArticleId),
		AuthorId:  uuid.MustParse(comment.AuthorId),
		Body:      comment.Body,
		CreatedAt: time.UnixMilli(comment.CreatedAt),
		UpdatedAt: time.UnixMilli(comment.UpdatedAt),
	}
}
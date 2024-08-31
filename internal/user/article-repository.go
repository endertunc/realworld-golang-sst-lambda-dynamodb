package user

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"realworld-aws-lambda-dynamodb-golang/internal/database"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
)

type DynamodbArticleRepository struct {
	db database.DynamoDBStore
}

var _ ArticleRepositoryInterface = DynamodbArticleRepository{}

var articleTable = aws.String("article")
var commentTable = aws.String("commentTable")
var favoriteTable = aws.String("favorite")

var articleSlugGSI = aws.String("article_slug_gsi")
var commentArticleGSI = aws.String("comment_article_gsi")

type DynamodbFavoriteArticleItem struct {
	UserId    uuid.UUID `dynamodbav:"userId"`
	ArticleId uuid.UUID `dynamodbav:"articleId"`
}

func (d DynamodbArticleRepository) FindArticleBySlug(c context.Context, slug string) (domain.Article, error) {
	input := &dynamodb.QueryInput{
		TableName:              articleTable,
		IndexName:              articleSlugGSI,
		KeyConditionExpression: aws.String("slug = :slug"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":slug": &ddbtypes.AttributeValueMemberS{Value: slug},
		},
	}

	result, err := d.db.Client.Query(c, input)
	if err != nil {
		return domain.Article{}, errutil.ErrDynamoQuery.Errorf("FindArticleBySlug - query error: %w", err)
	}

	if len(result.Items) == 0 {
		return domain.Article{}, fmt.Errorf("article not found")
	}

	var article domain.Article
	err = attributevalue.UnmarshalMap(result.Items[0], &article)
	if err != nil {
		return domain.Article{}, errutil.ErrDynamoQuery.Errorf("FindArticleBySlug - unmarshal error: %w", err)
	}

	return article, nil
}

func (d DynamodbArticleRepository) FindArticleById(c context.Context, articleId uuid.UUID) (domain.Article, error) {
	input := &dynamodb.GetItemInput{
		TableName: articleTable,
		Key: map[string]ddbtypes.AttributeValue{
			"pk": &ddbtypes.AttributeValueMemberS{Value: articleId.String()},
		},
	}

	result, err := d.db.Client.GetItem(c, input)
	if err != nil {
		return domain.Article{}, errutil.ErrDynamoQuery.Errorf("FindArticleById - get item error: %w", err)
	}

	if result.Item == nil {
		return domain.Article{}, fmt.Errorf("article not found")
	}

	var article domain.Article
	err = attributevalue.UnmarshalMap(result.Item, &article)
	if err != nil {
		return domain.Article{}, errutil.ErrDynamoQuery.Errorf("FindArticleById - unmarshal error: %w", err)
	}

	return article, nil
}

// ToDo there can only be one article with the same slug
func (d DynamodbArticleRepository) CreateArticle(c context.Context, article domain.Article) (domain.Article, error) {
	articleAttributes, err := attributevalue.MarshalMap(article)
	if err != nil {
		return domain.Article{}, errutil.ErrDynamoQuery.Errorf("CreateArticle - marshal error: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: articleTable,
		Item:      articleAttributes,
	}

	_, err = d.db.Client.PutItem(c, input)
	if err != nil {
		return domain.Article{}, errutil.ErrDynamoQuery.Errorf("CreateArticle - put item error: %w", err)
	}

	return article, nil
}

func (d DynamodbArticleRepository) DeleteArticleById(c context.Context, articleId uuid.UUID) error {
	input := &dynamodb.DeleteItemInput{
		Key: map[string]ddbtypes.AttributeValue{
			"pk": &ddbtypes.AttributeValueMemberS{Value: articleId.String()},
		},
		TableName: aws.String("article"),
	}

	_, err := d.db.Client.DeleteItem(c, input)
	if err != nil {
		return errutil.ErrDynamoQuery.Errorf("DeleteArticleById - delete item error: %w", err)
	}

	return nil
}

func (d DynamodbArticleRepository) DeleteCommentByArticleIdAndCommentId(c context.Context, loggedInUserId uuid.UUID, articleId uuid.UUID, commentId uuid.UUID) error {
	input := &dynamodb.DeleteItemInput{
		TableName: commentTable,
		Key: map[string]ddbtypes.AttributeValue{
			"commentId": &ddbtypes.AttributeValueMemberS{Value: commentId.String()},
			"articleId": &ddbtypes.AttributeValueMemberS{Value: articleId.String()},
		},
	}

	_, err := d.db.Client.DeleteItem(c, input)
	if err != nil {
		return errutil.ErrDynamoQuery.Errorf("DeleteCommentByArticleIdAndCommentId - delete item error: %w", err)
	}

	return nil
}

func (d DynamodbArticleRepository) GetCommentsByArticleId(c context.Context, articleId uuid.UUID) ([]domain.Comment, error) {

	input := &dynamodb.QueryInput{
		TableName:              commentTable,
		IndexName:              commentArticleGSI,
		KeyConditionExpression: aws.String("articleId = :articleId"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":articleId": &ddbtypes.AttributeValueMemberS{Value: articleId.String()},
		},
	}

	result, err := d.db.Client.Query(c, input)
	if err != nil {
		return nil, errutil.ErrDynamoQuery.Errorf("GetCommentsByArticleId - query error: %w", err)
	}

	var comments []domain.Comment
	for _, item := range result.Items {
		var comment domain.Comment
		err = attributevalue.UnmarshalMap(item, &comment)
		if err != nil {
			return nil, errutil.ErrDynamoQuery.Errorf("GetCommentsByArticleId - unmarshal error: %w", err)
		}
		comments = append(comments, comment)
	}

	return comments, nil
}

func (d DynamodbArticleRepository) CreateComment(c context.Context, comment domain.Comment) error {
	commentAttributes, err := attributevalue.MarshalMap(comment)
	if err != nil {
		return errutil.ErrDynamoQuery.Errorf("CreateComment - marshal error: %w", err)
	}

	input := &dynamodb.PutItemInput{
		Item:      commentAttributes,
		TableName: commentTable,
	}

	_, err = d.db.Client.PutItem(c, input)
	if err != nil {
		return errutil.ErrDynamoQuery.Errorf("CreateComment - put item error: %w", err)
	}

	return nil
}

func (d DynamodbArticleRepository) UnfavoriteArticle(c context.Context, loggedInUserId uuid.UUID, articleId uuid.UUID) error {
	input := &dynamodb.DeleteItemInput{
		TableName: favoriteTable,
		Key: map[string]ddbtypes.AttributeValue{
			"userId":    &ddbtypes.AttributeValueMemberS{Value: loggedInUserId.String()},
			"articleId": &ddbtypes.AttributeValueMemberS{Value: articleId.String()},
		},
	}

	_, err := d.db.Client.DeleteItem(c, input)
	if err != nil {
		return errutil.ErrDynamoQuery.Errorf("UnfavoriteArticle - delete item error: %w", err)
	}

	return nil
}

func (d DynamodbArticleRepository) FavoriteArticle(c context.Context, loggedInUserId uuid.UUID, articleId uuid.UUID) error {
	favorite := DynamodbFavoriteArticleItem{
		UserId:    loggedInUserId,
		ArticleId: articleId,
	}

	favoriteArticleAttributes, err := attributevalue.MarshalMap(favorite)
	if err != nil {
		return errutil.ErrDynamoQuery.Errorf("FavoriteArticle - marshal error: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: favoriteTable,
		Item:      favoriteArticleAttributes,
	}

	_, err = d.db.Client.PutItem(c, input)
	if err != nil {
		return errutil.ErrDynamoQuery.Errorf("FavoriteArticle - put item error: %w", err)
	}

	return nil
}

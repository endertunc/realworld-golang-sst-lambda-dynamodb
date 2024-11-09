package repository

import (
	"context"
	"errors"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"realworld-aws-lambda-dynamodb-golang/internal/database"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

type DynamodbArticleRepository struct {
	db *database.DynamoDBStore
}

func NewDynamodbArticleRepository(db *database.DynamoDBStore) DynamodbArticleRepository {
	return DynamodbArticleRepository{db: db}
}

type ArticleRepositoryInterface interface {
	FindArticleBySlug(ctx context.Context, email string) (domain.Article, error)
	FindArticleById(ctx context.Context, articleId uuid.UUID) (domain.Article, error)
	FindArticlesByIds(ctx context.Context, articleIds []uuid.UUID) ([]domain.Article, error)
	CreateArticle(ctx context.Context, article domain.Article) (domain.Article, error)
	DeleteArticleById(ctx context.Context, articleId uuid.UUID) error
	UnfavoriteArticle(ctx context.Context, loggedInUserId uuid.UUID, articleId uuid.UUID) error
	FavoriteArticle(ctx context.Context, loggedInUserId uuid.UUID, articleId uuid.UUID) error
	DeleteCommentByArticleIdAndCommentId(ctx context.Context, loggedInUserId uuid.UUID, articleId uuid.UUID, commentId uuid.UUID) error
	FindCommentsByArticleId(ctx context.Context, articleId uuid.UUID) ([]domain.Comment, error)
	CreateComment(ctx context.Context, comment domain.Comment) error
	FindCommentByCommentIdAndArticleId(ctx context.Context, commentId, articleId uuid.UUID) (domain.Comment, error)
	IsFavorited(ctx context.Context, articleId, userId uuid.UUID) (bool, error)
	IsFavoritedBulk(ctx context.Context, userId uuid.UUID, articleIds []uuid.UUID) (mapset.Set[uuid.UUID], error)
}

var _ ArticleRepositoryInterface = DynamodbArticleRepository{}

type DynamodbArticleItem struct {
	Id             string   `dynamodbav:"pk"`
	Title          string   `dynamodbav:"title"`
	Slug           string   `dynamodbav:"slug"`
	Description    string   `dynamodbav:"description"`
	Body           string   `dynamodbav:"body"`
	TagList        []string `dynamodbav:"tagList"`
	FavoritesCount int      `dynamodbav:"favoritesCount"`
	AuthorId       string   `dynamodbav:"authorId"`
	// ToDo @ender should we convert everything to milliseconds precision?
	CreatedAt int64 `dynamodbav:"createdAt"`
	UpdatedAt int64 `dynamodbav:"updatedAt"`
}

type DynamodbCommentItem struct {
	Id        string    `dynamodbav:"commentId"`
	ArticleId string    `dynamodbav:"articleId"`
	AuthorId  string    `dynamodbav:"authorId"`
	Body      string    `dynamodbav:"body"`
	CreatedAt time.Time `dynamodbav:"createdAt,unixtime"`
	UpdatedAt time.Time `dynamodbav:"updatedAt,unixtime"`
}

var articleTable = "article"
var commentTable = "comment"
var favoriteTable = "favorite"

var articleSlugGSI = aws.String("article_slug_gsi")
var commentArticleGSI = aws.String("comment_article_gsi")

type DynamodbFavoriteArticleItem struct {
	UserId    string `dynamodbav:"userId"`
	ArticleId string `dynamodbav:"articleId"`
}

func (d DynamodbArticleRepository) FindArticleBySlug(c context.Context, slug string) (domain.Article, error) {
	input := &dynamodb.QueryInput{
		TableName:              &articleTable,
		IndexName:              articleSlugGSI,
		KeyConditionExpression: aws.String("slug = :slug"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":slug": &ddbtypes.AttributeValueMemberS{Value: slug},
		},
		Select: ddbtypes.SelectAllAttributes,
	}

	result, err := d.db.Client.Query(c, input)
	if err != nil {
		return domain.Article{}, fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}

	if len(result.Items) == 0 {
		return domain.Article{}, errutil.ErrArticleNotFound
	}

	var dynamodbArticle DynamodbArticleItem
	err = attributevalue.UnmarshalMap(result.Items[0], &dynamodbArticle)
	if err != nil {
		return domain.Article{}, fmt.Errorf("%w: %w", errutil.ErrDynamoMarshalling, err)
	}

	domainArticle := toDomainArticle(dynamodbArticle)

	return domainArticle, nil
}

func (d DynamodbArticleRepository) FindArticleById(c context.Context, articleId uuid.UUID) (domain.Article, error) {
	input := &dynamodb.GetItemInput{
		TableName: &articleTable,
		Key: map[string]ddbtypes.AttributeValue{
			"pk": &ddbtypes.AttributeValueMemberS{Value: articleId.String()},
		},
	}

	result, err := d.db.Client.GetItem(c, input)
	if err != nil {
		return domain.Article{}, fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}

	if result.Item == nil {
		return domain.Article{}, errutil.ErrArticleNotFound
	}

	var article domain.Article
	err = attributevalue.UnmarshalMap(result.Item, &article)
	if err != nil {
		return domain.Article{}, fmt.Errorf("%w: %w", errutil.ErrDynamoMapping, err)
	}

	return article, nil
}

// ToDo there can only be one article with the same slug
func (d DynamodbArticleRepository) CreateArticle(ctx context.Context, article domain.Article) (domain.Article, error) {
	dynamodbArticleItem := toDynamodbArticleItem(article)
	articleAttributes, err := attributevalue.MarshalMap(dynamodbArticleItem)
	if err != nil {
		return domain.Article{}, fmt.Errorf("%w: %w", errutil.ErrDynamoMarshalling, err)
	}

	input := &dynamodb.PutItemInput{
		TableName: &articleTable,
		Item:      articleAttributes,
	}

	// ToDo @ender check if we can use the returned article from the put item?
	_, err = d.db.Client.PutItem(ctx, input)
	if err != nil {
		return domain.Article{}, fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
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
		return fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}

	return nil
}

// ToDo @ender delete only existing comments???
// ToDo @ender loggedin user id is not used
func (d DynamodbArticleRepository) DeleteCommentByArticleIdAndCommentId(c context.Context, loggedInUserId uuid.UUID, articleId uuid.UUID, commentId uuid.UUID) error {
	input := &dynamodb.DeleteItemInput{
		TableName: &commentTable,
		Key: map[string]ddbtypes.AttributeValue{
			"commentId": &ddbtypes.AttributeValueMemberS{Value: commentId.String()},
			"articleId": &ddbtypes.AttributeValueMemberS{Value: articleId.String()},
		},
	}

	_, err := d.db.Client.DeleteItem(c, input)
	if err != nil {
		return fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}

	return nil
}

/*
 * ToDo Ender by design, dynamodb returns item in any order,
 *  it's not necessary in our case but we could sort the comments by createdAt field.
 */

func (d DynamodbArticleRepository) FindCommentsByArticleId(c context.Context, articleId uuid.UUID) ([]domain.Comment, error) {

	input := &dynamodb.QueryInput{
		TableName:              &commentTable,
		IndexName:              commentArticleGSI,
		KeyConditionExpression: aws.String("articleId = :articleId"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":articleId": &ddbtypes.AttributeValueMemberS{Value: articleId.String()},
		},
	}

	result, err := d.db.Client.Query(c, input)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}

	comments := make([]domain.Comment, 0, len(result.Items))
	err = attributevalue.UnmarshalListOfMaps(result.Items, &comments)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errutil.ErrDynamoMarshalling, err)
	}

	return comments, nil
}

func (d DynamodbArticleRepository) CreateComment(ctx context.Context, comment domain.Comment) error {
	dynamodbCommentItem := toDynamodbCommentItem(comment)
	commentAttributes, err := attributevalue.MarshalMap(dynamodbCommentItem)
	if err != nil {
		return fmt.Errorf("%w: %w", errutil.ErrDynamoMarshalling, err)
	}

	input := &dynamodb.PutItemInput{
		TableName: &commentTable,
		Item:      commentAttributes,
	}

	_, err = d.db.Client.PutItem(ctx, input)
	if err != nil {
		return fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}

	return nil
}

// UnfavoriteArticle deletes the favorite item from the favorite table and decrements the favoritesCount of the article
// if the favorite item does not exist, it does not decrement the favoritesCount and returns an ErrAlreadyUnfavorited error
func (d DynamodbArticleRepository) UnfavoriteArticle(ctx context.Context, loggedInUserId uuid.UUID, articleId uuid.UUID) error {
	transactWriteItems := dynamodb.TransactWriteItemsInput{
		TransactItems: []ddbtypes.TransactWriteItem{
			{
				Delete: &ddbtypes.Delete{
					TableName: &favoriteTable,
					Key: map[string]ddbtypes.AttributeValue{
						"userId":    &ddbtypes.AttributeValueMemberS{Value: loggedInUserId.String()},
						"articleId": &ddbtypes.AttributeValueMemberS{Value: articleId.String()},
					},
					ConditionExpression: aws.String("attribute_exists(userId) AND attribute_exists(articleId)"),
				},
			},
			{
				Update: &ddbtypes.Update{
					TableName: &articleTable,
					Key: map[string]ddbtypes.AttributeValue{
						"pk": &ddbtypes.AttributeValueMemberS{Value: articleId.String()},
					},
					UpdateExpression: aws.String("SET favoritesCount = favoritesCount - :dec"),
					ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
						":dec": &ddbtypes.AttributeValueMemberN{Value: "1"},
					},
				},
			},
		},
	}

	_, err := d.db.Client.TransactWriteItems(ctx, &transactWriteItems, func(o *dynamodb.Options) {
		o.RetryMaxAttempts = 1 // we don't want to retry this operation due to the favoritesCount decrement
	})
	if err != nil {
		var transactionCanceledErr *ddbtypes.TransactionCanceledException
		if errors.As(err, &transactionCanceledErr) {
			for index, reason := range transactionCanceledErr.CancellationReasons {
				if reason.Code != nil && *reason.Code == conditionalCheckFailed && index == 0 {
					return fmt.Errorf("%w: %w", errutil.ErrAlreadyUnfavorited, err)
				}
			}
		}
		return fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}

	return nil
}

// FavoriteArticle creates a favorite item in the favorite table and increments the favoritesCount of the article
// if the favorite item already exists, it does not increment the favoritesCount and returns an ErrAlreadyFavorited error
func (d DynamodbArticleRepository) FavoriteArticle(ctx context.Context, loggedInUserId uuid.UUID, articleId uuid.UUID) error {
	favorite := DynamodbFavoriteArticleItem{
		UserId:    loggedInUserId.String(),
		ArticleId: articleId.String(),
	}

	favoriteArticleAttributes, err := attributevalue.MarshalMap(favorite)
	if err != nil {
		return fmt.Errorf("%w: %w", errutil.ErrDynamoMarshalling, err)
	}

	transactWriteItems := dynamodb.TransactWriteItemsInput{
		TransactItems: []ddbtypes.TransactWriteItem{
			{
				Put: &ddbtypes.Put{
					TableName:           &favoriteTable,
					Item:                favoriteArticleAttributes,
					ConditionExpression: aws.String("attribute_not_exists(userId) AND attribute_not_exists(articleId)"),
				},
			},
			{
				Update: &ddbtypes.Update{
					TableName: &articleTable,
					Key: map[string]ddbtypes.AttributeValue{
						"pk": &ddbtypes.AttributeValueMemberS{Value: articleId.String()},
					},
					UpdateExpression: aws.String("SET favoritesCount = favoritesCount + :inc"),
					ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
						":inc": &ddbtypes.AttributeValueMemberN{Value: "1"},
					},
				},
			},
		},
	}

	_, err = d.db.Client.TransactWriteItems(ctx, &transactWriteItems, func(o *dynamodb.Options) {
		o.RetryMaxAttempts = 1 // we don't want to retry this operation due to the favoritesCount increment
	})

	if err != nil {
		var transactionCanceledErr *ddbtypes.TransactionCanceledException
		if errors.As(err, &transactionCanceledErr) {
			for index, reason := range transactionCanceledErr.CancellationReasons {
				if reason.Code != nil && *reason.Code == conditionalCheckFailed && index == 0 {
					return fmt.Errorf("%w: %w", errutil.ErrAlreadyFavorited, err)
				}
			}
			return fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
		}
		return fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}

	return nil
}

func (d DynamodbArticleRepository) FindCommentByCommentIdAndArticleId(ctx context.Context, commentId, articleId uuid.UUID) (domain.Comment, error) {
	input := &dynamodb.GetItemInput{
		TableName: &commentTable,
		Key: map[string]ddbtypes.AttributeValue{
			"commentId": &ddbtypes.AttributeValueMemberS{Value: commentId.String()},
			"articleId": &ddbtypes.AttributeValueMemberS{Value: articleId.String()},
		},
	}

	result, err := d.db.Client.GetItem(ctx, input)
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

func (d DynamodbArticleRepository) IsFavorited(ctx context.Context, articleId, userId uuid.UUID) (bool, error) {

	input := &dynamodb.QueryInput{
		TableName:              &favoriteTable,
		KeyConditionExpression: aws.String("userId = :userId AND articleId = :articleId"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":userId":    &ddbtypes.AttributeValueMemberS{Value: userId.String()},
			":articleId": &ddbtypes.AttributeValueMemberS{Value: articleId.String()},
		},
		Select: ddbtypes.SelectCount,
	}

	result, err := d.db.Client.Query(ctx, input)
	if err != nil {
		return false, fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}

	return result.Count > 0, nil
}

func (d DynamodbArticleRepository) FindArticlesByIds(ctx context.Context, articleIds []uuid.UUID) ([]domain.Article, error) {
	// short circuit if articleIds is empty, no need to query
	// also, dynamodb will throw a validation error if we try to query with empty keys
	if len(articleIds) == 0 {
		return []domain.Article{}, nil
	}

	keys := make([]map[string]ddbtypes.AttributeValue, 0, len(articleIds))
	for _, articleId := range articleIds {
		keys = append(keys, map[string]ddbtypes.AttributeValue{
			"pk": &ddbtypes.AttributeValueMemberS{Value: articleId.String()},
		})
	}

	input := &dynamodb.BatchGetItemInput{
		RequestItems: map[string]ddbtypes.KeysAndAttributes{
			articleTable: {
				Keys: keys,
			},
		},
	}

	result, err := d.db.Client.BatchGetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}

	dynamodbArticleItems := make([]DynamodbArticleItem, 0, len(result.Responses[articleTable]))
	err = attributevalue.UnmarshalListOfMaps(result.Responses[articleTable], &dynamodbArticleItems)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errutil.ErrDynamoMarshalling, err)
	}

	articles := make([]domain.Article, 0, len(dynamodbArticleItems))
	for _, dynamodbArticleItem := range dynamodbArticleItems {
		articles = append(articles, toDomainArticle(dynamodbArticleItem))
	}

	return articles, nil

}

func (d DynamodbArticleRepository) IsFavoritedBulk(ctx context.Context, userId uuid.UUID, articleIds []uuid.UUID) (mapset.Set[uuid.UUID], error) {
	set := mapset.NewThreadUnsafeSet[uuid.UUID]()

	// short circuit if articleIds is empty, no need to query
	// also, dynamodb will throw a validation error if we try to query with empty keys
	if len(articleIds) == 0 {
		return set, nil
	}

	keys := make([]map[string]ddbtypes.AttributeValue, 0, len(articleIds))
	for _, articleId := range articleIds {
		keys = append(keys, map[string]ddbtypes.AttributeValue{
			"userId":    &ddbtypes.AttributeValueMemberS{Value: userId.String()},
			"articleId": &ddbtypes.AttributeValueMemberS{Value: articleId.String()},
		})
	}

	input := &dynamodb.BatchGetItemInput{
		RequestItems: map[string]ddbtypes.KeysAndAttributes{
			favoriteTable: {
				Keys: keys,
			},
		},
	}

	result, err := d.db.Client.BatchGetItem(ctx, input)
	if err != nil {
		return set, fmt.Errorf("%w: %w", errutil.ErrDynamoQuery, err)
	}

	dynamodbFavoriteArticleItems := make([]DynamodbFavoriteArticleItem, 0, len(result.Responses[favoriteTable]))
	err = attributevalue.UnmarshalListOfMaps(result.Responses[favoriteTable], &dynamodbFavoriteArticleItems)
	if err != nil {
		return set, fmt.Errorf("%w: %w", errutil.ErrDynamoMarshalling, err)
	}
	for _, item := range dynamodbFavoriteArticleItems {
		// ToDo @ender must parse again...
		set.Add(uuid.MustParse(item.ArticleId))
	}

	return set, nil
}

func toDynamodbArticleItem(article domain.Article) DynamodbArticleItem {
	return DynamodbArticleItem{
		Id:             article.Id.String(),
		Title:          article.Title,
		Slug:           article.Slug,
		Description:    article.Description,
		Body:           article.Body,
		TagList:        article.TagList,
		FavoritesCount: article.FavoritesCount,
		AuthorId:       article.AuthorId.String(),
		CreatedAt:      article.CreatedAt.UnixMilli(),
		UpdatedAt:      article.UpdatedAt.UnixMilli(),
	}
}

func toDomainArticle(article DynamodbArticleItem) domain.Article {
	return domain.Article{
		Id:             uuid.MustParse(article.Id), // ToDo @ender we should never use uuid.MustParse - it panics
		Title:          article.Title,
		Slug:           article.Slug,
		Description:    article.Description,
		Body:           article.Body,
		TagList:        article.TagList,
		FavoritesCount: article.FavoritesCount,
		AuthorId:       uuid.MustParse(article.AuthorId),
		CreatedAt:      time.UnixMilli(article.CreatedAt),
		UpdatedAt:      time.UnixMilli(article.UpdatedAt),
	}
}

func toDynamodbCommentItem(article domain.Comment) DynamodbCommentItem {
	return DynamodbCommentItem{
		Id:        article.Id.String(),
		ArticleId: article.ArticleId.String(),
		AuthorId:  article.AuthorId.String(),
		Body:      article.Body,
		CreatedAt: article.CreatedAt,
		UpdatedAt: article.UpdatedAt,
	}
}

func toDomainComment(article DynamodbCommentItem) domain.Comment {
	return domain.Comment{
		Id:        uuid.MustParse(article.Id),
		ArticleId: uuid.MustParse(article.ArticleId),
		AuthorId:  uuid.MustParse(article.AuthorId),
		Body:      article.Body,
		CreatedAt: article.CreatedAt,
		UpdatedAt: article.UpdatedAt,
	}
}

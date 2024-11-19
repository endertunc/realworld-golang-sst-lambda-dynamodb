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

type dynamodbArticleRepository struct {
	db *database.DynamoDBStore
}

type ArticleRepositoryInterface interface {
	FindArticleBySlug(ctx context.Context, email string) (domain.Article, error)
	FindArticleById(ctx context.Context, articleId uuid.UUID) (domain.Article, error)
	FindArticlesByIds(ctx context.Context, articleIds []uuid.UUID) ([]domain.Article, error)
	FindArticlesByAuthor(ctx context.Context, authorId uuid.UUID, limit int, nextPageToken *string) ([]domain.Article, *string, error)

	CreateArticle(ctx context.Context, article domain.Article) (domain.Article, error)
	DeleteArticleById(ctx context.Context, articleId uuid.UUID) error

	UnfavoriteArticle(ctx context.Context, loggedInUserId uuid.UUID, articleId uuid.UUID) error
	FavoriteArticle(ctx context.Context, loggedInUserId uuid.UUID, articleId uuid.UUID) error

	IsFavorited(ctx context.Context, articleId, userId uuid.UUID) (bool, error)
	IsFavoritedBulk(ctx context.Context, userId uuid.UUID, articleIds []uuid.UUID) (mapset.Set[uuid.UUID], error)
	FindArticlesFavoritedByUser(ctx context.Context, userId uuid.UUID, limit int, nextPageToken *string) ([]uuid.UUID, *string, error)
}

var _ ArticleRepositoryInterface = dynamodbArticleRepository{}

func NewDynamodbArticleRepository(db *database.DynamoDBStore) ArticleRepositoryInterface {
	return dynamodbArticleRepository{db: db}
}

type DynamodbArticleItem struct {
	Id             DynamodbUUID `dynamodbav:"pk"`
	Title          string       `dynamodbav:"title"`
	Slug           string       `dynamodbav:"slug"`
	Description    string       `dynamodbav:"description"`
	Body           string       `dynamodbav:"body"`
	TagList        []string     `dynamodbav:"tagList"`
	FavoritesCount int          `dynamodbav:"favoritesCount"`
	AuthorId       DynamodbUUID `dynamodbav:"authorId"`
	// ToDo @ender should we convert everything to milliseconds precision?
	CreatedAt int64 `dynamodbav:"createdAt"`
	UpdatedAt int64 `dynamodbav:"updatedAt"`
}

var articleTable = "article"
var favoriteTable = "favorite"

var articleSlugGSI = aws.String("article_slug_gsi")
var articleAuthorIdGSI = aws.String("article_author_gsi")
var favoriteUserIdCreatedAtGSI = aws.String("favorite_user_id_created_at_gsi")

type DynamodbFavoriteArticleItem struct {
	UserId    DynamodbUUID `dynamodbav:"userId"`
	ArticleId DynamodbUUID `dynamodbav:"articleId"`
	CreatedAt int64        `dynamodbav:"createdAt"`
}

func (d dynamodbArticleRepository) FindArticleBySlug(ctx context.Context, slug string) (domain.Article, error) {
	input := &dynamodb.QueryInput{
		TableName:              &articleTable,
		IndexName:              articleSlugGSI,
		KeyConditionExpression: aws.String("slug = :slug"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":slug": &ddbtypes.AttributeValueMemberS{Value: slug},
		},
	}

	article, err := QueryOne(ctx, d.db.Client, input, toDomainArticle)
	if err != nil {
		if errors.Is(err, ErrDynamodbItemNotFound) {
			return domain.Article{}, errutil.ErrArticleNotFound
		}
		return domain.Article{}, err
	}
	return article, nil
}

func (d dynamodbArticleRepository) FindArticleBySlugTBD(ctx context.Context, slug string) (domain.Article, error) {
	input := dynamodb.QueryInput{
		TableName:              &articleTable,
		IndexName:              articleSlugGSI,
		KeyConditionExpression: aws.String("slug = :slug"),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":slug": &ddbtypes.AttributeValueMemberS{Value: slug},
		},
	}
	article, err := QueryOne(ctx, d.db.Client, &input, toDomainArticle)
	if err != nil {
		if errors.Is(err, ErrDynamodbItemNotFound) {
			return domain.Article{}, errutil.ErrArticleNotFound
		}
		return domain.Article{}, err
	}
	return article, nil
}

func (d dynamodbArticleRepository) FindArticleById(ctx context.Context, articleId uuid.UUID) (domain.Article, error) {
	input := &dynamodb.GetItemInput{
		TableName: &articleTable,
		Key: map[string]ddbtypes.AttributeValue{
			"pk": &ddbtypes.AttributeValueMemberS{Value: articleId.String()},
		},
	}

	result, err := d.db.Client.GetItem(ctx, input)
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
func (d dynamodbArticleRepository) CreateArticle(ctx context.Context, article domain.Article) (domain.Article, error) {
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

func (d dynamodbArticleRepository) DeleteArticleById(c context.Context, articleId uuid.UUID) error {
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

// UnfavoriteArticle deletes the favorite item from the favorite table and decrements the favoritesCount of the article
// if the favorite item does not exist, it does not decrement the favoritesCount and returns an ErrAlreadyUnfavorited error
func (d dynamodbArticleRepository) UnfavoriteArticle(ctx context.Context, loggedInUserId uuid.UUID, articleId uuid.UUID) error {
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
func (d dynamodbArticleRepository) FavoriteArticle(ctx context.Context, loggedInUserId uuid.UUID, articleId uuid.UUID) error {
	favorite := DynamodbFavoriteArticleItem{
		UserId:    DynamodbUUID(loggedInUserId),
		ArticleId: DynamodbUUID(articleId),
		CreatedAt: time.Now().UnixMilli(),
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
				// ToDo @ender err.Error() doesnt give much information about the nature of the underlying issue.
				//  we should come up with a better way to retain the root cause of the error inside the CancellationReasons
				//  in all place where we use transactWriteItems
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

func (d dynamodbArticleRepository) IsFavorited(ctx context.Context, articleId, userId uuid.UUID) (bool, error) {

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

func (d dynamodbArticleRepository) FindArticlesByIds(ctx context.Context, articleIds []uuid.UUID) ([]domain.Article, error) {
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

func (d dynamodbArticleRepository) IsFavoritedBulk(ctx context.Context, userId uuid.UUID, articleIds []uuid.UUID) (mapset.Set[uuid.UUID], error) {
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

	articleIds, err := BatchGetItems(ctx, d.db.Client, favoriteTable, keys, func(item DynamodbFavoriteArticleItem) uuid.UUID {
		return uuid.UUID(item.ArticleId)
	})
	if err != nil {
		return nil, err
	}

	for _, articleId := range articleIds {
		set.Add(articleId)
	}
	return set, nil
}

func (d dynamodbArticleRepository) FindArticlesByAuthor(ctx context.Context, authorId uuid.UUID, limit int, nextPageToken *string) ([]domain.Article, *string, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(articleTable),
		IndexName:              articleAuthorIdGSI,
		KeyConditionExpression: aws.String("authorId = :authorId"),
		Limit:                  aws.Int32(int32(limit)),
		ScanIndexForward:       aws.Bool(false),
		ExpressionAttributeValues: map[string]ddbtypes.AttributeValue{
			":authorId": &ddbtypes.AttributeValueMemberS{Value: authorId.String()},
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

	articles, nextPageToken, err := QueryMany(ctx, d.db.Client, input, toDomainArticle)
	if err != nil {
		return nil, nil, err
	}

	return articles, nextPageToken, nil
}

func (d dynamodbArticleRepository) FindArticlesFavoritedByUser(ctx context.Context, userId uuid.UUID, limit int, nextPageToken *string) ([]uuid.UUID, *string, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(favoriteTable),
		IndexName:              favoriteUserIdCreatedAtGSI,
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

	articleIds, nextPageToken, err := QueryMany(ctx, d.db.Client, input, func(item DynamodbFavoriteArticleItem) uuid.UUID {
		return uuid.UUID(item.ArticleId)
	})
	if err != nil {
		return nil, nil, err
	}

	return articleIds, nextPageToken, nil
}

func toDynamodbArticleItem(article domain.Article) DynamodbArticleItem {
	return DynamodbArticleItem{
		Id:             DynamodbUUID(article.Id),
		Title:          article.Title,
		Slug:           article.Slug,
		Description:    article.Description,
		Body:           article.Body,
		TagList:        article.TagList,
		FavoritesCount: article.FavoritesCount,
		AuthorId:       DynamodbUUID(article.AuthorId),
		CreatedAt:      article.CreatedAt.UnixMilli(),
		UpdatedAt:      article.UpdatedAt.UnixMilli(),
	}
}

func toDomainArticle(article DynamodbArticleItem) domain.Article {
	return domain.Article{
		Id:             uuid.UUID(article.Id),
		Title:          article.Title,
		Slug:           article.Slug,
		Description:    article.Description,
		Body:           article.Body,
		TagList:        article.TagList,
		FavoritesCount: article.FavoritesCount,
		AuthorId:       uuid.UUID(article.AuthorId),
		CreatedAt:      time.UnixMilli(article.CreatedAt),
		UpdatedAt:      time.UnixMilli(article.UpdatedAt),
	}
}

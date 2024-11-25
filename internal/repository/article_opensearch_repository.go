package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"realworld-aws-lambda-dynamodb-golang/internal/database"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"realworld-aws-lambda-dynamodb-golang/internal/errutil"
	"realworld-aws-lambda-dynamodb-golang/internal/test"
	"strings"
	"time"
)

type articleOpensearchRepository struct {
	db *database.OpenSearchStore
}

type ArticleOpensearchRepositoryInterface interface {
	FindAllArticles(ctx context.Context, limit int, offset *int) ([]domain.Article, *int, error)
	FindArticlesByTag(ctx context.Context, tag string, limit int, offset *int) ([]domain.Article, *int, error)
	FindAllTags(ctx context.Context) ([]string, error)
}

var _ ArticleOpensearchRepositoryInterface = articleOpensearchRepository{}

func NewArticleOpensearchRepository(db *database.OpenSearchStore) ArticleOpensearchRepositoryInterface {
	return articleOpensearchRepository{db: db}
}

type OpensearchArticleDocument struct {
	Id             uuid.UUID `json:"pk"`
	Title          string    `json:"title"`
	Slug           string    `json:"slug"`
	Description    string    `json:"description"`
	Body           string    `json:"body"`
	TagList        []string  `json:"tagList"`
	FavoritesCount int       `json:"favoritesCount"`
	AuthorId       uuid.UUID `json:"authorId"`
	CreatedAt      int64     `json:"createdAt"`
	UpdatedAt      int64     `json:"updatedAt"`
}

type TagAggregationsResult struct {
	TagList struct {
		Buckets []struct {
			Key string `json:"key"`
		} `json:"buckets"`
	} `json:"tagList"`
}

var (
	articleIndex = "article"
)

func (o articleOpensearchRepository) FindAllArticles(ctx context.Context, limit int, offset *int) ([]domain.Article, *int, error) {
	from := 0
	if offset != nil {
		from = *offset
	}

	query := strings.NewReader(fmt.Sprintf(`
	{
		"from": %d,
		"size": %d,
		"query": {
			"match_all": {}
		},
		"sort":[
			{
				"createdAt": {
        			"order": "desc"
				}
			}
		],
		"track_total_hits": true
	}`, from, limit))

	request := opensearchapi.SearchReq{
		Indices: []string{articleIndex},
		Body:    query,
	}

	response, err := o.db.Client.Search(ctx, &request)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", errutil.ErrOpensearchQuery, err)
	}

	test.PrintAsJSON(response)

	articles := make([]domain.Article, 0)
	for _, hit := range response.Hits.Hits {
		article := OpensearchArticleDocument{}
		err := json.Unmarshal(hit.Source, &article)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: %w", errutil.ErrOpensearchMarshalling, err)
		}
		articles = append(articles, article.toDomainArticle())
	}

	return articles, nil, nil
}

func (o articleOpensearchRepository) FindArticlesByTag(ctx context.Context, tag string, limit int, offset *int) ([]domain.Article, *int, error) {
	from := 0
	if offset != nil {
		from = *offset
	}

	query := strings.NewReader(fmt.Sprintf(`
	{
		"from": %d,
		"size": %d,
		"query": {
			"match": {
				"tagList": "%s"
			}
		},
		"sort":[
			{
				"createdAt": {
        			"order": "desc"
				}
			}
		],
		"track_total_hits": true
	}`, from, limit, tag))

	request := opensearchapi.SearchReq{
		Indices: []string{articleIndex},
		Body:    query,
	}

	response, err := o.db.Client.Search(ctx, &request)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", errutil.ErrOpensearchQuery, err)
	}

	articles := make([]domain.Article, 0)
	for _, hit := range response.Hits.Hits {
		article := OpensearchArticleDocument{}
		err := json.Unmarshal(hit.Source, &article)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: %w", errutil.ErrOpensearchMarshalling, err)
		}
		articles = append(articles, article.toDomainArticle())
	}

	return articles, nil, nil
}

// "size: 0" at root means we don't want documents to be returned, just the aggregation
// "size: 100" at terms level means we want to get the top 100 tags which is enough for our use case
var query = strings.NewReader(`
	{
		"size" : 0,
	  	"aggs": {
	    	"tagList": {
	      		"terms": { 
					"size": 100,
					"field": "tagList.keyword"
				}
	    	}
	  	}
	}`)

func (o articleOpensearchRepository) FindAllTags(ctx context.Context) ([]string, error) {
	request := opensearchapi.SearchReq{
		Indices: []string{articleIndex},
		Body:    query,
	}

	response, err := o.db.Client.Search(ctx, &request)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errutil.ErrOpensearchQuery, err)
	}

	tagAggregationsResult := TagAggregationsResult{}
	err = json.Unmarshal(response.Aggregations, &tagAggregationsResult)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errutil.ErrOpensearchMarshalling, err)
	}

	tags := make([]string, 0)
	for _, bucket := range tagAggregationsResult.TagList.Buckets {
		tags = append(tags, bucket.Key)
	}

	return tags, nil
}

func (articleDocument OpensearchArticleDocument) toDomainArticle() domain.Article {
	return domain.Article{
		Id:             articleDocument.Id,
		Title:          articleDocument.Title,
		Slug:           articleDocument.Slug,
		Description:    articleDocument.Description,
		Body:           articleDocument.Body,
		TagList:        articleDocument.TagList,
		FavoritesCount: articleDocument.FavoritesCount,
		AuthorId:       articleDocument.AuthorId,
		CreatedAt:      time.UnixMilli(articleDocument.CreatedAt),
		UpdatedAt:      time.UnixMilli(articleDocument.UpdatedAt),
	}
}

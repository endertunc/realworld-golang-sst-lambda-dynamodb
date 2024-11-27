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
	"strings"
	"time"
)

type articleOpensearchRepository struct {
	db *database.OpenSearchStore
}

type ArticleOpensearchRepositoryInterface interface {
	FindAllArticles(ctx context.Context, limit int, offset *string) ([]domain.Article, *string, error)
	FindArticlesByTag(ctx context.Context, tag string, limit int, offset *string) ([]domain.Article, *string, error)
	FindAllTags(ctx context.Context) ([]string, error)
}

var _ ArticleOpensearchRepositoryInterface = articleOpensearchRepository{} //nolint:golint,exhaustruct

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

func (o articleOpensearchRepository) FindAllArticles(ctx context.Context, limit int, offset *string) ([]domain.Article, *string, error) {
	matchAll := map[string]any{
		"match_all": map[string]any{},
	}
	queryBody, err := prepareQueryWithPagination(matchAll, limit, offset)
	if err != nil {
		return nil, nil, err
	}

	searchReq := opensearchapi.SearchReq{
		Indices: []string{articleIndex},
		Body:    strings.NewReader(queryBody),
	}

	searchResp, err := o.db.Client.Search(ctx, &searchReq)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", errutil.ErrOpensearchQuery, err)
	}

	articles, newNextPageToken, err := parseSearchArticleResponse(searchResp, limit)
	if err != nil {
		return nil, nil, err
	}

	return articles, newNextPageToken, nil
}

func (o articleOpensearchRepository) FindArticlesByTag(ctx context.Context, tag string, limit int, nextPageToken *string) ([]domain.Article, *string, error) {
	matchByTag := map[string]any{
		"match": map[string]any{
			"tagList": tag,
		},
	}

	queryBody, err := prepareQueryWithPagination(matchByTag, limit, nextPageToken)
	if err != nil {
		return nil, nil, err
	}

	searchReq := opensearchapi.SearchReq{
		Indices: []string{articleIndex},
		Body:    strings.NewReader(queryBody),
	}

	searchResp, err := o.db.Client.Search(ctx, &searchReq)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", errutil.ErrOpensearchQuery, err)
	}

	articles, newNextPageToken, err := parseSearchArticleResponse(searchResp, limit)
	if err != nil {
		return nil, nil, err
	}

	return articles, newNextPageToken, nil
}

func parseSearchArticleResponse(response *opensearchapi.SearchResp, limit int) ([]domain.Article, *string, error) {
	articles := make([]domain.Article, 0)
	for _, hit := range response.Hits.Hits {
		var article OpensearchArticleDocument
		err := json.Unmarshal(hit.Source, &article)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: %w", errutil.ErrOpensearchMarshalling, err)
		}
		articles = append(articles, article.toDomainArticle())
	}
	// records last item's sort value as nextPageToken
	// if we get fewer documents than limit, then there is no next page
	var nextPageToken *string
	if limit == len(response.Hits.Hits) {
		bytes, err := json.Marshal(response.Hits.Hits[len(response.Hits.Hits)-1].Sort)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: %w", errutil.ErrOpensearchMarshalling, err)
		}
		// ToDo @ender let's base64 encode this
		s := string(bytes)
		nextPageToken = &s
	}
	return articles, nextPageToken, nil
}

// using only the provided api surface from opensearch-go there is no way to pass `search_after`
// therefore I decided to simply build the query myself as json object which is represented as a map[string]any in golang.
func prepareQueryWithPagination(query map[string]any, limit int, nextPageToken *string) (string, error) {
	queryMap := map[string]any{
		"size":  limit,
		"query": query,
		"sort": []map[string]any{
			{"createdAt": "desc"},
		},
	}
	if nextPageToken != nil {
		searchAfter := make([]any, 0)
		err := json.Unmarshal([]byte(*nextPageToken), &searchAfter)
		if err != nil {
			return "", fmt.Errorf("%w: %w", errutil.ErrOpensearchMarshalling, err)
		}
		queryMap["search_after"] = searchAfter
	}

	queryBody, err := json.Marshal(queryMap)
	if err != nil {
		return "", fmt.Errorf("%w: %w", errutil.ErrOpensearchMarshalling, err)
	}

	return string(queryBody), nil
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

	var tagAggregationsResult TagAggregationsResult
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

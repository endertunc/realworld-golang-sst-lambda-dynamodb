package repository

import (
	"context"
	"encoding/json"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"realworld-aws-lambda-dynamodb-golang/internal/database"
	"realworld-aws-lambda-dynamodb-golang/internal/domain"
	"strings"
	"testing"
	"time"
)

var (
	osStore = database.NewOpensearchStore()
	repo    = NewArticleOpensearchRepository(osStore)
)

func TestArticleOpensearchRepository_FindAllArticles(t *testing.T) {
	withOpensearchCleanup(t, osStore, func() {
		// setup articles with different creation times
		article1 := generateOpensearchArticleDocument()
		article2 := generateOpensearchArticleDocument()
		article3 := generateOpensearchArticleDocument()

		article1.CreatedAt = time.Now().Unix()
		article2.CreatedAt = time.Now().Add(-time.Hour * 24).Unix()
		article3.CreatedAt = time.Now().Add(-time.Hour * 48).Unix()

		createArticleDocument(t, osStore, article3)
		createArticleDocument(t, osStore, article2)
		createArticleDocument(t, osStore, article1)

		t.Run("should return all articles with pagination", func(t *testing.T) {
			assert.EventuallyWithT(t, func(ct *assert.CollectT) {
				// test first page
				limit := 2
				articles, nextPageToken, err := repo.FindAllArticles(context.Background(), limit, nil)
				assert.NoError(ct, err)
				assert.NotEmpty(ct, nextPageToken)
				assert.Equal(ct, 2, len(articles))
				// articles should be sorted by createdAt desc
				expectedArticles := []domain.Article{article1.toDomainArticle(), article2.toDomainArticle()}
				assert.Equal(ct, expectedArticles, articles)

				// test second page
				articles, nextPageToken, err = repo.FindAllArticles(context.Background(), limit, nextPageToken)
				assert.NoError(ct, err)
				assert.Empty(ct, nextPageToken)
				assert.Equal(ct, 1, len(articles))
				expectedArticles = []domain.Article{article3.toDomainArticle()}
				assert.Equal(ct, expectedArticles, articles)
			}, 5*time.Second, 500*time.Millisecond)
		})
	})
}

func TestArticleOpensearchRepository_FindArticlesByTag(t *testing.T) {
	withOpensearchCleanup(t, osStore, func() {
		// setup articles with different tags and creation times
		article1 := generateOpensearchArticleDocument()
		article2 := generateOpensearchArticleDocument()
		article3 := generateOpensearchArticleDocument()
		article4 := generateOpensearchArticleDocument()

		article1.TagList = []string{"tag1", "tag2"}
		article2.TagList = []string{"tag2", "tag3"}
		article3.TagList = []string{}
		article4.TagList = []string{"tag1", "tag2"}

		article1.CreatedAt = time.Now().Unix()
		article2.CreatedAt = time.Now().Add(-time.Hour * 24).Unix()
		article3.CreatedAt = time.Now().Add(-time.Hour * 48).Unix()
		article4.CreatedAt = time.Now().Add(-time.Hour * 72).Unix()

		createArticleDocument(t, osStore, article1)
		createArticleDocument(t, osStore, article2)
		createArticleDocument(t, osStore, article3)
		createArticleDocument(t, osStore, article4)
		//{"query":{"match":{"tagList":"tag2"}},"size":2,"sort":[{"createdAt":"desc"}]}
		//{"query":{"match":{"tagList":"tag2"}},"search_after":[1732546752],"size":2,"sort":[{"createdAt":"desc"}]}
		//{"query":{"match":{"match_all":{}  }},"size":2,"sort":[{"createdAt":"desc"}]}
		t.Run("should return articles by tag with pagination", func(t *testing.T) {
			assert.EventuallyWithT(t, func(ct *assert.CollectT) {
				// test first page
				limit := 2
				articles, nextPageToken, err := repo.FindArticlesByTag(context.Background(), "tag2", limit, nil)
				require.NoError(ct, err)
				assert.Equal(ct, 2, len(articles))
				assert.NotEmpty(ct, nextPageToken)
				// articles should be sorted by createdAt desc
				expectedArticles := []domain.Article{article1.toDomainArticle(), article2.toDomainArticle()}
				assert.Equal(ct, expectedArticles, articles)

				// test second page
				articles, nextPageToken, err = repo.FindArticlesByTag(context.Background(), "tag2", limit, nextPageToken)
				require.NoError(ct, err)
				assert.Empty(ct, nextPageToken)
				assert.Equal(ct, 1, len(articles))
				expectedArticles = []domain.Article{article4.toDomainArticle()} // article3 is not tagged with tag2
				assert.Equal(ct, expectedArticles, articles)

			}, 5*time.Second, 500*time.Millisecond)

		})

		t.Run("should return empty list for non-existent tag", func(t *testing.T) {
			articles, nextTokenPage, err := repo.FindArticlesByTag(context.Background(), "non-existent", 10, nil)
			require.NoError(t, err)
			assert.Empty(t, articles)
			assert.Empty(t, nextTokenPage)
		})
	})
}

func TestArticleOpensearchRepository_FindAllTags(t *testing.T) {

	withOpensearchCleanup(t, osStore, func() {
		// setup articles with different tags and creation times
		article1 := generateOpensearchArticleDocument()
		article2 := generateOpensearchArticleDocument()
		article3 := generateOpensearchArticleDocument()

		article1.TagList = []string{"tag1", "tag2"}
		article2.TagList = []string{"tag2", "tag3"}
		article3.TagList = []string{"tag3", "tag4"}

		createArticleDocument(t, osStore, article1)
		createArticleDocument(t, osStore, article2)
		createArticleDocument(t, osStore, article3)

		t.Run("should return all unique tags", func(t *testing.T) {
			assert.EventuallyWithT(t, func(ct *assert.CollectT) {
				tags, err := repo.FindAllTags(context.Background())
				require.NoError(ct, err)
				assert.Equal(ct, 4, len(tags))
				assert.ElementsMatch(ct, []string{"tag1", "tag2", "tag3", "tag4"}, tags)
			}, 5*time.Second, 500*time.Millisecond)
		})
	})
}

func withOpensearchCleanup(t *testing.T, db *database.OpenSearchStore, testFunc func()) {
	cleanupOpensearch(t, db)
	testFunc()
}

func cleanupOpensearch(t *testing.T, db *database.OpenSearchStore) {
	// delete all documents from the index
	query := strings.NewReader(`
	{
		"query": {
			"match_all": {}
		}
	}`)

	request := opensearchapi.DocumentDeleteByQueryReq{
		Indices: []string{articleIndex},
		Body:    query,
	}

	var deleteResp opensearchapi.DocumentDeleteByQueryResp
	_, err := db.Client.Client.Do(context.Background(), &request, &deleteResp)
	require.NoError(t, err)

	// refresh the index to make sure all changes are visible
	refreshReq := opensearchapi.IndicesRefreshReq{
		Indices: []string{articleIndex},
	}
	var refreshResp opensearchapi.IndicesRefreshResp
	_, err = db.Client.Client.Do(context.Background(), &refreshReq, &refreshResp)
	require.NoError(t, err)
}

func createArticleDocument(t *testing.T, db *database.OpenSearchStore, articleItem OpensearchArticleDocument) {
	articleJson, err := json.Marshal(articleItem)
	require.NoError(t, err)

	request := opensearchapi.IndexReq{
		Index:      articleIndex,
		DocumentID: articleItem.Id.String(),
		Body:       strings.NewReader(string(articleJson)),
	}

	var indexResp opensearchapi.IndexResp
	_, err = db.Client.Client.Do(context.Background(), &request, &indexResp)
	require.NoError(t, err)
	return
}

func generateOpensearchArticleDocument() OpensearchArticleDocument {
	title := gofakeit.LoremIpsumSentence(gofakeit.Number(5, 10))
	date := gofakeit.PastDate()
	return OpensearchArticleDocument{
		Id:             uuid.New(),
		Title:          title,
		Slug:           slug.Make(title),
		Description:    gofakeit.LoremIpsumSentence(gofakeit.Number(10, 20)),
		Body:           gofakeit.LoremIpsumParagraph(2, 20, 100, "\n"),
		TagList:        []string{gofakeit.LoremIpsumWord(), gofakeit.LoremIpsumWord()},
		FavoritesCount: gofakeit.Number(0, 100),
		AuthorId:       uuid.New(),
		CreatedAt:      date.UnixMilli(),
		UpdatedAt:      date.UnixMilli(),
	}
}

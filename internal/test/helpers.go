package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func init() {
	loadEnv()
}

// loadEnv loads the .env file from the project root, regardless of the current working directory
func loadEnv() {
	rootDir, err := findProjectRoot()
	if err != nil {
		log.Fatalf("error finding project root: %v", err)
	}
	envPath := filepath.Join(rootDir, ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("no .env file found at %s: %v", envPath, err)
	}
}

// findProjectRoot attempts to find the project root by looking for a go.mod file
func findProjectRoot() (string, error) {
	// Start from the current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Traverse up the directory tree looking for go.mod
	for {
		if _, err := os.Stat(filepath.Join(currentDir, "go.mod")); err == nil {
			return currentDir, nil
		}
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			return "", fmt.Errorf("could not find project root (go.mod file)")
		}
		currentDir = parentDir
	}
}

// This really feels like an overkill tbh...
var apiUrl = sync.OnceValue(func() string {
	value, found := os.LookupEnv("API_URL")
	if !found {
		panic("API_URL env variable is not set")
	}
	return value
})

var DynamodbClient = sync.OnceValue(func() *dynamodb.Client {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("eu-west-1"), config.WithSharedConfigProfile("nl2-golang"))
	if err != nil {
		log.Fatalf("error loading AWS configuration: %v", err)
	}

	client := dynamodb.NewFromConfig(cfg)
	return client
})

func truncateTable(t *testing.T, tableName string, pkName string, skName *string) {
	ctx := context.Background()

	// Scan the table to get items. It's only used for testing purposes.
	scanInput := &dynamodb.ScanInput{
		TableName:       aws.String(tableName),
		AttributesToGet: []string{pkName},
	}

	// Add sort key to attributes if it exists
	if skName != nil {
		scanInput.AttributesToGet = append(scanInput.AttributesToGet, *skName)
	}

	var writeRequests []types.WriteRequest
	var lastEvaluatedKey map[string]types.AttributeValue

	// Keep scanning and deleting until no more items are left
	for {
		// If there's a last evaluated key from previous scan, use it
		if lastEvaluatedKey != nil {
			scanInput.ExclusiveStartKey = lastEvaluatedKey
		}

		result, err := DynamodbClient().Scan(ctx, scanInput)
		if err != nil {
			t.Fatalf("failed to scan table: %v", err)
		}

		if len(result.Items) == 0 {
			return // No more items to delete
		}

		// Process items for deletion
		for i, item := range result.Items {
			key := make(map[string]types.AttributeValue)

			// Extract the primary key
			if pkValue, ok := item[pkName]; ok {
				key[pkName] = pkValue
			} else {
				t.Fatalf("primary key not found in item")
			}

			// Extract the sort key if it exists
			if skName != nil {
				if skValue, ok := item[*skName]; ok {
					key[*skName] = skValue
				} else {
					t.Fatalf("sort key not found in item")
				}
			}

			writeRequests = append(writeRequests, types.WriteRequest{
				DeleteRequest: &types.DeleteRequest{
					Key: key,
				},
			})

			// Perform the batch delete operation when we have 25 items or it's the last item
			if len(writeRequests) == 25 || i == len(result.Items)-1 {
				_, err = DynamodbClient().BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
					RequestItems: map[string][]types.WriteRequest{
						tableName: writeRequests,
					},
				})
				if err != nil {
					t.Fatalf("failed to batch delete items: %v", err)
				}
				writeRequests = make([]types.WriteRequest, 0)
			}
		}

		// Check if we need to continue scanning
		if result.LastEvaluatedKey == nil {
			break // No more pages to scan
		}
		lastEvaluatedKey = result.LastEvaluatedKey
	}
}

func cleanupDynamodbTables(t *testing.T) {
	truncateTable(t, "user", "pk", nil)
	truncateTable(t, "follower", "follower", aws.String("followee"))
	truncateTable(t, "article", "pk", nil)
	truncateTable(t, "comment", "commentId", aws.String("articleId"))
	truncateTable(t, "favorite", "userId", aws.String("articleId"))
	truncateTable(t, "feed", "userId", aws.String("createdAt"))
}

func beforeEach(t *testing.T) {
	cleanupDynamodbTables(t)
}

func afterEach(t *testing.T) {
}

func WithSetupAndTeardown(t *testing.T, testFunc func()) {
	beforeEach(t)
	t.Cleanup(func() {
		afterEach(t)
	})
	testFunc()
}

var client = &http.Client{}

// Nothing is used to indicate that the response body should not be parsed
type Nothing struct{}

// ExecuteRequest will skip parsing the response body if the generic response type is Nothing
func ExecuteRequest[T any](t *testing.T, method, path string, reqBody any, expectedStatusCode int, token *string) T {
	var respBody T
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest(method, apiUrl()+path, bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if token != nil {
		req.Header.Set("Authorization", "Token "+*token)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if expectedStatusCode != resp.StatusCode {
		buf := new(strings.Builder)
		_, _ = io.Copy(buf, resp.Body)
		t.Logf("didn't get expected status code. response body was: %v", buf.String())
	}

	require.Equal(t, expectedStatusCode, resp.StatusCode)

	// `Nothing` type is an experimental approach that I think works very well in this use case
	switch any(respBody).(type) {
	case Nothing:
		return respBody // skip parsing the response body
	default:
		err = json.NewDecoder(resp.Body).Decode(&respBody)
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		return respBody
	}
}

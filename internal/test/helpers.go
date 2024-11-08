package test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"realworld-aws-lambda-dynamodb-golang/internal/domain/dto"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

// ToDo @ender this actually doesn't work...
var _ = godotenv.Load("../../.env")

// This really feels like an overkill tbh...
var apiUrl = sync.OnceValue(func() string {
	value, found := os.LookupEnv("API_URL")
	if !found {
		panic("API_URL env variable is not set")
	}
	return value
})

var dynamodbClient = sync.OnceValue(func() *dynamodb.Client {
	slog.Info("initializing dynamodb client...")
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("eu-west-1"), config.WithSharedConfigProfile("nl2-golang"))
	if err != nil {
		log.Fatalf("error loading AWS configuration: %v", err)
	}

	client := dynamodb.NewFromConfig(cfg)
	slog.Info("dynamodb client is initialized")
	return client
})

func truncateTable(t *testing.T, tableName string, pkName string, skName *string) {
	ctx := context.Background()

	// Scan the table to get all items. It's only used for testing purposes.
	scanInput := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}

	result, err := dynamodbClient().Scan(ctx, scanInput)
	if err != nil {
		t.Fatalf("failed to scan table: %v", err)
	}

	// Batch delete items
	var writeRequests []ddbtypes.WriteRequest
	for _, item := range result.Items {
		key := make(map[string]ddbtypes.AttributeValue)

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

		writeRequests = append(writeRequests, ddbtypes.WriteRequest{
			DeleteRequest: &ddbtypes.DeleteRequest{
				Key: key,
			},
		})

	}

	// Perform the batch delete operation
	if len(writeRequests) > 0 {
		_, err = dynamodbClient().BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]ddbtypes.WriteRequest{
				tableName: writeRequests,
			},
		})

		if err != nil {
			t.Fatalf("failed to batch delete items: %v", err)
		}
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
	t.Log("running before each")
	cleanupDynamodbTables(t)
}

func afterEach(t *testing.T) {
	t.Log("running after each")
}

func WithSetupAndTeardown(t *testing.T, testFunc func()) {
	beforeEach(t)
	t.Cleanup(func() {
		afterEach(t)
	})
	testFunc()
}

// CreateAndLoginUser creates a new user and logs them in, returning the user data and authentication token
func CreateAndLoginUser(t *testing.T, user dto.NewUserRequestUserDto) (dto.NewUserRequestUserDto, string) {
	CreateUserEntity(t, user)

	loginReqBody := dto.LoginRequestBodyDTO{
		User: dto.LoginRequestUserDto{
			Email:    user.Email,
			Password: user.Password,
		},
	}
	var loginRespBody dto.UserResponseBodyDTO
	MakeRequestAndParseResponse(t, loginReqBody, "POST", "/api/users/login", http.StatusOK, &loginRespBody)

	return user, loginRespBody.User.Token
}

func MakeRequestAndParseResponse(t *testing.T, reqBody interface{}, method, path string, expectedStatusCode int, respBody interface{}) {
	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(method, apiUrl()+path, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{} // ToDo should be created once...
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	require.Equal(t, expectedStatusCode, resp.StatusCode)

	if respBody != nil {
		err = json.NewDecoder(resp.Body).Decode(respBody)
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
	}
}

func MakeAuthenticatedRequestAndParseResponse(t *testing.T, reqBody interface{}, method, path string, expectedStatusCode int, respBody interface{}, token string) {
	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(method, apiUrl()+path, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token "+token)

	client := &http.Client{} // ToDo should be created once...
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	require.Equal(t, expectedStatusCode, resp.StatusCode)

	if respBody != nil {
		err = json.NewDecoder(resp.Body).Decode(respBody)
		if err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
	}
}

func MakeRequestAndCheckError(t *testing.T, reqBody interface{}, method, path string, expectedStatusCode int, expectedMessage string) {
	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(method, apiUrl()+path, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{} // ToDo should be created once...
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyString := string(bodyBytes)

	assert.Equal(t, expectedStatusCode, resp.StatusCode)
	assert.Contains(t, bodyString, expectedMessage)
}

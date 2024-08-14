package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"log"
	"realworld-aws-lambda-dynamodb-golang/internal/database"
	"strings"
)

const physicalResourceID = "create-opensearch-indices"

func lambdaHandler(ctx context.Context, event cfn.Event) (string, map[string]interface{}, error) {
	openSearchStore := database.NewOpenSearchStore(ctx)
	indexName := "go-test-index"

	// Define index mapping.
	mapping := strings.NewReader(`{
	 "settings": {
	   "index": {
	        "number_of_shards": 4
	        }
	      }
	 }`)

	// Create an index with non-default settings.
	createResp, err := openSearchStore.Client.Indices.Create(
		ctx,
		opensearchapi.IndicesCreateReq{
			Index: indexName,
			Body:  mapping,
		},
	)
	if err != nil {
		log.Fatal("Error creating index: ", err)
	}

	log.Printf("created index: %s", createResp.Index)

	delResp, err := openSearchStore.Client.Indices.Delete(ctx, opensearchapi.IndicesDeleteReq{Indices: []string{indexName}})
	if err != nil {
		log.Fatal("Error deleting index: ", err)
	}

	log.Printf("deleted index: %#v", delResp.Acknowledged)

	switch event.RequestType {
	case cfn.RequestCreate:
		return physicalResourceID, nil, nil
	case cfn.RequestUpdate:
		return physicalResourceID, nil, nil
	case cfn.RequestDelete:
		return physicalResourceID, nil, nil
	default:
		return "", nil, fmt.Errorf("unknown request type: %s", event.RequestType)
	}
}

func main() {
	lambda.Start(cfn.LambdaWrap(lambdaHandler))
}

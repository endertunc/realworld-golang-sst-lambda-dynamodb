package database

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type DynamoDBStore struct {
	Client *dynamodb.Client
}

func NewDynamoDBStore() *DynamoDBStore {
	// should the context be passed in here?
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("error loading AWS configuration: %v", err)
	}

	client := dynamodb.NewFromConfig(cfg)
	return &DynamoDBStore{
		Client: client,
	}
}

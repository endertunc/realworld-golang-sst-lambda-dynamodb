package database

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/opensearch-project/opensearch-go/v4"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	requestsigner "github.com/opensearch-project/opensearch-go/v4/signer/awsv2"
	"log"
)

type OpenSearchStore struct {
	Client *opensearchapi.Client
}

func NewOpensearchStore() *OpenSearchStore {
	cfg, err := config.LoadDefaultConfig(context.Background())

	if err != nil {
		log.Fatalf("error loading AWS configuration: %v", err)
	}

	signer, err := requestsigner.NewSignerWithService(cfg, "es")
	if err != nil {
		log.Fatalf("error creating request signer: %v", err)
	}

	// opensearch client uses OPENSEARCH_URL env variable by default
	client, err := opensearchapi.NewClient(
		opensearchapi.Config{
			Client: opensearch.Config{
				Signer: signer,
			},
		},
	)

	if err != nil {
		log.Fatalf("error creating OpenSearch client: %v", err)
	}

	return &OpenSearchStore{
		Client: client,
	}
}

package main

import (
	"log"
	"os"
	"realworld-aws-lambda-dynamodb-golang/internal/api/openapi"
)

func main() {
	spec := openapi.GenerateAPISpec()

	specJSON, err := spec.MarshalYAML()
	if err != nil {
		log.Fatalf("Error marshaling spec to YAML: %v", err)
	}

	err = os.WriteFile("docs/spec.yaml", specJSON, 0644)
	if err != nil {
		log.Fatalf("Error writing spec file: %v", err)
	}
}

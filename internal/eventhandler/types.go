package eventhandler

// https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/example_serverless_DynamoDB_Lambda_batch_item_failures_section.html
type BatchItemFailure struct {
	ItemIdentifier string `json:"ItemIdentifier"`
}

type BatchResult struct {
	BatchItemFailures []BatchItemFailure `json:"BatchItemFailures"`
}

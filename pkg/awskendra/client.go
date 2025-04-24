package awskendra

import (
	"context"
)

// Client defines the interface for interacting with AWS Kendra.
// It abstracts away the underlying SDK and queuing mechanisms.
type Client interface {
	// MakeQuery performs a Kendra query, handling potential queuing and result processing.
	// Filters map key is the Kendra attribute key (e.g., "_authors"),
	// value is a slice of strings to filter by for that attribute.
	MakeQuery(ctx context.Context, query string, filters map[string][]string, pageNum int) (KendraResults, error)

	// GetSuggestions retrieves query suggestions from Kendra.
	GetSuggestions(ctx context.Context, query string) (KendraSuggestions, error)
}

// Config holds specific settings for the AWS Clients.
type Config struct {
	Credentials      Provider // Credentials is the provider for retrieving AWS credentials.
	Region           string   // Region provides the AWS region of the Client.
	IndexID          string   // IndexID provides the ID of the Kendra Index.
	ModelID          string   // ModelID provides the ModelID of the BedrockClient.
	RetryMaxAttempts int      // RetryMaxAttempts provides the number of attempts a client should make to AWS before giving up.
	KeywordsFilePath string   // The path to the Keywords file for AWS Bedrock.
}

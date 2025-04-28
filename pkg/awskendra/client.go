package awskendra

import (
	"context"
)

// KendraClient defines the interface for interacting with AWS Kendra.
// It abstracts away the underlying SDK and queuing mechanisms.
type KendraClient interface {
	// MakeQuery performs a Kendra query, handling potential queuing and result processing.
	// Filters map key is the Kendra attribute key (e.g., "_authors"),
	// value is a slice of strings to filter by for that attribute.
	MakeQuery(ctx context.Context, query string, filters map[string][]string, pageNum int) (KendraResults, error)

	// GetSuggestions retrieves query suggestions from Kendra.
	GetSuggestions(ctx context.Context, query string) (KendraSuggestions, error)
}

type Config struct {
	Credentials      Provider
	Region           string
	IndexID          string
	ModelID          string
	BucketName       string
	RetryMaxAttempts int
	KeywordsFilePath string
}

package kendra

import (
	"context"
	"fmt"

	"github.com/DSSD-Madison/gmu/pkg/core/config"
	"github.com/DSSD-Madison/gmu/pkg/core/logger"
	"github.com/DSSD-Madison/gmu/pkg/model/search"
	"github.com/aws/aws-sdk-go-v2/service/kendra"
	"github.com/aws/aws-sdk-go-v2/service/kendra/types"
)

// Client defines the interface for interacting with AWS Kendra.
// It abstracts away the underlying SDK and queuing mechanisms.
type Client interface {
	// MakeQuery performs a Kendra query, handling potential queuing and result processing.
	// Filters map key is the Kendra attribute key (e.g., "_authors"),
	// value is a slice of strings to filter by for that attribute.
	MakeQuery(ctx context.Context, query string, filters map[string][]string, pageNum int) (search.Results, error)

	// GetSuggestions retrieves query suggestions from Kendra.
	GetSuggestions(ctx context.Context, query string) (search.Suggestions, error)
}

type KendraClient struct {
	awsClient  *kendra.Client
	queryQueue QueryExecutor
	config     config.Config
	log        logger.Logger
}

func NewClient(cfg config.Config, log logger.Logger) (Client, error) {
	pkgLogger := log.With("package", "awskendra")

	opts := kendra.Options{
		Credentials:      cfg.AWS.Credentials,
		Region:           cfg.AWS.Region,
		RetryMaxAttempts: cfg.AWS.RetryMaxAttempts,
	}

	awsClient := kendra.New(opts)
	if awsClient == nil {
		pkgLogger.Error("Failed to create AWS Kendra SDK client instance.")
		return nil, fmt.Errorf("error creating AWS Kendra SDK client")
	}
	pkgLogger.Info("AWS Kendra SDK Client initialized")

	workers := 2
	buffer := 5
	queryQueue := NewKendraQueryQueue(awsClient, pkgLogger, workers, buffer)
	pkgLogger.Info("Kendra query queue initialized", "workers", workers, "buffer", buffer)

	return &KendraClient{
		awsClient:  awsClient,
		queryQueue: queryQueue,
		config:     cfg,
		log:        pkgLogger,
	}, nil
}

func (c *KendraClient) GetSuggestions(ctx context.Context, query string) (search.Suggestions, error) {
	c.log.DebugContext(ctx, "Requesting Kendra suggestions", "query", query)
	kendraQuery := kendra.GetQuerySuggestionsInput{
		IndexId:   &c.config.AWS.KendraIndexID,
		QueryText: &query,
	}

	out, err := c.awsClient.GetQuerySuggestions(ctx, &kendraQuery)
	if err != nil {
		c.log.ErrorContext(ctx, "Kendra GetSuggestions API call failed", "error", err)
		return search.Suggestions{}, err
	}

	suggestions := convertSuggestions(*out)
	c.log.DebugContext(ctx, "Kendra suggestions retrieved", "count", len(suggestions.Suggestions))
	return suggestions, nil
}

func (c *KendraClient) MakeQuery(ctx context.Context, query string, filters map[string][]string, pageNum int) (search.Results, error) {
	c.log.DebugContext(ctx, "Building kendra query", "query", query, "page", pageNum, "filter_count", len(filters))

	kendraFilters := types.AttributeFilter{}
	if len(filters) > 0 {
		kendraFilters.AndAllFilters = make([]types.AttributeFilter, 0, len(filters))
		for k, filterCategory := range filters {
			if len(filterCategory) == 0 {
				continue
			}
			key := k

			var subFilter types.AttributeFilter

			if key == "_file_type" || k == "Source" {
				subFilter.OrAllFilters = make([]types.AttributeFilter, len(filterCategory))
				for i, strVal := range filterCategory {
					val := strVal
					subFilter.OrAllFilters[i] = types.AttributeFilter{
						EqualsTo: &types.DocumentAttribute{
							Key: &key,
							Value: &types.DocumentAttributeValue{
								StringValue: &val,
							},
						},
					}
				}
			} else {
				subFilter.ContainsAny = &types.DocumentAttribute{
					Key: &key,
					Value: &types.DocumentAttributeValue{
						StringListValue: filterCategory,
					},
				}
			}
			kendraFilters.AndAllFilters = append(kendraFilters.AndAllFilters, subFilter)
		}
	}

	page := int32(pageNum)
	kendraQueryInput := kendra.QueryInput{
		AttributeFilter: nil,
		IndexId:         &c.config.AWS.KendraIndexID,
		QueryText:       &query,
		PageNumber:      &page,
	}
	if len(kendraFilters.AndAllFilters) > 0 {
		kendraQueryInput.AttributeFilter = &kendraFilters
	}

	c.log.DebugContext(ctx, "Enqueuing Kendra query")
	queryResult := c.queryQueue.EnqueueQuery(ctx, kendraQueryInput)
	if queryResult.Error != nil {
		c.log.ErrorContext(ctx, "Kendra query failed during execution", "error", queryResult.Error)
		return search.Results{}, queryResult.Error
	}

	results := convertToSearchResults(queryResult.Results, pageNum)
	results.Query = query

	return results, nil
}

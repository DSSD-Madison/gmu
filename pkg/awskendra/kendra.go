package awskendra

import (
	"context"
	"fmt"
	"strings"

	"github.com/DSSD-Madison/gmu/pkg/logger"
	"github.com/aws/aws-sdk-go-v2/service/kendra"
	"github.com/aws/aws-sdk-go-v2/service/kendra/types"
)

// kendraClientImpl provides an implementation of the Client interface.
// It wraps the base AWS kendra.Client, incorporating request handling via
// a QueryExecutor, specific configuration settings, and dedicated logging.
type kendraClientImpl struct {
	awsClient  *kendra.Client // awsClient is the underlying AWS Kendra SDK client.
	queryQueue QueryExecutor  // queryQueue handles the execution and queuing of queries.
	config     Config         // config holds specific settings for this client's behavior.
	log        logger.Logger  // log provides dedicated logging for this client's logging.
}

// New initializes a New Kendra Client using the provided config and logger.
func New(config Config, log logger.Logger) (Client, error) {
	pkgLogger := log.With("package", "awskendra")

	opts := kendra.Options{
		Credentials:      config.Credentials,
		Region:           config.Region,
		RetryMaxAttempts: config.RetryMaxAttempts,
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

	return &kendraClientImpl{
		awsClient:  awsClient,
		queryQueue: queryQueue,
		config:     config,
		log:        pkgLogger,
	}, nil
}

// GetSuggestions initiates a request to Kendra for suggestions using the provided query.
func (c *kendraClientImpl) GetSuggestions(ctx context.Context, query string) (KendraSuggestions, error) {
	c.log.DebugContext(ctx, "Requesting Kendra suggestions", "query", query)
	kendraQuery := kendra.GetQuerySuggestionsInput{
		IndexId:   &c.config.IndexID,
		QueryText: &query,
	}

	out, err := c.awsClient.GetQuerySuggestions(ctx, &kendraQuery)
	if err != nil {
		c.log.ErrorContext(ctx, "Kendra GetSuggestions API call failed", "error", err)
		return KendraSuggestions{}, err
	}

	suggestions := querySuggestionsOutputToSuggestions(*out)
	c.log.DebugContext(ctx, "Kendra suggestions retrieved", "count", len(suggestions.Suggestions))
	return suggestions, nil
}

// queryOutputToResults converts a Kendra query output to KendraResults.
func queryOutputToResults(out kendra.QueryOutput) KendraResults {
	kendraResults := KendraResults{
		Results: make(map[string]KendraResult),
		Filters: make([]FilterCategory, len(out.FacetResults)),
	}

	for _, item := range out.ResultItems {
		title := TrimExtension(*item.DocumentTitle.Text)

		var res KendraResult

		if result, ok := kendraResults.Results[title]; !ok {
			res = KendraResult{
				Title:    title,
				Excerpts: make([]Excerpt, 0),
				Link:     *item.DocumentURI,
			}
		} else {
			res = result
		}

		pageNum := 0

		for _, a := range item.DocumentAttributes {
			if *a.Key == "_excerpt_page_number" {
				pageNum = int(*a.Value.LongValue)
			}
		}

		res.Excerpts = append(res.Excerpts, Excerpt{
			Text:    *item.DocumentExcerpt.Text,
			PageNum: pageNum,
		})
		kendraResults.Results[res.Title] = res
	}

	kendraResults.Count = int(*out.TotalNumberOfResults)

	filterNamesMap := map[string]string{
		"Author":     "Authors",
		"Keyword":    "Keywords",
		"Region":     "Regions",
		"Category":   "Categories",
		"Source":     "Source",
		"_file_type": "File Type",
	}

	for i, facetRes := range out.FacetResults {
		Name, ok := filterNamesMap[*facetRes.DocumentAttributeKey]
		if !ok {
			Name = *facetRes.DocumentAttributeKey
		}
		filterCategory := FilterCategory{
			Category: *facetRes.DocumentAttributeKey,
			Options:  make([]FilterOption, len(facetRes.DocumentAttributeValueCountPairs)),
			Name:     Name,
		}
		for j, attribute := range facetRes.DocumentAttributeValueCountPairs {
			filterCategory.Options[j] = FilterOption{
				Label: *attribute.DocumentAttributeValue.StringValue,
				Count: *attribute.Count,
			}
		}
		kendraResults.Filters[i] = filterCategory
	}

	return kendraResults
}

// MakeQuery takes a query, filter map, page number, and request context and initiates a request to AWS Kendra,
// returning KendraResults.
func (c *kendraClientImpl) MakeQuery(ctx context.Context, query string, filters map[string][]string, pageNum int) (KendraResults, error) {
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
		IndexId:         &c.config.IndexID,
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
		return KendraResults{}, queryResult.Error
	}

	c.log.DebugContext(ctx, "Kendra query executed successfully", "result_count", queryResult.Results.Count)

	results := queryResult.Results
	calculatedPages := (results.Count + 9) / 10
	totalPages := min(calculatedPages, 10)

	results.PageStatus = PageStatus{
		CurrentPage: pageNum,
		PrevPage:    pageNum - 1,
		NextPage:    pageNum + 1,
		HasPrev:     pageNum > 1,
		HasNext:     pageNum < totalPages,
		TotalPages:  totalPages,
	}

	results.Query = query
	// results.UrlData.Query = results.Query
	return results, nil
}

// querySuggestionsOutputToSuggestions converts the suggestions output returned
// by a Kendra query for suggestions into KendraSuggestions.
func querySuggestionsOutputToSuggestions(out kendra.GetQuerySuggestionsOutput) KendraSuggestions {
	suggestions := KendraSuggestions{
		Suggestions: make([]string, 0),
	}

	for _, item := range out.Suggestions {
		suggestions.Suggestions = append(suggestions.Suggestions, *item.Value.Text.Text)
	}

	return suggestions
}

// TrimExtension trims defined extensions from a string,
// returning the trimmed string.
func TrimExtension(s string) string {
	if strings.HasSuffix(s, ".pdf") {
		return strings.TrimSuffix(s, ".pdf")
	}
	if strings.HasSuffix(s, ".docx") {
		return strings.TrimSuffix(s, ".docx")
	}
	return s
}

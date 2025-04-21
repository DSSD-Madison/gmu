package awskendra

import (
	"context"
	"fmt"
	"strings"

	"github.com/DSSD-Madison/gmu/pkg/cache"
	"github.com/DSSD-Madison/gmu/pkg/logger"
	"github.com/aws/aws-sdk-go-v2/service/kendra"
	"github.com/aws/aws-sdk-go-v2/service/kendra/types"
)

type kendraClientImpl struct {
	awsClient  *kendra.Client
	queryQueue QueryExecutor
	cache      cache.Cache[KendraResults]
	config     Config
	log        logger.Logger
}

func NewKendraClient(config Config, log logger.Logger) (KendraClient, error) {
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

	queryCache := cache.NewGeneric[KendraResults](pkgLogger)
	pkgLogger.Info("Kendra query cache initialized")

	workers := 2
	buffer := 5
	queryQueue := NewKendraQueryQueue(awsClient, queryCache, pkgLogger, workers, buffer)
	pkgLogger.Info("Kendra query queue initialized", "workers", workers, "buffer", buffer)

	return &kendraClientImpl{
		awsClient:  awsClient,
		queryQueue: queryQueue,
		cache:      queryCache,
		config:     config,
		log:        pkgLogger,
	}, nil
}

func (c *kendraClientImpl) GetSuggestions(ctx context.Context, query string) (KendraSuggestions, error) {
	c.log.DebugContext(ctx, "Requesting Kendra suggestions", "query", query)
	kendraQuery := kendra.GetQuerySuggestionsInput{
		IndexId:   &c.config.IndexID,
		QueryText: &query,
	}

	if true {
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
		"_authors":         "Authors",
		"_file_type":       "File Type",
		"source":           "Source",
		"Subject_Keywords": "Subject Keywords",
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

			if key == "_file_type" {
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

	cacheResult, exist := c.cache.Get(query)
	if exist {
		c.log.Info("Query found in cache", "query", query, "page", pageNum)
		return cacheResult, nil
	}

	c.log.DebugContext(ctx, "Enqueuing Kendra query")
	queryResult := c.queryQueue.EnqueueQuery(ctx, kendraQueryInput)
	if queryResult.Error != nil {
		c.log.ErrorContext(ctx, "Kendra query failed during execution", "error", queryResult.Error)
		return KendraResults{}, queryResult.Error
	}

	c.log.DebugContext(ctx, "Kendra query executed successfully", "result_count", queryResult.Results.Count)

	results := queryResult.Results
	// results.UrlData.Query = results.Query
	return results, nil
}

func querySuggestionsOutputToSuggestions(out kendra.GetQuerySuggestionsOutput) KendraSuggestions {
	suggestions := KendraSuggestions{
		Suggestions: make([]string, 0),
	}

	for _, item := range out.Suggestions {
		suggestions.Suggestions = append(suggestions.Suggestions, *item.Value.Text.Text)
	}

	return suggestions
}

func TrimExtension(s string) string {
	idx := strings.LastIndex(s, ".pdf")

	// No such file extension exists
	if idx == -1 {
		return s
	}

	return s[:idx]
}

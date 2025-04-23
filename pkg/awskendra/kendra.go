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

var filterCategoriesMap = map[string]string{
	"_file_type": "File Type",
	"Category":   "Category",
	"Author":     "Author",
	"Source":     "Source",
	"Region":     "Region",
	"Keyword":    "Keyword",
}

func documentAttributesToFilterCategories(attrs []types.DocumentAttribute) []FilterCategory {
	filters := make([]FilterCategory, 0)

	for _, attr := range attrs {
		if attr.Key != nil {
			if key, ok := filterCategoriesMap[*attr.Key]; ok {

				if attr.Value != nil {
					filter := FilterCategory{
						Category: key,
						Options:  []FilterOption{},
						Name:     key,
					}
					if attr.Value.StringValue != nil { // single item only
						filter.Options = append(filter.Options, FilterOption{
							Label:    *attr.Value.StringValue,
							Selected: false,
							Count:    0,
						})
					} else { // list of options
						for _, strListVal := range attr.Value.StringListValue {
							filter.Options = append(filter.Options, FilterOption{
								Label:    strListVal,
								Selected: false,
								Count:    0,
							})
						}
					}
					filters = append(filters, filter)
				}
			}
		}
	}

	// for _, filter := range filters {
	// 	fmt.Printf("name: %s\n", filter.Name)
	// 	for _, option := range filter.Options {
	// 		fmt.Printf("	->%s\n", option.Label)
	// 	}
	// 	fmt.Println()
	// }

	return filters
}

func queryOutputToResults(out kendra.QueryOutput) KendraResults {
	kendraResults := KendraResults{
		Results: make(map[string]KendraResult),
		Filters: make([]FilterCategory, len(out.FacetResults)),
	}

	for _, item := range out.ResultItems {
		// for _, attr := range item.DocumentAttributes {
		// 	fmt.Printf("attributes ")
		// 	if attr.Key != nil {
		// 		fmt.Printf("key %s ", *attr.Key)
		// 	}
		// 	if attr.Value != nil {
		// 		fmt.Printf("strs: ")
		// 		for _, str := range attr.Value.StringListValue {
		// 			fmt.Printf("%s ", str)
		// 		}
		// 	}
		// 	if attr.Value.StringValue != nil {
		// 		fmt.Printf("stringvalue: %s", *attr.Value.StringValue)
		// 	}
		// 	fmt.Println()
		// }
		title := TrimExtension(*item.DocumentTitle.Text)

		var res KendraResult

		if result, ok := kendraResults.Results[title]; !ok {
			res = KendraResult{
				Title:    title,
				Excerpts: make([]Excerpt, 0),
				Link:     *item.DocumentURI,
				Filters:  documentAttributesToFilterCategories(item.DocumentAttributes),
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
	// cacheResult, exist := c.cache.Get(query)
	// if exist {
	// 	c.log.Info("Query found in cache", "query", query, "page", pageNum)
	// return cacheResult, nil
	// }

	c.log.DebugContext(ctx, "Enqueuing Kendra query")
	queryResult := c.queryQueue.EnqueueQuery(ctx, kendraQueryInput)
	if queryResult.Error != nil {
		c.log.ErrorContext(ctx, "Kendra query failed during execution", "error", queryResult.Error)
		return KendraResults{}, queryResult.Error
	}

	c.log.DebugContext(ctx, "Kendra query executed successfully", "result_count", queryResult.Results.Count)

	results := queryResult.Results
	applyFilters(results, filters)
	// results.UrlData.Query = results.Query
	return results, nil
}

func applyFilters(results KendraResults, filtersMap map[string][]string) KendraResults {
	// iterate over each item in the filters map
	// iterate over each item returned from the results
	// find which
	removed := 0
	for key, result := range results.Results {
		r := result
		for _, resFilterCat := range result.Filters {
			fmt.Println(resFilterCat.Name)
			if filters, ok := filtersMap[resFilterCat.Name]; ok { // if user selected filters category contains the result's filterCategory
				fmt.Println(resFilterCat.Name)
				for _, selectedFilter := range filters {
					fmt.Println("->", selectedFilter)
				}
			} else {
				delete(results.Results, key)
				removed++
				break
			}
		}
		results.Results[key] = r
	}

	fmt.Printf("len of results after applying Filters: %d, removed %d\n", len(results.Results), removed)

	return results
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

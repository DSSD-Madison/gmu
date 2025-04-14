package awskendra

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/kendra"
	"github.com/aws/aws-sdk-go-v2/service/kendra/types"
)

type KendraClient struct {
	client *kendra.Client
	config Config
}

type QueryExecutor interface {
	EnqueueQuery(query kendra.QueryInput) QueryResult
}

type KendraClientExecutor struct {
	client *kendra.Client
}

func (e *KendraClientExecutor) EnqueueQuery(query kendra.QueryInput) QueryResult {
	out, err := e.client.Query(context.TODO(), &query)
	if err != nil {
		log.Printf("Kendra Query Failed: %q", err)
		return QueryResult{
			Results: KendraResults{},
			Error:   err,
		}
	}

	results := queryOutputToResults(*out)

	return QueryResult{
		Results: results,
		Error:   nil,
	}
}

func NewKendraClient(config Config) (*KendraClient, error) {
	opts := kendra.Options{
		Credentials: config.Credentials,
		Region:      config.Region,
	}

	client := kendra.New(opts)
	if client == nil {
		err := fmt.Errorf("Error making kendra client")
		return &KendraClient{}, err
	}

	return &KendraClient{client, config}, nil
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

func (c KendraClient) MakeQuery(executor QueryExecutor, query string, filters map[string][]string, pageNum int) KendraResults {
	kendraFilters := types.AttributeFilter{
		AndAllFilters: make([]types.AttributeFilter, len(filters)),
	}
	andAllIndex := 0
	for k, filterCategory := range filters {
		// _file_type is a string so we can't use ContainsAny
		if k == "_file_type" {
			kendraFilters.AndAllFilters[andAllIndex] = types.AttributeFilter{
				OrAllFilters: make([]types.AttributeFilter, len(filterCategory)),
			}
			for orAllIndex, str := range filterCategory {
				kendraFilters.AndAllFilters[andAllIndex].OrAllFilters[orAllIndex] = types.AttributeFilter{
					EqualsTo: &types.DocumentAttribute{
						Key: &k,
						Value: &types.DocumentAttributeValue{
							StringValue: &str,
						},
					},
				}
			}
		} else {
			kendraFilters.AndAllFilters[andAllIndex] = types.AttributeFilter{
				ContainsAny: &types.DocumentAttribute{
					Key: &k,
					Value: &types.DocumentAttributeValue{
						StringListValue: filterCategory,
					},
				},
			}
		}
		andAllIndex += 1
	}
	page := int32(pageNum)
	kendraQuery := kendra.QueryInput{
		AttributeFilter: &kendraFilters,
		IndexId:         &c.config.IndexID,
		QueryText:       &query,
		PageNumber:      &page,
	}

	queryResults := executor.EnqueueQuery(kendraQuery)
	if queryResults.Error != nil {
		return KendraResults{}
	}
	results := queryResults.Results

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
	results.UrlData.Query = results.Query
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

func (c KendraClient) GetSuggestions(query string) (KendraSuggestions, error) {
	kendraQuery := kendra.GetQuerySuggestionsInput{
		IndexId:   &c.config.IndexID,
		QueryText: &query,
	}
	out, err := c.client.GetQuerySuggestions(context.TODO(), &kendraQuery)
	if err != nil {
		log.Printf("Kendra Suggestions Query Failed %+v", err)
		return KendraSuggestions{}, err
	}

	suggestions := querySuggestionsOutputToSuggestions(*out)
	return suggestions, nil
}

func TrimExtension(s string) string {
	idx := strings.LastIndex(s, ".pdf")

	// No such file extension exists
	if idx == -1 {
		return s
	}

	return s[:idx]
}

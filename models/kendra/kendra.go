package kendra

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/kendra"
	"github.com/aws/aws-sdk-go-v2/service/kendra/types"

	"github.com/DSSD-Madison/gmu/internal"
	"github.com/DSSD-Madison/gmu/models/environment"
)

// client is the global Kendra client used in the program.
var client = kendra.New(kendra.Options{
	Credentials: environment.Provider(),
	Region:      environment.Region(),
})

// indexId holds the indexId for the Kendra index.
var indexId = environment.IndexId()

// queryOutputToResults parses the Kendra query output into KendraResults
// to be used for displaying the results page.
func queryOutputToResults(out kendra.QueryOutput) KendraResults {
	return KendraResults{
		Results: processResults(out.ResultItems),
		Filters: processFilters(out.FacetResults),
		Count:   int(*out.TotalNumberOfResults),
	}
}

func processResults(resultItems []types.QueryResultItem) (results map[string]KendraResult) {
	results = make(map[string]KendraResult)
	for _, item := range resultItems {
		title := internal.TrimExtension(*item.DocumentTitle.Text)

		var res KendraResult

		if result, ok := results[title]; !ok {
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
		results[res.Title] = res
	}
	return
}

func processFilters(facetResults []types.FacetResult) (filters []FilterCategory) {
	filters = make([]FilterCategory, len(facetResults))
	filterNamesMap := map[string]string{
		"_authors":         "Authors",
		"_file_type":       "File Type",
		"source":           "Source",
		"Subject_Keywords": "Subject Keywords",
	}

	for i, facetRes := range facetResults {
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
		filters[i] = filterCategory
	}
	return
}

// MakeQuery builds a query to Kendra.
func MakeQuery(query string, filters map[string][]string) KendraResults {
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
	kendraQuery := kendra.QueryInput{
		AttributeFilter: &kendraFilters,
		IndexId:         &indexId,
		QueryText:       &query,
	}
	out, err := client.Query(context.TODO(), &kendraQuery)

	// TODO: this needs to be fixed to a proper error
	if err != nil {
		log.Printf("Kendra Query Failed %+filterCategory", err)
	}

	results := queryOutputToResults(*out)
	results.Query = query
	return results
}

// querySuggestionsOutputToSuggestions parses Kendra query suggestions into KendraSuggestions.
func querySuggestionsOutputToSuggestions(out kendra.GetQuerySuggestionsOutput) KendraSuggestions {
	suggestions := KendraSuggestions{
		Suggestions: make([]string, 0),
	}

	for _, item := range out.Suggestions {
		suggestions.Suggestions = append(suggestions.Suggestions, *item.Value.Text.Text)
	}

	return suggestions
}

// GetSuggestions queries Kendra for Suggestions using the provided query.
func GetSuggestions(query string) (KendraSuggestions, error) {
	kendraQuery := kendra.GetQuerySuggestionsInput{
		IndexId:   &indexId,
		QueryText: &query,
	}
	out, err := client.GetQuerySuggestions(context.TODO(), &kendraQuery)
	if err != nil {
		log.Printf("Kendra Suggestions Query Failed %+v", err)
		return KendraSuggestions{}, err
	}

	suggestions := querySuggestionsOutputToSuggestions(*out)
	return suggestions, nil
}

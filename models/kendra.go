package models

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/kendra"
	"github.com/aws/aws-sdk-go-v2/service/kendra/types"

	"github.com/DSSD-Madison/gmu/internal"
)

var opts = kendra.Options{
	Credentials: prov,
	Region:      region,
}

var client = kendra.New(opts)

type Excerpt struct {
	Text    string
	PageNum int
}

type KendraResult struct {
	Title    string
	Excerpts []Excerpt
	Link     string
}

type KendraResults struct {
	Results map[string]KendraResult
	Query   string
	Count   int
	Filters []FilterCategory
}

type KendraSuggestions struct {
	Suggestions []string
}

func queryOutputToResults(out kendra.QueryOutput) KendraResults {
	results := KendraResults{
		Results: make(map[string]KendraResult),
		Filters: make([]FilterCategory, len(out.FacetResults)),
	}

	for _, item := range out.ResultItems {
		title := internal.TrimExtension(*item.DocumentTitle.Text)

		var res KendraResult

		if _, ok := results.Results[title]; !ok {
			res = KendraResult{
				Title:    internal.TrimExtension(*item.DocumentTitle.Text),
				Excerpts: make([]Excerpt, 0),
				Link:     *item.DocumentURI,
			}
		} else {
			res = results.Results[title]
		}

		pagenum := 0

		for _, a := range item.DocumentAttributes {
			if *a.Key == "_excerpt_page_number" {
				pagenum = int(*a.Value.LongValue)
			}
		}

		res.Excerpts = append(res.Excerpts, Excerpt{
			Text:    *item.DocumentExcerpt.Text,
			PageNum: pagenum,
		})
		results.Results[res.Title] = res

		results.Count = int(*out.TotalNumberOfResults)
	}

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
		results.Filters[i] = filterCategory
	}

	return results
}

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

func querySuggestionsOutputToSuggestions(out kendra.GetQuerySuggestionsOutput) KendraSuggestions {
	suggestions := KendraSuggestions{
		Suggestions: make([]string, 0),
	}

	for _, item := range out.Suggestions {
		suggestions.Suggestions = append(suggestions.Suggestions, *item.Value.Text.Text)
	}

	return suggestions
}

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

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

type KendraResult struct {
	Title   string
	Excerpt string
	Link    string
}

type KendraResults struct {
	QueryOutput kendra.QueryOutput
	Results     []KendraResult
	Query       string
	Count       int
	Filters     []FilterCategory
}

func queryOutputToResults(out kendra.QueryOutput) KendraResults {
	results := KendraResults{
		Results: make([]KendraResult, len(out.ResultItems)),
		Filters: make([]FilterCategory, len(out.FacetResults)),
	}

	for i, item := range out.ResultItems {
		res := KendraResult{
			Title:   internal.TrimExtension(*item.DocumentTitle.Text),
			Excerpt: *item.DocumentExcerpt.Text,
			Link:    *item.DocumentURI,
		}
		results.Results[i] = res
		results.Count = int(*out.TotalNumberOfResults)

	}

	for i, facetRes := range out.FacetResults {
		filterCategory := FilterCategory{
			Category: *facetRes.DocumentAttributeKey,
			Options:  make([]FilterOption, len(facetRes.DocumentAttributeValueCountPairs)),
		}
		for j, p := range facetRes.DocumentAttributeValueCountPairs {
			filterCategory.Options[j] = FilterOption{
				Label: *p.DocumentAttributeValue.StringValue,
				Count: *p.Count,
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
	for k, v := range filters {
		// _file_type is a string so we can't use ContainsAny
		if k == "_file_type" {
			kendraFilters.AndAllFilters[andAllIndex] = types.AttributeFilter{
				OrAllFilters: make([]types.AttributeFilter, len(v)),
			}
			for orAllIndex, str := range v {
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
						StringListValue: v,
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
		log.Printf("Kendra Query Failed %+v", err)
	}

	results := queryOutputToResults(*out)
	results.Query = query
	return results
}

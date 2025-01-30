package models

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/kendra"
	"github.com/aws/aws-sdk-go-v2/service/kendra/types"

	"github.com/DSSD-Madison/gmu/internal"
)

func Ptr[T any](p T) *T {
	return &p
}

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

	for i, f := range out.FacetResults {
		filterCategory := FilterCategory{
			Category: *f.DocumentAttributeKey,
			Options:  make([]FilterOption, len(f.DocumentAttributeValueCountPairs)),
		}
		for j, p := range f.DocumentAttributeValueCountPairs {
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
	if filters != nil {
		i := 0
		for k, v := range filters {
			kendraFilters.AndAllFilters[i] = types.AttributeFilter{
				ContainsAll: &types.DocumentAttribute{
					Key: &k,
					Value: &types.DocumentAttributeValue{
						StringListValue: v,
					},
				},
			}
			i += 1
		}
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

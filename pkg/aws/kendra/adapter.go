package kendra

import (
	"strings"

	"github.com/DSSD-Madison/gmu/pkg/model/search"
	awskendra "github.com/aws/aws-sdk-go-v2/service/kendra"
)

func convertToSearchResults(out awskendra.QueryOutput, pageNum int) search.Results {
	results := search.Results{
		Results: make(map[string]search.Result),
		Filters: make([]search.FilterCategory, len(out.FacetResults)),
	}

	for _, item := range out.ResultItems {
		title := trimExtension(*item.DocumentTitle.Text)

		var res search.Result

		if result, ok := results.Results[title]; !ok {
			res = search.Result{
				Title:    title,
				Excerpts: make([]search.Excerpt, 0),
				Link:     *item.DocumentURI,
			}
			results.Order = append(results.Order, title)
		} else {
			res = result
		}

		pageNum := 0
		for _, a := range item.DocumentAttributes {
			if *a.Key == "_excerpt_page_number" {
				pageNum = int(*a.Value.LongValue)
			}
		}

		res.Excerpts = append(res.Excerpts, search.Excerpt{
			Text:    *item.DocumentExcerpt.Text,
			PageNum: pageNum,
		})
		results.Results[res.Title] = res
	}

	// Set total count
	results.Count = int(*out.TotalNumberOfResults)

	// Convert Filters
	filterNamesMap := map[string]string{
		"Author":     "Authors",
		"Keyword":    "Keywords",
		"Region":     "Regions",
		"Category":   "Categories",
		"Source":     "Source",
		"_file_type": "File Type",
	}

	for i, facetRes := range out.FacetResults {
		name, ok := filterNamesMap[*facetRes.DocumentAttributeKey]
		if !ok {
			name = *facetRes.DocumentAttributeKey
		}

		filterCategory := search.FilterCategory{
			Category: *facetRes.DocumentAttributeKey,
			Options:  make([]search.FilterOption, len(facetRes.DocumentAttributeValueCountPairs)),
			Name:     name,
		}

		for j, attribute := range facetRes.DocumentAttributeValueCountPairs {
			filterCategory.Options[j] = search.FilterOption{
				Label: *attribute.DocumentAttributeValue.StringValue,
				Count: *attribute.Count,
			}
		}

		results.Filters[i] = filterCategory
	}

	// Calculate pagination
	calculatedPages := (results.Count + 9) / 10
	totalPages := calculatedPages
	if totalPages > 10 {
		totalPages = 10
	}

	results.PageStatus = search.PageStatus{
		CurrentPage: pageNum,
		PrevPage:    pageNum - 1,
		NextPage:    pageNum + 1,
		HasPrev:     pageNum > 1,
		HasNext:     pageNum < totalPages,
		TotalPages:  totalPages,
	}

	return results
}

func convertSuggestions(out awskendra.GetQuerySuggestionsOutput) search.Suggestions {
	suggestions := search.Suggestions{
		Suggestions: make([]string, len(out.Suggestions)),
	}

	for _, item := range out.Suggestions {
		suggestions.Suggestions = append(suggestions.Suggestions, *item.Value.Text.Text)
	}

	return suggestions
}

func trimExtension(s string) string {
	if strings.HasSuffix(s, ".pdf") {
		return strings.TrimSuffix(s, ".pdf")
	}
	if strings.HasSuffix(s, ".docx") {
		return strings.TrimSuffix(s, ".docx")
	}
	return s
}

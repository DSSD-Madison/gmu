package models

import (
	"io"
	"testing"

	_ "github.com/DSSD-Madison/gmu/testing_init"
)

func mockResults() KendraResults {
	return KendraResults{
		Results: []KendraResult{
			{
				Title:   "title 1",
				Excerpt: "excerpt 1",
				Link:    "link 1",
				PageNum: 1,
			},
			{
				Title:   "title 2",
				Excerpt: "excerpt 2",
				Link:    "link 2",
				PageNum: 2,
			},
			{
				Title:   "title 3",
				Excerpt: "excerpt 3",
				Link:    "link 3",
				PageNum: 3,
			},
		},
		Query: "query",
		Count: 3,
		Filters: []FilterCategory{
			{
				Category: "category",
				Options: []FilterOption{
					{
						Label: "label",
						Count: 1,
					},
				},
			},
		},
	}
}

func mockSuggestions() KendraSuggestions {
	return KendraSuggestions{
		Suggestions: []string{
			"suggestion1",
			"suggestion2",
		},
	}
}

func TestTemplateExecution(t *testing.T) {
	tmpls := NewTemplate()
	tests := []struct {
		name    string
		tmplStr string
		data    interface{}
	}{
		{"Search Template", "search", mockResults()},
		{"Results and Filters Template", "results", mockResults()},
		{"Suggestions Template", "suggestions", mockSuggestions()},
		{"Home", "index", nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmplStr := test.tmplStr
			tmpl, ok := tmpls.templates[tmplStr]
			if !ok {
				t.Errorf("Could not find template %s\n", tmplStr)
			}
			err := tmpl.Execute(io.Discard, test.data)
			if err != nil {
				t.Errorf("Failed to execute template %s: %q\n", tmplStr, err)
			}
		})
	}
}

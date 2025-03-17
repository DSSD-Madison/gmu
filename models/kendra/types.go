package kendra

// Excerpt provides the text and page number associated with a result.
type Excerpt struct {
	Text    string
	PageNum int
}

// KendraResult holds the necessary data for each individual result of a search.
type KendraResult struct {
	Title    string
	Excerpts []Excerpt
	Link     string
}

// KendraResults holds the results and other information
// that will be provided with a request.
type KendraResults struct {
	Results map[string]KendraResult
	Query   string
	Count   int
	Filters []FilterCategory
}

// KendraSuggestions is a wrapper for holding a list of Kendra Suggestions.
type KendraSuggestions struct {
	Suggestions []string
}

// FilterOption represents a clickable filter on the results page.
type FilterOption struct {
	Label string
	Count int32
}

// FilterCategory represents a dropdown category on the results page.
type FilterCategory struct {
	Category string
	Options  []FilterOption
	Name     string
}

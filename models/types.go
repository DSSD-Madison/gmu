package models

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
	Results     map[string]KendraResult
	Query       string
	Count       int
	CurrentPage int
	TotalPages  int
	Filters     []FilterCategory
}

type KendraSuggestions struct {
	Suggestions []string
}

type FilterOption struct {
	Label string
	Count int32
}

type FilterCategory struct {
	Category string
	Options  []FilterOption
	Name     string
}

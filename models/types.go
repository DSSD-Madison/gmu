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
	HasPrev     bool
	HasNext     bool
	PrevPage    int
	NextPage    int
	Filters     []FilterCategory
}

type KendraSuggestions struct {
	Suggestions []string
}

type FilterOption struct {
	Label    string
	Selected bool
	Count    int32
}

type FilterCategory struct {
	Category string
	Options  []FilterOption
	Name     string
}

type Filter struct {
	Name            string
	SelectedFilters []string
}

type UrlData struct {
	Query   string
	Filters []Filter
	Page    int
}

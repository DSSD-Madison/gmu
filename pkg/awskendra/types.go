package awskendra

// Excerpt holds information about a
// Result's excerpt for use in rendering the UI
type Excerpt struct {
	Text    string
	PageNum int
}

// KendraResult holds information about a single
// Result from AWS Kendra used for rendering the UI
type KendraResult struct {
	Title       string
	Excerpts    []Excerpt
	Link        string
	Image       string
	Authors     []string
	Regions     []string
	Keywords    []string
	PublishDate string
	Categories  []string
	Abstract    string
	Source      string
	UUID        string
}

// PageStatus holds information about the state of the page
// for use in Pagination
type PageStatus struct {
	CurrentPage int
	HasPrev     bool
	HasNext     bool
	PrevPage    int
	NextPage    int
	TotalPages  int
}

// KendraResults holds information about the entire
// response to a search query.
type KendraResults struct {
	IsStoringUrl bool
	Results      map[string]KendraResult
	Query        string
	Count        int
	PageStatus   PageStatus
	Filters      []FilterCategory
	UrlData      UrlData
}

// KendraSuggestions simply holds a list of
// suggestion strings for use in suggesting
// searches in the UI.
type KendraSuggestions struct {
	Suggestions []string
}

// FilterOption represents a filter option in a
// filter category. Selected is used to determine whether
// the user has selected this filter or not. Count holds
// the number of items in this filter.
type FilterOption struct {
	Label    string
	Selected bool
	Count    int32
}

// FilterCategory represents a category/grouping of FilterOptions.
// This is displayed in the UI as a togglable dropdown of filters.
type FilterCategory struct {
	Category string
	Options  []FilterOption
	Name     string
}

// Filter is a data processing type used for processing Kendra Search output.
type Filter struct {
	Name            string
	SelectedFilters []string
}

type UrlData struct {
	IsStoringUrl bool
	Query        string
	Filters      []Filter
	Page         int
}

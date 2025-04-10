package awskendra

type Excerpt struct {
	Text    string
	PageNum int
}

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
}

type PageStatus struct {
	CurrentPage int
	HasPrev     bool
	HasNext     bool
	PrevPage    int
	NextPage    int
	TotalPages  int
}

type KendraResults struct {
	IsStoringUrl bool
	Results      map[string]KendraResult
	Query        string
	Count        int
	PageStatus   PageStatus
	Filters      []FilterCategory
	UrlData      UrlData
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
	IsStoringUrl bool
	Query        string
	Filters      []Filter
	Page         int
}

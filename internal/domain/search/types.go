package search

type Suggestions struct {
	Suggestions []string
}

type Result struct {
	ID          string
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

type Excerpt struct {
	Text    string
	PageNum int
}

type PageStatus struct {
	CurrentPage int
	HasPrev     bool
	HasNext     bool
	PrevPage    int
	NextPage    int
	TotalPages  int
}

type Results struct {
	IsStoringURL bool
	Results      map[string]Result
	Order        []string
	Query        string
	Count        int
	PageStatus   PageStatus
	Filters      []FilterCategory
	URLData      UrlData
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
	IsStoringURL bool
	Query        string
	Filters      []Filter
	Page         int
}

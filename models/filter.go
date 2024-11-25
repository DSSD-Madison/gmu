package models 

type FilterOption struct {
	Label string
	Count int
}

type FilterCategory struct {
	Category string
	Options  []FilterOption
}

package models

import "net/url"

func UrlValuesToFilters(vals url.Values) (filters []Filter) {
	for name, selected := range vals {
		filters = append(filters, Filter{
			Name: name,
			SelectedFilters: selected,
		})
	}
	return
}

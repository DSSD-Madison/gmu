package util

import (
	"strings"

	db "github.com/DSSD-Madison/gmu/pkg/db/generated"
	"github.com/DSSD-Madison/gmu/web/components"
)

func ToAuthorPairs(all []db.Author, selected []string) []components.Pair {
	seen := make(map[string]bool)
	var out []components.Pair
	for _, name := range selected {
		norm := strings.ToLower(name)
		if seen[norm] {
			continue
		}
		for _, a := range all {
			if strings.EqualFold(a.Name, name) {
				out = append(out, components.Pair{
					ID:   a.ID.String(),
					Name: a.Name,
				})
				seen[norm] = true
				break
			}
		}
	}
	return out
}

func ToKeywordPairs(all []db.Keyword, selected []string) []components.Pair {
	seen := make(map[string]bool)
	var out []components.Pair
	for _, name := range selected {
		norm := strings.ToLower(name)
		if seen[norm] {
			continue
		}
		for _, k := range all {
			if strings.EqualFold(k.Name, name) {
				out = append(out, components.Pair{
					ID:   k.ID.String(),
					Name: k.Name,
				})
				seen[norm] = true
				break
			}
		}
	}
	return out
}

func ToRegionPairs(all []db.Region, selected []string) []components.Pair {
	seen := make(map[string]bool)
	var out []components.Pair
	for _, name := range selected {
		norm := strings.ToLower(name)
		if seen[norm] {
			continue
		}
		for _, r := range all {
			if strings.EqualFold(r.Name, name) {
				out = append(out, components.Pair{
					ID:   r.ID.String(),
					Name: r.Name,
				})
				seen[norm] = true
				break
			}
		}
	}
	return out
}

func ToCategoryPairs(all []db.Category, selected []string) []components.Pair {
	seen := make(map[string]bool)
	var out []components.Pair
	for _, name := range selected {
		norm := strings.ToLower(name)
		if seen[norm] {
			continue
		}
		for _, c := range all {
			if strings.EqualFold(c.Name, name) {
				out = append(out, components.Pair{
					ID:   c.ID.String(),
					Name: c.Name,
				})
				seen[norm] = true
				break
			}
		}
	}
	return out
}

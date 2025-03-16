package routes

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/DSSD-Madison/gmu/models"
)

const MinQueryLength = 3

func SearchSuggestions(c echo.Context) error {
	query := c.FormValue("query")

	if len(query) == 0 {
		return nil
	}
	suggestions, err := models.GetSuggestions(query)
	// TODO: add error status code
	if err != nil {
		return nil
	}
	return c.Render(http.StatusOK, "suggestions", suggestions)
}

func Search(c echo.Context) error {
	query := c.FormValue("query")
	pageNum := c.FormValue("page")

	filters, _ := c.FormParams()

	delete(filters, "query")
	delete(filters, "page")

	if len(query) == 0 {
		return Home(c)
	}

	num, err := strconv.Atoi(strings.TrimSpace(pageNum))
	if err != nil {
		num = 1
	}

	var filterList []models.Filter

	for key, values := range filters {
		filterList = append(filterList, models.Filter{
			Name:            key,
			SelectedFilters: values,
		})
	}

	fmt.Println(filters)

	urlData := models.UrlData{
		Query:   query,
		Filters: filterList,
		Page:    num,
	}
	if len(query) < MinQueryLength {
		return echo.NewHTTPError(http.StatusBadRequest, "Query too short")
	}
	// Check if the request is coming from HTMX
	target := c.Request().Header.Get("HX-Target")

	if target == "root" || target == "" {
		return c.Render(http.StatusOK, "search-standalone", urlData)
	} else if target == "results-container" {
		tempResults := models.MakeQuery(query, nil, 1)

		results := models.MakeQuery(query, filters, num)

		results.Filters = tempResults.Filters

		for i, cat := range results.Filters {
			if selectedOptions, exists := filters[cat.Category]; exists {
				for idx, o := range cat.Options {
					for _, selected := range selectedOptions {
						if o.Label == selected {
							results.Filters[i].Options[idx].Selected = true
							break
						}
					}
				}
			}
		}

		return c.Render(http.StatusOK, "results", results)
	} else if target == "results-content-container" {
		results := models.MakeQuery(query, filters, num)
		return c.Render(http.StatusOK, "results-container", results)
	} else if target == "results-and-pagination" {
		results := models.MakeQuery(query, filters, num)
		return c.Render(http.StatusOK, "results-and-pagination", results)
	} else {
		return c.Render(http.StatusOK, "search", urlData)
	}

}

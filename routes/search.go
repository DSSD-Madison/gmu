package routes

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/DSSD-Madison/gmu/db"
	"github.com/DSSD-Madison/gmu/internal/db_helpers"
	"github.com/DSSD-Madison/gmu/components"
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
	return models.Render(c, http.StatusOK, components.Suggestions(suggestions))
}

func Search(c echo.Context, db_querier *db.Queries) error {
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

	urlData := models.UrlData{
		Query:        query,
		Filters:      filterList,
		Page:         num,
		IsStoringUrl: true,
	}
	if len(query) < MinQueryLength {
		return echo.NewHTTPError(http.StatusBadRequest, "Query too short")
	}
	// Check if the request is coming from HTMX
	target := c.Request().Header.Get("HX-Target")

	if target == "root" {
		return models.Render(c, http.StatusOK, components.Search(models.KendraResults{UrlData: urlData}))
	} else if target == "results-container" {
		if len(filterList) == 0 {
			results, err := getResults(c, db_querier, query, filters, num)
			if err != nil {
				return err
			}
			return models.Render(c, http.StatusOK, components.ResultsPage(results))
		}
		tempResults := models.MakeQuery(query, nil, 1)
		results, err := getResults(c, db_querier, query, filters, num)
		if err != nil {
			return err
		}
		results.Filters = tempResults.Filters
		selectFilters(filters, &results)
		return models.Render(c, http.StatusOK, components.ResultsPage(results))
	} else if target == "results-content-container" {
		results, err := getResults(c, db_querier, query, filters, num)
		if err != nil {
			return err
		}
		return models.Render(c, http.StatusOK, components.ResultsContainer(results))
	} else if target == "results-and-pagination" {
		results, err := getResults(c, db_querier, query, filters, num)
		if err != nil {
			return err
		}
		return models.Render(c, http.StatusOK, components.ResultsAndPagination(results))
	} else {
		return models.Render(c, http.StatusOK, components.SearchHome(models.KendraResults{UrlData: urlData}))
	}

}

func getResults(c echo.Context, queries *db.Queries, query string, filters url.Values, num int) (models.KendraResults, error) {
	results := models.MakeQuery(query, filters, num)
	err := db_helpers.AddImagesToResults(results, c, queries)
	if err != nil {
		return models.KendraResults{}, err
	}
	return results, nil
}

func selectFilters(filters url.Values, results *models.KendraResults) {
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
}



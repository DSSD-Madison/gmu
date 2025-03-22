package routes

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/DSSD-Madison/gmu/pkg/awskendra"
	"github.com/DSSD-Madison/gmu/pkg/db"
	db_util "github.com/DSSD-Madison/gmu/pkg/db/util"
	"github.com/DSSD-Madison/gmu/web/components"
)

const MinQueryLength = 3

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

	var filterList []awskendra.Filter

	for key, values := range filters {
		filterList = append(filterList, awskendra.Filter{
			Name:            key,
			SelectedFilters: values,
		})
	}

	urlData := awskendra.UrlData{
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
		return awskendra.Render(c, http.StatusOK, components.Search(awskendra.KendraResults{UrlData: urlData}))
	} else if target == "results-container" {
		if len(filterList) == 0 {
			results, err := getResults(c, db_querier, query, filters, num)
			if err != nil {
				return err
			}
			return awskendra.Render(c, http.StatusOK, components.ResultsPage(results))
		}
		tempResults := awskendra.MakeQuery(query, nil, 1)
		results, err := getResults(c, db_querier, query, filters, num)
		if err != nil {
			return err
		}
		results.Filters = tempResults.Filters
		selectFilters(filters, &results)
		return awskendra.Render(c, http.StatusOK, components.ResultsPage(results))
	} else if target == "results-content-container" {
		results, err := getResults(c, db_querier, query, filters, num)
		if err != nil {
			return err
		}
		return awskendra.Render(c, http.StatusOK, components.ResultsContainer(results))
	} else if target == "results-and-pagination" {
		results, err := getResults(c, db_querier, query, filters, num)
		if err != nil {
			return err
		}
		return awskendra.Render(c, http.StatusOK, components.ResultsAndPagination(results))
	} else {
		return awskendra.Render(c, http.StatusOK, components.SearchHome(awskendra.KendraResults{UrlData: urlData}))
	}

}

func getResults(c echo.Context, queries *db.Queries, query string, filters url.Values, num int) (awskendra.KendraResults, error) {
	results := awskendra.MakeQuery(query, filters, num)
	db_util.AddImagesToResults(results, c, queries)
	return results, nil
}

func selectFilters(filters url.Values, results *awskendra.KendraResults) {
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

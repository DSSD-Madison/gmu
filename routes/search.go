package routes

import (
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"

	"github.com/DSSD-Madison/gmu/pkg/awskendra"
	db_util "github.com/DSSD-Madison/gmu/pkg/db/util"
	"github.com/DSSD-Madison/gmu/web"
	"github.com/DSSD-Madison/gmu/web/components"
)

const MinQueryLength = 3

type searchRequest struct {
	query         string
	pageNum       int
	kendraFilters []awskendra.Filter
	filters       url.Values
	urlData       awskendra.UrlData
	target        string
}

func parseSearchRequest(c echo.Context) searchRequest {
	query := c.FormValue("query")
	pageNumStr := c.FormValue("page")

	filters, _ := c.FormParams()
	delete(filters, "query")
	delete(filters, "page")

	pageNum := parsePageNum(pageNumStr)

	filterList := convertFilterstoKendra(filters)

	urlData := awskendra.UrlData{
		Query:        query,
		Filters:      filterList,
		Page:         pageNum,
		IsStoringUrl: true,
	}

	// Check if the request is coming from HTMX
	target := c.Request().Header.Get("HX-Target")

	return searchRequest{
		query:         query,
		pageNum:       pageNum,
		kendraFilters: filterList,
		filters:       filters,
		urlData:       urlData,
		target:        target,
	}
}

func (h *Handler) Search(c echo.Context) error {
	r := parseSearchRequest(c)
	if r.query == "" {
		return h.Home(c)
	}

	if len(r.query) < MinQueryLength {
		return echo.NewHTTPError(http.StatusBadRequest, "Query too short")
	}

	results, err := selectResultsFromTarget(r.target, r, h, c)
	if err != nil {
		return err
	}
	component, err := selectComponentTarget(r.target, r, results)
	if err != nil {
		return err
	}
	return web.Render(c, http.StatusOK, component)
}

func parsePageNum(pageNumStr string) int {
	num, err := strconv.Atoi(strings.TrimSpace(pageNumStr))
	if err != nil || num < 1 {
		return 1
	}
	return num
}

func selectResultsFromTarget(target string, r searchRequest, h *Handler, c echo.Context) (awskendra.KendraResults, error) {
	switch target {
	case "root", "":
		return awskendra.KendraResults{UrlData: r.urlData}, nil
	case "results-container", "results-content-container", "results-and-pagination":
		results, err := h.getResults(c, r.query, r.filters, r.pageNum)
		if err != nil {
			return awskendra.KendraResults{}, err
		}
		if r.target == "results-container" && len(r.kendraFilters) > 0 {
			tempResults := h.kendra.MakeQuery(r.query, nil, 1)
			results.Filters = tempResults.Filters
			selectFilters(r.filters, &results)
		}
		return results, nil
	default:
		return awskendra.KendraResults{}, fmt.Errorf("Failed to select results from target header.")
	}
}

func selectComponentTarget(target string, r searchRequest, results awskendra.KendraResults) (templ.Component, error) {
	switch target {
	case "root":
		return components.Search(awskendra.KendraResults{UrlData: r.urlData}), nil
	case "":
		return components.SearchHome(awskendra.KendraResults{UrlData: r.urlData}), nil
	case "results-container", "results-content-container":
		return components.ResultsPage(results), nil
	case "results-and-pagination":
		return components.ResultsAndPagination(results), nil
	default:
		return templ.NopComponent, fmt.Errorf("Failed to determine target component from target header.")
	}
}

func convertFilterstoKendra(filters url.Values) []awskendra.Filter {
	var filterList []awskendra.Filter
	for key, values := range filters {
		filterList = append(filterList, awskendra.Filter{
			Name:            key,
			SelectedFilters: values,
		})
	}
	return filterList
}

func (h *Handler) getResults(c echo.Context, query string, filters url.Values, num int) (awskendra.KendraResults, error) {
	results := h.kendra.MakeQuery(query, filters, num)
	db_util.AddImagesToResults(results, c, h.db)
	return results, nil
}

func selectFilters(filters url.Values, results *awskendra.KendraResults) {
	for i, cat := range results.Filters {
		if selectedOptions, exists := filters[cat.Category]; exists {
			for idx, o := range cat.Options {
				if slices.Contains(selectedOptions, o.Label) {
					results.Filters[i].Options[idx].Selected = true
				}
			}
		}
	}
}

package routes

import (
	"net/http"
	"net/url"
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

func (h *Handler) Search(c echo.Context) error {
	query := c.FormValue("query")
	pageNumStr := c.FormValue("page")

	filters, _ := c.FormParams()
	delete(filters, "query")
	delete(filters, "page")

	if query == "" {
		return h.Home(c)
	}

	if len(query) < MinQueryLength {
		return echo.NewHTTPError(http.StatusBadRequest, "Query too short")
	}

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
	if target == "root" {
		return web.Render(c, http.StatusOK, components.Search(awskendra.KendraResults{UrlData: urlData}))
	}

	results, err := h.getResults(c, query, filters, pageNum)
	if err != nil {
		return err
	}

	component := selectComponentTarget(target, urlData, results)

	if target == "results-container" && len(filterList) > 0 {
		tempResults := h.kendra.MakeQuery(query, nil, 1)
		results.Filters = tempResults.Filters
		selectFilters(filters, &results)
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

func selectComponentTarget(target string, urlData awskendra.UrlData, results awskendra.KendraResults) templ.Component {
	switch target {
	case "results-container", "results-content-container":
		return components.ResultsPage(results)
	case "results-and-pagination":
		return components.ResultsAndPagination(results)
	default:
		return components.SearchHome(awskendra.KendraResults{UrlData: urlData})
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

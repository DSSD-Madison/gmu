package routes

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/DSSD-Madison/gmu/pkg/middleware"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"

	"github.com/DSSD-Madison/gmu/pkg/awskendra"
	"github.com/DSSD-Madison/gmu/pkg/logger"
	"github.com/DSSD-Madison/gmu/pkg/services"
	"github.com/DSSD-Madison/gmu/web"
	"github.com/DSSD-Madison/gmu/web/components"
)

const MinQueryLength = 3

type SearchHandler struct {
	log      logger.Logger
	searcher services.Searcher
}

func NewSearchHandler(log logger.Logger, searcher services.Searcher) *SearchHandler {
	handlerLogger := log.With("Handler", "Search")
	return &SearchHandler{
		log:      handlerLogger,
		searcher: searcher,
	}
}

type searchRequest struct {
	query   string
	pageNum int
	filters url.Values
	urlData awskendra.UrlData
	target  string
}

func parseSearchRequest(c echo.Context) (searchRequest, error) {
	query := c.FormValue("query")
	pageNumStr := c.FormValue("page")

	filters, err := c.FormParams()
	if err != nil {
		filters = make(url.Values)
	}
	delete(filters, "query")
	delete(filters, "page")

	pageNum := parsePageNum(pageNumStr)

	kendraFilterList := convertFilterstoKendra(filters)

	urlData := awskendra.UrlData{
		Query:        query,
		Filters:      kendraFilterList,
		Page:         pageNum,
		IsStoringUrl: true,
	}

	if query != "" && len(query) < MinQueryLength {
		return searchRequest{}, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Query must be at least %d characters", MinQueryLength))
	}

	target := c.Request().Header.Get("HX-Target")

	return searchRequest{
		query:   query,
		pageNum: pageNum,
		filters: filters,
		urlData: urlData,
		target:  target,
	}, nil
}

func (h *SearchHandler) Search(c echo.Context) error {
	ctx := c.Request().Context()
	req, err := parseSearchRequest(c)
	if err != nil {
		h.log.WarnContext(ctx, "Failed to parse search request", "error", err)

		if httpErr, ok := err.(*echo.HTTPError); ok {
			return httpErr
		}
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid search parameters")
	}
	if req.query == "" {
		h.log.DebugContext(ctx, "No search query provided, rendering initial search component")
		return web.Render(c, http.StatusOK, components.Search(awskendra.KendraResults{UrlData: req.urlData}))
	}

	h.log.InfoContext(ctx, "Performing search", "query", req.query, "page", req.pageNum, "filters", req.filters)

	results, err := selectResultsFromTarget(ctx, h, req)
	if err != nil {
		h.log.ErrorContext(ctx, "Search service failed", "query", req.query, "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Search failed")
	}

	results.UrlData = req.urlData

	isAuthorized, isMaster := middleware.GetSessionFlags(c)

	component, err := selectComponentTarget(req.target, req, results, isAuthorized, isMaster)
	if err != nil {
		h.log.ErrorContext(ctx, "Failed to select component target", "target", req.target, "error", err)
		// Fallback or error
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal error")
	}

	h.log.DebugContext(ctx, "Rendering search results", "target", req.target, "result_count", results.Count)
	return web.Render(c, http.StatusOK, component)
}

func parsePageNum(pageNumStr string) int {
	num, err := strconv.Atoi(strings.TrimSpace(pageNumStr))
	if err != nil || num < 1 {
		return 1
	}
	return num
}

func selectResultsFromTarget(ctx context.Context, h *SearchHandler, req searchRequest) (awskendra.KendraResults, error) {
	if h == nil {
		return awskendra.KendraResults{}, fmt.Errorf("Cannot get results from nil handler")
	}
	switch req.target {
	case "root", "":
		return awskendra.KendraResults{UrlData: req.urlData}, nil
	case "results-container", "results-content-container", "results-and-pagination":
		results, err := h.searcher.SearchDocuments(ctx, req.query, req.filters, req.pageNum)
		if err != nil {
			h.log.ErrorContext(ctx, "Search service failed", "query", req.query, "error", err)
			return awskendra.KendraResults{}, err
		}
		kendraFilters := convertFilterstoKendra(req.filters)
		if req.target == "results-container" && len(kendraFilters) > 0 {
			tempResults, err := h.searcher.SearchDocuments(ctx, req.query, nil, 1)
			if err != nil {
				h.log.ErrorContext(ctx, "Search service failed", "query", req.query, "error", err)
				return awskendra.KendraResults{}, err
			}
			results.Filters = tempResults.Filters
			selectFilters(req.filters, &results)
		}
		return results, nil
	default:
		h.log.ErrorContext(ctx, "Failed to select results from target header")
		return awskendra.KendraResults{}, fmt.Errorf("Failed to select results from target header")
	}
}

func selectComponentTarget(target string, r searchRequest, results awskendra.KendraResults, isAuthorized bool, isMaster bool) (templ.Component, error) {
	switch target {
	case "root":
		return components.Search(results), nil
	case "":
		return components.SearchHome(awskendra.KendraResults{UrlData: r.urlData}, isAuthorized, isMaster), nil
	case "results-container", "results-content-container":
		return components.ResultsPage(results, isAuthorized), nil
	case "results-and-pagination":
		return components.ResultsAndPagination(results, isAuthorized), nil
	default:
		return nil, fmt.Errorf("unknown HX-Target for search: %s", target)
	}
}

func convertFilterstoKendra(filters url.Values) []awskendra.Filter {
	filterList := make([]awskendra.Filter, len(filters))
	i := 0
	for key, values := range filters {
		filterList[i].Name = key
		filterList[i].SelectedFilters = values
		i += 1
	}
	return filterList
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

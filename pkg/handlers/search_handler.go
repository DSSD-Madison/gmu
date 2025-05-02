package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"

	"github.com/DSSD-Madison/gmu/pkg/core/logger"
	"github.com/DSSD-Madison/gmu/pkg/model/search"
	"github.com/DSSD-Madison/gmu/pkg/services"
	"github.com/DSSD-Madison/gmu/web"
	"github.com/DSSD-Madison/gmu/web/components"
)

const MinQueryLength = 3

type SearchHandler struct {
	log            logger.Logger
	searcher       services.Searcher
	sessionManager services.SessionManager
}

func NewSearchHandler(log logger.Logger, searcher services.Searcher, sessionManager services.SessionManager) *SearchHandler {
	handlerLogger := log.With("Handler", "Search")
	return &SearchHandler{
		log:            handlerLogger,
		sessionManager: sessionManager,
		searcher:       searcher,
	}
}

type searchRequest struct {
	query   string
	pageNum int
	filters url.Values
	urlData search.UrlData
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

	urlData := search.UrlData{
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
		return web.Render(c, http.StatusOK, components.Search(search.Results{UrlData: req.urlData}))
	}

	h.log.InfoContext(ctx, "Performing search", "query", req.query, "page", req.pageNum, "filters", req.filters)

	results, err := selectResultsFromTarget(ctx, h, req)
	if err != nil {
		h.log.ErrorContext(ctx, "Search service failed", "query", req.query, "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Search failed")
	}

	results.UrlData = req.urlData

	isAuthorized := h.sessionManager.IsAuthenticated(c)
	isMaster := h.sessionManager.IsMaster(c)

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

func selectResultsFromTarget(ctx context.Context, h *SearchHandler, req searchRequest) (search.Results, error) {
	if h == nil {
		return search.Results{}, fmt.Errorf("Cannot get results from nil handler")
	}
	switch req.target {
	case "root", "":
		return search.Results{UrlData: req.urlData}, nil
	case "results-container", "results-content-container", "results-and-pagination":
		results, err := h.searcher.SearchDocuments(ctx, req.query, req.filters, req.pageNum)
		if err != nil {
			h.log.ErrorContext(ctx, "Search service failed", "query", req.query, "error", err)
			return search.Results{}, err
		}
		kendraFilters := convertFilterstoKendra(req.filters)
		if req.target == "results-container" && len(kendraFilters) > 0 {
			tempResults, err := h.searcher.SearchDocuments(ctx, req.query, nil, 1)
			if err != nil {
				h.log.ErrorContext(ctx, "Search service failed", "query", req.query, "error", err)
				return search.Results{}, err
			}
			results.Filters = tempResults.Filters
			selectFilters(req.filters, &results)
		}
		return results, nil
	default:
		h.log.ErrorContext(ctx, "Failed to select results from target header")
		return search.Results{}, fmt.Errorf("Failed to select results from target header")
	}
}

func selectComponentTarget(target string, r searchRequest, results search.Results, isAuthorized bool, isMaster bool) (templ.Component, error) {
	switch target {
	case "root":
		return components.Search(results), nil
	case "":
		return components.SearchHome(search.Results{UrlData: r.urlData}, isAuthorized, isMaster), nil
	case "results-container", "results-content-container":
		return components.ResultsPage(results, isAuthorized), nil
	case "results-and-pagination":
		return components.ResultsAndPagination(results, isAuthorized), nil
	default:
		return nil, fmt.Errorf("unknown HX-Target for search: %s", target)
	}
}

func convertFilterstoKendra(filters url.Values) []search.Filter {
	filterList := make([]search.Filter, len(filters))
	i := 0
	for key, values := range filters {
		filterList[i].Name = key
		filterList[i].SelectedFilters = values
		i += 1
	}
	return filterList
}

func selectFilters(filters url.Values, results *search.Results) {
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

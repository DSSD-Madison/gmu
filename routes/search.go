package routes

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/DSSD-Madison/gmu/internal/db_helpers"
	"github.com/DSSD-Madison/gmu/components"
	"github.com/DSSD-Madison/gmu/models"
)

const MinQueryLength = 3

func (h *Handler) SearchSuggestions(c echo.Context) error {
	query := c.FormValue("query")

	if len(query) == 0 {
		return nil
	}
	suggestions, err := h.kendra.GetSuggestions(query)
	// TODO: add error status code
	if err != nil {
		return nil
	}
	return models.Render(c, http.StatusOK, components.Suggestions(suggestions))
}

func (h *Handler) Search(c echo.Context) error {
	query := c.FormValue("query")
	pageNum := c.FormValue("page")

	filters, _ := c.FormParams()

	delete(filters, "query")
	delete(filters, "page")

	if len(query) == 0 {
		return h.Home(c)
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
			results, err := h.getResults(c, query, filters, num)
			if err != nil {
				return err
			}
			return models.Render(c, http.StatusOK, components.ResultsPage(results))
		}
		tempResults := h.kendra.MakeQuery(query, nil, 1)
		results, err := h.getResults(c, query, filters, num)
		if err != nil {
			return err
		}
		results.Filters = tempResults.Filters
		selectFilters(filters, &results)
		return models.Render(c, http.StatusOK, components.ResultsPage(results))
	} else if target == "results-content-container" {
		results, err := h.getResults(c, query, filters, num)
		if err != nil {
			return err
		}
		return models.Render(c, http.StatusOK, components.ResultsContainer(results))
	} else if target == "results-and-pagination" {
		results, err := h.getResults(c, query, filters, num)
		if err != nil {
			return err
		}
		return models.Render(c, http.StatusOK, components.ResultsAndPagination(results))
	} else {
		return models.Render(c, http.StatusOK, components.SearchHome(models.KendraResults{UrlData: urlData}))
	}

}

func (h Handler) getResults(c echo.Context, query string, filters url.Values, num int) (models.KendraResults, error) {
	results := h.kendra.MakeQuery(query, filters, num)
	err := db_helpers.AddImagesToResults(results, c, h.db)
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



package routes

import (
	"net/http"
	"strconv"

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

	if len(query) == 0 {
		return Home(c)
	}

	num, err := strconv.Atoi(pageNum)
	if err != nil {
		num = 1
	}

	if len(query) < MinQueryLength {
		return echo.NewHTTPError(http.StatusBadRequest, "Query too short")
	}
	// Check if the request is coming from HTMX
	target := c.Request().Header.Get("HX-Target")

	if target == "root" || target == "" {
		return c.Render(http.StatusOK, "search-standalone", query)
	} else if target == "results-container" {
		results := models.MakeQuery(query, nil, num)
		return c.Render(http.StatusOK, "results", results)
	} else {
		return c.Render(http.StatusOK, "search", query)
	}

}

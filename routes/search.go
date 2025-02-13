package routes

import (
	"net/http"

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
	context := c.FormValue("context")

	if len(query) < MinQueryLength {
		return echo.NewHTTPError(http.StatusBadRequest, "Query too short")
	}
	if context == "home" {
		return c.Render(http.StatusOK, "search", query)
	}
	results := models.MakeQuery(query, nil)

	return c.Render(http.StatusOK, "results", results)
}

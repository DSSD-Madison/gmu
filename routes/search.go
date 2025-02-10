package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/DSSD-Madison/gmu/models"
)

const MinQueryLength = 3

func fetchSearchPage(c echo.Context) error {
	query := c.FormValue("query")
	if len(query) < MinQueryLength {
		return echo.NewHTTPError(http.StatusBadRequest, "Query too short")
	}
	return c.Render(http.StatusOK, "search", query)
}

func Search(c echo.Context) error {
	query := c.FormValue("query")
	if len(query) < MinQueryLength {
		return echo.NewHTTPError(http.StatusBadRequest, "Query too short")
	}
	results := models.MakeQuery(query, nil)

	return c.Render(http.StatusOK, "results", results)
}

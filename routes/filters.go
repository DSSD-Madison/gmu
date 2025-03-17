package routes

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/DSSD-Madison/gmu/db"
	"github.com/DSSD-Madison/gmu/models"
)

func Filters(c echo.Context, queries *db.Queries) error {
	// Ensure form values are parsed
	if err := c.Request().ParseForm(); err != nil {
		return c.String(http.StatusBadRequest, "Failed to parse form")
	}

	// Retrieve and process filters
	selectedFilters := make(map[string][]string)
	query := ""
	for key, values := range c.Request().Form {
		// All form values get collapsed into a key. We need to put the query string into the URL
		if key == "query" {
			query = values[0]
			continue
		}
		cleanKey := strings.TrimPrefix(key, "filters[")
		cleanKey = strings.TrimSuffix(cleanKey, "][]")

		selectedFilters[cleanKey] = values
	}

	results := models.MakeQuery(query, selectedFilters)
	err := addImagesToResults(results, c, queries)
	if err != nil {
		return err;
	}

	return c.Render(http.StatusOK, "results-container", results)
}

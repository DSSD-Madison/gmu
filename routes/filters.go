package routes

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/DSSD-Madison/gmu/models"
)

func Filters(c echo.Context) error {
	// Ensure form values are parsed
	if err := c.Request().ParseForm(); err != nil {
		return c.String(http.StatusBadRequest, "Failed to parse form")
	}

	// Retrieve and process filters
	selectedFilters := make(map[string][]string)
	query := ""
	for key, values := range c.Request().Form {
		if key == "query" {
			query = values[0]
			continue
		}
		cleanKey := strings.TrimPrefix(key, "filters[")
		cleanKey = strings.TrimSuffix(cleanKey, "][]")

		fmt.Println(cleanKey, values)
		selectedFilters[cleanKey] = values
	}

	results := models.MakeQuery(query, selectedFilters)

	return c.Render(http.StatusOK, "results:results-container", results)
}

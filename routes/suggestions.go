package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/DSSD-Madison/gmu/pkg/awskendra"
	"github.com/DSSD-Madison/gmu/web/components"
)

func SearchSuggestions(c echo.Context) error {
	query := c.FormValue("query")

	if len(query) == 0 {
		return nil
	}
	suggestions, err := awskendra.GetSuggestions(query)
	// TODO: add error status code
	if err != nil {
		return nil
	}
	return awskendra.Render(c, http.StatusOK, components.Suggestions(suggestions))
}


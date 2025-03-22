package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/DSSD-Madison/gmu/web"
	"github.com/DSSD-Madison/gmu/web/components"
)

func (h *Handler) SearchSuggestions(c echo.Context) error {
	query := c.FormValue("query")

	if query == "" {
		return nil
	}

	suggestions, err := h.kendra.GetSuggestions(query)
	// TODO: add error status code
	if err != nil {
		return nil
	}

	return web.Render(c, http.StatusOK, components.Suggestions(suggestions))
}

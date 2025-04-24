package routes

import (
	"github.com/DSSD-Madison/gmu/pkg/handlers"
	"github.com/labstack/echo/v4"
)

func RegisterSuggestionsRoutes(e *echo.Echo, suggestionsHandler *handlers.SuggestionsHandler) {
	e.POST("/search/suggestions", suggestionsHandler.SearchSuggestions)
}

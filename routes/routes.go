package routes

import (
	"github.com/labstack/echo/v4"
)

// InitRoutes registers all the application routes
func InitRoutes(e *echo.Echo, homeHandler *HomeHandler, searchHandler *SearchHandler, suggestionsHandler *SuggestionsHandler) {
	e.GET("/", homeHandler.Home)

	e.GET("/search", searchHandler.Search)
	e.POST("/search/suggestions", suggestionsHandler.SearchSuggestions)
}

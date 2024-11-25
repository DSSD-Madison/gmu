package routes

import (
	"github.com/labstack/echo/v4"
)

// InitRoutes registers all the application routes
func InitRoutes(e *echo.Echo) {
	// Home Route
	e.GET("/", Home)

	// Search Routes
	e.POST("/search", Search)
	e.POST("/fetchSearchPage", fetchSearchPage)
	e.POST("/search/suggestions", SearchSuggestions)

	// Filters Route
	e.POST("/filters", Filters)
}

package routes

import (
	"github.com/DSSD-Madison/gmu/db"
	"github.com/labstack/echo/v4"
)

// InitRoutes registers all the application routes
func InitRoutes(e *echo.Echo, queries *db.Queries) {
	// Home Route
	e.GET("/", Home)

	// Search Routes
	e.GET("/search", func(c echo.Context) error {
		// Call the Search handler and pass both Echo context and queries
		return Search(c, queries)
	})
	e.POST("/search/suggestions", SearchSuggestions)

	// Filters Route
	e.GET("/POST", func(c echo.Context) error {
		// Call the Search handler and pass both Echo context and queries
		return Filters(c, queries)
	})
}

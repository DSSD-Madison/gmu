package routes

import (
	"github.com/DSSD-Madison/gmu/db"
	"github.com/labstack/echo/v4"
)

// InitRoutes registers all the application routes
func InitRoutes(e *echo.Echo, db_querier *db.Queries) {
	e.GET("/", Home)

	e.GET("/search", func(c echo.Context) error {
		return Search(c, db_querier)
	})
	e.POST("/search/suggestions", SearchSuggestions)
}

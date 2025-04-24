package routes

import (
	"github.com/DSSD-Madison/gmu/pkg/handlers"
	"github.com/DSSD-Madison/gmu/pkg/middleware"
	"github.com/labstack/echo/v4"
)

func RegisterDatabaseRoutes(e *echo.Echo, databaseHandler *handlers.DatabaseHandler) {
	// --- Database Search Routes ---
	e.GET("/authors", databaseHandler.DatabaseSearchAuthors, middleware.RequireAuth)
	e.GET("/keywords", databaseHandler.DatabaseSearchKeywords, middleware.RequireAuth)
	e.GET("/regions", databaseHandler.DatabaseSearchRegions, middleware.RequireAuth)
	e.GET("/categories", databaseHandler.DatabaseSearchCategories, middleware.RequireAuth)
}

package routes

import (
	"github.com/DSSD-Madison/gmu/pkg/handlers"
	"github.com/DSSD-Madison/gmu/pkg/services"
	"github.com/labstack/echo/v4"
)

func RegisterDatabaseRoutes(e *echo.Echo, databaseHandler *handlers.DatabaseHandler, sessionManager services.SessionManager) {
	// --- Database Search Routes ---
	e.GET("/authors", databaseHandler.DatabaseSearchAuthors, sessionManager.RequireAuth)
	e.GET("/keywords", databaseHandler.DatabaseSearchKeywords, sessionManager.RequireAuth)
	e.GET("/regions", databaseHandler.DatabaseSearchRegions, sessionManager.RequireAuth)
	e.GET("/categories", databaseHandler.DatabaseSearchCategories, sessionManager.RequireAuth)
}

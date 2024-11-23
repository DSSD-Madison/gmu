package routes

import (
	"github.com/labstack/echo/v4"
)

// InitRoutes registers all the application routes
func InitRoutes(e *echo.Echo) {
	// Home Route
	e.GET("/", Home)

	// Search Route
	e.POST("/loadsearch", LoadSearch)
	e.POST("/search", Search)
}

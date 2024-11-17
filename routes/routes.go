package routes

import (
	"github.com/labstack/echo/v4"
)

// InitRoutes registers all the application routes
func InitRoutes(e *echo.Echo) {
	// Home Route
	e.GET("/", Home)

	// Search Route
	e.POST("/search", Search)

	// Book Route
	e.GET("/page/:id", Book)

	// Filter Route
	e.POST("/filter/:id", Filter)
}

package routes

import (
	"github.com/labstack/echo/v4"
)

// InitRoutes registers all the application routes
func InitRoutes(e *echo.Echo, h Handler) {
	e.GET("/", h.Home)

	e.GET("/search", h.Search)
	e.POST("/search/suggestions", h.SearchSuggestions)
}

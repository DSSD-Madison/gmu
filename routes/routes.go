package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/DSSD-Madison/gmu/routes/api"
	"github.com/DSSD-Madison/gmu/routes/pages"
)

// InitRoutes registers all the application routes
func InitRoutes(e *echo.Echo) {
	/*
		Frontend
	*/
	e.GET("/", pages.Home)
	e.GET("/search", pages.Search)

	e.GET("/results", pages.Results)

	e.GET("/test", pages.Test)

	/*
		API
	*/
	apiGroup := e.Group("/api")
	api.RegisterFiltersRoutes(apiGroup)
}

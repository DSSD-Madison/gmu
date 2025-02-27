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
	e.POST("/search/suggestions", SearchSuggestions)

	// Filters Route
	e.POST("/filters", Filters)

	// Document Management Routes
	e.GET("/document", DocumentView)

	e.GET("/document/new", DocumentNew)
	e.PUT("/document/new", DocumentNewPut)

	e.GET("/document/delete", DocumentDelete)
	e.DELETE("/document/delete", DocumentDeleteDelete)

	e.GET("/document/edit", DocumentEdit)
	e.PATCH("/document/edit", DocumentEditPatch)
}

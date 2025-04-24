package routes

import (
	"github.com/DSSD-Madison/gmu/pkg/handlers"
	"github.com/labstack/echo/v4"
)

func RegisterSearchRoutes(e *echo.Echo, searchHandler *handlers.SearchHandler) {
	e.GET("/search", searchHandler.Search)
}

package routes

import (
	"github.com/DSSD-Madison/gmu/internal/infra/http/handlers"
	"github.com/labstack/echo/v4"
)

func RegisterHomeRoutes(e *echo.Echo, homeHandler *handlers.HomeHandler) {
	e.GET("/", homeHandler.Home)
}

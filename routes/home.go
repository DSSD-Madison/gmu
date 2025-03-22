package routes

import (
	"net/http"

	"github.com/DSSD-Madison/gmu/web/components"
	"github.com/DSSD-Madison/gmu/pkg/awskendra"
	"github.com/labstack/echo/v4"
)

func Home(c echo.Context) error {
	return awskendra.Render(c, http.StatusOK, components.Home())
}
